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
| -46 | `m8-flink-operator` optional Flink Kubernetes Operator |
| -45 | `m8-flink` optional Flink session runtime |
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

## Optional System Components

Optional platform dependencies live under:

- `gitops/components/system/<component>`

They are enabled by applying an Argo CD `Application` manifest from:

- `gitops/optional/<component>/application.yaml`

Flink is the first optional system component:

- Operator Application: `m8-flink-operator`
- Runtime component: `gitops/components/system/flink`
- Enabling Application: `gitops/optional/flink/application.yaml`
- Target namespace: `m8-data`

Enable it after bootstrap:

```bash
kubectl apply -f gitops/optional/flink/application.yaml
```

Flink is intentionally not installed by the default root `ApplicationSet`. Production deployments must mirror and digest-pin the operator and runtime images from the release catalog and replace local filesystem checkpoints with object storage. Runtime clusters are declared through `flink.apache.org/v1beta1` `FlinkDeployment`; jobs should be added separately as `FlinkSessionJob` manifests.

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
