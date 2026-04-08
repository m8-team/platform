package organizationquery

import (
	"context"
	"fmt"

	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	organizationmapper "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/mapper"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type GetInteractor struct {
	Reader port.OrganizationReader
}

func (i GetInteractor) Execute(ctx context.Context, input organizationboundary.GetOrganizationInput) (organizationboundary.GetOrganizationOutput, error) {
	entity, err := i.Reader.GetByID(ctx, input.ID, input.IncludeDeleted)
	if err != nil {
		return organizationboundary.GetOrganizationOutput{}, fmt.Errorf("get organization: %w", err)
	}
	return organizationboundary.GetOrganizationOutput{
		Organization: organizationmapper.ToBoundary(entity),
	}, nil
}
