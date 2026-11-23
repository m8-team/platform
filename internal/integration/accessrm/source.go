// Package accessrm is the integration boundary for resource ownership facts.
// It intentionally lives outside both bounded contexts.
package accessrm

import (
	"context"
	"errors"
	"fmt"

	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
	rmcatalog "github.com/m8-team/platform/internal/resourcemanager/adapter/resourcecatalog"
	rmquery "github.com/m8-team/platform/internal/resourcemanager/app/query"
	rmresource "github.com/m8-team/platform/internal/resourcemanager/domain/resource"
)

var ErrUntrustedResource = errors.New("resource manager resource tenant, kind or lifecycle mismatch")

type Source struct {
	resources *rmquery.AuthoritativeResource
}

func New(resources *rmquery.AuthoritativeResource) (*Source, error) {
	if resources == nil {
		return nil, ErrUntrustedResource
	}
	return &Source{resources: resources}, nil
}

var _ consistency.ResourceRevisionReader = (*Source)(nil)
var _ projection.SnapshotReader = (*Source)(nil)

func (s *Source) load(ctx context.Context, tenant domain.TenantID, ref domain.Resource, active bool) (rmresource.Snapshot, error) {
	if ctx == nil || s == nil || s.resources == nil || tenant.Validate() != nil || ref.Validate() != nil {
		return rmresource.Snapshot{}, ErrUntrustedResource
	}
	if err := ctx.Err(); err != nil {
		return rmresource.Snapshot{}, err
	}
	if _, err := projection.KindForResourceType(ref.Type()); err != nil {
		return rmresource.Snapshot{}, ErrUntrustedResource
	}
	var snapshot rmresource.Snapshot
	var err error
	if active {
		snapshot, err = s.resources.GetActive(ctx, string(tenant), rmresource.Kind(ref.Type()), ref.ID())
	} else {
		snapshot, err = s.resources.Get(ctx, string(tenant), rmresource.Kind(ref.Type()), ref.ID())
	}
	if err != nil {
		if errors.Is(err, rmquery.ErrResourceScopeMismatch) ||
			errors.Is(err, rmquery.ErrResourceInactive) ||
			errors.Is(err, rmresource.ErrInvalidSnapshot) {
			return rmresource.Snapshot{}, fmt.Errorf("%w: %v", ErrUntrustedResource, err)
		}
		return rmresource.Snapshot{}, err
	}
	return snapshot, nil
}

// AuthoritativeRevision fails closed for deleted/suspended resources and
// permits no revision supplied by a caller or an event payload.
func (s *Source) AuthoritativeRevision(ctx context.Context, tenant domain.TenantID, ref domain.Resource) (uint64, error) {
	snapshot, err := s.load(ctx, tenant, ref, true)
	if err != nil {
		return 0, err
	}
	return uint64(snapshot.Revision), nil
}

// LoadCurrent includes tombstones so projection can remove obsolete tuples.
func (s *Source) LoadCurrent(ctx context.Context, key projection.Key) (projection.Snapshot, error) {
	if err := key.Validate(); err != nil {
		return projection.Snapshot{}, err
	}
	kind, err := projection.ResourceTypeForKind(key.Kind)
	if err != nil {
		return projection.Snapshot{}, err
	}
	ref, err := domain.NewResource(kind, key.AggregateID)
	if err != nil {
		return projection.Snapshot{}, err
	}
	snapshot, err := s.load(ctx, key.TenantID, ref, false)
	if err != nil {
		return projection.Snapshot{}, err
	}
	state := &projection.ResourceState{
		TenantID: key.TenantID, Resource: ref,
		Revision: uint64(snapshot.Revision), Lifecycle: string(snapshot.State),
	}
	return projection.Snapshot{Key: key, Revision: snapshot.Revision, Resource: state}, nil
}

// NewSource wires Resource Manager's authoritative query and its repository
// adapter to the Access-facing revision and snapshot ports. Neither domain
// imports this integration package.
func NewSource(catalog rmcatalog.Catalog) (*Source, error) {
	repo, err := rmcatalog.New(catalog)
	if err != nil {
		return nil, err
	}
	query, err := rmquery.NewAuthoritativeResource(repo)
	if err != nil {
		return nil, err
	}
	return New(query)
}
