package workspace

import "context"

type ListParams struct {
	OrganizationID string
	PageSize       int32
	PageToken      string
	Filter         string
	OrderBy        string
	ShowDeleted    bool
}

type Page struct {
	Items         []Workspace
	NextPageToken string
	TotalSize     int32
}

type Repository interface {
	GetByID(ctx context.Context, id string, includeDeleted bool) (Workspace, error)
	Create(ctx context.Context, aggregate Workspace) error
	Update(ctx context.Context, aggregate Workspace) error
	List(ctx context.Context, params ListParams) (Page, error)
}
