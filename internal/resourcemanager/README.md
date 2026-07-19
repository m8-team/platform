# M8 Resource Manager

## Responsibility

Owns organizations, workspaces, services, service environment assignment, resource catalog metadata and platform-level resource intent.

## Owns

- Organization
- Workspace
- Service
- Immutable environment assignment for each Service
- Resource catalog metadata
- Platform-level desired and actual state summary

## Does Not Own

- Users
- Authentication sessions
- Permissions and authorization relationships
- Provider-specific provisioning drivers
- Runtime clusters and placements

## Main APIs

- CreateOrganization
- GetOrganization
- UpdateOrganization
- DeleteOrganization
- CreateWorkspace
- GetWorkspace
- UpdateWorkspace
- DeleteWorkspace
- CreateService
- GetService
- UpdateService
- DeleteService

`OrganizationService` is implemented end to end for gRPC and REST/JSON:
create, get, list, update, soft-delete, and undelete. The REST gateway shares
the process HTTP listener with the health endpoints and uses the canonical
`/resource-manager/v1/organizations...` routes. Mutations return an
already-completed `google.longrunning.Operation` containing common operation
metadata and the typed response declared by the protobuf contract.

## Organization Semantics

- IDs are server-generated canonical UUIDs.
- New organizations start in `ACTIVE` with version `1`.
- `name`, `description`, and `labels` are the only mutable fields.
- Update and delete use optimistic compare-and-swap; API version `0` means no
  client precondition.
- Delete retains a tombstone for the configured retention period. Get can read
  that tombstone, while List excludes it unless `show_deleted=true`.
- Delete is rejected when the hierarchy adapter reports a non-deleted
  Workspace.
- Undelete is accepted only for a retained tombstone and clears its deletion
  timestamps.
- List uses stable keyset pagination. Page tokens are HMAC-signed and bound to
  the caller authorization scope, effective page size, filter, ordering, and
  `show_deleted` flag.

The supported v1 filter subset is equality on `state`, `name`, and
`labels.<key>`, joined with `AND`. Ordering accepts one of `id`, `name`,
`create_time`, or `update_time`, optionally followed by `asc` or `desc`.

## Events Published

The event names below are module-owned contracts. The local in-memory adapter
does not publish them yet; production publication requires the transactional
outbox described under Current Adapter Scope.

- OrganizationCreated
- OrganizationUpdated
- OrganizationDeleted
- WorkspaceCreated
- WorkspaceUpdated
- WorkspaceDeleted
- ServiceCreated
- ServiceUpdated
- ServiceDeleted

## Events Consumed

- None currently.

## Module Configuration

The Resource Manager module is composed through `resourcemanager.Module(config)` and validates its module-level configuration during Fx application construction.

The `resource-manager` process also accepts:

- `M8_HTTP_ADDR` (default `:8080`; REST listener);
- `M8_HEALTH_HTTP_ADDR` (default `:8081`; health, readiness and startup listener);
- `M8_GRPC_ADDR` (default `:9090`);
- `M8_RM_ALLOW_UNAUTHENTICATED` (default `false`);
- `M8_RM_SOFT_DELETE_RETENTION` (default `720h`);
- `M8_RM_PAGE_TOKEN_KEY` (optional, at least 32 bytes).

Authorization is deny-by-default. `M8_RM_ALLOW_UNAUTHENTICATED=true` selects an
explicit local-development adapter and must not be enabled in production.
The current `organizations.list` permission is a global-list grant. An Access
adapter that supports per-organization visibility must add that visibility
constraint to the repository query as well as its `ScopeKey`.

## Current Adapter Scope

The current composition deliberately uses an in-process repository and an
empty Workspace-child projection. It is suitable for local development and
contract tests, not durable production storage. A production composition must
replace both with the module-owned database/projection, implement a
transactional outbox and persistent idempotency, and connect authorization to
M8 Access. It must also add audit publication, request telemetry, and the
deployment's rate-limit policy. The returned LRO is completed synchronously;
an Operations polling backend is not registered yet.
