package event

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Topic_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func writeWithProducer(ctx context.Context, db *ydb.Driver, tx query.TxActor, path, prefix string, events ...*eventv1.Event) error {
	messages, _, err := encodeBatch(ctx, DefaultConfig(nil), events)
	if err != nil {
		return err
	}
	options := []topicoptions.MultiWriterOption{topicoptions.WithWriterPartitionByKey(topicoptions.BoundPartitionChooser())}
	if prefix != "" {
		options = append(options, topicoptions.WithProducerIDPrefix(prefix))
	}
	writer, err := db.Topic().StartTransactionalWriter(tx, path,
		topicoptions.WithWriterMessageMaxBytesSize(DefaultConfig(nil).MaxEventBytes),
		topicoptions.WithWriteToManyPartitions(options...),
	)
	if err != nil {
		return err
	}
	return writer.Write(ctx, messages...)
}

func TestIntegrationStableProducerSequence(t *testing.T) {
	ctx, db, path := liveDatabase(t)
	var loseCommit, lostCommit atomic.Bool
	intercept := func(ctx context.Context, method string, req, reply any, conn *grpc.ClientConn,
		invoke grpc.UnaryInvoker, options ...grpc.CallOption) error {
		err := invoke(ctx, method, req, reply, conn, options...)
		if method == Ydb_Query_V1.QueryService_CommitTransaction_FullMethodName && err == nil {
			if response, ok := reply.(*Ydb_Query.CommitTransactionResponse); ok &&
				response.GetStatus() == Ydb.StatusIds_SUCCESS && loseCommit.CompareAndSwap(true, false) {
				lostCommit.Store(true)
				return status.Error(codes.DeadlineExceeded, "injected successful commit response loss")
			}
		}
		return err
	}
	_, replacement, _ := liveDatabase(t, ydb.With(config.WithGrpcOptions(grpc.WithChainUnaryInterceptor(intercept))))
	createTestTopic(t, ctx, db, path, 1)
	secondPath := path + "_second"
	createTestTopic(t, ctx, db, secondPath, 1)
	table := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	rollback := errors.New("rollback after staging")
	const prefix = "m8.test.events.0.0"
	var expected, expectedSecond []*eventv1.Event
	for attempt := range 5 {
		driver := db
		if attempt >= 3 {
			driver = replacement
		}
		first, second, forwarded := testEvent(t), testEvent(t), testEvent(t)
		first.Id, second.Id, forwarded.Id = uuid.NewString(), uuid.NewString(), uuid.NewString()
		owner, _ := NewTransactor(driver, 12*time.Second)
		if attempt == 3 {
			loseCommit.Store(true)
		}
		callbacks := 0
		err := owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
			callbacks++
			if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+table+" (id) VALUES ($id);",
				query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(first.Id).Build())); err != nil {
				return err
			}
			if err := writeWithProducer(ctx, driver, tx, path, prefix, first, second); err != nil {
				return err
			}
			if err := writeWithProducer(ctx, driver, tx, secondPath, prefix, forwarded); err != nil {
				return err
			}
			if attempt == 1 {
				return rollback
			}
			if attempt == 2 && callbacks == 1 {
				return retry.RetryableError(errors.New("injected retry after staging"))
			}
			return nil
		})
		if attempt == 1 {
			if !errors.Is(err, rollback) {
				t.Fatalf("rollback: %v", err)
			}
			if stats := owner.Stats(); stats.Failed != 1 || stats.Attempts != 1 || stats.Committed != 0 {
				t.Fatalf("rollback stats: %+v", stats)
			}
			continue
		}
		if attempt == 3 {
			if !lostCommit.Load() || err == nil {
				t.Fatalf("expected ambiguity from an intercepted successful response, got %v", err)
			}
			row, lookupErr := db.Query().QueryRow(ctx, "DECLARE $id AS Utf8; SELECT id FROM "+table+" WHERE id = $id;",
				query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(first.Id).Build()))
			if lookupErr != nil {
				t.Fatal(lookupErr)
			}
			var id string
			if scanErr := row.Scan(&id); scanErr != nil || id != first.Id {
				t.Fatalf("business state after loss of commit response: id=%s error=%v", id, scanErr)
			}
			expected = append(expected, first, second)
			expectedSecond = append(expectedSecond, forwarded)
			continue
		}
		if err != nil {
			t.Fatalf("transaction %d: %v", attempt, err)
		}
		stats := owner.Stats()
		if stats.Attempts != uint64(callbacks) || stats.Retries != uint64(callbacks-1) ||
			stats.Committed != 1 || stats.Failed != 0 || stats.Duration <= 0 || attempt == 2 && callbacks != 2 {
			t.Fatalf("transaction %d, callbacks=%d: stats=%+v", attempt, callbacks, stats)
		}
		expected = append(expected, first, second)
		expectedSecond = append(expectedSecond, forwarded)
	}

	row, err := replacement.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table)
	if err != nil {
		t.Fatal(err)
	}
	var state uint64
	if err := row.Scan(&state); err != nil || state != uint64(len(expectedSecond)) {
		t.Fatalf("business rows: %d, error: %v", state, err)
	}
	assertStableProducerEvents(t, ctx, replacement, path, prefix+"-0", expected)
	assertStableProducerEvents(t, ctx, replacement, secondPath, prefix+"-0", expectedSecond)
}

func assertStableProducerEvents(t *testing.T, ctx context.Context, db *ydb.Driver, path, producer string, expected []*eventv1.Event) {
	t.Helper()
	end, _, err := topicEnds(ctx, db, path)
	if err != nil || end != int64(len(expected)) {
		t.Fatalf("topic %s end: %d, want %d, error: %v", path, end, len(expected), err)
	}
	reader, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(path))
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
	read, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var lastSeq int64
	for offset, want := range expected {
		msg, err := reader.ReadMessage(read)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(msg)
		if err != nil {
			t.Fatal(err)
		}
		got := new(eventv1.Event)
		if err := proto.Unmarshal(data, got); err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(got, want) || msg.Offset != int64(offset) || msg.ProducerID != producer || msg.SeqNo <= lastSeq {
			t.Fatalf("unexpected publication: offset=%d ProducerID=%q SeqNo=%d event=%s", msg.Offset, msg.ProducerID, msg.SeqNo, got.Id)
		}
		lastSeq = msg.SeqNo
	}
	t.Logf("topic %s: events=%d producer IDs=1 last sequence=%d", path, end, lastSeq)
}

type heldProducerStream struct {
	grpc.ClientStream
	started chan<- struct{}
	release <-chan struct{}
	held    *atomic.Bool
}

func (s heldProducerStream) SendMsg(message any) error {
	if request, ok := message.(*Ydb_Topic.StreamWriteMessage_FromClient); ok &&
		request.GetWriteRequest() != nil && s.held.CompareAndSwap(false, true) {
		close(s.started)
		select {
		case <-s.release:
		case <-s.Context().Done():
			return s.Context().Err()
		}
	}
	return s.ClientStream.SendMsg(message)
}

func TestIntegrationDuplicateProducerIdentityCanSuppressEvent(t *testing.T) {
	firstSequenced := make(chan struct{})
	release := make(chan struct{}, 1)
	defer close(release)
	var held atomic.Bool
	intercept := func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn,
		method string, streamer grpc.Streamer, options ...grpc.CallOption) (grpc.ClientStream, error) {
		stream, err := streamer(ctx, desc, conn, method, options...)
		if err == nil && method == Ydb_Topic_V1.TopicService_StreamWrite_FullMethodName {
			return heldProducerStream{stream, firstSequenced, release, &held}, nil
		}
		return stream, err
	}
	ctx, db, path := liveDatabase(t, ydb.With(config.WithGrpcOptions(grpc.WithChainStreamInterceptor(intercept))))
	_, other, _ := liveDatabase(t)
	createTestTopic(t, ctx, db, path, 1)
	table := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	const identity = "m8.test.overlapping.0.0"
	run := func(driver *ydb.Driver, e *eventv1.Event, timeout time.Duration) error {
		owner, _ := NewTransactor(driver, timeout)
		return owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
			if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+table+" (id) VALUES ($id);",
				query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(e.Id).Build())); err != nil {
				return err
			}
			return writeWithProducer(ctx, driver, tx, path, identity, e)
		})
	}
	first, second := testEvent(t), testEvent(t)
	first.Id, second.Id = uuid.NewString(), uuid.NewString()
	result := make(chan error, 1)
	go func() { result <- run(db, first, 12*time.Second) }()
	select {
	case <-firstSequenced:
	case err := <-result:
		t.Fatalf("transaction did not assign a sequence number: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	secondErr := run(other, second, 5*time.Second)
	// Resume the first writer after the competing replica attempts to commit.
	release <- struct{}{}
	firstErr := <-result
	for _, e := range []*eventv1.Event{first, second} {
		row, err := other.Query().QueryRow(ctx, "DECLARE $id AS Utf8; SELECT id FROM "+table+" WHERE id = $id;",
			query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(e.Id).Build()))
		if err != nil {
			t.Fatal(err)
		}
		var id string
		if err := row.Scan(&id); err != nil || id != e.Id {
			t.Fatalf("missing business state for %s: %s, %v", e.Id, id, err)
		}
	}
	end, _, err := topicEnds(ctx, other, path)
	t.Logf("first=%v second=%v business rows=2 topic end=%d", firstErr, secondErr, end)
	// Pin the known behavior as a release blocker on this SDK/server pair.
	// Reassess the ownership decision if a future version changes the outcome.
	if firstErr != nil || secondErr != nil || err != nil || end != 1 {
		t.Fatalf("expected two successful commits with one suppressed event: %v, %v, end=%d, describe=%v",
			firstErr, secondErr, end, err)
	}
	reader, err := other.Topic().StartReader(testConsumer, topicoptions.ReadTopic(path))
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
	read, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	msg, err := reader.ReadMessage(read)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(msg)
	if err != nil {
		t.Fatal(err)
	}
	got := new(eventv1.Event)
	if err := proto.Unmarshal(data, got); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(got, second) || msg.Offset != 0 || msg.SeqNo != 1 || msg.ProducerID != identity+"-0" {
		t.Fatalf("unexpected surviving event: offset=%d SeqNo=%d ProducerID=%q event=%s",
			msg.Offset, msg.SeqNo, msg.ProducerID, got.Id)
	}
}
