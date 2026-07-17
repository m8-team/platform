package domain

import "strings"

type Permission struct {
	name string
}

func NewPermission(name string) (Permission, error) {
	permission := Permission{name: strings.TrimSpace(name)}
	if err := permission.Validate(); err != nil {
		return Permission{}, err
	}

	return permission, nil
}

func (p Permission) Validate() error {
	if p.name == "" {
		return ErrPermissionRequired
	}

	return nil
}

func (p Permission) Name() string {
	return p.name
}

type ModelRevision struct {
	value string
}

func NewModelRevision(value string) (ModelRevision, error) {
	revision := ModelRevision{value: strings.TrimSpace(value)}
	if err := revision.Validate(); err != nil {
		return ModelRevision{}, err
	}

	return revision, nil
}

func (r ModelRevision) Validate() error {
	if r.value == "" {
		return ErrModelRevisionRequired
	}

	return nil
}

func (r ModelRevision) String() string {
	return r.value
}

func (r ModelRevision) Equal(other ModelRevision) bool {
	return r.value == other.value
}
