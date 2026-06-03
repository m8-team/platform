package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/m8platform/platform/internal/platform/health"
	healthhttp "github.com/m8platform/platform/internal/platform/health/adapters/http"
	"go.uber.org/fx"
)

type HealthHTTPConfig struct {
	Address string
}

func healthHTTPModule(cfg HealthHTTPConfig) fx.Option {
	return fx.Module(
		"resource-manager-health-http",
		fx.Supply(cfg.normalized()),
		fx.Invoke(registerHealthHTTPServer),
	)
}

func registerHealthHTTPServer(lifecycle fx.Lifecycle, registry health.Registry, cfg HealthHTTPConfig) error {
	cfg = cfg.normalized()
	if cfg.Address == "" {
		return fmt.Errorf("%w: health http address is empty", ErrInvalidConfigValue)
	}

	mux := http.NewServeMux()
	healthhttp.NewHandler(registry).RegisterRoutes(mux)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: mux,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			listener, err := net.Listen("tcp", cfg.Address)
			if err != nil {
				return fmt.Errorf("listen health http %s: %w", cfg.Address, err)
			}

			go func() {
				_ = server.Serve(listener)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := server.Shutdown(ctx); err != nil {
				return fmt.Errorf("shutdown health http: %w", err)
			}

			return nil
		},
	})

	return nil
}

func (c HealthHTTPConfig) normalized() HealthHTTPConfig {
	c.Address = strings.TrimSpace(c.Address)
	if c.Address == "" {
		c.Address = defaultHealthHTTPAddress
	}

	return c
}
