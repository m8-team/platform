package project

import "context"

type ListParams struct {
	WorkspaceID string
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type Page struct {
	Items         []Project
	NextPageToken string
	TotalSize     int32
}

type Repository interface {
	GetByID(ctx context.Context, id string, includeDeleted bool) (Project, error)
	Create(ctx context.Context, aggregate Project) error
	Update(ctx context.Context, aggregate Project) error
	List(ctx context.Context, params ListParams) (Page, error)
}
