package query

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/resource"
)

var (
	ErrResourceReaderRequired = errors.New("authoritative resource repository is required")
	ErrInvalidResourceQuery   = errors.New("invalid resource query")
	ErrResourceScopeMismatch  = errors.New("resource does not match the trusted tenant, kind or identity")
	ErrResourceInactive       = errors.New("resource is not active")
)

// AuthoritativeResource is Resource Manager's CQRS query handler for the
// tenant-scoped snapshot consumed by Access and other internal projections.
// The tenant is supplied by trusted service identity, not a request header.
type AuthoritativeResource struct {
	repository ports.ResourceSnapshotRepository
}

func NewAuthoritativeResource(repository ports.ResourceSnapshotRepository) (*AuthoritativeResource, error) {
	if repository == nil {
		return nil, ErrResourceReaderRequired
	}
	return &AuthoritativeResource{repository: repository}, nil
}

func (q *AuthoritativeResource) Get(ctx context.Context, tenantID string, kind resource.Kind, id string) (resource.Snapshot, error) {
	if ctx == nil || q == nil || q.repository == nil {
		return resource.Snapshot{}, ErrResourceReaderRequired
	}
	if err := ctx.Err(); err != nil {
		return resource.Snapshot{}, err
	}
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(id) == "" || !kind.Valid() {
		return resource.Snapshot{}, ErrInvalidResourceQuery
	}
	snapshot, err := q.repository.Load(ctx, kind, id)
	if err != nil {
		return resource.Snapshot{}, fmt.Errorf("load authoritative resource: %w", err)
	}
	if err := snapshot.Validate(); err != nil {
		return resource.Snapshot{}, err
	}
	if snapshot.TenantID != tenantID || snapshot.Kind != kind || snapshot.ID != id {
		return resource.Snapshot{}, ErrResourceScopeMismatch
	}
	return snapshot, nil
}

// GetActive is for security-critical decisions. Use Get for projection
// tombstones so revoked or deleted resources remove their graph tuples.
func (q *AuthoritativeResource) GetActive(ctx context.Context, tenantID string, kind resource.Kind, id string) (resource.Snapshot, error) {
	snapshot, err := q.Get(ctx, tenantID, kind, id)
	if err != nil {
		return resource.Snapshot{}, err
	}
	if !snapshot.Active() {
		return resource.Snapshot{}, ErrResourceInactive
	}
	return snapshot, nil
}
