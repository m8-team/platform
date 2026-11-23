# Access Policy Consistency v2

**Scope:** trusted revision publication, contiguous grant/role-binding projection evidence
and two-owner PDP consistency checks. Source-of-truth Access persistence must
remain independent of SpiceDB. A positive AuthZEN decision requires BOTH
Resource Manager freshness and Access policy freshness.

## Authoritative journal contract

`Store.PublishPolicyMutationTx(ctx, tx, manifest)` MUST be invoked on the SAME
YDB serializable transaction as the authoritative Access state mutation and its
transactional Outbox append. Calling it as an unrelated standalone transaction
is unsafe and forbidden. It increments the tenant-wide `head_revision`, which
instantly makes the PDP fail closed until all projection jobs converge.

Each manifest contains:
- a unique per-tenant mutation ID and expected policy head (CAS);
- one job for the affected AccessGrant or RoleBinding, specifying its expected
  **projection** revision (not only domain object revision);
- on role permission updates, one job for **every affected RoleBinding**,
  including binds of differing resource scopes; no success while any is pending.

The owner repository must discover and persist the **complete fanout in the
same transaction**. The journal validates uniqueness and tenant identity, but
it cannot detect a RoleBinding omitted by an unimplemented owner repository.
That is a separate ownership/completeness release gate.

A new tenant has NO trusted policy state by default; PDP denies. The first
publication must be an audited `Bootstrap` manifest containing the complete
current inventory of effective grants and bindings, with a nonempty job set.
Empty inventory bootstrapping requires a separately reviewed owner attestation
workflow, not an implicit `revision=0` permit.

## Delivery and durable certification

1. Commit Access mutation, policy head, complete manifest and outbox in YDB.
2. Consume one job with its declared `policy_revision` and per-key revision.
3. Load the owning Access snapshot. Project to SpiceDB with the official SDK
   under the graph guard. Commit YDB checkpoint and ZedToken.
4. `CommitPolicyJob` uses the authoritative manifest and certified checkpoint;
   it rejects a missing job, stale checkpoint and an old fence epoch.
5. Advance `applied_revision` only across **contiguous** policy revisions
   whose complete fanout is certified. Do NOT combine opaque ZedTokens.
6. ACK Kafka/Outbox delivery only after step 5 commits. Replay is safe.
7. If a role has more than 1000 bindings, the transaction MUST fail closed
   pending a reviewed large-fanout protocol; silent truncation is forbidden.

## PDP policy

`consistency.PolicyBarrier` reads both `head_revision` and
`applied_revision` from one trusted YDB read. It blocks missing journal rows,
not-yet-applied revokes, partial role fanout and YDB failures. After Resource
Manager's separate revision check, SpiceDB receives `fullyConsistent`,
because YDB has distinct opaque graph tokens for RM and Access jobs.

For a positive graph decision, the barrier re-reads Access policy revision and
Resource Manager authority. A concurrent revoke or ownership change invalidates
the result. This establishes a check-time consistency guarantee, not a
cross-service transaction: target services MUST still enforce authoritative
state and resource revisions for their business mutations.

This v2 uses a **tenant-wide barrier**, which is conservative and may create
temporary denials during heavy policy write traffic. Consider audited
subject/resource key partitioning only after correctness and scaling evidence.

## Migration and acceptance gates

Apply `schema_policy_v2.sql` **after** the base Access schema, in a reviewed
YDB migration transaction/window. Do not run twice or use production
credentials for test jobs. In isolated GitHub CI, the integration test applies
the exact migration and verifies partial role fanout, out-of-order completion,
revoke lag, idempotency and checkpoint epochs.

Before rollout:
- implement authoritative Access grant/role/binding repositories with matching
  YDB transaction + Outbox; verify complete role fanout and bootstrap inventory;
- connect the real Kafka/Outbox consumer and verified OIDC/mTLS PrincipalResolver;
- demonstrate failure injection between SpiceDB write and YDB checkpoint,
  and during multi-replica takeover;
- verify tenant isolation, shadow PDP decisions, source/checkpoint/graph
  reconciliation, alerting and rollback; secure human security/ops approval.

Until those prerequisites exist this implementation is **not production ready**
and an uninitialized policy journal MUST remain fail-closed.
