package outbox

import (
	"context"

	eventadapter "github.com/m8platform/platform/internal/adapters/outbound/events"
)

type Dispatcher struct {
	Writer    *Writer
	Publisher *eventadapter.Publisher
}

func (d Dispatcher) RunOnce(ctx context.Context) error {
	if d.Writer == nil || d.Publisher == nil {
		return nil
	}
	for _, record := range d.Writer.Snapshot() {
		if err := d.Publisher.Publish(ctx, eventadapter.MapOutboxRecord(record)); err != nil {
			return err
		}
	}
	return nil
}
