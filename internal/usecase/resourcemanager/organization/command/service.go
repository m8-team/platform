package organizationcommand

import (
	"context"

	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/requestmeta"
)

type CommandService struct {
	CreateHandler     organizationboundary.CreateHandler
	UpdateHandler     organizationboundary.UpdateHandler
	DeleteHandler     organizationboundary.DeleteHandler
	UndeleteHandler   organizationboundary.UndeleteHandler
	MetadataValidator MetadataValidator
}

func (s CommandService) CreateOrganization(ctx context.Context, input organizationboundary.CreateOrganizationInput) (organizationboundary.CreateOrganizationOutput, error) {
	if err := s.validateMetadata(input.Metadata); err != nil {
		return organizationboundary.CreateOrganizationOutput{}, err
	}
	return s.CreateHandler.Execute(ctx, input)
}

func (s CommandService) UpdateOrganization(ctx context.Context, input organizationboundary.UpdateOrganizationInput) (organizationboundary.UpdateOrganizationOutput, error) {
	if err := s.validateMetadata(input.Metadata); err != nil {
		return organizationboundary.UpdateOrganizationOutput{}, err
	}
	return s.UpdateHandler.Execute(ctx, input)
}

func (s CommandService) DeleteOrganization(ctx context.Context, input organizationboundary.DeleteOrganizationInput) (organizationboundary.DeleteOrganizationOutput, error) {
	if err := s.validateMetadata(input.Metadata); err != nil {
		return organizationboundary.DeleteOrganizationOutput{}, err
	}
	return s.DeleteHandler.Execute(ctx, input)
}

func (s CommandService) UndeleteOrganization(ctx context.Context, input organizationboundary.UndeleteOrganizationInput) (organizationboundary.UndeleteOrganizationOutput, error) {
	if err := s.validateMetadata(input.Metadata); err != nil {
		return organizationboundary.UndeleteOrganizationOutput{}, err
	}
	return s.UndeleteHandler.Execute(ctx, input)
}

func (s CommandService) validateMetadata(metadata requestmeta.RequestMetadata) error {
	if s.MetadataValidator == nil {
		return nil
	}
	return s.MetadataValidator.Validate(metadata)
}
