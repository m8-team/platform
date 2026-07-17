package domain

import (
	"errors"
	"fmt"
)

var (
	ErrRecordNotFound      = errors.New("idempotency record not found")
	ErrRecordAlreadyExists = errors.New("idempotency record already exists")

	ErrKeyRequired         = errors.New("idempotency key is required")
	ErrScopeRequired       = errors.New("idempotency scope is required")
	ErrTenantRequired      = errors.New("tenant id is required")
	ErrFingerprintRequired = errors.New("request fingerprint is required")

	ErrConflict          = errors.New("idempotency conflict")
	ErrInProgress        = errors.New("idempotency request is already in progress")
	ErrExpired           = errors.New("idempotency key expired")
	ErrLeaseRequired     = errors.New("idempotency lease token is required")
	ErrLeaseLost         = errors.New("idempotency lease lost")
	ErrInvalidState      = errors.New("invalid idempotency state")
	ErrInvalidTransition = errors.New("invalid idempotency transition")

	ErrInvalidStatus       = errors.New("invalid idempotency status")
	ErrInvalidDecision     = errors.New("invalid idempotency decision")
	ErrInvalidReplayPolicy = errors.New("invalid idempotency replay policy")
	ErrInvalidResultKind   = errors.New("invalid idempotency result kind")
)

type ConflictError struct {
	Scope           Scope
	Key             Key
	ExpectedHash    string
	ActualHash      string
	ExpectedVersion string
	ActualVersion   string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf(
		"%v: scope=%s key=%s expected_hash=%s actual_hash=%s",
		ErrConflict,
		e.Scope.Identity(),
		e.Key.Key,
		e.ExpectedHash,
		e.ActualHash,
	)
}

func (e *ConflictError) Unwrap() error {
	return ErrConflict
}

type InProgressError struct {
	Scope      Scope
	Key        Key
	RetryAfter string
}

func (e *InProgressError) Error() string {
	return fmt.Sprintf(
		"%v: scope=%s key=%s retry_after=%s",
		ErrInProgress,
		e.Scope.Identity(),
		e.Key.Key,
		e.RetryAfter,
	)
}

func (e *InProgressError) Unwrap() error {
	return ErrInProgress
}

type LeaseError struct {
	Scope Scope
	Key   Key
	Cause error
}

func (e *LeaseError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%v: scope=%s key=%s", ErrLeaseLost, e.Scope.Identity(), e.Key.Key)
	}

	return fmt.Sprintf("%v: scope=%s key=%s cause=%v", ErrLeaseLost, e.Scope.Identity(), e.Key.Key, e.Cause)
}

func (e *LeaseError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}

	return ErrLeaseLost
}

type StateError struct {
	Scope  Scope
	Key    Key
	Status Status
	Action string
}

func (e *StateError) Error() string {
	return fmt.Sprintf(
		"%v: scope=%s key=%s status=%s action=%s",
		ErrInvalidState,
		e.Scope.Identity(),
		e.Key.Key,
		e.Status,
		e.Action,
	)
}

func (e *StateError) Unwrap() error {
	return ErrInvalidState
}
