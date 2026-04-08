package postgres

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type HierarchyReader struct {
	Store *Store
}

func (r HierarchyReader) GetOrganizationNode(_ context.Context, id string) (ports.HierarchyNode, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	entity, ok := r.Store.organizations[id]
	if !ok {
		return ports.HierarchyNode{ID: id}, nil
	}
	return ports.HierarchyNode{
		ID:      id,
		Exists:  true,
		Deleted: entity.IsDeleted(),
	}, nil
}

func (r HierarchyReader) GetWorkspaceNode(_ context.Context, id string) (ports.HierarchyNode, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	entity, ok := r.Store.workspaces[id]
	if !ok {
		return ports.HierarchyNode{ID: id}, nil
	}
	return ports.HierarchyNode{
		ID:      id,
		Exists:  true,
		Deleted: entity.IsDeleted(),
	}, nil
}

func (r HierarchyReader) HasActiveWorkspaces(_ context.Context, organizationID string) (bool, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	for _, entity := range r.Store.workspaces {
		if entity.OrganizationID == organizationID && !entity.IsDeleted() {
			return true, nil
		}
	}
	return false, nil
}

func (r HierarchyReader) HasActiveProjects(_ context.Context, workspaceID string) (bool, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	for _, entity := range r.Store.projects {
		if entity.WorkspaceID == workspaceID && !entity.IsDeleted() {
			return true, nil
		}
	}
	return false, nil
}
