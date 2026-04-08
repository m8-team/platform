package port

import (
	"context"
	"encoding/json"
	"time"
)

type OutboxRecord struct {
	ID                string
	EventType         string
	EventVersion      int
	AggregateType     string
	AggregateID       string
	ParentAggregateID string
	OccurredAt        time.Time
	Actor             string
	CorrelationID     string
	CausationID       string
	IdempotencyKey    string
	ETagOrRevision    string
	Payload           json.RawMessage
}

type OutboxWriter interface {
	Append(ctx context.Context, record OutboxRecord) error
}
