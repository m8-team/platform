# Resource Manager

> **Contract replacement in progress (2026-10-25):** the earlier Organization → Workspace → per-environment Service API has been removed from protobuf source. The new **canonical v1** is defined in [Resource Manager Specification](../../architecture/resource-manager-spec.md). Regenerated gRPC/REST/OpenAPI reference pages are pending and the legacy API-reference pages below are **not** applicable to the new v1.

## Canonical ownership model

```text
Tenant / Account
  ├── Organization (mandatory ownership and tenant isolation root)
  │     ├── Workspace (optional)
  │     ├── Environment (production/testing/staging/etc; not a parent scope)
  │     └── ResourceDescriptors (Agents, Workflows, etc by owner policy)
  └── OrganizationGroup (optional; affiliation to organizations, not permission inheritance)
```

`WorkloadService` is a logical service owned by a Workspace and is deployed/bound to Environments separately. `EnvironmentBinding` records approved target versions/configuration references; runtime rollout belongs to Delivery.

Resource Manager is the source of truth for ownership and lifecycle metadata, not for employee records, authorizations, secrets or cross-company Service Agreements.

## New v1 protobuf contract

[ResourceManagerService](https://github.com/m8-team/platform/blob/main/api/proto/m8/platform/resourcemanager/v1/resource_manager_service.proto) (link resolves after merge) contains RPC definitions together with their request and response messages. Reusable entity contracts live in `resource_manager.proto`, `tenant.proto`, `organization.proto`, `organization_group.proto`, `workspace.proto`, `environment.proto`, `service.proto` and `resource.proto`. Organization relationship and org-structure APIs follow the same convention.

The existing legacy RPC/REST reference pages are retained solely as historical material until replacement documentation is generated. Do not use them for new client development.
