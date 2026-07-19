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
)

const DefaultSoftDeleteRetention = 30 * 24 * time.Hour

var (
	ErrOrganizationRepositoryRequired  = errors.New("organization repository is required")
	ErrOrganizationAuthorizerRequired  = errors.New("organization authorizer is required")
	ErrOrganizationClockRequired       = errors.New("organization clock is required")
	ErrOrganizationIDGeneratorRequired = errors.New("organization id generator is required")
	ErrWorkspaceChildrenRequired       = errors.New("workspace children reader is required")
	ErrInvalidSoftDeleteRetention      = errors.New("soft-delete retention must be positive")
	ErrInvalidPageTokenKey             = errors.New("page token key must contain at least 32 bytes")
	ErrOrganizationHasWorkspaces       = errors.New("organization has non-deleted workspaces")
	ErrGeneratedOrganizationID         = errors.New("generated organization id is invalid")
	ErrAuthorizationScopeRequired      = errors.New("authorization scope key is required")
)

type OrganizationServiceConfig struct {
	SoftDeleteRetention time.Duration
	PageTokenKey        []byte
}

// OrganizationService coordinates organization use cases. Persistence,
// authorization, time, identity generation, and hierarchy checks are explicit
// ports so business behavior remains independent from gRPC and storage.
type OrganizationService struct {
	repository        ports.OrganizationRepository
	authorizer        ports.Authorizer
	clock             ports.Clock
	idGenerator       ports.IDGenerator
	workspaceChildren ports.WorkspaceChildren
	retention         time.Duration
	pageTokens        pageTokenCodec
}

func NewOrganizationService(
	repository ports.OrganizationRepository,
	authorizer ports.Authorizer,
	clock ports.Clock,
	idGenerator ports.IDGenerator,
	workspaceChildren ports.WorkspaceChildren,
	config OrganizationServiceConfig,
) (*OrganizationService, error) {
	if repository == nil {
		return nil, ErrOrganizationRepositoryRequired
	}
	if authorizer == nil {
		return nil, ErrOrganizationAuthorizerRequired
	}
	if clock == nil {
		return nil, ErrOrganizationClockRequired
	}
	if idGenerator == nil {
		return nil, ErrOrganizationIDGeneratorRequired
	}
	if workspaceChildren == nil {
		return nil, ErrWorkspaceChildrenRequired
	}
	if config.SoftDeleteRetention <= 0 {
		return nil, ErrInvalidSoftDeleteRetention
	}
	if len(config.PageTokenKey) < minimumPageTokenKeyLength {
		return nil, ErrInvalidPageTokenKey
	}

	return &OrganizationService{
		repository:        repository,
		authorizer:        authorizer,
		clock:             clock,
		idGenerator:       idGenerator,
		workspaceChildren: workspaceChildren,
		retention:         config.SoftDeleteRetention,
		pageTokens:        newPageTokenCodec(config.PageTokenKey),
	}, nil
}

func (s *OrganizationService) Create(
	ctx context.Context,
	cmd command.CreateOrganization,
) (*organization.Organization, error) {
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

func (s *OrganizationService) Get(
	ctx context.Context,
	q query.GetOrganization,
) (*organization.Organization, error) {
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

func (s *OrganizationService) List(
	ctx context.Context,
	q query.ListOrganizations,
) (query.ListOrganizationsResult, error) {
	if err := s.authorize(ctx, ports.ActionListOrganizations, organization.ID{}); err != nil {
		return query.ListOrganizationsResult{}, err
	}
	authorizationScope, err := s.authorizer.ScopeKey(ctx)
	if err != nil {
		return query.ListOrganizationsResult{}, fmt.Errorf("load authorization scope: %w", err)
	}
	if strings.TrimSpace(authorizationScope) == "" {
		return query.ListOrganizationsResult{}, ErrAuthorizationScopeRequired
	}

	options, requestHash, err := normalizeListQuery(q, authorizationScope)
	if err != nil {
		return query.ListOrganizationsResult{}, err
	}
	if q.PageToken != "" {
		cursor, err := s.pageTokens.decode(q.PageToken, requestHash)
		if err != nil {
			return query.ListOrganizationsResult{}, err
		}
		options.After = cursor
	}

	result, err := s.repository.List(ctx, options)
	if err != nil {
		return query.ListOrganizationsResult{}, fmt.Errorf("list organizations: %w", err)
	}

	nextPageToken := ""
	if result.Next != nil {
		nextPageToken, err = s.pageTokens.encode(*result.Next, requestHash)
		if err != nil {
			return query.ListOrganizationsResult{}, fmt.Errorf("encode organization page token: %w", err)
		}
	}

	return query.ListOrganizationsResult{
		Organizations: result.Organizations,
		NextPageToken: nextPageToken,
		TotalSize:     result.TotalSize,
	}, nil
}

func (s *OrganizationService) Update(
	ctx context.Context,
	cmd command.UpdateOrganization,
) (*organization.Organization, error) {
	if err := cmd.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionUpdateOrganization, cmd.ID); err != nil {
		return nil, err
	}

	value, err := s.repository.Get(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("get organization for update: %w", err)
	}
	expectedStoredVersion := value.Version()
	if err := value.Update(organization.UpdateParams{
		Name:            cmd.Name,
		Description:     cmd.Description,
		Labels:          cmd.Labels,
		Now:             s.clock.Now().UTC(),
		ExpectedVersion: cmd.ExpectedVersion,
	}); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, value, expectedStoredVersion); err != nil {
		return nil, fmt.Errorf("update organization: %w", err)
	}

	return value.Clone(), nil
}

func (s *OrganizationService) Delete(
	ctx context.Context,
	cmd command.DeleteOrganization,
) (*organization.Organization, error) {
	if err := cmd.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionDeleteOrganization, cmd.ID); err != nil {
		return nil, err
	}

	value, err := s.repository.Get(ctx, cmd.ID)
	if err != nil {
		if cmd.AllowMissing && errors.Is(err, ports.ErrOrganizationNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get organization for delete: %w", err)
	}
	if value.IsDeleted() {
		if cmd.AllowMissing {
			return value, nil
		}
		return nil, organization.ErrOrganizationAlreadyDeleted
	}
	if err := value.CheckVersion(cmd.ExpectedVersion); err != nil {
		return nil, err
	}

	hasChildren, err := s.workspaceChildren.HasNonDeleted(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("check organization workspaces: %w", err)
	}
	if hasChildren {
		return nil, ErrOrganizationHasWorkspaces
	}

	expectedStoredVersion := value.Version()
	now := s.clock.Now().UTC()
	if err := value.Delete(organization.DeleteParams{
		Now:             now,
		PurgeTime:       now.Add(s.retention),
		ExpectedVersion: cmd.ExpectedVersion,
	}); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, value, expectedStoredVersion); err != nil {
		return nil, fmt.Errorf("delete organization: %w", err)
	}

	return value.Clone(), nil
}

func (s *OrganizationService) Undelete(
	ctx context.Context,
	cmd command.UndeleteOrganization,
) (*organization.Organization, error) {
	if err := cmd.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ports.ActionUndeleteOrganization, cmd.ID); err != nil {
		return nil, err
	}

	value, err := s.repository.Get(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("get organization for undelete: %w", err)
	}
	expectedStoredVersion := value.Version()
	if err := value.Undelete(organization.UndeleteParams{
		Now:             s.clock.Now().UTC(),
		ExpectedVersion: expectedStoredVersion,
	}); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, value, expectedStoredVersion); err != nil {
		return nil, fmt.Errorf("undelete organization: %w", err)
	}

	return value.Clone(), nil
}

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
