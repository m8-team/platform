package modulekit

type ConsumerRegistration struct {
	Name    string
	Handler any
}

type ConsumerRegistrar interface {
	RegisterConsumer(consumer ConsumerRegistration)
}

type ConsumerRegistry struct {
	consumers []ConsumerRegistration
}

func NewConsumerRegistry() *ConsumerRegistry {
	return &ConsumerRegistry{consumers: make([]ConsumerRegistration, 0)}
}

func (r *ConsumerRegistry) RegisterConsumer(consumer ConsumerRegistration) {
	if r == nil {
		return
	}
	r.consumers = append(r.consumers, consumer)
}

func (r *ConsumerRegistry) Consumers() []ConsumerRegistration {
	if r == nil {
		return nil
	}
	consumers := make([]ConsumerRegistration, len(r.consumers))
	copy(consumers, r.consumers)
	return consumers
}
