package idempotency

import (
	"context"
	"time"
)

type Scope struct {
	TenantID       string
	Scope          string
	ActorID        string
	OrganizationID string
	WorkspaceID    string
	ProjectID      string
}

type Key struct {
	Key       string
	RequestID string
}

type Fingerprint struct {
	Algorithm               string
	Hash                    string
	CanonicalizationVersion string
	Method                  string
	Route                   string
	SchemaVersion           string
}

type Status string

const (
	StatusReceived        Status = "RECEIVED"
	StatusProcessing      Status = "PROCESSING"
	StatusCompleted       Status = "COMPLETED"
	StatusFailedRetryable Status = "FAILED_RETRYABLE"
	StatusFailedFinal     Status = "FAILED_FINAL"
	StatusUnknown         Status = "UNKNOWN"
	StatusExpired         Status = "EXPIRED"
)

type Decision string

const (
	DecisionExecute    Decision = "EXECUTE"
	DecisionReplay     Decision = "REPLAY"
	DecisionInProgress Decision = "IN_PROGRESS"
	DecisionConflict   Decision = "CONFLICT"
	DecisionRecover    Decision = "RECOVER"
	DecisionExpired    Decision = "EXPIRED"
)

type BeginOptions struct {
	TTL          time.Duration
	LockTTL      time.Duration
	Owner        string
	ReplayPolicy ReplayPolicy
	Labels       map[string]string
}

type ReplayPolicy string

const (
	ReplaySuccessOnly           ReplayPolicy = "SUCCESS_ONLY"
	ReplaySuccessAndFinalErrors ReplayPolicy = "SUCCESS_AND_FINAL_ERRORS"
	ReplayAll                   ReplayPolicy = "ALL"
)

type BeginResult struct {
	Decision   Decision
	Record     *Record
	LeaseToken string
	RetryAfter time.Duration
	Reason     string
}

type Store interface {
	Begin(
		ctx context.Context,
		scope Scope,
		key Key,
		fingerprint Fingerprint,
		options BeginOptions,
	) (*BeginResult, error)

	Commit(
		ctx context.Context,
		scope Scope,
		key Key,
		leaseToken string,
		result Result,
	) (*Record, error)

	Abort(
		ctx context.Context,
		scope Scope,
		key Key,
		leaseToken string,
		err error,
		retryable bool,
	) (*Record, error)

	Touch(
		ctx context.Context,
		scope Scope,
		key Key,
		leaseToken string,
		extendBy time.Duration,
	) error

	Get(
		ctx context.Context,
		scope Scope,
		key Key,
	) (*Record, error)
}
