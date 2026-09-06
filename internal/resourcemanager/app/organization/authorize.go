package organizationapp

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func (s *OrganizationService) authorize(
	ctx context.Context,
	action ports.AuthorizationAction,
	id organization.ID,
) error {
	if err := s.authorizer.Authorize(ctx, ports.AuthorizationRequest{
		Action:         action,
		OrganizationID: id,
	}); err != nil {
		return fmt.Errorf("authorize %s: %w", action, err)
	}

	return nil
}
