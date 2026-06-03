package grpc

import (
	"context"
	"strings"
	"time"

	platformhealth "github.com/m8platform/platform/internal/platform/health"
	grpc_health "google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

const defaultPeriod = 5 * time.Second

type Adapter struct {
	registry     platformhealth.Registry
	server       *grpc_health.Server
	period       time.Duration
	serviceNames []string
}

type Option func(*Adapter)

func NewAdapter(registry platformhealth.Registry, opts ...Option) *Adapter {
	adapter := &Adapter{
		registry: registry,
		server:   grpc_health.NewServer(),
		period:   defaultPeriod,
	}

	for _, opt := range opts {
		opt(adapter)
	}
	if adapter.period <= 0 {
		adapter.period = defaultPeriod
	}

	return adapter
}

func WithPeriod(period time.Duration) Option {
	return func(adapter *Adapter) {
		adapter.period = period
	}
}

func WithServiceNames(names ...string) Option {
	return func(adapter *Adapter) {
		adapter.serviceNames = append(adapter.serviceNames, normalizeServiceNames(names)...)
	}
}

func (a *Adapter) Server() *grpc_health.Server {
	return a.server
}

func (a *Adapter) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	a.update(ctx)

	ticker := time.NewTicker(a.period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.server.Shutdown()
			return
		case <-ticker.C:
			a.update(ctx)
		}
	}
}

func (a *Adapter) update(ctx context.Context) {
	status := grpcStatus(platformhealth.StatusUnhealthy)
	if a.registry != nil {
		status = grpcStatus(a.registry.Snapshot(ctx, platformhealth.KindReadiness).Status)
	}

	a.server.SetServingStatus("", status)
	for _, serviceName := range a.serviceNames {
		a.server.SetServingStatus(serviceName, status)
	}
}

func grpcStatus(status platformhealth.Status) grpc_health_v1.HealthCheckResponse_ServingStatus {
	switch status {
	case platformhealth.StatusHealthy, platformhealth.StatusDegraded:
		return grpc_health_v1.HealthCheckResponse_SERVING
	case platformhealth.StatusUnhealthy:
		return grpc_health_v1.HealthCheckResponse_NOT_SERVING
	case platformhealth.StatusUnknown:
		return grpc_health_v1.HealthCheckResponse_UNKNOWN
	default:
		return grpc_health_v1.HealthCheckResponse_UNKNOWN
	}
}

func normalizeServiceNames(names []string) []string {
	result := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			result = append(result, name)
		}
	}

	return result
}
