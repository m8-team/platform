package organizationquery

import (
	"context"

	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	organizationmapper "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/mapper"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type ListInteractor struct {
	Reader port.OrganizationReader
}

func (i ListInteractor) Execute(ctx context.Context, input organizationboundary.ListOrganizationsInput) (organizationboundary.ListOrganizationsOutput, error) {
	page, err := i.Reader.List(ctx, port.OrganizationListParams{
		PageSize:    input.PageSize,
		PageToken:   input.PageToken,
		Filter:      input.Filter,
		OrderBy:     input.OrderBy,
		ShowDeleted: input.ShowDeleted,
	})
	if err != nil {
		return organizationboundary.ListOrganizationsOutput{}, err
	}

	items := make([]organizationboundary.Organization, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, organizationmapper.ToBoundary(item))
	}

	return organizationboundary.ListOrganizationsOutput{
		Organizations: items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}
