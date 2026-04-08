package grpcpresenter

import (
	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

type ProjectPresenter struct{}

func (ProjectPresenter) PresentGet(output boundaries.GetProjectOutput) *resourcemanagerv1.Project {
	return mapProject(output.Project)
}

func (ProjectPresenter) PresentCreate(output boundaries.CreateProjectOutput) *resourcemanagerv1.Project {
	return mapProject(output.Project)
}

func (ProjectPresenter) PresentUpdate(output boundaries.UpdateProjectOutput) *resourcemanagerv1.Project {
	return mapProject(output.Project)
}

func (ProjectPresenter) PresentUndelete(output boundaries.UndeleteProjectOutput) *resourcemanagerv1.Project {
	return mapProject(output.Project)
}

func (ProjectPresenter) PresentList(output boundaries.ListProjectsOutput) *resourcemanagerv1.ListProjectsResponse {
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
