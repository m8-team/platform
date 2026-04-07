package idempotency

import (
	"context"
	"sync"
	"time"

	"github.com/m8platform/platform/internal/ports"
)

// Store is an in-memory placeholder. Production wiring should replace it with
// a transactional PostgreSQL-backed implementation.
type Store struct {
	mu      sync.Mutex
	records map[string]time.Time
}

func NewStore() *Store {
	return &Store{records: make(map[string]time.Time)}
}

func (s *Store) Reserve(_ context.Context, scope string, key string, ttl time.Duration) (ports.IdempotencyReservation, error) {
	if s == nil {
		return ports.IdempotencyReservation{}, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	compound := scope + ":" + key
	if expiresAt, ok := s.records[compound]; ok && expiresAt.After(time.Now().UTC()) {
		return ports.IdempotencyReservation{Scope: scope, Key: key, Duplicate: true}, nil
	}
	s.records[compound] = time.Now().UTC().Add(ttl)
	return ports.IdempotencyReservation{Scope: scope, Key: key}, nil
}

func (s *Store) MarkCompleted(_ context.Context, _ ports.IdempotencyReservation) error {
	return nil
}
