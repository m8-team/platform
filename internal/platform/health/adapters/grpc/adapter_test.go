package grpc

import (
	"context"
	"testing"

	platformhealth "github.com/m8platform/platform/internal/platform/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

func TestAdapterGRPCStatusMapping(t *testing.T) {
	tests := []struct {
		name string
		in   platformhealth.Status
		want grpc_health_v1.HealthCheckResponse_ServingStatus
	}{
		{
			name: "healthy serving",
			in:   platformhealth.StatusHealthy,
			want: grpc_health_v1.HealthCheckResponse_SERVING,
		},
		{
			name: "degraded serving",
			in:   platformhealth.StatusDegraded,
			want: grpc_health_v1.HealthCheckResponse_SERVING,
		},
		{
			name: "unhealthy not serving",
			in:   platformhealth.StatusUnhealthy,
			want: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		},
		{
			name: "unknown unknown",
			in:   platformhealth.StatusUnknown,
			want: grpc_health_v1.HealthCheckResponse_UNKNOWN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAdapter(fakeRegistry{status: tt.in}, WithServiceNames("svc"))
			adapter.update(context.Background())

			resp, err := adapter.Server().Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: "svc"})
			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}
			if resp.GetStatus() != tt.want {
				t.Fatalf("Serving status = %s, want %s", resp.GetStatus(), tt.want)
			}
		})
	}
}

type fakeRegistry struct {
	status platformhealth.Status
}

func (r fakeRegistry) Register(platformhealth.Check) error {
	return nil
}

func (r fakeRegistry) Snapshot(context.Context, platformhealth.CheckKind) platformhealth.Snapshot {
	return platformhealth.Snapshot{Status: r.status}
}
