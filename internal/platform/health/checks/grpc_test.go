package checks

import (
	"context"
	"errors"
	"testing"

	"github.com/m8platform/platform/internal/platform/health"
	"google.golang.org/grpc"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

func TestGRPCHealthCheckerStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status grpc_health_v1.HealthCheckResponse_ServingStatus
		want   health.Status
	}{
		{name: "serving healthy", status: grpc_health_v1.HealthCheckResponse_SERVING, want: health.StatusHealthy},
		{name: "not serving unhealthy", status: grpc_health_v1.HealthCheckResponse_NOT_SERVING, want: health.StatusUnhealthy},
		{name: "unknown unhealthy", status: grpc_health_v1.HealthCheckResponse_UNKNOWN, want: health.StatusUnhealthy},
		{name: "service unknown unhealthy", status: grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN, want: health.StatusUnhealthy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewGRPCHealthChecker("dependency", fakeHealthClient{status: tt.status}, "svc")
			result := checker.Check(context.Background())
			if result.Status != tt.want {
				t.Fatalf("Status = %s, want %s", result.Status, tt.want)
			}
		})
	}
}

func TestGRPCHealthCheckerError(t *testing.T) {
	checker := NewGRPCHealthChecker("dependency", fakeHealthClient{err: errors.New("down")}, "svc")

	result := checker.Check(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
	if result.Error == "" {
		t.Fatal("Error is empty, want grpc error")
	}
}

func TestGRPCHealthCheckerNilClient(t *testing.T) {
	checker := NewGRPCHealthChecker("dependency", nil, "svc")

	result := checker.Check(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
	if result.Error != errGRPCHealthClientRequired.Error() {
		t.Fatalf("Error = %q, want %q", result.Error, errGRPCHealthClientRequired.Error())
	}
}

type fakeHealthClient struct {
	status grpc_health_v1.HealthCheckResponse_ServingStatus
	err    error
}

func (c fakeHealthClient) Check(context.Context, *grpc_health_v1.HealthCheckRequest, ...grpc.CallOption) (*grpc_health_v1.HealthCheckResponse, error) {
	if c.err != nil {
		return nil, c.err
	}

	return &grpc_health_v1.HealthCheckResponse{Status: c.status}, nil
}

func (c fakeHealthClient) List(context.Context, *grpc_health_v1.HealthListRequest, ...grpc.CallOption) (*grpc_health_v1.HealthListResponse, error) {
	return nil, errors.New("list is not implemented")
}

func (c fakeHealthClient) Watch(context.Context, *grpc_health_v1.HealthCheckRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[grpc_health_v1.HealthCheckResponse], error) {
	return nil, errors.New("watch is not implemented")
}
