package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/domain/project"
	"github.com/m8platform/platform/internal/domain/workspace"
	"github.com/m8platform/platform/internal/ports"
)

type TxManager struct{}

func (TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type Clock struct {
	Value time.Time
}

func (c Clock) Now() time.Time {
	if c.Value.IsZero() {
		return time.Unix(1_700_000_000, 0).UTC()
	}
	return c.Value.UTC()
}

type UUIDGenerator struct {
	mu     sync.Mutex
	Values []string
	index  int
}

func (g *UUIDGenerator) NewString() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.index >= len(g.Values) {
		return "00000000-0000-0000-0000-000000000000"
	}
	value := g.Values[g.index]
	g.index++
	return value
}

type IdempotencyStore struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{seen: make(map[string]struct{})}
}

func (s *IdempotencyStore) Reserve(_ context.Context, scope string, key string, _ time.Duration) (ports.IdempotencyReservation, error) {
	if key == "" {
		return ports.IdempotencyReservation{}, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	compound := scope + ":" + key
	if _, ok := s.seen[compound]; ok {
		return ports.IdempotencyReservation{Scope: scope, Key: key, Duplicate: true}, nil
	}
	s.seen[compound] = struct{}{}
	return ports.IdempotencyReservation{Scope: scope, Key: key}, nil
}

func (s *IdempotencyStore) MarkCompleted(context.Context, ports.IdempotencyReservation) error {
	return nil
}

type OutboxWriter struct {
	mu      sync.Mutex
	Records []ports.OutboxRecord
}

func (o *OutboxWriter) Append(_ context.Context, record ports.OutboxRecord) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.Records = append(o.Records, record)
	return nil
}

type OrganizationRepository struct {
	Items map[string]organization.Organization
}

func NewOrganizationRepository() *OrganizationRepository {
	return &OrganizationRepository{Items: make(map[string]organization.Organization)}
}

func (r *OrganizationRepository) GetByID(_ context.Context, id string, includeDeleted bool) (organization.Organization, error) {
	item, ok := r.Items[id]
	if !ok {
		return organization.Organization{}, organization.ErrNotFound
	}
	if item.IsDeleted() && !includeDeleted {
		return organization.Organization{}, organization.ErrNotFound
	}
	return item, nil
}

func (r *OrganizationRepository) Create(_ context.Context, aggregate organization.Organization) error {
	r.Items[aggregate.ID] = aggregate
	return nil
}

func (r *OrganizationRepository) Update(_ context.Context, aggregate organization.Organization) error {
	r.Items[aggregate.ID] = aggregate
	return nil
}

func (r *OrganizationRepository) List(_ context.Context, _ organization.ListParams) (organization.Page, error) {
	items := make([]organization.Organization, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, item)
	}
	return organization.Page{Items: items, TotalSize: int32(len(items))}, nil
}

type WorkspaceRepository struct {
	Items map[string]workspace.Workspace
}

func NewWorkspaceRepository() *WorkspaceRepository {
	return &WorkspaceRepository{Items: make(map[string]workspace.Workspace)}
}

func (r *WorkspaceRepository) GetByID(_ context.Context, id string, includeDeleted bool) (workspace.Workspace, error) {
	item, ok := r.Items[id]
	if !ok {
		return workspace.Workspace{}, workspace.ErrNotFound
	}
	if item.IsDeleted() && !includeDeleted {
		return workspace.Workspace{}, workspace.ErrNotFound
	}
	return item, nil
}

func (r *WorkspaceRepository) Create(_ context.Context, aggregate workspace.Workspace) error {
	r.Items[aggregate.ID] = aggregate
	return nil
}

func (r *WorkspaceRepository) Update(_ context.Context, aggregate workspace.Workspace) error {
	r.Items[aggregate.ID] = aggregate
	return nil
}

func (r *WorkspaceRepository) List(_ context.Context, params workspace.ListParams) (workspace.Page, error) {
	items := make([]workspace.Workspace, 0)
	for _, item := range r.Items {
		if params.OrganizationID != "" && item.OrganizationID != params.OrganizationID {
			continue
		}
		if item.IsDeleted() && !params.ShowDeleted {
			continue
		}
		items = append(items, item)
	}
	return workspace.Page{Items: items, TotalSize: int32(len(items))}, nil
}

type ProjectRepository struct {
	Items map[string]project.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{Items: make(map[string]project.Project)}
}

func (r *ProjectRepository) GetByID(_ context.Context, id string, includeDeleted bool) (project.Project, error) {
	item, ok := r.Items[id]
	if !ok {
		return project.Project{}, project.ErrNotFound
	}
	if item.IsDeleted() && !includeDeleted {
		return project.Project{}, project.ErrNotFound
	}
	return item, nil
}

func (r *ProjectRepository) Create(_ context.Context, aggregate project.Project) error {
	r.Items[aggregate.ID] = aggregate
	return nil
}

func (r *ProjectRepository) Update(_ context.Context, aggregate project.Project) error {
	r.Items[aggregate.ID] = aggregate
	return nil
}

func (r *ProjectRepository) List(_ context.Context, params project.ListParams) (project.Page, error) {
	items := make([]project.Project, 0)
	for _, item := range r.Items {
		if params.WorkspaceID != "" && item.WorkspaceID != params.WorkspaceID {
			continue
		}
		if item.IsDeleted() && !params.ShowDeleted {
			continue
		}
		items = append(items, item)
	}
	return project.Page{Items: items, TotalSize: int32(len(items))}, nil
}

type HierarchyRepository struct {
	Organizations    map[string]ports.HierarchyNode
	Workspaces       map[string]ports.HierarchyNode
	ActiveWorkspaces map[string]bool
	ActiveProjects   map[string]bool
}

func NewHierarchyRepository() *HierarchyRepository {
	return &HierarchyRepository{
		Organizations:    make(map[string]ports.HierarchyNode),
		Workspaces:       make(map[string]ports.HierarchyNode),
		ActiveWorkspaces: make(map[string]bool),
		ActiveProjects:   make(map[string]bool),
	}
}

func (r *HierarchyRepository) GetOrganizationNode(_ context.Context, id string) (ports.HierarchyNode, error) {
	node, ok := r.Organizations[id]
	if !ok {
		return ports.HierarchyNode{ID: id}, nil
	}
	return node, nil
}

func (r *HierarchyRepository) GetWorkspaceNode(_ context.Context, id string) (ports.HierarchyNode, error) {
	node, ok := r.Workspaces[id]
	if !ok {
		return ports.HierarchyNode{ID: id}, nil
	}
	return node, nil
}

func (r *HierarchyRepository) HasActiveWorkspaces(_ context.Context, organizationID string) (bool, error) {
	return r.ActiveWorkspaces[organizationID], nil
}

func (r *HierarchyRepository) HasActiveProjects(_ context.Context, workspaceID string) (bool, error) {
	return r.ActiveProjects[workspaceID], nil
}
