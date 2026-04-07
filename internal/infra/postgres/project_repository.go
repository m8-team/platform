package postgres

import (
	"context"
	"database/sql"

	"github.com/m8platform/platform/internal/domain/project"
	"github.com/m8platform/platform/internal/ports"
)

type ProjectRepository struct {
	DB *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{DB: db}
}

func (r *ProjectRepository) GetByID(context.Context, string, bool) (project.Project, error) {
	return project.Project{}, ports.ErrNotImplemented
}

func (r *ProjectRepository) Create(context.Context, project.Project) error {
	return ports.ErrNotImplemented
}

func (r *ProjectRepository) Update(context.Context, project.Project) error {
	return ports.ErrNotImplemented
}

func (r *ProjectRepository) List(context.Context, project.ListParams) (project.Page, error) {
	return project.Page{}, ports.ErrNotImplemented
}
