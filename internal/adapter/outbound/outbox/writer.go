package outbox

import (
	"context"
	"sync"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type Writer struct {
	mu      sync.Mutex
	records []port.OutboxRecord
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Append(_ context.Context, record port.OutboxRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.records = append(w.records, record)
	return nil
}

func (w *Writer) Snapshot() []port.OutboxRecord {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]port.OutboxRecord, len(w.records))
	copy(out, w.records)
	return out
}
