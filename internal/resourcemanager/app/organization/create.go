package organizationapp

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func (s *OrganizationService) Create(
	ctx context.Context,
	cmd CreateOrganization,
) (out *organization.Organization, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.organization.create")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, ports.ActionCreateOrganization, organization.ID{}); err != nil {
		return nil, err
	}

	id := s.idGenerator.NewID()
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGeneratedOrganizationID, err)
	}

	value, err := organization.New(organization.CreateParams{
		ID:          id,
		Name:        cmd.Name,
		Description: cmd.Description,
		Labels:      cmd.Labels,
		Now:         s.clock.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}
	if err := s.repository.Create(ctx, value); err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	return value.Clone(), nil
}
