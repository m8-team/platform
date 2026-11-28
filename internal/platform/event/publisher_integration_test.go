package event

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"google.golang.org/grpc"
)

func TestIntegrationPublisherConflictingCommitSamePod(t *testing.T) {
	held, resume := make(chan struct{}), make(chan struct{}, 1)
	var pause, aborted atomic.Bool
	intercept := func(ctx context.Context, method string, req, reply any, conn *grpc.ClientConn,
		next grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		isCommit := method == Ydb_Query_V1.QueryService_CommitTransaction_FullMethodName
		if isCommit && pause.CompareAndSwap(true, false) {
			close(held)
			select {
			case <-resume:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		err := next(ctx, method, req, reply, conn, opts...)
		if response, ok := reply.(*Ydb_Query.CommitTransactionResponse); isCommit && ok &&
			response.GetStatus() == Ydb.StatusIds_ABORTED {
			aborted.Store(true)
		}
		return err
	}
	ctx, old, path := liveDatabase(t, ydb.With(config.WithGrpcOptions(grpc.WithChainUnaryInterceptor(intercept))))
	_, newProcess, _ := liveDatabase(t)
	createTestTopic(t, ctx, newProcess, path, 1)
	state := createTestTable(t, ctx, newProcess, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	createTestTable(t, ctx, newProcess, path+"_owner", "replica Utf8 NOT NULL, pod_uid Utf8 NOT NULL, generation Uint64 NOT NULL, PRIMARY KEY (replica)")
	createTestTable(t, ctx, newProcess, path+"_guard", "replica Utf8 NOT NULL, revision Uint64 NOT NULL, PRIMARY KEY (replica)")
	slot := ReplicaSlot{"test", uuid.NewString(), "publisher-0", path + "_owner"}
	uid := uuid.NewString()
	if err := newProcess.Query().Exec(ctx, declareReplicaKey+"DECLARE $uid AS Utf8; "+
		"INSERT INTO `"+slot.TablePath+"` (replica, pod_uid, generation) VALUES ($replica, $uid, 1u);",
		query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(slot.key()).Param("$uid").Text(uid).Build())); err != nil {
		t.Fatal(err)
	}
	newPublisher := func(db *ydb.Driver) *Publisher {
		t.Helper()
		owner, err := NewTransactor(db, 15*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		claim, err := slot.Bind(ctx, owner, uid)
		if err != nil {
			t.Fatal(err)
		}
		p, err := NewPublisher(owner, DefaultConfig(map[string]string{"event": path}), claim, path+"_guard")
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	first, second := newPublisher(old), newPublisher(newProcess)
	write := func(ctx context.Context, tx query.TxActor, id string) error {
		return tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+state+" (id) VALUES ($id);",
			query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(id).Build()))
	}
	run := func(p *Publisher, event *eventv1.Event) error {
		return p.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor, send func(string, ...*eventv1.Event) error) error {
			if err := write(ctx, tx, event.Id); err != nil {
				return err
			}
			return send("event", event)
		})
	}
	late, winner := testEvent(t), testEvent(t)
	late.Id, winner.Id = uuid.NewString(), uuid.NewString()
	result, done := make(chan error, 1), make(chan struct{})
	pause.Store(true)
	go func() {
		defer close(done)
		result <- run(first, late)
	}()
	defer func() {
		close(resume)
		<-done
	}()
	select {
	case <-held:
	case err := <-result:
		t.Fatalf("old attempt did not reach commit: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := run(second, winner); err != nil {
		t.Fatalf("new publisher could not commit while the previous outcome was unresolved: %v", err)
	}
	resume <- struct{}{}
	if err := <-result; err != nil || !aborted.Load() || first.owner.Stats().Retries == 0 {
		t.Fatalf("expected server ABORTED followed by successful retry: aborted=%t, error=%v", aborted.Load(), err)
	}
	rejected := testEvent(t)
	rejected.Id = uuid.NewString()
	if err := second.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor, send func(string, ...*eventv1.Event) error) error {
		if err := write(ctx, tx, rejected.Id); err != nil {
			return err
		}
		if err := send("event", rejected); err != nil {
			return err
		}
		_ = send("event", rejected)
		return nil
	}); !errors.Is(err, ErrInvalidTopic) {
		t.Fatalf("repeated batch should abort even if the callback ignores the error: %v", err)
	}
	row, err := newProcess.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+state)
	if err != nil {
		t.Fatal(err)
	}
	var count uint64
	if err := row.Scan(&count); err != nil || count != 2 {
		t.Fatalf("committed state rows=%d, error=%v", count, err)
	}
	for _, event := range []*eventv1.Event{winner, late} {
		row, err := newProcess.Query().QueryRow(ctx, "DECLARE $id AS Utf8; SELECT id FROM "+state+" WHERE id = $id;",
			query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(event.Id).Build()))
		if err != nil {
			t.Fatal(err)
		}
		var id string
		if err := row.Scan(&id); err != nil || id != event.Id {
			t.Fatalf("missing committed state for event %s: %v", event.Id, err)
		}
	}
	verifyFence(t, ctx, second.owner, second, uid, 1, 2)
	assertStableProducerEvents(t, ctx, newProcess, path, second.producer+"-0", []*eventv1.Event{winner, late})
}

func TestIntegrationPublisherCommitDeadline(t *testing.T) {
	var pause, intercepted atomic.Bool
	intercept := func(ctx context.Context, method string, req, reply any, conn *grpc.ClientConn,
		next grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if method == Ydb_Query_V1.QueryService_CommitTransaction_FullMethodName && pause.CompareAndSwap(true, false) {
			intercepted.Store(true)
			<-ctx.Done()
			return ctx.Err()
		}
		return next(ctx, method, req, reply, conn, opts...)
	}
	ctx, db, path := liveDatabase(t, ydb.With(config.WithGrpcOptions(grpc.WithChainUnaryInterceptor(intercept))))
	createTestTopic(t, ctx, db, path, 1)
	state := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	p := livePublisher(t, ctx, db, path, DefaultConfig(map[string]string{"events": path}), 5*time.Second)
	publish := func(e *eventv1.Event) error {
		e.Id = uuid.NewString()
		return p.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor, send func(string, ...*eventv1.Event) error) error {
			if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+state+" (id) VALUES ($id);",
				query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(e.Id).Build())); err != nil {
				return err
			}
			return send("events", e)
		})
	}
	pause.Store(true)
	if err := publish(testEvent(t)); !errors.Is(err, context.DeadlineExceeded) || !intercepted.Load() {
		t.Fatalf("expected a deadline at the CommitTransaction boundary: %v", err)
	}
	next := testEvent(t)
	if err := publish(next); err != nil {
		t.Fatalf("subsequent publication could not acquire lane and commit: %v", err)
	}
	assertStableProducerEvents(t, ctx, db, path, p.producer+"-0", []*eventv1.Event{next})
	verifyFence(t, ctx, p.owner, p, p.ownership.uid, 1, 1)
	row, err := db.Query().QueryRow(ctx, "SELECT id FROM "+state)
	if err != nil {
		t.Fatal(err)
	}
	var id string
	if err := row.Scan(&id); err != nil || id != next.Id {
		t.Fatalf("unexpected state after the expired commit request: %q, %v", id, err)
	}
}
