package workspace

import "errors"

var (
	ErrNotFound           = errors.New("workspace not found")
	ErrInvalidID          = errors.New("workspace id must be a valid uuid")
	ErrInvalidParentID    = errors.New("workspace organization id must be a valid uuid")
	ErrInvalidState       = errors.New("workspace state is invalid")
	ErrImmutableID        = errors.New("workspace id is immutable")
	ErrImmutableParent    = errors.New("workspace organization_id is immutable")
	ErrETagMismatch       = errors.New("workspace etag mismatch")
	ErrDeleted            = errors.New("workspace is soft-deleted")
	ErrAlreadyDeleted     = errors.New("workspace is already deleted")
	ErrNotDeleted         = errors.New("workspace is not deleted")
	ErrInvalidUpdatePath  = errors.New("workspace update path is not allowed")
	ErrInvalidPurgeWindow = errors.New("workspace purge time must be after delete time")
)
