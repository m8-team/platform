package topics

import (
	"context"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	eventsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/events/v1"
	supportv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/support/v1"
	tenantentity "github.com/m8platform/platform/iam/internal/module/tenant/entity"
	tenantmodel "github.com/m8platform/platform/iam/internal/module/tenant/model"
	legacytopics "github.com/m8platform/platform/iam/internal/topics"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SupportGrantEventPublisher struct {
	publisher *legacytopics.Publisher
	topic     string
}

func NewSupportGrantEventPublisher(publisher *legacytopics.Publisher, topic string) *SupportGrantEventPublisher {
	return &SupportGrantEventPublisher{
		publisher: publisher,
		topic:     topic,
	}
}

func (p *SupportGrantEventPublisher) PublishSupportGrantCreated(ctx context.Context, event tenantmodel.SupportGrantCreatedEvent) error {
	if p == nil || p.publisher == nil {
		return nil
	}
	return p.publisher.PublishProto(ctx, p.topic, &eventsv1.SupportGrantCreated{
		Meta: &eventsv1.EventMeta{
			EventId:       event.EventID,
			OccurredAt:    timestamppb.New(event.OccurredAt.UTC()),
			CorrelationId: event.Grant.ID,
			TenantId:      event.Grant.TenantID,
		},
		Grant: supportGrantToProto(event.Grant),
	})
}

func (p *SupportGrantEventPublisher) PublishSupportGrantRevoked(ctx context.Context, event tenantmodel.SupportGrantRevokedEvent) error {
	if p == nil || p.publisher == nil {
		return nil
	}
	return p.publisher.PublishProto(ctx, p.topic, &eventsv1.SupportGrantRevoked{
		Meta: &eventsv1.EventMeta{
			EventId:       event.EventID,
			OccurredAt:    timestamppb.New(event.OccurredAt.UTC()),
			CorrelationId: event.Grant.ID,
			TenantId:      event.Grant.TenantID,
		},
		SupportGrantId: event.Grant.ID,
	})
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
