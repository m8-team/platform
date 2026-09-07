package retry

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

type Policy struct {
	MaxAttempts       int
	Initial, MaxDelay time.Duration
	If                func(error) bool
}

// Do has no default classifier or hidden attempts. MaxAttempts includes first call.
func Do(ctx context.Context, p Policy, fn func(context.Context) error) error {
	if p.MaxAttempts < 1 || p.Initial < 0 || p.MaxDelay < p.Initial || fn == nil {
		return errors.New("invalid retry policy")
	}
	delay := p.Initial
	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := fn(ctx)
		if err == nil || attempt >= p.MaxAttempts || p.If == nil || !p.If(err) {
			return err
		}
		pause := delay / 2
		if pause > 0 {
			pause += time.Duration(rand.Int64N(int64(pause)))
		}
		timer := time.NewTimer(pause)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if delay > p.MaxDelay/2 {
			delay = p.MaxDelay
		} else {
			delay *= 2
		}
	}
}
