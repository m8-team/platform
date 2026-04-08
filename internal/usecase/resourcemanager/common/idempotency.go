package common

import (
	"context"
	"fmt"
	"time"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

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
		return ports.IdempotencyReservation{}, ErrDuplicateRequest
	}
	return reservation, nil
}

func CompleteIdempotency(
	ctx context.Context,
	store ports.IdempotencyStore,
	reservation ports.IdempotencyReservation,
) error {
	if store == nil || reservation.Key == "" {
		return nil
	}
	return store.MarkCompleted(ctx, reservation)
}
