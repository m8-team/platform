package checks

import (
	"context"
	nethttp "net/http"

	"github.com/m8platform/platform/internal/platform/health"
)

type httpChecker struct {
	name   string
	client *nethttp.Client
	url    string
}

func NewHTTPChecker(name string, client *nethttp.Client, url string) health.Checker {
	if client == nil {
		client = nethttp.DefaultClient
	}

	return &httpChecker{
		name:   name,
		client: client,
		url:    url,
	}
}

func (c *httpChecker) Check(ctx context.Context) health.Result {
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, c.url, nil)
	if err != nil {
		return health.Result{
			Name:    c.name,
			Status:  health.StatusUnhealthy,
			Message: "http health request failed",
			Error:   err.Error(),
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return health.Result{
			Name:    c.name,
			Status:  health.StatusUnhealthy,
			Message: "http health request failed",
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return health.Result{
			Name:   c.name,
			Status: health.StatusHealthy,
			Metadata: map[string]string{
				"status_code": resp.Status,
			},
		}
	}

	return health.Result{
		Name:    c.name,
		Status:  health.StatusUnhealthy,
		Message: "http health endpoint returned non-success status",
		Error:   resp.Status,
		Metadata: map[string]string{
			"status_code": resp.Status,
		},
	}
}
