package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health "google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

func TestModuleRegistersAndServesBeforeStart(t *testing.T) {
	t.Parallel()

	healthServer := grpc_health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	registrationCalled := false
	registration := func() Registration {
		return func(registrar grpc.ServiceRegistrar) error {
			registrationCalled = true
			grpc_health_v1.RegisterHealthServer(registrar, healthServer)
			return nil
		}
	}

	var managed *Server
	var raw *grpc.Server
	app := fx.New(
		Module(Config{Address: "127.0.0.1:0"}),
		fx.Provide(fx.Annotate(registration, fx.ResultTags(RegistrationResultTag))),
		fx.Populate(&managed, &raw),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	if !registrationCalled {
		t.Fatal("registration was not called during application construction")
	}
	if _, ok := raw.GetServiceInfo()["grpc.health.v1.Health"]; !ok {
		t.Fatal("health service is not registered before application start")
	}
	if managed.Address() != nil {
		t.Fatalf("Address() before Start = %v, want nil", managed.Address())
	}

	startCtx, cancelStart := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStart()
	if err := app.Start(startCtx); err != nil {
		t.Fatalf("start app: %v", err)
	}
	stopped := false
	t.Cleanup(func() {
		if stopped {
			return
		}
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := app.Stop(stopCtx); err != nil {
			t.Errorf("cleanup app: %v", err)
		}
	})

	address := managed.Address()
	if address == nil {
		t.Fatal("Address() after Start = nil")
	}

	client := grpc_health_v1.NewHealthClient(mustDial(t, address.String()))
	requestCtx, cancelRequest := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelRequest()
	response, err := client.Check(requestCtx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("health Check() error = %v", err)
	}
	if response.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("health status = %s, want SERVING", response.GetStatus())
	}

	stopCtx, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStop()
	if err := app.Stop(stopCtx); err != nil {
		t.Fatalf("stop app: %v", err)
	}
	stopped = true
}

func TestModuleRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	app := fx.New(
		Module(Config{}),
		fx.NopLogger,
	)
	if !errors.Is(app.Err(), ErrAddressRequired) {
		t.Fatalf("app error = %v, want %v", app.Err(), ErrAddressRequired)
	}
}

func TestModulePropagatesRegistrationFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("registration failed")
	registration := func() Registration {
		return func(grpc.ServiceRegistrar) error {
			return wantErr
		}
	}

	app := fx.New(
		Module(Config{Address: "127.0.0.1:0"}),
		fx.Provide(fx.Annotate(registration, fx.ResultTags(RegistrationResultTag))),
		fx.NopLogger,
	)
	if !errors.Is(app.Err(), wantErr) {
		t.Fatalf("app error = %v, want %v", app.Err(), wantErr)
	}
}

func mustDial(t *testing.T, target string) *grpc.ClientConn {
	t.Helper()

	connection, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("create gRPC client: %v", err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close gRPC client: %v", err)
		}
	})

	return connection
}
