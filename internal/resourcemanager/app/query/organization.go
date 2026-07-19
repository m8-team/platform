package query

import (
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

type GetOrganization struct {
	ID organization.ID
}

// ListOrganizations keeps the public filter/order syntax at the application
// boundary. The use case validates and converts it into repository criteria.
type ListOrganizations struct {
	PageSize    int
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type ListOrganizationsResult struct {
	Organizations []*organization.Organization
	NextPageToken string
	TotalSize     int
}
