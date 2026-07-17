package idempotency

import (
	"errors"
	"sync"
)

var ErrConflict = errors.New("idempotency key was used with a different request")

type Record struct {
	RequestHash string
	ResultID    string
}

type Store interface {
	Get(scope, key string) (Record, bool)
	Put(scope, key string, record Record) error
}

type MemoryStore struct {
	mu      sync.Mutex
	records map[string]Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Record)}
}

func (s *MemoryStore) Get(scope, key string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[scope+":"+key]
	return rec, ok
}

func (s *MemoryStore) Put(scope, key string, record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	compound := scope + ":" + key
	if existing, ok := s.records[compound]; ok && existing.RequestHash != record.RequestHash {
		return ErrConflict
	}
	s.records[compound] = record
	return nil
}
