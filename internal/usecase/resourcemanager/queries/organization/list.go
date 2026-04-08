package organizationqry

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type ListInteractor struct {
	Repository      ports.OrganizationRepository
	FilterValidator ports.FilterValidator
	OrderValidator  ports.OrderValidator
}

func (i ListInteractor) Execute(ctx context.Context, input boundaries.ListOrganizationsInput) (boundaries.ListOrganizationsOutput, error) {
	if i.FilterValidator != nil {
		if err := i.FilterValidator.Validate(input.Filter); err != nil {
			return boundaries.ListOrganizationsOutput{}, err
		}
	}
	if i.OrderValidator != nil {
		if err := i.OrderValidator.Validate(input.OrderBy); err != nil {
			return boundaries.ListOrganizationsOutput{}, err
		}
	}

	page, err := i.Repository.List(ctx, ports.OrganizationListParams{
		PageSize:    input.PageSize,
		PageToken:   input.PageToken,
		Filter:      input.Filter,
		OrderBy:     input.OrderBy,
		ShowDeleted: input.ShowDeleted,
	})
	if err != nil {
		return boundaries.ListOrganizationsOutput{}, err
	}

	items := make([]boundaries.Organization, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, usecasecommon.OrganizationToBoundary(item))
	}

	return boundaries.ListOrganizationsOutput{
		Organizations: items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}
