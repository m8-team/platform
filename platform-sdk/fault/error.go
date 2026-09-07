// Package fault defines safe application errors without a transport dependency.
package fault

import "fmt"

type Code string

const (
	InvalidArgument    Code = "INVALID_ARGUMENT"
	NotFound           Code = "NOT_FOUND"
	AlreadyExists      Code = "ALREADY_EXISTS"
	PermissionDenied   Code = "PERMISSION_DENIED"
	Unauthenticated    Code = "UNAUTHENTICATED"
	Conflict           Code = "CONFLICT"
	FailedPrecondition Code = "FAILED_PRECONDITION"
	ResourceExhausted  Code = "RESOURCE_EXHAUSTED"
	Unavailable        Code = "UNAVAILABLE"
	Internal           Code = "INTERNAL"
)

type Violation struct{ Field, Description string }

// Message and Violations are explicitly client-safe. Cause is never serialized.
// Retryability belongs to the caller's operation, not a blanket error flag.
type Error struct {
	Code       Code
	Message    string
	Cause      error
	Violations []Violation
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }
func (e *Error) Unwrap() error { return e.Cause }
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t != nil && e != nil && t.Code == e.Code
}
func New(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}
