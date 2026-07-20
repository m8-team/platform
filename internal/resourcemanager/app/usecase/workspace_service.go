package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/app/command"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/query"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

var (
	ErrWorkspaceRepositoryRequired  = errors.New("workspace repository is required")
	ErrOrganizationLookupRequired   = errors.New("organization lookup is required")
	ErrWorkspaceAuthorizerRequired  = errors.New("workspace authorizer is required")
	ErrWorkspaceClockRequired       = errors.New("workspace clock is required")
	ErrWorkspaceIDGeneratorRequired = errors.New("workspace id generator is required")
	ErrGeneratedWorkspaceID         = errors.New("generated workspace id is invalid")
	ErrWorkspaceAuthorizationScope  = errors.New("workspace authorization scope key is required")
)

type WorkspaceServiceConfig struct {
	SoftDeleteRetention time.Duration
	PageTokenKey        []byte
}

// WorkspaceService coordinates workspace use cases. The domain aggregate and
// persistence port are workspace-owned; the parent Organization is accessed
// only through the explicit lookup port.
type WorkspaceService struct {
	repository    ports.WorkspaceRepository
	organizations ports.OrganizationLookup
	authorizer    ports.Authorizer
	clock         ports.Clock
	idGenerator   ports.WorkspaceIDGenerator
	retention     time.Duration
	pageTokens    workspacePageTokenCodec
}

func NewWorkspaceService(
	repository ports.WorkspaceRepository,
	organizations ports.OrganizationLookup,
	authorizer ports.Authorizer,
	clock ports.Clock,
	idGenerator ports.WorkspaceIDGenerator,
	config WorkspaceServiceConfig,
) (*WorkspaceService, error) {
	if repository == nil {
		return nil, ErrWorkspaceRepositoryRequired
	}
	if organizations == nil {
		return nil, ErrOrganizationLookupRequired
	}
	if authorizer == nil {
		return nil, ErrWorkspaceAuthorizerRequired
	}
	if clock == nil {
		return nil, ErrWorkspaceClockRequired
	}
	if idGenerator == nil {
		return nil, ErrWorkspaceIDGeneratorRequired
	}
	if config.SoftDeleteRetention <= 0 {
		return nil, ErrInvalidSoftDeleteRetention
	}
	if len(config.PageTokenKey) < minimumPageTokenKeyLength {
		return nil, ErrInvalidPageTokenKey
	}
	if workspaceFilterParserError != nil {
		return nil, fmt.Errorf("initialize workspace filter parser: %w", workspaceFilterParserError)
	}

	return &WorkspaceService{
		repository:    repository,
		organizations: organizations,
		authorizer:    authorizer,
		clock:         clock,
		idGenerator:   idGenerator,
		retention:     config.SoftDeleteRetention,
		pageTokens:    newWorkspacePageTokenCodec(config.PageTokenKey),
	}, nil
}

func (s *WorkspaceService) Create(ctx context.Context, cmd command.CreateWorkspace) (*workspace.Workspace, error) {
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

func (s *WorkspaceService) Get(ctx context.Context, q query.GetWorkspace) (*workspace.Workspace, error) {
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

func (s *WorkspaceService) List(ctx context.Context, q query.ListWorkspaces) (query.ListWorkspacesResult, error) {
	if err := s.authorize(ctx, ports.ActionListWorkspaces, q.OrganizationID, workspace.ID{}); err != nil {
		return query.ListWorkspacesResult{}, err
	}
	if !q.OrganizationID.IsZero() {
		if err := q.OrganizationID.Validate(); err != nil {
			return query.ListWorkspacesResult{}, err
		}
		if _, err := s.organizations.Get(ctx, q.OrganizationID); err != nil {
			return query.ListWorkspacesResult{}, fmt.Errorf("get parent organization for list: %w", err)
		}
	}
	authorizationScope, err := s.authorizer.ScopeKey(ctx)
	if err != nil {
		return query.ListWorkspacesResult{}, fmt.Errorf("load authorization scope: %w", err)
	}
	if strings.TrimSpace(authorizationScope) == "" {
		return query.ListWorkspacesResult{}, ErrWorkspaceAuthorizationScope
	}

	options, requestHash, err := normalizeWorkspaceListQuery(q, authorizationScope)
	if err != nil {
		return query.ListWorkspacesResult{}, err
	}
	if q.PageToken != "" {
		cursor, err := s.pageTokens.decode(q.PageToken, requestHash)
		if err != nil {
			return query.ListWorkspacesResult{}, err
		}
		options.After = cursor
	}

	result, err := s.repository.List(ctx, options)
	if err != nil {
		return query.ListWorkspacesResult{}, fmt.Errorf("list workspaces: %w", err)
	}
	nextPageToken := ""
	if result.Next != nil {
		nextPageToken, err = s.pageTokens.encode(*result.Next, requestHash)
		if err != nil {
			return query.ListWorkspacesResult{}, fmt.Errorf("encode workspace page token: %w", err)
		}
	}
	return query.ListWorkspacesResult{
		Workspaces:    result.Workspaces,
		NextPageToken: nextPageToken,
		TotalSize:     result.TotalSize,
	}, nil
}

func (s *WorkspaceService) Update(ctx context.Context, cmd command.UpdateWorkspace) (*workspace.Workspace, error) {
	if err := cmd.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionUpdateWorkspace, organization.ID{}, cmd.ID); err != nil {
		return nil, err
	}
	value, err := s.repository.Get(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("get workspace for update: %w", err)
	}
	expectedStoredVersion := value.Version()
	if err := value.Update(workspace.UpdateParams{
		Name:            cmd.Name,
		Description:     cmd.Description,
		Labels:          cmd.Labels,
		Now:             s.clock.Now().UTC(),
		ExpectedVersion: cmd.ExpectedVersion,
	}); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, value, expectedStoredVersion); err != nil {
		return nil, fmt.Errorf("update workspace: %w", err)
	}
	return value.Clone(), nil
}

func (s *WorkspaceService) Delete(ctx context.Context, cmd command.DeleteWorkspace) (*workspace.Workspace, error) {
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

func (s *WorkspaceService) Undelete(ctx context.Context, cmd command.UndeleteWorkspace) (*workspace.Workspace, error) {
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
			ExpectedVersion: expectedStoredVersion,
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
