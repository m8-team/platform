package entity

type Permission struct {
	ID          string
	DisplayName string
}

type Role struct {
	ID            string
	DisplayName   string
	Description   string
	Permissions   []Permission
	RelationNames []string
}

var defaultRoles = map[string]Role{
	"tenant-owner": {
		ID:          "tenant-owner",
		DisplayName: "Tenant Owner",
		Description: "Full tenant management access.",
		Permissions: []Permission{
			{ID: "tenant.manage", DisplayName: "Manage tenant"},
			{ID: "project.read", DisplayName: "Read project"},
			{ID: "project.write", DisplayName: "Write project"},
		},
		RelationNames: []string{"owner"},
	},
	"tenant-admin": {
		ID:          "tenant-admin",
		DisplayName: "Tenant Admin",
		Description: "Administrative access for tenant resources.",
		Permissions: []Permission{
			{ID: "tenant.manage", DisplayName: "Manage tenant"},
			{ID: "project.read", DisplayName: "Read project"},
		},
		RelationNames: []string{"admin"},
	},
	"project-viewer": {
		ID:          "project-viewer",
		DisplayName: "Project Viewer",
		Description: "Read-only project access.",
		Permissions: []Permission{
			{ID: "project.read", DisplayName: "Read project"},
		},
		RelationNames: []string{"viewer"},
	},
	"project-editor": {
		ID:          "project-editor",
		DisplayName: "Project Editor",
		Description: "Read-write project access.",
		Permissions: []Permission{
			{ID: "project.read", DisplayName: "Read project"},
			{ID: "project.write", DisplayName: "Write project"},
		},
		RelationNames: []string{"editor"},
	},
	"support-operator": {
		ID:          "support-operator",
		DisplayName: "Support Operator",
		Description: "Temporary support access.",
		Permissions: []Permission{
			{ID: "project.read", DisplayName: "Read project"},
			{ID: "support.case.read", DisplayName: "Read support case"},
		},
		RelationNames: []string{"support"},
	},
}

func DefaultRoles() []Role {
	roles := make([]Role, 0, len(defaultRoles))
	for _, role := range defaultRoles {
		roles = append(roles, role)
	}
	return roles
}

func ResolveRole(roleID string) (Role, bool) {
	role, ok := defaultRoles[roleID]
	return role, ok
}

func PermissionsForRole(roleID string) []string {
	role, ok := ResolveRole(roleID)
	if !ok {
		return nil
	}
	permissions := make([]string, 0, len(role.Permissions))
	for _, permission := range role.Permissions {
		permissions = append(permissions, permission.ID)
	}
	return permissions
}
