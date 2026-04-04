package audit

import (
	"context"

	auditv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/audit/v1"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/storage/ydb"
)

type Service struct {
	auditv1.UnimplementedAuditServiceServer

	store core.DocumentStore
}

func NewService(store core.DocumentStore) *Service {
	return &Service{store: store}
}

func (s *Service) ListAuditEvents(ctx context.Context, req *auditv1.ListAuditEventsRequest) (*auditv1.ListAuditEventsResponse, error) {
	events, next, err := core.ListProto(ctx, s.store, ydb.TableAuditEvents, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *auditv1.AuditEvent {
		return &auditv1.AuditEvent{}
	})
	if err != nil {
		return nil, err
	}
	return &auditv1.ListAuditEventsResponse{Events: events, NextPageToken: next}, nil
}

func (s *Service) GetAuditEvent(ctx context.Context, req *auditv1.GetAuditEventRequest) (*auditv1.AuditEvent, error) {
	event := &auditv1.AuditEvent{}
	if err := core.LoadProto(ctx, s.store, ydb.TableAuditEvents, req.GetAuditEventId(), event); err != nil {
		return nil, err
	}
	return event, nil
}
