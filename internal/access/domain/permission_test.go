package domain

import (
	"errors"
	"testing"
)

func TestPermissionDefinitionRequiresExplicitGraphMapping(t *testing.T) {
	p := domainPermission(t, "m8.workspace.read")
	d, err := NewPermissionDefinition(p, "workspace", "view")
	if err != nil {
		t.Fatal(err)
	}
	if d.Permission().Name() != "m8.workspace.read" || d.ResourceType() != "workspace" || d.GraphPermission() != "view" {
		t.Fatalf("wrong permission definition: %+v", d)
	}
	if _, err := NewPermissionDefinition(p, "", "view"); !errors.Is(err, ErrInvalidPermissionDefinition) {
		t.Fatalf("missing resource type: %v", err)
	}
	if _, err := NewPermission("x y"); !errors.Is(err, ErrInvalidPermissionName) {
		t.Fatalf("invalid permission name: %v", err)
	}
}
