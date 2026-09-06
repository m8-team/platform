package main

import (
	"github.com/m8-team/platform/internal/platform/bootstrap"
	"github.com/m8-team/platform/internal/platform/health"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"github.com/m8-team/platform/internal/platform/telemetry"
	"github.com/m8-team/platform/internal/resourcemanager"
	grpcadapter "github.com/m8-team/platform/internal/resourcemanager/adapter/grpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func NewApp(cfg Config) *fx.App {
	return fx.New(appOptions(cfg)...)
}

func appOptions(cfg Config) []fx.Option {
	options := []fx.Option{
		telemetryModule(cfg.TraceEndpoint),
		fx.Provide(bootstrap.NewSupervisor),
		health.FxModule,
		resourceManagerHTTPModule(cfg.HTTP),
		healthHTTPModule(cfg.HealthHTTP),
		grpcserver.Module(cfg.GRPC, grpc.ChainUnaryInterceptor(telemetry.Unary)),
		resourceManagerModule(resourcemanager.Config{
			ServiceName:         "resource-manager",
			Debug:               cfg.Debug,
			SoftDeleteRetention: cfg.SoftDeleteRetention,
			PageTokenKey:        cfg.PageTokenKey,
		}, cfg.AllowUnauthenticated),
		grpcadapter.Module(),
	}

	if !cfg.Debug {
		options = append([]fx.Option{fx.NopLogger}, options...)
	}

	return options
}
