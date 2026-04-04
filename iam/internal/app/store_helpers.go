package app

import (
	"context"
	"fmt"
	"time"

	auditv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/audit/v1"
	opsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/ops/v1"
	"google.golang.org/protobuf/proto"
)

const DefaultPageSize = 50

func LoadProto(ctx context.Context, store DocumentStore, table string, id string, target proto.Message) error {
	document, err := store.GetDocument(ctx, table, id)
	if err != nil {
		return err
	}
	return UnmarshalProto(document.Payload, target)
}

func SaveProto(ctx context.Context, store DocumentStore, table string, id string, tenantID string, message proto.Message, now time.Time) error {
	payload, err := MarshalProto(message)
	if err != nil {
		return err
	}
	return store.UpsertDocument(ctx, table, StoredDocument{
		ID:        id,
		TenantID:  tenantID,
		Payload:   payload,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	})
}

func ListProto[T proto.Message](ctx context.Context, store DocumentStore, table string, tenantID string, pageSize int, pageToken string, newItem func() T) ([]T, string, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	offset := DecodePageToken(pageToken)
	documents, next, err := store.ListDocuments(ctx, table, tenantID, offset, pageSize)
	if err != nil {
		return nil, "", err
	}
	items := make([]T, 0, len(documents))
	for _, document := range documents {
		item := newItem()
		if err := UnmarshalProto(document.Payload, item); err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}
	return items, next, nil
}

func NewOperation(now time.Time, tenantID string, operationType string, resourceType string, resourceID string) *opsv1.Operation {
	return &opsv1.Operation{
		OperationId:   fmt.Sprintf("op-%d", now.UnixNano()),
		TenantId:      tenantID,
		OperationType: operationType,
		Status:        opsv1.OperationStatus_OPERATION_STATUS_SUCCEEDED,
		ResourceType:  resourceType,
		ResourceId:    resourceID,
		CreatedAt:     Timestamp(now),
		UpdatedAt:     Timestamp(now),
		CompletedAt:   Timestamp(now),
	}
}

func PersistOperation(ctx context.Context, store DocumentStore, operation *opsv1.Operation, now time.Time) error {
	return SaveProto(ctx, store, "operations", operation.GetOperationId(), operation.GetTenantId(), operation, now)
}

func NewAuditEvent(now time.Time, tenantID string, eventType string, actor string, operationID string, reason string) *auditv1.AuditEvent {
	return &auditv1.AuditEvent{
		AuditEventId:  fmt.Sprintf("audit-%d", now.UnixNano()),
		TenantId:      tenantID,
		EventType:     eventType,
		Actor:         actor,
		OperationId:   operationID,
		Reason:        reason,
		OccurredAt:    Timestamp(now),
		CorrelationId: operationID,
	}
}

func PersistAuditEvent(ctx context.Context, store DocumentStore, event *auditv1.AuditEvent, now time.Time) error {
	return SaveProto(ctx, store, "audit_events", event.GetAuditEventId(), event.GetTenantId(), event, now)
}
