package ports

import (
	"context"
	"encoding/json"
	"time"
)

// OutboxRecord is the storage-facing representation of a domain event that
// must be published asynchronously.
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

type EventPublisher interface {
	Publish(ctx context.Context, record OutboxRecord) error
}
