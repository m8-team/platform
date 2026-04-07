package common

import "errors"

var (
	ErrDuplicateRequest = errors.New("duplicate request")
	ErrInvalidMask      = errors.New("invalid update mask")
)
