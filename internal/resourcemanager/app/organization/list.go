package organizationapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func (s *OrganizationService) List(
	ctx context.Context,
	q ListOrganizations,
) (out ListOrganizationsResult, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.organization.list")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return ListOrganizationsResult{}, err
	}

	if err := s.authorize(ctx, ports.ActionListOrganizations, organization.ID{}); err != nil {
		return ListOrganizationsResult{}, err
	}
	authorizationScope, err := s.authorizer.ScopeKey(ctx)
	if err != nil {
		return ListOrganizationsResult{}, fmt.Errorf("load authorization scope: %w", err)
	}
	if strings.TrimSpace(authorizationScope) == "" {
		return ListOrganizationsResult{}, ErrAuthorizationScopeRequired
	}

	options, requestHash, err := normalizeListQuery(s.filterParser, q, authorizationScope)
	if err != nil {
		return ListOrganizationsResult{}, err
	}
	if q.PageToken != "" {
		cursor, err := s.pageTokens.decode(q.PageToken, requestHash)
		if err != nil {
			return ListOrganizationsResult{}, err
		}
		options.After = cursor
	}

	result, err := s.repository.List(ctx, options)
	if err != nil {
		return ListOrganizationsResult{}, fmt.Errorf("list organizations: %w", err)
	}

	nextPageToken := ""
	if result.Next != nil {
		nextPageToken, err = s.pageTokens.encode(*result.Next, requestHash)
		if err != nil {
			return ListOrganizationsResult{}, fmt.Errorf("encode organization page token: %w", err)
		}
	}

	return ListOrganizationsResult{
		Organizations: result.Organizations,
		NextPageToken: nextPageToken,
		TotalSize:     result.TotalSize,
	}, nil
}
