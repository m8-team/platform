package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

const (
	DefaultIdempotencyTTL = 24 * time.Hour
	DefaultPurgeWindow    = 30 * 24 * time.Hour
)

func WriteOutboxRecord(
	ctx context.Context,
	writer ports.OutboxWriter,
	record ports.OutboxRecord,
) error {
	if writer == nil {
		return nil
	}
	return writer.Append(ctx, record)
}

func NewOutboxRecord(
	uuid ports.UUIDGenerator,
	metadata boundaries.RequestMetadata,
	eventType string,
	aggregateType string,
	aggregateID string,
	parentAggregateID string,
	etag string,
	occurredAt time.Time,
	payload any,
) (ports.OutboxRecord, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return ports.OutboxRecord{}, fmt.Errorf("marshal payload: %w", err)
	}

	id := ""
	if uuid != nil {
		id = uuid.NewString()
	}

	return ports.OutboxRecord{
		ID:                id,
		EventType:         eventType,
		EventVersion:      1,
		AggregateType:     aggregateType,
		AggregateID:       aggregateID,
		ParentAggregateID: parentAggregateID,
		OccurredAt:        occurredAt.UTC(),
		Actor:             metadata.Actor,
		CorrelationID:     metadata.CorrelationID,
		CausationID:       metadata.CausationID,
		IdempotencyKey:    metadata.IdempotencyKey,
		ETagOrRevision:    etag,
		Payload:           body,
	}, nil
}
