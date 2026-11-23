package role

import (
	"errors"
	"github.com/m8-team/platform/internal/access/domain/shared"
	"strings"
	"time"
)

var (
	ErrInvalidRoleBinding = errors.New("invalid role binding")
	ErrRoleTenantMismatch = errors.New("role binding belongs to another role or tenant")
	ErrBindingNotActive   = errors.New("role binding is not active")
)

type BindingState string

const (
	BindingActive  BindingState = "ACTIVE"
	BindingRevoked BindingState = "REVOKED"
)

// RoleBinding is its own aggregate, allowing independent assignments and revocations.
type RoleBinding struct {
	id        string
	tenantID  shared.TenantID
	roleID    string
	subject   shared.Subject
	scope     shared.Resource
	revision  shared.Revision
	state     BindingState
	revokedAt time.Time
}
type RoleBindingInput struct {
	ID       string
	TenantID shared.TenantID
	RoleID   string
	Subject  shared.Subject
	Scope    shared.Resource
	Revision int64
}

func NewRoleBinding(input RoleBindingInput) (RoleBinding, error) {
	b := RoleBinding{id: strings.TrimSpace(input.ID), tenantID: input.TenantID, roleID: strings.TrimSpace(input.RoleID),
		subject: input.Subject, scope: input.Scope, revision: shared.Revision(input.Revision), state: BindingActive}
	return b, b.Validate()
}
func (b RoleBinding) Validate() error {
	if !shared.ValidUUID(b.id) || !shared.ValidUUID(b.roleID) || b.tenantID.Validate() != nil || b.revision.Validate() != nil || b.subject.Validate() != nil || b.scope.Validate() != nil {
		return ErrInvalidRoleBinding
	}
	if b.state != BindingActive && b.state != BindingRevoked {
		return ErrInvalidRoleBinding
	}
	if b.state == BindingActive && !b.revokedAt.IsZero() || b.state == BindingRevoked && b.revokedAt.IsZero() {
		return ErrInvalidRoleBinding
	}
	return nil
}
func (b RoleBinding) ValidateForRole(r Role) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if b.roleID != r.id || b.tenantID != r.tenantID {
		return ErrRoleTenantMismatch
	}
	// Resource Manager must separately verify that scope belongs to the same tenant.
	return nil
}
func (b RoleBinding) Revoke(expected int64, at time.Time) (RoleBinding, error) {
	if b.state != BindingActive {
		return RoleBinding{}, ErrBindingNotActive
	}
	if at.IsZero() {
		return RoleBinding{}, ErrInvalidRoleBinding
	}
	next, err := b.revision.Next(expected)
	if err != nil {
		return RoleBinding{}, err
	}
	updated := b
	updated.revision = next
	updated.state = BindingRevoked
	updated.revokedAt = at.UTC()
	return updated, updated.Validate()
}
func (b RoleBinding) ID() string                { return b.id }
func (b RoleBinding) TenantID() shared.TenantID { return b.tenantID }
func (b RoleBinding) RoleID() string            { return b.roleID }
func (b RoleBinding) Subject() shared.Subject   { return b.subject }
func (b RoleBinding) Scope() shared.Resource    { return b.scope }
func (b RoleBinding) Revision() int64           { return b.revision.Int64() }
func (b RoleBinding) State() BindingState       { return b.state }
func (b RoleBinding) RevokedAt() time.Time      { return b.revokedAt }
