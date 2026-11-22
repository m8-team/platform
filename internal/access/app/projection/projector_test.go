package projection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

const tenantA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
const tenantB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
const aliceID = "alice"

var fixedNow = time.Date(2026, 11, 22, 19, 0, 0, 0, time.UTC)

type memoryWriter struct {
	edges      map[domain.RelationshipKey]domain.Relationship
	calls      int
	batches    [][]ports.RelationshipChange
	failure    error
	emptyToken bool
}

func (w *memoryWriter) WriteRelationships(ctx context.Context, tenant domain.TenantID, changes []ports.RelationshipChange) (ports.RelationshipWriteResult, error) {
	if err := ctx.Err(); err != nil {
		return ports.RelationshipWriteResult{}, err
	}
	if w.failure != nil {
		return ports.RelationshipWriteResult{}, w.failure
	}
	if w.edges == nil {
		w.edges = map[domain.RelationshipKey]domain.Relationship{}
	}
	for _, c := range changes {
		if c.Relationship.TenantID() != tenant {
			panic("cross-tenant tuple")
		}
		if c.Operation == ports.RelationshipDelete {
			delete(w.edges, c.Relationship.Key())
		} else {
			w.edges[c.Relationship.Key()] = c.Relationship
		}
	}
	w.calls++
	w.batches = append(w.batches, append([]ports.RelationshipChange(nil), changes...))
	if w.emptyToken {
		return ports.RelationshipWriteResult{}, nil
	}
	return ports.RelationshipWriteResult{ZedToken: "zed-" + string(rune('0'+w.calls))}, nil
}

func tenant(t *testing.T, s string) domain.TenantID {
	t.Helper()
	v, e := domain.NewTenantID(s)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func subject(t *testing.T, typ, id string) domain.Subject {
	t.Helper()
	v, e := domain.NewSubject(typ, id)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func resource(t *testing.T, typ, id string) domain.Resource {
	t.Helper()
	v, e := domain.NewResource(typ, id)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func permission(t *testing.T, name string) domain.Permission {
	t.Helper()
	v, e := domain.NewPermission(name)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func projector(t *testing.T, w *memoryWriter) *Projector {
	t.Helper()
	defs := []domain.PermissionDefinition{}
	for _, item := range []struct{ res, action, graph string }{
		{"workspace", "m8.workspace.read", "view"},
		{"workspace", "m8.workspace.update", "edit"},
		{"workspace", "m8.workspace.delete", "delete"},
		{"workspace", "m8.workspace.manage", "manage"},
	} {
		d, e := domain.NewPermissionDefinition(permission(t, item.action), item.res, item.graph)
		if e != nil {
			t.Fatal(e)
		}
		defs = append(defs, d)
	}
	catalog := &catalog{defs: defs}
	p, e := New(catalog, w, func() time.Time { return fixedNow })
	if e != nil {
		t.Fatal(e)
	}
	return p
}

type catalog struct{ defs []domain.PermissionDefinition }

func (c *catalog) ResolvePermission(ctx context.Context, tn domain.TenantID, p domain.Permission, rt string) (domain.PermissionDefinition, error) {
	if err := ctx.Err(); err != nil {
		return domain.PermissionDefinition{}, err
	}
	for _, d := range c.defs {
		if d.Permission() == p && d.ResourceType() == rt {
			return d, nil
		}
	}
	return domain.PermissionDefinition{}, ports.ErrPermissionNotRegistered
}
func makeGrant(t *testing.T, id string, tn domain.TenantID) domain.AccessGrant {
	t.Helper()
	g, e := domain.NewAccessGrant(domain.AccessGrantInput{
		ID: id, TenantID: tn, Subject: subject(t, "user", aliceID),
		Target:      domain.GrantTarget{OrganizationID: "33333333-3333-4333-8333-333333333333", Resource: resource(t, "workspace", "w1")},
		Permissions: []domain.Permission{permission(t, "m8.workspace.read")},
		Reason:      domain.GrantReasonSupportCase, ReasonRef: "case-1", RequestedBy: subject(t, "user", "admin"),
		CreatedAt: fixedNow.Add(-time.Hour), ValidFrom: fixedNow.Add(-time.Hour), ExpiresAt: fixedNow.Add(time.Hour),
	})
	if e != nil {
		t.Fatal(e)
	}
	return g
}
func approved(t *testing.T, g domain.AccessGrant) domain.AccessGrant {
	t.Helper()
	p, e := g.SubmitForApproval(g.Revision(), fixedNow.Add(-50*time.Minute))
	if e != nil {
		t.Fatal(e)
	}
	a, e := p.Approve(p.Revision(), subject(t, "user", "security-admin"), fixedNow.Add(-40*time.Minute))
	if e != nil {
		t.Fatal(e)
	}
	return a
}
func active(t *testing.T, g domain.AccessGrant) domain.AccessGrant {
	t.Helper()
	a, e := g.Activate(g.Revision(), "zed-stage", fixedNow.Add(-30*time.Minute))
	if e != nil {
		t.Fatal(e)
	}
	return a
}
func hasEdge(w *memoryWriter, resType, relation, subjType string) bool {
	for _, r := range w.edges {
		if r.Resource().Type() == resType && r.Relation() == relation && r.Subject().Type() == subjType {
			return true
		}
	}
	return false
}

func TestGrantApprovedActiveRevokeIndependent(t *testing.T) {
	w := &memoryWriter{}
	p := projector(t, w)
	g1 := makeGrant(t, "11111111-1111-4111-8111-111111111111", tenant(t, tenantA))
	a1 := approved(t, g1)
	stage, e := p.ProjectGrant(context.Background(), nil, &a1)
	if e != nil || !stage.Changed {
		t.Fatalf("stage: %+v %v", stage, e)
	}
	if len(w.edges) != 1 || hasEdge(w, "access_grant", "holder", "user") {
		t.Fatalf("approved grant granted access: %+v", w.edges)
	}
	active1 := active(t, a1)
	if _, e = p.ProjectGrant(context.Background(), &a1, &active1); e != nil {
		t.Fatal(e)
	}
	if len(w.edges) != 2 || !hasEdge(w, "access_grant", "holder", "user") {
		t.Fatalf("active grant not projected: %+v", w.edges)
	}
	for _, r := range w.edges {
		if r.Resource().Type() == "access_grant" && !r.ExpiresAt().Equal(active1.ExpiresAt()) {
			t.Fatalf("expiration missing: %v", r.ExpiresAt())
		}
	}
	g2 := makeGrant(t, "22222222-2222-4222-8222-222222222222", tenant(t, tenantA))
	a2 := approved(t, g2)
	active2 := active(t, a2)
	if _, e = p.ProjectGrant(context.Background(), nil, &active2); e != nil {
		t.Fatal(e)
	}
	if len(w.edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(w.edges))
	}
	revoked1, e := active1.Revoke(active1.Revision(), subject(t, "user", "security-admin"), fixedNow.Add(-20*time.Minute))
	if e != nil {
		t.Fatal(e)
	}
	res, e := p.ProjectGrant(context.Background(), &active1, &revoked1)
	if e != nil || res.Changes != 2 {
		t.Fatalf("revoke: %+v %v", res, e)
	}
	if len(w.edges) != 2 {
		t.Fatalf("revocation must preserve grant 2, found %d", len(w.edges))
	}
	for _, r := range w.edges {
		if r.Resource().ID() != g2.ID() && r.Subject().ID() != g2.ID() {
			t.Fatalf("unrelated edge remains: %+v", r.Key())
		}
	}
	res, e = p.ProjectGrant(context.Background(), &revoked1, &revoked1)
	if e != nil || res.Changed || w.calls != 4 {
		t.Fatalf("noop replay must not write: %+v %v calls=%d", res, e, w.calls)
	}
}

func TestRoleBindingFanoutAndIsolation(t *testing.T) {
	w := &memoryWriter{}
	p := projector(t, w)
	tn := tenant(t, tenantA)
	roleID := "44444444-4444-4444-8444-444444444444"
	oldRole, e := domain.NewRole(domain.RoleInput{ID: roleID, TenantID: tn, Name: "reader", Revision: 1, Permissions: []domain.Permission{permission(t, "m8.workspace.read")}})
	if e != nil {
		t.Fatal(e)
	}
	bind := func(id string) domain.RoleBinding {
		b, e := domain.NewRoleBinding(domain.RoleBindingInput{ID: id, RoleID: roleID, TenantID: tn, Subject: subject(t, "user", aliceID), Scope: resource(t, "workspace", "w1"), Revision: 1})
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	b1 := bind("11111111-1111-4111-8111-111111111111")
	b2 := bind("22222222-2222-4222-8222-222222222222")
	s1, s2 := RoleBindingState{oldRole, b1}, RoleBindingState{oldRole, b2}
	for _, s := range []*RoleBindingState{&s1, &s2} {
		if _, e := p.ProjectRoleBinding(context.Background(), nil, s); e != nil {
			t.Fatal(e)
		}
	}
	if len(w.edges) != 4 {
		t.Fatalf("got %d edges", len(w.edges))
	}
	if _, e := p.ProjectRoleBinding(context.Background(), &s1, nil); e != nil {
		t.Fatal(e)
	}
	if len(w.edges) != 2 {
		t.Fatalf("removed wrong binding: got %d edges", len(w.edges))
	}
	newRole, e := oldRole.WithPermissions(1, []domain.Permission{permission(t, "m8.workspace.update")})
	if e != nil {
		t.Fatal(e)
	}
	s2New := RoleBindingState{newRole, b2}
	if _, e = p.ProjectRoleBinding(context.Background(), &s2, &s2New); e != nil {
		t.Fatal(e)
	}
	if hasEdge(w, "workspace", "viewer", "access_binding") || !hasEdge(w, "workspace", "editor", "access_binding") {
		t.Fatalf("role update did not replace permission edges")
	}
}

func TestProjectionRejectsUnsafeInputAndDoesNotWrite(t *testing.T) {
	tn := tenant(t, tenantA)
	for _, tc := range []struct {
		name          string
		before, after *domain.AccessGrant
		want          error
	}{
		{"unmatched id", ptr(makeGrant(t, "11111111-1111-4111-8111-111111111111", tn)), ptr(makeGrant(t, "22222222-2222-4222-8222-222222222222", tn)), ErrInvalidProjection},
		{"cross tenant", ptr(makeGrant(t, "11111111-1111-4111-8111-111111111111", tn)), ptr(makeGrant(t, "11111111-1111-4111-8111-111111111111", tenant(t, tenantB))), ErrInvalidProjection},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &memoryWriter{}
			p := projector(t, w)
			_, e := p.ProjectGrant(context.Background(), tc.before, tc.after)
			if !errors.Is(e, tc.want) || w.calls != 0 {
				t.Fatalf("err=%v calls=%d", e, w.calls)
			}
		})
	}
	g := approved(t, makeGrant(t, "11111111-1111-4111-8111-111111111111", tn))
	w := &memoryWriter{}
	p := projector(t, w)
	unsupported := makeGrant(t, "22222222-2222-4222-8222-222222222222", tn)
	_ = unsupported
	// Expired replay must not create a new holder edge.
	expiredClock, err := New(p.catalog, w, func() time.Time { return fixedNow.Add(2 * time.Hour) })
	if err != nil {
		t.Fatal(err)
	}
	ag := active(t, g)
	res, e := expiredClock.ProjectGrant(context.Background(), nil, &ag)
	if e != nil || res.Changed || w.calls != 0 {
		t.Fatalf("expired projection must not write, got %+v %v", res, e)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = p.ProjectGrant(canceled, nil, &ag); !errors.Is(e, context.Canceled) {
		t.Fatalf("cancellation not honored: %v", e)
	}
}
func ptr[T any](v T) *T { return &v }

func TestGrantConditionalFailsClosed(t *testing.T) {
	tn := tenant(t, tenantA)
	g, e := domain.NewAccessGrant(domain.AccessGrantInput{
		ID: "11111111-1111-4111-8111-111111111111", TenantID: tn,
		Subject:     subject(t, "user", "alice"),
		Target:      domain.GrantTarget{OrganizationID: "33333333-3333-4333-8333-333333333333", Resource: resource(t, "workspace", "w1")},
		Permissions: []domain.Permission{permission(t, "m8.workspace.read")},
		Reason:      domain.GrantReasonSupportCase, ReasonRef: "case", RequestedBy: subject(t, "user", "admin"),
		CreatedAt: fixedNow.Add(-time.Hour), ValidFrom: fixedNow.Add(-time.Hour), ExpiresAt: fixedNow.Add(time.Hour),
		Constraints: domain.GrantConstraints{RequireMFA: true},
	})
	if e != nil {
		t.Fatal(e)
	}
	a := active(t, approved(t, g))
	w := &memoryWriter{}
	p := projector(t, w)
	if _, e := p.ProjectGrant(context.Background(), nil, &a); !errors.Is(e, ErrGrantConditions) || w.calls != 0 {
		t.Fatalf("conditional grant allowed: %v calls=%d", e, w.calls)
	}
}

func TestFailedWriteNeverReturnsSuccess(t *testing.T) {
	w := &memoryWriter{failure: errors.New("network")}
	p := projector(t, w)
	grant := approved(t, makeGrant(t, "11111111-1111-4111-8111-111111111111", tenant(t, tenantA)))
	if _, err := p.ProjectGrant(context.Background(), nil, &grant); err == nil {
		t.Fatal("write failure ignored")
	}
	w.failure = nil
	w.emptyToken = true
	if _, err := p.ProjectGrant(context.Background(), nil, &grant); !errors.Is(err, ErrMissingZedToken) {
		t.Fatalf("missing token: %v", err)
	}
}
