package grpc

import (
	"context"

	auditv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/audit/v1"
	opsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/ops/v1"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/storage/ydb"
)

type AuditServer struct {
	auditv1.UnimplementedAuditServiceServer

	store core.DocumentStore
}

func NewAuditServer(store core.DocumentStore) *AuditServer {
	return &AuditServer{store: store}
}

func (s *AuditServer) ListAuditEvents(ctx context.Context, req *auditv1.ListAuditEventsRequest) (*auditv1.ListAuditEventsResponse, error) {
	events, next, err := core.ListProto(ctx, s.store, ydb.TableAuditEvents, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *auditv1.AuditEvent {
		return &auditv1.AuditEvent{}
	})
	if err != nil {
		return nil, err
	}
	return &auditv1.ListAuditEventsResponse{Events: events, NextPageToken: next}, nil
}

func (s *AuditServer) GetAuditEvent(ctx context.Context, req *auditv1.GetAuditEventRequest) (*auditv1.AuditEvent, error) {
	event := &auditv1.AuditEvent{}
	if err := core.LoadProto(ctx, s.store, ydb.TableAuditEvents, req.GetAuditEventId(), event); err != nil {
		return nil, err
	}
	return event, nil
}

type OperationsServer struct {
	opsv1.UnimplementedOperationsServiceServer

	store core.DocumentStore
}

func NewOperationsServer(store core.DocumentStore) *OperationsServer {
	return &OperationsServer{store: store}
}

func (s *OperationsServer) ListOperations(ctx context.Context, req *opsv1.ListOperationsRequest) (*opsv1.ListOperationsResponse, error) {
	operations, next, err := core.ListProto(ctx, s.store, ydb.TableOperations, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *opsv1.Operation {
		return &opsv1.Operation{}
	})
	if err != nil {
		return nil, err
	}
	return &opsv1.ListOperationsResponse{Operations: operations, NextPageToken: next}, nil
}

func (s *OperationsServer) GetOperation(ctx context.Context, req *opsv1.GetOperationRequest) (*opsv1.Operation, error) {
	operation := &opsv1.Operation{}
	if err := core.LoadProto(ctx, s.store, ydb.TableOperations, req.GetOperationId(), operation); err != nil {
		return nil, err
	}
	return operation, nil
}
