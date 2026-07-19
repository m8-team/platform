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

type HTTPConfig struct {
	Address string
}

type resourceManagerHTTPHandler struct {
	http.Handler
}

type healthHTTPHandler struct {
	http.Handler
}

const yaRuHealthCheckName = "Ping ya.ru"

func resourceManagerHTTPModule(cfg HTTPConfig) fx.Option {
	return fx.Module(
		"resource-manager-http",
		fx.Supply(cfg.normalized()),
		fx.Provide(newResourceManagerHTTPHandler),
		fx.Invoke(registerResourceManagerHTTPServer),
	)
}

func healthHTTPModule(cfg HealthHTTPConfig) fx.Option {
	return fx.Module(
		"resource-manager-health-http",
		fx.Supply(cfg.normalized()),
		fx.Provide(newHealthHTTPHandler),
		fx.Invoke(registerResourceManagerHealthChecks),
		fx.Invoke(registerHealthHTTPServer),
	)
}

func newResourceManagerHTTPHandler(
	organizationServer resourcemanagerpb.OrganizationServiceServer,
) (resourceManagerHTTPHandler, error) {
	if organizationServer == nil {
		return resourceManagerHTTPHandler{}, errors.New("organization HTTP service is required")
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
		return resourceManagerHTTPHandler{}, fmt.Errorf("register organization HTTP gateway: %w", err)
	}

	return resourceManagerHTTPHandler{Handler: gateway}, nil
}

func newHealthHTTPHandler(registry health.Registry) healthHTTPHandler {
	mux := http.NewServeMux()
	healthhttp.NewHandler(registry).RegisterRoutes(mux)
	return healthHTTPHandler{Handler: mux}
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

func registerResourceManagerHTTPServer(
	lifecycle fx.Lifecycle,
	handler resourceManagerHTTPHandler,
	cfg HTTPConfig,
) error {
	cfg = cfg.normalized()
	return registerHTTPServer(lifecycle, "resource manager http", cfg.Address, handler.Handler)
}

func registerHealthHTTPServer(lifecycle fx.Lifecycle, handler healthHTTPHandler, cfg HealthHTTPConfig) error {
	cfg = cfg.normalized()
	return registerHTTPServer(lifecycle, "health http", cfg.Address, handler.Handler)
}

func registerHTTPServer(lifecycle fx.Lifecycle, name, address string, handler http.Handler) error {
	if address == "" {
		return fmt.Errorf("%w: %s address is empty", ErrInvalidConfigValue, name)
	}

	server := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			listener, err := net.Listen("tcp", address)
			if err != nil {
				return fmt.Errorf("listen %s %s: %w", name, address, err)
			}

			go func() {
				_ = server.Serve(listener)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := server.Shutdown(ctx); err != nil {
				return fmt.Errorf("shutdown %s: %w", name, err)
			}

			return nil
		},
	})

	return nil
}

func (c HTTPConfig) normalized() HTTPConfig {
	c.Address = strings.TrimSpace(c.Address)
	if c.Address == "" {
		c.Address = defaultHTTPAddress
	}

	return c
}

func (c HealthHTTPConfig) normalized() HealthHTTPConfig {
	c.Address = strings.TrimSpace(c.Address)
	if c.Address == "" {
		c.Address = defaultHealthHTTPAddress
	}

	return c
}
