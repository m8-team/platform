package spicedb

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/structpb"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	legacyauthz "github.com/m8platform/platform/iam/internal/authz"
	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	legacyspicedb "github.com/m8platform/platform/iam/internal/spicedb"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	"github.com/m8platform/platform/iam/internal/usecase/port"
)

type AuthorizationChecker struct {
	client *legacyspicedb.Client
}

func NewAuthorizationChecker(client *legacyspicedb.Client) *AuthorizationChecker {
	return &AuthorizationChecker{client: client}
}

func (c *AuthorizationChecker) CheckAccess(ctx context.Context, query model.AccessCheckQuery) (model.AccessCheckResult, error) {
	if c == nil || c.client == nil {
		return model.AccessCheckResult{}, port.ErrAuthorizationUnavailable
	}

	var caveatContext *structpb.Struct
	if len(query.CaveatContext) > 0 {
		ctxStruct, err := structpb.NewStruct(query.CaveatContext)
		if err != nil {
			return model.AccessCheckResult{}, err
		}
		caveatContext = ctxStruct
	}

	result, err := c.client.Check(ctx, &authzv1.CheckAccessRequest{
		Subject: &authzv1.SubjectRef{
			TenantId: query.Subject.TenantID,
			Type:     subjectTypeFromString(query.Subject.Type),
			Id:       query.Subject.ID,
		},
		Resource: &authzv1.ResourceRef{
			TenantId: query.Resource.TenantID,
			Type:     resourceTypeFromString(query.Resource.Type),
			Id:       query.Resource.ID,
		},
		Permission:    query.Permission,
		CaveatContext: caveatContext,
	})
	if err != nil {
		if errors.Is(err, legacyspicedb.ErrNotConfigured) || errors.Is(err, legacyspicedb.ErrNotImplemented) {
			return model.AccessCheckResult{}, port.ErrAuthorizationUnavailable
		}
		return model.AccessCheckResult{}, err
	}

	return model.AccessCheckResult{
		Decision:          permissionDecisionFromProto(result.GetDecision()),
		Permission:        result.GetPermission(),
		CacheHit:          result.GetCacheHit(),
		ZedToken:          result.GetZedToken(),
		CaveatExpressions: result.GetCaveatExpressions(),
	}, nil
}

func subjectTypeFromString(value string) authzv1.SubjectType {
	switch value {
	case authzv1.SubjectType_SUBJECT_TYPE_GROUP.String():
		return authzv1.SubjectType_SUBJECT_TYPE_GROUP
	case authzv1.SubjectType_SUBJECT_TYPE_SERVICE_ACCOUNT.String():
		return authzv1.SubjectType_SUBJECT_TYPE_SERVICE_ACCOUNT
	case authzv1.SubjectType_SUBJECT_TYPE_FEDERATED_USER.String():
		return authzv1.SubjectType_SUBJECT_TYPE_FEDERATED_USER
	case authzv1.SubjectType_SUBJECT_TYPE_SYSTEM.String():
		return authzv1.SubjectType_SUBJECT_TYPE_SYSTEM
	default:
		return authzv1.SubjectType_SUBJECT_TYPE_USER_ACCOUNT
	}
}

func resourceTypeFromString(value string) authzv1.ResourceType {
	switch value {
	case authzv1.ResourceType_RESOURCE_TYPE_PROJECT.String():
		return authzv1.ResourceType_RESOURCE_TYPE_PROJECT
	case authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE.String():
		return authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE
	default:
		return authzv1.ResourceType_RESOURCE_TYPE_TENANT
	}
}

func permissionDecisionFromProto(value authzv1.PermissionDecision) authzentity.PermissionDecision {
	switch value {
	case authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW:
		return authzentity.PermissionDecisionAllow
	case authzv1.PermissionDecision_PERMISSION_DECISION_CONDITIONAL:
		return authzentity.PermissionDecisionConditional
	default:
		return authzentity.PermissionDecisionDeny
	}
}

type RolePermissionResolver struct{}

func (RolePermissionResolver) Permissions(roleID string) []string {
	return legacyauthz.PermissionsForRole(roleID)
}
