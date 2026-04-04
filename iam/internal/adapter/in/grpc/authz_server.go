package grpc

import (
	"context"
	"fmt"
	"slices"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	legacyauthz "github.com/m8platform/platform/iam/internal/authz"
	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	authzuc "github.com/m8platform/platform/iam/internal/usecase/authz"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	"github.com/m8platform/platform/iam/internal/usecase/port"
)

type AuthorizationServer struct {
	*legacyauthz.Service

	checkAccess *authzuc.CheckAccessUseCase
	bindings    port.AccessBindingRepository
	roles       port.RolePermissionResolver
}

func NewAuthorizationServer(
	legacy *legacyauthz.Service,
	checkAccess *authzuc.CheckAccessUseCase,
	bindings port.AccessBindingRepository,
	roles port.RolePermissionResolver,
) *AuthorizationServer {
	return &AuthorizationServer{
		Service:     legacy,
		checkAccess: checkAccess,
		bindings:    bindings,
		roles:       roles,
	}
}

func (s *AuthorizationServer) CheckAccess(ctx context.Context, req *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error) {
	if s.checkAccess == nil || req.GetSubject() == nil || req.GetResource() == nil {
		return s.Service.CheckAccess(ctx, req)
	}

	result, err := s.checkAccess.Execute(ctx, accessCheckQueryFromProto(req))
	if err != nil {
		return nil, err
	}

	return &authzv1.AccessCheckResult{
		Decision:          permissionDecisionToProto(result.Decision),
		Permission:        result.Permission,
		CacheHit:          result.CacheHit,
		ZedToken:          result.ZedToken,
		CaveatExpressions: result.CaveatExpressions,
	}, nil
}

func (s *AuthorizationServer) BatchCheckAccess(ctx context.Context, req *authzv1.BatchCheckAccessRequest) (*authzv1.BatchCheckAccessResponse, error) {
	if s.checkAccess == nil {
		return s.Service.BatchCheckAccess(ctx, req)
	}

	results := make([]*authzv1.AccessCheckResult, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		result, err := s.CheckAccess(ctx, item)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return &authzv1.BatchCheckAccessResponse{Results: results}, nil
}

func (s *AuthorizationServer) ExplainAccess(ctx context.Context, req *authzv1.ExplainAccessRequest) (*authzv1.ExplainAccessResponse, error) {
	if s.checkAccess == nil || s.bindings == nil || s.roles == nil || req.GetSubject() == nil || req.GetResource() == nil {
		return s.Service.ExplainAccess(ctx, req)
	}

	result, err := s.checkAccess.Execute(ctx, model.AccessCheckQuery{
		Subject: authzentity.SubjectRef{
			TenantID: req.GetSubject().GetTenantId(),
			Type:     req.GetSubject().GetType().String(),
			ID:       req.GetSubject().GetId(),
		},
		Resource: authzentity.ResourceRef{
			TenantID: req.GetResource().GetTenantId(),
			Type:     req.GetResource().GetType().String(),
			ID:       req.GetResource().GetId(),
		},
		Permission: req.GetPermission(),
	})
	if err != nil {
		return nil, err
	}

	bindings, err := s.bindings.ListByResource(ctx, authzentity.ResourceRef{
		TenantID: req.GetResource().GetTenantId(),
		Type:     req.GetResource().GetType().String(),
		ID:       req.GetResource().GetId(),
	})
	if err != nil {
		return nil, err
	}

	pathIDs := make([]string, 0, len(bindings))
	summary := "no matching access path found"
	subject := authzentity.SubjectRef{
		TenantID: req.GetSubject().GetTenantId(),
		Type:     req.GetSubject().GetType().String(),
		ID:       req.GetSubject().GetId(),
	}
	for _, binding := range bindings {
		if binding.Subject.Equals(subject) && slices.Contains(s.roles.Permissions(binding.RoleID), req.GetPermission()) {
			pathIDs = append(pathIDs, binding.ID)
			summary = fmt.Sprintf("subject %s has %s via role %s", subject.ID, req.GetPermission(), binding.RoleID)
		}
	}

	return &authzv1.ExplainAccessResponse{
		Decision: permissionDecisionToProto(result.Decision),
		Summary:  summary,
		PathIds:  pathIDs,
		Revision: fmt.Sprintf("local-%d", time.Now().Unix()),
	}, nil
}

func accessCheckQueryFromProto(req *authzv1.CheckAccessRequest) model.AccessCheckQuery {
	query := model.AccessCheckQuery{
		Subject: authzentity.SubjectRef{
			TenantID: req.GetSubject().GetTenantId(),
			Type:     req.GetSubject().GetType().String(),
			ID:       req.GetSubject().GetId(),
		},
		Resource: authzentity.ResourceRef{
			TenantID: req.GetResource().GetTenantId(),
			Type:     req.GetResource().GetType().String(),
			ID:       req.GetResource().GetId(),
		},
		Permission: req.GetPermission(),
	}
	if req.GetCaveatContext() != nil {
		query.CaveatContext = req.GetCaveatContext().AsMap()
	}
	return query
}

func permissionDecisionToProto(value authzentity.PermissionDecision) authzv1.PermissionDecision {
	switch value {
	case authzentity.PermissionDecisionAllow:
		return authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW
	case authzentity.PermissionDecisionConditional:
		return authzv1.PermissionDecision_PERMISSION_DECISION_CONDITIONAL
	default:
		return authzv1.PermissionDecision_PERMISSION_DECISION_DENY
	}
}
