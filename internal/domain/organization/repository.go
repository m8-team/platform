package organization

import "context"

type ListParams struct {
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type Page struct {
	Items         []Organization
	NextPageToken string
	TotalSize     int32
}

type Repository interface {
	GetByID(ctx context.Context, id string, includeDeleted bool) (Organization, error)
	Create(ctx context.Context, aggregate Organization) error
	Update(ctx context.Context, aggregate Organization) error
	List(ctx context.Context, params ListParams) (Page, error)
}
