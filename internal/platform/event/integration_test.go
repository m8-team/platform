package event

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Topic_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const testConsumer = "event-test"

var fixturePath = regexp.MustCompile(`^(/[a-zA-Z0-9_][a-zA-Z0-9_.-]*)+$`)

func liveDatabase(t testing.TB, options ...ydb.Option) (context.Context, *ydb.Driver, string) {
	t.Helper()
	dsn, prefix := os.Getenv("M8_EVENT_TEST_YDB_DSN"), os.Getenv("M8_EVENT_TEST_YDB_PREFIX")
	if dsn == "" && prefix == "" && os.Getenv("M8_EVENT_TEST_YDB_REQUIRED") != "1" {
		t.Skip("disposable YDB not configured")
	}
	endpoint, err := url.Parse(dsn)
	if err != nil || !fixturePath.MatchString(prefix) || !fixturePath.MatchString(endpoint.Path) ||
		(prefix != endpoint.Path && !strings.HasPrefix(prefix, endpoint.Path+"/")) {
		t.Fatal("invalid disposable YDB DSN or fixture prefix")
	}
	switch endpoint.Hostname() {
	case "localhost", "127.0.0.1", "::1":
	default:
		t.Fatal("fixture requires a disposable loopback YDB endpoint")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	openCtx, stop := context.WithTimeout(ctx, 45*time.Second)
	defer stop()
	db, err := ydb.Open(openCtx, dsn, options...)
	if err != nil {
		t.Fatalf("connect to disposable YDB: %v", err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Close(cleanup); err != nil {
			t.Errorf("close fixture driver: %v", err)
		}
	})
	return ctx, db, prefix + "/event_" + uuid.NewString()
}

func createTestTable(t testing.TB, ctx context.Context, db *ydb.Driver, name, columns string) string {
	t.Helper()
	table := "`" + name + "`"
	if err := db.Query().Exec(ctx, "CREATE TABLE "+table+" ("+columns+");"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Query().Exec(cleanup, "DROP TABLE "+table+";"); err != nil {
			t.Errorf("drop test table: %v", err)
		}
	})
	return table
}

func createTestTopic(t testing.TB, ctx context.Context, db *ydb.Driver, path string, partitions int64) {
	t.Helper()
	createFixtureTopic(t, ctx, db, path,
		topicoptions.CreateWithMinActivePartitions(partitions),
		topicoptions.CreateWithMaxActivePartitions(partitions),
		topicoptions.CreateWithAutoPartitioningSettings(topictypes.AutoPartitioningSettings{
			AutoPartitioningStrategy: topictypes.AutoPartitioningStrategyScaleUp,
		}),
	)
}

func createFixtureTopic(t testing.TB, ctx context.Context, db *ydb.Driver, path string, options ...topicoptions.CreateOption) {
	t.Helper()
	options = append(options, topicoptions.CreateWithConsumer(topictypes.Consumer{Name: testConsumer}))
	if err := db.Topic().Create(ctx, path, options...); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Topic().Drop(cleanup, path); err != nil {
			t.Errorf("drop test topic: %v", err)
		}
	})
}

func TestIntegrationAtomicPublication(t *testing.T) {
	ctx, db, topicPath := liveDatabase(t)
	table := createTestTable(t, ctx, db, topicPath+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	createTestTopic(t, ctx, db, topicPath, 1)
	createTestTopic(t, ctx, db, topicPath+"_forwarded", 1)
	cfg := DefaultConfig(map[string]string{"resource": topicPath, "forwarded": topicPath + "_forwarded"})
	cfg.Topics["missing"] = topicPath + "_missing"
	p := livePublisher(t, ctx, db, topicPath, cfg, 25*time.Second)
	writeState := func(ctx context.Context, tx *Transaction, id string) error {
		return tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+table+" (id) VALUES ($id);",
			ydb.ParamsBuilder().Param("$id").Text(id).Build())
	}
	var err error
	assertCount := func(want uint64) {
		t.Helper()
		row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table+";")
		if err != nil {
			t.Fatal(err)
		}
		var got uint64
		if err := row.Scan(&got); err != nil || got != want {
			t.Fatalf("state count got=%d want=%d err=%v", got, want, err)
		}
	}
	makeEvent := func() *eventv1.Event {
		e := testEvent(t)
		e.Id = uuid.NewString()
		return e
	}
	rollback := errors.New("intentional rollback")
	rolledBack := makeEvent()
	err = p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := writeState(ctx, tx, rolledBack.Id); err != nil {
			return err
		}
		if err := tx.Publish("resource", rolledBack); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("expected rollback after staged event, got %v", err)
	}
	assertCount(0)

	bad := makeEvent()
	bad.Type = "invalid"
	err = p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := writeState(ctx, tx, bad.Id); err != nil {
			return err
		}
		return tx.Publish("resource", bad)
	})
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected validation failure, got %v", err)
	}
	assertCount(0)

	missing := makeEvent()
	err = p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := writeState(ctx, tx, missing.Id); err != nil {
			return err
		}
		return tx.Publish("missing", missing)
	})
	if err == nil {
		t.Fatal("unexpected success for nonexistent topic")
	}
	assertCount(0)

	first, second := makeEvent(), makeEvent()
	err = p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := writeState(ctx, tx, first.Id); err != nil {
			return err
		}
		return tx.Publish("resource", first, second)
	})
	if err != nil {
		t.Fatalf("commit state and events: %v", err)
	}
	assertCount(1)

	reader, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(topicPath))
	if err != nil {
		t.Fatal(err)
	}
	closeReader := func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := reader.Close(cleanup); err != nil {
			t.Errorf("close fixture reader: %v", err)
		}
	}
	defer closeReader()
	restartReader := func() {
		closeReader()
		next, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(topicPath))
		if err != nil {
			t.Fatal(err)
		}
		reader = next
	}
	expected := map[string]*eventv1.Event{first.Id: first, second.Id: second}
	consume := func(ctx context.Context, tx *Transaction, output string) error {
		batch, err := tx.ReadBatch(ctx, reader, 8*time.Second)
		if err != nil {
			return err
		}
		forward := make([]*eventv1.Event, 0, len(batch.Messages))
		for _, msg := range batch.Messages {
			data, err := io.ReadAll(msg)
			if err != nil {
				return err
			}
			got := new(eventv1.Event)
			if err := proto.Unmarshal(data, got); err != nil {
				return err
			}
			if !proto.Equal(expected[got.Id], got) {
				return errors.New("unexpected or rolled-back event visible")
			}
			if err := writeState(ctx, tx, "consumed-"+got.Id); err != nil {
				return err
			}
			forward = append(forward, got)
		}
		return tx.Publish(output, forward...)
	}
	if err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := consume(ctx, tx, "forwarded"); err != nil {
			return err
		}
		return rollback
	}); !errors.Is(err, rollback) {
		t.Fatalf("consumer rollback: %v", err)
	}
	assertCount(1)
	restartReader()
	if err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		return consume(ctx, tx, "unconfigured")
	}); !errors.Is(err, ErrInvalidTopic) {
		t.Fatalf("consumer rollback after output error: %v", err)
	}
	assertCount(1)
	if total, _, err := topicEnds(ctx, db, topicPath+"_forwarded"); err != nil || total != 0 {
		t.Fatalf("output survived rollback: %d, %v", total, err)
	}
	var retryInjected bool
	var replayPending bool
	recoveryStart := time.Now()
	for i := 0; i < 2; i++ {
		row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table+";")
		if err != nil {
			t.Fatal(err)
		}
		var count uint64
		if err := row.Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count == 3 {
			break
		}
		if err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
			if replayPending {
				replayPending = false
				// A fresh SDK reader starts from the durable offset. The
				// automatic reconnect after consecutive rollbacks can exceed
				// the read budget in the pinned SDK.
				restartReader()
			}
			if err := consume(ctx, tx, "forwarded"); err != nil {
				return err
			}
			if !retryInjected {
				retryInjected, replayPending = true, true
				return retry.RetryableError(errors.New("retry after staging input and output"))
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	assertCount(3)
	if !retryInjected || replayPending {
		t.Fatal("expected the SDK to retry the consuming callback")
	}
	t.Logf("reader recovery and replay after output failure and injected retry: %s", time.Since(recoveryStart))
	total, _, err := topicEnds(ctx, db, topicPath+"_forwarded")
	if err != nil || total != 2 {
		t.Fatalf("consumer publication and state must commit with offsets: %d, %v", total, err)
	}
	restartReader()
	if err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		// Even an ignored read error must abort the attempt.
		_, _ = tx.ReadBatch(ctx, reader, 250*time.Millisecond)
		return nil
	}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("empty input did not abort on bounded wait: %v", err)
	}
	sentinel := makeEvent()
	if err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		return tx.Publish("resource", sentinel)
	}); err != nil {
		t.Fatal(err)
	}
	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(msg)
	if err != nil {
		t.Fatal(err)
	}
	got := new(eventv1.Event)
	if err := proto.Unmarshal(data, got); err != nil || !proto.Equal(got, sentinel) {
		t.Fatalf("restarted reader did not begin at sentinel: %v", err)
	}
}

func TestIntegrationTransactionDeadline(t *testing.T) {
	ctx, db, path := liveDatabase(t)
	table := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	createTestTopic(t, ctx, db, path, 1)
	p := livePublisher(t, ctx, db, path, DefaultConfig(map[string]string{"events": path}), 3*time.Second)
	event := testEvent(t)
	staged := false
	err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := tx.Exec(ctx, "INSERT INTO "+table+" (id) VALUES ('deadline');", nil); err != nil {
			return err
		}
		if err := tx.Publish("events", event); err != nil {
			return err
		}
		staged = true
		<-ctx.Done()
		return ctx.Err()
	})
	if !staged || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline after staging: staged=%t, error=%v", staged, err)
	}
	row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table)
	if err != nil {
		t.Fatal(err)
	}
	var count uint64
	if err := row.Scan(&count); err != nil || count != 0 {
		t.Fatalf("state after deadline: %d, %v", count, err)
	}
	total, _, err := topicEnds(ctx, db, path)
	if err != nil || total != 0 {
		t.Fatalf("topic after deadline: %d, %v", total, err)
	}
	if stats := p.Stats(); stats.StagedEvents != 1 || stats.Errors != 0 {
		t.Fatalf("publication metrics must describe staging even across rollback: %+v", stats)
	}
}

type failingWriteStream struct {
	grpc.ClientStream
	failAfter <-chan struct{}
	injected  *atomic.Bool
}

func (s failingWriteStream) RecvMsg(message any) error {
	if err := s.ClientStream.RecvMsg(message); err != nil {
		return err
	}
	if reply, ok := message.(*Ydb_Topic.StreamWriteMessage_FromServer); ok && reply.GetWriteResponse() != nil {
		select {
		case <-s.failAfter:
			s.injected.Store(true)
			return status.Error(codes.PermissionDenied, "injected topic write acknowledgement error")
		case <-s.Context().Done():
			return s.Context().Err()
		}
	}
	return nil
}

func TestIntegrationWriterFlushFailure(t *testing.T) {
	release := make(chan struct{})
	var injected atomic.Bool
	intercept := func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn,
		method string, stream grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		result, err := stream(ctx, desc, conn, method, opts...)
		if err == nil && method == Ydb_Topic_V1.TopicService_StreamWrite_FullMethodName {
			return failingWriteStream{ClientStream: result, failAfter: release, injected: &injected}, nil
		}
		return result, err
	}
	ctx, db, path := liveDatabase(t, ydb.With(config.WithGrpcOptions(grpc.WithChainStreamInterceptor(intercept))))
	table := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	createTestTopic(t, ctx, db, path, 1)
	createTestTopic(t, ctx, db, path+"_other", 1)
	p := livePublisher(t, ctx, db, path, DefaultConfig(map[string]string{"events": path, "other": path + "_other"}), 8*time.Second)
	callbackReturned := false
	err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := tx.Exec(ctx, "INSERT INTO "+table+" (id) VALUES ('writer-failure');", nil); err != nil {
			return err
		}
		if err := tx.Publish("events", testEvent(t)); err != nil {
			return err
		}
		if err := tx.Publish("other", testEvent(t)); err != nil {
			return err
		}
		if !callbackReturned {
			close(release)
			callbackReturned = true
		}
		return nil
	})
	if err == nil || !callbackReturned || !injected.Load() {
		t.Fatalf("expected a writer flush failure after callback success: returned=%t, error=%v", callbackReturned, err)
	}
	t.Logf("flush failure after callback success: %s", strings.SplitN(err.Error(), " at ", 2)[0])
	row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table)
	if err != nil {
		t.Fatal(err)
	}
	var count uint64
	if err := row.Scan(&count); err != nil || count != 0 {
		t.Fatalf("state after flush failure: %d, %v", count, err)
	}
	for _, topic := range []string{path, path + "_other"} {
		total, _, err := topicEnds(ctx, db, topic)
		if err != nil || total != 0 {
			t.Fatalf("topic after flush failure: %d, %v", total, err)
		}
	}
}

func livePublisher(t testing.TB, ctx context.Context, db *ydb.Driver, prefix string, cfg Config, timeout time.Duration) *Publisher {
	t.Helper()
	createTestTable(t, ctx, db, prefix+"_owner", "replica Utf8 NOT NULL, pod_uid Utf8 NOT NULL, generation Uint64 NOT NULL, PRIMARY KEY (replica)")
	createTestTable(t, ctx, db, prefix+"_guard", "replica Utf8 NOT NULL, revision Uint64 NOT NULL, PRIMARY KEY (replica)")
	return bindTestPublisher(t, ctx, db, ReplicaSlot{"test", uuid.NewString(), "publisher-0", prefix + "_owner"}, prefix+"_guard", cfg, timeout)
}

func bindTestPublisher(t testing.TB, ctx context.Context, db *ydb.Driver, slot ReplicaSlot, guard string, cfg Config, timeout time.Duration) *Publisher {
	t.Helper()
	owner, err := NewTransactor(db, timeout)
	if err != nil {
		t.Fatal(err)
	}
	uid := uuid.NewString()
	if err := slot.Reconcile(ctx, owner, func(context.Context) (string, bool, error) { return uid, false, nil }); err != nil {
		t.Fatal(err)
	}
	claim, err := slot.Bind(ctx, owner, uid)
	if err != nil {
		t.Fatal(err)
	}
	p, err := newPublisher(owner, cfg, claim, guard)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func verifyFence(t testing.TB, ctx context.Context, owner *Transactor, p *Publisher, uid string, generation, revision uint64) {
	t.Helper()
	if err := owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		gotUID, gotGeneration, err := p.ownership.slot.read(ctx, tx)
		if err != nil {
			return err
		}
		gotRevision, err := p.guardRevision(ctx, tx)
		if err != nil {
			return err
		}
		if gotUID != uid || gotGeneration != generation || gotRevision != revision {
			return fmt.Errorf("unexpected fence: uid=%q generation=%d revision=%d; want %q/%d/%d", gotUID, gotGeneration, gotRevision, uid, generation, revision)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationEventSchema(t *testing.T) {
	ctx, db, prefix := liveDatabase(t)
	schema, err := os.ReadFile("../../../deploy/event/schema.yql")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Scheme().MakeDirectory(ctx, prefix); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, table := range []string{"replica_owners", "replica_publication_guards"} {
			_ = db.Query().Exec(cleanup, "DROP TABLE IF EXISTS `"+prefix+"/"+table+"`;")
		}
		_ = db.Scheme().RemoveDirectory(cleanup, prefix)
	}()
	migrate := func() {
		t.Helper()
		if err := db.Query().Exec(ctx, "PRAGMA TablePathPrefix = '"+prefix+"'; "+string(schema)); err != nil {
			t.Fatal(err)
		}
	}
	migrate()
	slot := ReplicaSlot{"test", uuid.NewString(), "publisher-0", prefix + "/replica_owners"}
	p := bindTestPublisher(t, ctx, db, slot, prefix+"/replica_publication_guards", DefaultConfig(map[string]string{"test": prefix + "/test"}), 5*time.Second)
	var escaped func(string, ...*eventv1.Event) error
	if err := p.WithinTransaction(ctx, func(_ context.Context, tx *Transaction) error {
		if _, ok := any(tx).(query.Transaction); ok {
			t.Error("business callback can explicitly commit")
		}
		if _, ok := any(tx).(query.TxActor); ok {
			t.Error("business callback can unwrap the SDK transaction handle")
		}
		escaped = tx.Publish
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := escaped("test"); !errors.Is(err, ErrTransactionClosed) {
		t.Fatalf("expired publication callback: %v", err)
	}
	migrate()
	verifyFence(t, ctx, p.owner, p, p.ownership.uid, 1, 1)
	if err := p.ownership.slot.Reconcile(ctx, p.owner, func(context.Context) (string, bool, error) { return "", true, nil }); err != nil {
		t.Fatal(err)
	}
	if err := p.CheckReadiness(ctx); !errors.Is(err, ErrReplicaOwnership) {
		t.Fatalf("retired publisher is ready: %v", err)
	}
	verifyFence(t, ctx, p.owner, p, retiredReplica, 2, 1)
}

func TestIntegrationNativePrematureCommit(t *testing.T) {
	ctx, db, path := liveDatabase(t)
	state := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	createTestTopic(t, ctx, db, path, 1)
	callbackFailure := errors.New("callback failed after state write")
	err := db.Query().DoTx(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := tx.Exec(ctx, "INSERT INTO "+state+" (id) VALUES ('native');", query.WithCommit()); err != nil {
			return err
		}
		return callbackFailure
	})
	if !errors.Is(err, callbackFailure) {
		t.Fatalf("expected callback failure after native commit: %v", err)
	}
	p := livePublisher(t, ctx, db, path, DefaultConfig(map[string]string{"events": path}), 8*time.Second)
	if err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := tx.Exec(ctx, "INSERT INTO "+state+" (id) VALUES ('yql');", nil); err != nil {
			return err
		}
		if err := tx.Exec(ctx, "COMMIT;", nil); err == nil {
			t.Error("Query Service accepted transaction control in YQL")
		}
		return callbackFailure
	}); !errors.Is(err, callbackFailure) {
		t.Fatal(err)
	}
	row, err := db.Query().QueryRow(ctx, "SELECT id FROM "+state)
	if err != nil {
		t.Fatal(err)
	}
	var id string
	if err := row.Scan(&id); err != nil || id != "native" {
		t.Fatalf("expected only the native control state row, got %q: %v", id, err)
	}
	if total, _, err := topicEnds(ctx, db, path); err != nil || total != 0 {
		t.Fatalf("unexpected topic contents after premature commit: %d, %v", total, err)
	}
}

func TestIntegrationWriterInitializationFailure(t *testing.T) {
	var armed atomic.Bool
	var interrupted atomic.Int64
	intercept := func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn,
		method string, stream grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if armed.Load() && method == Ydb_Topic_V1.TopicService_StreamWrite_FullMethodName {
			interrupted.Add(1)
			return nil, status.Error(codes.Unavailable, "injected loss of transport before writer initialization")
		}
		return stream(ctx, desc, conn, method, opts...)
	}
	ctx, db, path := liveDatabase(t, ydb.With(config.WithGrpcOptions(grpc.WithChainStreamInterceptor(intercept))))
	createTestTopic(t, ctx, db, path, 1)
	state := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	p := livePublisher(t, ctx, db, path, DefaultConfig(map[string]string{"events": path}), 5*time.Second)
	e := testEvent(t)
	publish := func() error {
		return p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
			if err := tx.Exec(ctx, "INSERT INTO "+state+" (id) VALUES ('business');", nil); err != nil {
				return err
			}
			_ = tx.Publish("events", e)
			return nil
		})
	}
	armed.Store(true)
	err := publish()
	armed.Store(false)
	if err == nil || interrupted.Load() == 0 {
		t.Fatalf("expected failure during writer stream open: %v", err)
	}
	if _, err := db.Query().QueryRow(ctx, "SELECT id FROM "+state); !errors.Is(err, query.ErrNoRows) {
		t.Fatalf("writer initialization failure left state: %v", err)
	}
	if total, _, err := topicEnds(ctx, db, path); err != nil || total != 0 {
		t.Fatalf("writer initialization failure left an event: %d, %v", total, err)
	}
	verifyFence(t, ctx, p.owner, p, p.ownership.uid, 1, 0)
	if err := publish(); err != nil {
		t.Fatalf("publish after reconnect: %v", err)
	}
	verifyFence(t, ctx, p.owner, p, p.ownership.uid, 1, 1)
	assertStableProducerEvents(t, ctx, db, path, p.producer+"-0", []*eventv1.Event{e})
}
