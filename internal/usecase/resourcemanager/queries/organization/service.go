package organizationqry

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

type QueryService struct {
	GetHandler  GetInteractor
	ListHandler ListInteractor
}

func (s QueryService) GetOrganization(ctx context.Context, input boundaries.GetOrganizationInput) (boundaries.GetOrganizationOutput, error) {
	return s.GetHandler.Execute(ctx, input)
}

func (s QueryService) ListOrganizations(ctx context.Context, input boundaries.ListOrganizationsInput) (boundaries.ListOrganizationsOutput, error) {
	return s.ListHandler.Execute(ctx, input)
}
