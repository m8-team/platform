package authz

import "strings"

type PermissionDecision string

const (
	PermissionDecisionDeny        PermissionDecision = "deny"
	PermissionDecisionAllow       PermissionDecision = "allow"
	PermissionDecisionConditional PermissionDecision = "conditional"
)

type SubjectRef struct {
	TenantID string
	Type     string
	ID       string
}

func (s SubjectRef) Equals(other SubjectRef) bool {
	return s.Type == other.Type && s.ID == other.ID
}

type ResourceRef struct {
	TenantID string
	Type     string
	ID       string
}

func (r ResourceRef) Equals(other ResourceRef) bool {
	return r.Type == other.Type && r.ID == other.ID
}

type AccessBinding struct {
	ID       string
	RoleID   string
	Subject  SubjectRef
	Resource ResourceRef
}

func (b AccessBinding) Grants(subject SubjectRef, permission string, permissions []string) bool {
	if !b.Subject.Equals(subject) {
		return false
	}
	for _, granted := range permissions {
		if strings.TrimSpace(granted) == permission {
			return true
		}
	}
	return false
}
