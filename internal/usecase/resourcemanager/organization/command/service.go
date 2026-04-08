package organizationcommand

import (
	"context"

	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
)

type CommandService struct {
	CreateHandler   organizationboundary.CreateHandler
	UpdateHandler   organizationboundary.UpdateHandler
	DeleteHandler   organizationboundary.DeleteHandler
	UndeleteHandler organizationboundary.UndeleteHandler
}

func (s CommandService) CreateOrganization(ctx context.Context, input organizationboundary.CreateOrganizationInput) (organizationboundary.CreateOrganizationOutput, error) {
	return s.CreateHandler.Execute(ctx, input)
}

func (s CommandService) UpdateOrganization(ctx context.Context, input organizationboundary.UpdateOrganizationInput) (organizationboundary.UpdateOrganizationOutput, error) {
	return s.UpdateHandler.Execute(ctx, input)
}

func (s CommandService) DeleteOrganization(ctx context.Context, input organizationboundary.DeleteOrganizationInput) (organizationboundary.DeleteOrganizationOutput, error) {
	return s.DeleteHandler.Execute(ctx, input)
}

func (s CommandService) UndeleteOrganization(ctx context.Context, input organizationboundary.UndeleteOrganizationInput) (organizationboundary.UndeleteOrganizationOutput, error) {
	return s.UndeleteHandler.Execute(ctx, input)
}
