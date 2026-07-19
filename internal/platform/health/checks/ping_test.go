package checks

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/platform/health"
)

func TestPingCheckSuccess(t *testing.T) {
	check := NewPingCheck("postgres", func(context.Context) error { return nil })

	result := check(context.Background())
	if result.Status != health.StatusHealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusHealthy)
	}
}

func TestPingCheckError(t *testing.T) {
	check := NewPingCheck("postgres", func(context.Context) error { return errors.New("down") })

	result := check(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
	if result.Error == "" {
		t.Fatal("Error is empty, want ping error")
	}
}

func TestPingCheckNilPing(t *testing.T) {
	check := NewPingCheck("postgres", nil)

	result := check(context.Background())
	if result.Status != health.StatusUnhealthy {
		t.Fatalf("Status = %s, want %s", result.Status, health.StatusUnhealthy)
	}
	if result.Error != errPingFuncRequired.Error() {
		t.Fatalf("Error = %q, want %q", result.Error, errPingFuncRequired.Error())
	}
}
