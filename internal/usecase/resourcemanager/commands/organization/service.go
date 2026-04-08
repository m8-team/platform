package organizationcmd

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

type CommandService struct {
	CreateHandler   CreateInteractor
	UpdateHandler   UpdateInteractor
	DeleteHandler   DeleteInteractor
	UndeleteHandler UndeleteInteractor
}

func (s CommandService) CreateOrganization(ctx context.Context, input boundaries.CreateOrganizationInput) (boundaries.CreateOrganizationOutput, error) {
	return s.CreateHandler.Execute(ctx, input)
}

func (s CommandService) UpdateOrganization(ctx context.Context, input boundaries.UpdateOrganizationInput) (boundaries.UpdateOrganizationOutput, error) {
	return s.UpdateHandler.Execute(ctx, input)
}

func (s CommandService) DeleteOrganization(ctx context.Context, input boundaries.DeleteOrganizationInput) (boundaries.DeleteOrganizationOutput, error) {
	return s.DeleteHandler.Execute(ctx, input)
}

func (s CommandService) UndeleteOrganization(ctx context.Context, input boundaries.UndeleteOrganizationInput) (boundaries.UndeleteOrganizationOutput, error) {
	return s.UndeleteHandler.Execute(ctx, input)
}
