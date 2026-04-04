package grpc

import (
	"context"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	supportv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/support/v1"
	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	tenantentity "github.com/m8platform/platform/iam/internal/entity/tenant"
	legacysupport "github.com/m8platform/platform/iam/internal/support"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	tenantuc "github.com/m8platform/platform/iam/internal/usecase/tenant"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SupportServer struct {
	*legacysupport.Service

	logger  *zap.Logger
	useCase *tenantuc.SupportAccessUseCase
}

func NewSupportServer(legacy *legacysupport.Service, logger *zap.Logger, useCase *tenantuc.SupportAccessUseCase) *SupportServer {
	return &SupportServer{
		Service: legacy,
		logger:  logger,
		useCase: useCase,
	}
}

func (s *SupportServer) GrantTemporaryAccess(ctx context.Context, req *supportv1.GrantTemporaryAccessRequest) (*supportv1.SupportGrant, error) {
	if s.useCase == nil {
		return s.Service.GrantTemporaryAccess(ctx, req)
	}
	result, err := s.useCase.Grant(ctx, model.GrantSupportAccessCommand{
		RequestID: req.GetRequestId(),
		TenantID:  req.GetTenantId(),
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
		RoleID:      req.GetRoleId(),
		TTL:         req.GetTtl().AsDuration(),
		Reason:      req.GetReason(),
		RequestedBy: req.GetRequestedBy(),
	})
	if err != nil {
		return nil, err
	}
	return supportGrantToProto(result.Grant), nil
}

func (s *SupportServer) ApproveTemporaryAccess(ctx context.Context, req *supportv1.ApproveTemporaryAccessRequest) (*supportv1.SupportGrant, error) {
	if s.useCase == nil {
		return s.Service.ApproveTemporaryAccess(ctx, req)
	}
	result, err := s.useCase.Approve(ctx, model.ApproveSupportAccessCommand{
		SupportGrantID: req.GetSupportGrantId(),
		ApprovalTicket: req.GetApprovalTicket(),
		Reason:         req.GetReason(),
		ApprovedBy:     req.GetApprovedBy(),
	})
	if err != nil {
		return nil, err
	}
	s.logWarnings("approve support access", result.Warnings)
	return supportGrantToProto(result.Grant), nil
}

func (s *SupportServer) RevokeTemporaryAccess(ctx context.Context, req *supportv1.RevokeTemporaryAccessRequest) (*supportv1.SupportGrant, error) {
	if s.useCase == nil {
		return s.Service.RevokeTemporaryAccess(ctx, req)
	}
	result, err := s.useCase.Revoke(ctx, model.RevokeSupportAccessCommand{
		SupportGrantID: req.GetSupportGrantId(),
		Reason:         req.GetReason(),
		RevokedBy:      req.GetRevokedBy(),
	})
	if err != nil {
		return nil, err
	}
	return supportGrantToProto(result.Grant), nil
}

func (s *SupportServer) ListSupportGrants(ctx context.Context, req *supportv1.ListSupportGrantsRequest) (*supportv1.ListSupportGrantsResponse, error) {
	if s.useCase == nil {
		return s.Service.ListSupportGrants(ctx, req)
	}
	grants, next, err := s.useCase.List(ctx, model.ListSupportGrantsQuery{
		TenantID:  req.GetTenantId(),
		PageSize:  int(req.GetPageSize()),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*supportv1.SupportGrant, 0, len(grants))
	for _, grant := range grants {
		items = append(items, supportGrantToProto(grant))
	}
	return &supportv1.ListSupportGrantsResponse{
		Grants:        items,
		NextPageToken: next,
	}, nil
}

func (s *SupportServer) logWarnings(operation string, warnings []error) {
	if s == nil || s.logger == nil {
		return
	}
	for _, warning := range warnings {
		if warning == nil {
			continue
		}
		s.logger.Warn(operation+" degraded", zap.Error(warning))
	}
}

func supportGrantToProto(grant tenantentity.SupportGrant) *supportv1.SupportGrant {
	return &supportv1.SupportGrant{
		SupportGrantId: grant.ID,
		TenantId:       grant.TenantID,
		Subject: &authzv1.SubjectRef{
			TenantId: grant.Subject.TenantID,
			Type:     supportSubjectTypeFromString(grant.Subject.Type),
			Id:       grant.Subject.ID,
		},
		Resource: &authzv1.ResourceRef{
			TenantId: grant.Resource.TenantID,
			Type:     supportResourceTypeFromString(grant.Resource.Type),
			Id:       grant.Resource.ID,
		},
		RoleId:         grant.RoleID,
		Ttl:            durationpb.New(grant.TTL),
		Status:         supportGrantStatusToProto(grant.Status),
		Reason:         grant.Reason,
		ApprovalTicket: grant.ApprovalTicket,
		ApprovedAt:     supportTimestamp(grant.ApprovedAt),
		ExpiresAt:      supportTimestamp(grant.ExpiresAt),
		CreatedAt:      supportTimestamp(grant.CreatedAt),
		UpdatedAt:      supportTimestamp(grant.UpdatedAt),
	}
}

func supportGrantStatusToProto(status tenantentity.SupportGrantStatus) supportv1.SupportGrantStatus {
	switch status {
	case tenantentity.SupportGrantStatusActive:
		return supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_ACTIVE
	case tenantentity.SupportGrantStatusRevoked:
		return supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_REVOKED
	case tenantentity.SupportGrantStatusExpired:
		return supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_EXPIRED
	case tenantentity.SupportGrantStatusDenied:
		return supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_DENIED
	default:
		return supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_PENDING_APPROVAL
	}
}

func supportSubjectTypeFromString(value string) authzv1.SubjectType {
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

func supportResourceTypeFromString(value string) authzv1.ResourceType {
	switch value {
	case authzv1.ResourceType_RESOURCE_TYPE_PROJECT.String():
		return authzv1.ResourceType_RESOURCE_TYPE_PROJECT
	case authzv1.ResourceType_RESOURCE_TYPE_ENVIRONMENT.String():
		return authzv1.ResourceType_RESOURCE_TYPE_ENVIRONMENT
	case authzv1.ResourceType_RESOURCE_TYPE_SECRET.String():
		return authzv1.ResourceType_RESOURCE_TYPE_SECRET
	case authzv1.ResourceType_RESOURCE_TYPE_BILLING_ACCOUNT.String():
		return authzv1.ResourceType_RESOURCE_TYPE_BILLING_ACCOUNT
	case authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE.String():
		return authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE
	default:
		return authzv1.ResourceType_RESOURCE_TYPE_TENANT
	}
}

func supportTimestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
}
