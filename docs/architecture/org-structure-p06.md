# M8 Resource Manager P0.6 — YDB, Organization Bootstrap and M8 Access

Protobuf package: `m8.platform.resourcemanager.v1` (unchanged).

## Implemented

- Introduced the official YDB Go SDK `github.com/ydb-platform/ydb-go-sdk/v3`; the existing `database/sql` OrgUnit store is now selected in the gRPC module when `M8_RM_YDB_DSN` is configured.
- YDB initialization fails instead of silently falling back to a process-local store when a configured DSN is inaccessible or invalid.
- Fx closes the `database/sql` pool on shutdown.
- `M8_RM_YDB_TABLE_PREFIX` selects the schema installation path (default `/local/m8`); apply `001_init.sql` and `002_org_structure.sql` before startup.
- Read and changeset use cases now share one `orgstructure.Store` regardless of the transport (in-memory dev / YDB configured).
- A `orgaccess.Gate` adapter maps organization/unit permission checks to the existing M8 Access `CheckPermissionHandler` contract and enforces revision equality, fail-closed decisions and no fail-open/degraded results.
- Mutation impact approval remains a **separate mandatory** `ImpactReviewer`; an ALLOW from M8 Access by itself cannot approve a department reorganization. Until a real reviewer is injected, write operations fail closed.
- On creating an Organization through the legacy development `ResourceManagerService`, an empty OrgUnit structure (revision 1) is initialized in the selected store, without creating an artificial top-level OrgUnit.
- Bootstrap retries are safe: an existing structure (including higher revisions) is never overwritten by a repeated create.
- Added integration test `TestIntegrationYDBOrgStructure` against a real, pre-migrated YDB database. It requires `M8_RM_TEST_YDB_DSN` and `M8_RM_TEST_YDB_PREFIX`.

## Configuration

```sh
M8_RM_YDB_DSN='grpcs://<endpoint>/<database>'
M8_RM_YDB_TABLE_PREFIX='/local/m8'
M8_RM_ACCESS_MODEL_REVISION='<deployed model revision>'
M8_RM_PAGE_TOKEN_KEY='<32+ byte stable secret>'
```

**Security:** An M8 Access checker must be supplied through Fx. Existing `internal/access` currently has a handler but does **not** create a production engine; without a checker all typed OrgUnit operations remain denied. Trusted gRPC authentication middleware must also set the verified actor context; it is not configured by this PR.

## Remaining production blockers

1. Organization records in the existing `v1engine` still use in-memory state. The initial OrgUnit revision and Organization are **not an atomic durable YDB transaction**; a retry/repair bootstrap is supplied for the development flow, but the durable organization repository and atomic creation flow are still required.
2. No production M8 Access PermissionEngine provider, authn interceptor or independent impact reviewer is installed; production OrgUnit calls remain denied until those are configured.
3. Real-YDB integration test is opt-in and only runs on an already migrated test DB. Do not claim live DB verification from a passing unit-only CI run.
4. Existing YDB migration, snapshot limits and transactional CAS must be stress-tested on real YDB, including concurrent writers and crash recovery.
5. Bitemporal OrgUnit history and retroactive/forward-dated changes remain future work; history RPC returns UNIMPLEMENTED.
6. M8 Access revision `modelRevision` is not the same as `OrgStructureRevision`; access integration needs a revision barrier before using derived access on an updated graph.

## Commands

```sh
make fmt-check-resourcemanager
go test ./internal/resourcemanager/... ./cmd/resource-manager/...
M8_RM_TEST_YDB_DSN='grpc://localhost:2136/local' \
M8_RM_TEST_YDB_PREFIX='/local/m8' \
go test ./internal/resourcemanager/adapter/ydb -run TestIntegrationYDBOrgStructure -count=1 -v
```

The integration-test database must have migrations 001 and 002 installed; never point the test at production.
