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
	expected := make(map[string]struct{})
	publish := func(ctx context.Context, count int, distinctKeys, stable bool) string {
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
			err = p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
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
		for _, e := range events {
			expected[e.Id] = struct{}{}
		}
		return events[0].Id
	}
	first := publish(ctx, 1, false, true)
	loadCtx, stop := context.WithTimeout(ctx, 80*time.Second)
	defer stop()
	var children map[int64]bool
	start := time.Now()
	// Use the existing split-inducing workload by default. The opt-in case
	// also reuses the sentinel's producer identity for all pressure writes.
	for children == nil {
		if loadCtx.Err() != nil {
			t.Fatalf("no split after %s with %d committed events (32 KiB payload, 256 KiB/s partition, %d%% threshold, batch %d)", time.Since(start), len(expected), utilization, batchSize)
		}
		publish(loadCtx, batchSize, false, stableLoad)
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

	after := publish(ctx, 1, false, true)
	publish(ctx, 32, true, true)
	description, err := db.Topic().Describe(ctx, path, topicoptions.IncludePartitionStats())
	if err != nil {
		t.Fatal(err)
	}
	var retained int64
	for _, partition := range description.Partitions {
		offsets := partition.PartitionStats.PartitionsOffset
		retained += offsets.End - offsets.Start
	}
	if retained != int64(len(expected)) {
		t.Fatalf("retained records after split: got %d, want %d", retained, len(expected))
	}
	reader, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(path))
	if err != nil {
		t.Fatal(err)
	}
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
		if _, ok := expected[event.Id]; !ok {
			t.Fatalf("unexpected or duplicate event at partition %d offset %d", msg.PartitionID(), msg.Offset)
		}
		delete(expected, event.Id)
		_, child := children[msg.PartitionID()]
		stableID := fmt.Sprintf("%s-%d", producer, msg.PartitionID())
		if !proto.Equal(event.Payload, payload) {
			t.Fatalf("unexpected payload for %s", event.Id)
		}
		if stableLoad && msg.ProducerID != stableID {
			t.Fatalf("expected stable source %q, got %q", stableID, msg.ProducerID)
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
}
