// Package permission owns the public action vocabulary and publishable authorization model.
package permission

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrPermissionRequired    = errors.New("permission required")
	ErrInvalidPermissionName = errors.New("invalid permission name")
	ErrDuplicatePermission   = errors.New("duplicate permission")
)
var permissionPattern = regexp.MustCompile("^[a-z][a-z0-9_.:-]{0,127}$")

// Permission is a business action, not a SpiceDB relation.
type Permission struct{ name string }

func NewPermission(name string) (Permission, error) {
	p := Permission{name: strings.TrimSpace(name)}
	return p, p.Validate()
}
func (p Permission) Validate() error {
	if p.name == "" {
		return ErrPermissionRequired
	}
	if !permissionPattern.MatchString(p.name) {
		return ErrInvalidPermissionName
	}
	return nil
}
func (p Permission) Name() string { return p.name }

// ValidateSet rejects duplicate permission names and optionally an empty set.
func ValidateSet(permissions []Permission, requireNonEmpty bool) error {
	if requireNonEmpty && len(permissions) == 0 {
		return ErrPermissionRequired
	}
	found := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		if err := p.Validate(); err != nil {
			return err
		}
		if _, ok := found[p.Name()]; ok {
			return ErrDuplicatePermission
		}
		found[p.Name()] = struct{}{}
	}
	return nil
}
