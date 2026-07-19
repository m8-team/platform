package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/platform/health"
	healthhttp "github.com/m8-team/platform/internal/platform/health/adapters/http"
	"github.com/m8-team/platform/internal/platform/health/checks"
	"go.uber.org/fx"
	"google.golang.org/protobuf/encoding/protojson"
)

type HealthHTTPConfig struct {
	Address string
}

const yaRuHealthCheckName = "Ping ya.ru"

func healthHTTPModule(cfg HealthHTTPConfig) fx.Option {
	return fx.Module(
		"resource-manager-http",
		fx.Supply(cfg.normalized()),
		fx.Provide(newResourceManagerHTTPHandler),
		fx.Invoke(registerResourceManagerHealthChecks),
		fx.Invoke(registerHealthHTTPServer),
	)
}

func newResourceManagerHTTPHandler(
	registry health.Registry,
	organizationServer resourcemanagerpb.OrganizationServiceServer,
) (http.Handler, error) {
	if organizationServer == nil {
		return nil, errors.New("organization HTTP service is required")
	}

	gateway := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: false,
			},
		}),
	)
	if err := resourcemanagerpb.RegisterOrganizationServiceHandlerServer(
		context.Background(),
		gateway,
		organizationServer,
	); err != nil {
		return nil, fmt.Errorf("register organization HTTP gateway: %w", err)
	}

	mux := http.NewServeMux()
	healthhttp.NewHandler(registry).RegisterRoutes(mux)
	mux.Handle("/", gateway)

	return mux, nil
}

func registerResourceManagerHealthChecks(registry health.Registry) error {
	return health.Register(registry, health.Config{
		Spec: health.Spec{
			Name: yaRuHealthCheckName,
			Target: health.Target{
				Kind:   health.TargetKindDependency,
				Name:   "ya.ru",
				Module: "resource-manager",
			},
			Kinds:       []health.Kind{health.KindReadiness},
			Criticality: health.CriticalityOptional,
			Timeout:     1 * time.Second,
			Interval:    10 * time.Second,
		},
		Check: checks.NewHTTPCheck("ya.ru", http.DefaultClient, "https://ya.ru"),
	})
}

func registerHealthHTTPServer(lifecycle fx.Lifecycle, handler http.Handler, cfg HealthHTTPConfig) error {
	cfg = cfg.normalized()
	if cfg.Address == "" {
		return fmt.Errorf("%w: health http address is empty", ErrInvalidConfigValue)
	}

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: handler,
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
