package retry

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

func TestExplicitClassification(t *testing.T) {
	errFailure := errors.New("failure")
	for _, classify := range []bool{false, true} {
		t.Run(map[bool]string{false: "no implicit retry", true: "bounded retry"}[classify], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls := 0
				policy := Policy{MaxAttempts: 3, Initial: time.Second, MaxDelay: 2 * time.Second}
				if classify {
					policy.If = func(err error) bool { return errors.Is(err, errFailure) }
				}
				err := Do(context.Background(), policy, func(context.Context) error { calls++; return errFailure })
				want := 1
				if classify {
					want = 3
				}
				if calls != want || !errors.Is(err, errFailure) {
					t.Fatalf("calls=%d error=%v", calls, err)
				}
			})
		})
	}
}
func TestDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := Do(ctx, Policy{MaxAttempts: 10, Initial: time.Second, MaxDelay: time.Second, If: func(error) bool { return true }}, func(context.Context) error { return errors.New("temporary") })
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	})
}
