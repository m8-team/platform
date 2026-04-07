package query

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/domain/workspace"
	"github.com/m8platform/platform/internal/ports"
)

type GetWorkspace struct {
	ID string
}

type GetWorkspaceHandler struct {
	Repository ports.WorkspaceRepository
}

func (h GetWorkspaceHandler) Handle(ctx context.Context, q GetWorkspace) (workspace.Workspace, error) {
	aggregate, err := h.Repository.GetByID(ctx, q.ID, true)
	if err != nil {
		return workspace.Workspace{}, fmt.Errorf("get workspace: %w", err)
	}
	return aggregate, nil
}

type ListWorkspaces struct {
	OrganizationID string
	PageSize       int32
	PageToken      string
	Filter         string
	OrderBy        string
	ShowDeleted    bool
}

type ListWorkspacesHandler struct {
	Repository   ports.WorkspaceRepository
	FilterParser ports.FilterParser
	OrderParser  ports.OrderParser
}

func (h ListWorkspacesHandler) Handle(ctx context.Context, q ListWorkspaces) (workspace.Page, error) {
	if h.FilterParser != nil {
		if err := h.FilterParser.Validate(q.Filter); err != nil {
			return workspace.Page{}, err
		}
	}
	if h.OrderParser != nil {
		if err := h.OrderParser.Validate(q.OrderBy); err != nil {
			return workspace.Page{}, err
		}
	}
	page, err := h.Repository.List(ctx, workspace.ListParams{
		OrganizationID: q.OrganizationID,
		PageSize:       q.PageSize,
		PageToken:      q.PageToken,
		Filter:         q.Filter,
		OrderBy:        q.OrderBy,
		ShowDeleted:    q.ShowDeleted,
	})
	if err != nil {
		return workspace.Page{}, fmt.Errorf("list workspaces: %w", err)
	}
	return page, nil
}
