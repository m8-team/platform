package query

import (
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

type GetWorkspace struct{ ID workspace.ID }
type ListWorkspaces struct {
	OrganizationID             organization.ID
	PageSize                   int
	PageToken, Filter, OrderBy string
	ShowDeleted                bool
}
type ListWorkspacesResult struct {
	Workspaces    []*workspace.Workspace
	NextPageToken string
	TotalSize     int
}
