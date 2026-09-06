package organizationapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func (s *OrganizationService) Delete(
	ctx context.Context,
	cmd DeleteOrganization,
) (out *organization.Organization, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.organization.delete")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := cmd.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionDeleteOrganization, cmd.ID); err != nil {
		return nil, err
	}

	var result *organization.Organization
	err := s.workspaceChildren.WithOrganizationLock(ctx, cmd.ID, func(ctx context.Context) error {
		value, err := s.repository.Get(ctx, cmd.ID)
		if err != nil {
			if cmd.AllowMissing && errors.Is(err, ports.ErrOrganizationNotFound) {
				return nil
			}
			return fmt.Errorf("get organization for delete: %w", err)
		}
		if value.IsDeleted() {
			if cmd.AllowMissing {
				result = value
				return nil
			}
			return organization.ErrOrganizationAlreadyDeleted
		}
		if err := value.CheckVersion(cmd.ExpectedVersion); err != nil {
			return err
		}

		hasChildren, err := s.workspaceChildren.HasNonDeleted(ctx, cmd.ID)
		if err != nil {
			return fmt.Errorf("check organization workspaces: %w", err)
		}
		if hasChildren {
			return ErrOrganizationHasWorkspaces
		}

		expectedStoredVersion := value.Version()
		now := s.clock.Now().UTC()
		if err := value.Delete(organization.DeleteParams{
			Now:             now,
			PurgeTime:       now.Add(s.retention),
			ExpectedVersion: cmd.ExpectedVersion,
		}); err != nil {
			return err
		}
		if err := s.repository.Update(ctx, value, expectedStoredVersion); err != nil {
			return fmt.Errorf("delete organization: %w", err)
		}
		result = value.Clone()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
