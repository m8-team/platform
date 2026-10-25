// Package scope defines pure domain placement rules for the canonical v1 model.
// No access control, storage or remote provider calls belong in this package.
package scope

import (
	"errors"
	"fmt"
)

type Kind string

const (
	KindTenant       Kind = "tenant"
	KindGroup        Kind = "group"
	KindOrganization Kind = "organization"
	KindWorkspace    Kind = "workspace"
	KindProject      Kind = "project"
	KindEnvironment  Kind = "environment"
	KindWorkload     Kind = "workload_service"
	KindDescriptor   Kind = "resource_descriptor"
)

var ErrInvalidPlacement = errors.New("invalid resource placement")

type Parent struct {
	Kind Kind
	ID   string
}

// AllowedParent validates only containment shape. Group membership and
// Environment bindings are separate edges and never grant permissions.
func AllowedParent(child, parent Kind) bool {
	switch child {
	case KindTenant:
		return parent == ""
	case KindGroup, KindOrganization:
		return parent == KindTenant
	case KindWorkspace, KindEnvironment:
		return parent == KindOrganization
	case KindProject:
		return parent == KindOrganization || parent == KindWorkspace
	case KindWorkload:
		return parent == KindWorkspace || parent == KindProject
	case KindDescriptor:
		return parent == KindOrganization || parent == KindWorkspace || parent == KindProject
	default:
		return false
	}
}

// ValidatePlacement checks the parent-kind matrix. Adapters/application must
// also parse IDs, resolve parent lifecycle/tenant and run authorization.
func ValidatePlacement(child Kind, parent Parent) error {
	if !AllowedParent(child, parent.Kind) {
		return fmt.Errorf("%w: %s cannot belong to %s", ErrInvalidPlacement, child, parent.Kind)
	}
	if parent.Kind == "" && parent.ID != "" || parent.Kind != "" && parent.ID == "" {
		return fmt.Errorf("%w: parent kind and id must be supplied together", ErrInvalidPlacement)
	}
	return nil
}
