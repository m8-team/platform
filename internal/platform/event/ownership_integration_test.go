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
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Topic_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"google.golang.org/grpc"
)

// The fault boundary models an old process that can still reach YDB after
// losing ownership. A privileged fixture write stands in for the handoff.
func TestIntegrationStableProducerOwnershipHandoff(t *testing.T) {
	for _, scenario := range []struct {
		name             string
		holdBeforeWrite  bool
		replacementFirst bool
	}{
		{name: "late_write_after_replacement_commit", holdBeforeWrite: true, replacementFirst: true},
		{name: "late_commit_after_replacement_commit", replacementFirst: true},
		{name: "stale_commit_before_replacement_writer"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			testStableProducerOwnershipHandoff(t, scenario.holdBeforeWrite, scenario.replacementFirst)
		})
	}
}

func testStableProducerOwnershipHandoff(t *testing.T, holdBeforeWrite, replacementFirst bool) {
	delayed := make(chan struct{})
	resume := make(chan struct{}, 1)
	var held, aborted atomic.Bool
	stream := func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string,
		next grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		s, err := next(ctx, desc, conn, method, opts...)
		if err == nil && holdBeforeWrite && method == Ydb_Topic_V1.TopicService_StreamWrite_FullMethodName {
			return heldProducerStream{s, delayed, resume, &held}, nil
		}
		return s, err
	}
	unary := func(ctx context.Context, method string, req, reply any, conn *grpc.ClientConn,
		next grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		isCommit := method == Ydb_Query_V1.QueryService_CommitTransaction_FullMethodName
		if isCommit && !holdBeforeWrite && held.CompareAndSwap(false, true) {
			close(delayed)
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
	ctx, old, path := liveDatabase(t, ydb.With(config.WithGrpcOptions(
		grpc.WithChainStreamInterceptor(stream), grpc.WithChainUnaryInterceptor(unary))))
	_, replacement, _ := liveDatabase(t)
	createTestTopic(t, ctx, replacement, path, 1)
	state := createTestTable(t, ctx, replacement, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	fence := createTestTable(t, ctx, replacement, path+"_owner", "replica Utf8 NOT NULL, generation Uint64 NOT NULL, PRIMARY KEY (replica)")
	if err := replacement.Query().Exec(ctx, "INSERT INTO "+fence+" (replica, generation) VALUES ('replica-0'u, 1u);"); err != nil {
		t.Fatal(err)
	}
	const producer = "m8.test.handoff.0.0"
	stale := errors.New("producer ownership changed")
	publish := func(owner *Transactor, generation uint64, e *eventv1.Event) error {
		return owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
			row, err := tx.QueryRow(ctx, "SELECT generation FROM "+fence+" WHERE replica = 'replica-0'u;")
			if err != nil {
				return err
			}
			var current uint64
			if err := row.Scan(&current); err != nil {
				return err
			}
			if current != generation {
				return stale
			}
			if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+state+" (id) VALUES ($id);",
				query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(e.Id).Build())); err != nil {
				return err
			}
			return writeWithProducer(ctx, owner.db, tx, path, producer, e)
		})
	}
	oldOwner, err := NewTransactor(old, 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	newOwner, err := NewTransactor(replacement, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	first, next, sentinel := testEvent(t), testEvent(t), testEvent(t)
	first.Id, next.Id, sentinel.Id = uuid.NewString(), uuid.NewString(), uuid.NewString()
	result := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		result <- publish(oldOwner, 1, first)
	}()
	defer func() {
		close(resume)
		<-done
	}()
	select {
	case <-delayed:
	case err := <-result:
		t.Fatalf("old transaction completed before handoff: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := replacement.Query().Exec(ctx, "UPDATE "+fence+" SET generation = 2u WHERE replica = 'replica-0'u;"); err != nil {
		t.Fatal(err)
	}
	if replacementFirst {
		if err := publish(newOwner, 2, next); err != nil {
			t.Fatalf("replacement publication while old request is delayed: %v", err)
		}
	}
	resume <- struct{}{}
	if err := <-result; !errors.Is(err, stale) {
		t.Fatalf("expected retry to reject stale generation, got %v", err)
	}
	if !aborted.Load() || oldOwner.Stats().Retries != 1 {
		t.Fatalf("expected server ABORTED and one retry, got aborted=%v, stats=%+v", aborted.Load(), oldOwner.Stats())
	}
	if err := publish(oldOwner, 1, first); !errors.Is(err, stale) {
		t.Fatalf("expected a fresh transaction with the old generation to fail, got %v", err)
	}
	if !replacementFirst {
		if err := publish(newOwner, 2, next); err != nil {
			t.Fatalf("replacement publication after stale abort: %v", err)
		}
	}
	if err := publish(newOwner, 2, sentinel); err != nil {
		t.Fatalf("publication after stale attempt: %v", err)
	}
	for _, e := range []*eventv1.Event{next, sentinel} {
		row, err := replacement.Query().QueryRow(ctx,
			"DECLARE $id AS Utf8; SELECT id FROM "+state+" WHERE id = $id;",
			query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(e.Id).Build()))
		if err != nil {
			t.Fatal(err)
		}
		var id string
		if err := row.Scan(&id); err != nil || id != e.Id {
			t.Fatalf("missing business state for %s: %s, %v", e.Id, id, err)
		}
	}
	row, err := replacement.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+state)
	if err != nil {
		t.Fatal(err)
	}
	var rows uint64
	if err := row.Scan(&rows); err != nil || rows != 2 {
		t.Fatalf("business rows: %d, error: %v", rows, err)
	}
	assertStableProducerEvents(t, ctx, replacement, path, producer+"-0", []*eventv1.Event{next, sentinel})
}
