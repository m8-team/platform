package event

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestIntegrationActivePartitionSplit(t *testing.T) {
	ctx, db, path := liveDatabase(t)
	stableLoad := os.Getenv("M8_EVENT_TEST_STABLE_SPLIT_LOAD") == "1"
	batchSize, utilization := 8, 30
	if stableLoad {
		batchSize, utilization = 24, 15
	}
	createFixtureTopic(t, ctx, db, path,
		topicoptions.CreateWithMinActivePartitions(1),
		topicoptions.CreateWithMaxActivePartitions(2),
		topicoptions.CreateWithSupportedCodecs(topictypes.CodecRaw),
		topicoptions.CreateWithPartitionWriteSpeedBytesPerSecond(256<<10),
		topicoptions.CreateWithPartitionWriteBurstBytes(256<<10),
		topicoptions.CreateWithAutoPartitioningSettings(topictypes.AutoPartitioningSettings{
			AutoPartitioningStrategy: topictypes.AutoPartitioningStrategyScaleUp,
			AutoPartitioningWriteSpeedStrategy: topictypes.AutoPartitioningWriteSpeedStrategy{
				StabilizationWindow:  2 * time.Second,
				UpUtilizationPercent: int32(utilization),
			},
		}),
	)

	p := livePublisher(t, ctx, db, path, DefaultConfig(map[string]string{"events": path}), 20*time.Second)
	replicas := []*Publisher{p}
	if stableLoad {
		for i := 1; i < 4; i++ {
			slot := p.ownership.slot
			slot.PodName = fmt.Sprintf("publisher-%d", i)
			replicas = append(replicas, bindTestPublisher(t, ctx, db, slot, p.guard, p.cfg, 20*time.Second))
		}
	}
	tr, err := NewTransactor(db, 20*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	producer := p.producer
	initial, err := db.Topic().Describe(ctx, path)
	if err != nil || len(initial.Partitions) != 1 || !initial.Partitions[0].Active {
		t.Fatalf("expected one active parent partition: %v, %v", initial.Partitions, err)
	}
	parent := initial.Partitions[0].PartitionID
	payload, err := anypb.New(wrapperspb.String(strings.Repeat("x", 32<<10)))
	if err != nil {
		t.Fatal(err)
	}
	template := testEvent(t)
	template.Payload = payload
	expected := make(map[string]string)
	batches := make(map[string]uint64)
	publish := func(ctx context.Context, count int, distinctKeys, stable bool, publisher *Publisher) string {
		t.Helper()
		events := make([]*eventv1.Event, count)
		for i := range events {
			e := &eventv1.Event{
				Id: uuid.NewString(), Type: template.Type, Source: template.Source,
				AggregateId: template.AggregateId, AggregateType: template.AggregateType,
				AggregateRevision: template.AggregateRevision, OccurredAt: template.OccurredAt,
				Payload: payload,
			}
			if distinctKeys {
				e.AggregateId = fmt.Sprintf("aggregate-%d", i)
			}
			events[i] = e
		}
		var err error
		if stable {
			err = publisher.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
				return tx.Publish("events", events...)
			})
		} else {
			err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
				return writeWithProducer(ctx, db, tx, path, "", events...)
			})
		}
		if err != nil {
			t.Fatalf("publish during split fixture after %d confirmed events: %v", len(expected), err)
		}
		if stable {
			batches[publisher.producer]++
		}
		for _, e := range events {
			expected[e.Id] = publisher.producer
		}
		return events[0].Id
	}
	first := publish(ctx, 1, false, true, p)
	loadCtx, stop := context.WithTimeout(ctx, 80*time.Second)
	defer stop()
	var children map[int64]bool
	start := time.Now()
	// The protected workload alternates independent slots while each slot
	// reuses its stable identity and serializes its own writes.
	for batch := 0; children == nil; batch++ {
		if loadCtx.Err() != nil {
			t.Fatalf("no split after %s with %d committed events (32 KiB payload, 256 KiB/s partition, %d%% threshold, batch %d)", time.Since(start), len(expected), utilization, batchSize)
		}
		publish(loadCtx, batchSize, false, stableLoad, replicas[batch%len(replicas)])
		desc, err := db.Topic().Describe(loadCtx, path)
		if err != nil {
			t.Fatal(err)
		}
		for _, partition := range desc.Partitions {
			if partition.PartitionID == parent && !partition.Active && len(partition.ChildPartitionIDs) == 2 {
				children = make(map[int64]bool, 2)
				for _, child := range partition.ChildPartitionIDs {
					children[child] = false
				}
				break
			}
		}
	}
	t.Logf("active split observed in %s after %d committed events", time.Since(start).Round(time.Millisecond), len(expected))

	after := publish(ctx, 1, false, true, p)
	publish(ctx, 32, true, true, p)
	description, err := db.Topic().Describe(ctx, path, topicoptions.IncludePartitionStats())
	if err != nil {
		t.Fatal(err)
	}
	var retained int64
	ends := make(map[int64]int64, len(description.Partitions))
	for _, partition := range description.Partitions {
		offsets := partition.PartitionStats.PartitionsOffset
		retained += offsets.End - offsets.Start
		ends[partition.PartitionID] = offsets.End
	}
	if retained != int64(len(expected)) {
		t.Fatalf("retained records after split: got %d, want %d", retained, len(expected))
	}
	startReader := func() *topicreader.Reader {
		t.Helper()
		reader, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(path),
			topicoptions.WithReaderCommitMode(topicoptions.CommitModeSync))
		if err != nil {
			t.Fatal(err)
		}
		return reader
	}
	reader := startReader()
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := reader.Close(closeCtx); err != nil {
			t.Error(err)
		}
	}()
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	seenParent := false
	sequences := make(map[string]int64)
	for len(expected) != 0 {
		msg, err := reader.ReadMessage(readCtx)
		if err != nil {
			t.Fatalf("%d committed events not read across split: %v", len(expected), err)
		}
		data, err := io.ReadAll(msg)
		if err != nil {
			t.Fatal(err)
		}
		event := new(eventv1.Event)
		if err := proto.Unmarshal(data, event); err != nil {
			t.Fatal(err)
		}
		expectedPrefix, ok := expected[event.Id]
		if !ok {
			t.Fatalf("unexpected or duplicate event at partition %d offset %d", msg.PartitionID(), msg.Offset)
		}
		delete(expected, event.Id)
		_, child := children[msg.PartitionID()]
		stableID := fmt.Sprintf("%s-%d", producer, msg.PartitionID())
		if !proto.Equal(event.Payload, payload) {
			t.Fatalf("unexpected payload for %s", event.Id)
		}
		if stableLoad {
			if msg.ProducerID != fmt.Sprintf("%s-%d", expectedPrefix, msg.PartitionID()) {
				t.Fatalf("expected prefix %q, got %q", expectedPrefix, msg.ProducerID)
			}
			if last := sequences[msg.ProducerID]; msg.SeqNo <= last {
				t.Fatalf("sequence for %s: got %d after %d", msg.ProducerID, msg.SeqNo, last)
			}
			sequences[msg.ProducerID] = msg.SeqNo
		}
		switch event.Id {
		case first:
			if msg.ProducerID != stableID {
				t.Fatalf("pre-split producer: got %q, want %q", msg.ProducerID, stableID)
			}
			if msg.PartitionID() != parent {
				t.Fatal("pre-split sentinel was not read from parent")
			}
			seenParent = true
		case after:
			if msg.ProducerID != stableID {
				t.Fatalf("post-split producer: got %q, want %q", msg.ProducerID, stableID)
			}
			if !seenParent || !child {
				t.Fatal("post-split sentinel must follow parent delivery and come from a child")
			}
		}
		if child {
			children[msg.PartitionID()] = true
		}
		if err := reader.Commit(readCtx, msg); err != nil {
			t.Fatal(err)
		}
	}
	for id, read := range children {
		if !read {
			t.Fatalf("no event read from child %d", id)
		}
	}
	if stableLoad {
		if len(sequences) > len(replicas)*len(ends) {
			t.Fatalf("source budget exceeded: got %d distinct producer IDs", len(sequences))
		}
		if os.Getenv("M8_EVENT_TEST_MONITOR_URL") != "" {
			_, _, sample := observePartitionSourceIDs(t, ctx, path, len(ends))
			count := sample(ctx)
			if count == 0 || count > len(replicas)*len(ends) {
				t.Fatalf("retained sources outside expected bounds: %d", count)
			}
			t.Logf("split retained %d source IDs and delivered messages from %d source IDs", count, len(sequences))
		}
	}
	if err := reader.Close(readCtx); err != nil {
		t.Fatal(err)
	}
	position, err := db.Topic().DescribeTopicConsumer(ctx, path, testConsumer, topicoptions.IncludeConsumerStats())
	if err != nil {
		t.Fatal(err)
	}
	if len(position.Partitions) != len(ends) {
		t.Fatalf("consumer describes %d partitions, want %d", len(position.Partitions), len(ends))
	}
	for _, partition := range position.Partitions {
		want, known := ends[partition.PartitionID]
		if got := partition.PartitionConsumerStats.CommittedOffset; !known || got != want {
			t.Fatalf("partition %d: durable offset %d, want %d", partition.PartitionID, got, want)
		}
	}
	reader = startReader()
	id := publish(ctx, 1, false, true, p)
	msg, err := reader.ReadMessage(readCtx)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(msg)
	if err != nil {
		t.Fatal(err)
	}
	resumed := new(eventv1.Event)
	_, child := children[msg.PartitionID()]
	if err := proto.Unmarshal(data, resumed); err != nil || resumed.Id != id ||
		!proto.Equal(resumed.Payload, payload) || !child || msg.Offset != ends[msg.PartitionID()] {
		t.Fatalf("consumer resumed at %q (want %q), partition %d offset %d (want %d): %v",
			resumed.Id, id, msg.PartitionID(), msg.Offset, ends[msg.PartitionID()], err)
	}
	if stableLoad && (msg.ProducerID != fmt.Sprintf("%s-%d", p.producer, msg.PartitionID()) ||
		msg.SeqNo <= sequences[msg.ProducerID]) {
		t.Fatalf("resumed publication did not continue its stable source sequence: %d", msg.SeqNo)
	}
	for _, publisher := range replicas {
		verifyFence(t, ctx, publisher.owner, publisher, publisher.ownership.uid, 1, batches[publisher.producer])
	}
}
