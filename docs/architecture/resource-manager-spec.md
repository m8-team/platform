# Resource Manager Specification for M8 Platform

Status: Draft for implementation  
Document path: `docs/architecture/resource-manager-spec.md`  
Target domain: `M8 Platform / Control Plane / Resource Manager`

## Overview

Resource Manager is the canonical system domain in the M8 Platform control plane responsible for the platform resource catalog and its hierarchical ownership model.

In the current contract set, Resource Manager owns exactly three resource types:

- `Organization`
- `Workspace`
- `Project`

The canonical hierarchy is fixed as:

```text
Organization
  -> Workspace
       -> Project
```

This document is intentionally grounded in the existing protobuf contracts and service APIs:

- `organization.proto`
- `organization_service.proto`
- `workspace.proto`
- `workspace_service.proto`
- `project.proto`
- `project_service.proto`

The goal of this specification is not to invent a new model, but to define how the current contracts should be implemented as a coherent domain, application layer, persistence model, and event source for adjacent platform services.

## Scope

Resource Manager is in scope for the following responsibilities:

- canonical storage of `Organization`, `Workspace`, and `Project`
- canonical parent-child hierarchy between these resources
- lifecycle management for these resources
- soft-delete and undelete semantics
- optimistic concurrency through `etag`
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
- mutable fields: `name`, `description`, `annotations`
- server-controlled fields: `state`, `create_time`, `update_time`, `delete_time`, `purge_time`
- delete is soft-delete only
- delete is blocked while there are non-deleted workspaces

### Aggregate: Workspace

- identity field: `id`
- parent field: `organization_id`
- type: UUID string
- parent link is immutable after creation
- mutable fields: `name`, `description`, `annotations`
- server-controlled fields: `state`, `create_time`, `update_time`, `delete_time`, `purge_time`
- delete is soft-delete only
- delete is blocked while there are non-deleted projects
- undelete requires existing non-deleted parent organization

### Aggregate: Project

- identity field: `id`
- parent field: `workspace_id`
- type: UUID string
- parent link is immutable after creation
- mutable fields: `name`, `description`, `annotations`
- server-controlled fields: `state`, `create_time`, `update_time`, `delete_time`, `purge_time`
- delete is soft-delete only
- undelete requires existing non-deleted parent workspace
- enum contains `ARCHIVED`, but public API has no archive RPC yet

## Invariants

- `Organization` has no parent.
- `Workspace` always belongs to exactly one `Organization`.
- `Project` always belongs to exactly one `Workspace`.
- Parent link is immutable after creation.
- Child creation requires an existing, non-deleted parent.
- Undelete of a child requires an existing, non-deleted parent.
- `etag` changes on every successful mutating write.
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
- `CreateProject`
- `UpdateProject`
- `DeleteProject`
- `UndeleteProject`

Queries:

- `GetOrganization`
- `ListOrganizations`
- `GetWorkspace`
- `ListWorkspaces`
- `GetProject`
- `ListProjects`

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

Project events:

- `project.created`
- `project.updated`
- `project.state_changed`
- `project.archived`
- `project.unarchived`
- `project.deleted`
- `project.undeleted`
- `project.purged`

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
