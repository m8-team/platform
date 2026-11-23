# M8 Access Catalog — YDB Persistence + CQRS v1

**Status:** implementation proposed in stacked PR, dependent on Universal Catalog API PR #67. No production deployment, runtime composition or Kafka delivery is performed by this PR.

## Authoritative boundaries

- **Access Catalog** stores immutable authorization *definitions* from registered module publishers (resource types, permissions, role templates). It does not write source module business objects, tenant grants, role assignments or policy decisions.
- **YDB authoritative command state:** `catalog_heads`, `catalog_releases`, `catalog_receipts`.
- **CQRS read projection:** `catalog_read_index` keyed by `(module_id,revision)`, updated in exactly the same serializable YDB transaction as the release.
- **Durable outbox:** `catalog_outbox`, keyed by `(module_id,revision)`, updated within that same transaction. This outbox is *not* a SpiceDB policy mutation, and its publication does not advance the existing policy consistency barrier.
- All catalog rows are isolated by module namespace. Trusted publishing/reading authorization is enforced by the **application** ports from PR #67; direct Store methods are internal, never public APIs.

## Persistence contract

1. The app authenticates service identity, verifies its own namespace and constructs a validated immutable Domain Manifest. The YDB store separately validates it and rejects graph permissions other than `use`.
2. `Publish` validates the request and computes a stable SHA256 digest of the canonical storage DTO **plus expected head revision**. The canonical DTO includes authenticated publisher identity.
3. In one `SERIALIZABLE` YDB transaction it first loads `catalog_receipts`. For the same key and exact digest it returns the original revision, even if newer releases have been committed since. Reuse of a key for other content fails `ErrCatalogConflict`.
4. Otherwise it checks `catalog_heads` against `expected_revision` (initial publication expects zero), rejects existing release version, and inserts an immutable release, new head, receipt, read projection and outbox event.
5. Any write/query/commit failure aborts the entire transaction. The remote caller is never allowed to substitute a newer expected revision automatically.
6. The event has deterministic identity `catalog/{module}/{revision}`, `ModuleManifestPublished` kind and full canonical manifest snapshot. Consumers must deduplicate event ID; broker delivery is at least once. The final YDB commit is the only event-publication authority.

## Read-side contract

- `Get` reads one immutable `(module_id,version)` release from the authoritative release table and reconstructs a validated Domain Manifest; malformed data fails closed.
- `List` reads the revision-keyed materialization. The first page captures the current module head; subsequent pages use a signed cursor including principal type/ID, module ID, last emitted revision, snapshot upper revision bound and expiration.
- Cursor signatures use HMAC-SHA256 with a required injected secret of at least 32 bytes. The secret must originate from a secret manager, not a repository file. Cursors are subject/module bound, expire within a configured lifetime (1 minute to 24 hours), and use fixed snapshot upper bounds to exclude later releases.
- `List` is bounded to 100 records per page; callers must be authorized by `app/catalog.API` before accessing these repository methods. Store itself never establishes identity from headers.
- Catalog definitions are global namespace descriptors, **not tenant permissions**. Tenant-scoped entitlement queries and object-level access decisions remain in the separate Access/PDP pipelines.

## Migration and operational rollout

Migration is **additive** at `internal/access/adapters/ydbcatalog/schema.sql` and must be applied to the selected YDB namespace after creating `m8-access`. It creates five new tables without altering existing Access lease, policy or relationship tables. Do not apply migration to production until service config and rollback have been reviewed.

Wiring (later PR): create YDB SQL driver, load signed-cursor key from secret manager, construct Store via `ydbcatalog.New`, wire it as `catalog.Registry`, and supply authenticated `catalog.IdentityResolver` and `catalog.CatalogAuthorizer` to the gRPC composition root. No automatic migration or silent in-memory fallback.

The durable outbox exposes `Pending(ctx,module,after,limit)` and `MarkPublished(ctx,module,revision,eventID)` for a separately managed worker. **Only mark published after external broker ack.** A crash after broker confirmation but before ack duplicates the delivery; the consumer must deduplicate. For scalable operation, add explicit per-module dispatcher ownership/lease and dead-letter reporting before production.

### Rollback

Disable catalog RPC registration/dispatcher, keep published definitions and immutable event history for audit, and restart the old Access runtime. The migration does not affect current AuthZEN/SpiceDB policy checks. Do not drop catalog tables while a publisher or outbox consumer can still access them.

## Tests

- Unit: canonical storage DTO serialization round-trip, malformed JSON, size limits and version-sensitive request hash; cursor signature, cross-subject/namespace denial, expiry and configuration.
- Real disposable YDB (`-tags=integration`): apply exact migration, publish, repeat identical idempotent request, conflicting reuse, stale head CAS, immutable release, stable revision pagination, cross-principal cursor denial, pending outbox and idempotent ack, two concurrent publishers with one winner and no partial read/outbox projection.
- Existing Access YDB/SpiceDB integration workflow additionally runs this adapter's tests. No claim of full Kafka delivery, live mTLS publisher credentials, or production load tests is made.

## Further CQRS milestones

YDB stores and projection workers for **RoleBundle, RoleAssignment, AccessGrant and AccessPolicy** require separate write commands and approved external contracts. Their changes *can* grant or revoke entitlements and therefore MUST participate in the certified tenant policy journal, projection fencing, resource ownership and policy consistency barrier. They intentionally are **not** implemented as side effects of this catalog migration.
