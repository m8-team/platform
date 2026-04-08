package events

import (
	"encoding/json"
	"time"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type Envelope struct {
	EventID           string          `json:"event_id"`
	EventType         string          `json:"event_type"`
	EventVersion      int             `json:"event_version"`
	AggregateType     string          `json:"aggregate_type"`
	AggregateID       string          `json:"aggregate_id"`
	ParentAggregateID string          `json:"parent_aggregate_id,omitempty"`
	OccurredAt        time.Time       `json:"occurred_at"`
	Actor             string          `json:"actor,omitempty"`
	CorrelationID     string          `json:"correlation_id,omitempty"`
	CausationID       string          `json:"causation_id,omitempty"`
	IdempotencyKey    string          `json:"idempotency_key,omitempty"`
	ETagOrRevision    string          `json:"etag_or_revision,omitempty"`
	Payload           json.RawMessage `json:"payload"`
}

func MapOutboxRecord(record ports.OutboxRecord) Envelope {
	return Envelope{
		EventID:           record.ID,
		EventType:         record.EventType,
		EventVersion:      record.EventVersion,
		AggregateType:     record.AggregateType,
		AggregateID:       record.AggregateID,
		ParentAggregateID: record.ParentAggregateID,
		OccurredAt:        record.OccurredAt,
		Actor:             record.Actor,
		CorrelationID:     record.CorrelationID,
		CausationID:       record.CausationID,
		IdempotencyKey:    record.IdempotencyKey,
		ETagOrRevision:    record.ETagOrRevision,
		Payload:           record.Payload,
	}
}
