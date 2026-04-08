package boundary

import (
	"context"
	"time"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/requestmeta"
)

type Project struct {
	ID          string
	WorkspaceID string
	State       string
	Name        string
	Description string
	CreateTime  time.Time
	UpdateTime  time.Time
	DeleteTime  *time.Time
	PurgeTime   *time.Time
	ETag        string
	Annotations map[string]string
}

type CreateProjectInput struct {
	Metadata    requestmeta.RequestMetadata
	WorkspaceID string
	Name        string
	Description string
	Annotations map[string]string
}

type CreateProjectOutput struct {
	Project Project
}

type UpdateProjectInput struct {
	Metadata    requestmeta.RequestMetadata
	ID          string
	WorkspaceID string
	ETag        string
	UpdateMask  []string
	Name        *string
	Description *string
	Annotations map[string]string
}

type UpdateProjectOutput struct {
	Project Project
}

type DeleteProjectInput struct {
	Metadata     requestmeta.RequestMetadata
	ID           string
	ETag         string
	AllowMissing bool
}

type DeleteProjectOutput struct{}

type UndeleteProjectInput struct {
	Metadata requestmeta.RequestMetadata
	ID       string
}

type UndeleteProjectOutput struct {
	Project Project
}

type GetProjectInput struct {
	ID string
}

type GetProjectOutput struct {
	Project Project
}

type ListProjectsInput struct {
	WorkspaceID string
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type ListProjectsOutput struct {
	Projects      []Project
	NextPageToken string
	TotalSize     int32
}

type ProjectCommandUseCase interface {
	CreateProject(context.Context, CreateProjectInput) (CreateProjectOutput, error)
	UpdateProject(context.Context, UpdateProjectInput) (UpdateProjectOutput, error)
	DeleteProject(context.Context, DeleteProjectInput) (DeleteProjectOutput, error)
	UndeleteProject(context.Context, UndeleteProjectInput) (UndeleteProjectOutput, error)
}

type ProjectQueryUseCase interface {
	GetProject(context.Context, GetProjectInput) (GetProjectOutput, error)
	ListProjects(context.Context, ListProjectsInput) (ListProjectsOutput, error)
}
