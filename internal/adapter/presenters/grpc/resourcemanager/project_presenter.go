package grpcpresenter

import (
	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
)

type ProjectPresenter struct{}

func (ProjectPresenter) PresentGet(output boundary.GetProjectOutput) *resourcemanagerv1.Project {
	return mapProject(output.Project)
}

func (ProjectPresenter) PresentCreate(output boundary.CreateProjectOutput) *resourcemanagerv1.Project {
	return mapProject(output.Project)
}

func (ProjectPresenter) PresentUpdate(output boundary.UpdateProjectOutput) *resourcemanagerv1.Project {
	return mapProject(output.Project)
}

func (ProjectPresenter) PresentUndelete(output boundary.UndeleteProjectOutput) *resourcemanagerv1.Project {
	return mapProject(output.Project)
}

func (ProjectPresenter) PresentList(output boundary.ListProjectsOutput) *resourcemanagerv1.ListProjectsResponse {
	items := make([]*resourcemanagerv1.Project, 0, len(output.Projects))
	for _, item := range output.Projects {
		items = append(items, mapProject(item))
	}
	return &resourcemanagerv1.ListProjectsResponse{
		Projects:      items,
		NextPageToken: output.NextPageToken,
		TotalSize:     output.TotalSize,
	}
}
