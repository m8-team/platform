package events

import (
	"encoding/json"
	"time"
)

// Envelope is the canonical outbox payload shape published by Resource Manager.
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
