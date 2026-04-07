package postgres

import (
	"context"
	"database/sql"

	"github.com/m8platform/platform/internal/ports"
)

type HierarchyRepository struct {
	DB *sql.DB
}

func NewHierarchyRepository(db *sql.DB) *HierarchyRepository {
	return &HierarchyRepository{DB: db}
}

func (r *HierarchyRepository) GetOrganizationNode(context.Context, string) (ports.HierarchyNode, error) {
	return ports.HierarchyNode{}, ports.ErrNotImplemented
}

func (r *HierarchyRepository) GetWorkspaceNode(context.Context, string) (ports.HierarchyNode, error) {
	return ports.HierarchyNode{}, ports.ErrNotImplemented
}

func (r *HierarchyRepository) HasActiveWorkspaces(context.Context, string) (bool, error) {
	return false, ports.ErrNotImplemented
}

func (r *HierarchyRepository) HasActiveProjects(context.Context, string) (bool, error) {
	return false, ports.ErrNotImplemented
}
