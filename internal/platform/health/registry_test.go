package health

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRegistryRegisterValidCheck(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Register(testCheck("postgres", CheckKindReadiness, StatusHealthy)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

func TestRegistryRejectsEmptyName(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(Check{
		Name:    " ",
		Kinds:   []CheckKind{CheckKindReadiness},
		Checker: CheckerFunc(func(context.Context) Result { return Result{Status: StatusHealthy} }),
	})
	if !errors.Is(err, ErrCheckNameRequired) {
		t.Fatalf("Register() error = %v, want %v", err, ErrCheckNameRequired)
	}
}

func TestRegistryRejectsNilChecker(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(Check{
		Name:  "postgres",
		Kinds: []CheckKind{CheckKindReadiness},
	})
	if !errors.Is(err, ErrCheckCheckerRequired) {
		t.Fatalf("Register() error = %v, want %v", err, ErrCheckCheckerRequired)
	}
}

func TestRegistryRejectsDuplicateName(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Register(testCheck("postgres", CheckKindReadiness, StatusHealthy)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err := registry.Register(testCheck("postgres", CheckKindDeep, StatusHealthy))
	if !errors.Is(err, ErrDuplicateCheck) {
		t.Fatalf("Register() error = %v, want %v", err, ErrDuplicateCheck)
	}
}

func TestRegistryDefaults(t *testing.T) {
	registry := NewRegistry().(*DefaultRegistry)

	if err := registry.Register(testCheck("postgres", CheckKindReadiness, StatusHealthy)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	registered := registry.checks["postgres"]
	if registered.Criticality != CriticalityRequired {
		t.Fatalf("Criticality = %s, want %s", registered.Criticality, CriticalityRequired)
	}
	if registered.Timeout != DefaultTimeout {
		t.Fatalf("Timeout = %s, want %s", registered.Timeout, DefaultTimeout)
	}
	if registered.Interval != DefaultInterval {
		t.Fatalf("Interval = %s, want %s", registered.Interval, DefaultInterval)
	}
}

func TestRegistrySnapshotRunsOnlyMatchingKind(t *testing.T) {
	registry := NewRegistry()
	readinessRan := false
	livenessRan := false

	if err := RegisterChecks(registry,
		Check{
			Name:  "readiness",
			Kinds: []CheckKind{CheckKindReadiness},
			Checker: CheckerFunc(func(context.Context) Result {
				readinessRan = true
				return Result{Status: StatusHealthy}
			}),
		},
		Check{
			Name:  "liveness",
			Kinds: []CheckKind{CheckKindLiveness},
			Checker: CheckerFunc(func(context.Context) Result {
				livenessRan = true
				return Result{Status: StatusHealthy}
			}),
		},
	); err != nil {
		t.Fatalf("RegisterChecks() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), CheckKindReadiness)
	if snapshot.Status != StatusHealthy {
		t.Fatalf("Snapshot().Status = %s, want %s", snapshot.Status, StatusHealthy)
	}
	if len(snapshot.Results) != 1 || snapshot.Results[0].Name != "readiness" {
		t.Fatalf("Results = %+v, want readiness only", snapshot.Results)
	}
	if !readinessRan {
		t.Fatal("readiness check did not run")
	}
	if livenessRan {
		t.Fatal("liveness check ran for readiness snapshot")
	}
}

func TestRegistrySnapshotHandlesTimeout(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(Check{
		Name:    "slow",
		Kinds:   []CheckKind{CheckKindReadiness},
		Timeout: 5 * time.Millisecond,
		Checker: CheckerFunc(func(context.Context) Result {
			time.Sleep(50 * time.Millisecond)
			return Result{Status: StatusHealthy}
		}),
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), CheckKindReadiness)
	if snapshot.Status != StatusUnhealthy {
		t.Fatalf("Snapshot().Status = %s, want %s", snapshot.Status, StatusUnhealthy)
	}
	if len(snapshot.Results) != 1 {
		t.Fatalf("Results len = %d, want 1", len(snapshot.Results))
	}
	result := snapshot.Results[0]
	if result.Status != StatusUnhealthy {
		t.Fatalf("Result status = %s, want %s", result.Status, StatusUnhealthy)
	}
	if !strings.Contains(result.Message, "timed out") {
		t.Fatalf("Message = %q, want timeout message", result.Message)
	}
}

func TestRegistrySnapshotHandlesPanic(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(Check{
		Name:  "panic",
		Kinds: []CheckKind{CheckKindReadiness},
		Checker: CheckerFunc(func(context.Context) Result {
			panic("boom")
		}),
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), CheckKindReadiness)
	if snapshot.Status != StatusUnhealthy {
		t.Fatalf("Snapshot().Status = %s, want %s", snapshot.Status, StatusUnhealthy)
	}
	if len(snapshot.Results) != 1 {
		t.Fatalf("Results len = %d, want 1", len(snapshot.Results))
	}
	if snapshot.Results[0].Status != StatusUnhealthy {
		t.Fatalf("Result status = %s, want %s", snapshot.Results[0].Status, StatusUnhealthy)
	}
	if !strings.Contains(snapshot.Results[0].Error, "boom") {
		t.Fatalf("Error = %q, want panic value", snapshot.Results[0].Error)
	}
}

func TestRegistrySnapshotSortsResultsByName(t *testing.T) {
	registry := NewRegistry()

	if err := RegisterChecks(registry,
		testCheck("b", CheckKindReadiness, StatusHealthy),
		testCheck("a", CheckKindReadiness, StatusHealthy),
	); err != nil {
		t.Fatalf("RegisterChecks() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), CheckKindReadiness)
	if len(snapshot.Results) != 2 {
		t.Fatalf("Results len = %d, want 2", len(snapshot.Results))
	}
	if snapshot.Results[0].Name != "a" || snapshot.Results[1].Name != "b" {
		t.Fatalf("Results order = %s, %s; want a, b", snapshot.Results[0].Name, snapshot.Results[1].Name)
	}
}

func TestRegistrySnapshotAggregatesStatus(t *testing.T) {
	registry := NewRegistry()

	if err := RegisterChecks(registry,
		testCheck("postgres", CheckKindReadiness, StatusHealthy),
		Check{
			Name:        "redis",
			Kinds:       []CheckKind{CheckKindReadiness},
			Criticality: CriticalityOptional,
			Checker:     CheckerFunc(func(context.Context) Result { return Result{Status: StatusUnhealthy} }),
		},
	); err != nil {
		t.Fatalf("RegisterChecks() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), CheckKindReadiness)
	if snapshot.Status != StatusDegraded {
		t.Fatalf("Snapshot().Status = %s, want %s", snapshot.Status, StatusDegraded)
	}
}

func testCheck(name string, kind CheckKind, status Status) Check {
	return Check{
		Name:  name,
		Kinds: []CheckKind{kind},
		Checker: CheckerFunc(func(context.Context) Result {
			return Result{Status: status}
		}),
	}
}
