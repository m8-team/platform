package workspacequery

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type ListInteractor struct {
	Repository      port.WorkspaceRepository
	FilterValidator port.FilterValidator
	OrderValidator  port.OrderValidator
}

func (i ListInteractor) Execute(ctx context.Context, input boundary.ListWorkspacesInput) (boundary.ListWorkspacesOutput, error) {
	if i.FilterValidator != nil {
		if err := i.FilterValidator.Validate(input.Filter); err != nil {
			return boundary.ListWorkspacesOutput{}, err
		}
	}
	if i.OrderValidator != nil {
		if err := i.OrderValidator.Validate(input.OrderBy); err != nil {
			return boundary.ListWorkspacesOutput{}, err
		}
	}

	page, err := i.Repository.List(ctx, port.WorkspaceListParams{
		OrganizationID: input.OrganizationID,
		PageSize:       input.PageSize,
		PageToken:      input.PageToken,
		Filter:         input.Filter,
		OrderBy:        input.OrderBy,
		ShowDeleted:    input.ShowDeleted,
	})
	if err != nil {
		return boundary.ListWorkspacesOutput{}, err
	}

	items := make([]boundary.Workspace, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, usecasecommon.WorkspaceToBoundary(item))
	}

	return boundary.ListWorkspacesOutput{
		Workspaces:    items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}
