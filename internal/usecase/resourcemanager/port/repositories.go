package port

import (
	"context"

	"github.com/m8platform/platform/internal/entity/resourcemanager/organization"
)

type OrganizationListParams struct {
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type OrganizationPage struct {
	Items         []organization.Entity
	NextPageToken string
	TotalSize     int32
}

type OrganizationReader interface {
	GetByID(ctx context.Context, id string, includeDeleted bool) (organization.Entity, error)
	List(ctx context.Context, params OrganizationListParams) (OrganizationPage, error)
}

type OrganizationWriter interface {
	Create(ctx context.Context, entity organization.Entity) error
	Update(ctx context.Context, entity organization.Entity) error
	SoftDelete(ctx context.Context, entity organization.Entity) error
	Undelete(ctx context.Context, entity organization.Entity) error
}

type OrganizationRepository interface {
	OrganizationReader
	OrganizationWriter
}

type HierarchyReader interface {
	HasActiveWorkspaces(ctx context.Context, organizationID string) (bool, error)
}
