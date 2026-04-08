package grpcpresenter

import (
	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

type WorkspacePresenter struct{}

func (WorkspacePresenter) PresentGet(output boundaries.GetWorkspaceOutput) *resourcemanagerv1.Workspace {
	return mapWorkspace(output.Workspace)
}

func (WorkspacePresenter) PresentCreate(output boundaries.CreateWorkspaceOutput) *resourcemanagerv1.Workspace {
	return mapWorkspace(output.Workspace)
}

func (WorkspacePresenter) PresentUpdate(output boundaries.UpdateWorkspaceOutput) *resourcemanagerv1.Workspace {
	return mapWorkspace(output.Workspace)
}

func (WorkspacePresenter) PresentUndelete(output boundaries.UndeleteWorkspaceOutput) *resourcemanagerv1.Workspace {
	return mapWorkspace(output.Workspace)
}

func (WorkspacePresenter) PresentList(output boundaries.ListWorkspacesOutput) *resourcemanagerv1.ListWorkspacesResponse {
	items := make([]*resourcemanagerv1.Workspace, 0, len(output.Workspaces))
	for _, item := range output.Workspaces {
		items = append(items, mapWorkspace(item))
	}
	return &resourcemanagerv1.ListWorkspacesResponse{
		Workspaces:    items,
		NextPageToken: output.NextPageToken,
		TotalSize:     output.TotalSize,
	}
}
