package project

import "errors"

var (
	ErrNotFound           = errors.New("project not found")
	ErrInvalidID          = errors.New("project id must be a valid uuid")
	ErrInvalidParentID    = errors.New("project workspace id must be a valid uuid")
	ErrImmutableID        = errors.New("project id is immutable")
	ErrImmutableParent    = errors.New("project workspace_id is immutable")
	ErrInvalidState       = errors.New("project state is invalid")
	ErrDeleted            = errors.New("project is soft-deleted")
	ErrAlreadyDeleted     = errors.New("project is already deleted")
	ErrNotDeleted         = errors.New("project is not deleted")
	ErrETagMismatch       = errors.New("project etag mismatch")
	ErrInvalidUpdatePath  = errors.New("project update path is not allowed")
	ErrInvalidPurgeWindow = errors.New("project purge time must be after delete time")
)
