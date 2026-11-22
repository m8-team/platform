package projection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain"
)

type changingSource struct {
	snapshots map[Key]Snapshot
	err       error
}

func (s *changingSource) LoadCurrent(ctx context.Context, k Key) (Snapshot, error) {
	if s.err != nil {
		return Snapshot{}, s.err
	}
	return s.snapshots[k], nil
}

type fakeStore struct {
	cp      Checkpoint
	commits int
	fail    bool
}

func (s *fakeStore) WithLocked(ctx context.Context, key Key, fn func(Checkpoint) (Checkpoint, error)) error {
	if s.cp.Key == (Key{}) {
		s.cp.Key = key
	}
	next, err := fn(s.cp)
	if err != nil {
		return err
	}
	if s.fail {
		return errors.New("checkpoint write failure")
	}
	s.cp = next
	s.commits++
	return nil
}
func testKey(t *testing.T, id string) Key {
	return Key{TenantID: tenant(t, tenantA), Kind: KindGrant, AggregateID: id}
}
func TestWorkerCrashReplayAndIndependentGrantRevoke(t *testing.T) {
	ctx := context.Background()
	writer := &memoryWriter{}
	proj := projector(t, writer)
	id1 := "11111111-1111-4111-8111-111111111111"
	id2 := "22222222-2222-4222-8222-222222222222"
	k1, k2 := testKey(t, id1), testKey(t, id2)
	g1 := active(t, approved(t, makeGrant(t, id1, k1.TenantID)))
	g2 := active(t, approved(t, makeGrant(t, id2, k2.TenantID)))
	source := &changingSource{snapshots: map[Key]Snapshot{k1: {Key: k1, Revision: 1, Grant: &g1}, k2: {Key: k2, Revision: 1, Grant: &g2}}}
	cp1, cp2 := &fakeStore{}, &fakeStore{}
	// A simulated failure AFTER the graph write but BEFORE durable checkpoint.
	cp1.fail = true
	worker, _ := NewWorker(proj, source, cp1, func() time.Time { return fixedNow })
	e1 := Event{"event-1", k1, 1}
	if err := worker.Process(ctx, e1); err == nil {
		t.Fatal("expected checkpoint failure")
	}
	if len(writer.edges) != 2 || cp1.cp.Revision != 0 {
		t.Fatal("graph write/checkpoint failure window not exercised")
	}
	cp1.fail = false
	if err := worker.Process(ctx, e1); err != nil {
		t.Fatal(err)
	}
	if len(writer.edges) != 2 || cp1.cp.Revision != 1 || cp1.cp.ZedToken == "" {
		t.Fatalf("replayed grant: %+v, edges=%d", cp1.cp, len(writer.edges))
	}
	calls := writer.calls
	if err := worker.Process(ctx, e1); err != nil {
		t.Fatal(err)
	}
	if writer.calls != calls {
		t.Fatal("duplicate changed graph")
	}
	worker2, _ := NewWorker(proj, source, cp2, func() time.Time { return fixedNow })
	if err := worker2.Process(ctx, Event{"event-2", k2, 1}); err != nil {
		t.Fatal(err)
	}
	if len(writer.edges) != 4 {
		t.Fatal("grants collapsed into shared edges")
	}
	revoked, err := g1.Revoke(g1.Revision(), subject(t, "user", "security-admin"), fixedNow.Add(-20*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	source.snapshots[k1] = Snapshot{Key: k1, Revision: 2, Grant: &revoked}
	if err := worker.Process(ctx, Event{"revoke-1", k1, 2}); err != nil {
		t.Fatal(err)
	}
	if len(writer.edges) != 2 {
		t.Fatalf("revocation must leave other grant, got %d", len(writer.edges))
	}
	for _, edge := range writer.edges {
		if edge.Resource().ID() != id2 && edge.Subject().ID() != id2 {
			t.Fatalf("foreign graph edge remains: %+v", edge.Key())
		}
	}
}
func TestWorkerGapRequiresReconcile(t *testing.T) {
	ctx := context.Background()
	writer := &memoryWriter{}
	p := projector(t, writer)
	key := testKey(t, "11111111-1111-4111-8111-111111111111")
	g := active(t, approved(t, makeGrant(t, key.AggregateID, key.TenantID)))
	src := &changingSource{snapshots: map[Key]Snapshot{key: {Key: key, Revision: 3, Grant: &g}}}
	store := &fakeStore{}
	w, _ := NewWorker(p, src, store, func() time.Time { return fixedNow })
	if err := w.Process(ctx, Event{"late-3", key, 3}); !errors.Is(err, ErrRevisionGap) {
		t.Fatalf("want gap, got %v", err)
	}
	if writer.calls != 0 {
		t.Fatal("gap must not mutate graph")
	}
	if err := w.ReconcileKey(ctx, key); err != nil {
		t.Fatal(err)
	}
	if store.cp.Revision != 3 || store.cp.ZedToken == "" {
		t.Fatalf("reconciliation checkpoint: %+v", store.cp)
	}
	if err := w.Process(ctx, Event{"old-2", key, 2}); err != nil {
		t.Fatal(err)
	}
	if writer.calls != 1 {
		t.Fatal("stale event rewrote graph")
	}
	if err := w.ReconcileKey(ctx, key); err != nil {
		t.Fatal(err)
	}
	if writer.calls != 2 {
		t.Fatal("reconciliation must retouch desired graph edges")
	}
}
func TestWorkerRejectsCrossTenantSnapshotAndDoesNotWrite(t *testing.T) {
	writer := &memoryWriter{}
	p := projector(t, writer)
	key := testKey(t, "11111111-1111-4111-8111-111111111111")
	grant := makeGrant(t, key.AggregateID, tenant(t, tenantB))
	src := &changingSource{snapshots: map[Key]Snapshot{key: {Key: key, Revision: 1, Grant: &grant}}}
	store := &fakeStore{}
	w, _ := NewWorker(p, src, store, func() time.Time { return fixedNow })
	if err := w.Process(context.Background(), Event{"event", key, 1}); err == nil {
		t.Fatal("cross tenant snapshot accepted")
	}
	if writer.calls > 0 || store.cp.Revision != 0 {
		t.Fatal("unauthorized graph write")
	}
}
func TestWorkerDoesNotAdvanceOnSpiceDBError(t *testing.T) {
	writer := &memoryWriter{failure: errors.New("SpiceDB unavailable")}
	p := projector(t, writer)
	key := testKey(t, "11111111-1111-4111-8111-111111111111")
	grant := active(t, approved(t, makeGrant(t, key.AggregateID, key.TenantID)))
	src := &changingSource{snapshots: map[Key]Snapshot{key: {Key: key, Revision: 1, Grant: &grant}}}
	store := &fakeStore{}
	w, _ := NewWorker(p, src, store, func() time.Time { return fixedNow })
	if err := w.Process(context.Background(), Event{"event", key, 1}); err == nil {
		t.Fatal("wanted failed SpiceDB write")
	}
	if store.cp.Revision != 0 {
		t.Fatal("advanced checkpoint without graph success")
	}
}
func TestRecordedEdgesPreserveExpiryAndCaveat(t *testing.T) {
	tn := tenant(t, tenantA)
	r, err := domain.NewRelationship(domain.RelationshipInput{TenantID: tn, Resource: resource(t, "access_grant", "id"), Relation: "holder", Subject: subject(t, "user", "alice"), ExpiresAt: fixedNow.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	cp := Checkpoint{Key: testKey(t, "11111111-1111-4111-8111-111111111111"), Revision: 1, EventID: "event", Edges: recordEdges([]domain.Relationship{r})}
	edges, err := cp.relationships()
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 1 || !edges[0].ExpiresAt().Equal(r.ExpiresAt()) {
		t.Fatal("expiry lost during checkpoint")
	}
}

type fakeDelivery struct {
	event        Event
	acknowledged int
}

func (d *fakeDelivery) Event() Event              { return d.event }
func (d *fakeDelivery) Ack(context.Context) error { d.acknowledged++; return nil }

type singleFeed struct {
	delivery Delivery
	returned bool
}

func (f *singleFeed) Next(context.Context) (Delivery, error) {
	if f.returned {
		return nil, errors.New("end of feed")
	}
	f.returned = true
	return f.delivery, nil
}
func TestWorkerRunDoesNotAcknowledgeUntilDurableCheckpoint(t *testing.T) {
	writer := &memoryWriter{}
	p := projector(t, writer)
	k := testKey(t, "11111111-1111-4111-8111-111111111111")
	g := active(t, approved(t, makeGrant(t, k.AggregateID, k.TenantID)))
	src := &changingSource{snapshots: map[Key]Snapshot{k: {Key: k, Revision: 1, Grant: &g}}}
	store := &fakeStore{fail: true}
	w, _ := NewWorker(p, src, store, func() time.Time { return fixedNow })
	d := &fakeDelivery{event: Event{"event", k, 1}}
	if err := w.Run(context.Background(), &singleFeed{delivery: d}); err == nil {
		t.Fatal("expected error")
	}
	if d.acknowledged != 0 {
		t.Fatal("ack before durable checkpoint")
	}
	store.fail = false
	if err := w.Run(context.Background(), &singleFeed{delivery: d}); err == nil {
		t.Fatal("expected end of feed")
	}
	if d.acknowledged != 1 || store.cp.Revision != 1 {
		t.Fatal("not acknowledged after commit")
	}
}

func TestWorkerRejectsSnapshotAheadUntilReconciliation(t *testing.T) {
	writer := &memoryWriter{}
	p := projector(t, writer)
	k := testKey(t, "11111111-1111-4111-8111-111111111111")
	grant := active(t, approved(t, makeGrant(t, k.AggregateID, k.TenantID)))
	source := &changingSource{snapshots: map[Key]Snapshot{k: {Key: k, Revision: 2, Grant: &grant}}}
	store := &fakeStore{}
	w, _ := NewWorker(p, source, store, func() time.Time { return fixedNow })
	if err := w.Process(context.Background(), Event{EventID: "rev-1", Key: k, Revision: 1}); !errors.Is(err, ErrSourceAhead) {
		t.Fatalf("ahead should not silently skip rev: %v", err)
	}
	if writer.calls != 0 || store.cp.Revision != 0 {
		t.Fatal("ahead source silently advanced projection")
	}
	if err := w.ReconcileKey(context.Background(), k); err != nil {
		t.Fatal(err)
	}
	if store.cp.Revision != 2 || store.cp.ZedToken == "" {
		t.Fatal("reconciliation failed to reach source head")
	}
}

func TestProjectionKeyRejectsMalformedUUID(t *testing.T) {
	k := testKey(t, "11111111-1111-4111-8111-111111111111")
	k.AggregateID = "11111111-1111-4111-8111-zzzzzzzzzzzz"
	if err := k.Validate(); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("invalid ID: %v", err)
	}
}
