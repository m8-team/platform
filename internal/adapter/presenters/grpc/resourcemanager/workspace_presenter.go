package grpcpresenter

import (
	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
)

type WorkspacePresenter struct{}

func (WorkspacePresenter) PresentGet(output boundary.GetWorkspaceOutput) *resourcemanagerv1.Workspace {
	return mapWorkspace(output.Workspace)
}

func (WorkspacePresenter) PresentCreate(output boundary.CreateWorkspaceOutput) *resourcemanagerv1.Workspace {
	return mapWorkspace(output.Workspace)
}

func (WorkspacePresenter) PresentUpdate(output boundary.UpdateWorkspaceOutput) *resourcemanagerv1.Workspace {
	return mapWorkspace(output.Workspace)
}

func (WorkspacePresenter) PresentUndelete(output boundary.UndeleteWorkspaceOutput) *resourcemanagerv1.Workspace {
	return mapWorkspace(output.Workspace)
}

func (WorkspacePresenter) PresentList(output boundary.ListWorkspacesOutput) *resourcemanagerv1.ListWorkspacesResponse {
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
