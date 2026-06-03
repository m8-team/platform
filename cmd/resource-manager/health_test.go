package main

import (
	"context"
	"testing"

	"github.com/m8platform/platform/internal/platform/health"
)

func TestRegisterResourceManagerHealthChecks(t *testing.T) {
	registry := &capturingHealthRegistry{}

	if err := registerResourceManagerHealthChecks(registry); err != nil {
		t.Fatalf("registerResourceManagerHealthChecks() error = %v", err)
	}

	if len(registry.checks) != 1 {
		t.Fatalf("checks len = %d, want 1", len(registry.checks))
	}

	check := registry.checks[0]
	if check.Name != yaRuHealthCheckName {
		t.Fatalf("Name = %q, want %q", check.Name, yaRuHealthCheckName)
	}
	if check.Target.Kind != health.TargetDependency {
		t.Fatalf("Target.Kind = %s, want %s", check.Target.Kind, health.TargetDependency)
	}
	if check.Target.Name != "ya.ru" {
		t.Fatalf("Target.Name = %q, want ya.ru", check.Target.Name)
	}
	if check.Target.Module != "resource-manager" {
		t.Fatalf("Target.Module = %q, want resource-manager", check.Target.Module)
	}
	if check.Criticality != health.CriticalityOptional {
		t.Fatalf("Criticality = %s, want %s", check.Criticality, health.CriticalityOptional)
	}
	if len(check.Kinds) != 1 || check.Kinds[0] != health.CheckKindDeep {
		t.Fatalf("Kinds = %+v, want [%s]", check.Kinds, health.CheckKindDeep)
	}
	if check.Checker == nil {
		t.Fatal("Checker is nil")
	}
}

type capturingHealthRegistry struct {
	checks []health.Check
}

func (r *capturingHealthRegistry) Register(check health.Check) error {
	r.checks = append(r.checks, check)
	return nil
}

func (r *capturingHealthRegistry) Snapshot(context.Context, health.CheckKind) health.Snapshot {
	return health.Snapshot{}
}
