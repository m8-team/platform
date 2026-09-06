package bootstrap

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/fx"
)

func TestFailureStopsHostAndJoinsSibling(t *testing.T) {
	failure := errors.New("listener failed")
	joined := make(chan struct{})
	app := fx.New(fx.Provide(NewSupervisor), fx.Invoke(func(lc fx.Lifecycle, s *Supervisor) {
		lc.Append(fx.Hook{OnStart: func(context.Context) error {
			s.Go("sibling", func(ctx context.Context) error { <-ctx.Done(); close(joined); return nil })
			s.Go("listener", func(context.Context) error { return failure })
			return nil
		}})
	}), fx.NopLogger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case signal := <-app.Wait():
		if signal.ExitCode != 1 {
			t.Fatalf("exit code = %d", signal.ExitCode)
		}
	case <-ctx.Done():
		t.Fatal("failure did not stop host")
	}
	if err := app.Stop(ctx); !errors.Is(err, failure) {
		t.Fatalf("Stop = %v", err)
	}
	select {
	case <-joined:
	default:
		t.Fatal("sibling was not joined")
	}
}

func TestNormalCancellationIsNotFailure(t *testing.T) {
	app := fx.New(fx.Provide(NewSupervisor), fx.Invoke(func(lc fx.Lifecycle, s *Supervisor) {
		lc.Append(fx.Hook{OnStart: func(context.Context) error {
			s.Go("worker", func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })
			return nil
		}})
	}), fx.NopLogger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := app.Stop(ctx); err != nil {
		t.Fatalf("normal stop = %v", err)
	}
}
