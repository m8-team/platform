package checks

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/platform/health"
)

var errPingFuncRequired = errors.New("ping function is required")

type PingFunc func(ctx context.Context) error

func NewPingCheck(name string, ping PingFunc) health.Check {
	return func(ctx context.Context) health.Result {
		if ping == nil {
			return health.Result{
				Name:    name,
				Status:  health.StatusUnhealthy,
				Message: "ping function is not configured",
				Error:   errPingFuncRequired.Error(),
			}
		}

		if err := ping(ctx); err != nil {
			return health.Result{
				Name:    name,
				Status:  health.StatusUnhealthy,
				Message: "ping failed",
				Error:   err.Error(),
			}
		}

		return health.Result{
			Name:   name,
			Status: health.StatusHealthy,
		}
	}
}
