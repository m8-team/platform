package authorization

import (
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/domain/shared"
)

func TestInstanceCollectionOrganizationAndTenantBoundaries(t *testing.T) {
	tenant, _ := shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	other, _ := shared.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	org := "22222222-2222-4222-8222-222222222222"
	invoice, _ := shared.NewResource("finance.invoice", "invoice-123")
	verified := ResolvedResource{TenantID: tenant, Resource: invoice, OrganizationID: org, SourceRevision: 7}
	for _, tc := range []struct {
		kind       ScopeKind
		resourceID string
		orgID      string
		ids        []string
		ok         bool
	}{
		{ScopeInstance, "invoice-123", "", nil, true},
		{ScopeInstance, "invoice-999", "", nil, false},
		{ScopeCollection, "", "", []string{"invoice-123", "invoice-124"}, true},
		{ScopeCollection, "", "", []string{"invoice-124"}, false},
		{ScopeOrganization, "", org, nil, true},
		{ScopeResourceKind, "", "", nil, true},
		{ScopeTenant, "", "", nil, true},
	} {
		scope, err := NewScope(ScopeInput{Kind: tc.kind, TenantID: tenant, ResourceType: "finance.invoice",
			ResourceID: tc.resourceID, OrganizationID: tc.orgID, CollectionIDs: tc.ids})
		if err != nil {
			t.Fatal(err)
		}
		ok, err := scope.Matches(verified)
		if err != nil || ok != tc.ok {
			t.Fatalf("%v expected %v, got %v %v", tc.kind, tc.ok, ok, err)
		}
		verified.TenantID = other
		ok, err = scope.Matches(verified)
		if err != nil || ok {
			t.Fatalf("%v leaked across tenant", tc.kind)
		}
		verified.TenantID = tenant
	}
	verified.SourceRevision = 0
	scope, _ := NewScope(ScopeInput{Kind: ScopeTenant, TenantID: tenant, ResourceType: "finance.invoice"})
	if _, err := scope.Matches(verified); !errors.Is(err, ErrUnverifiedResource) {
		t.Fatal(err)
	}
}
func TestSubtreeRequiresOwnerProvidedAncestry(t *testing.T) {
	tenant, _ := shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	root, _ := shared.NewResource("workspace", "workspace-a")
	invoice, _ := shared.NewResource("finance.invoice", "invoice-1")
	scope, _ := NewScope(ScopeInput{Kind: ScopeSubtree, TenantID: tenant, ResourceType: "finance.invoice", Root: root})
	verified := ResolvedResource{TenantID: tenant, Resource: invoice, SourceRevision: 5}
	ok, err := scope.Matches(verified)
	if err != nil || ok {
		t.Fatal("absence of ancestry must not match")
	}
	verified.Ancestors = []shared.Resource{root}
	ok, err = scope.Matches(verified)
	if err != nil || !ok {
		t.Fatal(err)
	}
}
