package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/m8-team/platform-sdk/fault"
	"github.com/m8-team/platform-sdk/inbox"
	"github.com/m8-team/platform-sdk/outbox"
	"github.com/m8-team/platform-sdk/testkit"
	"github.com/m8-team/platform-sdk/transaction"
	"github.com/m8-team/product-service/internal/product/domain"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace/noop"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

// This test requires an isolated database with migrations/001_init.sql applied.
// Never point TEST_DATABASE_URL at a production database.
func TestPostgresAtomicDelivery(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL absent: real PostgreSQL integration not run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	tenant := domain.TenantID("test_" + rand.Text())
	hash := sha256.Sum256([]byte("Coffee"))
	key := rand.Text()
	repo := New(db, noop.NewTracerProvider().Tracer("test"), propagation.TraceContext{})
	const count = 8
	results := make(chan domain.Product, count)
	errs := make(chan error, count)
	var wg sync.WaitGroup
	for range count {
		wg.Go(func() {
			p, e := repo.Create(ctx, domain.Product{ID: domain.ID("prd_" + rand.Text()), Tenant: tenant, Name: "Coffee"}, key, hash)
			results <- p
			errs <- e
		})
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var id domain.ID
	for p := range results {
		if id != "" && p.ID != id {
			t.Fatal("duplicate request produced different products")
		}
		id = p.ID
	}
	var rows int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM product.products WHERE tenant=$1`, tenant).Scan(&rows); err != nil || rows != 1 {
		t.Fatalf("rows=%d err=%v", rows, err)
	}
	_, err = repo.Create(ctx, domain.Product{ID: "different", Tenant: tenant, Name: "Tea"}, key, sha256.Sum256([]byte("Tea")))
	if !errors.Is(err, &fault.Error{Code: fault.Conflict}) {
		t.Fatalf("expected conflict: %v", err)
	}
	// Inbox marker and side effect roll back together, then a replay executes once.
	consumer := "test_" + rand.Text()
	message := rand.Text()
	sentinel := errors.New("rollback")
	err = transaction.Run(ctx, db, func(tx *sql.Tx) error {
		_, e := inbox.Execute(ctx, tx, consumer, message, func(context.Context) error { return sentinel })
		return e
	})
	if !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	calls := 0
	for range 2 {
		err = transaction.Run(ctx, db, func(tx *sql.Tx) error {
			_, e := inbox.Execute(ctx, tx, consumer, message, func(context.Context) error { calls++; return nil })
			return e
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatalf("inbox calls=%d", calls)
	}
	// Acknowledged dispatch marks the event. Repeated Step cannot publish it again.
	recorder := &testkit.Recorder{}
	worker, err := outbox.New(db, recorder, slog.New(slog.NewTextHandler(io.Discard, nil)), outbox.Config{PollInterval: time.Millisecond, PublishTimeout: time.Second, Lease: 2 * time.Second, MaxAttempts: 3})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for range 100 {
		ok, e := worker.Step(ctx)
		if e != nil {
			t.Fatal(e)
		}
		for _, event := range recorder.Events() {
			if event.AggregateID == string(id) {
				found = true
			}
		}
		if found || !ok {
			break
		}
	}
	if !found {
		t.Fatal("created event not dispatched")
	}
	var published bool
	if err = db.QueryRowContext(ctx, `SELECT published_at IS NOT NULL FROM platform.outbox WHERE id=$1`, string(id)+":created").Scan(&published); err != nil || !published {
		t.Fatalf("published=%v err=%v", published, err)
	}
}
