package domain

import (
	"errors"
	"testing"
)

func TestRoleImmutablePermissionsAndRevision(t *testing.T) {
	perms := []Permission{domainPermission(t, "read")}
	input := RoleInput{ID: testRoleID, TenantID: TenantID(testTenantID), Name: "workspace_viewer", Permissions: perms, Revision: 1}
	role, err := NewRole(input)
	if err != nil {
		t.Fatal(err)
	}
	perms[0] = domainPermission(t, "delete")
	if role.Permissions()[0].Name() != "read" {
		t.Fatal("constructor leaked caller-owned slice")
	}
	copyPerms := role.Permissions()
	copyPerms[0] = domainPermission(t, "delete")
	if role.Permissions()[0].Name() != "read" {
		t.Fatal("getter leaked mutable role slice")
	}
	if _, err := role.WithPermissions(2, []Permission{domainPermission(t, "update")}); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale revision: %v", err)
	}
	updated, err := role.WithPermissions(1, []Permission{domainPermission(t, "update")})
	if err != nil {
		t.Fatal(err)
	}
	if role.Revision() != 1 || updated.Revision() != 2 || updated.Permissions()[0].Name() != "update" {
		t.Fatal("immutable revision update failed")
	}
}

func TestRoleRejectsInvalidAndDuplicatePermissions(t *testing.T) {
	p := domainPermission(t, "read")
	input := RoleInput{ID: testRoleID, TenantID: TenantID(testTenantID), Name: "operator", Revision: 1, Permissions: []Permission{p, p}}
	if _, err := NewRole(input); !errors.Is(err, ErrDuplicatePermission) {
		t.Fatalf("duplicate: %v", err)
	}
	input.Permissions = []Permission{p}
	input.TenantID = ""
	if _, err := NewRole(input); !errors.Is(err, ErrInvalidTenantID) {
		t.Fatalf("missing tenant: %v", err)
	}
	input.TenantID = TenantID(testTenantID)
	input.Revision = 0
	if _, err := NewRole(input); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("missing revision: %v", err)
	}
}

func TestRoleBindingExplicitTenantAndScope(t *testing.T) {
	input := RoleBindingInput{ID: testBindingID, TenantID: TenantID(testTenantID), RoleID: testRoleID, Subject: domainSubject(t), Scope: domainResource(t), Revision: 1}
	binding, err := NewRoleBinding(input)
	if err != nil {
		t.Fatal(err)
	}
	if binding.TenantID() != TenantID(testTenantID) || binding.Scope().Type() != "workspace" {
		t.Fatalf("invalid binding: %+v", binding)
	}
	role, err := NewRole(RoleInput{ID: testRoleID, TenantID: TenantID(testOtherTenantID), Name: "operator", Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := binding.ValidateForRole(role); !errors.Is(err, ErrRoleTenantMismatch) {
		t.Fatalf("cross-tenant binding: %v", err)
	}
	input.Scope = Resource{}
	if _, err := NewRoleBinding(input); !errors.Is(err, ErrResourceTypeRequired) {
		t.Fatalf("scope cannot be omitted: %v", err)
	}
}
