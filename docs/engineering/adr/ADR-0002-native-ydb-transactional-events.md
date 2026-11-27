# ADR-0002 — Native YDB Transactional Events

- Status: Proposed implementation; architecture directed by the user
- Date: 2026-11-27
- Affected: shared event infrastructure and future module-by-module migration
- Requirements: ARCH-001, ARCH-002, ARCH-003, ARCH-006, DATA-002, DATA-003

## Decision

Use one versioned Protobuf Event envelope and a flat shared Go package.
YDB coordinates one native serializable read-write transaction containing
state, optional command receipt and required topic writes. The transaction
owner supplies the same transaction to the Publisher, propagates all
publication errors and owns commit outcomes. The SDK handles transactional
writer initialization, staging, flush, cleanup and supported retries.

Module owners define event types and typed constructors and own trusted
tenant and source context, business state, command idempotency and consumer
reconciliation. The shared layer validates the envelope and decoded payload,
enforces per-call limits, resolves configured topics and provides staging
counters. Logical identity, partition locality and cross-replica ordering
remain distinct.

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
restart and command receipt reconciliation. Require the Event Layer workflow
and a production-equivalent check of producer metadata capacity before
rollout. Server restart, active partition splitting and infrastructure
failover remain deployment verification gates. See the
[Event Layer README](../../../internal/platform/event/README.md) for the
test commands, operational constraints and per-module migration boundary.
