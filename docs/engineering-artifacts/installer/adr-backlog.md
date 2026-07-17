---
title: "M8 Installer ADR Backlog"
---

# ADR Backlog

| ADR | Status | Decision |
| --- | --- | --- |
| ADR-I001 | Proposed | Use two-phase installation: direct bootstrap, then GitOps reconciliation. |
| ADR-I002 | Proposed | Use `PlatformInstallation` as the installation API and desired-state contract. |
| ADR-I003 | Proposed | Use signed `PlatformRelease` catalog with digest-pinned artifacts. |
| ADR-I004 | Proposed | Use Server-Side Apply for direct Kubernetes mutations. |
| ADR-I005 | Proposed | Use Helm SDK and OCI APIs directly, never shell out to `helm` or `kubectl`. |
| ADR-I006 | Proposed | Store resumable operation state in `InstallationOperation`. |
| ADR-I007 | Proposed | Use DAG-based sync waves with observed readiness gates. |
| ADR-I008 | Proposed | Keep CLI out of steady-state reconciliation after Argo CD handoff. |
| ADR-I009 | Proposed | Treat secrets as external references and redact diagnostics by default. |
| ADR-I010 | Proposed | Block automatic rollback after irreversible data migration boundaries. |
| ADR-I011 | Proposed | Make air-gapped bundle completeness a release gate. |
| ADR-I012 | Proposed | Keep multi-cluster roles in the API from 1.0 while deferring full runtime. |

Each ADR must follow the repository ADR format: status, context, decision and consequences.

