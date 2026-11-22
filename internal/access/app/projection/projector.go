// Package projection converts authoritative M8 RoleBindings and AccessGrants
// into deterministic, tenant-scoped SpiceDB relationship changes. It does not
// own entity persistence, event offsets, source revisions or caller identity.
package projection

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

var (
	ErrInvalidProjection   = errors.New("invalid access projection transition")
	ErrUnsupportedResource = errors.New("resource kind or permission is not in SpiceDB schema v1")
	ErrUnsupportedSubject  = errors.New("subject kind is not supported by SpiceDB schema v1")
	ErrGrantConditions     = errors.New("grant conditions require a trusted PDP evaluator before activation")
	ErrProjectionTooLarge  = errors.New("access projection exceeds one atomic SpiceDB batch")
	ErrMissingZedToken     = errors.New("SpiceDB relationship write has no ZedToken")
)

const maxUpdates = 1000

// RoleBindingState is a verified snapshot. The role belongs to the Access
// Control Plane and its binding belongs to the same tenant.
type RoleBindingState struct {
	Role    domain.Role
	Binding domain.RoleBinding
}

// Result contains the graph token from a completed atomic batch, if changed.
// A no-op does NOT grant a new source projection watermark or ZedToken.
type Result struct {
	Changed  bool
	ZedToken string
	Changes  int
}

// Projector is deliberately stateless: caller must load committed before/after
// snapshots, persist source watermark and ZedToken durably and gate PDP reads
// on that watermark. It must not advance a watermark on an unsuccessful write.
type Projector struct {
	catalog ports.PermissionCatalog
	writer  ports.RelationshipWriter
	now     func() time.Time
}

func New(catalog ports.PermissionCatalog, writer ports.RelationshipWriter, now func() time.Time) (*Projector, error) {
	if catalog == nil || writer == nil || now == nil {
		return nil, ErrInvalidProjection
	}
	return &Projector{catalog: catalog, writer: writer, now: now}, nil
}

// ProjectRoleBinding applies a CAS-safe graph delta derived from before/after.
// For a role permission change callers MUST reproject every binding using the
// old and new role snapshots; a role edit alone grants no permissions.
func (p *Projector) ProjectRoleBinding(ctx context.Context, before, after *RoleBindingState) (Result, error) {
	if before == nil && after == nil {
		return Result{}, ErrInvalidProjection
	}
	if before != nil && after != nil && (before.Binding.ID() != after.Binding.ID() || before.Binding.TenantID() != after.Binding.TenantID() || after.Binding.Revision() < before.Binding.Revision()) {
		return Result{}, ErrInvalidProjection
	}
	old, err := p.bindingRelationships(ctx, before)
	if err != nil {
		return Result{}, err
	}
	desired, err := p.bindingRelationships(ctx, after)
	if err != nil {
		return Result{}, err
	}
	tenant := chooseTenant(before, after)
	return p.apply(ctx, tenant, old, desired)
}

func chooseTenant(before, after *RoleBindingState) domain.TenantID {
	if after != nil {
		return after.Binding.TenantID()
	}
	return before.Binding.TenantID()
}

func allowedSubject(s domain.Subject) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if s.Type() != "user" && s.Type() != "service_account" {
		return ErrUnsupportedSubject
	}
	return nil
}

func (p *Projector) bindingRelationships(ctx context.Context, snapshot *RoleBindingState) ([]domain.Relationship, error) {
	if snapshot == nil {
		return nil, nil
	}
	role, binding := snapshot.Role, snapshot.Binding
	if err := binding.ValidateForRole(role); err != nil {
		return nil, err
	}
	if err := allowedSubject(binding.Subject()); err != nil {
		return nil, err
	}
	holder, err := domain.NewResource("access_binding", binding.ID())
	if err != nil {
		return nil, err
	}
	rel, err := newRelationship(binding.TenantID(), holder, "holder", binding.Subject(), "", time.Time{})
	if err != nil {
		return nil, err
	}
	result := []domain.Relationship{rel}
	node, err := domain.NewSubject("access_binding", binding.ID())
	if err != nil {
		return nil, err
	}
	for _, permission := range role.Permissions() {
		relation, err := p.directRelation(ctx, binding.TenantID(), binding.Scope(), permission)
		if err != nil {
			return nil, err
		}
		r, err := newRelationship(binding.TenantID(), binding.Scope(), relation, node, "holder", time.Time{})
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// ProjectGrant stages approved grants without a holder relation; only ACTIVE
// grants have an expiring holder. Revoked, rejected and expired grants have no
// graph representation. Each grant ID is a distinct node: revoking one grant
// never deletes another grant's effective permission, even for equal subjects.
func (p *Projector) ProjectGrant(ctx context.Context, before, after *domain.AccessGrant) (Result, error) {
	if before == nil && after == nil {
		return Result{}, ErrInvalidProjection
	}
	if before != nil && after != nil && (before.ID() != after.ID() || before.TenantID() != after.TenantID() || after.Revision() < before.Revision()) {
		return Result{}, ErrInvalidProjection
	}
	old, err := p.grantRelationships(ctx, before, false)
	if err != nil {
		return Result{}, err
	}
	desired, err := p.grantRelationships(ctx, after, true)
	if err != nil {
		return Result{}, err
	}
	var tenant domain.TenantID
	if after != nil {
		tenant = after.TenantID()
	} else {
		tenant = before.TenantID()
	}
	return p.apply(ctx, tenant, old, desired)
}

func (p *Projector) grantRelationships(ctx context.Context, grant *domain.AccessGrant, current bool) ([]domain.Relationship, error) {
	if grant == nil {
		return nil, nil
	}
	if err := grant.Validate(); err != nil {
		return nil, err
	}
	state := grant.State()
	if state != domain.GrantStateApproved && state != domain.GrantStateActive {
		return nil, nil
	}
	if err := allowedSubject(grant.Subject()); err != nil {
		return nil, err
	}
	if current && !p.now().Before(grant.ExpiresAt()) {
		return nil, nil
	}
	// No silently ignored dynamic checks. An ACTIVE conditional grant is not
	// projected until a reviewed caveat/risk/case evaluator exists.
	c := grant.Constraints()
	if c.RequireMFA || c.RequireActiveCase || c.RequireActiveAssignment || c.ServiceAgreementID != "" || c.ConditionPolicyRef != "" {
		return nil, ErrGrantConditions
	}
	node, err := domain.NewSubject("access_grant", grant.ID())
	if err != nil {
		return nil, err
	}
	output := make([]domain.Relationship, 0, len(grant.Permissions())+1)
	for _, permission := range grant.Permissions() {
		relation, err := p.directRelation(ctx, grant.TenantID(), grant.Target().Resource, permission)
		if err != nil {
			return nil, err
		}
		edge, err := newRelationship(grant.TenantID(), grant.Target().Resource, relation, node, "holder", time.Time{})
		if err != nil {
			return nil, err
		}
		output = append(output, edge)
	}
	if state == domain.GrantStateActive {
		resource, err := domain.NewResource("access_grant", grant.ID())
		if err != nil {
			return nil, err
		}
		holder, err := newRelationship(grant.TenantID(), resource, "holder", grant.Subject(), "", grant.ExpiresAt())
		if err != nil {
			return nil, err
		}
		output = append(output, holder)
	}
	return output, nil
}

func newRelationship(tenant domain.TenantID, resource domain.Resource, relation string, subject domain.Subject, subjectRelation string, expiresAt time.Time) (domain.Relationship, error) {
	return domain.NewRelationship(domain.RelationshipInput{TenantID: tenant, Resource: resource, Relation: relation, Subject: subject, SubjectRelation: subjectRelation, ExpiresAt: expiresAt})
}

// directRelations are the only writable permission edges in m8.zed v1.
// GraphPermission itself is a computed SpiceDB permission, never a relation.
var directRelations = map[string]map[string]string{
	"organization":     {"view": "viewer", "edit": "editor", "manage": "manager", "delete": "deleter"},
	"workspace":        {"view": "viewer", "edit": "editor", "manage": "manager", "delete": "deleter"},
	"environment":      {"view": "viewer", "edit": "editor", "manage": "manager", "deploy": "deployer"},
	"workload_service": {"view": "viewer", "edit": "editor", "manage": "manager", "delete": "deleter", "deploy": "deployer"},
}

func (p *Projector) directRelation(ctx context.Context, tenant domain.TenantID, scope domain.Resource, permission domain.Permission) (string, error) {
	if err := scope.Validate(); err != nil {
		return "", err
	}
	defs, ok := directRelations[scope.Type()]
	if !ok {
		return "", ErrUnsupportedResource
	}
	def, err := p.catalog.ResolvePermission(ctx, tenant, permission, scope.Type())
	if err != nil {
		return "", err
	}
	if err := def.Validate(); err != nil {
		return "", err
	}
	if def.Permission() != permission || def.ResourceType() != scope.Type() {
		return "", ErrUnsupportedResource
	}
	relation, ok := defs[def.GraphPermission()]
	if !ok {
		return "", ErrUnsupportedResource
	}
	return relation, nil
}

func (p *Projector) apply(ctx context.Context, tenant domain.TenantID, before, after []domain.Relationship) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if err := tenant.Validate(); err != nil {
		return Result{}, err
	}
	beforeSet, err := indexRelationships(tenant, before)
	if err != nil {
		return Result{}, err
	}
	afterSet, err := indexRelationships(tenant, after)
	if err != nil {
		return Result{}, err
	}
	changes := make([]ports.RelationshipChange, 0, len(beforeSet)+len(afterSet))
	for key, old := range beforeSet {
		if _, exists := afterSet[key]; !exists {
			changes = append(changes, ports.RelationshipChange{Operation: ports.RelationshipDelete, Relationship: old})
		}
	}
	for key, desired := range afterSet {
		old, exists := beforeSet[key]
		if !exists || old.ConditionRef() != desired.ConditionRef() || !old.ExpiresAt().Equal(desired.ExpiresAt()) {
			changes = append(changes, ports.RelationshipChange{Operation: ports.RelationshipTouch, Relationship: desired})
		}
	}
	if len(changes) == 0 {
		return Result{}, nil
	}
	if len(changes) > maxUpdates {
		return Result{}, ErrProjectionTooLarge
	}
	// Deterministic mutations simplify retries, audit and test fixtures.
	sort.Slice(changes, func(i, j int) bool {
		a, b := changes[i].Relationship.Key(), changes[j].Relationship.Key()
		return fmt.Sprint(a, changes[i].Operation) < fmt.Sprint(b, changes[j].Operation)
	})
	written, err := p.writer.WriteRelationships(ctx, tenant, changes)
	if err != nil {
		return Result{}, err
	}
	if written.ZedToken == "" {
		return Result{}, ErrMissingZedToken
	}
	return Result{Changed: true, ZedToken: written.ZedToken, Changes: len(changes)}, nil
}

func indexRelationships(tenant domain.TenantID, entries []domain.Relationship) (map[domain.RelationshipKey]domain.Relationship, error) {
	result := make(map[domain.RelationshipKey]domain.Relationship, len(entries))
	for _, r := range entries {
		if err := r.Validate(); err != nil {
			return nil, err
		}
		if r.TenantID() != tenant {
			return nil, ErrInvalidProjection
		}
		if previous, exists := result[r.Key()]; exists && (previous.ConditionRef() != r.ConditionRef() || !previous.ExpiresAt().Equal(r.ExpiresAt())) {
			return nil, ErrInvalidProjection
		}
		result[r.Key()] = r
	}
	return result, nil
}
