package hierarchy

import "errors"

var (
	ErrParentNotFound        = errors.New("parent resource not found")
	ErrParentDeleted         = errors.New("parent resource is soft-deleted")
	ErrDeleteBlocked         = errors.New("delete is blocked by active child resources")
	ErrUndeleteParentInvalid = errors.New("undelete is blocked because parent resource is missing or deleted")
)
