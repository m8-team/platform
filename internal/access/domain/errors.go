package domain

import "errors"

var (
	ErrSubjectTypeRequired       = errors.New("subject type is required")
	ErrSubjectIDRequired         = errors.New("subject id is required")
	ErrPermissionRequired        = errors.New("permission is required")
	ErrResourceTypeRequired      = errors.New("resource type is required")
	ErrResourceIDRequired        = errors.New("resource id is required")
	ErrModelRevisionRequired     = errors.New("authorization model revision is required")
	ErrInvalidDecision           = errors.New("invalid access decision")
	ErrInvalidFailMode           = errors.New("invalid access fail mode")
	ErrInvalidEngineFailureKind  = errors.New("invalid permission engine failure kind")
	ErrFailOpenReferenceRequired = errors.New("fail-open reference is required")
	ErrFailOpenForCriticalCheck  = errors.New("fail-open is not allowed for critical permission checks")
	ErrModelRevisionMismatch     = errors.New("permission decision model revision does not match request")
)
