package event

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIntegrationCommandReceipt(t *testing.T) {
	t.Run("concurrent", func(t *testing.T) { testCommandReceipt(t, false) })
	t.Run("discarded_commit_response", func(t *testing.T) { testCommandReceipt(t, true) })
}

func testCommandReceipt(t *testing.T, inject bool) {
	var failBeforeSend, beforeSendIntercepted, discardResponse, discarded atomic.Bool
	intercept := func(ctx context.Context, method string, req, reply any, conn *grpc.ClientConn,
		invoke grpc.UnaryInvoker, options ...grpc.CallOption) error {
		const commit = Ydb_Query_V1.QueryService_CommitTransaction_FullMethodName
		if method == commit && failBeforeSend.Load() {
			beforeSendIntercepted.Store(true)
			return status.Error(codes.Unavailable, "injected transport failure before commit request")
		}
		err := invoke(ctx, method, req, reply, conn, options...)
		if method == commit && err == nil && discardResponse.Load() {
			if response, ok := reply.(*Ydb_Query.CommitTransactionResponse); ok &&
				response.GetStatus() == Ydb.StatusIds_SUCCESS && discardResponse.CompareAndSwap(true, false) {
				discarded.Store(true)
				return status.Error(codes.DeadlineExceeded, "injected loss of successful commit response")
			}
		}
		return err
	}
	ctx, db, path := liveDatabase(t, ydb.With(config.WithGrpcOptions(grpc.WithChainUnaryInterceptor(intercept))))
	_, replica, _ := liveDatabase(t)
	createTestTopic(t, ctx, db, path, 1)
	state := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	receipts := createTestTable(t, ctx, db, path+"_receipts",
		"id Utf8 NOT NULL, hash Utf8 NOT NULL, event_id Utf8 NOT NULL, PRIMARY KEY (id)")
	cfg := DefaultConfig(map[string]string{"events": path})
	p := livePublisher(t, ctx, db, path, cfg, 20*time.Second)
	slot := p.ownership.slot
	slot.PodName = "publisher-1"
	p2 := bindTestPublisher(t, ctx, replica, slot, p.guard, cfg, 20*time.Second)
	e := testEvent(t)
	e.Id = uuid.NewString() // Stable even if the SDK invokes the callback again.
	wrongHash := errors.New("command ID already used for a different request")
	rollback := errors.New("intentional rollback before commit")
	run := func(publisher *Publisher, hash, alias string, beforeWrite func(context.Context) error, fail bool) error {
		return publisher.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
			row, err := tx.QueryRow(ctx, "SELECT hash, event_id FROM "+receipts+" WHERE id = 'command';", nil)
			if err == nil {
				var storedHash, storedEvent string
				if err := row.Scan(&storedHash, &storedEvent); err != nil {
					return err
				}
				if storedHash != hash {
					return wrongHash
				}
				if storedEvent != e.Id {
					return errors.New("receipt did not return the original event identity")
				}
				return nil
			}
			if !errors.Is(err, query.ErrNoRows) {
				return err
			}
			if beforeWrite != nil {
				if err := beforeWrite(ctx); err != nil {
					return err
				}
			}
			if err := tx.Exec(ctx, "INSERT INTO "+state+" (id) VALUES ('command');", nil); err != nil {
				return err
			}
			params := ydb.ParamsBuilder().Param("$hash").Text(hash).Param("$event").Text(e.Id).Build()
			if err := tx.Exec(ctx, "DECLARE $hash AS Utf8; DECLARE $event AS Utf8; INSERT INTO "+receipts+
				" (id, hash, event_id) VALUES ('command', $hash, $event);", params); err != nil {
				return err
			}
			if err := tx.Publish(alias, e); err != nil {
				return err
			}
			if fail {
				return rollback
			}
			return nil
		})
	}
	assertCommitted := func(want uint64) {
		t.Helper()
		for _, table := range []string{state, receipts} {
			row, err := replica.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table)
			if err != nil {
				t.Fatal(err)
			}
			var count uint64
			if err := row.Scan(&count); err != nil || count != want {
				t.Fatalf("durable rows: got %d, want %d, error %v", count, want, err)
			}
		}
		count, _, err := topicEnds(ctx, replica, path)
		if err != nil || count != int64(want) {
			t.Fatalf("topic end: got %d, want %d, error %v", count, want, err)
		}
	}
	if err := run(p, "hash", "invalid", nil, false); !errors.Is(err, ErrInvalidTopic) {
		t.Fatal(err)
	}
	assertCommitted(0)
	if err := run(p, "hash", "events", nil, true); !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	assertCommitted(0)

	if inject {
		failBeforeSend.Store(true)
		savedTimeout := p.owner.timeout
		p.owner.timeout = 5 * time.Second
		err := run(p, "hash", "events", nil, false)
		p.owner.timeout = savedTimeout
		failBeforeSend.Store(false)
		if err == nil || !beforeSendIntercepted.Load() {
			t.Fatalf("expected intercepted commit request failure: %v", err)
		}
		assertCommitted(0)

		discardResponse.Store(true)
		err = run(p, "hash", "events", nil, false)
		if !discarded.Load() {
			t.Fatalf("successful commit response was not intercepted: %v", err)
		}
		t.Logf("SDK returned success after discarding the first reply: %t", err == nil)
		assertCommitted(1)
		t.Log("gRPC interceptor dropped a successful YDB CommitTransaction response before the SDK observed it")
	} else {
		read := make(chan struct{}, 2)
		release := make(chan struct{})
		var contenders atomic.Int32
		beforeWrite := func(ctx context.Context) error {
			if contenders.Add(1) > 2 {
				return nil
			}
			read <- struct{}{}
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		results := make(chan error, 2)
		go func() { results <- run(p, "hash", "events", beforeWrite, false) }()
		go func() { results <- run(p2, "hash", "events", beforeWrite, false) }()
		ready := 0
		deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		for ready < 2 && deadline.Err() == nil {
			select {
			case <-read:
				ready++
			case <-deadline.Done():
			}
		}
		close(release)
		for range 2 {
			if err := <-results; err != nil {
				t.Error(err)
			}
		}
		if ready != 2 {
			t.Fatal("both replicas must read the missing receipt before either writes")
		}
		assertCommitted(1)

	}

	// Re-submit through a separate driver; the durable receipt supplies the original result.
	if err := run(p2, "hash", "events", nil, false); err != nil {
		t.Fatal(err)
	}
	if err := run(p2, "different-hash", "events", nil, false); !errors.Is(err, wrongHash) {
		t.Fatal(err)
	}
	assertCommitted(1)
}
