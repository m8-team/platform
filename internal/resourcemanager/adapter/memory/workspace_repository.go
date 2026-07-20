package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

type WorkspaceRepository struct {
	mu     sync.RWMutex
	values map[workspace.ID]*workspace.Workspace
}

func NewWorkspaceRepository() *WorkspaceRepository {
	return &WorkspaceRepository{values: make(map[workspace.ID]*workspace.Workspace)}
}
func (r *WorkspaceRepository) Create(ctx context.Context, w *workspace.Workspace) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if w == nil {
		return workspace.ErrNilWorkspace
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.values[w.ID()]; ok {
		return ports.ErrWorkspaceAlreadyExists
	}
	r.values[w.ID()] = w.Clone()
	return nil
}
func (r *WorkspaceRepository) Get(ctx context.Context, id workspace.ID) (*workspace.Workspace, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.values[id]
	if !ok {
		return nil, fmt.Errorf("%w: workspace_id=%s", ports.ErrWorkspaceNotFound, id)
	}
	return w.Clone(), nil
}
func (r *WorkspaceRepository) Update(ctx context.Context, w *workspace.Workspace, expected types.Version) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if w == nil {
		return workspace.ErrNilWorkspace
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.values[w.ID()]
	if !ok {
		return ports.ErrWorkspaceNotFound
	}
	if !current.Version().Equal(expected) {
		return ports.ErrWorkspaceVersionConflict
	}
	r.values[w.ID()] = w.Clone()
	return nil
}
func (r *WorkspaceRepository) List(ctx context.Context, organizationID organization.ID, showDeleted bool) ([]*workspace.Workspace, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*workspace.Workspace, 0)
	for _, w := range r.values {
		if w.OrganizationID().Equal(organizationID) && (showDeleted || !w.IsDeleted()) {
			result = append(result, w.Clone())
		}
	}
	return result, nil
}

var _ ports.WorkspaceRepository = (*WorkspaceRepository)(nil)
