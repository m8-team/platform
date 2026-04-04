package support

import (
	"context"
	"fmt"
	"time"

	eventsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/events/v1"
	supportv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/support/v1"
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/storage/ydb"
	"github.com/m8platform/platform/iam/internal/temporalx"
	"go.uber.org/zap"
)

type Service struct {
	supportv1.UnimplementedSupportAccessServiceServer

	store     core.DocumentStore
	publisher core.EventPublisher
	workflows core.WorkflowStarter
	logger    *zap.Logger
	now       func() time.Time
	topics    config.TopicsConfig
}

func NewService(store core.DocumentStore, publisher core.EventPublisher, workflows core.WorkflowStarter, logger *zap.Logger, cfg config.Config) *Service {
	return &Service{
		store:     store,
		publisher: publisher,
		workflows: workflows,
		logger:    logger,
		now:       time.Now,
		topics:    cfg.Topics,
	}
}

func (s *Service) GrantTemporaryAccess(ctx context.Context, req *supportv1.GrantTemporaryAccessRequest) (*supportv1.SupportGrant, error) {
	now := s.now()
	grant := &supportv1.SupportGrant{
		SupportGrantId: fmt.Sprintf("support-%d", now.UnixNano()),
		TenantId:       req.GetTenantId(),
		Subject:        req.GetSubject(),
		Resource:       req.GetResource(),
		RoleId:         req.GetRoleId(),
		Ttl:            req.GetTtl(),
		Status:         supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_PENDING_APPROVAL,
		Reason:         req.GetReason(),
		CreatedAt:      core.Timestamp(now),
		UpdatedAt:      core.Timestamp(now),
	}
	if err := core.SaveProto(ctx, s.store, ydb.TableWorkflowLocks, grant.GetSupportGrantId(), grant.GetTenantId(), grant, now); err != nil {
		return nil, err
	}
	event := &eventsv1.SupportGrantCreated{
		Meta: &eventsv1.EventMeta{
			EventId:       grant.GetSupportGrantId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: grant.GetSupportGrantId(),
			TenantId:      grant.GetTenantId(),
		},
		Grant: grant,
	}
	if err := s.publisher.PublishProto(ctx, s.topics.SupportGrants, event); err != nil {
		return nil, err
	}
	return grant, nil
}

func (s *Service) ApproveTemporaryAccess(ctx context.Context, req *supportv1.ApproveTemporaryAccessRequest) (*supportv1.SupportGrant, error) {
	grant := &supportv1.SupportGrant{}
	if err := core.LoadProto(ctx, s.store, ydb.TableWorkflowLocks, req.GetSupportGrantId(), grant); err != nil {
		return nil, err
	}
	now := s.now()
	grant.Status = supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_ACTIVE
	grant.ApprovalTicket = req.GetApprovalTicket()
	grant.ApprovedAt = core.Timestamp(now)
	grant.ExpiresAt = core.Timestamp(now.Add(grant.GetTtl().AsDuration()))
	grant.UpdatedAt = core.Timestamp(now)
	if err := core.SaveProto(ctx, s.store, ydb.TableWorkflowLocks, grant.GetSupportGrantId(), grant.GetTenantId(), grant, now); err != nil {
		return nil, err
	}
	if s.workflows != nil {
		workflowID := fmt.Sprintf("grant-support-%s", grant.GetSupportGrantId())
		if _, err := s.workflows.StartWorkflow(ctx, temporalx.GrantSupportAccessWorkflowName, workflowID, temporalx.GrantTemporarySupportAccessInput{
			SupportGrantID: grant.GetSupportGrantId(),
			TenantID:       grant.GetTenantId(),
			RequestedBy:    req.GetApprovedBy(),
			Reason:         req.GetReason(),
			TTL:            grant.GetTtl().AsDuration(),
		}); err != nil {
			s.logger.Warn("support workflow start failed", zap.Error(err))
		}
	}
	return grant, nil
}

func (s *Service) RevokeTemporaryAccess(ctx context.Context, req *supportv1.RevokeTemporaryAccessRequest) (*supportv1.SupportGrant, error) {
	grant := &supportv1.SupportGrant{}
	if err := core.LoadProto(ctx, s.store, ydb.TableWorkflowLocks, req.GetSupportGrantId(), grant); err != nil {
		return nil, err
	}
	now := s.now()
	grant.Status = supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_REVOKED
	grant.UpdatedAt = core.Timestamp(now)
	if err := core.SaveProto(ctx, s.store, ydb.TableWorkflowLocks, grant.GetSupportGrantId(), grant.GetTenantId(), grant, now); err != nil {
		return nil, err
	}
	event := &eventsv1.SupportGrantRevoked{
		Meta: &eventsv1.EventMeta{
			EventId:       fmt.Sprintf("revoke-%d", now.UnixNano()),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: grant.GetSupportGrantId(),
			TenantId:      grant.GetTenantId(),
		},
		SupportGrantId: grant.GetSupportGrantId(),
	}
	if err := s.publisher.PublishProto(ctx, s.topics.SupportGrants, event); err != nil {
		return nil, err
	}
	return grant, nil
}

func (s *Service) ListSupportGrants(ctx context.Context, req *supportv1.ListSupportGrantsRequest) (*supportv1.ListSupportGrantsResponse, error) {
	grants, next, err := core.ListProto(ctx, s.store, ydb.TableWorkflowLocks, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *supportv1.SupportGrant {
		return &supportv1.SupportGrant{}
	})
	if err != nil {
		return nil, err
	}
	return &supportv1.ListSupportGrantsResponse{Grants: grants, NextPageToken: next}, nil
}
