package event

import (
	"context"
	"errors"
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
	if err := db.Topic().Create(ctx, path,
		topicoptions.CreateWithMinActivePartitions(partitions),
		topicoptions.CreateWithMaxActivePartitions(partitions),
		topicoptions.CreateWithAutoPartitioningSettings(topictypes.AutoPartitioningSettings{
			AutoPartitioningStrategy: topictypes.AutoPartitioningStrategyScaleUp,
		}),
		topicoptions.CreateWithConsumer(topictypes.Consumer{Name: testConsumer}),
	); err != nil {
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
	p, err := NewPublisher(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTransactor(db, 12*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	writeState := func(ctx context.Context, tx query.TxActor, id string) error {
		return tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+table+" (id) VALUES ($id);",
			query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(id).Build()))
	}
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
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := writeState(ctx, tx, rolledBack.Id); err != nil {
			return err
		}
		if err := p.Publish(ctx, tx, "resource", rolledBack); err != nil {
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
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := writeState(ctx, tx, bad.Id); err != nil {
			return err
		}
		return p.Publish(ctx, tx, "resource", bad)
	})
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected validation failure, got %v", err)
	}
	assertCount(0)

	missing := makeEvent()
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := writeState(ctx, tx, missing.Id); err != nil {
			return err
		}
		return p.Publish(ctx, tx, "missing", missing)
	})
	if err == nil {
		t.Fatal("unexpected success for nonexistent topic")
	}
	assertCount(0)

	first, second := makeEvent(), makeEvent()
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := writeState(ctx, tx, first.Id); err != nil {
			return err
		}
		return p.Publish(ctx, tx, "resource", first, second)
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
	consume := func(ctx context.Context, tx query.TxActor, fail bool) error {
		batch, err := reader.PopMessagesBatchTx(ctx, tx)
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
		if err := p.Publish(ctx, tx, "forwarded", forward...); err != nil {
			return err
		}
		if fail {
			return rollback
		}
		return nil
	}
	if err := tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		return consume(ctx, tx, true)
	}); !errors.Is(err, rollback) {
		t.Fatalf("consumer rollback: %v", err)
	}
	assertCount(1)
	restartReader()
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
		if err := tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
			return consume(ctx, tx, false)
		}); err != nil {
			t.Fatal(err)
		}
	}
	assertCount(3)
	total, _, err := topicEnds(ctx, db, topicPath+"_forwarded")
	if err != nil || total != 2 {
		t.Fatalf("consumer publication and state must commit with offsets: %d, %v", total, err)
	}
	restartReader()
	sentinel := makeEvent()
	if err := tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		return p.Publish(ctx, tx, "resource", sentinel)
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
	p, err := NewPublisher(db, DefaultConfig(map[string]string{"events": path}))
	if err != nil {
		t.Fatal(err)
	}
	tr, _ := NewTransactor(db, 3*time.Second)
	event := testEvent(t)
	staged := false
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := tx.Exec(ctx, "INSERT INTO "+table+" (id) VALUES ('deadline');"); err != nil {
			return err
		}
		if err := p.Publish(ctx, tx, "events", event); err != nil {
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
	p, err := NewPublisher(db, DefaultConfig(map[string]string{"events": path, "other": path + "_other"}))
	if err != nil {
		t.Fatal(err)
	}
	tr, _ := NewTransactor(db, 8*time.Second)
	callbackReturned := false
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := tx.Exec(ctx, "INSERT INTO "+table+" (id) VALUES ('writer-failure');"); err != nil {
			return err
		}
		if err := p.Publish(ctx, tx, "events", testEvent(t)); err != nil {
			return err
		}
		if err := p.Publish(ctx, tx, "other", testEvent(t)); err != nil {
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
