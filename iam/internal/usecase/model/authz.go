package model

import (
	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
)

type AccessCheckQuery struct {
	Subject       authzentity.SubjectRef
	Resource      authzentity.ResourceRef
	Permission    string
	CaveatContext map[string]any
}

type AccessCheckResult struct {
	Decision          authzentity.PermissionDecision
	Permission        string
	CacheHit          bool
	ZedToken          string
	CaveatExpressions []string
}
