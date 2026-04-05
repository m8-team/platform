package audit

import (
	"context"
	"fmt"
	"time"

	auditv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/audit/v1"
	opsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/ops/v1"
	foundationprotokit "github.com/m8platform/platform/iam/internal/foundation/protokit"
	foundationstore "github.com/m8platform/platform/iam/internal/foundation/store"
)

func NewOperation(now time.Time, tenantID string, operationType string, resourceType string, resourceID string) *opsv1.Operation {
	return &opsv1.Operation{
		OperationId:   fmt.Sprintf("op-%d", now.UnixNano()),
		TenantId:      tenantID,
		OperationType: operationType,
		Status:        opsv1.OperationStatus_OPERATION_STATUS_SUCCEEDED,
		ResourceType:  resourceType,
		ResourceId:    resourceID,
		CreatedAt:     foundationprotokit.Timestamp(now),
		UpdatedAt:     foundationprotokit.Timestamp(now),
		CompletedAt:   foundationprotokit.Timestamp(now),
	}
}

func PersistOperation(ctx context.Context, store foundationstore.DocumentStore, operation *opsv1.Operation, now time.Time) error {
	return foundationstore.SaveProto(ctx, store, "operations", operation.GetOperationId(), operation.GetTenantId(), operation, now)
}

func NewEvent(now time.Time, tenantID string, eventType string, actor string, operationID string, reason string) *auditv1.AuditEvent {
	return &auditv1.AuditEvent{
		AuditEventId:  fmt.Sprintf("audit-%d", now.UnixNano()),
		TenantId:      tenantID,
		EventType:     eventType,
		Actor:         actor,
		OperationId:   operationID,
		Reason:        reason,
		OccurredAt:    foundationprotokit.Timestamp(now),
		CorrelationId: operationID,
	}
}

func PersistEvent(ctx context.Context, store foundationstore.DocumentStore, event *auditv1.AuditEvent, now time.Time) error {
	return foundationstore.SaveProto(ctx, store, "audit_events", event.GetAuditEventId(), event.GetTenantId(), event, now)
}
