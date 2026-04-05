package model

import (
	"github.com/m8platform/platform/iam/internal/module/authz/entity"
	"github.com/m8platform/platform/iam/internal/shared/principal"
	"github.com/m8platform/platform/iam/internal/shared/resource"
)

type AccessCheckQuery struct {
	Subject       principal.Principal
	Resource      resource.Ref
	Permission    string
	CaveatContext map[string]any
}

type AccessCheckResult struct {
	Decision          entity.PermissionDecision
	Permission        string
	CacheHit          bool
	ZedToken          string
	CaveatExpressions []string
}
