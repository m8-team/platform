package checks

import (
	"context"
	nethttp "net/http"

	"github.com/m8platform/platform/internal/platform/health"
)

func NewHTTPCheck(name string, client *nethttp.Client, url string) health.Check {
	if client == nil {
		client = nethttp.DefaultClient
	}

	return func(ctx context.Context) health.Result {
		req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, url, nil)
		if err != nil {
			return health.Result{
				Name:    name,
				Status:  health.StatusUnhealthy,
				Message: "http health request failed",
				Error:   err.Error(),
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			return health.Result{
				Name:    name,
				Status:  health.StatusUnhealthy,
				Message: "http health request failed",
				Error:   err.Error(),
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return health.Result{
				Name:   name,
				Status: health.StatusHealthy,
			}
		}

		return health.Result{
			Name:    name,
			Status:  health.StatusUnhealthy,
			Message: "http health endpoint returned non-success status",
			Error:   resp.Status,
		}
	}
}
