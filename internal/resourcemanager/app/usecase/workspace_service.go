package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/app/command"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/query"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

var ErrParentOrganizationDeleted = errors.New("parent organization is deleted")

type WorkspaceService struct {
	repository    ports.WorkspaceRepository
	organizations ports.OrganizationLookup
	authorizer    ports.Authorizer
	clock         ports.Clock
	ids           ports.IDGenerator
	retention     time.Duration
}

func NewWorkspaceService(repository ports.WorkspaceRepository, organizations ports.OrganizationLookup, authorizer ports.Authorizer, clock ports.Clock, ids ports.IDGenerator, cfg OrganizationServiceConfig) (*WorkspaceService, error) {
	if repository == nil || organizations == nil || authorizer == nil || clock == nil || ids == nil {
		return nil, errors.New("workspace service dependencies are required")
	}
	if cfg.SoftDeleteRetention <= 0 {
		return nil, ErrInvalidSoftDeleteRetention
	}
	return &WorkspaceService{repository: repository, organizations: organizations, authorizer: authorizer, clock: clock, ids: ids, retention: cfg.SoftDeleteRetention}, nil
}
func (s *WorkspaceService) auth(ctx context.Context, action ports.AuthorizationAction, org organization.ID, id workspace.ID) error {
	if err := s.authorizer.Authorize(ctx, ports.AuthorizationRequest{Action: action, OrganizationID: org, WorkspaceID: id}); err != nil {
		return fmt.Errorf("authorize %s: %w", action, err)
	}
	return nil
}
func (s *WorkspaceService) Create(ctx context.Context, c command.CreateWorkspace) (*workspace.Workspace, error) {
	if err := c.OrganizationID.Validate(); err != nil {
		return nil, err
	}
	if err := s.auth(ctx, ports.ActionCreateWorkspace, c.OrganizationID, workspace.ID{}); err != nil {
		return nil, err
	}
	parent, err := s.organizations.Get(ctx, c.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("get parent organization: %w", err)
	}
	if parent.IsDeleted() {
		return nil, ErrParentOrganizationDeleted
	}
	w, err := workspace.New(workspace.CreateParams{ID: s.ids.NewID(), OrganizationID: c.OrganizationID, Name: c.Name, Description: c.Description, Labels: c.Labels, Now: s.clock.Now().UTC()})
	if err != nil {
		return nil, err
	}
	if err = s.repository.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	return w.Clone(), nil
}
func (s *WorkspaceService) Get(ctx context.Context, q query.GetWorkspace) (*workspace.Workspace, error) {
	if err := q.ID.Validate(); err != nil {
		return nil, err
	}
	if err := s.auth(ctx, ports.ActionGetWorkspace, organization.ID{}, q.ID); err != nil {
		return nil, err
	}
	w, err := s.repository.Get(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get workspace: %w", err)
	}
	return w, nil
}
func (s *WorkspaceService) List(ctx context.Context, q query.ListWorkspaces) (query.ListWorkspacesResult, error) {
	if err := q.OrganizationID.Validate(); err != nil {
		return query.ListWorkspacesResult{}, err
	}
	if q.PageSize < 0 || q.PageSize > 1000 {
		return query.ListWorkspacesResult{}, ErrInvalidOrganizationPageSize
	}
	if q.Filter != "" {
		return query.ListWorkspacesResult{}, fmt.Errorf("%w: workspace filters are not supported yet", ErrInvalidOrganizationFilter)
	}
	if q.OrderBy != "" && q.OrderBy != "id" && q.OrderBy != "id asc" && q.OrderBy != "id desc" {
		return query.ListWorkspacesResult{}, ErrInvalidOrganizationOrderBy
	}
	if err := s.auth(ctx, ports.ActionListWorkspaces, q.OrganizationID, workspace.ID{}); err != nil {
		return query.ListWorkspacesResult{}, err
	}
	values, err := s.repository.List(ctx, q.OrganizationID, q.ShowDeleted)
	if err != nil {
		return query.ListWorkspacesResult{}, err
	}
	desc := strings.HasSuffix(q.OrderBy, "desc")
	sort.Slice(values, func(i, j int) bool {
		if desc {
			return values[i].ID().String() > values[j].ID().String()
		}
		return values[i].ID().String() < values[j].ID().String()
	})
	start := 0
	if q.PageToken != "" {
		for start < len(values) && values[start].ID().String() != q.PageToken {
			start++
		}
		if start == len(values) {
			return query.ListWorkspacesResult{}, ErrInvalidOrganizationPageToken
		}
		start++
	}
	size := q.PageSize
	if size == 0 {
		size = 50
	}
	end := min(start+size, len(values))
	token := ""
	if end < len(values) {
		token = values[end-1].ID().String()
	}
	return query.ListWorkspacesResult{Workspaces: values[start:end], NextPageToken: token, TotalSize: len(values)}, nil
}
func (s *WorkspaceService) Update(ctx context.Context, c command.UpdateWorkspace) (*workspace.Workspace, error) {
	w, err := s.repository.Get(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	if err = s.auth(ctx, ports.ActionUpdateWorkspace, w.OrganizationID(), c.ID); err != nil {
		return nil, err
	}
	v := w.Version()
	if err = w.Update(c.Name, c.Description, c.Labels, s.clock.Now().UTC(), c.ExpectedVersion); err != nil {
		return nil, err
	}
	if err = s.repository.Update(ctx, w, v); err != nil {
		return nil, err
	}
	return w.Clone(), nil
}
func (s *WorkspaceService) Delete(ctx context.Context, c command.DeleteWorkspace) (*workspace.Workspace, error) {
	w, err := s.repository.Get(ctx, c.ID)
	if err != nil {
		if c.AllowMissing && errors.Is(err, ports.ErrWorkspaceNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if err = s.auth(ctx, ports.ActionDeleteWorkspace, w.OrganizationID(), c.ID); err != nil {
		return nil, err
	}
	if w.IsDeleted() {
		if c.AllowMissing {
			return w, nil
		}
		return nil, organization.ErrOrganizationAlreadyDeleted
	}
	v := w.Version()
	now := s.clock.Now().UTC()
	if err = w.Delete(now, now.Add(s.retention), c.ExpectedVersion); err != nil {
		return nil, err
	}
	if err = s.repository.Update(ctx, w, v); err != nil {
		return nil, err
	}
	return w.Clone(), nil
}
func (s *WorkspaceService) Undelete(ctx context.Context, c command.UndeleteWorkspace) (*workspace.Workspace, error) {
	w, err := s.repository.Get(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	if err = s.auth(ctx, ports.ActionUndeleteWorkspace, w.OrganizationID(), c.ID); err != nil {
		return nil, err
	}
	v := w.Version()
	if err = w.Undelete(s.clock.Now().UTC()); err != nil {
		return nil, err
	}
	if err = s.repository.Update(ctx, w, v); err != nil {
		return nil, err
	}
	return w.Clone(), nil
}
