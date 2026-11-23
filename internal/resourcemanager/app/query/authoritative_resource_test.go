package query

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/resourcemanager/domain/resource"
)

type testResourceRepository struct {
	snapshot resource.Snapshot
	err      error
}

func (f testResourceRepository) Load(context.Context, resource.Kind, string) (resource.Snapshot, error) {
	return f.snapshot, f.err
}

func TestAuthoritativeResourceDenyByDefault(t *testing.T) {
	_, err := NewAuthoritativeResource(nil)
	if !errors.Is(err, ErrResourceReaderRequired) {
		t.Fatalf("nil repository accepted: %v", err)
	}
	snapshot := resource.Snapshot{Kind: resource.Workspace, ID: "w1", TenantID: "t1", State: resource.StateActive, Revision: 4}
	q, _ := NewAuthoritativeResource(testResourceRepository{snapshot: snapshot})
	got, err := q.GetActive(context.Background(), "t1", resource.Workspace, "w1")
	if err != nil || got.Revision != 4 {
		t.Fatalf("trusted read: %+v %v", got, err)
	}
	for _, tc := range []struct {
		tenant string
		kind   resource.Kind
		id     string
	}{
		{"other", resource.Workspace, "w1"},
		{"t1", resource.Organization, "w1"},
		{"t1", resource.Workspace, "other"},
	} {
		if _, err := q.Get(context.Background(), tc.tenant, tc.kind, tc.id); !errors.Is(err, ErrResourceScopeMismatch) {
			t.Fatalf("mismatch %+v: %v", tc, err)
		}
	}
	snapshot.State = resource.StateDeleted
	q, _ = NewAuthoritativeResource(testResourceRepository{snapshot: snapshot})
	if _, err := q.GetActive(context.Background(), "t1", resource.Workspace, "w1"); !errors.Is(err, ErrResourceInactive) {
		t.Fatalf("deleted resource permitted: %v", err)
	}
	if _, err := q.Get(context.Background(), "t1", resource.Workspace, "w1"); err != nil {
		t.Fatalf("tombstone must be projectable: %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := q.Get(cancelled, "t1", resource.Workspace, "w1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("ignored cancellation: %v", err)
	}
	snapshot.Revision = 0
	q, _ = NewAuthoritativeResource(testResourceRepository{snapshot: snapshot})
	if _, err := q.Get(context.Background(), "t1", resource.Workspace, "w1"); !errors.Is(err, resource.ErrInvalidSnapshot) {
		t.Fatalf("invalid revision accepted: %v", err)
	}
}
