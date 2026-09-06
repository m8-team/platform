package workspaceapp

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func (s *WorkspaceService) Create(ctx context.Context, cmd CreateWorkspace) (out *workspace.Workspace, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.workspace.create")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := cmd.OrganizationID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionCreateWorkspace, cmd.OrganizationID, workspace.ID{}); err != nil {
		return nil, err
	}

	var result *workspace.Workspace
	err := s.repository.WithOrganizationLock(ctx, cmd.OrganizationID, func(ctx context.Context) error {
		parent, err := s.organizations.Get(ctx, cmd.OrganizationID)
		if err != nil {
			return fmt.Errorf("get parent organization: %w", err)
		}
		if err := parent.CanCreateWorkspace(); err != nil {
			return err
		}

		id := s.idGenerator.NewWorkspaceID()
		if err := id.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrGeneratedWorkspaceID, err)
		}
		value, err := workspace.New(workspace.CreateParams{
			ID:             id,
			OrganizationID: cmd.OrganizationID,
			Name:           cmd.Name,
			Description:    cmd.Description,
			Labels:         cmd.Labels,
			Now:            s.clock.Now().UTC(),
		})
		if err != nil {
			return err
		}
		if err := s.repository.Create(ctx, value); err != nil {
			return fmt.Errorf("create workspace: %w", err)
		}
		result = value.Clone()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
