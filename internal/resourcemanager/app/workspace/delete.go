package workspaceapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func (s *WorkspaceService) Delete(ctx context.Context, cmd DeleteWorkspace) (out *workspace.Workspace, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.workspace.delete")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := cmd.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionDeleteWorkspace, organization.ID{}, cmd.ID); err != nil {
		return nil, err
	}
	value, err := s.repository.Get(ctx, cmd.ID)
	if err != nil {
		if cmd.AllowMissing && errors.Is(err, ports.ErrWorkspaceNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get workspace for delete: %w", err)
	}
	if value.IsDeleted() {
		if cmd.AllowMissing {
			return value, nil
		}
		return nil, workspace.ErrWorkspaceAlreadyDeleted
	}

	expectedStoredVersion := value.Version()
	now := s.clock.Now().UTC()
	if err := value.Delete(workspace.DeleteParams{
		Now:             now,
		PurgeTime:       now.Add(s.retention),
		ExpectedVersion: cmd.ExpectedVersion,
	}); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, value, expectedStoredVersion); err != nil {
		return nil, fmt.Errorf("delete workspace: %w", err)
	}
	return value.Clone(), nil
}
