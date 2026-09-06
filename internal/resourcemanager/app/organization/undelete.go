package organizationapp

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func (s *OrganizationService) Undelete(
	ctx context.Context,
	cmd UndeleteOrganization,
) (out *organization.Organization, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.organization.undelete")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := cmd.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionUndeleteOrganization, cmd.ID); err != nil {
		return nil, err
	}

	value, err := s.repository.Get(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("get organization for undelete: %w", err)
	}
	expectedStoredVersion := value.Version()
	if err := value.Undelete(organization.UndeleteParams{
		Now:             s.clock.Now().UTC(),
		ExpectedVersion: cmd.ExpectedVersion,
	}); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, value, expectedStoredVersion); err != nil {
		return nil, fmt.Errorf("undelete organization: %w", err)
	}

	return value.Clone(), nil
}
