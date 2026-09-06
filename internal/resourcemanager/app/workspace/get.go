package workspaceapp

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func (s *WorkspaceService) Get(ctx context.Context, q GetWorkspace) (out *workspace.Workspace, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.workspace.get")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := q.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionGetWorkspace, organization.ID{}, q.ID); err != nil {
		return nil, err
	}
	value, err := s.repository.Get(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get workspace: %w", err)
	}
	return value, nil
}
