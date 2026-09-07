package kafka

import (
	"context"
	"errors"
	"github.com/m8-team/platform-sdk/logging"
	"github.com/m8-team/platform-sdk/retry"
	"github.com/m8-team/platform-sdk/telemetry"
	"github.com/twmb/franz-go/pkg/kgo"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestFailedDispositionNeverCommits(t *testing.T) {
	p, err := telemetry.New(context.Background(), logging.Info{Name: "test"}, telemetry.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Shutdown(context.Background())
	for _, mode := range []string{"panic", "timeout", "dlq-failure"} {
		t.Run(mode, func(t *testing.T) {
			// Nil Client is intentional: any attempt to commit on these paths panics.
			c := Consumer{Telemetry: p, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Timeout: time.Millisecond, Retry: retry.Policy{MaxAttempts: 1}}
			switch mode {
			case "panic":
				c.Handler = func(context.Context, Message) error { panic("secret") }
			case "timeout":
				c.Handler = func(ctx context.Context, _ Message) error { <-ctx.Done(); return nil }
			case "dlq-failure":
				c.Handler = func(context.Context, Message) error { return errors.New("terminal") }
				c.DeadLetter = func(context.Context, Message, error) error { return errors.New("broker unavailable") }
			}
			if err := c.process(context.Background(), &kgo.Record{Topic: "test"}); err == nil {
				t.Fatal("failure was swallowed")
			}
		})
	}
}
