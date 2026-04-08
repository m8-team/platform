package projectquery

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type ListInteractor struct {
	Repository      port.ProjectRepository
	FilterValidator port.FilterValidator
	OrderValidator  port.OrderValidator
}

func (i ListInteractor) Execute(ctx context.Context, input boundary.ListProjectsInput) (boundary.ListProjectsOutput, error) {
	if i.FilterValidator != nil {
		if err := i.FilterValidator.Validate(input.Filter); err != nil {
			return boundary.ListProjectsOutput{}, err
		}
	}
	if i.OrderValidator != nil {
		if err := i.OrderValidator.Validate(input.OrderBy); err != nil {
			return boundary.ListProjectsOutput{}, err
		}
	}

	page, err := i.Repository.List(ctx, port.ProjectListParams{
		WorkspaceID: input.WorkspaceID,
		PageSize:    input.PageSize,
		PageToken:   input.PageToken,
		Filter:      input.Filter,
		OrderBy:     input.OrderBy,
		ShowDeleted: input.ShowDeleted,
	})
	if err != nil {
		return boundary.ListProjectsOutput{}, err
	}

	items := make([]boundary.Project, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, usecasecommon.ProjectToBoundary(item))
	}

	return boundary.ListProjectsOutput{
		Projects:      items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}
