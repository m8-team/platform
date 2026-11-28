package event

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"buf.build/go/protovalidate"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var benchmarkRoutingKey string

func BenchmarkEventOperations(b *testing.B) {
	attrs := sampleAttributes(b)
	payload := wrapperspb.String(strings.Repeat("x", 1024))
	event, err := New(attrs, payload)
	if err != nil {
		b.Fatal(err)
	}

	for _, operation := range []struct {
		name string
		run  func() error
	}{
		{"new", func() error { _, err := New(attrs, payload); return err }},
		{"envelope", func() error { return protovalidate.Validate(event) }},
		{"payload", func() error { return protovalidate.Validate(payload) }},
		{"pack", func() error { _, err := anypb.New(payload); return err }},
		{"unpack", func() error { _, err := event.Payload.UnmarshalNew(); return err }},
		{"marshal", func() error { _, err := proto.Marshal(event); return err }},
		{"routing", func() error { benchmarkRoutingKey = routingKey(event); return nil }},
	} {
		b.Run(operation.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := operation.run(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkNativeYDB(b *testing.B) {
	for _, variant := range []struct {
		name                    string
		batch, payload, workers int
		partitions              int64
		lanes                   int
	}{
		{"state_only", 0, 0, 1, 1, 0},
		{"single_writer", 1, 256, 1, 1, 0},
		{"keyed_writer", 1, 256, 1, 1, 0},
		{"protected", 1, 256, 1, 1, 0},
		{"protected_contended", 10, 1024, 4, 4, 0},
		{"protected_batch_100", 100, 256, 1, 16, 0},
		{"stable_single_writer", 1, 256, 1, 1, 1},
		{"stable_keyed_writer", 1, 256, 1, 1, 1},
		{"stable_keyed_two_lanes", 1, 256, 4, 1, 2},
		{"batch_10", 10, 1024, 1, 4, 0},
		{"batch_50_concurrent", 50, 16384, 4, 16, 0},
		{"batch_100_concurrent", 100, 256, 16, 4, 0},
	} {
		b.Run(variant.name, func(b *testing.B) {
			var retries, writerInitNs, writerInits, producers atomic.Uint64
			var seenProducers sync.Map
			var queryExecs, ownershipNS, guardNS atomic.Uint64
			ctx, db, path := liveDatabase(b, ydb.WithTraceTopic(trace.Topic{
				OnWriterInitStream: func(info trace.TopicWriterInitStreamStartInfo) func(trace.TopicWriterInitStreamDoneInfo) {
					start := time.Now()
					if _, known := seenProducers.LoadOrStore(info.ProducerID, struct{}{}); !known {
						producers.Add(1)
					}
					return func(trace.TopicWriterInitStreamDoneInfo) {
						writerInitNs.Add(uint64(time.Since(start)))
						writerInits.Add(1)
					}
				},
			}), ydb.WithTraceQuery(trace.Query{
				OnTxQueryRow: func(info trace.QueryTxQueryRowStartInfo) func(trace.QueryTxQueryRowDoneInfo) {
					queryExecs.Add(1)
					start := time.Now()
					return func(trace.QueryTxQueryRowDoneInfo) {
						if strings.Contains(info.Query, "SELECT pod_uid") {
							ownershipNS.Add(uint64(time.Since(start)))
						} else if strings.Contains(info.Query, "SELECT revision") {
							guardNS.Add(uint64(time.Since(start)))
						}
					}
				},
				OnTxExec: func(info trace.QueryTxExecStartInfo) func(trace.QueryTxExecDoneInfo) {
					queryExecs.Add(1)
					start := time.Now()
					return func(trace.QueryTxExecDoneInfo) {
						if strings.Contains(info.Query, "_guard") {
							guardNS.Add(uint64(time.Since(start)))
						}
					}
				},
			}))
			table := createTestTable(b, ctx, db, path+"_state", "id Uint64 NOT NULL, PRIMARY KEY (id)")
			createTestTopic(b, ctx, db, path, variant.partitions)
			cfg := DefaultConfig(map[string]string{"events": path})
			tr, _ := NewTransactor(db, 30*time.Second)
			var protected *Publisher
			var before TransactionStats
			if strings.HasPrefix(variant.name, "protected") {
				protected = livePublisher(b, ctx, db, path, cfg, 30*time.Second)
				before = protected.owner.Stats()
			}
			template := testEvent(b)
			payload, err := anypb.New(wrapperspb.String(strings.Repeat("x", variant.payload)))
			if err != nil {
				b.Fatal(err)
			}
			template.Payload = payload
			durations := make([]time.Duration, b.N)
			var laneWait atomic.Uint64
			lanes := make(chan string, variant.lanes)
			for lane := range variant.lanes {
				lanes <- fmt.Sprintf("m8.test.bench.0.%d", lane)
			}
			failures := make(chan error, variant.workers)
			b.ReportAllocs()
			queryExecs.Store(0)
			ownershipNS.Store(0)
			guardNS.Store(0)
			b.ResetTimer()
			var workers sync.WaitGroup
			for worker := range variant.workers {
				workers.Go(func() {
					for i := worker; i < b.N; i += variant.workers {
						start := time.Now()
						events := make([]*eventv1.Event, variant.batch)
						for j := range events {
							e := proto.CloneOf(template)
							e.Id = fmt.Sprintf("34b169b7-3969-47f1-bc60-%012x", i*variant.batch+j)
							e.AggregateId = fmt.Sprintf("aggregate-%d", j)
							events[j] = e
						}
						params := ydb.ParamsBuilder().Param("$id").Uint64(uint64(i)).Build()
						producer := ""
						if variant.lanes > 0 {
							waitStart := time.Now()
							select {
							case producer = <-lanes:
							case <-ctx.Done():
								failures <- ctx.Err()
								return
							}
							laneWait.Add(uint64(time.Since(waitStart)))
						}
						writeState := func(ctx context.Context, tx query.TxActor) error {
							return tx.Exec(ctx, "DECLARE $id AS Uint64; INSERT INTO "+table+
								" (id) VALUES ($id);", query.WithParameters(params))
						}
						if protected != nil {
							err := protected.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor, send func(string, ...*eventv1.Event) error) error {
								if err := writeState(ctx, tx); err != nil {
									return err
								}
								return send("events", events...)
							})
							if err != nil {
								failures <- err
								return
							}
							durations[i] = time.Since(start)
							continue
						}
						attempt := 0
						err := tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
							attempt++
							if attempt > 1 {
								retries.Add(1)
							}
							if err := writeState(ctx, tx); err != nil {
								return err
							}
							switch variant.name {
							case "state_only":
								return nil
							case "stable_keyed_writer", "stable_keyed_two_lanes":
								return writeWithProducer(ctx, db, tx, path, producer, events...)
							case "single_writer", "stable_single_writer":
								messages, _, err := encodeBatch(ctx, cfg, events)
								if err != nil {
									return err
								}
								options := []topicoptions.WriterOption{topicoptions.WithWriterMessageMaxBytesSize(cfg.MaxEventBytes)}
								if producer != "" {
									options = append(options, topicoptions.WithWriterProducerID(producer))
								}
								writer, err := db.Topic().StartTransactionalWriter(tx, path, options...)
								if err != nil {
									return err
								}
								messages[0].Key = ""
								return writer.Write(ctx, messages...)
							default:
								return writeWithProducer(ctx, db, tx, path, "", events...)
							}
						})
						if producer != "" {
							lanes <- producer
						}
						if err != nil {
							failures <- err
							return
						}
						durations[i] = time.Since(start)
					}
				})
			}
			workers.Wait()
			b.StopTimer()
			close(failures)
			for err := range failures {
				b.Fatal(err)
			}
			end, _, err := topicEnds(ctx, db, path)
			if err != nil || end != int64(b.N*variant.batch) {
				b.Fatalf("topic end %d, want %d, error %v", end, b.N*variant.batch, err)
			}
			if variant.lanes > 0 {
				b.ReportMetric(float64(laneWait.Load())/float64(b.N)/float64(time.Millisecond), "lane-wait-ms/tx")
			}
			ownerStats := tr.Stats()
			if protected != nil {
				ownerStats = protected.owner.Stats()
				retries.Store(ownerStats.Retries - before.Retries)
				b.ReportMetric(float64(protected.Stats().LaneWait)/float64(b.N)/float64(time.Millisecond), "lane-wait-ms/tx")
				b.ReportMetric(float64(ownershipNS.Load())/float64(b.N)/float64(time.Millisecond), "ownership-ms/tx")
				b.ReportMetric(float64(guardNS.Load())/float64(b.N)/float64(time.Millisecond), "guard-ms/tx")
			}
			b.ReportMetric(float64(queryExecs.Load())/float64(b.N), "queries/tx")
			b.ReportMetric(float64(ownerStats.Duration-before.Duration)/float64(b.N)/float64(time.Millisecond), "ydb-ms/tx")
			b.ReportMetric(float64(ownerStats.CommitConflicts-before.CommitConflicts)/float64(b.N), "commit-conflicts/tx")
			slices.Sort(durations)
			for _, percentile := range []int{50, 95, 99} {
				index := (b.N*percentile+99)/100 - 1
				b.ReportMetric(float64(durations[index])/float64(time.Millisecond), fmt.Sprintf("p%d-ms", percentile))
			}
			tps := float64(b.N) / b.Elapsed().Seconds()
			b.ReportMetric(tps, "tx/s")
			b.ReportMetric(tps*float64(variant.batch), "events/s")
			b.ReportMetric(float64(retries.Load())/float64(b.N), "retry/tx")
			b.ReportMetric(0, "errors/tx")
			if count := writerInits.Load(); count > 0 {
				b.ReportMetric(float64(writerInitNs.Load())/float64(count)/float64(time.Millisecond), "init-ms/stream")
				b.ReportMetric(float64(producers.Load())/float64(b.N), "producer-ids/tx")
			}
		})
	}
}
