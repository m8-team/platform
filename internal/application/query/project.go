package query

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/domain/project"
	"github.com/m8platform/platform/internal/ports"
)

type GetProject struct {
	ID string
}

type GetProjectHandler struct {
	Repository ports.ProjectRepository
}

func (h GetProjectHandler) Handle(ctx context.Context, q GetProject) (project.Project, error) {
	aggregate, err := h.Repository.GetByID(ctx, q.ID, true)
	if err != nil {
		return project.Project{}, fmt.Errorf("get project: %w", err)
	}
	return aggregate, nil
}

type ListProjects struct {
	WorkspaceID string
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type ListProjectsHandler struct {
	Repository   ports.ProjectRepository
	FilterParser ports.FilterParser
	OrderParser  ports.OrderParser
}

func (h ListProjectsHandler) Handle(ctx context.Context, q ListProjects) (project.Page, error) {
	if h.FilterParser != nil {
		if err := h.FilterParser.Validate(q.Filter); err != nil {
			return project.Page{}, err
		}
	}
	if h.OrderParser != nil {
		if err := h.OrderParser.Validate(q.OrderBy); err != nil {
			return project.Page{}, err
		}
	}
	page, err := h.Repository.List(ctx, project.ListParams{
		WorkspaceID: q.WorkspaceID,
		PageSize:    q.PageSize,
		PageToken:   q.PageToken,
		Filter:      q.Filter,
		OrderBy:     q.OrderBy,
		ShowDeleted: q.ShowDeleted,
	})
	if err != nil {
		return project.Page{}, fmt.Errorf("list projects: %w", err)
	}
	return page, nil
}
