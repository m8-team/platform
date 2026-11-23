package role

import (
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

func TestVersionedCompositeProfiles(t *testing.T) {
	employee := BundleRef{Name: "profiles.employee", Version: "1.0.0"}
	accountant := BundleRef{Name: "profiles.accountant", Version: "1.0.0"}
	base := TemplateRef{Name: "identity.employee", Version: "1.0.0"}
	finance := TemplateRef{Name: "finance.accountant", Version: "2.0.0"}
	b1, err := NewBundle(employee, []TemplateRef{base}, nil)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := NewBundle(accountant, []TemplateRef{finance}, []BundleRef{employee})
	if err != nil {
		t.Fatal(err)
	}
	roles, err := ExpandProfiles(accountant, []Bundle{b1, b2})
	if err != nil || len(roles) != 2 || roles[0] != base || roles[1] != finance {
		t.Fatalf("bad bundle expansion: %v %+v", err, roles)
	}
	if _, err = ExpandProfiles(accountant, []Bundle{b2}); !errors.Is(err, ErrUnresolvedBundle) {
		t.Fatal(err)
	}
	cyclicA, _ := NewBundle(employee, nil, []BundleRef{accountant})
	if _, err = ExpandProfiles(accountant, []Bundle{cyclicA, b2}); !errors.Is(err, ErrCyclicBundle) {
		t.Fatal(err)
	}
}
func TestAssignmentIsIndependentAndScoped(t *testing.T) {
	tenant, _ := shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	subject, _ := shared.NewSubject("user", "alice")
	scope, err := authorization.NewScope(authorization.ScopeInput{Kind: authorization.ScopeInstance,
		TenantID: tenant, ResourceType: "finance.invoice", ResourceID: "invoice-123"})
	if err != nil {
		t.Fatal(err)
	}
	input := AssignmentInput{ID: "22222222-2222-4222-8222-222222222222", TenantID: tenant,
		Subject: subject, Target: AssignmentTarget{Kind: AssignmentBundle, Name: "profiles.accountant", Version: "1.0.0"}, Scope: scope}
	assigned, err := NewAssignment(input)
	if err != nil {
		t.Fatal(err)
	}
	revoked, err := assigned.Revoke(1, time.Date(2026, 11, 23, 16, 0, 0, 0, time.UTC))
	if err != nil || revoked.State() != AssignmentRevoked || assigned.State() != AssignmentActive {
		t.Fatal(err)
	}
	if _, err = revoked.Revoke(2, time.Now()); !errors.Is(err, ErrAssignmentNotActive) {
		t.Fatal(err)
	}
	input.Scope, _ = authorization.NewScope(authorization.ScopeInput{Kind: authorization.ScopeInstance,
		TenantID: shared.TenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), ResourceType: "finance.invoice", ResourceID: "invoice-123"})
	if _, err = NewAssignment(input); !errors.Is(err, ErrInvalidAssignment) {
		t.Fatal(err)
	}
}
