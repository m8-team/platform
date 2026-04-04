package authz

import (
	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
)

var defaultRoles = map[string]*authzv1.Role{
	"tenant-owner": {
		RoleId:      "tenant-owner",
		DisplayName: "Tenant Owner",
		Description: "Full tenant management access.",
		Permissions: []*authzv1.Permission{
			{Id: "tenant.manage", DisplayName: "Manage tenant"},
			{Id: "project.read", DisplayName: "Read project"},
			{Id: "project.write", DisplayName: "Write project"},
		},
		RelationNames: []string{"owner"},
	},
	"tenant-admin": {
		RoleId:      "tenant-admin",
		DisplayName: "Tenant Admin",
		Description: "Administrative access for tenant resources.",
		Permissions: []*authzv1.Permission{
			{Id: "tenant.manage", DisplayName: "Manage tenant"},
			{Id: "project.read", DisplayName: "Read project"},
		},
		RelationNames: []string{"admin"},
	},
	"project-viewer": {
		RoleId:      "project-viewer",
		DisplayName: "Project Viewer",
		Description: "Read-only project access.",
		Permissions: []*authzv1.Permission{
			{Id: "project.read", DisplayName: "Read project"},
		},
		RelationNames: []string{"viewer"},
	},
	"project-editor": {
		RoleId:      "project-editor",
		DisplayName: "Project Editor",
		Description: "Read-write project access.",
		Permissions: []*authzv1.Permission{
			{Id: "project.read", DisplayName: "Read project"},
			{Id: "project.write", DisplayName: "Write project"},
		},
		RelationNames: []string{"editor"},
	},
	"support-operator": {
		RoleId:      "support-operator",
		DisplayName: "Support Operator",
		Description: "Temporary support access.",
		Permissions: []*authzv1.Permission{
			{Id: "project.read", DisplayName: "Read project"},
			{Id: "support.case.read", DisplayName: "Read support case"},
		},
		RelationNames: []string{"support"},
	},
}

func DefaultRoles() []*authzv1.Role {
	roles := make([]*authzv1.Role, 0, len(defaultRoles))
	for _, role := range defaultRoles {
		roles = append(roles, role)
	}
	return roles
}

func ResolveRole(roleID string) (*authzv1.Role, bool) {
	role, ok := defaultRoles[roleID]
	return role, ok
}

func PermissionsForRole(roleID string) []string {
	role, ok := ResolveRole(roleID)
	if !ok {
		return nil
	}
	permissions := make([]string, 0, len(role.GetPermissions()))
	for _, permission := range role.GetPermissions() {
		permissions = append(permissions, permission.GetId())
	}
	return permissions
}
