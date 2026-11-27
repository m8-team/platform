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
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
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
	publisher, err := NewPublisher(db, DefaultConfig(topics))
	if err != nil {
		t.Fatal(err)
	}
	owner, err := NewTransactor(db, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		p := probes[c.alias]
		if err := owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
			return publisher.Publish(ctx, tx, c.alias, p.event)
		}); err != nil {
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
}
