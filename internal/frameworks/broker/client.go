package broker

import "context"

type Client interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

type NopClient struct{}

func (NopClient) Publish(context.Context, string, []byte) error {
	return nil
}
