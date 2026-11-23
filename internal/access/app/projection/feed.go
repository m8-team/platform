package projection

import "context"

// Delivery is a Kafka or Outbox event whose acknowledgement MUST be deferred
// until Process has durably committed its checkpoint.
type Delivery interface {
	Event() Event
	Ack(context.Context) error
}
type EventFeed interface {
	Next(context.Context) (Delivery, error)
}

func consumeEvents(ctx context.Context, feed EventFeed, process func(context.Context, Event) error) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		delivery, err := feed.Next(ctx)
		if err != nil {
			return err
		}
		if delivery == nil {
			return ErrInvalidEvent
		}
		if err := process(ctx, delivery.Event()); err != nil {
			return err
		}
		if err := delivery.Ack(ctx); err != nil {
			return err
		}
	}
}
