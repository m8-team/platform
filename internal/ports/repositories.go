package ports

import (
	"context"

	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/domain/project"
	"github.com/m8platform/platform/internal/domain/workspace"
)

type OrganizationRepository = organization.Repository
type WorkspaceRepository = workspace.Repository
type ProjectRepository = project.Repository

// HierarchyNode exposes only the parent-state facts required for cross-aggregate
// policies.
type HierarchyNode struct {
	ID      string
	Exists  bool
	Deleted bool
}

type HierarchyRepository interface {
	GetOrganizationNode(ctx context.Context, id string) (HierarchyNode, error)
	GetWorkspaceNode(ctx context.Context, id string) (HierarchyNode, error)
	HasActiveWorkspaces(ctx context.Context, organizationID string) (bool, error)
	HasActiveProjects(ctx context.Context, workspaceID string) (bool, error)
}
