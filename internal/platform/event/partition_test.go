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
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
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
			p, err := NewPublisher(db, cfg)
			if err != nil {
				t.Fatal(err)
			}
			q, err := NewPublisher(replica, cfg)
			if err != nil {
				t.Fatal(err)
			}
			tr, _ := NewTransactor(db, 30*time.Second)
			other, _ := NewTransactor(replica, 30*time.Second)
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
				err := tr.WithinTransaction(ctx, func(txCtx context.Context, tx query.TxActor) error {
					if err := tx.Exec(txCtx, "INSERT INTO "+table+" (id) VALUES ('initial');"); err != nil {
						return err
					}
					if err := p.Publish(txCtx, tx, "primary", events...); err != nil {
						return err
					}
					if err := p.Publish(txCtx, tx, "other", events[0]); err != nil {
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
			if err := other.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
				return q.Publish(ctx, tx, "primary", ordered[1:]...)
			}); err != nil {
				t.Fatal(err)
			}
			start := make(chan struct{})
			results := make(chan error, 8)
			for i := range 8 {
				publisher, owner := p, tr
				if i%2 == 1 {
					publisher, owner = q, other
				}
				e := testEvent(t)
				e.Id = uuid.NewString()
				if i >= 4 {
					e.AggregateId = fmt.Sprint(i)
				}
				go func() {
					<-start
					results <- owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
						return publisher.Publish(ctx, tx, "primary", e)
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
	createTestTopic(t, ctx, db, path, 4)
	cfg := DefaultConfig(map[string]string{"events": path})
	p, err := NewPublisher(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	tr, _ := NewTransactor(db, 30*time.Second)
	events := make([]*eventv1.Event, cfg.MaxBatchBytes/cfg.MaxEventBytes)
	for i := range events {
		e := testEvent(t)
		e.Id = uuid.NewString()
		e.AggregateId = fmt.Sprintf("aggregate-%d", i)
		payload := strings.Repeat("x", cfg.MaxEventBytes)
		e.Payload, err = anypb.New(wrapperspb.String(payload))
		if err != nil {
			t.Fatal(err)
		}
		e.Payload, err = anypb.New(wrapperspb.String(payload[:len(payload)-(proto.Size(e)-cfg.MaxEventBytes)]))
		if err != nil || proto.Size(e) != cfg.MaxEventBytes {
			t.Fatalf("size of limit fixture: got %d, error %v", proto.Size(e), err)
		}
		events[i] = e
	}
	if err := tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		return p.Publish(ctx, tx, "events", events...)
	}); err != nil {
		t.Fatal(err)
	}
	count, _, err := topicEnds(ctx, db, path)
	if err != nil || count != int64(len(events)) || p.Stats().StagedBytes != uint64(cfg.MaxBatchBytes) {
		t.Fatalf("batch at byte limit: count=%d stats=%+v error=%v", count, p.Stats(), err)
	}
}
