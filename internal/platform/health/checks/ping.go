package checks

import (
	"context"
	"errors"

	"github.com/m8platform/platform/internal/platform/health"
)

var ErrPingFuncRequired = errors.New("ping function is required")

type PingFunc func(ctx context.Context) error

type pingChecker struct {
	name string
	ping PingFunc
}

func NewPingChecker(name string, ping PingFunc) health.Checker {
	return &pingChecker{name: name, ping: ping}
}

func (c *pingChecker) Check(ctx context.Context) health.Result {
	if c.ping == nil {
		return health.Result{
			Name:    c.name,
			Status:  health.StatusUnhealthy,
			Message: "ping function is not configured",
			Error:   ErrPingFuncRequired.Error(),
		}
	}

	if err := c.ping(ctx); err != nil {
		return health.Result{
			Name:    c.name,
			Status:  health.StatusUnhealthy,
			Message: "ping failed",
			Error:   err.Error(),
		}
	}

	return health.Result{
		Name:   c.name,
		Status: health.StatusHealthy,
	}
}
