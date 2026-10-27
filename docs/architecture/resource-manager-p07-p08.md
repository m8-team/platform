# P0.7/P0.8 — Unified YDB Resource Catalog and atomic Organization

Canonical Protobuf package: `m8.platform.resourcemanager.v1`.

## Resource ownership

Resource Manager owns tenant, group, organization, workspace, project, environment and workload service records. Employees and temporal memberships remain in Identity/Directory; `org_unit_closure` continues to describe OrgUnit-to-OrgUnit edges only.

## Single durable YDB catalog

After applying `001_init.sql`, `002_org_structure.sql` and `003_unified_resources.sql` the gRPC Fx module, when configured with `M8_RM_YDB_DSN`, uses the durable `ydb.Catalog` for all currently implemented v1 resource methods.

The catalog stores resource identity, parent identity, tenant ID, derived owning organization, revision, lifecycle and environment key in `resource_records`. Its associations live in distinct group-affiliation and environment-binding tables. It preserves the existing typed resource contracts.

The old in-memory backend is retained exclusively for local tests and explicitly chosen development runs; a configured YDB connection failure never silently falls back to memory.

## Atomic CreateOrganization

A serializable YDB transaction writes:
1. Organization record, with immutable ID and tenant ownership.
2. `org_structure_heads` for that Organization at revision 1 and a unique immutable empty snapshot ID (zero OrgUnits).
3. A tenant-scoped idempotency key bound to a payload digest and the created resource identity.
4. A `m8.resourcemanager.resource.created.v1` Outbox event.

All four are committed or rolled back together. Retrying the same request returns the same Organization; changing the payload under the same key returns Conflict. The previous two-step OrgUnit bootstrap is bypassed for the durable backend.

### Related persistence

- All resource creations use the same transactional idempotency/outbox protocol.
- Update, state changes and MoveProject use serializable revision checks.
- Cross-organization project moves and environment bindings are prohibited.
- Environment key uniqueness is materialized in its own YDB table.
- Group affiliation never grants authorization.

## Verify

```sh
make fmt-check-resourcemanager
go test -race ./internal/resourcemanager/... ./cmd/resource-manager/...
# Real-YDB CI runs:
M8_RM_TEST_YDB_DSN=grpc://127.0.0.1:2136/local \
M8_RM_TEST_YDB_PREFIX=/local/m8 \
M8_RM_TEST_YDB_BOOTSTRAP=1 \
go test -v ./internal/resourcemanager/adapter/ydb -run TestIntegrationUnifiedCatalogAndAtomicOrganization
```

The test bootstraps schema only in an explicitly disposable local-YDB environment. Production migrations must be applied by the real schema migration process, not test code.

## Not included yet

- Authorization: the legacy general ResourceManagerService still fails closed outside `AllowUnauthenticated`; trusted server actor extraction and resource-by-resource M8 Access policies belong to P0.9.
- Filtered ListResources pagination with privacy-aware scoping, durable ResourceKindPolicy registration, and strong Access derived-revision barriers.
- Complete historical data migration and rollout, transactional outbox dispatch/replay workers and snapshots GC.
- Long-running workflow state and bitemporal OrgUnit history. Real YDB failure injection and load tests must follow.
- Some existing `v1engine` helpers survive as type/validation utilities and dev-only backend; they are not authoritative production storage.

Never enable `AllowUnauthenticated` with a configured YDB connection.

## Extended verification scenarios

The live-YDB integration suite executes the actual `003_unified_resources.sql` migration and verifies concurrent create attempts using one idempotency key, OrgUnit revision-1 creation, durable read-back from a new repository instance, optimistic conflicts, environment uniqueness, project moves, lifecycle restrictions, group unlink, binding removal, and scoped Outbox event counts. A failed serializable concurrent attempt is expected to be safely retryable by the client with the original idempotency key.
