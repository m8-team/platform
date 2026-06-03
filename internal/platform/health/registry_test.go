package health

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRegistryRegisterValidCheck(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Register(testCheck("postgres", KindReadiness, StatusHealthy)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

func TestRegistryRejectsEmptyName(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(Check{
		Spec: CheckSpec{
			Name:  " ",
			Kinds: []Kind{KindReadiness},
		},
		Checker: CheckerFunc(func(context.Context) Result { return Result{Status: StatusHealthy} }),
	})
	if !errors.Is(err, ErrCheckNameRequired) {
		t.Fatalf("Register() error = %v, want %v", err, ErrCheckNameRequired)
	}
}

func TestRegistryRejectsNilChecker(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(Check{
		Spec: CheckSpec{
			Name:  "postgres",
			Kinds: []Kind{KindReadiness},
		},
	})
	if !errors.Is(err, ErrCheckCheckerRequired) {
		t.Fatalf("Register() error = %v, want %v", err, ErrCheckCheckerRequired)
	}

	var checker CheckerFunc
	err = registry.Register(Check{
		Spec: CheckSpec{
			Name:  "typed-nil",
			Kinds: []Kind{KindReadiness},
		},
		Checker: checker,
	})
	if !errors.Is(err, ErrCheckCheckerRequired) {
		t.Fatalf("Register() typed nil error = %v, want %v", err, ErrCheckCheckerRequired)
	}
}

func TestRegistryRejectsDuplicateName(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Register(testCheck("postgres", KindReadiness, StatusHealthy)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err := registry.Register(testCheck("postgres", KindReadiness, StatusHealthy))
	if !errors.Is(err, ErrDuplicateCheck) {
		t.Fatalf("Register() error = %v, want %v", err, ErrDuplicateCheck)
	}
}

func TestRegistryDefaults(t *testing.T) {
	registry := NewRegistry().(*registry)

	if err := registry.Register(testCheck("postgres", KindReadiness, StatusHealthy)); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	registered := registry.checks["postgres"]
	if registered.Spec.Criticality != CriticalityRequired {
		t.Fatalf("Criticality = %s, want %s", registered.Spec.Criticality, CriticalityRequired)
	}
	if registered.Spec.Timeout != defaultTimeout {
		t.Fatalf("Timeout = %s, want %s", registered.Spec.Timeout, defaultTimeout)
	}
	if registered.Spec.Interval != defaultInterval {
		t.Fatalf("Interval = %s, want %s", registered.Spec.Interval, defaultInterval)
	}
}

func TestRegistrySnapshotRunsOnlyMatchingKind(t *testing.T) {
	registry := NewRegistry()
	readinessRan := false
	livenessRan := false

	if err := RegisterChecks(registry,
		Check{
			Spec: CheckSpec{
				Name:  "readiness",
				Kinds: []Kind{KindReadiness},
			},
			Checker: CheckerFunc(func(context.Context) Result {
				readinessRan = true
				return Result{Status: StatusHealthy}
			}),
		},
		Check{
			Spec: CheckSpec{
				Name:  "liveness",
				Kinds: []Kind{KindLiveness},
			},
			Checker: CheckerFunc(func(context.Context) Result {
				livenessRan = true
				return Result{Status: StatusHealthy}
			}),
		},
	); err != nil {
		t.Fatalf("RegisterChecks() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), KindReadiness)
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
		Spec: CheckSpec{
			Name:    "slow",
			Kinds:   []Kind{KindReadiness},
			Timeout: 5 * time.Millisecond,
		},
		Checker: CheckerFunc(func(context.Context) Result {
			time.Sleep(50 * time.Millisecond)
			return Result{Status: StatusHealthy}
		}),
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), KindReadiness)
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
		Spec: CheckSpec{
			Name:  "panic",
			Kinds: []Kind{KindReadiness},
		},
		Checker: CheckerFunc(func(context.Context) Result {
			panic("boom")
		}),
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), KindReadiness)
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
		testCheck("b", KindReadiness, StatusHealthy),
		testCheck("a", KindReadiness, StatusHealthy),
	); err != nil {
		t.Fatalf("RegisterChecks() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), KindReadiness)
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
		testCheck("postgres", KindReadiness, StatusHealthy),
		Check{
			Spec: CheckSpec{
				Name:        "redis",
				Kinds:       []Kind{KindReadiness},
				Criticality: CriticalityOptional,
			},
			Checker: CheckerFunc(func(context.Context) Result { return Result{Status: StatusUnhealthy} }),
		},
	); err != nil {
		t.Fatalf("RegisterChecks() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), KindReadiness)
	if snapshot.Status != StatusDegraded {
		t.Fatalf("Snapshot().Status = %s, want %s", snapshot.Status, StatusDegraded)
	}
}

func TestRegistrySnapshotLatencyUsesMilliseconds(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Register(Check{
		Spec: CheckSpec{
			Name:  "slow",
			Kinds: []Kind{KindReadiness},
		},
		Checker: CheckerFunc(func(context.Context) Result {
			time.Sleep(15 * time.Millisecond)
			return Result{Status: StatusHealthy}
		}),
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	snapshot := registry.Snapshot(context.Background(), KindReadiness)
	if len(snapshot.Results) != 1 {
		t.Fatalf("Results len = %d, want 1", len(snapshot.Results))
	}

	latency := int64(snapshot.Results[0].Latency)
	if latency < 10 || latency > 500 {
		t.Fatalf("Latency = %d, want milliseconds", latency)
	}
}

func TestRegistryConcurrentRegisterAndSnapshot(t *testing.T) {
	registry := NewRegistry()

	var wait sync.WaitGroup
	for i := 0; i < 50; i++ {
		wait.Add(1)
		go func(i int) {
			defer wait.Done()

			_ = registry.Register(testCheck(fmt.Sprintf("check-%03d", i), KindReadiness, StatusHealthy))
		}(i)
	}

	for i := 0; i < 50; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()

			_ = registry.Snapshot(context.Background(), KindReadiness)
		}()
	}

	wait.Wait()

	snapshot := registry.Snapshot(context.Background(), KindReadiness)
	if snapshot.Status != StatusHealthy {
		t.Fatalf("Snapshot().Status = %s, want %s", snapshot.Status, StatusHealthy)
	}
	for i := 1; i < len(snapshot.Results); i++ {
		if snapshot.Results[i-1].Name > snapshot.Results[i].Name {
			t.Fatalf("results are not sorted: %s before %s", snapshot.Results[i-1].Name, snapshot.Results[i].Name)
		}
	}
}

func testCheck(name string, kind Kind, status Status) Check {
	return Check{
		Spec: CheckSpec{
			Name:  name,
			Kinds: []Kind{kind},
		},
		Checker: CheckerFunc(func(context.Context) Result {
			return Result{Status: status}
		}),
	}
}
