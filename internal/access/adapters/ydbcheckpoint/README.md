# YDB distributed projection checkpoint and SpiceDB fencing (v1)

Status: **implementation for review, not deployed**. Run `schema.sql` in the M8 Access YDB database before instantiating `Store`. This does not migrate checkpoints from `filecheckpoint`; perform an explicit backfill/reconciliation before enabling worker replicas. The database path supplied to `New` is the absolute YDB TablePathPrefix. Use the existing `database/sql` YDB driver and a unique per-process `ownerID` (e.g. workload UID + random UUID), never a tenant/client-controlled identifier.

## Fenced processing sequence

1. YDB serializable transaction reads one `(tenant,kind,aggregate)` row, rejects an unexpired lease and increments its `lease_epoch` before committing. Fresh rows start at epoch 1. Lease expiry uses YDB server time, not node wall clocks.
2. Worker installs `projection_guard#owner@projection_epoch` in SpiceDB with a transaction containing `MUST_MATCH` on the observed previous guard (or `MUST_NOT_MATCH` for a new guard). A smaller epoch cannot replace a larger observed epoch. Any concurrently overwritten observation fails the precondition.
3. Every `WriteRelationships` issued by this worker carries a **graph-side** `MUST_MATCH` precondition on its current guard. Prior lease owners cannot write after takeover because the guard has changed.
4. On each fresh epoch, worker reads the **actual fully consistent** graph relationships owned by the grant/binding, compares with the authoritative Access snapshot, repairs with `TOUCH`/`DELETE`, and validates the returned ZedToken. Reading graph state is mandatory: a checkpoint snapshot alone cannot find orphan tuples written before a failed checkpoint commit.
5. A background YDB lease renewal cancels operations on failure. A serializable commit checks `(ownerID,epoch,lease_until)` and writes `source revision + graph token + full projected edges + fence_epoch` in the same transaction that releases the lease.
6. Kafka offset/outbox ack occurs only if this commit succeeds. An expired/lost lease never acknowledges. `LoadCheckpoint` rejects checkpoints whose `fence_epoch != lease_epoch` (takeover in progress or failed), keeping dependent authorization checks fail-closed.

**Ownership:** Access owns the binding/grant projection checkpoints. Resource Manager owns the authoritative source revision. The separate `resource_projection_watermarks` table is populated only by the trusted RM projection worker, after its relevant SpiceDB batch succeeded. Do not let public callers forge/advance watermarks. The current work provides the persistence API, not a Resource Manager projection worker or its complete ownership reconciliation.

## Source freshness

`consistency.Barrier` fetches the current source revision from an authoritative tenant-checking Resource Manager reader, then checks `resource_projection_watermarks` in YDB. Only a watermark with revision >= authoritative revision and a non-empty graph ZedToken can proceed to the PDP. The barrier pins `at_least_as_fresh` or `fully_consistent` if two independent opaque ZedTokens cannot be ordered. Missing/mismatched tenant, source errors, lag or missing watermark fail closed. This must run in an authenticated server/service path, in addition to business-owner conditional-write checks against owner/tenant/revision (SEC-010).

## Rollout requirements

1. Deploy updated `m8.zed` schema with `projection_epoch` and `projection_guard` to SpiceDB; verify with `zed schema validate` and live schema/relationship integration tests.
2. Run `schema.sql` under the chosen YDB TablePathPrefix; verify serializable transactions, YDB driver handling of `CurrentUtcTimestamp`, and conflict semantics against a real database.
3. **Stop all old unfenced projection writers**. A legacy writer not using guard preconditions can bypass graph fencing. Ensure service identities and permissions prevent arbitrary graph writes from other modules.
4. Enumerate all current grant/binding keys and tombstones from authoritative Access storage; `ReconcileKey` each. Do not publish readiness while pending keys remain or before the source watermark catches up.
5. Implement the Resource Manager source revision/tenant resolver and its event/snapshot projector with its own durable progress and deletion/ownership handling. Wire `consistency.Barrier` into the authenticated Access PDP path; the provided package does not automatically replace the existing public AuthZEN transport adapter.
6. Test two replicas contending for a single aggregate, lease loss during an in-flight SpiceDB write, stale graph writer, recovery after SpiceDB success/YDB failure, revocation and tenant/resource move, and full reconciliation after disaster recovery.
7. For a source revision changing after the authoritative read but before the owning service mutation, use the owning service's transactional tenant/owner/revision CAS and optional re-evaluation. A remote PDP check and local state mutation are not one distributed transaction.

## Unresolved rollout gates

- Live YDB + live SpiceDB end-to-end tests and migrations/backfill are required before production use.
- Kafka/Outbox concrete transport and trustworthy snapshot/discovery adapters are not included. The existing `eventfeed.Receiver` and `projection.SnapshotReader` ports need implementations.
- Reconciliation reads only graph namespaces listed in `m8.zed` v1. Any new resource kind needs schema and reader updates together.
- The Resource Manager watermark is currently an internal storage interface, **not** a proof that an upstream RM projection currently exists. Do not enable an allow-path until the source projector and authoritative revision resolver are wired and verified.
