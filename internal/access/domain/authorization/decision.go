package authorization

import (
	"errors"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrInvalidDecision          = errors.New("invalid permission decision")
	ErrInvalidEngineFailureKind = errors.New("invalid engine failure kind")
	ErrModelRevisionMismatch    = errors.New("permission decision model revision mismatch")
)

type Decision string

const (
	Allow Decision = "ALLOW"
	Deny  Decision = "DENY"
)

func (d Decision) Valid() bool  { return d == Allow || d == Deny }
func (d Decision) Allows() bool { return d == Allow }

type EngineFailureKind string

const (
	FailureTimeout     EngineFailureKind = "TIMEOUT"
	FailureUnavailable EngineFailureKind = "UNAVAILABLE"
)

func (k EngineFailureKind) Valid() bool { return k == FailureTimeout || k == FailureUnavailable }

type PermissionDecision struct {
	zedToken      string
	decision      Decision
	modelRevision shared.ModelRevision
	reason        string
	degraded      bool
	failureKind   EngineFailureKind
}

func NewPermissionDecision(d Decision, rev shared.ModelRevision, reason string) (PermissionDecision, error) {
	if !d.Valid() || rev.Validate() != nil {
		return PermissionDecision{}, ErrInvalidDecision
	}
	return PermissionDecision{decision: d, modelRevision: rev, reason: reason}, nil
}
func NewEngineFailureDecision(req CheckPermissionRequest, kind EngineFailureKind, reason string) (PermissionDecision, error) {
	if !kind.Valid() {
		return PermissionDecision{}, ErrInvalidEngineFailureKind
	}
	if err := req.Validate(); err != nil {
		return PermissionDecision{}, err
	}
	decision := Deny
	if req.FailMode() == FailAllow {
		decision = Allow
	}
	return PermissionDecision{decision: decision, modelRevision: req.ModelRevision(), reason: reason, degraded: true, failureKind: kind}, nil
}
func (d PermissionDecision) EnsureModelRevision(expected shared.ModelRevision) error {
	if expected.Validate() != nil || !d.modelRevision.Equal(expected) {
		return ErrModelRevisionMismatch
	}
	return nil
}
func (d PermissionDecision) WithZedToken(token string) (PermissionDecision, error) {
	if token == "" || len(token) > 4096 {
		return PermissionDecision{}, ErrInvalidZedToken
	}
	d.zedToken = token
	return d, nil
}
func (d PermissionDecision) Decision() Decision                  { return d.decision }
func (d PermissionDecision) Allowed() bool                       { return d.decision.Allows() }
func (d PermissionDecision) ModelRevision() shared.ModelRevision { return d.modelRevision }
func (d PermissionDecision) Reason() string                      { return d.reason }
func (d PermissionDecision) Degraded() bool                      { return d.degraded }
func (d PermissionDecision) FailureKind() EngineFailureKind      { return d.failureKind }
func (d PermissionDecision) ZedToken() string                    { return d.zedToken }
