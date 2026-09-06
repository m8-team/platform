package workspaceapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func (s *WorkspaceService) List(ctx context.Context, q ListWorkspaces) (out ListWorkspacesResult, resultErr error) {
	ctx, finish := s.observer.Start(ctx, "resourcemanager.workspace.list")
	defer func() { finish(resultErr) }()
	if err := ctx.Err(); err != nil {
		return ListWorkspacesResult{}, err
	}

	if err := s.authorize(ctx, ports.ActionListWorkspaces, q.OrganizationID, workspace.ID{}); err != nil {
		return ListWorkspacesResult{}, err
	}
	if !q.OrganizationID.IsZero() {
		if err := q.OrganizationID.Validate(); err != nil {
			return ListWorkspacesResult{}, err
		}
		if _, err := s.organizations.Get(ctx, q.OrganizationID); err != nil {
			return ListWorkspacesResult{}, fmt.Errorf("get parent organization for list: %w", err)
		}
	}
	authorizationScope, err := s.authorizer.ScopeKey(ctx)
	if err != nil {
		return ListWorkspacesResult{}, fmt.Errorf("load authorization scope: %w", err)
	}
	if strings.TrimSpace(authorizationScope) == "" {
		return ListWorkspacesResult{}, ErrWorkspaceAuthorizationScope
	}

	options, requestHash, err := normalizeWorkspaceListQuery(s.filterParser, q, authorizationScope)
	if err != nil {
		return ListWorkspacesResult{}, err
	}
	if q.PageToken != "" {
		cursor, err := s.pageTokens.decode(q.PageToken, requestHash)
		if err != nil {
			return ListWorkspacesResult{}, err
		}
		options.After = cursor
	}

	result, err := s.repository.List(ctx, options)
	if err != nil {
		return ListWorkspacesResult{}, fmt.Errorf("list workspaces: %w", err)
	}
	nextPageToken := ""
	if result.Next != nil {
		nextPageToken, err = s.pageTokens.encode(*result.Next, requestHash)
		if err != nil {
			return ListWorkspacesResult{}, fmt.Errorf("encode workspace page token: %w", err)
		}
	}
	return ListWorkspacesResult{
		Workspaces:    result.Workspaces,
		NextPageToken: nextPageToken,
		TotalSize:     result.TotalSize,
	}, nil
}
