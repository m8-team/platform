package domain

import "fmt"

type Decision string

const (
	DecisionAllow Decision = "ALLOW"
	DecisionDeny  Decision = "DENY"
)

func (d Decision) IsValid() bool {
	switch d {
	case DecisionAllow, DecisionDeny:
		return true
	default:
		return false
	}
}

func (d Decision) String() string {
	return string(d)
}

func (d Decision) Allows() bool {
	return d == DecisionAllow
}

type FailMode string

const (
	FailModeDeny  FailMode = "DENY"
	FailModeAllow FailMode = "ALLOW"
)

func (m FailMode) WithDefault() FailMode {
	if m == "" {
		return FailModeDeny
	}

	return m
}

func (m FailMode) IsValid() bool {
	switch m {
	case FailModeDeny, FailModeAllow:
		return true
	default:
		return false
	}
}

func (m FailMode) String() string {
	return string(m)
}

func (m FailMode) Decision() Decision {
	if m == FailModeAllow {
		return DecisionAllow
	}

	return DecisionDeny
}

type EngineFailureKind string

const (
	EngineFailureTimeout     EngineFailureKind = "TIMEOUT"
	EngineFailureUnavailable EngineFailureKind = "UNAVAILABLE"
)

func (k EngineFailureKind) IsValid() bool {
	switch k {
	case EngineFailureTimeout, EngineFailureUnavailable:
		return true
	default:
		return false
	}
}

func (k EngineFailureKind) String() string {
	return string(k)
}

type PermissionDecision struct {
	decision      Decision
	modelRevision ModelRevision
	reason        string
	degraded      bool
	failureKind   EngineFailureKind
}

func NewPermissionDecision(decision Decision, modelRevision ModelRevision, reason string) (PermissionDecision, error) {
	if !decision.IsValid() {
		return PermissionDecision{}, fmt.Errorf("%w: %q", ErrInvalidDecision, decision)
	}
	if err := modelRevision.Validate(); err != nil {
		return PermissionDecision{}, err
	}

	return PermissionDecision{
		decision:      decision,
		modelRevision: modelRevision,
		reason:        reason,
	}, nil
}

func NewEngineFailureDecision(
	request CheckPermissionRequest,
	kind EngineFailureKind,
	reason string,
) (PermissionDecision, error) {
	if !kind.IsValid() {
		return PermissionDecision{}, fmt.Errorf("%w: %q", ErrInvalidEngineFailureKind, kind)
	}
	if err := request.Validate(); err != nil {
		return PermissionDecision{}, err
	}

	return PermissionDecision{
		decision:      request.FailMode().Decision(),
		modelRevision: request.ModelRevision(),
		reason:        reason,
		degraded:      true,
		failureKind:   kind,
	}, nil
}

func (d PermissionDecision) EnsureModelRevision(expected ModelRevision) error {
	if err := expected.Validate(); err != nil {
		return err
	}
	if !d.modelRevision.Equal(expected) {
		return ErrModelRevisionMismatch
	}

	return nil
}

func (d PermissionDecision) Decision() Decision {
	return d.decision
}

func (d PermissionDecision) Allowed() bool {
	return d.decision.Allows()
}

func (d PermissionDecision) ModelRevision() ModelRevision {
	return d.modelRevision
}

func (d PermissionDecision) Reason() string {
	return d.reason
}

func (d PermissionDecision) Degraded() bool {
	return d.degraded
}

func (d PermissionDecision) FailureKind() EngineFailureKind {
	return d.failureKind
}
