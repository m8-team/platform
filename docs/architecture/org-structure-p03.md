# OrgUnit P0.3 — YDB storage adapter (draft)

The implementation uses the canonical `m8.platform.resourcemanager.v1` domain and the tables from `002_org_structure.sql`. It requires a `*sql.DB` connected using the official YDB Go SDK database/sql driver; this change does **not** add a driver or service wiring.

## Implemented in Go

- Tenant-and-organization-scoped reads of active head, immutable OrgUnit snapshot and closure materialization.
- Snapshot integrity check: recompute adjacency closure and compare every stored ancestor/descendant/depth edge.
- Persistent ChangeSet replay keyed by `(tenant_id, organization_id, change_set_id)`, validated with request hash.
- Snapshot staging under unpredictable unique IDs, in a transaction, before publication.
- Serializable publication transaction: re-read head and compare both revision and snapshot ID; write ChangeSet receipt, newly created stable OrgUnit identities, head update and outbox event.
- An unchanged ChangeSet records an idempotent receipt without producing a structure-change event.
- Privileged organization initialization creates the initial snapshot and stable OrgUnit identities.
- SQL table-path prefix validation and a temporary 10,000-row staging bound to fail safely until scalable chunked staging is implemented.

## Deployment blockers

1. **No live YDB verification yet**: YQL statements, JsonDocument casting, named parameters, multi-statement transactions and conflict handling need integration tests against the deployed YDB version.
2. **No production gRPC handlers, authz policy implementation or service bootstrap**. The upstream `DecisionGate` must be fail-closed. Do not expose `Initialize` as an unguarded endpoint.
3. The current code does **not** populate `org_unit_versions` with full bitemporal corrections. It supports only immediate changes; replaying past snapshots is not equivalent to backdated HR history. Scheduled/backdated synchronization requires a separate temporal writer.
4. Staging currently uses one transaction and a bounded graph size; a chunked immutable snapshot builder, completeness markers, retention-safe garbage collector and repair tooling are required before large hierarchies.
5. During concurrent updates YDB may report a serializable transaction conflict instead of the typed `ErrRevisionConflict`. Add a retry-or-reload wrapper that never repeats a non-idempotent effect.
6. Add end-to-end tests for crash between staging/publishing, partial staging, concurrent writers, retained history, outbox delivery, access revision fences and tenant separation.

## Correct usage

- Migrate baseline 001 then 002.
- Inject an authenticated actor and a production `orgstructure.DecisionGate` before invoking the application service.
- Use the application service for writes; never call repository `Commit` directly from HTTP/gRPC.
- Keep Person, OrganizationMembership and OrgUnitAssignment in Identity/Directory, never in the closure table.

## Validation

Isolated local Go 1.23 harness: `go test -race ./internal/resourcemanager/...` and `go vet ./internal/resourcemanager/...` passed. Repository-wide Go 1.26 CI, SDK integration, Buf generation and live YDB tests remain unverified.
