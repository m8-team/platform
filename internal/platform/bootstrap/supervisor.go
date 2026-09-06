// Package bootstrap owns the lifetime of process-level background tasks.
package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
)

// Supervisor joins every registered task before the process finishes stopping.
// Register its lifecycle hook before server hooks so servers drain before Wait.
type Supervisor struct {
	group    *errgroup.Group
	ctx      context.Context
	shutdown fx.Shutdowner
}

func NewSupervisor(lifecycle fx.Lifecycle, shutdown fx.Shutdowner) *Supervisor {
	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)
	s := &Supervisor{group: group, ctx: ctx, shutdown: shutdown}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		cancel()
		return group.Wait()
	}})
	return s
}

// Go is called during startup. Task errors stop the whole host with a failure
// exit code; a service cannot silently remain healthy after losing a listener.
func (s *Supervisor) Go(name string, run func(context.Context) error) {
	s.group.Go(func() error {
		if err := run(s.ctx); err != nil {
			if s.ctx.Err() != nil && errors.Is(err, context.Canceled) {
				return nil
			}
			if shutdownErr := s.shutdown.Shutdown(fx.ExitCode(1)); shutdownErr != nil {
				return fmt.Errorf("%s: %w; request shutdown: %w", name, err, shutdownErr)
			}
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	})
}
