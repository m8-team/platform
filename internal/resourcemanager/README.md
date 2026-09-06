# M8 Resource Manager

## Architecture and development

Requires Go 1.27; the repository selects toolchain Go 1.27.1. Run
`go run ./cmd/resource-manager`, `go test ./...`, and
`go test -race ./internal/resourcemanager/... ./cmd/resource-manager/...`.

The module can be hosted in a modular monolith or an independently deployed API.
`resourcemanager.New(config, dependencies)` is the explicit composition boundary;
it does not import Fx or choose storage/authorization adapters. The executable
chooses local adapters in `cmd/resource-manager/module.go` and uses Fx only to
connect process lifecycle and transports.

Application slices are organized by resource:

```text
internal/resourcemanager/
  domain/organization/       aggregate, lifecycle, typed ID, persistence snapshot
  domain/workspace/          workspace aggregate and immutable parent relationship
  app/organization/         create/get/list/update/delete/undelete, commands, queries
  app/workspace/            corresponding workspace slices
  app/ports/                persistence, hierarchy and authorization contracts
  app/integration/          cross-resource use-case and hierarchy tests
  adapter/                  gRPC, memory, authorization and system implementations
  module.go                 explicit constructor
```

Each use case has its own file. Filter parsers belong to service instances;
there is no package-level initialized CEL parser. Domain behavior remains free
of transport, logging and tracing. Observability is attached to application
boundaries, so both REST and gRPC calls produce the same operation signals.

The local repository stores immutable copies and clones only the returned List
page. UUID ordering compares bytes without allocating formatted strings. Label
filtering reads individual values without copying maps. Hierarchy locks are keyed
by organization, honor context cancellation and are removed when no callers wait.
These optimizations preserve atomic version checks and detached return values.

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
- ListWorkspaces
- UpdateWorkspace
- DeleteWorkspace
- UndeleteWorkspace
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

`WorkspaceService` is also implemented end to end for gRPC and REST/JSON. It
supports create, get, list, update, soft-delete, and undelete, preserves the
immutable parent Organization relationship, and permits creation only below an
`ACTIVE` Organization. Workspace mutations use the same version and retention
semantics as Organization mutations.

## Organization Semantics

- IDs are server-generated canonical UUIDs.
- New organizations start in `ACTIVE` with version `1`.
- `name`, `description`, and `labels` are the only mutable fields. Create accepts
  `OrganizationInput` / `WorkspaceInput`, not the response resource message.
  Labels are limited to 64 entries, nonempty keys up to 128 characters and values
  up to 256 characters, enforced by both the domain and protobuf contracts.
- Update, delete and undelete use optimistic compare-and-swap; API version `0` means no
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

List filters use CEL syntax. The supported v1 subset is equality on `state`,
`name`, and `labels.<key>` or `labels["key"]`, joined with `&&`. State also
supports membership expressions such as `state in ["ACTIVE", "SUSPENDED"]`.
CEL parsing and resource limits come from `internal/platform/filter`; Resource
Manager translates its neutral predicates into the module-owned typed
repository filter and validates states, operators, and duplicate conditions.
Ordering accepts one of `id`, `name`, `create_time`, or `update_time`,
optionally followed by `asc` or `desc`.

## Workspace Semantics

- Workspace has its own typed ID, state, errors, aggregate, and persistence
  snapshot; it does not reuse the Organization aggregate.
- The parent `organization_id` is required, immutable, and preserved by
  persistence clones and rehydration.
- Create and undelete require an `ACTIVE` parent Organization.
- Workspace and Organization hierarchy mutations share a coordination boundary
  so an Organization cannot be deleted concurrently with child creation or
  restoration in the in-memory adapter.
- Organization delete is rejected while a non-deleted Workspace exists.
- List supports the same CEL subset and ordering fields as Organization List.
  Its HMAC-signed keyset token is also bound to `organization_id`, preventing a
  cursor issued for one parent from being reused under another.

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

The Resource Manager module is composed through `resourcemanager.New(config, dependencies)`
and validates configuration and required dependencies before serving requests.

The `resource-manager` process also accepts:

- `M8_HTTP_ADDR` (default `:8080`; REST listener);
- `M8_HEALTH_HTTP_ADDR` (default `:8081`; health, readiness and startup listener);
- `M8_GRPC_ADDR` (default `:9090`);
- `M8_RM_ALLOW_UNAUTHENTICATED` (default `false`);
- `M8_RM_SOFT_DELETE_RETENTION` (default `720h`);
- `M8_RM_PAGE_TOKEN_KEY` (optional, at least 32 bytes).
- `M8_OTLP_HTTP_ENDPOINT` (optional OTLP HTTP trace endpoint, for example
  `http://localhost:4318/v1/traces`). Without it, traces have no external exporter.

## Observability and process lifetime

Application operations emit structured JSON logs to stderr with operation,
outcome, duration, request ID and trace ID. Payloads, credentials and raw error
strings are not logged. W3C trace context is accepted over HTTP and gRPC; new
traces use a parent-based 10% sampler. The management listener exposes JSON
operation counters (`calls`, `failures`, cumulative `duration_seconds`) at
`/metrics`, separate from public resource routes. These are local counters,
not a Prometheus/OpenMetrics endpoint. Production should restrict management access.

HTTP listeners set header/read/write/idle timeouts. HTTP and gRPC serving tasks
are supervised: unexpected listener errors request process shutdown with a
nonzero exit code. Shutdown drains servers, force-closes timed-out HTTP servers,
joins background work and flushes telemetry. Health no longer probes `ya.ru`;
only actual dependencies should contribute readiness checks.

Authorization is deny-by-default. `M8_RM_ALLOW_UNAUTHENTICATED=true` selects an
explicit local-development adapter and must not be enabled in production.
The current `organizations.list` permission is a global-list grant. An Access
adapter that supports per-organization visibility must add that visibility
constraint to the repository query as well as its `ScopeKey`.

## Current Adapter Scope

The current composition deliberately uses in-process Organization and Workspace
repositories. The Workspace repository also supplies the hierarchy check and
coordination lock consumed by Organization and Workspace mutations. This is
suitable for local development and contract tests, not durable production
storage. A production composition must replace these with module-owned durable
storage and one transaction/coordinator spanning parent state checks and child
mutations. It must also implement a transactional outbox and persistent
idempotency, and connect authorization to M8 Access. Audit publication and the
deployment's rate-limit policy are not wired yet. Returned
LROs are completed synchronously; an Operations polling backend is not
registered yet.

## Contract migration

Compatibility is intentionally not preserved for create payload protobuf types:
clients must regenerate against `OrganizationInput` / `WorkspaceInput`. The
HTTP create body still contains `name`, `description` and `labels`; server-owned
fields are rejected as unknown input. Undelete now accepts `version` just like
delete. Existing label sets exceeding the new limits need normalization before
rehydration. Page tokens issued for ID filters should be discarded and lists
restarted because identifiers now participate in the canonical request hash.

Service APIs remain contract-only. The unused `doman/service` stubs were removed;
no service persistence or handlers are implied by the existence of protobufs.
