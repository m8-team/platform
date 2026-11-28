package event

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func topicEnds(ctx context.Context, db *ydb.Driver, path string) (total int64, occupied int, err error) {
	description, err := db.Topic().Describe(ctx, path, topicoptions.IncludePartitionStats())
	if err != nil {
		return 0, 0, err
	}
	for _, partition := range description.Partitions {
		end := partition.PartitionStats.PartitionsOffset.End
		total += end
		if end > 0 {
			occupied++
		}
	}
	return total, occupied, nil
}

func TestIntegrationPartitionsAndReplicas(t *testing.T) {
	for _, partitions := range []int64{4, 16} {
		t.Run(fmt.Sprint(partitions), func(t *testing.T) {
			ctx, db, path := liveDatabase(t)
			_, replica, _ := liveDatabase(t)
			table := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
			createTestTopic(t, ctx, db, path, partitions)
			createTestTopic(t, ctx, db, path+"_other", 1)
			cfg := DefaultConfig(map[string]string{"primary": path, "other": path + "_other"})
			p := livePublisher(t, ctx, db, path, cfg, 30*time.Second)
			slot := p.ownership.slot
			slot.PodName = "publisher-1"
			q := bindTestPublisher(t, ctx, replica, slot, p.guard, cfg, 30*time.Second)
			check := func(first, second int64) error {
				for _, topic := range []struct {
					path string
					want int64
				}{{path, first}, {path + "_other", second}} {
					total, occupied, err := topicEnds(ctx, replica, topic.path)
					if err != nil {
						return err
					}
					if total != topic.want || topic.want >= 96 && occupied != int(partitions) {
						return fmt.Errorf("offset sum %d, want %d, occupied partitions %d", total, topic.want, occupied)
					}
				}
				return nil
			}
			events := make([]*eventv1.Event, 96)
			for i := range events {
				e := testEvent(t)
				e.Id = uuid.NewString()
				e.AggregateId = fmt.Sprintf("aggregate-%d", i)
				events[i] = e
			}
			rollback := errors.New("intentional multi-topic rollback")
			for _, fail := range []bool{true, false} {
				err := p.WithinTransaction(ctx, func(txCtx context.Context, tx *Transaction) error {
					if err := tx.Exec(txCtx, "INSERT INTO "+table+" (id) VALUES ('initial');", nil); err != nil {
						return err
					}
					if err := tx.Publish("primary", events...); err != nil {
						return err
					}
					if err := tx.Publish("other", events[0]); err != nil {
						return err
					}
					// The callback holds the transaction before commit while an
					// independent driver queries authoritative partition end offsets.
					if err := check(0, 0); err != nil {
						return err
					}
					if fail {
						return rollback
					}
					return nil
				})
				if fail {
					if !errors.Is(err, rollback) {
						t.Fatalf("rollback: %v", err)
					}
					if err := check(0, 0); err != nil {
						t.Fatal(err)
					}
					row, err := replica.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table)
					var count uint64
					if err != nil {
						t.Fatal(err)
					}
					if err := row.Scan(&count); err != nil || count != 0 {
						t.Fatalf("state after rollback: %d, %v", count, err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
			}
			if err := check(96, 1); err != nil {
				t.Fatal(err)
			}
			row, err := replica.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table)
			var count uint64
			if err != nil {
				t.Fatal(err)
			}
			if err := row.Scan(&count); err != nil || count != 1 {
				t.Fatalf("state after commit: %d, %v", count, err)
			}
			ordered := []*eventv1.Event{events[0], proto.CloneOf(events[0]), proto.CloneOf(events[0])}
			for _, e := range ordered[1:] {
				e.Id = uuid.NewString()
			}
			if err := q.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
				return tx.Publish("primary", ordered[1:]...)
			}); err != nil {
				t.Fatal(err)
			}
			start := make(chan struct{})
			results := make(chan error, 8)
			for i := range 8 {
				publisher := p
				if i%2 == 1 {
					publisher = q
				}
				e := testEvent(t)
				e.Id = uuid.NewString()
				if i >= 4 {
					e.AggregateId = fmt.Sprint(i)
				}
				go func() {
					<-start
					results <- publisher.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
						return tx.Publish("primary", e)
					})
				}()
			}
			close(start)
			for range 8 {
				if err := <-results; err != nil {
					t.Error(err)
				}
			}
			if err := check(106, 1); err != nil {
				t.Fatal(err)
			}
			reader, err := replica.Topic().StartReader(testConsumer, topicoptions.ReadTopic(path))
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := reader.Close(cleanup); err != nil {
					t.Error(err)
				}
			}()
			readCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			producers := make(map[string]struct{})
			ids := make(map[string]struct{}, 106)
			var index int
			var partition, offset, seq int64
			for range 106 {
				msg, err := reader.ReadMessage(readCtx)
				if err != nil {
					t.Fatal(err)
				}
				data, err := io.ReadAll(msg)
				if err != nil {
					t.Fatal(err)
				}
				e := new(eventv1.Event)
				if err := proto.Unmarshal(data, e); err != nil {
					t.Fatal(err)
				}
				if _, duplicate := ids[e.Id]; duplicate {
					t.Fatal("logical event duplicated")
				}
				ids[e.Id] = struct{}{}
				producers[msg.ProducerID] = struct{}{}
				if e.AggregateId != ordered[0].AggregateId {
					continue
				}
				if index >= len(ordered) || e.Id != ordered[index].Id {
					t.Fatal("unexpected order for one aggregate in sequential commits")
				}
				if index > 0 && (msg.PartitionID() != partition || msg.Offset <= offset) {
					t.Fatal("aggregate moved partition or offsets did not increase")
				}
				if index == 2 && msg.SeqNo <= seq {
					t.Fatal("sequence did not increase within one producer and transaction")
				}
				partition, offset, seq = msg.PartitionID(), msg.Offset, msg.SeqNo
				index++
			}
			if index != len(ordered) {
				t.Fatal("missing aggregate events")
			}
			t.Logf("%d partitions, 8 concurrent transactions across 2 drivers, %d distinct producer IDs for 106 committed events",
				partitions, len(producers))
		})
	}
}

func TestIntegrationDefaultByteLimits(t *testing.T) {
	ctx, db, path := liveDatabase(t)
	createTestTopic(t, ctx, db, path, 1)
	cfg := DefaultConfig(map[string]string{"events": path})
	p := livePublisher(t, ctx, db, path, cfg, 30*time.Second)
	events := make([]*eventv1.Event, cfg.MaxBatchBytes/cfg.MaxEventBytes)
	for i := range events {
		e := testEvent(t)
		e.Id = uuid.NewString()
		payload := strings.Repeat("x", cfg.MaxEventBytes)
		fullPayload, err := anypb.New(wrapperspb.String(payload))
		e.Payload = fullPayload
		if err != nil {
			t.Fatal(err)
		}
		e.Payload, err = anypb.New(wrapperspb.String(payload[:len(payload)-(proto.Size(e)-cfg.MaxEventBytes)]))
		if err != nil || proto.Size(e) != cfg.MaxEventBytes {
			t.Fatalf("size of limit fixture: got %d, error %v", proto.Size(e), err)
		}
		events[i] = e
	}
	publish := func(events ...*eventv1.Event) error {
		return p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
			_ = tx.Publish("events", events...)
			return nil // Failed staging must still abort.
		})
	}
	if err := publish(events...); err != nil {
		t.Fatal(err)
	}
	count, _, err := topicEnds(ctx, db, path)
	if err != nil || count != int64(len(events)) || p.Stats().StagedBytes != uint64(cfg.MaxBatchBytes) {
		t.Fatalf("batch at byte limit: count=%d stats=%+v error=%v", count, p.Stats(), err)
	}
	oversized := proto.CloneOf(events[0])
	payload := new(wrapperspb.StringValue)
	if err := events[0].Payload.UnmarshalTo(payload); err != nil {
		t.Fatal(err)
	}
	oversized.Payload, err = anypb.New(wrapperspb.String(payload.Value + "x"))
	if err != nil || proto.Size(oversized) != cfg.MaxEventBytes+1 {
		t.Fatalf("expected a serialized envelope one byte above the limit: %v", err)
	}
	if err := publish(oversized); !errors.Is(err, ErrEventTooLarge) {
		t.Fatalf("oversized envelope: %v", err)
	}

	extra := testEvent(t)
	extra.Id = uuid.NewString()
	last := proto.CloneOf(events[len(events)-1])
	last.Payload, err = anypb.New(wrapperspb.String(payload.Value[:len(payload.Value)-proto.Size(extra)+1]))
	if err != nil || proto.Size(last)+proto.Size(extra) != cfg.MaxEventBytes+1 {
		t.Fatalf("expected a serialized batch one byte above the limit: %v", err)
	}
	tooManyBytes := append(append([]*eventv1.Event(nil), events[:len(events)-1]...), last, extra)
	if err := publish(tooManyBytes...); !errors.Is(err, ErrBatchTooLarge) {
		t.Fatalf("oversized byte batch: %v", err)
	}
	small := make([]*eventv1.Event, cfg.MaxBatchEvents+1)
	for i := range small {
		small[i] = testEvent(t)
		small[i].Id = uuid.NewString()
	}
	if err := publish(small[:cfg.MaxBatchEvents]...); err != nil {
		t.Fatal(err)
	}
	if err := publish(small...); !errors.Is(err, ErrBatchTooLarge) {
		t.Fatalf("oversized event count: %v", err)
	}
	verifyFence(t, ctx, p.owner, p, p.ownership.uid, 1, 2)
	count, _, err = topicEnds(ctx, db, path)
	if err != nil || count != int64(len(events)+cfg.MaxBatchEvents) {
		t.Fatalf("unexpected topic offsets after limit tests: %d, %v", count, err)
	}
}
