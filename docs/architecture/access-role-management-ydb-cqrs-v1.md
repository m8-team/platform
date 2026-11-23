# M8 Access — RoleBundle and Scoped RoleAssignment Persistence v1

**Status:** PROPOSED. This change adds durable YDB state and an application boundary for composite role profiles and staged scoped assignments. It does **not** activate permissions, publish public Protobuf RPCs, or alter production PDP/SpiceDB code.

## Domain and ownership

The module-owned Catalog remains the authority for immutable, version-pinned role templates. Access owns:

- Immutable global `RoleBundle` / `AccessProfile` releases, with nested version-pinned profiles and permissions sourced from published template contracts.
- Independent tenant `ProfileAssignment` aggregates, each identified by tenant + UUID. One profile can cover multiple modules, but every resource type must receive its **own explicit Scope**. An assignment may bind `finance.invoice` to `ORGANIZATION` and `documents.file` to `INSTANCE` without implying rights over either module's unrelated resources.
- Audit/outbox events and CQRS materialized indexes.

Identity and the authoritative resource owners retain their existing responsibilities. Access must never infer organization, ancestry or ownership from a caller-supplied claim.

## Safety-critical activation boundary

**All new ProfileAssignment state is `PENDING_PROJECTION` or `REVOKED` only.** No `ACTIVE` state exists in this increment. Saving a staged assignment is *not* a permission grant and must never be consumed as a positive PDP decision.

The next stage must atomically publish tenant policy-journal revisions and full fanout jobs, project and reconcile each assignment into SpiceDB with distributed fencing, and certify ZedToken + resource revision in the existing Policy Consistency Barrier. Only after that can a reviewed `ACTIVE` transition and public API be enabled. This design prevents an incomplete storage integration from creating an authorization bypass.

The existing `policyjournal.Mutation` only permits `Grant` and `RoleBinding` keys. This PR deliberately does not masquerade a new ProfileAssignment as an existing RoleBinding; a new projection kind, semantic relationships and policy journal contract need an explicit reviewed migration.

## Application service

`app/rolemanagement.Service` defines trusted inbound ports:

- `Identity` derives an authenticated principal.
- `Authorizer` authorizes publishing global profiles and managing/reading a tenant subject's assignments.
- `TemplateResolver` binds every template version to verified module catalog definitions and allowed scopes.
- `OwnerScopeVerifier` obtains a trusted authoritative revision for every selected scope without a direct Resource Manager dependency.
- `BundleRepository` / `AssignmentRepository` are persistence ports. No infrastructure SDK in the domain or app package.

Publication checks nested profile cycles and all referenced immutable versions before writing. Assignment requires *exactly* the resource types used by every role in the expanded bundle, no hidden wildcard or extra type. Missing owner proof fails closed. `RevokeAssignment` changes only the named independent UUID.

These interfaces are not automatically satisfied or wired by public HTTP handlers. Production composition must provide actual authenticated adapters and authoritative owner/catalog resolvers before exposing any API.

## YDB write model

Apply additive `internal/access/adapters/ydbroles/schema.sql` after creating `m8-access`:

| Entity | Write table | Read / audit tables |
| --- | --- | --- |
| Bundle | `role_bundle_heads`, `role_bundles` | `role_bundle_read_index`, `role_bundle_receipts` |
| Assignment | `profile_assignments` | `profile_assignment_subjects`, `profile_assignment_receipts` |
| Events | `role_management_outbox` | Immutable per-stream/tenant/aggregate/revision events |

Every write uses one YDB `SERIALIZABLE` transaction. Bundle publication is version immutable and guarded by the expected per-profile head. Creation is idempotent on tenant+assignment ID+request key, and revocation requires the expected assignment revision. Exact idempotency retries reconstruct the historical committed snapshot from the durable outbox, not the aggregate's newest mutable state.

A read model is written in the same transaction as the authoritative aggregate and outbox. `ListBundleVersions` and `ListBySubject` are bounded to 100 records; the latter is only callable after authorization in the app service. Paging is by explicit revision or UUID, not an access-token claim. These are stored request views, never entitlement evaluations.

Outbox events `RoleBundlePublished`, `ProfileAssignmentStaged`, and `ProfileAssignmentRevoked` are stored atomically. Dispatch is at least once, and `MarkPublished` is for a future trusted dispatcher after broker acknowledgment. Event consumers MUST NOT translate a `ProfileAssignmentStaged` event to a SpiceDB allow relationship.

## Rollout and rollback

Do not deploy the new migration automatically. The current AuthZEN/PDP, RoleBinding projection, YDB policy journal and SpiceDB graph remain intact. Disable new service composition to roll back; retain assignment rows for audit. No existing production writes are redirected.

## Acceptance tests

- Domain multi-module scope isolation, revoke CAS, restoration and no active status.
- Storage snapshot/manifest codec, immutable profile revisions, idempotent retries, request conflicts, subject CQRS index, historical replay.
- Disposable live YDB migration, concurrency single-CAS winner, outbox ACK idempotency and rollback of all writes on an unavailable outbox.
- Existing SpiceDB/YDB policy barrier regression workflow remains enabled.

### Next milestones

1. Contract-first publish/list/assign/revoke Protobuf APIs, authenticated and permission-checked composition root, no production activity until trusted resolvers exist.
2. Policy Consistency v3 with explicit profile assignment projection kinds, full per-role/per-scope fanout, distributed worker fencing and reconciliation. Future activation/revocation must advance the policy head atomically with all authoritative writes.
3. Contextual policy YDB persistence and verified IP/MFA/resource attributes at AuthZEN; tenant business attributes remain at owner modules.
