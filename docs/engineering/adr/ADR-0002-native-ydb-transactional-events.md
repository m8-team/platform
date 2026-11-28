# ADR-0002 — Native YDB Transactional Events

- Status: Proposed implementation; architecture directed by the user
- Date: 2026-11-27
- Affected: shared event infrastructure and future module-by-module migration
- Requirements: ARCH-001, ARCH-002, ARCH-003, ARCH-006, DATA-002, DATA-003

## Decision

Use one versioned Protobuf Event envelope and a flat shared Go package.
YDB coordinates one native serializable read-write transaction containing
state, optional command receipt and required topic writes. The transaction
owner runs the transaction through the protected Publisher. The Publisher
supplies the business callback with a scoped actor and publication function,
and aborts an attempt after a publication failure. The SDK handles transactional
writer initialization, staging, flush, cleanup and supported retries.

Module owners define event types and typed constructors and own trusted
tenant and source context, business state, command idempotency and consumer
reconciliation. The shared layer validates the envelope and decoded payload,
enforces per-call limits, resolves configured topics and provides staging
counters. Logical identity, partition locality and cross-replica ordering
remain distinct.

## Stable Producer Identity decision

Stable, reusable producer identities use the unified Publisher API.
The pinned SDK/server pair allows two independent transactions using the
same Producer ID to commit business state while YDB suppresses one of the
events. The regression holds a sequenced write before send, commits a
competing write, then releases the delayed write. Both state rows exist
and only the competing event is present.

Kubernetes ordinal stability alone cannot establish ownership. The
publisher holds a single local lane from before opening the transaction to
after SDK attempts and writer completion. It checks ownership inside every
attempt and advances a separate row keyed by the replica. This row makes
unresolved attempts across same-Pod process restarts conflict. In the event
of a conflict the losing attempt rolls back, and SDK retry initializes a
fresh writer. Only the Producer ID is reused. The SDK owns its partition
suffix and sequence initialization. Use one nonempty batch per physical
topic per attempt.

A StatefulSet preserves ordinal exclusivity by waiting for confirmed
termination, which can block replacement during a network partition. A
Kubernetes Lease or a YDB Coordination Node semaphore alone cannot fence
requests already in flight to YDB. A live experiment verifies that a
privileged generation change invalidates a stale publication transaction
when that transaction reads the generation under serializable isolation.
The fixture verifies state and delivery after the stale attempt aborts and
the replacement reuses the same Producer ID. The handoff implementation
binds the generation to the Pod UID returned by an authenticated Kubernetes
API read for a pinned StatefulSet incarnation. It compares that observation
against a serializable ownership row and keeps the publisher on a read-only
credential for that row. The live Kubernetes fixture exercises node
partition and Pod replacement, then both kinds of uncertainty around a
same-Pod process restart. The reconciler uses a named StatefulSet and Pod GET
with authenticated TLS and a separate YDB ownership writer. It reads the
ownership generation before observing Kubernetes under a bounded deadline,
so SDK retries obtain a fresh observation.

A missing or deleting Pod clears its active binding without removing
generation history. A deleted, deleting or replaced StatefulSet permanently
retires the old incarnation's rows. Provisioning supplies predecessor UIDs
and their maximum authorized ordinal ranges, and the reconciler retires
those slots before assigning the new incarnation. The old ownership read
conflicts with retirement even though a recreated StatefulSet uses a new
Producer ID. Publisher credentials cannot change the ownership table.
Separate ownership and guard tables preserve distinct ACLs. See the
[deployment and recovery runbook](../../../deploy/event/README.md).

See the [v1.1 verification](../../../internal/platform/event/README.md#v11-stable-producer-identity-verification)
for the counterexample, sequential reuse, metadata measurements, rollout
conditions and unverified scenarios.

## Alternatives

- **Native keyed MultiWriter** routes an aggregate within configured partition
  bounds and supports batches spanning partitions. It opens a stream for each
  touched partition and creates producer IDs per transaction attempt.
- **Native Single Writer** has one stream initialization per writer. It can
  support a fixed-partition workload, while multi-partition routing would need
  additional coordination or an explicit partition policy. The benchmark
  measures the end-to-end cost of both approaches.
- **Native CDC** exposes storage changes and has its own schemas and lifecycle.
  Business event identity and semantics would still require owner-defined
  mapping, and CDC cannot express all domain facts from a row change.
- **Traditional outbox** supports systems without native topics and existing
  cutover paths, at the cost of relay, checkpoints, cleanup and reconciliation
  infrastructure. Existing module outboxes remain in place.

## Compatibility and release evidence

The tested server image is local YDB 26.3.1.25 with Go SDK v3.157.1. Provision
keyed topics with partition bounds. The SDK API is experimental. The wire
fields and their numbers are preserved. The contract narrows UUID spelling
and metadata size. The routing key format changes to length-prefixed identity
bytes passed to the SDK hasher. Check for external writers using the old key
before a routing change on an existing stream.

Live tests cover table/topic commit and rollback, offset visibility during
staging, multiple topics, 1/4/16 partitions, independent replicas, reader
restart and command receipt reconciliation. The Event Layer CI also exercises
load-induced partition splitting, recovery after a hard server restart and
loss of a tablet-hosting process in a replicated test cluster. The metadata
soak directly samples retained source records from the YDB partition tablet.
A separate live test confirms TTL cleanup below the partition start offset
and retention of records whose offsets still exist. Size the hottest
partition using observed attempt rate, source TTL and message retention, and
verify the actual production failure domains and recovery budget before
enabling writes. See the [Event Layer README](../../../internal/platform/event/README.md)
for commands, scope of evidence and migration constraints.

### v1.2.1 transaction and deployment boundary

`Publisher.WithinTransaction` passes an attempt-local `*event.Transaction`
with typed query parameters, materialized single-row reads, publication and
bounded native transactional batch reading. The native transaction stays in
the Publisher and remains under SDK lifecycle control. The scope invalidates
before completion. A native SDK negative control committed a state write via
`query.WithCommit()` despite a later callback error. The corresponding
protected YQL test rejected `COMMIT`. The boundary assumes trusted code and
driver defaults. Business modules still own event obligation, command receipts,
authorization and deterministic identities.

Fencing paths derive from the connected YDB database name. The ownership
table also holds a controller activation row. Replacement activation verifies
the known slots against actual stored ownership and requires the prior
activation's retirement. The same row fences positive slot writes from old
reconciler configurations. Upgrade requires a verified canonical inventory and
fencing binaries that predate this barrier. Readiness inspects both schemas and
does not stage events.

Pilot and production conditions are documented in the [Event Layer README](../../../internal/platform/event/README.md)
and [deployment runbook](../../../deploy/event/README.md). Keep target-cluster
telemetry, protected split, independent failure domains and coordinated
backup/restore as explicit gates.
