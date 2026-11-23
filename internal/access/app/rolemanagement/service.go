// Package rolemanagement verifies universal profiles and scoped assignments
// before asking the authoritative repository to record non-effective requests.
package rolemanagement

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrUnavailable     = errors.New("role management unavailable")
	ErrUnauthenticated = errors.New("authenticated role management principal required")
	ErrForbidden       = errors.New("role management operation denied")
	ErrUnresolvedRole  = errors.New("role template missing or invalid")
	ErrUnverifiedScope = errors.New("unverified assignment scope")
	ErrInvalidProfile  = errors.New("invalid or unsatisfied composite role profile")
)

type Identity interface {
	AuthenticatedPrincipal(context.Context) (shared.Subject, error)
}
type Authorizer interface {
	CanPublishProfile(context.Context, shared.Subject, role.BundleRef) (bool, error)
	// CanManageTenant is checked before resolving an assignment ID to avoid
	// revealing whether another tenant's assignment exists.
	CanManageTenant(context.Context, shared.Subject, shared.TenantID) (bool, error)
	CanManageSubject(context.Context, shared.Subject, shared.TenantID, shared.Subject) (bool, error)
	CanReadSubject(context.Context, shared.Subject, shared.TenantID, shared.Subject) (bool, error)
}

// VerifiedPermission comes from a trusted published module catalog, not from
// an untrusted request or an arbitrary resourced owner-provided header.
type VerifiedPermission struct {
	Action       permission.Permission
	ResourceType string
	Scopes       []authorization.ScopeKind
}
type VerifiedTemplate struct {
	Template    role.Template
	Permissions []VerifiedPermission
}
type TemplateResolver interface {
	ResolveTemplate(context.Context, role.TemplateRef) (VerifiedTemplate, error)
}

// OwnerScopeVerifier obtains authoritative resource ownership and revision.
// This is a trusted port: Access has NO direct Resource Manager dependency.
type OwnerScopeVerifier interface {
	VerifyScope(context.Context, authorization.Scope) (int64, error)
}
type BundleRecord struct {
	Bundle   role.Bundle
	Revision int64
}
type BundleRepository interface {
	GetBundle(context.Context, role.BundleRef) (BundleRecord, error)
	PublishBundle(context.Context, role.Bundle, int64, string) (BundleRecord, error)
}
type AssignmentRepository interface {
	CreateAssignment(context.Context, role.ProfileAssignment, string) (role.ProfileAssignment, error)
	GetAssignment(context.Context, shared.TenantID, string) (role.ProfileAssignment, error)
	RevokeAssignment(context.Context, shared.TenantID, string, int64, string) (role.ProfileAssignment, error)
	ListBySubject(context.Context, shared.TenantID, shared.Subject, string, int) ([]role.ProfileAssignment, error)
}
type Service struct {
	identity    Identity
	auth        Authorizer
	templates   TemplateResolver
	owner       OwnerScopeVerifier
	bundles     BundleRepository
	assignments AssignmentRepository
}

func New(identity Identity, auth Authorizer, templates TemplateResolver, owner OwnerScopeVerifier, bundles BundleRepository, assignments AssignmentRepository) (*Service, error) {
	if identity == nil || auth == nil || templates == nil || owner == nil || bundles == nil || assignments == nil {
		return nil, ErrUnavailable
	}
	return &Service{identity: identity, auth: auth, templates: templates, owner: owner, bundles: bundles, assignments: assignments}, nil
}
func (s *Service) principal(ctx context.Context) (shared.Subject, error) {
	if s == nil || s.identity == nil || s.auth == nil {
		return shared.Subject{}, ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return shared.Subject{}, err
	}
	actor, err := s.identity.AuthenticatedPrincipal(ctx)
	if err != nil || actor.Validate() != nil {
		return shared.Subject{}, ErrUnauthenticated
	}
	return actor, nil
}
func (s *Service) checkManagement(ctx context.Context, tenant shared.TenantID, subject shared.Subject) (shared.Subject, error) {
	if tenant.Validate() != nil || subject.Validate() != nil {
		return shared.Subject{}, ErrInvalidProfile
	}
	actor, err := s.principal(ctx)
	if err != nil {
		return shared.Subject{}, err
	}
	ok, err := s.auth.CanManageSubject(ctx, actor, tenant, subject)
	if err != nil || !ok {
		return shared.Subject{}, ErrForbidden
	}
	return actor, nil
}

func validateResolved(ref role.TemplateRef, got VerifiedTemplate) error {
	if ref.Validate() != nil || got.Template.Validate() != nil || got.Template.Ref() != ref ||
		len(got.Permissions) != len(got.Template.Permissions()) {
		return ErrUnresolvedRole
	}
	byAction := make(map[string]VerifiedPermission, len(got.Permissions))
	for _, p := range got.Permissions {
		if p.Action.Validate() != nil || !shared.ValidQualifiedName(p.ResourceType) ||
			len(p.Scopes) == 0 || byAction[p.Action.Name()].ResourceType != "" {
			return ErrUnresolvedRole
		}
		seen := map[authorization.ScopeKind]bool{}
		for _, scope := range p.Scopes {
			if !scope.Valid() || seen[scope] {
				return ErrUnresolvedRole
			}
			seen[scope] = true
		}
		byAction[p.Action.Name()] = p
	}
	for _, p := range got.Template.Permissions() {
		definition, ok := byAction[p.Name()]
		if !ok {
			return ErrUnresolvedRole
		}
		for _, scope := range got.Template.AllowedScopes() {
			allowed := false
			for _, candidate := range definition.Scopes {
				if candidate == scope {
					allowed = true
				}
			}
			if !allowed {
				return ErrUnresolvedRole
			}
		}
	}
	return nil
}
func (s *Service) resolveTemplates(ctx context.Context, refs []role.TemplateRef) ([]VerifiedTemplate, error) {
	var out []VerifiedTemplate
	for _, ref := range refs {
		verified, err := s.templates.ResolveTemplate(ctx, ref)
		if err != nil || validateResolved(ref, verified) != nil {
			return nil, ErrUnresolvedRole
		}
		out = append(out, verified)
	}
	return out, nil
}
func (s *Service) collectBundles(ctx context.Context, ref role.BundleRef, depth int, index map[string]role.Bundle, visiting map[string]bool) error {
	if depth > 32 || len(index) > 1000 {
		return ErrInvalidProfile
	}
	if visiting[ref.Key()] {
		return role.ErrCyclicBundle
	}
	if _, ok := index[ref.Key()]; ok {
		return nil
	}
	visiting[ref.Key()] = true
	rec, err := s.bundles.GetBundle(ctx, ref)
	if err != nil || rec.Bundle.Validate() != nil || rec.Bundle.Ref() != ref {
		return ErrInvalidProfile
	}
	for _, child := range rec.Bundle.Includes() {
		if err := s.collectBundles(ctx, child, depth+1, index, visiting); err != nil {
			return err
		}
	}
	visiting[ref.Key()] = false
	index[ref.Key()] = rec.Bundle
	return nil
}
func (s *Service) expandPublished(ctx context.Context, ref role.BundleRef) ([]role.TemplateRef, error) {
	if ref.Validate() != nil {
		return nil, ErrInvalidProfile
	}
	index := map[string]role.Bundle{}
	if err := s.collectBundles(ctx, ref, 0, index, map[string]bool{}); err != nil {
		return nil, err
	}
	all := make([]role.Bundle, 0, len(index))
	for _, bundle := range index {
		all = append(all, bundle)
	}
	refs, err := role.ExpandProfiles(ref, all)
	if err != nil {
		return nil, err
	}
	return refs, nil
}

// PublishBundle rejects cycles and missing pinned nested/template versions
// before committing. Published dependencies are immutable; a profile edit
// creates a new version rather than changing existing entitlements.
func (s *Service) PublishBundle(ctx context.Context, b role.Bundle, expected int64, key string) (BundleRecord, error) {
	if b.Validate() != nil || expected < 0 {
		return BundleRecord{}, ErrInvalidProfile
	}
	actor, err := s.principal(ctx)
	if err != nil {
		return BundleRecord{}, err
	}
	allowed, err := s.auth.CanPublishProfile(ctx, actor, b.Ref())
	if err != nil || !allowed {
		return BundleRecord{}, ErrForbidden
	}
	index := map[string]role.Bundle{b.Ref().Key(): b}
	for _, child := range b.Includes() {
		if err := s.collectBundles(ctx, child, 0, index, map[string]bool{b.Ref().Key(): true}); err != nil {
			return BundleRecord{}, err
		}
	}
	all := make([]role.Bundle, 0, len(index))
	for _, item := range index {
		all = append(all, item)
	}
	refs, err := role.ExpandProfiles(b.Ref(), all)
	if err != nil {
		return BundleRecord{}, err
	}
	if _, err := s.resolveTemplates(ctx, refs); err != nil {
		return BundleRecord{}, err
	}
	return s.bundles.PublishBundle(ctx, b, expected, key)
}

// CreateAssignment checks ALL role permissions against typed per-resource
// scopes and trusted owner proof. It persists only PENDING_PROJECTION records.
func (s *Service) CreateAssignment(ctx context.Context, in role.ProfileAssignmentInput, key string) (role.ProfileAssignment, error) {
	if _, err := s.checkManagement(ctx, in.TenantID, in.Subject); err != nil {
		return role.ProfileAssignment{}, err
	}
	a, err := role.NewProfileAssignment(in)
	if err != nil {
		return role.ProfileAssignment{}, err
	}
	var refs []role.TemplateRef
	switch in.Target.Kind {
	case role.AssignmentTemplate:
		refs = []role.TemplateRef{{Name: in.Target.Name, Version: in.Target.Version}}
	case role.AssignmentBundle:
		refs, err = s.expandPublished(ctx, role.BundleRef{Name: in.Target.Name, Version: in.Target.Version})
		if err != nil {
			return role.ProfileAssignment{}, err
		}
	default:
		return role.ProfileAssignment{}, ErrInvalidProfile
	}
	resolved, err := s.resolveTemplates(ctx, refs)
	if err != nil {
		return role.ProfileAssignment{}, err
	}
	scopes := make(map[string]authorization.Scope, len(a.Scopes()))
	for _, scope := range a.Scopes() {
		proof, err := s.owner.VerifyScope(ctx, scope)
		if err != nil || proof < 1 {
			return role.ProfileAssignment{}, ErrUnverifiedScope
		}
		scopes[scope.ResourceType()] = scope
	}
	used := make(map[string]bool)
	for _, template := range resolved {
		for _, p := range template.Permissions {
			scope, ok := scopes[p.ResourceType]
			if !ok {
				return role.ProfileAssignment{}, ErrInvalidProfile
			}
			allowed := false
			for _, kind := range p.Scopes {
				if scope.Kind() == kind {
					allowed = true
				}
			}
			if !allowed || !template.Template.SupportsScope(scope.Kind()) {
				return role.ProfileAssignment{}, ErrInvalidProfile
			}
			used[p.ResourceType] = true
		}
	}
	if len(used) != len(scopes) {
		return role.ProfileAssignment{}, ErrInvalidProfile
	}
	return s.assignments.CreateAssignment(ctx, a, key)
}
func (s *Service) RevokeAssignment(ctx context.Context, tenant shared.TenantID, id string, expected int64, key string) (role.ProfileAssignment, error) {
	if tenant.Validate() != nil || !shared.ValidUUID(id) {
		return role.ProfileAssignment{}, ErrInvalidProfile
	}
	actor, err := s.principal(ctx)
	if err != nil {
		return role.ProfileAssignment{}, err
	}
	tenantAllowed, err := s.auth.CanManageTenant(ctx, actor, tenant)
	if err != nil || !tenantAllowed {
		return role.ProfileAssignment{}, ErrForbidden
	}
	a, err := s.assignments.GetAssignment(ctx, tenant, id)
	if err != nil {
		return role.ProfileAssignment{}, err
	}
	ok, err := s.auth.CanManageSubject(ctx, actor, tenant, a.Subject())
	if err != nil || !ok {
		return role.ProfileAssignment{}, ErrForbidden
	}
	return s.assignments.RevokeAssignment(ctx, tenant, id, expected, key)
}
func (s *Service) ListBySubject(ctx context.Context, tenant shared.TenantID, subject shared.Subject, after string, limit int) ([]role.ProfileAssignment, error) {
	if tenant.Validate() != nil || subject.Validate() != nil {
		return nil, ErrInvalidProfile
	}
	actor, err := s.principal(ctx)
	if err != nil {
		return nil, err
	}
	allowed, err := s.auth.CanReadSubject(ctx, actor, tenant, subject)
	if err != nil || !allowed {
		return nil, ErrForbidden
	}
	return s.assignments.ListBySubject(ctx, tenant, subject, after, limit)
}

// CompileActivationPlan is an authorized, non-mutating command preparation.
// It resolves every exact version-pinned template and owner-verified INSTANCE
// scope again at activation time. Non-instance and contextual assignments are
// deliberately unsupported until their owner graph and PDP policy exist.
// This result must be committed atomically with the tenant policy journal.
func (s *Service) CompileActivationPlan(ctx context.Context, tenant shared.TenantID, assignmentID string) ([]projection.ProfileCapability, error) {
	if s == nil || s.assignments == nil || s.templates == nil || s.owner == nil {
		return nil, ErrUnavailable
	}
	if tenant.Validate() != nil || !shared.ValidUUID(assignmentID) {
		return nil, ErrInvalidProfile
	}
	actor, err := s.principal(ctx)
	if err != nil {
		return nil, err
	}
	canManage, err := s.auth.CanManageTenant(ctx, actor, tenant)
	if err != nil || !canManage {
		return nil, ErrForbidden
	}
	a, err := s.assignments.GetAssignment(ctx, tenant, assignmentID)
	if err != nil {
		return nil, err
	}
	allowed, err := s.auth.CanManageSubject(ctx, actor, tenant, a.Subject())
	if err != nil || !allowed {
		return nil, ErrForbidden
	}
	if a.Status() != role.ProfileAssignmentPending || len(a.PolicyIDs()) != 0 {
		return nil, ErrInvalidProfile
	}
	var refs []role.TemplateRef
	switch a.Target().Kind {
	case role.AssignmentTemplate:
		refs = []role.TemplateRef{{Name: a.Target().Name, Version: a.Target().Version}}
	case role.AssignmentBundle:
		refs, err = s.expandPublished(ctx, role.BundleRef{Name: a.Target().Name, Version: a.Target().Version})
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidProfile
	}
	templates, err := s.resolveTemplates(ctx, refs)
	if err != nil {
		return nil, err
	}
	scopes := map[string]authorization.Scope{}
	proofs := map[string]int64{}
	for _, scope := range a.Scopes() {
		if scope.Kind() != authorization.ScopeInstance {
			return nil, projection.ErrUnsupportedProfileScope
		}
		revision, err := s.owner.VerifyScope(ctx, scope)
		if err != nil || revision < 1 {
			return nil, ErrUnverifiedScope
		}
		scopes[scope.ResourceType()] = scope
		proofs[scope.ResourceType()] = revision
	}
	caps := make([]projection.ProfileCapability, 0)
	seen := map[string]bool{}
	used := map[string]bool{}
	for _, t := range templates {
		for _, p := range t.Permissions {
			scope, ok := scopes[p.ResourceType]
			if !ok || !t.Template.SupportsScope(scope.Kind()) {
				return nil, ErrInvalidProfile
			}
			scopeAllowed := false
			for _, allowed := range p.Scopes {
				if allowed == scope.Kind() {
					scopeAllowed = true
				}
			}
			if !scopeAllowed {
				return nil, ErrInvalidProfile
			}
			cap := projection.ProfileCapability{ResourceType: p.ResourceType, ResourceID: scope.ResourceID(),
				Action: p.Action.Name(), OwnerRevision: proofs[p.ResourceType]}
			id, err := projection.CapabilityObjectID(cap.ResourceType, cap.ResourceID, cap.Action)
			if err != nil {
				return nil, err
			}
			if !seen[id] {
				caps = append(caps, cap)
				seen[id] = true
			}
			used[p.ResourceType] = true
		}
	}
	if len(used) != len(scopes) || len(caps) == 0 || len(caps) > 999 {
		return nil, ErrInvalidProfile
	}
	if err := projection.ValidateProfileCapabilities(a, caps); err != nil {
		return nil, err
	}
	return caps, nil
}
