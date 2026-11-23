package projection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

type fencedMemoryStore struct {
	cp         Checkpoint
	generation uint64
	failCommit bool
}

func (s *fencedMemoryStore) WithLocked(context.Context, Key, func(Checkpoint) (Checkpoint, error)) error {
	return errors.New("unfenced path must not be called")
}
func (s *fencedMemoryStore) WithFenced(ctx context.Context, key Key, fn func(context.Context, Checkpoint, ports.GraphFence) (Checkpoint, error)) error {
	if s.cp.Key == (Key{}) {
		s.cp.Key = key
	}
	s.generation++
	next, err := fn(ctx, s.cp, ports.GraphFence{TenantID: key.TenantID, Kind: string(key.Kind), AggregateID: key.AggregateID, Epoch: s.generation})
	if err != nil {
		return err
	}
	if s.failCommit {
		return errors.New("durable commit lost")
	}
	s.cp = next
	return nil
}

type guardedMemoryWriter struct {
	memoryWriter
	active   uint64
	installs int
	writes   int
}

func (w *guardedMemoryWriter) EstablishFence(_ context.Context, guard ports.GraphFence) (string, error) {
	if guard.Epoch <= w.active {
		return "", errors.New("stale graph guard")
	}
	w.active = guard.Epoch
	w.installs++
	return "guard-zed", nil
}
func (w *guardedMemoryWriter) WriteRelationships(ctx context.Context, tenant domain.TenantID, changes []ports.RelationshipChange) (ports.RelationshipWriteResult, error) {
	fence, ok := ports.GraphFenceFromContext(ctx)
	if !ok || fence.Epoch != w.active || fence.TenantID != tenant {
		return ports.RelationshipWriteResult{}, ports.ErrInvalidGraphFence
	}
	w.writes++
	return w.memoryWriter.WriteRelationships(ctx, tenant, changes)
}
func fencedProjector(t *testing.T, w *guardedMemoryWriter) *Projector {
	t.Helper()
	p := projector(t, &memoryWriter{})
	p.writer = w
	return p
}
func (w *guardedMemoryWriter) ReadProjectionEdges(_ context.Context, f ports.GraphFence) ([]domain.Relationship, error) {
	if f.Epoch != w.active {
		return nil, ports.ErrInvalidGraphFence
	}
	kind := "access_binding"
	if f.Kind == "ACCESS_GRANT" {
		kind = "access_grant"
	}
	var edges []domain.Relationship
	for _, edge := range w.edges {
		if edge.TenantID() != f.TenantID {
			continue
		}
		if (edge.Resource().Type() == kind && edge.Resource().ID() == f.AggregateID) || (edge.Subject().Type() == kind && edge.Subject().ID() == f.AggregateID) {
			edges = append(edges, edge)
		}
	}
	return edges, nil
}
func TestFencedWorkerReconcilesEveryNewEpoch(t *testing.T) {
	k := testKey(t, "11111111-1111-4111-8111-111111111111")
	grant := active(t, approved(t, makeGrant(t, k.AggregateID, k.TenantID)))
	writer := &guardedMemoryWriter{}
	projector := fencedProjector(t, writer)
	source := &changingSource{snapshots: map[Key]Snapshot{k: {Key: k, Revision: 1, Grant: &grant}}}
	store := &fencedMemoryStore{}
	worker, err := NewWorker(projector, source, store, func() time.Time { return fixedNow })
	if err != nil {
		t.Fatal(err)
	}
	event := Event{EventID: "test", Key: k, Revision: 1}
	if err := worker.Process(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if store.cp.FenceEpoch != 1 || store.cp.ZedToken == "" || store.cp.Revision != 1 || writer.writes != 1 {
		t.Fatalf("bad initial projection %+v writes=%d", store.cp, writer.writes)
	}
	if err := worker.Process(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if store.cp.FenceEpoch != 2 || writer.writes != 2 || writer.installs != 2 {
		t.Fatalf("takeover did not repair existing graph %+v writes=%d", store.cp, writer.writes)
	}
}
func TestFencedWorkerDoesNotPublishAfterLostCheckpoint(t *testing.T) {
	k := testKey(t, "11111111-1111-4111-8111-111111111111")
	grant := active(t, approved(t, makeGrant(t, k.AggregateID, k.TenantID)))
	writer := &guardedMemoryWriter{}
	projector := fencedProjector(t, writer)
	source := &changingSource{snapshots: map[Key]Snapshot{k: {Key: k, Revision: 1, Grant: &grant}}}
	store := &fencedMemoryStore{failCommit: true}
	worker, _ := NewWorker(projector, source, store, func() time.Time { return fixedNow })
	if err := worker.Process(context.Background(), Event{EventID: "test", Key: k, Revision: 1}); err == nil {
		t.Fatal("checkpoint loss must not ACK")
	}
	if store.cp.Revision != 0 || store.cp.FenceEpoch != 0 {
		t.Fatal("committed after failure")
	}
	store.failCommit = false
	if err := worker.ReconcileKey(context.Background(), k); err != nil {
		t.Fatal(err)
	}
	if store.cp.FenceEpoch != 2 || store.cp.Revision != 1 || writer.writes != 2 {
		t.Fatalf("did not reconcile after failed commit: %+v", store.cp)
	}
}
func TestFencedWorkerRejectsWriterWithoutGraphGuard(t *testing.T) {
	k := testKey(t, "11111111-1111-4111-8111-111111111111")
	writer := &memoryWriter{}
	projector := projector(t, writer)
	source := &changingSource{snapshots: map[Key]Snapshot{}}
	store := &fencedMemoryStore{}
	worker, _ := NewWorker(projector, source, store, func() time.Time { return fixedNow })
	err := worker.ReconcileKey(context.Background(), k)
	if !errors.Is(err, ports.ErrInvalidGraphFence) {
		t.Fatalf("unfenced SpiceDB allowed: %v", err)
	}
}

func TestFencedWorkerRepairsUncheckpointedRevokedGrant(t *testing.T) {
	key := testKey(t, "11111111-1111-4111-8111-111111111111")
	grant := active(t, approved(t, makeGrant(t, key.AggregateID, key.TenantID)))
	source := &changingSource{snapshots: map[Key]Snapshot{key: {Key: key, Revision: 1, Grant: &grant}}}
	writer := &guardedMemoryWriter{}
	proj := fencedProjector(t, writer)
	store := &fencedMemoryStore{failCommit: true}
	worker, _ := NewWorker(proj, source, store, func() time.Time { return fixedNow })
	err := worker.Process(context.Background(), Event{EventID: "grant-activate", Key: key, Revision: 1})
	if err == nil || len(writer.edges) == 0 || store.cp.Revision != 0 {
		t.Fatalf("failed to simulate lost checkpoint: err=%v edges=%d", err, len(writer.edges))
	}
	revoked, err := grant.Revoke(grant.Revision(), subject(t, "user", "security-admin"), fixedNow.Add(-10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	source.snapshots[key] = Snapshot{Key: key, Revision: 2, Grant: &revoked}
	store.failCommit = false
	if err := worker.ReconcileKey(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	if len(writer.edges) != 0 || store.cp.Revision != 2 || store.cp.FenceEpoch != 2 {
		t.Fatalf("stale grant survived crash repair: %+v graph=%v", store.cp, writer.edges)
	}
}
