package organizationapp

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func (s *OrganizationService) Get(
	ctx context.Context,
	q GetOrganization,
) (out *organization.Organization, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.organization.get")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := q.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionGetOrganization, q.ID); err != nil {
		return nil, err
	}

	value, err := s.repository.Get(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}

	return value, nil
}
