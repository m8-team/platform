package command

import (
	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

type CreateWorkspace struct {
	OrganizationID    organization.ID
	Name, Description string
	Labels            map[string]string
}
type UpdateWorkspace struct {
	ID                workspace.ID
	ExpectedVersion   types.Version
	Name, Description *string
	Labels            *map[string]string
}
type DeleteWorkspace struct {
	ID              workspace.ID
	ExpectedVersion types.Version
	AllowMissing    bool
}
type UndeleteWorkspace struct{ ID workspace.ID }
