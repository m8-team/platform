// Package scope defines placement rules for the proposed Resource Manager v2
// hierarchy. It deliberately does not perform authorization or persistence.
package scope

import (
	"errors"
	"fmt"
)

// Kind is a canonical resource kind used for containment validation.
type Kind string

const (
	KindGroup        Kind = "group"
	KindOrganization Kind = "organization"
	KindWorkspace    Kind = "workspace"
	KindProject      Kind = "project"
	KindService      Kind = "service"
	KindResource     Kind = "resource"
)

var ErrInvalidPlacement = errors.New("invalid resource placement")

// Parent is a typed reference to a resource's immediate owner.
type Parent struct {
	Kind Kind
	ID   string
}

// AllowedParent enforces containment without inferring access rights.
// Group affiliation is a separate relationship, not a containment parent.
func AllowedParent(child, parent Kind) bool {
	switch child {
	case KindGroup, KindOrganization:
		return parent == ""
	case KindWorkspace:
		return parent == KindOrganization
	case KindProject:
		return parent == KindOrganization || parent == KindWorkspace
	case KindService:
		return parent == KindWorkspace || parent == KindProject
	case KindResource:
		return parent == KindOrganization || parent == KindWorkspace || parent == KindProject
	default:
		return false
	}
}

// ValidatePlacement checks the proposed v2 parent-kind matrix. API adapters
// must additionally parse IDs and verify parent existence, state, and policy.
func ValidatePlacement(child Kind, parent Parent) error {
	if !AllowedParent(child, parent.Kind) {
		return fmt.Errorf("%w: %s cannot have parent %s", ErrInvalidPlacement, child, parent.Kind)
	}
	if parent.Kind == "" && parent.ID != "" || parent.Kind != "" && parent.ID == "" {
		return fmt.Errorf("%w: parent kind and ID must be supplied together", ErrInvalidPlacement)
	}
	return nil
}
