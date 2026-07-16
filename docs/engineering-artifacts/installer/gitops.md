---
title: "M8 Installer GitOps Design"
---

# GitOps Design

## Decision

After bootstrap, Argo CD owns platform reconciliation. `m8ctl` may observe and diagnose the platform, but it must not continuously reconcile application resources.

## Bootstrap Layer

`m8ctl bootstrap` installs only:

1. Cilium when no compatible CNI exists and replacement is explicitly allowed.
2. Gateway API CRDs.
3. cert-manager.
4. trust-manager.
5. Argo CD.
6. External Secrets Operator.
7. Initial namespaces.
8. Root `AppProject`.
9. Root `ApplicationSet`.
10. M8 installer metadata CRDs and `PlatformInstallation`.

## Root Resources

Files:

- `gitops/root/appproject.yaml`
- `gitops/root/applicationset.yaml`

The root `ApplicationSet` emits component applications for data operators, data clusters, identity, observability, gateway, M8 services, routes, bootstrap data and smoke tests.

## Sync Waves

| Wave | Application |
| ---: | --- |
| -60 | `m8-data-operators` |
| -50 | `m8-data-clusters` |
| -40 | `m8-identity-authorization` |
| -30 | `m8-observability` |
| -20 | `m8-envoy-gateway` |
| -10 | `m8-shared-services` |
| 0 | `m8-applications` |
| 10 | `m8-routes-policies` |
| 20 | `m8-bootstrap-data` |
| 30 | `m8-smoke-tests` |

## Environment Overlays

The first production overlay is:

- `gitops/environments/production/platform-installation.yaml`

Future overlays should be Kustomize or Helm values overlays under `gitops/environments/<environment>`, never one-off CLI mutations.

## Application Services

Deployable M8 application services live under:

- `gitops/components/m8-applications/services`

The root `ApplicationSet` reconciles this tree through the `m8-m8-applications` Argo CD Application. Each service gets its own directory with a local `kustomization.yaml`, Deployment, Service, ServiceAccount and optional policy manifests. The initial service scaffold is:

- `gitops/components/m8-applications/services/resource-manager`

Infrastructure operators, data clusters, identity, authorization, observability and gateway resources must stay in their earlier sync-wave components, not in the application services tree.

## Health And Readiness

Argo CD health checks must be added for:

- `PlatformInstallation`;
- data clusters;
- Keycloak realm import;
- SpiceDB schema migration;
- Temporal namespace bootstrap;
- Gateway API route status;
- M8 module readiness.

## Security

- Private repositories use External Secrets references.
- Argo CD SSO is via Keycloak.
- Project RBAC grants the installer only bootstrap/sync permissions needed for handoff.
- OCI Helm chart sources must use digest-pinned release catalog entries.
