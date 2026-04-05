package entity

import "testing"

func TestPermissionsForRole(t *testing.T) {
	t.Parallel()

	permissions := PermissionsForRole("project-editor")
	if len(permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(permissions))
	}
	if permissions[0] != "project.read" || permissions[1] != "project.write" {
		t.Fatalf("unexpected permissions: %#v", permissions)
	}
}
