# M8 Access Universal Catalog API v1

**Status:** contract-first API implementation for review. This PR provides Protobuf contracts, generated Go/Connect/gRPC/Gateway clients, domain mappings, authenticated application commands and queries, and gRPC handlers. **There is no production YDB Registry adapter or service composition in this PR.** The API must not be exposed in production until those dependencies are wired and integration-tested.

## Scope and ownership

The catalog is a global, module-owned authorization **definition** registry, not a resource-instance database. A publisher may define resource kinds (e.g. `finance.invoice`), actions (e.g. `finance.invoice.view`) and reusable role templates (e.g. `finance.accountant@1.0.0`) within its verified namespace. Object existence, ownership, lifecycle and revision remain with Finance / the authoritative owning service and Resource Manager. Assigning roles, creating grants, publishing policies, building SpiceDB relationships or authorizing an end-user are **not** effects of publishing catalog definitions.

The authenticated service identity is determined by trusted workload middleware; client-supplied `manifest.module_id`, headers or identity properties do not prove namespace ownership. The `CatalogAuthorizer` must separately check `CanPublishModule`. Readers (users and services) must pass `CanReadModule` before reads; unauthorized results must never be revealed by listings or cursors.

## API operations

| RPC | HTTP binding | Authorization |
| --- | --- | --- |
| `PublishModuleManifest` | `POST /access/v1/catalog/manifests:publish` | verified service account and namespace publisher permission |
| `GetModuleManifest` | `GET /access/v1/catalog/modules/{module_id}/manifests/{version}` | catalog-read permission for namespace |
| `ListModuleManifests` | `GET /access/v1/catalog/modules/{module_id}/manifests` | catalog-read permission for namespace |
| `ListCatalogResourceTypes` | `GET /access/v1/catalog/modules/{module_id}/manifests/{version}/resource-types` | catalog-read permission for namespace |
| `ListCatalogPermissions` | `GET /access/v1/catalog/modules/{module_id}/manifests/{version}/permissions` | catalog-read permission for namespace |
| `ListCatalogRoleTemplates` | `GET /access/v1/catalog/modules/{module_id}/manifests/{version}/role-templates` | catalog-read permission for namespace |

`Publish` accepts the entire immutable module release; registrations of the same release cannot modify it. Fields `publisher_subject_id` and `revision` come from authoritative storage, not the caller. Publishing a new version requires a successful compare-and-swap of the module catalog head. For a new module, `expected_revision=0`. Successful publication returns its actual committed revision, including idempotent replays.

**Idempotency contract:** a retry with the same idempotency key and equivalent payload returns the original committed `CatalogManifestRecord`. Reusing the key for a different payload is `AlreadyExists`. A stale revision is `Aborted`. No partial publication is permitted.

**Listing contract:** module releases are sorted deterministically by committed catalog revision, then version. Default `page_size=20`, maximum 100. Tokens are opaque, integrity-protected and bound to authenticated principal, module and query; expiration and handling of changed permissions are the storage adapter's responsibility. The three definition-list methods read one immutable bounded snapshot (maximum 256 resource types, 1024 permissions, 256 role templates) and intentionally do not accept a cursor.

**Error mapping:** malformed requests `InvalidArgument`, missing verified principal `Unauthenticated`, unauthorized namespace `PermissionDenied`, absent release `NotFound`, version/idempotency conflict `AlreadyExists`, stale revision `Aborted`, infrastructure/invalid repository response `Unavailable`. Public errors are sanitized.

## Security restrictions

- Dynamic module permissions use `graph_permission=use` only in this API. A module cannot map to privileged builtin relations.
- Type and action names are namespaced. `module.Manifest.Validate` prevents foreign permissions, ambiguous actions or role scopes broader than their resource types.
- Publishing a role template does not create a RoleAssignment. No access is conferred without independent entitlement.
- An `INSTANCE` or `TENANT` scope in metadata is **not** owner proof; the PDP must still obtain the authoritative tenant, resource existence/lifecycle and revision.
- This API intentionally does not accept caller-supplied IP, MFA or resource attributes. Context-based PDP remains a separate milestone.
- Unavailable publisher identity, catalog authorization or Registry storage fail closed. Never silently install an in-memory production Registry.

## Wiring / delivery milestones

1. Implement a YDB `catalog.Registry` with immutable release rows, per-module revision CAS, idempotency conflict detection, transactional event outbox and policy-journal update where appropriate. Ensure retry, stale revision, concurrent publish and restart tests.
2. Bind `cataloggrpc.Server` and generated gRPC/Gateway services through the Access composition root with authenticated mTLS/OIDC identity resolution and a trusted namespace/read authorization adapter.
3. Publish `ModuleManifestPublished` only after durable commit; synchronize SpiceDB definitions in a separately reviewed Universal Projection/Schema ADR and migration.
4. Add API integration tests on real ephemeral YDB, TLS/authenticated service calls and cross-namespace read denial.
5. Review catalog release compatibility: old clients can ignore new RPCs; new clients cannot assume the published catalog is active until schema compilation and projection barriers report certified status.

## Testing

The PR adds application tests for namespace authorization, principal verification, immutable version publication, CAS conflicts and idempotent retries, plus adapter tests for Protobuf mapping, security-negative cases and read/discovery RPCs. Buf lint/breaking and generated-code verification are tracked in the PR CI.
