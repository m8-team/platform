package domain

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidPermissionName       = errors.New("invalid permission name")
	ErrInvalidPermissionDefinition = errors.New("invalid permission definition")
)

var permissionNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,127}$`)

// Permission is a named authorization action. It is deliberately independent
// of a resource's owner or any particular authorization engine.
type Permission struct{ name string }

func NewPermission(name string) (Permission, error) {
	p := Permission{name: strings.TrimSpace(name)}
	return p, p.Validate()
}

func (p Permission) Validate() error {
	if p.name == "" {
		return ErrPermissionRequired
	}
	if !permissionNamePattern.MatchString(p.name) {
		return ErrInvalidPermissionName
	}
	return nil
}

func (p Permission) Name() string { return p.name }

// PermissionDefinition is the versioned registry mapping a public action to an
// authorization graph permission. Its publication belongs to the Control Plane.
type PermissionDefinition struct {
	permission      Permission
	resourceType    string
	graphPermission string
}

func NewPermissionDefinition(permission Permission, resourceType, graphPermission string) (PermissionDefinition, error) {
	d := PermissionDefinition{permission: permission, resourceType: strings.TrimSpace(resourceType), graphPermission: strings.TrimSpace(graphPermission)}
	return d, d.Validate()
}

func (d PermissionDefinition) Validate() error {
	if err := d.permission.Validate(); err != nil {
		return err
	}
	if !relationNamePattern.MatchString(d.resourceType) || !relationNamePattern.MatchString(d.graphPermission) {
		return ErrInvalidPermissionDefinition
	}
	return nil
}

func (d PermissionDefinition) Permission() Permission  { return d.permission }
func (d PermissionDefinition) ResourceType() string    { return d.resourceType }
func (d PermissionDefinition) GraphPermission() string { return d.graphPermission }
