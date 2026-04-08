package workspaceqry

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type ListInteractor struct {
	Repository      ports.WorkspaceRepository
	FilterValidator ports.FilterValidator
	OrderValidator  ports.OrderValidator
}

func (i ListInteractor) Execute(ctx context.Context, input boundaries.ListWorkspacesInput) (boundaries.ListWorkspacesOutput, error) {
	if i.FilterValidator != nil {
		if err := i.FilterValidator.Validate(input.Filter); err != nil {
			return boundaries.ListWorkspacesOutput{}, err
		}
	}
	if i.OrderValidator != nil {
		if err := i.OrderValidator.Validate(input.OrderBy); err != nil {
			return boundaries.ListWorkspacesOutput{}, err
		}
	}

	page, err := i.Repository.List(ctx, ports.WorkspaceListParams{
		OrganizationID: input.OrganizationID,
		PageSize:       input.PageSize,
		PageToken:      input.PageToken,
		Filter:         input.Filter,
		OrderBy:        input.OrderBy,
		ShowDeleted:    input.ShowDeleted,
	})
	if err != nil {
		return boundaries.ListWorkspacesOutput{}, err
	}

	items := make([]boundaries.Workspace, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, usecasecommon.WorkspaceToBoundary(item))
	}

	return boundaries.ListWorkspacesOutput{
		Workspaces:    items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}
