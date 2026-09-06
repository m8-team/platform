package workspaceapp

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func (s *WorkspaceService) authorize(
	ctx context.Context,
	action ports.AuthorizationAction,
	organizationID organization.ID,
	workspaceID workspace.ID,
) error {
	if err := s.authorizer.Authorize(ctx, ports.AuthorizationRequest{
		Action:         action,
		OrganizationID: organizationID,
		WorkspaceID:    workspaceID,
	}); err != nil {
		return fmt.Errorf("authorize %s: %w", action, err)
	}
	return nil
}
