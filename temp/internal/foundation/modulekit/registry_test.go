package modulekit

import "testing"

type stubModule struct {
	name string
}

func (m stubModule) Name() string { return m.name }

func (m stubModule) RegisterHTTP(reg HTTPRegistrar) {
	reg.RegisterHTTPRoute(HTTPRoute{Method: "GET", Pattern: "/" + m.name})
}

func (m stubModule) RegisterGRPC(reg GRPCRegistrar) {
	reg.RegisterGRPCService(GRPCService{Name: m.name})
}

func (m stubModule) RegisterConsumers(reg ConsumerRegistrar) {
	reg.RegisterConsumer(ConsumerRegistration{Name: m.name})
}

func (m stubModule) RegisterWorkers(reg WorkerRegistrar) {
	reg.RegisterWorker(WorkerRegistration{Name: m.name})
}

func TestNewRegistryRegistersModules(t *testing.T) {
	registry := NewRegistry(stubModule{name: "tenant"}, stubModule{name: "iam"})

	if got := len(registry.Modules()); got != 2 {
		t.Fatalf("module count = %d, want 2", got)
	}
	if got := len(registry.HTTP().Routes()); got != 2 {
		t.Fatalf("http route count = %d, want 2", got)
	}
	if got := len(registry.GRPC().Services()); got != 2 {
		t.Fatalf("grpc service count = %d, want 2", got)
	}
	if got := len(registry.Consumers().Consumers()); got != 2 {
		t.Fatalf("consumer count = %d, want 2", got)
	}
	if got := len(registry.Workers().Workers()); got != 2 {
		t.Fatalf("worker count = %d, want 2", got)
	}
}
