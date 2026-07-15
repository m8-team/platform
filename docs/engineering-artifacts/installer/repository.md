---
title: "M8 Installer Repository Structure"
---

# Repository Structure

```text
cmd/m8ctl/                         CLI entrypoint.
api/installer/v1alpha1/            Kubernetes API types, defaulting and validation.
internal/installer/catalog/         PlatformRelease catalog loading and verification.
internal/installer/cli/             Command parsing, exit codes and command orchestration.
internal/installer/config/          Installation file loading and canonical digests.
internal/installer/diagnostics/     Redaction and future bundle creation.
internal/installer/gitops/          Argo CD port interfaces.
internal/installer/graph/           Deterministic dependency graph.
internal/installer/helm/            Helm SDK port interfaces.
internal/installer/kubernetes/      client-go adapter.
internal/installer/operations/      Operation model helpers and typed errors.
internal/installer/output/          Table, JSON and YAML output.
internal/installer/planner/         InstallationPlan model and generator.
internal/installer/preflight/       Preflight framework and checks.
internal/installer/registry/        OCI registry port interfaces.
internal/installer/security/        Signature verification boundary.
charts/m8-installer-crds/           Bootstrap CRD chart.
gitops/root/                        Root AppProject and ApplicationSet.
gitops/environments/production/     First production PlatformInstallation overlay.
catalog/releases/                   Signed PlatformRelease catalog entries.
docs/engineering-artifacts/installer/ Architecture, API, CLI, testing and plan docs.
```

The structure follows the repository AGENTS rules by keeping installer as a bounded context under `internal/installer` and avoiding catch-all `common` or `utils` packages.

