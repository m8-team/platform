package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/platform/bootstrap"
	"github.com/m8-team/platform/internal/platform/telemetry"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/platform/health"
	healthhttp "github.com/m8-team/platform/internal/platform/health/adapters/http"
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
		fx.Invoke(registerHealthHTTPServer),
	)
}

func newResourceManagerHTTPHandler(
	organizationServer resourcemanagerpb.OrganizationServiceServer,
	workspaceServer resourcemanagerpb.WorkspaceServiceServer,
) (resourceManagerHTTPHandler, error) {
	if organizationServer == nil {
		return resourceManagerHTTPHandler{}, errors.New("organization HTTP service is required")
	}
	if workspaceServer == nil {
		return resourceManagerHTTPHandler{}, errors.New("workspace HTTP service is required")
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
	if err := resourcemanagerpb.RegisterWorkspaceServiceHandlerServer(
		context.Background(),
		gateway,
		workspaceServer,
	); err != nil {
		return resourceManagerHTTPHandler{}, fmt.Errorf("register workspace HTTP gateway: %w", err)
	}

	return resourceManagerHTTPHandler{Handler: telemetry.HTTP(gateway)}, nil
}

func newHealthHTTPHandler(registry health.Registry, observer *telemetry.Observer) healthHTTPHandler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", observer)
	healthhttp.NewHandler(registry).RegisterRoutes(mux)
	return healthHTTPHandler{Handler: mux}
}

func registerResourceManagerHTTPServer(
	lifecycle fx.Lifecycle,
	supervisor *bootstrap.Supervisor,
	handler resourceManagerHTTPHandler,
	cfg HTTPConfig,
) error {
	cfg = cfg.normalized()
	return registerHTTPServer(lifecycle, supervisor, "resource manager http", cfg.Address, handler.Handler)
}

func registerHealthHTTPServer(lifecycle fx.Lifecycle, supervisor *bootstrap.Supervisor, handler healthHTTPHandler, cfg HealthHTTPConfig) error {
	cfg = cfg.normalized()
	return registerHTTPServer(lifecycle, supervisor, "health http", cfg.Address, handler.Handler)
}

func registerHTTPServer(lifecycle fx.Lifecycle, supervisor *bootstrap.Supervisor, name, address string, handler http.Handler) error {
	if address == "" {
		return fmt.Errorf("%w: %s address is empty", ErrInvalidConfigValue, name)
	}

	server := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			listener, err := net.Listen("tcp", address)
			if err != nil {
				return fmt.Errorf("listen %s %s: %w", name, address, err)
			}

			supervisor.Go(name, func(context.Context) error {
				err := server.Serve(listener)
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			})

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := server.Shutdown(ctx); err != nil {
				_ = server.Close()
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
