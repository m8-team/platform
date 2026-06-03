package checks

import (
	"context"
	"errors"

	"github.com/m8platform/platform/internal/platform/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

var ErrGRPCHealthClientRequired = errors.New("grpc health client is required")

type grpcHealthChecker struct {
	name    string
	client  grpc_health_v1.HealthClient
	service string
}

func NewGRPCHealthChecker(name string, client grpc_health_v1.HealthClient, service string) health.Checker {
	return &grpcHealthChecker{
		name:    name,
		client:  client,
		service: service,
	}
}

func (c *grpcHealthChecker) Check(ctx context.Context) health.Result {
	if c.client == nil {
		return health.Result{
			Name:    c.name,
			Status:  health.StatusUnhealthy,
			Message: "grpc health client is not configured",
			Error:   ErrGRPCHealthClientRequired.Error(),
		}
	}

	resp, err := c.client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: c.service})
	if err != nil {
		return health.Result{
			Name:    c.name,
			Status:  health.StatusUnhealthy,
			Message: "grpc health check failed",
			Error:   err.Error(),
		}
	}

	if resp.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING {
		return health.Result{
			Name:   c.name,
			Status: health.StatusHealthy,
			Metadata: map[string]string{
				"service": c.service,
				"status":  resp.GetStatus().String(),
			},
		}
	}

	return health.Result{
		Name:    c.name,
		Status:  health.StatusUnhealthy,
		Message: "grpc health service is not serving",
		Error:   resp.GetStatus().String(),
		Metadata: map[string]string{
			"service": c.service,
			"status":  resp.GetStatus().String(),
		},
	}
}
