package organizationboundary

import (
	"context"
	"time"
)

type Organization struct {
	ID          string
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

type CreateOrganizationInput struct {
	Name        string
	Description string
	Annotations map[string]string
}

type CreateOrganizationOutput struct {
	Organization Organization
}

type UpdateOrganizationInput struct {
	ID          string
	ETag        string
	UpdateMask  []string
	Name        *string
	Description *string
	Annotations *map[string]string
}

type UpdateOrganizationOutput struct {
	Organization Organization
}

type DeleteOrganizationInput struct {
	ID           string
	ETag         string
	AllowMissing bool
}

type DeleteOrganizationOutput struct{}

type UndeleteOrganizationInput struct {
	ID string
}

type UndeleteOrganizationOutput struct {
	Organization Organization
}

type GetOrganizationInput struct {
	ID             string
	IncludeDeleted bool
}

type GetOrganizationOutput struct {
	Organization Organization
}

type ListOrganizationsInput struct {
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type ListOrganizationsOutput struct {
	Organizations []Organization
	NextPageToken string
	TotalSize     int32
}

type CreateHandler interface {
	Execute(context.Context, CreateOrganizationInput) (CreateOrganizationOutput, error)
}

type UpdateHandler interface {
	Execute(context.Context, UpdateOrganizationInput) (UpdateOrganizationOutput, error)
}

type DeleteHandler interface {
	Execute(context.Context, DeleteOrganizationInput) (DeleteOrganizationOutput, error)
}

type UndeleteHandler interface {
	Execute(context.Context, UndeleteOrganizationInput) (UndeleteOrganizationOutput, error)
}

type GetHandler interface {
	Execute(context.Context, GetOrganizationInput) (GetOrganizationOutput, error)
}

type ListHandler interface {
	Execute(context.Context, ListOrganizationsInput) (ListOrganizationsOutput, error)
}

type OrganizationCommandUseCase interface {
	CreateOrganization(context.Context, CreateOrganizationInput) (CreateOrganizationOutput, error)
	UpdateOrganization(context.Context, UpdateOrganizationInput) (UpdateOrganizationOutput, error)
	DeleteOrganization(context.Context, DeleteOrganizationInput) (DeleteOrganizationOutput, error)
	UndeleteOrganization(context.Context, UndeleteOrganizationInput) (UndeleteOrganizationOutput, error)
}

type OrganizationQueryUseCase interface {
	GetOrganization(context.Context, GetOrganizationInput) (GetOrganizationOutput, error)
	ListOrganizations(context.Context, ListOrganizationsInput) (ListOrganizationsOutput, error)
}
