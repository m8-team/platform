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

The root `ApplicationSet` emits the default platform application. Optional platform dependencies are represented as separate Argo CD `Application` manifests under `gitops/optional`.

## Sync Waves

| Wave | Application |
| ---: | --- |
| -10 | `flink-operator` shared Flink Kubernetes Operator |
| -5 | `flink-1c` independent Flink installation |
| 0 | `m8-platform` default platform services and UI |

## Environment Overlays

The first production overlay is:

- `gitops/environments/production/platform-installation.yaml`

Future overlays should be Kustomize or Helm values overlays under `gitops/environments/<environment>`, never one-off CLI mutations.

## Application Services

Deployable M8 application services live under:

- `gitops/components/platform/services`

The root `ApplicationSet` reconciles this tree through the `m8-platform` Argo CD Application. Each service gets its own directory with a local `kustomization.yaml`, Deployment, Service, ServiceAccount and optional policy manifests. The initial service scaffold is:

- `gitops/components/platform/services/resource-manager`

Infrastructure operators, data clusters, identity, authorization, observability and gateway resources must stay in their earlier sync-wave components, not in the application services tree.

## Flink GitOps

Flink has a dedicated multi-installation delivery model under `deploy/flink`:

- shared runtime image: `deploy/flink/shared/image`;
- single operator chart: `deploy/flink/shared/operator`;
- installation runtime and SQL: `deploy/flink/installations/<name>`;
- operator and per-installation Applications: `deploy/flink/argocd`.

The operator is installed once in `flink-operator` and watches installation
namespaces. Each installation has a dedicated namespace, `FlinkDeployment`, SQL
Gateway, Kafka configuration, object-storage prefix, Secrets, resources, SQL,
image digest pin, and Argo CD lifecycle. Production CI builds and publishes the
runtime image but never reconciles the cluster; Argo CD owns deployment,
self-healing, and pruning.

Bootstrap the project and Applications in dependency order as documented in
`deploy/flink/README.md`. Flink remains independent of the default root
`ApplicationSet`, so enabling one installation does not install or modify any
other installation.

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
