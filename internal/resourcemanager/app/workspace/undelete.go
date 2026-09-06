package workspaceapp

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func (s *WorkspaceService) Undelete(ctx context.Context, cmd UndeleteWorkspace) (out *workspace.Workspace, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.workspace.undelete")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := cmd.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionUndeleteWorkspace, organization.ID{}, cmd.ID); err != nil {
		return nil, err
	}
	value, err := s.repository.Get(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("get workspace for undelete: %w", err)
	}

	var result *workspace.Workspace
	err = s.repository.WithOrganizationLock(ctx, value.OrganizationID(), func(ctx context.Context) error {
		parent, err := s.organizations.Get(ctx, value.OrganizationID())
		if err != nil {
			return fmt.Errorf("get parent organization for undelete: %w", err)
		}
		if err := parent.CanCreateWorkspace(); err != nil {
			return err
		}
		current, err := s.repository.Get(ctx, cmd.ID)
		if err != nil {
			return fmt.Errorf("reload workspace for undelete: %w", err)
		}
		expectedStoredVersion := current.Version()
		if err := current.Undelete(workspace.UndeleteParams{
			Now:             s.clock.Now().UTC(),
			ExpectedVersion: cmd.ExpectedVersion,
		}); err != nil {
			return err
		}
		if err := s.repository.Update(ctx, current, expectedStoredVersion); err != nil {
			return fmt.Errorf("undelete workspace: %w", err)
		}
		result = current.Clone()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
