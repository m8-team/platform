package role

import (
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

func TestProfileAssignmentIndependentMultiModuleScopes(t *testing.T) {
	tenant, _ := shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	subject, _ := shared.NewSubject("user", "alice")
	invoice, _ := authorization.NewScope(authorization.ScopeInput{Kind: authorization.ScopeOrganization,
		TenantID: tenant, ResourceType: "finance.invoice", OrganizationID: "22222222-2222-4222-8222-222222222222"})
	document, _ := authorization.NewScope(authorization.ScopeInput{Kind: authorization.ScopeInstance,
		TenantID: tenant, ResourceType: "documents.file", ResourceID: "file-42"})
	input := ProfileAssignmentInput{
		ID: "33333333-3333-4333-8333-333333333333", TenantID: tenant, Subject: subject,
		Target: AssignmentTarget{Kind: AssignmentBundle, Name: "profiles.accountant", Version: "1.0.0"},
		Scopes: []authorization.Scope{invoice, document},
	}
	a, err := NewProfileAssignment(input)
	if err != nil || a.Status() != ProfileAssignmentPending || len(a.Scopes()) != 2 {
		t.Fatalf("invalid staged assignment: %+v %v", a, err)
	}
	revoked, err := a.Revoke(1, time.Now())
	if err != nil || revoked.Revision() != 2 || revoked.Status() != ProfileAssignmentRevoked ||
		a.Status() != ProfileAssignmentPending {
		t.Fatalf("independent revocation: %v", err)
	}
	loaded, err := RestoreProfileAssignment(input, revoked.Revision(), revoked.Status(), revoked.RevokedAt())
	if err != nil || loaded.Status() != ProfileAssignmentRevoked {
		t.Fatal(err)
	}
	input.Scopes = []authorization.Scope{invoice, invoice}
	if _, err = NewProfileAssignment(input); !errors.Is(err, ErrInvalidProfileAssignment) {
		t.Fatalf("duplicate resource type accepted: %v", err)
	}
	input.Scopes = nil
	if _, err = NewProfileAssignment(input); !errors.Is(err, ErrInvalidProfileAssignment) {
		t.Fatalf("missing scope accepted: %v", err)
	}
}
