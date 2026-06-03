package checks

import (
	"context"
	"errors"
	"testing"

	"github.com/m8platform/platform/internal/platform/health"
)

func TestPingCheckerSuccess(t *testing.T) {
	checker := NewPingChecker("postgres", func(context.Context) error { return nil })

	result := checker.Check(context.Background())
	if result.Status != health.StatusHealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusHealthy)
	}
}

func TestPingCheckerError(t *testing.T) {
	checker := NewPingChecker("postgres", func(context.Context) error { return errors.New("down") })

	result := checker.Check(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
	if result.Error == "" {
		t.Fatal("Error is empty, want ping error")
	}
}

func TestPingCheckerNilPing(t *testing.T) {
	checker := NewPingChecker("postgres", nil)

	result := checker.Check(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
	if !errors.Is(ErrPingFuncRequired, ErrPingFuncRequired) || result.Error != ErrPingFuncRequired.Error() {
		t.Fatalf("Error = %q, want %q", result.Error, ErrPingFuncRequired.Error())
	}
}
