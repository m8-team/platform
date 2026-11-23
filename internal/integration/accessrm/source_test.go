package accessrm

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
	rmquery "github.com/m8-team/platform/internal/resourcemanager/app/query"
	rmresource "github.com/m8-team/platform/internal/resourcemanager/domain/resource"
)

type fakeResourceRepository struct {
	snapshot rmresource.Snapshot
	err      error
}

func (f fakeResourceRepository) Load(context.Context, rmresource.Kind, string) (rmresource.Snapshot, error) {
	return f.snapshot, f.err
}

func newSource(t *testing.T, snapshot rmresource.Snapshot) *Source {
	t.Helper()
	query, err := rmquery.NewAuthoritativeResource(fakeResourceRepository{snapshot: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	source, err := New(query)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestSourceFailClosedOnForeignTenantAndInactiveState(t *testing.T) {
	tenant, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	ref, _ := domain.NewResource("workspace", "11111111-1111-4111-8111-111111111111")
	snapshot := rmresource.Snapshot{
		Kind: rmresource.Workspace, ID: ref.ID(), TenantID: string(tenant),
		State: rmresource.StateActive, Revision: 3,
		OwnerKind: "organization", OwnerID: "22222222-2222-4222-8222-222222222222",
	}
	source := newSource(t, snapshot)
	got, err := source.AuthoritativeRevision(context.Background(), tenant, ref)
	if err != nil || got != 3 {
		t.Fatalf("revision %d %v", got, err)
	}
	key := projection.Key{TenantID: tenant, Kind: projection.KindRMWorkspace, AggregateID: ref.ID()}
	actual, err := source.LoadCurrent(context.Background(), key)
	if err != nil || actual.Resource == nil || actual.Revision != 3 || actual.Validate() != nil {
		t.Fatalf("bad source snapshot %+v %v", actual, err)
	}

	snapshot.TenantID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	source = newSource(t, snapshot)
	if _, err := source.AuthoritativeRevision(context.Background(), tenant, ref); !errors.Is(err, ErrUntrustedResource) {
		t.Fatalf("cross-tenant revision allowed: %v", err)
	}
	if _, err := source.LoadCurrent(context.Background(), key); !errors.Is(err, ErrUntrustedResource) {
		t.Fatalf("cross-tenant projection allowed: %v", err)
	}

	snapshot.TenantID = string(tenant)
	snapshot.State = rmresource.StateDeleted
	source = newSource(t, snapshot)
	if _, err := source.AuthoritativeRevision(context.Background(), tenant, ref); !errors.Is(err, ErrUntrustedResource) {
		t.Fatalf("deleted resource allowed: %v", err)
	}
	actual, err = source.LoadCurrent(context.Background(), key)
	if err != nil || actual.Resource == nil || actual.Resource.Lifecycle != "DELETED" {
		t.Fatalf("tombstone missing: %+v %v", actual, err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := source.AuthoritativeRevision(cancelled, tenant, ref); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled query reached resource: %v", err)
	}
}

func TestSourceRequiresResourceQueryAndCatalog(t *testing.T) {
	if _, err := New(nil); !errors.Is(err, ErrUntrustedResource) {
		t.Fatalf("accepted empty resource query: %v", err)
	}
	if _, err := NewSource(nil); err == nil {
		t.Fatal("accepted missing authoritative catalog")
	}
}
