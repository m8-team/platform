package projectqry

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type ListInteractor struct {
	Repository      ports.ProjectRepository
	FilterValidator ports.FilterValidator
	OrderValidator  ports.OrderValidator
}

func (i ListInteractor) Execute(ctx context.Context, input boundaries.ListProjectsInput) (boundaries.ListProjectsOutput, error) {
	if i.FilterValidator != nil {
		if err := i.FilterValidator.Validate(input.Filter); err != nil {
			return boundaries.ListProjectsOutput{}, err
		}
	}
	if i.OrderValidator != nil {
		if err := i.OrderValidator.Validate(input.OrderBy); err != nil {
			return boundaries.ListProjectsOutput{}, err
		}
	}

	page, err := i.Repository.List(ctx, ports.ProjectListParams{
		WorkspaceID: input.WorkspaceID,
		PageSize:    input.PageSize,
		PageToken:   input.PageToken,
		Filter:      input.Filter,
		OrderBy:     input.OrderBy,
		ShowDeleted: input.ShowDeleted,
	})
	if err != nil {
		return boundaries.ListProjectsOutput{}, err
	}

	items := make([]boundaries.Project, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, usecasecommon.ProjectToBoundary(item))
	}

	return boundaries.ListProjectsOutput{
		Projects:      items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}
