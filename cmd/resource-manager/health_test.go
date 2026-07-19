package main

import (
	"context"
	"testing"

	"github.com/m8-team/platform/internal/platform/health"
)

func TestRegisterResourceManagerHealthChecks(t *testing.T) {
	registry := &capturingHealthRegistry{}

	if err := registerResourceManagerHealthChecks(registry); err != nil {
		t.Fatalf("registerResourceManagerHealthChecks() error = %v", err)
	}

	if len(registry.registrations) != 1 {
		t.Fatalf("registrations len = %d, want 1", len(registry.registrations))
	}

	registration := registry.registrations[0]
	spec := registration.Spec
	if spec.Name != yaRuHealthCheckName {
		t.Fatalf("Name = %q, want %q", spec.Name, yaRuHealthCheckName)
	}
	if spec.Target.Kind != health.TargetKindDependency {
		t.Fatalf("Target.Kind = %s, want %s", spec.Target.Kind, health.TargetKindDependency)
	}
	if spec.Target.Name != "ya.ru" {
		t.Fatalf("Target.Name = %q, want ya.ru", spec.Target.Name)
	}
	if spec.Target.Module != "resource-manager" {
		t.Fatalf("Target.Module = %q, want resource-manager", spec.Target.Module)
	}
	if spec.Criticality != health.CriticalityOptional {
		t.Fatalf("Criticality = %s, want %s", spec.Criticality, health.CriticalityOptional)
	}
	if len(spec.Kinds) != 1 || spec.Kinds[0] != health.KindReadiness {
		t.Fatalf("Kinds = %+v, want [%s]", spec.Kinds, health.KindReadiness)
	}
	if registration.Check == nil {
		t.Fatal("Check is nil")
	}
}

type capturingHealthRegistry struct {
	registrations []health.Config
}

func (r *capturingHealthRegistry) Register(registration health.Config) error {
	r.registrations = append(r.registrations, registration)
	return nil
}

func (r *capturingHealthRegistry) Snapshot(context.Context, health.Kind) health.Snapshot {
	return health.Snapshot{}
}
