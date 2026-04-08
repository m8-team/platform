package boundaries

import (
	"context"
	"time"
)

type Workspace struct {
	ID             string
	OrganizationID string
	State          string
	Name           string
	Description    string
	CreateTime     time.Time
	UpdateTime     time.Time
	DeleteTime     *time.Time
	PurgeTime      *time.Time
	ETag           string
	Annotations    map[string]string
}

type CreateWorkspaceInput struct {
	Metadata       RequestMetadata
	OrganizationID string
	Name           string
	Description    string
	Annotations    map[string]string
}

type CreateWorkspaceOutput struct {
	Workspace Workspace
}

type UpdateWorkspaceInput struct {
	Metadata       RequestMetadata
	ID             string
	OrganizationID string
	ETag           string
	UpdateMask     []string
	Name           *string
	Description    *string
	Annotations    map[string]string
}

type UpdateWorkspaceOutput struct {
	Workspace Workspace
}

type DeleteWorkspaceInput struct {
	Metadata     RequestMetadata
	ID           string
	ETag         string
	AllowMissing bool
}

type DeleteWorkspaceOutput struct{}

type UndeleteWorkspaceInput struct {
	Metadata RequestMetadata
	ID       string
}

type UndeleteWorkspaceOutput struct {
	Workspace Workspace
}

type GetWorkspaceInput struct {
	ID string
}

type GetWorkspaceOutput struct {
	Workspace Workspace
}

type ListWorkspacesInput struct {
	OrganizationID string
	PageSize       int32
	PageToken      string
	Filter         string
	OrderBy        string
	ShowDeleted    bool
}

type ListWorkspacesOutput struct {
	Workspaces    []Workspace
	NextPageToken string
	TotalSize     int32
}

type WorkspaceCommandUseCase interface {
	CreateWorkspace(context.Context, CreateWorkspaceInput) (CreateWorkspaceOutput, error)
	UpdateWorkspace(context.Context, UpdateWorkspaceInput) (UpdateWorkspaceOutput, error)
	DeleteWorkspace(context.Context, DeleteWorkspaceInput) (DeleteWorkspaceOutput, error)
	UndeleteWorkspace(context.Context, UndeleteWorkspaceInput) (UndeleteWorkspaceOutput, error)
}

type WorkspaceQueryUseCase interface {
	GetWorkspace(context.Context, GetWorkspaceInput) (GetWorkspaceOutput, error)
	ListWorkspaces(context.Context, ListWorkspacesInput) (ListWorkspacesOutput, error)
}
