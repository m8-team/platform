package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/platform/idempotency"
	"github.com/m8-team/platform/internal/platform/operation"
	"github.com/m8-team/platform/services/m8-authentication/internal/domain"
)

type clients struct{ client domain.Client }

func (c clients) GetClient(context.Context, string) (domain.Client, bool, error) {
	return c.client, true, nil
}

type identity struct{}

func (identity) ResolveSubject(context.Context, string) (string, bool, error) {
	return "users/u-1", true, nil
}

type risk struct{ allowed bool }

func (r risk) Evaluate(context.Context, string, string, string) (domain.RiskDecision, error) {
	return domain.RiskDecision{Allowed: r.allowed, RequiredAAL: "AAL2", RequiredChallenge: "CIBA"}, nil
}

type txRepo struct{ items map[string]domain.Transaction }

func (r *txRepo) Save(_ context.Context, tx domain.Transaction) error {
	r.items[tx.ID] = tx
	return nil
}
func (r *txRepo) Get(_ context.Context, id string) (domain.Transaction, bool, error) {
	v, ok := r.items[id]
	return v, ok, nil
}

type opRepo struct {
	items map[string]operation.Operation
}

func (r *opRepo) Save(_ context.Context, op operation.Operation) error {
	r.items[op.ID] = op
	return nil
}
func (r *opRepo) Get(_ context.Context, id string) (operation.Operation, bool, error) {
	v, ok := r.items[id]
	return v, ok, nil
}

type outbox struct{ events []domain.Event }

func (o *outbox) Add(_ context.Context, event domain.Event) error {
	o.events = append(o.events, event)
	return nil
}

type uow struct{}

func (uow) Do(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type ids struct{ n int }

func (g *ids) NewID(prefix string) string { g.n++; return fmt.Sprintf("%s-%d", prefix, g.n) }

type clock struct{}

func (clock) Now() time.Time { return time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC) }

func newService(allowed bool) Service {
	return Service{
		Clients:      clients{client: domain.Client{ID: "clients/c-1", Enabled: true, CIBA: true}},
		Identity:     identity{},
		Risk:         risk{allowed: allowed},
		Transactions: &txRepo{items: map[string]domain.Transaction{}},
		Operations:   &opRepo{items: map[string]operation.Operation{}},
		Outbox:       &outbox{},
		Idempotency:  idempotency.NewMemoryStore(),
		UOW:          uow{},
		IDs:          &ids{},
		Clock:        clock{},
		TTL:          5 * time.Minute,
	}
}

func TestStartReauthenticationCreatesAtomicArtifacts(t *testing.T) {
	svc := newService(true)
	result, err := svc.StartReauthentication(context.Background(), StartReauthenticationCommand{
		ClientID: "clients/c-1", SubjectHint: "user@example.test", RequestedAAL: "AAL2",
		Reason: "REFRESH_UNAVAILABLE", IdempotencyKey: "key-1", CorrelationID: "corr-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AuthenticationID == "" || result.OperationID == "" {
		t.Fatal("missing identifiers")
	}
}

func TestStartReauthenticationIsIdempotent(t *testing.T) {
	svc := newService(true)
	cmd := StartReauthenticationCommand{
		ClientID: "clients/c-1", SubjectHint: "user@example.test", RequestedAAL: "AAL2",
		Reason: "REFRESH_UNAVAILABLE", IdempotencyKey: "key-1", CorrelationID: "corr-1",
	}
	first, err := svc.StartReauthentication(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.StartReauthentication(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Reused {
		t.Fatal("expected reused result")
	}
	if first.OperationID != second.OperationID {
		t.Fatal("operation IDs differ")
	}
}

func TestStartReauthenticationStopsOnRiskDeny(t *testing.T) {
	svc := newService(false)
	_, err := svc.StartReauthentication(context.Background(), StartReauthenticationCommand{
		ClientID: "clients/c-1", SubjectHint: "user@example.test", RequestedAAL: "AAL2",
		Reason: "REFRESH_UNAVAILABLE", IdempotencyKey: "key-1",
	})
	if err != domain.ErrRiskDenied {
		t.Fatalf("unexpected error: %v", err)
	}
}
