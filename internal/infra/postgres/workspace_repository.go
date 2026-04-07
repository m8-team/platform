package postgres

import (
	"context"
	"database/sql"

	"github.com/m8platform/platform/internal/domain/workspace"
	"github.com/m8platform/platform/internal/ports"
)

type WorkspaceRepository struct {
	DB *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) *WorkspaceRepository {
	return &WorkspaceRepository{DB: db}
}

func (r *WorkspaceRepository) GetByID(context.Context, string, bool) (workspace.Workspace, error) {
	return workspace.Workspace{}, ports.ErrNotImplemented
}

func (r *WorkspaceRepository) Create(context.Context, workspace.Workspace) error {
	return ports.ErrNotImplemented
}

func (r *WorkspaceRepository) Update(context.Context, workspace.Workspace) error {
	return ports.ErrNotImplemented
}

func (r *WorkspaceRepository) List(context.Context, workspace.ListParams) (workspace.Page, error) {
	return workspace.Page{}, ports.ErrNotImplemented
}
