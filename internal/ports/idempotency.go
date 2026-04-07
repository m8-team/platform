package ports

import (
	"context"
	"time"
)

// IdempotencyReservation is returned when a mutating command attempts to
// reserve a transport-provided idempotency key.
type IdempotencyReservation struct {
	Scope string
	Key   string

	// Duplicate is true when a previous request already reserved or completed
	// the same key in the same scope.
	Duplicate bool
}

type IdempotencyStore interface {
	Reserve(ctx context.Context, scope string, key string, ttl time.Duration) (IdempotencyReservation, error)
	MarkCompleted(ctx context.Context, reservation IdempotencyReservation) error
}
