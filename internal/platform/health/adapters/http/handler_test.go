package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/m8platform/platform/internal/platform/health"
)

func TestHandlerLivezReturnsJSONSnapshot(t *testing.T) {
	registry := health.NewRegistry()
	if err := registry.Register(check("app", health.KindLiveness, health.StatusHealthy, health.CriticalityRequired)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	resp := request(registry, nethttp.MethodGet, "/livez")
	if resp.Code != nethttp.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.Code, nethttp.StatusOK)
	}
	if got := resp.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var snapshot health.Snapshot
	if err := json.Unmarshal(resp.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("JSON decode error = %v", err)
	}
	if snapshot.Kind != health.KindLiveness {
		t.Fatalf("Snapshot kind = %s, want %s", snapshot.Kind, health.KindLiveness)
	}
	if snapshot.Status != health.StatusHealthy {
		t.Fatalf("Snapshot status = %s, want %s", snapshot.Status, health.StatusHealthy)
	}
}

func TestHandlerStatusCodes(t *testing.T) {
	tests := []struct {
		name        string
		checkStatus health.Status
		criticality health.Criticality
		wantStatus  health.Status
		wantCode    int
	}{
		{
			name:        "healthy",
			checkStatus: health.StatusHealthy,
			criticality: health.CriticalityRequired,
			wantStatus:  health.StatusHealthy,
			wantCode:    nethttp.StatusOK,
		},
		{
			name:        "degraded",
			checkStatus: health.StatusUnhealthy,
			criticality: health.CriticalityOptional,
			wantStatus:  health.StatusDegraded,
			wantCode:    nethttp.StatusOK,
		},
		{
			name:        "unhealthy",
			checkStatus: health.StatusUnhealthy,
			criticality: health.CriticalityRequired,
			wantStatus:  health.StatusUnhealthy,
			wantCode:    nethttp.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := health.NewRegistry()
			if err := registry.Register(check("dependency", health.KindReadiness, tt.checkStatus, tt.criticality)); err != nil {
				t.Fatalf("Register() error = %v", err)
			}

			resp := request(registry, nethttp.MethodGet, "/readyz")
			if resp.Code != tt.wantCode {
				t.Fatalf("status code = %d, want %d", resp.Code, tt.wantCode)
			}

			var snapshot health.Snapshot
			if err := json.Unmarshal(resp.Body.Bytes(), &snapshot); err != nil {
				t.Fatalf("JSON decode error = %v", err)
			}
			if snapshot.Status != tt.wantStatus {
				t.Fatalf("Snapshot status = %s, want %s", snapshot.Status, tt.wantStatus)
			}
		})
	}
}

func TestHandlerUnsupportedMethod(t *testing.T) {
	registry := health.NewRegistry()
	resp := request(registry, nethttp.MethodPost, "/livez")

	if resp.Code != nethttp.StatusMethodNotAllowed {
		t.Fatalf("status code = %d, want %d", resp.Code, nethttp.StatusMethodNotAllowed)
	}
	if got := resp.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var snapshot health.Snapshot
	if err := json.Unmarshal(resp.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("JSON decode error = %v", err)
	}
	if snapshot.Kind != health.KindLiveness {
		t.Fatalf("Snapshot kind = %s, want %s", snapshot.Kind, health.KindLiveness)
	}
}

func request(registry health.Registry, method string, path string) *httptest.ResponseRecorder {
	mux := nethttp.NewServeMux()
	NewHandler(registry).RegisterRoutes(mux)

	req := httptest.NewRequest(method, path, nil)
	resp := httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	return resp
}

func check(name string, kind health.Kind, status health.Status, criticality health.Criticality) health.Check {
	return health.Check{
		Spec: health.CheckSpec{
			Name:        name,
			Kinds:       []health.Kind{kind},
			Criticality: criticality,
		},
		Checker: health.CheckerFunc(func(context.Context) health.Result {
			return health.Result{Status: status}
		}),
	}
}
