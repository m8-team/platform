package organization

import "errors"

var (
	ErrEmptyOrganizationID      = errors.New("organization id is empty")
	ErrInvalidOrganizationID    = errors.New("organization id is invalid")
	ErrInvalidOrganizationState = errors.New("organization state is invalid")
	ErrEmptyTime                = errors.New("time is empty")

	ErrOrganizationNameTooLong        = errors.New("organization name is too long")
	ErrOrganizationDescriptionTooLong = errors.New("organization description is too long")
)
