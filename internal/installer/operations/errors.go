package operations

import (
	"errors"
	"fmt"
)

var (
	ErrTransient = errors.New("transient operation error")
	ErrPermanent = errors.New("permanent operation error")
)

type Error struct {
	Code        string
	Message     string
	Transient   bool
	Cause       error
	Remediation string
}

func (e Error) Error() string {
	if e.Cause == nil {
		return e.Code + ": " + e.Message
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e Error) Unwrap() error {
	return e.Cause
}

func (e Error) Is(target error) bool {
	if target == ErrTransient {
		return e.Transient
	}
	if target == ErrPermanent {
		return !e.Transient
	}
	return false
}

func Transient(code string, message string, cause error) Error {
	return Error{Code: code, Message: message, Transient: true, Cause: cause}
}

func Permanent(code string, message string, cause error) Error {
	return Error{Code: code, Message: message, Transient: false, Cause: cause}
}
