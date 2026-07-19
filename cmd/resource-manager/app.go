package main

import (
	"github.com/m8-team/platform/internal/platform/health"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"github.com/m8-team/platform/internal/resourcemanager"
	grpcadapter "github.com/m8-team/platform/internal/resourcemanager/adapter/grpc"
	"go.uber.org/fx"
)

func NewApp(cfg Config) *fx.App {
	return fx.New(appOptions(cfg)...)
}

func appOptions(cfg Config) []fx.Option {
	options := []fx.Option{
		health.FxModule,
		healthHTTPModule(cfg.HealthHTTP),
		grpcserver.Module(cfg.GRPC),
		resourcemanager.Module(resourcemanager.Config{
			ServiceName:          "resource-manager",
			Debug:                cfg.Debug,
			AllowUnauthenticated: cfg.AllowUnauthenticated,
			SoftDeleteRetention:  cfg.SoftDeleteRetention,
			PageTokenKey:         cfg.PageTokenKey,
		}),
		grpcadapter.Module(),
	}

	if !cfg.Debug {
		options = append([]fx.Option{fx.NopLogger}, options...)
	}

	return options
}
