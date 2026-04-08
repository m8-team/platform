package organization

import "errors"

var (
	ErrNotFound           = errors.New("organization not found")
	ErrInvalidID          = errors.New("organization id must be a valid uuid")
	ErrImmutableID        = errors.New("organization id is immutable")
	ErrInvalidState       = errors.New("organization state is invalid")
	ErrDeleted            = errors.New("organization is soft-deleted")
	ErrAlreadyDeleted     = errors.New("organization is already deleted")
	ErrNotDeleted         = errors.New("organization is not deleted")
	ErrETagMismatch       = errors.New("organization etag mismatch")
	ErrInvalidUpdatePath  = errors.New("organization update path is not allowed")
	ErrInvalidPurgeWindow = errors.New("organization purge time must be after delete time")
)
