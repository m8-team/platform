package main

import (
	"testing"

	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func TestNewAppBuilds(t *testing.T) {
	app := NewApp(Config{
		HealthHTTP: HealthHTTPConfig{Address: "127.0.0.1:0"},
		GRPC:       grpcserver.Config{Address: "127.0.0.1:0"},
	})
	if err := app.Err(); err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
}

func TestAppRegistersOrganizationServiceBeforeStart(t *testing.T) {
	cfg := Config{
		HealthHTTP: HealthHTTPConfig{Address: "127.0.0.1:0"},
		GRPC:       grpcserver.Config{Address: "127.0.0.1:0"},
	}
	var server *grpc.Server
	options := append(appOptions(cfg), fx.Populate(&server))
	app := fx.New(options...)
	if err := app.Err(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	serviceName := resourcemanagerpb.OrganizationService_ServiceDesc.ServiceName
	if _, exists := server.GetServiceInfo()[serviceName]; !exists {
		t.Fatalf("gRPC service %q is not registered", serviceName)
	}
}
