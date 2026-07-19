package ports

import (
	"context"
	"errors"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

var (
	ErrUnauthenticated  = errors.New("authentication is required")
	ErrPermissionDenied = errors.New("permission denied")
)

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() organization.ID
}

type AuthorizationAction string

const (
	ActionCreateOrganization   AuthorizationAction = "resourcemanager.organizations.create"
	ActionGetOrganization      AuthorizationAction = "resourcemanager.organizations.get"
	ActionListOrganizations    AuthorizationAction = "resourcemanager.organizations.list"
	ActionUpdateOrganization   AuthorizationAction = "resourcemanager.organizations.update"
	ActionDeleteOrganization   AuthorizationAction = "resourcemanager.organizations.delete"
	ActionUndeleteOrganization AuthorizationAction = "resourcemanager.organizations.undelete"
)

type AuthorizationRequest struct {
	Action         AuthorizationAction
	OrganizationID organization.ID
}

type Authorizer interface {
	Authorize(ctx context.Context, request AuthorizationRequest) error
	// ScopeKey returns a stable, non-secret fingerprint for the authenticated
	// caller and effective authorization scope. It is used to bind page tokens.
	ScopeKey(ctx context.Context) (string, error)
}

type WorkspaceChildren interface {
	HasNonDeleted(ctx context.Context, organizationID organization.ID) (bool, error)
}
