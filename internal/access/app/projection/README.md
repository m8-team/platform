# M8 Access Projection Worker (v1)

The projection module consumes committed Access outbox/Kafka **projection event envelopes**, loads current authoritative snapshots via `SnapshotReader`, computes per-grant or per-role-binding graph deltas and writes those deltas using the official [authzed-go](https://github.com/authzed/authzed-go) SDK. The event feed package is a transport boundary: an application still must supply a Kafka consumer or an Outbox poller implementing `eventfeed.Receiver`.

## Delivery and checkpoint contract

- Required envelope fields: `event_id`, `tenant_id`, `projection_kind` (`ROLE_BINDING` or `ACCESS_GRANT`), `aggregate_id`, and monotonic per-key `projection_revision`. Role definition changes MUST fan out new projection revisions for **each affected role binding**.
- Never trust tuple bodies from broker messages. `SnapshotReader.LoadCurrent` must obtain an authoritative tenant-checked snapshot or tombstone from Access storage. The event's revision must match the source head, otherwise reconcile explicitly.
- A checkpoint contains `tenant/kind/aggregate`, `source revision`, `event_id`, graph `ZedToken` and the previously projected relationships (including expiry). An atomic SpiceDB write is acknowledged first; then the checkpoint is durably stored; **only then** acknowledge the Kafka offset/outbox row.
- A crash between SpiceDB write and checkpoint commit is replayed with idempotent TOUCH/DELETE. Reconciliation re-TOUCHes desired tuples and removes tuples recorded in checkpoint but no longer desired.
- The filecheckpoint adapter uses local fsync + per-key flock. **Single-host only.** It is not distributed HA, cannot fence cross-host writers, and must not be used with replicas sharing a network filesystem. Production multi-replica requires transactional/lease-fenced YDB (or equivalent) checkpoint store plus partition ownership and an audit/reconciliation strategy.
- `ReconcileKey` does not enumerate unknown graph tuples or enumerate all aggregate keys; a separately authorized discovery/reconciliation controller must do that before claiming full repair.
- ZedToken does not attest to upstream resource-manager revision. A PDP consistency barrier needs a trusted upstream projection watermark and must deny critical operations until the source revision has converged.
- This PR does **not** implement a live Kafka client, an Outbox poller, Resource Manager projection, or deploy/verify `schema/m8.zed`. These are separate wiring/integration steps.

## SDK transport

The adapter uses `authzed-go/v1` `PermissionsServiceClient` over TLS and bearer credentials by default. A plaintext connection is only allowed with explicit `AllowInsecureLoopback` and a loopback endpoint. TLS and schema revision pinning must be verified at deployment. Call `Client.Close` on shutdown.

## Checks

```bash
go test ./internal/access/...
go vet ./internal/access/...
go mod tidy  # ensure dependency graph and checksums before release
```

The local file-checkpoint package is Unix-specific because it uses `syscall.Flock`; CI on Linux is supported.
