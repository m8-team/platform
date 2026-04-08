package common

import (
	"context"
	"fmt"
	"time"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

func ReserveIdempotency(
	ctx context.Context,
	store port.IdempotencyStore,
	scope string,
	key string,
	ttl time.Duration,
) (port.IdempotencyReservation, error) {
	if store == nil || key == "" {
		return port.IdempotencyReservation{}, nil
	}
	reservation, err := store.Reserve(ctx, scope, key, ttl)
	if err != nil {
		return port.IdempotencyReservation{}, fmt.Errorf("reserve idempotency key: %w", err)
	}
	if reservation.Duplicate {
		return port.IdempotencyReservation{}, ErrDuplicateRequest
	}
	return reservation, nil
}

func CompleteIdempotency(
	ctx context.Context,
	store port.IdempotencyStore,
	reservation port.IdempotencyReservation,
) error {
	if store == nil || reservation.Key == "" {
		return nil
	}
	return store.MarkCompleted(ctx, reservation)
}
