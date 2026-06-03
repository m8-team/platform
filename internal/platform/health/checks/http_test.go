package checks

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/m8platform/platform/internal/platform/health"
)

func TestHTTPCheckerStatuses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       health.Status
	}{
		{name: "2xx healthy", statusCode: nethttp.StatusNoContent, want: health.StatusHealthy},
		{name: "3xx healthy", statusCode: nethttp.StatusTemporaryRedirect, want: health.StatusHealthy},
		{name: "4xx unhealthy", statusCode: nethttp.StatusBadRequest, want: health.StatusUnhealthy},
		{name: "5xx unhealthy", statusCode: nethttp.StatusInternalServerError, want: health.StatusUnhealthy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			checker := NewHTTPChecker("dependency", server.Client(), server.URL)
			result := checker.Check(context.Background())
			if result.Status != tt.want {
				t.Fatalf("Status = %s, want %s", result.Status, tt.want)
			}
		})
	}
}

func TestHTTPCheckerUsesRequestContext(t *testing.T) {
	server := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		<-r.Context().Done()
		w.WriteHeader(nethttp.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	checker := NewHTTPChecker("dependency", server.Client(), server.URL)
	result := checker.Check(ctx)
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
}

func TestHTTPCheckerNetworkError(t *testing.T) {
	server := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusOK)
	}))
	url := server.URL
	server.Close()

	checker := NewHTTPChecker("dependency", server.Client(), url)
	result := checker.Check(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
	if result.Error == "" {
		t.Fatal("Error is empty, want network error")
	}
}

func TestHTTPCheckerInvalidURL(t *testing.T) {
	checker := NewHTTPChecker("dependency", nil, "http://%zz")

	result := checker.Check(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
	if result.Error == "" {
		t.Fatal("Error is empty, want invalid URL error")
	}
}
