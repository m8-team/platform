package ports

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, record OutboxRecord) error
}
