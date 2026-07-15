---
title: "M8 Installer 1.0 Implementation Plan"
---

# Implementation Plan

## Epic 1. Installer API And Contracts

Features:

- `PlatformInstallation` CRD.
- `PlatformRelease` CRD.
- `InstallationOperation` CRD.
- `Backup` and `RestorePlan` CRDs.

Tasks:

- Generate CRD schemas from Go types.
- Add CEL validation for production and air-gapped profile rules.
- Add status condition conventions.
- Add API examples for all profiles.

Acceptance:

- CRDs install on supported Kubernetes versions.
- Invalid profile/module combinations are rejected before mutation.

Risks:

- Manual schemas can drift; code generation is required before beta.

## Epic 2. Preflight And Planning

Features:

- Preflight framework.
- Kubernetes discovery checks.
- Release catalog verification.
- Deterministic installation plan.

Tasks:

- Add registry, Git, Vault and object storage checks.
- Add external PostgreSQL, YDB, Kafka and Redis readiness checks.
- Add VolumeSnapshotClass, LoadBalancer, DNS, NTP and webhook conflict checks.
- Sign saved plans.

Acceptance:

- `m8ctl preflight` emits table/json/yaml.
- `m8ctl plan` produces stable digests and sync waves.

Risks:

- Some checks require cloud-specific adapters.

## Epic 3. Bootstrap Executor

Features:

- Server-Side Apply engine.
- Helm SDK engine.
- Operation checkpoints.
- Argo CD handoff.

Tasks:

- Implement apply/verify/rollback step lifecycle.
- Store `InstallationOperation` after each safe checkpoint.
- Install CRDs and Argo CD without external `kubectl` or `helm`.
- Watch readiness using Kubernetes Watch API.

Acceptance:

- Clean cluster can bootstrap.
- Re-running bootstrap is idempotent.
- Interrupting bootstrap resumes from last checkpoint.

Risks:

- CNI installation has cluster-specific failure modes.

## Epic 4. GitOps Platform Reconciliation

Features:

- Root AppProject/ApplicationSet.
- Component applications.
- Production overlays.
- Custom Argo health checks.

Tasks:

- Add component manifests for operators and M8 modules.
- Add sync waves and health scripts.
- Add private repo and OCI registry references via External Secrets.
- Add Keycloak SSO for Argo CD.

Acceptance:

- Argo CD becomes source of truth after bootstrap.
- CLI does not continuously reconcile application resources.

Risks:

- Argo CD health for some CRDs must be custom.

## Epic 5. Upgrade And Rollback

Features:

- Upgrade planner.
- Migration classification.
- Rollback boundary enforcement.

Tasks:

- Resolve current version from `PlatformInstallation.status`.
- Validate upgrade path from `PlatformRelease`.
- Require verified backup before data migrations.
- Stop automatic rollback after irreversible boundary.

Acceptance:

- Non-rollbackable migrations are visible before execute.
- Rollback never crosses irreversible data migration.

Risks:

- Stateful components may require per-version recovery playbooks.

## Epic 6. Backup, Restore And DR

Features:

- Consistent backup orchestration.
- Native database backup integration.
- Restore planning and execution.
- Post-restore functional checks.

Tasks:

- Integrate Velero, PostgreSQL native backup, YDB backup, Kafka config export, Keycloak realm export, SpiceDB schema export.
- Add backup verification.
- Add restore order and maintenance gates.

Acceptance:

- Restore plan shows RPO/RTO, risks and replaced resources.
- Restore runs smoke tests.

Risks:

- Cross-store consistency requires clear quiesce/snapshot policy.

## Epic 7. Air-Gapped Bundle

Features:

- Bundle export.
- Bundle verify.
- Bundle import.

Tasks:

- Resolve all release artifacts.
- Copy images/charts/CRDs/manifests/SBOM/signatures/docs.
- Write manifest, checksums and signature.
- Import to private registry and rewrite references through catalog overlay.

Acceptance:

- Air-gapped install does not require internet access.
- Missing dependency is detected before import.

Risks:

- License and vulnerability metadata freshness must be handled offline.

## Epic 8. Doctor And Diagnostics

Features:

- Functional doctor checks.
- Sanitized diagnostics bundle.
- Operation telemetry.

Tasks:

- Implement TLS, SPIFFE, OIDC, YDB, PostgreSQL, Redis, Kafka, Temporal, SpiceDB, Gateway, OTel, Prometheus and backup checks.
- Add automatic redaction.
- Add trace and metric emission.

Acceptance:

- `doctor` validates real user-impacting flows.
- Diagnostic archive contains no secrets.

Risks:

- Smoke credentials must be created and rotated safely.

