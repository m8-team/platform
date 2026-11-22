package domain

import (
	"errors"
	"math"
	"strings"
)

var (
	ErrInvalidRole         = errors.New("invalid role")
	ErrInvalidRoleBinding  = errors.New("invalid role binding")
	ErrDuplicatePermission = errors.New("duplicate permission")
	ErrRevisionConflict    = errors.New("revision conflict")
	ErrRoleTenantMismatch  = errors.New("role binding belongs to another role or tenant")
)

// Role is a tenant-scoped permission set, not a global principal identity.
// Roles do not grant access without an explicit scoped RoleBinding.
type Role struct {
	id          string
	tenantID    TenantID
	name        string
	permissions []Permission
	revision    int64
}

type RoleInput struct {
	ID          string
	TenantID    TenantID
	Name        string
	Permissions []Permission
	Revision    int64
}

func NewRole(input RoleInput) (Role, error) {
	r := Role{id: strings.TrimSpace(input.ID), tenantID: input.TenantID,
		name: strings.TrimSpace(input.Name), permissions: append([]Permission(nil), input.Permissions...), revision: input.Revision}
	return r, r.Validate()
}

func (r Role) Validate() error {
	if !canonicalUUIDPattern.MatchString(r.id) {
		return ErrInvalidRole
	}
	if err := r.tenantID.Validate(); err != nil {
		return err
	}
	if !relationNamePattern.MatchString(r.name) || r.revision < 1 {
		return ErrInvalidRole
	}
	return validatePermissions(r.permissions, false)
}

func validatePermissions(permissions []Permission, nonEmpty bool) error {
	if nonEmpty && len(permissions) == 0 {
		return ErrPermissionRequired
	}
	seen := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		if err := p.Validate(); err != nil {
			return err
		}
		if _, exists := seen[p.Name()]; exists {
			return ErrDuplicatePermission
		}
		seen[p.Name()] = struct{}{}
	}
	return nil
}

// WithPermissions returns a new role revision; persistence must atomically
// compare expectedRevision before replacing authoritative state.
func (r Role) WithPermissions(expectedRevision int64, permissions []Permission) (Role, error) {
	if expectedRevision != r.revision {
		return Role{}, ErrRevisionConflict
	}
	if r.revision == math.MaxInt64 {
		return Role{}, ErrInvalidRole
	}
	updated := r
	updated.permissions = append([]Permission(nil), permissions...)
	updated.revision++
	if err := updated.Validate(); err != nil {
		return Role{}, err
	}
	return updated, nil
}

func (r Role) ID() string                { return r.id }
func (r Role) TenantID() TenantID        { return r.tenantID }
func (r Role) Name() string              { return r.name }
func (r Role) Revision() int64           { return r.revision }
func (r Role) Permissions() []Permission { return append([]Permission(nil), r.permissions...) }

// RoleBinding attaches a role to a subject *at a specified resource scope*.
// Membership in an organization group alone must not create a role binding.
type RoleBinding struct {
	id       string
	tenantID TenantID
	roleID   string
	subject  Subject
	scope    Resource
	revision int64
}

type RoleBindingInput struct {
	ID       string
	TenantID TenantID
	RoleID   string
	Subject  Subject
	Scope    Resource
	Revision int64
}

func NewRoleBinding(input RoleBindingInput) (RoleBinding, error) {
	b := RoleBinding{id: strings.TrimSpace(input.ID), tenantID: input.TenantID,
		roleID: strings.TrimSpace(input.RoleID), subject: input.Subject,
		scope: input.Scope, revision: input.Revision}
	return b, b.Validate()
}

func (b RoleBinding) Validate() error {
	if !canonicalUUIDPattern.MatchString(b.id) || !canonicalUUIDPattern.MatchString(b.roleID) || b.revision < 1 {
		return ErrInvalidRoleBinding
	}
	if err := b.tenantID.Validate(); err != nil {
		return err
	}
	if err := b.subject.Validate(); err != nil {
		return err
	}
	return b.scope.Validate()
}

// ValidateForRole ensures a binding cannot attach a role belonging to a
// different tenant. The owning service must still check the scope's tenant.
func (b RoleBinding) ValidateForRole(role Role) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if err := role.Validate(); err != nil {
		return err
	}
	if b.roleID != role.id || b.tenantID != role.tenantID {
		return ErrRoleTenantMismatch
	}
	return nil
}

func (b RoleBinding) ID() string         { return b.id }
func (b RoleBinding) TenantID() TenantID { return b.tenantID }
func (b RoleBinding) RoleID() string     { return b.roleID }
func (b RoleBinding) Subject() Subject   { return b.subject }
func (b RoleBinding) Scope() Resource    { return b.scope }
func (b RoleBinding) Revision() int64    { return b.revision }
