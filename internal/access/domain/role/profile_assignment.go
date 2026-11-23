package role

import (
	"errors"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidProfileAssignment = errors.New("invalid scoped profile assignment")

// ProfileAssignmentStatus deliberately has NO ACTIVE state. A persisted
// assignment cannot authorize anything until a separately reviewed policy
// journal + fenced graph projection explicitly activates it.
type ProfileAssignmentStatus string

const (
	ProfileAssignmentPending ProfileAssignmentStatus = "PENDING_PROJECTION"
	ProfileAssignmentRevoked ProfileAssignmentStatus = "REVOKED"
)

// ProfileAssignment binds an exact published template/profile version to
// independently scoped protected resource kinds. Multi-module bundles need
// one scope per resource type; a tenant-wide wildcard is forbidden.
type ProfileAssignment struct {
	id        string
	tenantID  shared.TenantID
	subject   shared.Subject
	target    AssignmentTarget
	scopes    []authorization.Scope
	policyIDs []string
	revision  shared.Revision
	status    ProfileAssignmentStatus
	revokedAt time.Time
}

type ProfileAssignmentInput struct {
	ID        string
	TenantID  shared.TenantID
	Subject   shared.Subject
	Target    AssignmentTarget
	Scopes    []authorization.Scope
	PolicyIDs []string
}

func NewProfileAssignment(in ProfileAssignmentInput) (ProfileAssignment, error) {
	return RestoreProfileAssignment(in, 1, ProfileAssignmentPending, time.Time{})
}

// RestoreProfileAssignment validates a complete persisted aggregate snapshot.
// Storage remains a reference record; it is not a PDP permission decision.
func RestoreProfileAssignment(in ProfileAssignmentInput, revision int64, status ProfileAssignmentStatus, revokedAt time.Time) (ProfileAssignment, error) {
	a := ProfileAssignment{
		id: strings.TrimSpace(in.ID), tenantID: in.TenantID, subject: in.Subject,
		target: in.Target, scopes: append([]authorization.Scope(nil), in.Scopes...),
		policyIDs: append([]string(nil), in.PolicyIDs...), revision: shared.Revision(revision),
		status: status, revokedAt: revokedAt.UTC(),
	}
	return a, a.Validate()
}

func (a ProfileAssignment) Validate() error {
	if !shared.ValidUUID(a.id) || a.tenantID.Validate() != nil ||
		a.subject.Validate() != nil || a.target.Validate() != nil ||
		a.revision.Validate() != nil || len(a.scopes) == 0 || len(a.scopes) > 64 ||
		len(a.policyIDs) > 64 {
		return ErrInvalidProfileAssignment
	}
	seenKinds := make(map[string]bool, len(a.scopes))
	for _, scope := range a.scopes {
		if scope.Validate() != nil || scope.TenantID() != a.tenantID ||
			seenKinds[scope.ResourceType()] {
			return ErrInvalidProfileAssignment
		}
		seenKinds[scope.ResourceType()] = true
	}
	seenPolicies := make(map[string]bool, len(a.policyIDs))
	for _, id := range a.policyIDs {
		if !shared.ValidUUID(id) || seenPolicies[id] {
			return ErrInvalidProfileAssignment
		}
		seenPolicies[id] = true
	}
	switch a.status {
	case ProfileAssignmentPending:
		if !a.revokedAt.IsZero() {
			return ErrInvalidProfileAssignment
		}
	case ProfileAssignmentRevoked:
		if a.revokedAt.IsZero() || a.revision.Int64() < 2 {
			return ErrInvalidProfileAssignment
		}
	default:
		return ErrInvalidProfileAssignment
	}
	return nil
}

func (a ProfileAssignment) Revoke(expected int64, at time.Time) (ProfileAssignment, error) {
	if a.status != ProfileAssignmentPending {
		return ProfileAssignment{}, ErrAssignmentNotActive
	}
	if at.IsZero() {
		return ProfileAssignment{}, ErrInvalidProfileAssignment
	}
	next, err := a.revision.Next(expected)
	if err != nil {
		return ProfileAssignment{}, err
	}
	a.revision = next
	a.status = ProfileAssignmentRevoked
	a.revokedAt = at.UTC()
	return a, a.Validate()
}

func (a ProfileAssignment) ID() string                { return a.id }
func (a ProfileAssignment) TenantID() shared.TenantID { return a.tenantID }
func (a ProfileAssignment) Subject() shared.Subject   { return a.subject }
func (a ProfileAssignment) Target() AssignmentTarget  { return a.target }
func (a ProfileAssignment) Scopes() []authorization.Scope {
	return append([]authorization.Scope(nil), a.scopes...)
}
func (a ProfileAssignment) PolicyIDs() []string             { return append([]string(nil), a.policyIDs...) }
func (a ProfileAssignment) Revision() int64                 { return a.revision.Int64() }
func (a ProfileAssignment) Status() ProfileAssignmentStatus { return a.status }
func (a ProfileAssignment) RevokedAt() time.Time            { return a.revokedAt }
