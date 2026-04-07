package events

import (
	"encoding/json"
	"fmt"
	"time"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/domain/project"
	"github.com/m8platform/platform/internal/domain/workspace"
	"github.com/m8platform/platform/internal/ports"
)

func OrganizationRecord(
	generator ports.UUIDGenerator,
	event organization.Event,
	meta appcommon.Metadata,
) (ports.OutboxRecord, error) {
	payload, err := json.Marshal(event.Aggregate)
	if err != nil {
		return ports.OutboxRecord{}, fmt.Errorf("marshal organization event payload: %w", err)
	}
	return buildRecord(generator, AggregateOrganization, event.Type, event.Aggregate.ID, "", event.Aggregate.ETag, event.OccurredAt, meta, payload), nil
}

func WorkspaceRecord(
	generator ports.UUIDGenerator,
	event workspace.Event,
	meta appcommon.Metadata,
) (ports.OutboxRecord, error) {
	payload, err := json.Marshal(event.Aggregate)
	if err != nil {
		return ports.OutboxRecord{}, fmt.Errorf("marshal workspace event payload: %w", err)
	}
	return buildRecord(generator, AggregateWorkspace, event.Type, event.Aggregate.ID, event.Aggregate.OrganizationID, event.Aggregate.ETag, event.OccurredAt, meta, payload), nil
}

func ProjectRecord(
	generator ports.UUIDGenerator,
	event project.Event,
	meta appcommon.Metadata,
) (ports.OutboxRecord, error) {
	payload, err := json.Marshal(event.Aggregate)
	if err != nil {
		return ports.OutboxRecord{}, fmt.Errorf("marshal project event payload: %w", err)
	}
	return buildRecord(generator, AggregateProject, event.Type, event.Aggregate.ID, event.Aggregate.WorkspaceID, event.Aggregate.ETag, event.OccurredAt, meta, payload), nil
}

func buildRecord(
	generator ports.UUIDGenerator,
	aggregateType string,
	eventType string,
	aggregateID string,
	parentAggregateID string,
	etag string,
	occurredAt time.Time,
	meta appcommon.Metadata,
	payload json.RawMessage,
) ports.OutboxRecord {
	id := ""
	if generator != nil {
		id = generator.NewString()
	}
	return ports.OutboxRecord{
		ID:                id,
		EventType:         eventType,
		EventVersion:      1,
		AggregateType:     aggregateType,
		AggregateID:       aggregateID,
		ParentAggregateID: parentAggregateID,
		OccurredAt:        occurredAt,
		Actor:             meta.Actor,
		CorrelationID:     meta.CorrelationID,
		CausationID:       meta.CausationID,
		IdempotencyKey:    meta.IdempotencyKey,
		ETagOrRevision:    etag,
		Payload:           payload,
	}
}
