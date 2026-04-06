package modulekit

type Module interface {
	Name() string
	RegisterHTTP(reg HTTPRegistrar)
	RegisterGRPC(reg GRPCRegistrar)
	RegisterConsumers(reg ConsumerRegistrar)
	RegisterWorkers(reg WorkerRegistrar)
}

type Registry struct {
	modules   []Module
	http      *HTTPRegistry
	grpc      *GRPCRegistry
	consumers *ConsumerRegistry
	workers   *WorkerRegistry
}

func NewRegistry(modules ...Module) *Registry {
	registry := &Registry{
		modules:   make([]Module, 0, len(modules)),
		http:      NewHTTPRegistry(),
		grpc:      NewGRPCRegistry(),
		consumers: NewConsumerRegistry(),
		workers:   NewWorkerRegistry(),
	}

	for _, module := range modules {
		if module == nil {
			continue
		}
		registry.modules = append(registry.modules, module)
		module.RegisterHTTP(registry.http)
		module.RegisterGRPC(registry.grpc)
		module.RegisterConsumers(registry.consumers)
		module.RegisterWorkers(registry.workers)
	}

	return registry
}

func (r *Registry) Modules() []Module {
	if r == nil {
		return nil
	}
	modules := make([]Module, len(r.modules))
	copy(modules, r.modules)
	return modules
}

func (r *Registry) HTTP() *HTTPRegistry {
	if r == nil {
		return nil
	}
	return r.http
}

func (r *Registry) GRPC() *GRPCRegistry {
	if r == nil {
		return nil
	}
	return r.grpc
}

func (r *Registry) Consumers() *ConsumerRegistry {
	if r == nil {
		return nil
	}
	return r.consumers
}

func (r *Registry) Workers() *WorkerRegistry {
	if r == nil {
		return nil
	}
	return r.workers
}
