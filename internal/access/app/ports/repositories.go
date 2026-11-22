package ports

import (
	"context"

	"github.com/m8-team/platform/internal/access/domain"
)

// RoleRepository stores authoritative M8 role state, NOT SpiceDB tuples.
// SaveRole must perform compare-and-swap with expectedRevision and deduplicate
// the idempotency key within the specified tenant.
type RoleRepository interface {
	GetRole(ctx context.Context, tenant domain.TenantID, id string) (domain.Role, error)
	SaveRole(ctx context.Context, role domain.Role, expectedRevision int64, idempotencyKey string) error
}

// RoleBindingRepository owns scoped assignments. A role binding is a distinct
// business object, not a SpiceDB relationship with implicit grant semantics.
type RoleBindingRepository interface {
	GetRoleBinding(ctx context.Context, tenant domain.TenantID, id string) (domain.RoleBinding, error)
	SaveRoleBinding(ctx context.Context, binding domain.RoleBinding, expectedRevision int64, idempotencyKey string) error
}

// GrantRepository owns approval history and grant lifecycle. SpiceDB alone
// cannot distinguish multiple independent grants of the same permission.
type GrantRepository interface {
	GetGrant(ctx context.Context, tenant domain.TenantID, id string) (domain.AccessGrant, error)
	SaveGrant(ctx context.Context, grant domain.AccessGrant, expectedRevision int64, idempotencyKey string) error
}
