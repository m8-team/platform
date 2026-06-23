package domain

import (
	"fmt"
	"time"
)

// Decision tells the caller what to do after Begin.
type Decision string

const (
	DecisionExecute    Decision = "EXECUTE"
	DecisionReplay     Decision = "REPLAY"
	DecisionInProgress Decision = "IN_PROGRESS"
	DecisionConflict   Decision = "CONFLICT"
	DecisionRecover    Decision = "RECOVER"
	DecisionExpired    Decision = "EXPIRED"
)

func (d Decision) String() string {
	return string(d)
}

func (d Decision) IsValid() bool {
	switch d {
	case DecisionExecute,
		DecisionReplay,
		DecisionInProgress,
		DecisionConflict,
		DecisionRecover,
		DecisionExpired:
		return true
	default:
		return false
	}
}

func ParseDecision(value string) (Decision, error) {
	decision := Decision(value)
	if !decision.IsValid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidDecision, value)
	}

	return decision, nil
}

// ReplayPolicy defines which results can be stored and replayed.
type ReplayPolicy string

const (
	ReplayPolicySuccessOnly           ReplayPolicy = "SUCCESS_ONLY"
	ReplayPolicySuccessAndFinalErrors ReplayPolicy = "SUCCESS_AND_FINAL_ERRORS"
	ReplayPolicyAll                   ReplayPolicy = "ALL"
)

func (p ReplayPolicy) String() string {
	return string(p)
}

func (p ReplayPolicy) IsValid() bool {
	switch p {
	case ReplayPolicySuccessOnly,
		ReplayPolicySuccessAndFinalErrors,
		ReplayPolicyAll:
		return true
	default:
		return false
	}
}

func ParseReplayPolicy(value string) (ReplayPolicy, error) {
	policy := ReplayPolicy(value)
	if !policy.IsValid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidReplayPolicy, value)
	}

	return policy, nil
}

// BeginOptions controls claim/lease behavior for a single idempotency scope.
type BeginOptions struct {
	TTL          time.Duration
	LockTTL      time.Duration
	Owner        string
	ReplayPolicy ReplayPolicy
	Labels       map[string]string
}

// BeginResult is returned by Store.Begin.
type BeginResult struct {
	Decision   Decision
	Record     *Record
	LeaseToken string
	RetryAfter time.Duration
	Reason     string
}

func NewExecuteResult(record *Record, leaseToken string) *BeginResult {
	return &BeginResult{
		Decision:   DecisionExecute,
		Record:     record,
		LeaseToken: leaseToken,
	}
}

func NewReplayResult(record *Record) *BeginResult {
	return &BeginResult{
		Decision: DecisionReplay,
		Record:   record,
	}
}

func NewInProgressResult(record *Record, retryAfter time.Duration) *BeginResult {
	return &BeginResult{
		Decision:   DecisionInProgress,
		Record:     record,
		RetryAfter: retryAfter,
	}
}

func NewConflictResult(record *Record, reason string) *BeginResult {
	return &BeginResult{
		Decision: DecisionConflict,
		Record:   record,
		Reason:   reason,
	}
}

func NewRecoverResult(record *Record, leaseToken string, reason string) *BeginResult {
	return &BeginResult{
		Decision:   DecisionRecover,
		Record:     record,
		LeaseToken: leaseToken,
		Reason:     reason,
	}
}

func NewExpiredResult(record *Record) *BeginResult {
	return &BeginResult{
		Decision: DecisionExpired,
		Record:   record,
	}
}
