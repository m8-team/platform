package events

import (
	"context"

	"github.com/m8platform/platform/internal/ports"
)

// Publisher is a placeholder adapter for the future event bus integration.
type Publisher struct{}

func (Publisher) Publish(_ context.Context, _ ports.OutboxRecord) error {
	return ports.ErrNotImplemented
}
