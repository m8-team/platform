package ports

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

var (
	ErrWorkspaceNotFound        = errors.New("workspace not found")
	ErrWorkspaceAlreadyExists   = errors.New("workspace already exists")
	ErrWorkspaceVersionConflict = errors.New("workspace version conflict")
)

type WorkspaceRepository interface {
	Create(context.Context, *workspace.Workspace) error
	Get(context.Context, workspace.ID) (*workspace.Workspace, error)
	Update(context.Context, *workspace.Workspace, types.Version) error
	List(context.Context, organization.ID, bool) ([]*workspace.Workspace, error)
}

type OrganizationLookup interface {
	Get(context.Context, organization.ID) (*organization.Organization, error)
}
