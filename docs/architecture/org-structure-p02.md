# OrgUnit P0.2 — revisioned ChangeSet and storage contract

**Status:** Draft; depends on PR #20. Package is `m8.platform.resourcemanager.v1`. No legacy API compatibility.

## Implemented

- Pure `orgunit.Apply` for CREATE/MOVE/RENAME/CLOSE, final-graph validation, stable IDs, source authority, optimistic revision.
- Authorized application boundary `orgstructure.Service` with caller-derived actor, idempotent request hash, mandatory Access/Impact policy gates and CAS store contract.
- In-memory adapter with atomic commit of snapshot/revision/idempotency/outbox effect and concurrency tests. The in-memory adapter is for contract testing only.
- A YDB DDL schema for immutable staged snapshots, active head pointer, closure table, change sets, and bitemporal OrgUnitVersion fields. Common operational outbox/idempotency already exist in 001_init.sql.
- Protobuf ChangeSet validate/apply source, not a registered endpoint.

## Correct YDB publish protocol (YDB implementation not yet present)

1. Read current `org_structure_heads` and the immutable rows under its `active_snapshot_id`, with tenant and organization authorization.
2. Validate source ownership, impact on Access, employees and ResourceOwnership and an approved ChangeSet. The impact-check port must fail closed if unavailable.
3. Stage the complete next adjacency and closure rows under a random unique `snapshot_id`; check counts, cycles, parent references and depth-0 self-edges. Competing builders use distinct IDs.
4. In one **serializable read-write transaction**, read the head and compare its revision AND active snapshot ID to those initially observed. Atomically persist new head, ChangeSet result, history publication record, idempotency response and an event in the existing outbox. On compare failure leave current head untouched and GC abandoned staged rows later.
5. Reader obtains active head first and queries only that snapshot ID. Version-history visibility must be tied to a published revision, not a half-written batch. The old snapshot is retained while readers/cursors or audit retention require it.
6. An unchanged structure preserves the revision and emits no resource-change event. A distinct plan ID with stale expected revision fails.

## Important limits

- DDL is **not** a running YDB repository; no SDK adapter, live migration execution or live DB integration tests have been performed.
- `effective_at` must equal server-recorded time for the current Go write slice. Scheduling and corrections backdated from 1C:ZUP require temporal history materialization; never fake a past revision.
- Cross-module impact checks and approved-ChangeSet verification have interfaces, not production implementations. No permissive policy may be deployed.
- Protobuf requires buf lint/generate, gRPC handler wiring and full CI verification.
- Identity/Directory still owns Person, Membership, OrgUnitAssignment and Team; the Closure Table contains OrgUnit-to-OrgUnit edges only.
