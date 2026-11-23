package projection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

func profileFixture(t *testing.T, id string) (role.ProfileAssignment, ProfileCapability) {
	t.Helper()
	tenant, err := shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	if err != nil {
		t.Fatal(err)
	}
	subject, err := shared.NewSubject("user", "alice")
	if err != nil {
		t.Fatal(err)
	}
	scope, err := authorization.NewScope(authorization.ScopeInput{Kind: authorization.ScopeInstance, TenantID: tenant, ResourceType: "finance.invoice", ResourceID: "invoice-123"})
	if err != nil {
		t.Fatal(err)
	}
	a, err := role.NewProfileAssignment(role.ProfileAssignmentInput{ID: id, TenantID: tenant, Subject: subject,
		Target: role.AssignmentTarget{Kind: role.AssignmentTemplate, Name: "finance.accountant", Version: "1.0.0"},
		Scopes: []authorization.Scope{scope}})
	if err != nil {
		t.Fatal(err)
	}
	return a, ProfileCapability{ResourceType: "finance.invoice", ResourceID: "invoice-123", Action: "finance.invoice.view", OwnerRevision: 7}
}
func TestProfileProjectionFailsClosedBeforeActivationAndAfterRevoke(t *testing.T) {
	a, c := profileFixture(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	p := &Projector{}
	edges, err := p.profileAssignmentRelationships(context.Background(), &ProfileAssignmentState{Assignment: a})
	if err != nil || len(edges) != 0 {
		t.Fatalf("staged request projected permissions: %+v %v", edges, err)
	}
	if _, err := p.profileAssignmentRelationships(context.Background(), &ProfileAssignmentState{Assignment: a, Capabilities: []ProfileCapability{c}}); !errors.Is(err, ErrInvalidProfileProjection) {
		t.Fatalf("capabilities before activation allowed: %v", err)
	}
	activated, err := a.Activate(1)
	if err != nil {
		t.Fatal(err)
	}
	edges, err = p.profileAssignmentRelationships(context.Background(), &ProfileAssignmentState{Assignment: activated, Capabilities: []ProfileCapability{c}})
	if err != nil || len(edges) != 2 {
		t.Fatalf("active instance expected 2 edges: %+v %v", edges, err)
	}
	revoked, err := activated.RevokeActive(2, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	edges, err = p.profileAssignmentRelationships(context.Background(), &ProfileAssignmentState{Assignment: revoked})
	if err != nil || len(edges) != 0 {
		t.Fatalf("revoked assignment projected permissions: %+v %v", edges, err)
	}
}
func TestProfileProjectionRejectsScopeWideningAndUnverifiedOwnerRevision(t *testing.T) {
	a, c := profileFixture(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	active, err := a.Activate(1)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []ProfileCapability{
		{ResourceType: c.ResourceType, ResourceID: "invoice-999", Action: c.Action, OwnerRevision: 7},
		{ResourceType: c.ResourceType, ResourceID: c.ResourceID, Action: c.Action, OwnerRevision: 0},
		{ResourceType: "documents.file", ResourceID: c.ResourceID, Action: c.Action, OwnerRevision: 7},
	} {
		if err := (ProfileAssignmentState{Assignment: active, Capabilities: []ProfileCapability{bad}}).Validate(); err == nil {
			t.Fatalf("accepted unverified/widened capability: %+v", bad)
		}
	}
	same, err := CapabilityObjectID(c.ResourceType, c.ResourceID, c.Action)
	if err != nil {
		t.Fatal(err)
	}
	other, _ := CapabilityObjectID(c.ResourceType, c.ResourceID, "finance.invoice.edit")
	if same == other {
		t.Fatal("distinct permissions collided")
	}
}
func TestProfileProjectionRejectsOrganizationScopeUntilOwnerGraphExists(t *testing.T) {
	tenant, _ := shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	subject, _ := shared.NewSubject("user", "alice")
	scope, _ := authorization.NewScope(authorization.ScopeInput{Kind: authorization.ScopeOrganization, TenantID: tenant, ResourceType: "finance.invoice", OrganizationID: "22222222-2222-4222-8222-222222222222"})
	a, err := role.NewProfileAssignment(role.ProfileAssignmentInput{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", TenantID: tenant, Subject: subject,
		Target: role.AssignmentTarget{Kind: role.AssignmentTemplate, Name: "finance.accountant", Version: "1.0.0"}, Scopes: []authorization.Scope{scope}})
	if err != nil {
		t.Fatal(err)
	}
	active, err := a.Activate(1)
	if err != nil {
		t.Fatal(err)
	}
	state := ProfileAssignmentState{Assignment: active, Capabilities: []ProfileCapability{{ResourceType: "finance.invoice", ResourceID: "invoice-123", Action: "finance.invoice.view", OwnerRevision: 7}}}
	if err := state.Validate(); !errors.Is(err, ErrUnsupportedProfileScope) {
		t.Fatalf("organization scope accidentally flattened: %v", err)
	}
}
