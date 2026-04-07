package common

import (
	"context"
	"fmt"
	"time"

	"github.com/m8platform/platform/internal/ports"
)

// ReserveIdempotency reserves a request-scoped idempotency key when one is
// provided. Empty keys are treated as "idempotency disabled".
func ReserveIdempotency(
	ctx context.Context,
	store ports.IdempotencyStore,
	scope string,
	key string,
	ttl time.Duration,
) (ports.IdempotencyReservation, error) {
	if store == nil || key == "" {
		return ports.IdempotencyReservation{}, nil
	}

	reservation, err := store.Reserve(ctx, scope, key, ttl)
	if err != nil {
		return ports.IdempotencyReservation{}, fmt.Errorf("reserve idempotency key: %w", err)
	}
	if reservation.Duplicate {
		return reservation, ErrDuplicateRequest
	}

	return reservation, nil
}

// CompleteIdempotency marks a previously acquired reservation as completed.
func CompleteIdempotency(
	ctx context.Context,
	store ports.IdempotencyStore,
	reservation ports.IdempotencyReservation,
) error {
	if store == nil || reservation.Key == "" || reservation.Scope == "" {
		return nil
	}
	return store.MarkCompleted(ctx, reservation)
}
