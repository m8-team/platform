package organization

import "errors"

var (
	ErrEmptyOrganizationID      = errors.New("organization id is empty")
	ErrInvalidOrganizationID    = errors.New("organization id is invalid")
	ErrInvalidOrganizationState = errors.New("organization state is invalid")
	ErrEmptyTime                = errors.New("time is empty")
	ErrInvalidVersion           = errors.New("version is invalid")
	ErrVersionConflict          = errors.New("organization version conflict")

	ErrOrganizationNameTooLong        = errors.New("organization name is too long")
	ErrOrganizationDescriptionTooLong = errors.New("organization description is too long")

	ErrTooManyLabels     = errors.New("too many labels")
	ErrEmptyLabelKey     = errors.New("label key is empty")
	ErrLabelKeyTooLong   = errors.New("label key is too long")
	ErrInvalidLabelKey   = errors.New("label key is invalid")
	ErrLabelValueTooLong = errors.New("label value is too long")

	ErrInvalidStateTransition        = errors.New("invalid organization state transition")
	ErrOrganizationCannotBeUpdated   = errors.New("organization cannot be updated in current state")
	ErrOrganizationCannotBeDeleted   = errors.New("organization cannot be deleted in current state")
	ErrOrganizationCannotBeUndeleted = errors.New("organization cannot be undeleted in current state")
	ErrOrganizationCannotBeSuspended = errors.New("organization cannot be suspended in current state")
	ErrOrganizationCannotBeResumed   = errors.New("organization cannot be resumed in current state")
)
