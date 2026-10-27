# Resource Manager P0.4 — authorized OrgUnit read model

**Canonical package:** `m8.platform.resourcemanager.v1`.  
**Status:** Go query application service implemented; gRPC transport pending generated stubs.

## Read API application

`internal/resourcemanager/app/orgstructure/query.go` provides:

- `Get` — stable OrgUnit identity and snapshot revision.
- `List` — organization list or descendant subtree; optional lifecycle filter.
- `Ancestors` — ordered authorized ancestors including self (depth 0).
- `Descendants` — authorized descendants and depth; paginated.
- HMAC-SHA256 signed keyset pagination bound to tenant, principal, organization, subtree, filter and active revision; 15-minute TTL.
- Strong revision guard: a subsequent page after reorganization returns `ErrRevisionConflict` (map to gRPC `ABORTED` and have client restart listing).
- Two distinct authorization checks: organization-level and individual OrgUnit-level; *filter before pagination* to avoid enumeration leaks.

`HierarchyReader` is satisfied by the existing YDB `Store.Load` from P0.3. Its snapshot loader must remain revision-consistent.

## Security invariants

1. Authentication supplies Actor. Never accept tenant/principal ID from request body as trusted.
2. Authorizer and signing key are mandatory at construction; no production permissive defaults.
3. Filtering by unit occurs before page boundaries. Unreadable ancestors and hidden subtrees do not reveal their contents.
4. Do not rely on a previously authorized page token as lasting permission; the gate runs on every request.
5. Do not serve `ListOrgUnitVersions` until a consistent bitemporal history projection is implemented.
6. Tenant-aware cache keys and outbox revision barriers are required before exposing org-derived decisions to M8 Access.

## Implementation left before deployment

- Buf regeneration for OrgStructureService and ChangeSetService: committed `.proto` files have no generated Go client/server types yet.
- Concrete gRPC adapter mapped to these use cases and Access-backed authorization gate.
- Durable page token key management/rotation, rate limits and query metrics.
- Live YDB SDK integration tests (serializable conflicts, failed publication, idempotency).
- Bitemporal history writer and read API.

## Test command

`go test -race ./internal/resourcemanager/app/orgstructure ./internal/resourcemanager/domain/orgunit`

The isolated package tests can run with Go 1.23, but full platform CI must validate the repository's current Go toolchain.
