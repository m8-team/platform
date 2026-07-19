package checks

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/platform/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

var errGRPCHealthClientRequired = errors.New("grpc health client is required")

func NewGRPCHealthCheck(name string, client grpc_health_v1.HealthClient, service string) health.Check {
	return func(ctx context.Context) health.Result {
		if client == nil {
			return health.Result{
				Name:    name,
				Status:  health.StatusUnhealthy,
				Message: "grpc health client is not configured",
				Error:   errGRPCHealthClientRequired.Error(),
			}
		}

		resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
		if err != nil {
			return health.Result{
				Name:    name,
				Status:  health.StatusUnhealthy,
				Message: "grpc health check failed",
				Error:   err.Error(),
			}
		}

		if resp.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING {
			return health.Result{
				Name:   name,
				Status: health.StatusHealthy,
			}
		}

		return health.Result{
			Name:    name,
			Status:  health.StatusUnhealthy,
			Message: "grpc health service is not serving",
			Error:   resp.GetStatus().String(),
		}
	}
}
