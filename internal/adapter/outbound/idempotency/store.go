package idempotency

import (
	"context"
	"sync"
	"time"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type entry struct {
	expiresAt time.Time
	completed bool
}

type Store struct {
	mu      sync.Mutex
	clock   port.Clock
	entries map[string]entry
}

func NewStore(clock port.Clock) *Store {
	return &Store{
		clock:   clock,
		entries: make(map[string]entry),
	}
}

func (s *Store) Reserve(_ context.Context, scope string, key string, ttl time.Duration) (port.IdempotencyReservation, error) {
	if key == "" {
		return port.IdempotencyReservation{}, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if s.clock != nil {
		now = s.clock.Now().UTC()
	}
	composite := scope + ":" + key
	if current, ok := s.entries[composite]; ok && current.expiresAt.After(now) {
		return port.IdempotencyReservation{
			Scope:     scope,
			Key:       key,
			Duplicate: true,
		}, nil
	}
	s.entries[composite] = entry{expiresAt: now.Add(ttl)}
	return port.IdempotencyReservation{Scope: scope, Key: key}, nil
}

func (s *Store) MarkCompleted(_ context.Context, reservation port.IdempotencyReservation) error {
	if reservation.Key == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	composite := reservation.Scope + ":" + reservation.Key
	current := s.entries[composite]
	current.completed = true
	s.entries[composite] = current
	return nil
}
