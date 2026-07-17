package domain

import (
	"fmt"
	"strings"
	"time"
)

// Scope defines the uniqueness boundary for an idempotency key.
//
// Main identity:
//
//	tenant_id + scope + idempotency_key
//
// Optional dimensions are useful for audit, debugging, metrics and policy,
// but they should not silently change the uniqueness contract unless the store
// explicitly includes them in its unique index.
type Scope struct {
	TenantID       string
	Scope          string
	ActorID        string
	OrganizationID string
	WorkspaceID    string
	ProjectID      string
}

func (s Scope) Identity() string {
	return s.TenantID + ":" + s.Scope
}

func (s Scope) Validate() error {
	if strings.TrimSpace(s.TenantID) == "" {
		return ErrTenantRequired
	}

	if strings.TrimSpace(s.Scope) == "" {
		return ErrScopeRequired
	}

	return nil
}

// Key is supplied by the client or upstream service.
type Key struct {
	Key       string
	RequestID string
}

func (k Key) Validate() error {
	if strings.TrimSpace(k.Key) == "" {
		return ErrKeyRequired
	}

	return nil
}

// Fingerprint is calculated by the server from a canonicalized request.
type Fingerprint struct {
	Algorithm               string
	Hash                    string
	CanonicalizationVersion string
	Method                  string
	Route                   string
	SchemaVersion           string
}

func (f Fingerprint) Validate() error {
	if strings.TrimSpace(f.Algorithm) == "" || strings.TrimSpace(f.Hash) == "" {
		return ErrFingerprintRequired
	}

	return nil
}

func (f Fingerprint) SameAs(other Fingerprint) bool {
	return f.Algorithm == other.Algorithm &&
		f.Hash == other.Hash &&
		f.CanonicalizationVersion == other.CanonicalizationVersion
}

// Record is the domain model stored by the idempotency store.
type Record struct {
	ID string

	Scope       Scope
	Key         Key
	Fingerprint Fingerprint

	Status Status

	Result *Result

	Owner          string
	LeaseTokenHash string
	LockedUntil    *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time

	Version int64
	Labels  map[string]string
}

func NewRecord(
	id string,
	scope Scope,
	key Key,
	fingerprint Fingerprint,
	status Status,
	now time.Time,
	expiresAt time.Time,
) (*Record, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}

	if err := key.Validate(); err != nil {
		return nil, err
	}

	if err := fingerprint.Validate(); err != nil {
		return nil, err
	}

	if !status.IsValid() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidStatus, status)
	}

	return &Record{
		ID:          id,
		Scope:       scope,
		Key:         key,
		Fingerprint: fingerprint,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   expiresAt,
		Version:     1,
		Labels:      map[string]string{},
	}, nil
}

func (r *Record) IsExpired(now time.Time) bool {
	return !r.ExpiresAt.IsZero() && !now.Before(r.ExpiresAt)
}

func (r *Record) IsLocked(now time.Time) bool {
	return r.LockedUntil != nil && now.Before(*r.LockedUntil)
}

func (r *Record) CanReplay() bool {
	return r.Status.CanReplay() && r.Result != nil
}

func (r *Record) EnsureFingerprint(fingerprint Fingerprint) error {
	if !r.Fingerprint.SameAs(fingerprint) {
		return &ConflictError{
			Scope:           r.Scope,
			Key:             r.Key,
			ExpectedHash:    r.Fingerprint.Hash,
			ActualHash:      fingerprint.Hash,
			ExpectedVersion: r.Fingerprint.CanonicalizationVersion,
			ActualVersion:   fingerprint.CanonicalizationVersion,
		}
	}

	return nil
}

func (r *Record) MarkProcessing(owner string, leaseTokenHash string, lockedUntil time.Time, now time.Time) {
	r.Status = StatusProcessing
	r.Owner = owner
	r.LeaseTokenHash = leaseTokenHash
	r.LockedUntil = &lockedUntil
	r.UpdatedAt = now
	r.Version++
}

func (r *Record) MarkCompleted(result Result, now time.Time) {
	r.Status = StatusCompleted
	r.Result = &result
	r.Owner = ""
	r.LeaseTokenHash = ""
	r.LockedUntil = nil
	r.UpdatedAt = now
	r.Version++
}

func (r *Record) MarkFailedFinal(result Result, now time.Time) {
	r.Status = StatusFailedFinal
	r.Result = &result
	r.Owner = ""
	r.LeaseTokenHash = ""
	r.LockedUntil = nil
	r.UpdatedAt = now
	r.Version++
}

func (r *Record) MarkFailedRetryable(result Result, now time.Time) {
	r.Status = StatusFailedRetryable
	r.Result = &result
	r.Owner = ""
	r.LeaseTokenHash = ""
	r.LockedUntil = nil
	r.UpdatedAt = now
	r.Version++
}

func (r *Record) MarkUnknown(now time.Time) {
	r.Status = StatusUnknown
	r.Owner = ""
	r.LeaseTokenHash = ""
	r.LockedUntil = nil
	r.UpdatedAt = now
	r.Version++
}

func (r *Record) MarkExpired(now time.Time) {
	r.Status = StatusExpired
	r.Owner = ""
	r.LeaseTokenHash = ""
	r.LockedUntil = nil
	r.UpdatedAt = now
	r.Version++
}

// ResultKind describes what exactly should be replayed.
type ResultKind string

const (
	ResultKindHTTP      ResultKind = "HTTP"
	ResultKindRPC       ResultKind = "RPC"
	ResultKindOperation ResultKind = "OPERATION"
	ResultKindResource  ResultKind = "RESOURCE"
	ResultKindError     ResultKind = "ERROR"
)

func (k ResultKind) IsValid() bool {
	switch k {
	case ResultKindHTTP,
		ResultKindRPC,
		ResultKindOperation,
		ResultKindResource,
		ResultKindError:
		return true
	default:
		return false
	}
}

// Result is a transport-neutral replay payload.
type Result struct {
	Kind ResultKind

	HTTP      *HTTPResult
	RPC       *RPCResult
	Operation *OperationResult
	Resource  *ResourceResult
	Error     *ErrorResult
}

func (r Result) Validate() error {
	if !r.Kind.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidResultKind, r.Kind)
	}

	switch r.Kind {
	case ResultKindHTTP:
		if r.HTTP == nil {
			return fmt.Errorf("%w: missing HTTP result", ErrInvalidResultKind)
		}
	case ResultKindRPC:
		if r.RPC == nil {
			return fmt.Errorf("%w: missing RPC result", ErrInvalidResultKind)
		}
	case ResultKindOperation:
		if r.Operation == nil {
			return fmt.Errorf("%w: missing operation result", ErrInvalidResultKind)
		}
	case ResultKindResource:
		if r.Resource == nil {
			return fmt.Errorf("%w: missing resource result", ErrInvalidResultKind)
		}
	case ResultKindError:
		if r.Error == nil {
			return fmt.Errorf("%w: missing error result", ErrInvalidResultKind)
		}
	}

	return nil
}

type HTTPResult struct {
	StatusCode  int
	Headers     map[string]string
	ContentType string
	Body        []byte
}

type RPCResult struct {
	// Response should be an application DTO or protobuf message.
	// The persistence adapter is responsible for serialization.
	Response any

	Code    int
	Message string
	Details []any
}

type OperationResult struct {
	OperationName      string
	TargetResourceName string
	Metadata           any
}

type ResourceResult struct {
	ResourceType string
	ResourceID   string
	ResourceName string
	Resource     any
}

type ErrorResult struct {
	Code    string
	Message string
	Details any
}
