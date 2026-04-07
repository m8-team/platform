package outbox

import (
	"context"
	"sync"

	"github.com/m8platform/platform/internal/ports"
)

// Store is an in-memory placeholder for an outbox table persisted in the same
// transaction as aggregate writes.
type Store struct {
	mu      sync.Mutex
	records []Record
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Append(_ context.Context, record ports.OutboxRecord) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, record)
	return nil
}

func (s *Store) Records() []Record {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Record, len(s.records))
	copy(out, s.records)
	return out
}
