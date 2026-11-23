package policyjournal

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

type fakeProjection struct {
	calls int
	err   error
}

func (p *fakeProjection) Process(context.Context, projection.Event) error { p.calls++; return p.err }

type fakeCertification struct {
	calls int
	err   error
}

func (c *fakeCertification) CommitPolicyJob(context.Context, uint64, projection.Key) error {
	c.calls++
	return c.err
}
func testPolicyEvent(t *testing.T) Event {
	t.Helper()
	tenant, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	return Event{PolicyRevision: 3, Projection: projection.Event{
		EventID:  "mutation-42-binding-1",
		Key:      projection.Key{TenantID: tenant, Kind: projection.KindRoleBinding, AggregateID: "11111111-1111-4111-8111-111111111111"},
		Revision: 7,
	}}
}

type singleDelivery struct {
	event     Event
	delivered bool
	acks      int
}

func (d *singleDelivery) Next(context.Context) (Delivery, error) {
	if d.delivered {
		return nil, io.EOF
	}
	d.delivered = true
	return d, nil
}
func (d *singleDelivery) Event() Event              { return d.event }
func (d *singleDelivery) Ack(context.Context) error { d.acks++; return nil }

func TestPolicyWorkerDoesNotAckBeforeDurableCertification(t *testing.T) {
	p := &fakeProjection{err: errors.New("SpiceDB failed")}
	c := &fakeCertification{}
	w, _ := NewWorker(p, c)
	event := testPolicyEvent(t)
	run := func(wantErr error, commits, acks int) {
		t.Helper()
		d := &singleDelivery{event: event}
		err := w.Run(t.Context(), d)
		if !errors.Is(err, wantErr) || c.calls != commits || d.acks != acks {
			t.Fatalf("Run returned %v, commits=%d, acknowledgements=%d; want %v, %d, %d", err, c.calls, d.acks, wantErr, commits, acks)
		}
	}
	run(p.err, 0, 0)
	p.err = nil
	c.err = errors.New("YDB checkpoint not durable")
	run(c.err, 1, 0)
	c.err = nil
	run(io.EOF, 2, 1)
}
func TestPolicyWorkerRequiresManifestRevision(t *testing.T) {
	p := &fakeProjection{}
	c := &fakeCertification{}
	w, _ := NewWorker(p, c)
	event := testPolicyEvent(t)
	event.PolicyRevision = 0
	if err := w.Process(context.Background(), event); !errors.Is(err, ErrInvalidPolicyDelivery) || p.calls != 0 {
		t.Fatal("unmanifested event accepted")
	}
	event = testPolicyEvent(t)
	event.Projection.Key.Kind = projection.KindRMWorkspace
	if err := w.Process(context.Background(), event); !errors.Is(err, ErrInvalidPolicyDelivery) {
		t.Fatal("resource event accepted as Access grant")
	}
}
