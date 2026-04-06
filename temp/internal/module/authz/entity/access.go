package entity

import (
	"strings"

	"github.com/m8platform/platform/iam/internal/shared/principal"
	"github.com/m8platform/platform/iam/internal/shared/resource"
)

type PermissionDecision string

const (
	PermissionDecisionDeny        PermissionDecision = "deny"
	PermissionDecisionAllow       PermissionDecision = "allow"
	PermissionDecisionConditional PermissionDecision = "conditional"
)

type AccessBinding struct {
	ID       string
	RoleID   string
	Subject  principal.Principal
	Resource resource.Ref
}

func (b AccessBinding) Grants(subject principal.Principal, permission string, permissions []string) bool {
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
