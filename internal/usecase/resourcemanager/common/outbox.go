package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

const (
	DefaultIdempotencyTTL = 24 * time.Hour
	DefaultPurgeWindow    = 30 * 24 * time.Hour
)

func WriteOutboxRecord(
	ctx context.Context,
	writer port.OutboxWriter,
	record port.OutboxRecord,
) error {
	if writer == nil {
		return nil
	}
	return writer.Append(ctx, record)
}

func NewOutboxRecord(
	uuid port.UUIDGenerator,
	eventType string,
	aggregateType string,
	aggregateID string,
	parentAggregateID string,
	etag string,
	occurredAt time.Time,
	payload any,
) (port.OutboxRecord, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return port.OutboxRecord{}, fmt.Errorf("marshal payload: %w", err)
	}

	id := ""
	if uuid != nil {
		id = uuid.NewString()
	}

	return port.OutboxRecord{
		ID:                id,
		EventType:         eventType,
		EventVersion:      1,
		AggregateType:     aggregateType,
		AggregateID:       aggregateID,
		ParentAggregateID: parentAggregateID,
		OccurredAt:        occurredAt.UTC(),
		ETagOrRevision:    etag,
		Payload:           body,
	}, nil
}
