// Package resourcecatalog maps the legacy runtime/YDB record representation
// into Resource Manager's domain snapshot read port. No caller may write state
// through this adapter.
package resourcecatalog

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/resource"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

var ErrCatalogRequired = errors.New("authoritative resource catalog is required")

// Catalog is the existing owner-backed YDB catalog's narrow read API.
type Catalog interface {
	Get(context.Context, v1engine.Kind, string) (v1engine.Record, error)
}

type Repository struct {
	catalog Catalog
}

func New(catalog Catalog) (*Repository, error) {
	if catalog == nil {
		return nil, ErrCatalogRequired
	}
	return &Repository{catalog: catalog}, nil
}

var _ ports.ResourceSnapshotRepository = (*Repository)(nil)

func (r *Repository) Load(ctx context.Context, kind resource.Kind, id string) (resource.Snapshot, error) {
	if r == nil || r.catalog == nil {
		return resource.Snapshot{}, ErrCatalogRequired
	}
	if err := ctx.Err(); err != nil {
		return resource.Snapshot{}, err
	}
	record, err := r.catalog.Get(ctx, v1engine.Kind(kind), id)
	if err != nil {
		return resource.Snapshot{}, err
	}
	snapshot := resource.Snapshot{}
	snapshot.Kind = resource.Kind(record.Kind)
	snapshot.ID = record.ID
	snapshot.TenantID = record.TenantID
	snapshot.OwnerKind = string(record.Owner.Kind)
	snapshot.OwnerID = record.Owner.ID
	snapshot.OrganizationID = record.OrganizationID
	snapshot.State = resource.State(record.State)
	snapshot.Revision = record.Revision
	if err := snapshot.Validate(); err != nil {
		return resource.Snapshot{}, err
	}
	return snapshot, nil
}
