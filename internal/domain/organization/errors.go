package organization

import "errors"

var (
	ErrNotFound           = errors.New("organization not found")
	ErrInvalidID          = errors.New("organization id must be a valid uuid")
	ErrInvalidState       = errors.New("organization state is invalid")
	ErrImmutableID        = errors.New("organization id is immutable")
	ErrETagMismatch       = errors.New("organization etag mismatch")
	ErrDeleted            = errors.New("organization is soft-deleted")
	ErrAlreadyDeleted     = errors.New("organization is already deleted")
	ErrNotDeleted         = errors.New("organization is not deleted")
	ErrInvalidUpdatePath  = errors.New("organization update path is not allowed")
	ErrMissingDeleteTime  = errors.New("organization delete time is missing")
	ErrInvalidPurgeWindow = errors.New("organization purge time must be after delete time")
)
