package app

import (
	"context"
	"errors"
	"github.com/m8-team/platform-sdk/health"
	"github.com/m8-team/platform-sdk/lifecycle"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"
)

func config() Config {
	return Config{StartupTimeout: time.Second, ShutdownTimeout: time.Second, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Health: health.New()}
}
func TestStartupRollback(t *testing.T) {
	var order []string
	sentinel := errors.New("bind failed")
	component := func(name string) lifecycle.Component {
		return lifecycle.Component{Name: name, Start: func(context.Context) error { order = append(order, "start "+name); return nil }, Stop: func(context.Context) error { order = append(order, "stop "+name); return nil }}
	}
	err := Run(context.Background(), config(), component("db"), component("kafka"), lifecycle.Component{Name: "http", Start: func(context.Context) error { return sentinel }})
	if !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if want := []string{"start db", "start kafka", "stop kafka", "stop db"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("%v", order)
	}
}

func TestCancelledCompositionClosesAcquiredResources(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	closed := false
	err := Run(ctx, config(), lifecycle.Component{Name: "allocated", Stop: func(context.Context) error { closed = true; return nil }}, lifecycle.Component{Name: "probe", Start: func(context.Context) error { t.Error("startup after cancellation"); return nil }})
	if !closed || !errors.Is(err, context.Canceled) {
		t.Fatalf("closed=%v error=%v", closed, err)
	}
}
func TestFatalTaskDrainsAndJoins(t *testing.T) {
	cfg := config()
	sentinel := errors.New("worker failed")
	stopped := false
	err := Run(context.Background(), cfg, lifecycle.Component{Name: "http", Run: func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }, Stop: func(context.Context) error {
		if cfg.Health.Ready(context.Background()) {
			t.Error("ready during shutdown")
		}
		stopped = true
		return nil
	}}, lifecycle.Component{Name: "worker", Run: func(context.Context) error { return sentinel }})
	if !errors.Is(err, sentinel) || !stopped {
		t.Fatalf("%v, stopped=%v", err, stopped)
	}
}
func TestShutdownBudget(t *testing.T) {
	cfg := config()
	cfg.ShutdownTimeout = 20 * time.Millisecond
	unblock := make(chan struct{})
	done := make(chan struct{})
	defer func() { close(unblock); <-done }()
	err := Run(context.Background(), cfg, lifecycle.Component{Name: "stuck", Run: func(context.Context) error { return errors.New("fatal") }, Stop: func(context.Context) error { defer close(done); <-unblock; return nil }})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}
