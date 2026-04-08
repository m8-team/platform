package outbox

import (
	"context"
	"sync"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type Writer struct {
	mu      sync.Mutex
	records []ports.OutboxRecord
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Append(_ context.Context, record ports.OutboxRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.records = append(w.records, record)
	return nil
}

func (w *Writer) Snapshot() []ports.OutboxRecord {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]ports.OutboxRecord, len(w.records))
	copy(out, w.records)
	return out
}
