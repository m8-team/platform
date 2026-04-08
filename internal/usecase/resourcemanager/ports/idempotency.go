package ports

import (
	"context"
	"time"
)

type IdempotencyReservation struct {
	Scope     string
	Key       string
	Duplicate bool
}

type IdempotencyStore interface {
	Reserve(ctx context.Context, scope string, key string, ttl time.Duration) (IdempotencyReservation, error)
	MarkCompleted(ctx context.Context, reservation IdempotencyReservation) error
}
