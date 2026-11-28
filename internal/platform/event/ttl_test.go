package event

import (
	"context"
	"io"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"google.golang.org/protobuf/proto"
)

func TestProducerMetadataTTL(t *testing.T) {
	if os.Getenv("M8_EVENT_TEST_MONITOR_URL") == "" && os.Getenv("M8_EVENT_TEST_TTL_REQUIRED") != "1" {
		t.Skip("set M8_EVENT_TEST_MONITOR_URL for disposable YDB tablet monitoring")
	}
	ctx, db, path := liveDatabase(t)

	const shortTTL = 10 * time.Second
	const longTTL = 5 * time.Minute
	type observation struct {
		count    func(context.Context) int
		event    *eventv1.Event
		commitAt time.Time
	}
	cases := []struct {
		alias     string
		retention time.Duration
		ttl       time.Duration
	}{
		{"expired", 2 * time.Second, shortTTL},
		{"young", 2 * time.Second, longTTL},
		{"retained", time.Hour, shortTTL},
	}
	topics := make(map[string]string, len(cases))
	probes := make(map[string]observation, len(cases))
	for _, c := range cases {
		topic := path + "_" + c.alias
		topics[c.alias] = topic
		createFixtureTopic(t, ctx, db, topic,
			topicoptions.CreateWithMinActivePartitions(1),
			topicoptions.CreateWithRetentionPeriod(c.retention),
			topicoptions.CreateWithAttributes(map[string]string{
				"_message_group_seqno_retention_period_ms": strconv.FormatInt(c.ttl.Milliseconds(), 10),
			}),
		)
		capacity, ttl, count := observeSourceIDs(t, ctx, topic)
		if time.Duration(ttl)*time.Second != c.ttl || capacity < 100 || count(ctx) != 0 {
			t.Fatalf("%s fixture has unexpected TTL, capacity or initial Source IDs: %ds, %d", c.alias, ttl, capacity)
		}
		e := testEvent(t)
		e.Id = uuid.NewString()
		probes[c.alias] = observation{count: count, event: e}
	}
	state := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	publisher := livePublisher(t, ctx, db, path, DefaultConfig(topics), 10*time.Second)
	publish := func(alias string, event *eventv1.Event) error {
		return publisher.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
			if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+state+" (id) VALUES ($id);",
				ydb.ParamsBuilder().Param("$id").Text(event.Id).Build()); err != nil {
				return err
			}
			return tx.Publish(alias, event)
		})
	}
	for _, c := range cases {
		p := probes[c.alias]
		if err := publish(c.alias, p.event); err != nil {
			t.Fatal(err)
		}
		p.commitAt = time.Now()
		probes[c.alias] = p
		if got := p.count(ctx); got != 1 {
			t.Fatalf("%s has %d Source IDs after commit, want 1", c.alias, got)
		}
	}

	offsets := func(alias string) (int64, int64) {
		t.Helper()
		desc, err := db.Topic().Describe(ctx, topics[alias], topicoptions.IncludePartitionStats())
		if err != nil || len(desc.Partitions) != 1 {
			t.Fatalf("describe %s: partitions=%d, err=%v", alias, len(desc.Partitions), err)
		}
		r := desc.Partitions[0].PartitionStats.PartitionsOffset
		return r.Start, r.End
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	timeout, cancel := context.WithTimeout(ctx, 70*time.Second)
	defer cancel()
	short := probes["expired"]
	young := probes["young"]
	retained := probes["retained"]
	for {
		start, end := offsets("expired")
		youngStart, youngEnd := offsets("young")
		count := short.count(ctx)
		if end != 1 || youngEnd != 1 {
			t.Fatalf("unexpected end offsets: expired=%d young=%d", end, youngEnd)
		}
		if start == 1 && youngStart == 1 && count == 0 && time.Since(retained.commitAt) >= shortTTL {
			break
		}
		select {
		case <-timeout.Done():
			t.Fatalf("retention or TTL cleanup did not finish: expired start=%d Source IDs=%d young start=%d: %v",
				start, count, youngStart, timeout.Err())
		case <-ticker.C:
		}
	}
	if got := young.count(ctx); got != 1 {
		t.Fatalf("young Source ID evicted before its TTL: %d", got)
	}
	if start, end := offsets("retained"); start != 0 || end != 1 {
		t.Fatalf("long-retention message disappeared: [%d, %d)", start, end)
	}
	if got := retained.count(ctx); got != 1 {
		t.Fatalf("Source ID for retained message unexpectedly removed: %d", got)
	}

	reader, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(topics["retained"]))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		closeCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := reader.Close(closeCtx); err != nil {
			t.Error(err)
		}
	}()
	msg, err := reader.ReadMessage(timeout)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(msg)
	if err != nil {
		t.Fatal(err)
	}
	e := new(eventv1.Event)
	if err := proto.Unmarshal(data, e); err != nil || !proto.Equal(e, retained.event) {
		t.Fatalf("retained message did not round-trip: %v", err)
	}
	t.Logf("after %s: expired offset range=[1,1), Source IDs=0; young offset range=[1,1), Source IDs=1; retained offset range=[0,1), Source IDs=1, read offset=%d",
		time.Since(short.commitAt).Round(time.Millisecond), msg.Offset)

	reopened, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(topics["expired"]))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		closeCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := reopened.Close(closeCtx); err != nil {
			t.Error(err)
		}
	}()
	next := proto.CloneOf(short.event)
	next.Id = uuid.NewString()
	if err := publish("expired", next); err != nil {
		t.Fatal(err)
	}
	if _, end := offsets("expired"); end != 2 {
		t.Fatalf("same identity after eviction: topic end=%d, want 2", end)
	}
	read, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	msg, err = reopened.ReadMessage(read)
	if err != nil {
		t.Fatal(err)
	}
	data, err = io.ReadAll(msg)
	if err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(data, e); err != nil || !proto.Equal(e, next) ||
		msg.Offset != 1 || msg.ProducerID != publisher.producer+"-0" {
		t.Fatalf("message after source eviction: offset=%d producer=%q seq=%d error=%v",
			msg.Offset, msg.ProducerID, msg.SeqNo, err)
	}
	verifyFence(t, ctx, publisher.owner, publisher, publisher.ownership.uid, 1, 4)
	row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+state)
	if err != nil {
		t.Fatal(err)
	}
	var count uint64
	if err := row.Scan(&count); err != nil || count != 4 {
		t.Fatalf("unexpected committed state after producer metadata eviction: %d, %v", count, err)
	}
	t.Logf("reused the evicted producer at offset %d with sequence %d", msg.Offset, msg.SeqNo)

}
