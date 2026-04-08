package events

import (
	"context"
	"encoding/json"

	"github.com/m8platform/platform/internal/frameworks/broker"
)

type Publisher struct {
	Client broker.Client
}

func (p Publisher) Publish(ctx context.Context, envelope Envelope) error {
	if p.Client == nil {
		return nil
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return p.Client.Publish(ctx, envelope.EventType, payload)
}
