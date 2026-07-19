# Resource Manager Specification for M8 Platform

Status: Draft for implementation  
Document path: `docs/architecture/resource-manager-spec.md`  
Target domain: `M8 Platform / Control Plane / Resource Manager`

## Overview

Resource Manager is the canonical system domain in the M8 Platform control plane responsible for the platform resource catalog and its hierarchical ownership model.

In the current contract set, Resource Manager owns exactly three resource types:

- `Organization`
- `Workspace`
- `Service`

The canonical hierarchy is fixed as:

```text
Organization
  -> Workspace
       -> Service
```

This document is intentionally grounded in the existing protobuf contracts and service APIs:

- `organization.proto`
- `organization_service.proto`
- `workspace.proto`
- `workspace_service.proto`
- `service.proto`
- `service_service.proto`

The goal of this specification is not to invent a new model, but to define how the current contracts should be implemented as a coherent domain, application layer, persistence model, and event source for adjacent platform services.

## Implementation Status

`OrganizationService` now has a clean-architecture vertical slice covering its
six protobuf RPCs. It includes lifecycle domain behavior, optimistic
concurrency, authorization and hierarchy ports, signed keyset pagination,
manual RPC-aware validation, gRPC status mapping, and completed LRO responses.

The executable currently composes process-local persistence and a placeholder
Workspace-child projection. These adapters are for local development and
contract testing only. Durable storage, transactional outbox/idempotency,
Access integration, and a pollable `google.longrunning.Operations` backend
remain production follow-up work.

## Scope

Resource Manager is in scope for the following responsibilities:

- canonical storage of `Organization`, `Workspace`, and `Service`
- canonical parent-child hierarchy between these resources
- immutable environment assignment for each `Service`
- lifecycle management for these resources
- soft-delete and undelete semantics
- optimistic concurrency through `version`
- emission of resource domain events for downstream consumers
- list/get query semantics for control plane and read models

Resource Manager is explicitly out of scope for:

- IAM roles, permissions, memberships, group assignments
- provisioning of infrastructure or tenant runtime resources
- billing calculations, invoices, charging, or quota enforcement logic
- audit storage and retention of audit records
- search indexing implementation details and UI read model ownership

In practical terms, Resource Manager is the source of truth for **resource existence, identity, parent linkage, and lifecycle state**, but not the source of truth for access control, deployment state, financial state, or audit evidence.

## Aggregate Model

The domain naturally fits three aggregates, each mapped directly to an existing proto message.

### Aggregate: Organization

- identity field: `id`
- type: UUID string
- assigned by server
- immutable after creation
- mutable fields: `name`, `description`, `labels`
- server-controlled fields: `state`, `create_time`, `update_time`, `delete_time`, `purge_time`
- delete is soft-delete only
- delete is blocked while there are non-deleted workspaces

### Aggregate: Workspace

- identity field: `id`
- parent field: `organization_id`
- type: UUID string
- parent link is immutable after creation
- mutable fields: `name`, `description`, `labels`
- server-controlled fields: `state`, `create_time`, `update_time`, `delete_time`, `purge_time`
- delete is soft-delete only
- delete is blocked while there are non-deleted services
- undelete requires existing non-deleted parent organization

### Aggregate: Service

- identity field: `id`
- parent field: `workspace_id`
- environment field: `environment`
- type: UUID string
- parent link is immutable after creation
- environment is a required immutable lowercase slug such as `dev`, `stage`, or `prod`
- the same logical application in different environments is represented by separate services
- mutable fields: `name`, `description`, `labels`
- server-controlled fields: `state`, `create_time`, `update_time`, `delete_time`, `purge_time`
- delete is soft-delete only
- undelete requires existing non-deleted parent workspace
- enum contains `ARCHIVED`, but public API has no archive RPC yet

## Invariants

- `Organization` has no parent.
- `Workspace` always belongs to exactly one `Organization`.
- `Service` always belongs to exactly one `Workspace`.
- `Service` always belongs to exactly one environment.
- `environment` is immutable and must match the API slug validation rules.
- Parent link is immutable after creation.
- Child creation requires an existing, non-deleted parent.
- Undelete of a child requires an existing, non-deleted parent.
- `version` changes on every successful mutating write.
- `show_deleted=false` excludes soft-deleted resources from list responses.
- There is no hidden recursive delete in v1.

## Commands and Queries

Commands:

- `CreateOrganization`
- `UpdateOrganization`
- `DeleteOrganization`
- `UndeleteOrganization`
- `CreateWorkspace`
- `UpdateWorkspace`
- `DeleteWorkspace`
- `UndeleteWorkspace`
- `CreateService`
- `UpdateService`
- `DeleteService`
- `UndeleteService`

Queries:

- `GetOrganization`
- `ListOrganizations`
- `GetWorkspace`
- `ListWorkspaces`
- `GetService`
- `ListServices`

## Domain Events

Organization events:

- `organization.created`
- `organization.updated`
- `organization.state_changed`
- `organization.deleted`
- `organization.undeleted`
- `organization.purged`

Workspace events:

- `workspace.created`
- `workspace.updated`
- `workspace.state_changed`
- `workspace.deleted`
- `workspace.undeleted`
- `workspace.purged`

Service events:

- `service.created`
- `service.updated`
- `service.state_changed`
- `service.archived`
- `service.unarchived`
- `service.deleted`
- `service.undeleted`
- `service.purged`

Envelope fields:

- `event_id`
- `event_type`
- `event_version`
- `aggregate_type`
- `aggregate_id`
- `parent_aggregate_id`
- `occurred_at`
- `actor`
- `correlation_id`
- `causation_id`
- `idempotency_key`
- `etag_or_revision`
- `payload`
