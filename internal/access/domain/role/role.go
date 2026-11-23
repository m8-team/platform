// Package role contains independent aggregate roots for role definitions and assignments.
package role

import (
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidRole = errors.New("invalid role")

// Role is a tenant-scoped set of actions. Without an explicit RoleBinding it grants nothing.
type Role struct {
	id          string
	tenantID    shared.TenantID
	name        string
	permissions []permission.Permission
	revision    shared.Revision
}
type RoleInput struct {
	ID          string
	TenantID    shared.TenantID
	Name        string
	Permissions []permission.Permission
	Revision    int64
}

func NewRole(input RoleInput) (Role, error) {
	r := Role{id: strings.TrimSpace(input.ID), tenantID: input.TenantID, name: strings.TrimSpace(input.Name), permissions: append([]permission.Permission(nil), input.Permissions...), revision: shared.Revision(input.Revision)}
	return r, r.Validate()
}
func (r Role) Validate() error {
	if !shared.ValidUUID(r.id) || r.tenantID.Validate() != nil || r.revision.Validate() != nil || !validRoleName(r.name) {
		return ErrInvalidRole
	}
	return permission.ValidateSet(r.permissions, false)
}
func validRoleName(name string) bool {
	if len(name) < 1 || len(name) > 128 || name[0] < 'a' || name[0] > 'z' {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}
func (r Role) WithPermissions(expected int64, p []permission.Permission) (Role, error) {
	next, err := r.revision.Next(expected)
	if err != nil {
		return Role{}, err
	}
	updated := r
	updated.revision = next
	updated.permissions = append([]permission.Permission(nil), p...)
	return updated, updated.Validate()
}
func (r Role) ID() string                { return r.id }
func (r Role) TenantID() shared.TenantID { return r.tenantID }
func (r Role) Name() string              { return r.name }
func (r Role) Revision() int64           { return r.revision.Int64() }
func (r Role) Permissions() []permission.Permission {
	return append([]permission.Permission(nil), r.permissions...)
}
