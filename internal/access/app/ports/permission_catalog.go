package ports

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/access/domain"
)

var ErrPermissionNotRegistered = errors.New("permission is not registered for resource type")

// PermissionCatalog maps a public M8 action to one published SpiceDB permission.
// The caller must supply a verified tenant scope; the catalog MUST NOT grant
// broader access when a mapping is missing.
type PermissionCatalog interface {
	ResolvePermission(ctx context.Context, tenant domain.TenantID, permission domain.Permission, resourceType string) (domain.PermissionDefinition, error)
}
