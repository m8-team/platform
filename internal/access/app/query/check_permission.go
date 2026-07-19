package query

import "github.com/m8-team/platform/internal/access/domain"

type CheckPermissionResult struct {
	Decision      string
	Allowed       bool
	ModelRevision string
	Reason        string
	Degraded      bool
	FailureKind   string
}

func NewCheckPermissionResult(decision domain.PermissionDecision) CheckPermissionResult {
	return CheckPermissionResult{
		Decision:      decision.Decision().String(),
		Allowed:       decision.Allowed(),
		ModelRevision: decision.ModelRevision().String(),
		Reason:        decision.Reason(),
		Degraded:      decision.Degraded(),
		FailureKind:   decision.FailureKind().String(),
	}
}
