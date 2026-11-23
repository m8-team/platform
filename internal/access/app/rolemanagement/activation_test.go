package rolemanagement

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/role"
)

func TestCompileActivationPlanIsVersionPinnedAndInstanceOnly(t *testing.T) {
	s, input, _, repo := setup(t)
	finance, _ := authorization.NewScope(authorization.ScopeInput{
		Kind: authorization.ScopeInstance, TenantID: input.TenantID,
		ResourceType: "finance.invoice", ResourceID: "invoice-123"})
	input.Scopes[0] = finance
	finAction, _ := permission.NewPermission("finance.invoice.view")
	docAction, _ := permission.NewPermission("documents.file.view")
	finRef := role.TemplateRef{Name: "finance.accountant", Version: "1.0.0"}
	docRef := role.TemplateRef{Name: "documents.reader", Version: "1.0.0"}
	finTemplate, _ := role.NewTemplate("finance", finRef, []permission.Permission{finAction}, []authorization.ScopeKind{authorization.ScopeInstance})
	docTemplate, _ := role.NewTemplate("documents", docRef, []permission.Permission{docAction}, []authorization.ScopeKind{authorization.ScopeInstance})
	s.templates = rolesStub{templates: map[string]VerifiedTemplate{
		finRef.Key(): {Template: finTemplate, Permissions: []VerifiedPermission{{Action: finAction, ResourceType: "finance.invoice", Scopes: []authorization.ScopeKind{authorization.ScopeInstance}}}},
		docRef.Key(): {Template: docTemplate, Permissions: []VerifiedPermission{{Action: docAction, ResourceType: "documents.file", Scopes: []authorization.ScopeKind{authorization.ScopeInstance}}}},
	}}
	a, err := role.NewProfileAssignment(input)
	if err != nil {
		t.Fatal(err)
	}
	repo.items[a.ID()] = a
	caps, err := s.CompileActivationPlan(context.Background(), a.TenantID(), a.ID())
	if err != nil || len(caps) != 2 {
		t.Fatalf("compiled unexpected plan %+v: %v", caps, err)
	}
	if err := projection.ValidateProfileCapabilities(a, caps); err != nil {
		t.Fatalf("compiled plan invalid: %v", err)
	}
	for _, cap := range caps {
		if cap.OwnerRevision != 7 {
			t.Fatalf("missing authoritative revision in plan %+v", cap)
		}
	}
	// Replacing the exact instance with a broad organization scope must deny.
	broad, _ := authorization.NewScope(authorization.ScopeInput{
		Kind: authorization.ScopeOrganization, TenantID: input.TenantID,
		ResourceType: "finance.invoice", OrganizationID: "22222222-2222-4222-8222-222222222222"})
	input.Scopes[0] = broad
	broadAssignment, err := role.NewProfileAssignment(input)
	if err != nil {
		t.Fatal(err)
	}
	repo.items[a.ID()] = broadAssignment
	if _, err := s.CompileActivationPlan(context.Background(), a.TenantID(), a.ID()); !errors.Is(err, projection.ErrUnsupportedProfileScope) {
		t.Fatalf("organization scope silently activated: %v", err)
	}
}
func TestCompileActivationChecksTenantBeforeIDDisclosure(t *testing.T) {
	s, input, _, repository := setup(t)
	s.auth = authStub{allow: false}
	_, err := s.CompileActivationPlan(context.Background(), input.TenantID, input.ID)
	if !errors.Is(err, ErrForbidden) || repository.reads != 0 {
		t.Fatalf("foreign assignment lookup was allowed/disclosed: %v reads=%d", err, repository.reads)
	}
}
