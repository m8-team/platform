package id

import "errors"

var (
	ErrEmptyID   = errors.New("id is empty")
	ErrInvalidID = errors.New("id is invalid")
)
