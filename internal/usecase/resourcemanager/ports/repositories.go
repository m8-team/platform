package ports

import (
	"context"

	"github.com/m8platform/platform/internal/entities/resourcemanager/organization"
	"github.com/m8platform/platform/internal/entities/resourcemanager/project"
	"github.com/m8platform/platform/internal/entities/resourcemanager/workspace"
)

type HierarchyNode struct {
	ID      string
	Exists  bool
	Deleted bool
}

type OrganizationListParams struct {
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type WorkspaceListParams struct {
	OrganizationID string
	PageSize       int32
	PageToken      string
	Filter         string
	OrderBy        string
	ShowDeleted    bool
}

type ProjectListParams struct {
	WorkspaceID string
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type OrganizationPage struct {
	Items         []organization.Entity
	NextPageToken string
	TotalSize     int32
}

type WorkspacePage struct {
	Items         []workspace.Entity
	NextPageToken string
	TotalSize     int32
}

type ProjectPage struct {
	Items         []project.Entity
	NextPageToken string
	TotalSize     int32
}

type OrganizationRepository interface {
	GetByID(ctx context.Context, id string, includeDeleted bool) (organization.Entity, error)
	Create(ctx context.Context, entity organization.Entity) error
	Update(ctx context.Context, entity organization.Entity) error
	List(ctx context.Context, params OrganizationListParams) (OrganizationPage, error)
}

type WorkspaceRepository interface {
	GetByID(ctx context.Context, id string, includeDeleted bool) (workspace.Entity, error)
	Create(ctx context.Context, entity workspace.Entity) error
	Update(ctx context.Context, entity workspace.Entity) error
	List(ctx context.Context, params WorkspaceListParams) (WorkspacePage, error)
}

type ProjectRepository interface {
	GetByID(ctx context.Context, id string, includeDeleted bool) (project.Entity, error)
	Create(ctx context.Context, entity project.Entity) error
	Update(ctx context.Context, entity project.Entity) error
	List(ctx context.Context, params ProjectListParams) (ProjectPage, error)
}

type HierarchyReader interface {
	GetOrganizationNode(ctx context.Context, id string) (HierarchyNode, error)
	GetWorkspaceNode(ctx context.Context, id string) (HierarchyNode, error)
	HasActiveWorkspaces(ctx context.Context, organizationID string) (bool, error)
	HasActiveProjects(ctx context.Context, workspaceID string) (bool, error)
}
