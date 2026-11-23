// Package grant owns independent, time-bounded access grants and their lifecycle.
package grant

import (
	"errors"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrInvalidGrant            = errors.New("invalid access grant")
	ErrInvalidGrantTransition  = errors.New("invalid access grant state transition")
	ErrGrantExpired            = errors.New("access grant expired")
	ErrGrantNotYetValid        = errors.New("access grant not yet valid")
	ErrGrantSelfApproval       = errors.New("grant requester or beneficiary cannot approve")
	ErrGrantProjectionRequired = errors.New("activation requires verified projection evidence")
)

type Reason string

const (
	ReasonProjectAssignment Reason = "PROJECT_ASSIGNMENT"
	ReasonSupportCase       Reason = "SUPPORT_CASE"
	ReasonExternalContract  Reason = "EXTERNAL_CONTRACT"
	ReasonEmergency         Reason = "EMERGENCY"
	ReasonOneTimeTask       Reason = "ONE_TIME_TASK"
)

func (r Reason) Valid() bool {
	switch r {
	case ReasonProjectAssignment, ReasonSupportCase, ReasonExternalContract, ReasonEmergency, ReasonOneTimeTask:
		return true
	default:
		return false
	}
}

// AccessGrant is an immutable aggregate snapshot. Its grant ID preserves independent
// revocation, even when multiple grants yield the same graph permission.
type AccessGrant struct {
	id            string
	tenantID      shared.TenantID
	subject       shared.Subject
	target        Target
	permissions   []permission.Permission
	reason        Reason
	reasonRef     string
	constraints   Constraints
	validFrom     time.Time
	expiresAt     time.Time
	state         State
	revision      shared.Revision
	requestedBy   shared.Subject
	approvedBy    shared.Subject
	revokedBy     shared.Subject
	projectionRef string
	createdAt     time.Time
	updatedAt     time.Time
	revokedAt     time.Time
}
type Input struct {
	ID          string
	TenantID    shared.TenantID
	Subject     shared.Subject
	Target      Target
	Permissions []permission.Permission
	Reason      Reason
	ReasonRef   string
	Constraints Constraints
	ValidFrom   time.Time
	ExpiresAt   time.Time
	RequestedBy shared.Subject
	CreatedAt   time.Time
}

func NewAccessGrant(input Input) (AccessGrant, error) {
	g := AccessGrant{
		id: strings.TrimSpace(input.ID), tenantID: input.TenantID, subject: input.Subject,
		target: input.Target, permissions: append([]permission.Permission(nil), input.Permissions...),
		reason: input.Reason, reasonRef: strings.TrimSpace(input.ReasonRef), constraints: input.Constraints,
		validFrom: input.ValidFrom.UTC(), expiresAt: input.ExpiresAt.UTC(), state: Requested, revision: 1,
		requestedBy: input.RequestedBy, createdAt: input.CreatedAt.UTC(), updatedAt: input.CreatedAt.UTC(),
	}
	return g, g.Validate()
}
func (g AccessGrant) Validate() error {
	if !shared.ValidUUID(g.id) || g.tenantID.Validate() != nil || g.subject.Validate() != nil ||
		g.requestedBy.Validate() != nil || g.target.Validate() != nil || !g.reason.Valid() ||
		g.reasonRef == "" || g.revision.Validate() != nil {
		return ErrInvalidGrant
	}
	if err := permission.ValidateSet(g.permissions, true); err != nil {
		return err
	}
	if err := g.constraints.Validate(); err != nil {
		return err
	}
	if g.createdAt.IsZero() || g.validFrom.IsZero() || g.expiresAt.IsZero() ||
		!g.validFrom.Before(g.expiresAt) || !g.createdAt.Before(g.expiresAt) ||
		g.updatedAt.Before(g.createdAt) || !g.validState() {
		return ErrInvalidGrant
	}
	return nil
}
func (g AccessGrant) validState() bool {
	switch g.state {
	case Requested, PendingApproval, Approved, Active, Rejected, Revoked, Expired:
	default:
		return false
	}
	if g.state == Active && g.projectionRef == "" {
		return false
	}
	if g.state == Revoked && (g.revokedAt.IsZero() || g.revokedBy.Validate() != nil) {
		return false
	}
	return true
}
func (g AccessGrant) ID() string                { return g.id }
func (g AccessGrant) TenantID() shared.TenantID { return g.tenantID }
func (g AccessGrant) Subject() shared.Subject   { return g.subject }
func (g AccessGrant) Target() Target            { return g.target }
func (g AccessGrant) Permissions() []permission.Permission {
	return append([]permission.Permission(nil), g.permissions...)
}
func (g AccessGrant) Reason() Reason              { return g.reason }
func (g AccessGrant) ReasonRef() string           { return g.reasonRef }
func (g AccessGrant) Constraints() Constraints    { return g.constraints }
func (g AccessGrant) ValidFrom() time.Time        { return g.validFrom }
func (g AccessGrant) ExpiresAt() time.Time        { return g.expiresAt }
func (g AccessGrant) State() State                { return g.state }
func (g AccessGrant) Revision() int64             { return g.revision.Int64() }
func (g AccessGrant) RequestedBy() shared.Subject { return g.requestedBy }
func (g AccessGrant) ApprovedBy() shared.Subject  { return g.approvedBy }
func (g AccessGrant) RevokedBy() shared.Subject   { return g.revokedBy }
func (g AccessGrant) ProjectionRef() string       { return g.projectionRef }
func (g AccessGrant) CreatedAt() time.Time        { return g.createdAt }
func (g AccessGrant) UpdatedAt() time.Time        { return g.updatedAt }
func (g AccessGrant) RevokedAt() time.Time        { return g.revokedAt }
