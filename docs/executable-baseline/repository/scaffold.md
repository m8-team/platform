---
title: Repository Scaffold
---

# Repository Scaffold

[Executable Baseline](../index.md) | [Repository](index.md)

{% note info %}

The repository scaffold is executable evidence. It is intentionally small and demonstrates service boundaries, a platform operation model, idempotency support and the AUTH-FR-017 pilot slice.

{% endnote %}

## Raw files

| Area | Path |
| --- | --- |
| Workspace | `go.work` |
| Module | `go.mod` |
| Platform idempotency | `internal/platform/idempotency/store.go` |
| Platform operations | `internal/platform/operation/operation.go` |
| Services | `services/` |

## Service Entrypoints

| Service | Entrypoint |
| --- | --- |
| m8-access | `services/m8-access/cmd/m8-access/main.go` |
| m8-audit | `services/m8-audit/cmd/m8-audit/main.go` |
| m8-authentication | `services/m8-authentication/cmd/m8-authentication/main.go` |
| m8-identity | `services/m8-identity/cmd/m8-identity/main.go` |
| m8-provisioning | `services/m8-provisioning/cmd/m8-provisioning/main.go` |
| m8-resource-manager | `services/m8-resource-manager/cmd/m8-resource-manager/main.go` |
| m8-risk-decision | `services/m8-risk-decision/cmd/m8-risk-decision/main.go` |
