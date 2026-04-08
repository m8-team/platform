package organizationquery

import (
	"context"

	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
)

type QueryService struct {
	GetHandler    organizationboundary.GetHandler
	ListHandler   organizationboundary.ListHandler
	ListValidator ListInputValidator
}

func (s QueryService) GetOrganization(ctx context.Context, input organizationboundary.GetOrganizationInput) (organizationboundary.GetOrganizationOutput, error) {
	return s.GetHandler.Execute(ctx, input)
}

func (s QueryService) ListOrganizations(ctx context.Context, input organizationboundary.ListOrganizationsInput) (organizationboundary.ListOrganizationsOutput, error) {
	if s.ListValidator != nil {
		if err := s.ListValidator.Validate(input); err != nil {
			return organizationboundary.ListOrganizationsOutput{}, err
		}
	}
	return s.ListHandler.Execute(ctx, input)
}
