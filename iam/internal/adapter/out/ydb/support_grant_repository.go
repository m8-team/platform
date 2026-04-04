package ydb

import (
	"context"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	supportv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/support/v1"
	legacycore "github.com/m8platform/platform/iam/internal/core"
	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	tenantentity "github.com/m8platform/platform/iam/internal/entity/tenant"
	legacystorage "github.com/m8platform/platform/iam/internal/storage/ydb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SupportGrantRepository struct {
	store legacycore.DocumentStore
}

func NewSupportGrantRepository(store legacycore.DocumentStore) *SupportGrantRepository {
	return &SupportGrantRepository{store: store}
}

func (r *SupportGrantRepository) Save(ctx context.Context, grant tenantentity.SupportGrant) error {
	payload, err := legacycore.MarshalProto(supportGrantToProto(grant))
	if err != nil {
		return err
	}
	return r.store.UpsertDocument(ctx, legacystorage.TableWorkflowLocks, legacycore.StoredDocument{
		ID:        grant.ID,
		TenantID:  grant.TenantID,
		Payload:   payload,
		CreatedAt: grant.CreatedAt.UTC(),
		UpdatedAt: grant.UpdatedAt.UTC(),
	})
}

func (r *SupportGrantRepository) GetByID(ctx context.Context, supportGrantID string) (tenantentity.SupportGrant, error) {
	document, err := r.store.GetDocument(ctx, legacystorage.TableWorkflowLocks, supportGrantID)
	if err != nil {
		return tenantentity.SupportGrant{}, err
	}
	record := &supportv1.SupportGrant{}
	if err := legacycore.UnmarshalProto(document.Payload, record); err != nil {
		return tenantentity.SupportGrant{}, err
	}
	return supportGrantFromProto(record), nil
}

func (r *SupportGrantRepository) ListByTenant(ctx context.Context, tenantID string, pageSize int, pageToken string) ([]tenantentity.SupportGrant, string, error) {
	if pageSize <= 0 {
		pageSize = legacycore.DefaultPageSize
	}
	offset := legacycore.DecodePageToken(pageToken)
	documents, next, err := r.store.ListDocuments(ctx, legacystorage.TableWorkflowLocks, tenantID, offset, pageSize)
	if err != nil {
		return nil, "", err
	}

	grants := make([]tenantentity.SupportGrant, 0, len(documents))
	for _, document := range documents {
		record := &supportv1.SupportGrant{}
		if err := legacycore.UnmarshalProto(document.Payload, record); err != nil {
			return nil, "", err
		}
		grants = append(grants, supportGrantFromProto(record))
	}
	return grants, next, nil
}

func supportGrantToProto(grant tenantentity.SupportGrant) *supportv1.SupportGrant {
	return &supportv1.SupportGrant{
		SupportGrantId: grant.ID,
		TenantId:       grant.TenantID,
		Subject: &authzv1.SubjectRef{
			TenantId: grant.Subject.TenantID,
			Type:     subjectTypeFromString(grant.Subject.Type),
			Id:       grant.Subject.ID,
		},
		Resource: &authzv1.ResourceRef{
			TenantId: grant.Resource.TenantID,
			Type:     resourceTypeFromString(grant.Resource.Type),
			Id:       grant.Resource.ID,
		},
		RoleId:         grant.RoleID,
		Ttl:            durationpb.New(grant.TTL),
		Status:         supportGrantStatusToProto(grant.Status),
		Reason:         grant.Reason,
		ApprovalTicket: grant.ApprovalTicket,
		ApprovedAt:     timestampOrNil(grant.ApprovedAt),
		ExpiresAt:      timestampOrNil(grant.ExpiresAt),
		CreatedAt:      timestampOrNil(grant.CreatedAt),
		UpdatedAt:      timestampOrNil(grant.UpdatedAt),
	}
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

func supportGrantFromProto(grant *supportv1.SupportGrant) tenantentity.SupportGrant {
	if grant == nil {
		return tenantentity.SupportGrant{}
	}
	return tenantentity.SupportGrant{
		ID:       grant.GetSupportGrantId(),
		TenantID: grant.GetTenantId(),
		Subject: authzentity.SubjectRef{
			TenantID: grant.GetSubject().GetTenantId(),
			Type:     grant.GetSubject().GetType().String(),
			ID:       grant.GetSubject().GetId(),
		},
		Resource: authzentity.ResourceRef{
			TenantID: grant.GetResource().GetTenantId(),
			Type:     grant.GetResource().GetType().String(),
			ID:       grant.GetResource().GetId(),
		},
		RoleID:         grant.GetRoleId(),
		TTL:            grant.GetTtl().AsDuration(),
		Status:         supportGrantStatusFromProto(grant.GetStatus()),
		Reason:         grant.GetReason(),
		ApprovalTicket: grant.GetApprovalTicket(),
		ApprovedAt:     timeFromProto(grant.GetApprovedAt()),
		ExpiresAt:      timeFromProto(grant.GetExpiresAt()),
		CreatedAt:      timeFromProto(grant.GetCreatedAt()),
		UpdatedAt:      timeFromProto(grant.GetUpdatedAt()),
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

func supportGrantStatusFromProto(status supportv1.SupportGrantStatus) tenantentity.SupportGrantStatus {
	switch status {
	case supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_ACTIVE:
		return tenantentity.SupportGrantStatusActive
	case supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_REVOKED:
		return tenantentity.SupportGrantStatusRevoked
	case supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_EXPIRED:
		return tenantentity.SupportGrantStatusExpired
	case supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_DENIED:
		return tenantentity.SupportGrantStatusDenied
	default:
		return tenantentity.SupportGrantStatusPendingApproval
	}
}

func timestampOrNil(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func timeFromProto(value *timestamppb.Timestamp) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.AsTime().UTC()
}
