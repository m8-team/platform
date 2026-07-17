---
title: "M8 Installer Test Strategy"
---

# Test Strategy

## Unit Tests

Implemented initial tests:

- `api/installer/v1alpha1`: profile defaulting, validation, release digest/floating-version rejection;
- `internal/installer/graph`: deterministic order and cycle detection;
- `internal/installer/planner`: sync wave order;
- `internal/installer/preflight`: fake cluster checks;
- `internal/installer/diagnostics`: secret redaction.

Required next unit coverage:

- config merge;
- version compatibility;
- operation error classification;
- plan signature verification;
- module dependency graph matrix;
- rollback boundary detection.

## Integration Tests

Use:

- fake Kubernetes clients;
- envtest for CRD installation and Server-Side Apply;
- Helm SDK action configuration with test storage driver;
- local OCI registry;
- Argo CD resource reconciliation tests without requiring a running Argo CD controller.

## E2E Tests

Temporary clusters:

- kind;
- k3d;
- Talos;
- managed Kubernetes smoke environments.

Scenarios:

- fresh install;
- repeated install;
- partial failure and resume;
- interrupted CLI;
- upgrade;
- failed upgrade and rollback;
- backup and restore;
- air-gapped import;
- existing CNI;
- existing cert-manager;
- external databases;
- multi-zone production profile.

## Chaos Tests

- delete operator pods during install;
- registry unavailable;
- Git unavailable;
- Vault unavailable;
- PostgreSQL unavailable;
- Kubernetes API temporary unavailability;
- network loss during upgrade.

## Security Tests

- secrets in logs and diagnostic bundles;
- unsigned release catalog;
- image without digest;
- malicious Helm chart;
- registry MITM;
- privilege escalation in generated manifests;
- webhook conflicts;
- expired policy exceptions.

## Current Verification Command

```bash
go test -mod=mod ./api/installer/v1alpha1 ./internal/installer/... ./cmd/m8ctl
```

Full `go test ./...` currently requires fixing the repository-level `replace github.com/m8-team/go-genproto => ./api/generate/go` target by adding a module file or changing generated import ownership.

