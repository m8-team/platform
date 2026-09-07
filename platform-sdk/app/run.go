package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/m8-team/platform-sdk/health"
	"github.com/m8-team/platform-sdk/lifecycle"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"runtime/pprof"
	"time"
)

type Config struct {
	StartupTimeout, ShutdownTimeout time.Duration
	Logger                          *slog.Logger
	Health                          *health.Registry
}

// Run owns tasks and cleanup. The caller owns signals, dependency construction
// and exit codes. Every dependency must appear before its consumers.
func Run(parent context.Context, cfg Config, components ...lifecycle.Component) error {
	if cfg.Logger == nil || cfg.Health == nil || cfg.StartupTimeout <= 0 || cfg.ShutdownTimeout <= 0 {
		return errors.New("logger, health and positive timeouts required")
	}
	names := map[string]bool{}
	for _, c := range components {
		if c.Name == "" || names[c.Name] {
			return errors.New("unique component names required")
		}
		names[c.Name] = true
	}
	startup, cancelStartup := context.WithTimeout(parent, cfg.StartupTimeout)
	defer cancelStartup()
	started := 0
	var result error
	for _, c := range components {
		// A component without Start transfers an already constructed resource.
		// Register its cleanup even if the caller cancelled during composition.
		if c.Start == nil {
			started++
			continue
		}
		if err := startup.Err(); err != nil {
			result = err
			break
		}
		if err := c.Start(startup); err != nil {
			result = fmt.Errorf("start %s: %w", c.Name, err)
			break
		}
		started++
	}
	if result == nil && startup.Err() != nil {
		result = startup.Err()
	}
	cancelStartup()
	group, groupCtx := errgroup.WithContext(context.WithoutCancel(parent))
	cancels := make([]context.CancelFunc, started)
	done := make([]chan struct{}, started)
	if result == nil {
		for i, c := range components[:started] {
			ctx, cancel := context.WithCancel(context.WithoutCancel(parent))
			cancels[i] = cancel
			done[i] = make(chan struct{})
			if c.Run == nil {
				close(done[i])
				continue
			}
			group.Go(func() error {
				defer close(done[i])
				var err error
				pprof.Do(ctx, pprof.Labels("component", c.Name), func(ctx context.Context) { err = c.Run(ctx) })
				if ctx.Err() != nil && errors.Is(err, context.Canceled) {
					return nil
				}
				if err == nil && ctx.Err() == nil {
					err = errors.New("runtime task stopped unexpectedly")
				}
				if err != nil {
					return fmt.Errorf("run %s: %w", c.Name, err)
				}
				return nil
			})
		}
		cfg.Health.Started()
		select {
		case <-parent.Done():
		case <-groupCtx.Done():
		}
	}
	cfg.Health.Drain()
	shutdown, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	for i := started - 1; i >= 0; i-- {
		c := components[i]
		if cancels[i] != nil {
			cancels[i]()
		}
		if c.Stop != nil {
			if err := stopComponent(shutdown, c); err != nil {
				result = errors.Join(result, fmt.Errorf("stop %s: %w", c.Name, err))
			}
		}
		if shutdown.Err() != nil && c.Force != nil {
			result = errors.Join(result, c.Force())
		}
		if done[i] != nil {
			select {
			case <-done[i]:
			case <-shutdown.Done():
				result = errors.Join(result, fmt.Errorf("join %s: %w", c.Name, shutdown.Err()))
			}
		}
	}
	if shutdown.Err() != nil {
		var dump bytes.Buffer
		_ = pprof.Lookup("goroutine").WriteTo(&dump, 2)
		cfg.Logger.Error("shutdown budget exceeded", "goroutines", dump.String())
		return errors.Join(result, shutdown.Err())
	}
	return errors.Join(result, group.Wait())
}

// A hook has an owner and a bounded join. Go cannot kill an uncooperative hook:
// after timeout the process owner must exit; Run must not be used to restart an
// application in the same process after a shutdown deadline has expired.
func stopComponent(ctx context.Context, c lifecycle.Component) error {
	done := make(chan error, 1)
	go func() {
		pprof.Do(ctx, pprof.Labels("component", c.Name, "phase", "stop"), func(ctx context.Context) { done <- c.Stop(ctx) })
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
