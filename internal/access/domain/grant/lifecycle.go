package grant

import (
	"time"

	"github.com/m8-team/platform/internal/access/domain/shared"
)

type State string

const (
	Requested       State = "REQUESTED"
	PendingApproval State = "PENDING_APPROVAL"
	Approved        State = "APPROVED"
	Active          State = "ACTIVE"
	Rejected        State = "REJECTED"
	Revoked         State = "REVOKED"
	Expired         State = "EXPIRED"
)

func (g AccessGrant) advance(expected int64, at time.Time, from []State, to State) (AccessGrant, error) {
	// Check expected revision first so a stale command never changes aggregate state.
	next, err := g.revision.Next(expected)
	if err != nil {
		return AccessGrant{}, err
	}
	accepted := false
	for _, s := range from {
		if g.state == s {
			accepted = true
			break
		}
	}
	if !accepted {
		return AccessGrant{}, ErrInvalidGrantTransition
	}
	if at.IsZero() || at.Before(g.updatedAt) {
		return AccessGrant{}, ErrInvalidGrant
	}
	updated := g
	updated.revision = next
	updated.state = to
	updated.updatedAt = at.UTC()
	return updated, nil
}
func (g AccessGrant) SubmitForApproval(expected int64, at time.Time) (AccessGrant, error) {
	if !at.Before(g.expiresAt) {
		return AccessGrant{}, ErrGrantExpired
	}
	return g.advance(expected, at, []State{Requested}, PendingApproval)
}
func (g AccessGrant) Approve(expected int64, actor shared.Subject, at time.Time) (AccessGrant, error) {
	if err := actor.Validate(); err != nil {
		return AccessGrant{}, err
	}
	if actor == g.requestedBy || actor == g.subject {
		return AccessGrant{}, ErrGrantSelfApproval
	}
	if !at.Before(g.expiresAt) {
		return AccessGrant{}, ErrGrantExpired
	}
	updated, err := g.advance(expected, at, []State{PendingApproval}, Approved)
	if err != nil {
		return AccessGrant{}, err
	}
	updated.approvedBy = actor
	return updated, updated.Validate()
}

// Activate accepts a projection proof already verified by the application consistency barrier.
// A nonempty string alone is not proof, and MUST NOT authorize a decision.
func (g AccessGrant) Activate(expected int64, projectionRef string, at time.Time) (AccessGrant, error) {
	if projectionRef == "" {
		return AccessGrant{}, ErrGrantProjectionRequired
	}
	if at.Before(g.validFrom) {
		return AccessGrant{}, ErrGrantNotYetValid
	}
	if !at.Before(g.expiresAt) {
		return AccessGrant{}, ErrGrantExpired
	}
	updated, err := g.advance(expected, at, []State{Approved}, Active)
	if err != nil {
		return AccessGrant{}, err
	}
	updated.projectionRef = projectionRef
	return updated, updated.Validate()
}
func (g AccessGrant) Reject(expected int64, at time.Time) (AccessGrant, error) {
	updated, err := g.advance(expected, at, []State{Requested, PendingApproval}, Rejected)
	if err != nil {
		return AccessGrant{}, err
	}
	return updated, updated.Validate()
}
func (g AccessGrant) Revoke(expected int64, actor shared.Subject, at time.Time) (AccessGrant, error) {
	if err := actor.Validate(); err != nil {
		return AccessGrant{}, err
	}
	updated, err := g.advance(expected, at, []State{Approved, Active}, Revoked)
	if err != nil {
		return AccessGrant{}, err
	}
	updated.revokedBy = actor
	updated.revokedAt = at.UTC()
	return updated, updated.Validate()
}
func (g AccessGrant) Expire(expected int64, at time.Time) (AccessGrant, error) {
	if at.Before(g.expiresAt) {
		return AccessGrant{}, ErrInvalidGrantTransition
	}
	updated, err := g.advance(expected, at, []State{Requested, PendingApproval, Approved, Active}, Expired)
	if err != nil {
		return AccessGrant{}, err
	}
	return updated, updated.Validate()
}

// EffectiveAt reports only the temporal state, NEVER the final authorization decision.
func (g AccessGrant) EffectiveAt(at time.Time) bool {
	return g.state == Active && !at.Before(g.validFrom) && at.Before(g.expiresAt)
}
