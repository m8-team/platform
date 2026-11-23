package authorization

import (
	"errors"
	"github.com/m8-team/platform/internal/access/domain/shared"
	"strings"
)

var (
	ErrInvalidScope       = errors.New("invalid authorization scope")
	ErrUnverifiedResource = errors.New("resource ownership must be resolved by its authoritative owner")
)

// ScopeKind is an entitlement boundary, not a Resource Manager OwnerScope.
// Project is deliberately not a Resource Manager owner scope.
type ScopeKind string

const (
	ScopeInstance     ScopeKind = "INSTANCE"
	ScopeCollection   ScopeKind = "COLLECTION"
	ScopeOrganization ScopeKind = "ORGANIZATION"
	ScopeSubtree      ScopeKind = "SUBTREE"
	ScopeResourceKind ScopeKind = "RESOURCE_KIND"
	ScopeTenant       ScopeKind = "TENANT"
)

func (k ScopeKind) Valid() bool {
	switch k {
	case ScopeInstance, ScopeCollection, ScopeOrganization, ScopeSubtree, ScopeResourceKind, ScopeTenant:
		return true
	default:
		return false
	}
}

// Scope requires an explicit resource type even for TENANT, so publishing
// a new module can never silently expand an existing tenant-wide assignment.
type Scope struct {
	kind           ScopeKind
	tenantID       shared.TenantID
	resourceType   string
	resourceID     string
	organizationID string
	collectionIDs  []string
	root           shared.Resource
}
type ScopeInput struct {
	Kind           ScopeKind
	TenantID       shared.TenantID
	ResourceType   string
	ResourceID     string
	OrganizationID string
	CollectionIDs  []string
	Root           shared.Resource
}

func NewScope(input ScopeInput) (Scope, error) {
	s := Scope{kind: input.Kind, tenantID: input.TenantID, resourceType: strings.TrimSpace(input.ResourceType),
		resourceID: strings.TrimSpace(input.ResourceID), organizationID: strings.TrimSpace(input.OrganizationID),
		collectionIDs: append([]string(nil), input.CollectionIDs...), root: input.Root}
	return s, s.Validate()
}
func (s Scope) Validate() error {
	if !s.kind.Valid() || s.tenantID.Validate() != nil || !shared.ValidQualifiedName(s.resourceType) {
		return ErrInvalidScope
	}
	switch s.kind {
	case ScopeInstance:
		if s.resourceID == "" || s.organizationID != "" || len(s.collectionIDs) != 0 || s.root.Type() != "" {
			return ErrInvalidScope
		}
	case ScopeCollection:
		if s.resourceID != "" || s.organizationID != "" || s.root.Type() != "" || len(s.collectionIDs) == 0 || len(s.collectionIDs) > 1000 {
			return ErrInvalidScope
		}
		seen := make(map[string]bool, len(s.collectionIDs))
		for _, id := range s.collectionIDs {
			if id == "" || strings.TrimSpace(id) != id || len(id) > 256 || seen[id] {
				return ErrInvalidScope
			}
			seen[id] = true
		}
	case ScopeOrganization:
		if !shared.ValidUUID(s.organizationID) || s.resourceID != "" || len(s.collectionIDs) != 0 || s.root.Type() != "" {
			return ErrInvalidScope
		}
	case ScopeSubtree:
		if s.root.Validate() != nil || s.resourceID != "" || s.organizationID != "" || len(s.collectionIDs) != 0 {
			return ErrInvalidScope
		}
	case ScopeResourceKind, ScopeTenant:
		if s.resourceID != "" || s.organizationID != "" || len(s.collectionIDs) != 0 || s.root.Type() != "" {
			return ErrInvalidScope
		}
	}
	if len(s.resourceID) > 256 {
		return ErrInvalidScope
	}
	return nil
}
func (s Scope) Kind() ScopeKind           { return s.kind }
func (s Scope) TenantID() shared.TenantID { return s.tenantID }
func (s Scope) ResourceType() string      { return s.resourceType }
func (s Scope) ResourceID() string        { return s.resourceID }
func (s Scope) OrganizationID() string    { return s.organizationID }
func (s Scope) CollectionIDs() []string   { return append([]string(nil), s.collectionIDs...) }
func (s Scope) Root() shared.Resource     { return s.root }

// ResolvedResource is evidence obtained from Resource Manager/the authoritative
// module. Never construct it from caller-supplied claims or unverified headers.
type ResolvedResource struct {
	TenantID       shared.TenantID
	Resource       shared.Resource
	OrganizationID string
	Ancestors      []shared.Resource
	SourceRevision int64
}

func (r ResolvedResource) Validate() error {
	if r.TenantID.Validate() != nil || r.Resource.Validate() != nil || !shared.ValidQualifiedName(r.Resource.Type()) || r.SourceRevision < 1 {
		return ErrUnverifiedResource
	}
	if r.OrganizationID != "" && !shared.ValidUUID(r.OrganizationID) {
		return ErrUnverifiedResource
	}
	for _, a := range r.Ancestors {
		if a.Validate() != nil {
			return ErrUnverifiedResource
		}
	}
	return nil
}

// Matches returns false for cross-tenant or cross-resource-type checks.
// Caller MUST resolve ownership/lifecycle freshly before creating evidence.
func (s Scope) Matches(r ResolvedResource) (bool, error) {
	if err := s.Validate(); err != nil {
		return false, err
	}
	if err := r.Validate(); err != nil {
		return false, err
	}
	if s.tenantID != r.TenantID || s.resourceType != r.Resource.Type() {
		return false, nil
	}
	switch s.kind {
	case ScopeInstance:
		return s.resourceID == r.Resource.ID(), nil
	case ScopeCollection:
		for _, id := range s.collectionIDs {
			if id == r.Resource.ID() {
				return true, nil
			}
		}
		return false, nil
	case ScopeOrganization:
		return s.organizationID == r.OrganizationID, nil
	case ScopeSubtree:
		if s.root == r.Resource {
			return true, nil
		}
		for _, a := range r.Ancestors {
			if a == s.root {
				return true, nil
			}
		}
		return false, nil
	case ScopeResourceKind, ScopeTenant:
		return true, nil
	}
	return false, ErrInvalidScope
}
