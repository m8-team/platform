package domain

import (
	"errors"
	"math"
	"strings"
	"time"
)

var (
	ErrInvalidGrant            = errors.New("invalid access grant")
	ErrInvalidGrantTransition  = errors.New("invalid access grant state transition")
	ErrGrantExpired            = errors.New("access grant has expired")
	ErrGrantNotYetValid        = errors.New("access grant is not yet valid")
	ErrGrantSelfApproval       = errors.New("access grant requester cannot approve their own request")
	ErrGrantProjectionRequired = errors.New("access grant activation requires projection evidence")
)

type GrantState string

const (
	GrantStateRequested       GrantState = "REQUESTED"
	GrantStatePendingApproval GrantState = "PENDING_APPROVAL"
	GrantStateApproved        GrantState = "APPROVED"
	GrantStateActive          GrantState = "ACTIVE"
	GrantStateRejected        GrantState = "REJECTED"
	GrantStateRevoked         GrantState = "REVOKED"
	GrantStateExpired         GrantState = "EXPIRED"
)

type GrantReason string

const (
	GrantReasonProjectAssignment GrantReason = "PROJECT_ASSIGNMENT"
	GrantReasonSupportCase       GrantReason = "SUPPORT_CASE"
	GrantReasonExternalContract  GrantReason = "EXTERNAL_CONTRACT"
	GrantReasonEmergency         GrantReason = "EMERGENCY"
	GrantReasonOneTimeTask       GrantReason = "ONE_TIME_TASK"
)

func (r GrantReason) Valid() bool {
	switch r {
	case GrantReasonProjectAssignment, GrantReasonSupportCase, GrantReasonExternalContract, GrantReasonEmergency, GrantReasonOneTimeTask:
		return true
	default:
		return false
	}
}

// GrantTarget binds the authorization scope to one authoritative organization.
// Validating this ID is not a substitute for resolving its tenant/owner via Resource Manager.
type GrantTarget struct {
	OrganizationID string
	Resource       Resource
	EnvironmentID  string
}

func (t GrantTarget) Validate() error {
	if !canonicalUUIDPattern.MatchString(t.OrganizationID) {
		return ErrInvalidGrant
	}
	if err := t.Resource.Validate(); err != nil {
		return err
	}
	if t.EnvironmentID != "" && !canonicalUUIDPattern.MatchString(t.EnvironmentID) {
		return ErrInvalidGrant
	}
	return nil
}

// GrantConstraints describe requirements that the PDP must verify from trusted
// sources. Merely observing an active grant does not satisfy these requirements.
type GrantConstraints struct {
	RequireMFA              bool
	RequireActiveCase       bool
	RequireActiveAssignment bool
	ServiceAgreementID      string
	ConditionPolicyRef      string
}

func (c GrantConstraints) Validate() error {
	if c.ConditionPolicyRef != "" && !relationNamePattern.MatchString(c.ConditionPolicyRef) {
		return ErrInvalidGrant
	}
	if len(c.ServiceAgreementID) > 256 {
		return ErrInvalidGrant
	}
	return nil
}

// AccessGrant is an immutable snapshot of a time-bounded authorization request.
// APPROVED is never an allow; ACTIVE can only be recorded after the application
// confirms projection, and even ACTIVE still requires a PDP decision.
type AccessGrant struct {
	id            string
	tenantID      TenantID
	subject       Subject
	target        GrantTarget
	permissions   []Permission
	reason        GrantReason
	reasonRef     string
	constraints   GrantConstraints
	validFrom     time.Time
	expiresAt     time.Time
	state         GrantState
	revision      int64
	requestedBy   Subject
	approvedBy    Subject
	revokedBy     Subject
	projectionRef string
	createdAt     time.Time
	updatedAt     time.Time
	revokedAt     time.Time
}

type AccessGrantInput struct {
	ID          string
	TenantID    TenantID
	Subject     Subject
	Target      GrantTarget
	Permissions []Permission
	Reason      GrantReason
	ReasonRef   string
	Constraints GrantConstraints
	ValidFrom   time.Time
	ExpiresAt   time.Time
	RequestedBy Subject
	CreatedAt   time.Time
}

func NewAccessGrant(input AccessGrantInput) (AccessGrant, error) {
	g := AccessGrant{
		id: strings.TrimSpace(input.ID), tenantID: input.TenantID,
		subject: input.Subject, target: input.Target,
		permissions: append([]Permission(nil), input.Permissions...),
		reason:      input.Reason, reasonRef: strings.TrimSpace(input.ReasonRef),
		constraints: input.Constraints, validFrom: input.ValidFrom.UTC(),
		expiresAt: input.ExpiresAt.UTC(), requestedBy: input.RequestedBy,
		createdAt: input.CreatedAt.UTC(), updatedAt: input.CreatedAt.UTC(),
		state: GrantStateRequested, revision: 1,
	}
	if err := g.Validate(); err != nil {
		return AccessGrant{}, err
	}
	return g, nil
}

func (g AccessGrant) Validate() error {
	if !canonicalUUIDPattern.MatchString(g.id) || g.revision < 1 || !g.reason.Valid() || g.reasonRef == "" {
		return ErrInvalidGrant
	}
	if err := g.tenantID.Validate(); err != nil {
		return err
	}
	if err := g.subject.Validate(); err != nil {
		return err
	}
	if err := g.requestedBy.Validate(); err != nil {
		return err
	}
	if err := g.target.Validate(); err != nil {
		return err
	}
	if err := validatePermissions(g.permissions, true); err != nil {
		return err
	}
	if err := g.constraints.Validate(); err != nil {
		return err
	}
	if g.createdAt.IsZero() || g.validFrom.IsZero() || g.expiresAt.IsZero() || !g.validFrom.Before(g.expiresAt) || !g.createdAt.Before(g.expiresAt) {
		return ErrInvalidGrant
	}
	return nil
}

func (g AccessGrant) advance(expectedRevision int64, now time.Time, from []GrantState, to GrantState) (AccessGrant, error) {
	if g.revision != expectedRevision {
		return AccessGrant{}, ErrRevisionConflict
	}
	allowed := false
	for _, state := range from {
		if g.state == state {
			allowed = true
			break
		}
	}
	if !allowed {
		return AccessGrant{}, ErrInvalidGrantTransition
	}
	if now.IsZero() || now.Before(g.updatedAt) || g.revision == math.MaxInt64 {
		return AccessGrant{}, ErrInvalidGrant
	}
	updated := g
	updated.state = to
	updated.revision++
	updated.updatedAt = now.UTC()
	return updated, nil
}

func (g AccessGrant) SubmitForApproval(expectedRevision int64, at time.Time) (AccessGrant, error) {
	if !at.Before(g.expiresAt) {
		return AccessGrant{}, ErrGrantExpired
	}
	return g.advance(expectedRevision, at, []GrantState{GrantStateRequested}, GrantStatePendingApproval)
}

func (g AccessGrant) Approve(expectedRevision int64, actor Subject, at time.Time) (AccessGrant, error) {
	if err := actor.Validate(); err != nil {
		return AccessGrant{}, err
	}
	if actor == g.requestedBy {
		return AccessGrant{}, ErrGrantSelfApproval
	}
	if !at.Before(g.expiresAt) {
		return AccessGrant{}, ErrGrantExpired
	}
	updated, err := g.advance(expectedRevision, at, []GrantState{GrantStatePendingApproval}, GrantStateApproved)
	if err != nil {
		return AccessGrant{}, err
	}
	updated.approvedBy = actor
	return updated, nil
}

// Activate requires proof of a successfully applied authorization projection.
// The application layer is responsible for validating that projection reference
// against the authoritative store/watermark before calling this transition.
func (g AccessGrant) Activate(expectedRevision int64, projectionRef string, at time.Time) (AccessGrant, error) {
	projectionRef = strings.TrimSpace(projectionRef)
	if projectionRef == "" {
		return AccessGrant{}, ErrGrantProjectionRequired
	}
	if at.Before(g.validFrom) {
		return AccessGrant{}, ErrGrantNotYetValid
	}
	if !at.Before(g.expiresAt) {
		return AccessGrant{}, ErrGrantExpired
	}
	updated, err := g.advance(expectedRevision, at, []GrantState{GrantStateApproved}, GrantStateActive)
	if err != nil {
		return AccessGrant{}, err
	}
	updated.projectionRef = projectionRef
	return updated, nil
}

func (g AccessGrant) Reject(expectedRevision int64, at time.Time) (AccessGrant, error) {
	return g.advance(expectedRevision, at, []GrantState{GrantStateRequested, GrantStatePendingApproval}, GrantStateRejected)
}

// Revoke is an authoritative denial: the PDP must suppress any stale positive
// graph results until projection convergence has been verified.
func (g AccessGrant) Revoke(expectedRevision int64, actor Subject, at time.Time) (AccessGrant, error) {
	if err := actor.Validate(); err != nil {
		return AccessGrant{}, err
	}
	updated, err := g.advance(expectedRevision, at, []GrantState{GrantStateApproved, GrantStateActive}, GrantStateRevoked)
	if err != nil {
		return AccessGrant{}, err
	}
	updated.revokedBy = actor
	updated.revokedAt = at.UTC()
	return updated, nil
}

func (g AccessGrant) Expire(expectedRevision int64, at time.Time) (AccessGrant, error) {
	if at.Before(g.expiresAt) {
		return AccessGrant{}, ErrInvalidGrantTransition
	}
	return g.advance(expectedRevision, at,
		[]GrantState{GrantStateRequested, GrantStatePendingApproval, GrantStateApproved, GrantStateActive}, GrantStateExpired)
}

// EffectiveAt checks only the grant's state and time interval. This result must
// never be used as authorization without separately evaluating constraints and PDP.
func (g AccessGrant) EffectiveAt(at time.Time) bool {
	return g.state == GrantStateActive && !at.Before(g.validFrom) && at.Before(g.expiresAt)
}

func (g AccessGrant) ID() string                    { return g.id }
func (g AccessGrant) TenantID() TenantID            { return g.tenantID }
func (g AccessGrant) Subject() Subject              { return g.subject }
func (g AccessGrant) Target() GrantTarget           { return g.target }
func (g AccessGrant) Permissions() []Permission     { return append([]Permission(nil), g.permissions...) }
func (g AccessGrant) Reason() GrantReason           { return g.reason }
func (g AccessGrant) ReasonRef() string             { return g.reasonRef }
func (g AccessGrant) Constraints() GrantConstraints { return g.constraints }
func (g AccessGrant) ValidFrom() time.Time          { return g.validFrom }
func (g AccessGrant) ExpiresAt() time.Time          { return g.expiresAt }
func (g AccessGrant) State() GrantState             { return g.state }
func (g AccessGrant) Revision() int64               { return g.revision }
func (g AccessGrant) RequestedBy() Subject          { return g.requestedBy }
func (g AccessGrant) ApprovedBy() Subject           { return g.approvedBy }
func (g AccessGrant) RevokedBy() Subject            { return g.revokedBy }
func (g AccessGrant) ProjectionRef() string         { return g.projectionRef }
func (g AccessGrant) CreatedAt() time.Time          { return g.createdAt }
func (g AccessGrant) UpdatedAt() time.Time          { return g.updatedAt }
func (g AccessGrant) RevokedAt() time.Time          { return g.revokedAt }
