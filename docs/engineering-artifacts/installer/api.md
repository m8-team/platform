---
title: "M8 Installer API Reference"
---

# API Reference

## Go Types

Canonical API types live in:

- `api/installer/v1alpha1/types.go`
- `api/installer/v1alpha1/validation.go`

Implemented resources:

- `PlatformInstallation`
- `PlatformRelease`
- `InstallationOperation`
- `Backup`
- `RestorePlan`

The package intentionally contains only API models, defaulting and validation. Kubernetes clients, Helm clients and execution logic live under `internal/installer`.

## CRD Schemas

Bootstrap CRDs live in:

- `charts/m8-installer-crds/Chart.yaml`
- `charts/m8-installer-crds/templates/crds.yaml`

The first schema covers all requested top-level sections:

```yaml
spec:
  cluster: {}
  network: {}
  gateway: {}
  certificates: {}
  trust: {}
  gitOps: {}
  secrets: {}
  databases: {}
  messaging: {}
  workflows: {}
  identity: {}
  authorization: {}
  observability: {}
  security: {}
  backup: {}
  modules: {}
  upgrade: {}
  airGap: {}
```

## Defaulting

Defaults are conservative and profile-aware:

| Profile | Defaults |
| --- | --- |
| `demo` | Single-node assumptions, Kubernetes Secrets allowed, in-cluster components. |
| `development` | In-cluster components, local-friendly defaults. |
| `staging` | External secrets, backup, reduced production-like stack. |
| `production` | At least three nodes, backup, supply-chain checks, security policies, SPIRE, HA data. |
| `air-gapped` | Production-like plus `airGap.enabled=true` and private registry requirement. |

## Validation

Validation rejects:

- missing `spec.platformVersion`;
- unsupported profiles;
- production without external secrets, backup or signature enforcement;
- air-gapped profile without private registry;
- CNI replacement without explicit permission;
- invalid data component modes;
- module dependency violations.

Module dependencies:

| Module | Requires |
| --- | --- |
| `authentication` | `identity`, Keycloak |
| `access` | SpiceDB |
| `provisioning` | Temporal |
| `gateway` | Envoy Gateway |
| `audit` | YDB |
| `operations` | Temporal |

## Status Model

`PlatformInstallation.status` includes:

- `observedGeneration`
- `phase`
- `platformVersion`
- `conditions`
- `components`
- `endpoints`
- `lastOperation`

Phases:

`Pending`, `Validating`, `Planning`, `Bootstrapping`, `Installing`, `Migrating`, `Verifying`, `Ready`, `Degraded`, `Failed`, `Upgrading`, `RollingBack`, `BackingUp`, `Restoring`, `Uninstalling`.

## Update And Deletion Behavior

- Safe field changes trigger plan recalculation and Argo CD reconciliation.
- Version changes route through `upgrade plan` and `upgrade execute`.
- Stateful mode changes require backup checks and may require maintenance windows.
- Deletion defaults to stateless uninstall only.
- Data deletion requires `--delete-data`, installation-name confirmation and a destructive action token.

