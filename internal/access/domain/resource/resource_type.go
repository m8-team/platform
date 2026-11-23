// Package resource describes protected resource *types*, never owns instance lifecycle.
package resource

import (
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidResourceType = errors.New("invalid module resource type")

// TypeDefinition is published by the owning module in its own namespace.
type TypeDefinition struct {
	moduleID string
	name     string
	scopes   []authorization.ScopeKind
}

func NewTypeDefinition(moduleID, name string, scopes []authorization.ScopeKind) (TypeDefinition, error) {
	d := TypeDefinition{moduleID: strings.TrimSpace(moduleID), name: strings.TrimSpace(name), scopes: append([]authorization.ScopeKind(nil), scopes...)}
	return d, d.Validate()
}
func (d TypeDefinition) Validate() error {
	if !shared.ValidModuleName(d.moduleID) || !shared.ValidQualifiedName(d.name) || !strings.HasPrefix(d.name, d.moduleID+".") || len(d.scopes) == 0 || len(d.scopes) > 8 {
		return ErrInvalidResourceType
	}
	seen := map[authorization.ScopeKind]bool{}
	for _, s := range d.scopes {
		if !s.Valid() || seen[s] {
			return ErrInvalidResourceType
		}
		seen[s] = true
	}
	return nil
}
func (d TypeDefinition) ModuleID() string { return d.moduleID }
func (d TypeDefinition) Name() string     { return d.name }
func (d TypeDefinition) Scopes() []authorization.ScopeKind {
	return append([]authorization.ScopeKind(nil), d.scopes...)
}
func (d TypeDefinition) Supports(scope authorization.ScopeKind) bool {
	for _, s := range d.scopes {
		if s == scope {
			return true
		}
	}
	return false
}
