# M8 Access Policy Consistency v3 — Universal Profile Entitlements

**Status: PROPOSED; security review required.** This increment adds a bounded, instance-level universal authorization graph, a tenant policy journal activation transaction, and a strict internal PDP barrier. It **does not** switch live AuthZEN handlers or grant access for broad hierarchical scopes.

## Source of truth and safe states

- Immutable, module-published resources/permissions/role templates stay in the Universal Catalog. Access owns role assignments and policy journal. The authoritative business module owns actual objects, tenant, existence, lifecycle and current revision.
- A staged `ProfileAssignment` remains `PENDING_PROJECTION` and projects **zero** relationships. Stage events are only administrative facts.
- `ACTIVE` may be reached only through `ydbroles.Store.ActivateWithPolicy` with an initialized, trusted `PolicyCommitter`; a single serializable YDB transaction updates assignment, subject read index, trusted source capability plan, request receipt, profile policy event and **tenant PolicyHead** with its exact projection job.
- Revoking an active assignment is a different method: `RevokeActiveWithPolicy`. It writes the state tombstone, empty capability source, policy event and incremented policy head in that same transaction. Existing simple `RevokeAssignment` is only for non-effective staged records and must fail for ACTIVE.
- The journal's existing `Bootstrap` path remains an explicitly audited operation. The activation writer refuses a zero/uninitialized policy head and never uses an implicit bootstrap.

## Universal SpiceDB model (instance scopes only)

`access_profile_assignment#holder@user` or `service_account`, unique per assignment. `access_capability#grantee@access_profile_assignment#holder`, where each capability object ID is the exact deterministic SHA-256 hash of (registered `resource_type`, concrete `instance_id`, `action`) plus tenant ID namespace. Graph permission `access_capability.use` represents **one concrete permission on one concrete resource**.

Independent assignments share the same capability object but have different holder nodes. Revoking one removes only its holder and grantee tuples, preserving other assignment grants. A full read of owned graph edges occurs after every YDB lease/fence generation (SpiceDB is not transactionally coupled to YDB). No direct bulk cross-tenant graph scan is permitted.

**No implicit hierarchy or wildcards.** Only `authorization.ScopeInstance` can become effective in v3. Collection, organization, subtree, resource-kind and tenant scopes reject activation with `ErrUnsupportedProfileScope`. Contextual policies (IP, MFA, dynamic attributes) also fail closed in v3 until their verified resolver and caveat/PDP contract exist. This is a deliberate restriction, not a missing default.

## Verified activation plan

`rolemanagement.Service.CompileActivationPlan` must be run by a trusted authorized administrator:

1. Validate authenticated tenant administration **before reading the assignment ID** and authorize the exact assigned subject.
2. Reload the pinned RoleBundle dependency graph and every pinned published RoleTemplate.
3. Re-resolve every resource's authoritative owner revision. Scope must be INSTANCE and match every permission's resource type and allowed scope.
4. Produce a bounded, duplicate-free set of `ProfileCapability` values with owner revision. Never source these from request bodies or Kafka payloads.
5. Pass the compiled plan through the internal activation API; the YDB transaction additionally validates the plan against assignment scopes.

Potential freshness races with the business owner's state are addressed by resolving the current resource state **again at PDP check time**. Owner source revision is not a token that Access can simply compare with SpiceDB's ZedToken.

## Consistency and event delivery

- PolicyJournal's projection kind `PROFILE_ASSIGNMENT` is distinct from `ROLE_BINDING` and `ACCESS_GRANT`. It has a dedicated YDB checkpoint with lease generation and SpiceDB graph-side epoch preconditions.
- The outbox `profile_policy_events` stores immutable policy revision and assignment revision. It must be delivered at least once and consumed by `policyjournal.Worker`. The worker certifies the full desired graph, checkpoint ZedToken, and policy job before advancing `Applied`. Out-of-order source revisions require reconciliation, not silent skipping.
- Before and after every positive internal `UniversalBarrier.Check`, the tenant PolicyHead must equal Applied, and the same head must be observed. It also verifies published permission, authoritative resource tenancy/existence and revision both before and after the fully consistent SpiceDB capability check. All errors and SpiceDB conditional checks fail closed.
- The exact application-level stage-to-active policy mutation is necessary because an authorization result cannot rely on the administrative CQRS state alone.
- `MarkPolicyPublished` is an outbox/broker ACK, not permission certification; only fenced checkpoint and policy job certification advance Applied.

## Database migration and rollback

Apply the additive `profile_assignment_projection` and `profile_policy_events` tables in `ydbroles/schema.sql` after planning the schema deployment. Existing Grant and RoleBinding tables and policies are unchanged. Update the SpiceDB schema to add `access_profile_assignment` and `access_capability` **before** starting profile workers, and verify graph schema compilation and ownership-filtered reconciliation in a disposable deployment.

Roll back by stopping profile activation APIs/dispatcher and keeping the tenant PolicyBarrier in fail-closed mode until all committed policy revisions converge or an audited compensation has completed. Do not drop policy events while Head > Applied.

## Status and remaining integration

- The code includes an application compiler, source snapshots, YDB policy transitions, fenced profile projection, generic SpiceDB check and a resource-and-policy freshness barrier with integration tests.
- **Not production-enabled:** public Protobuf activation/revoke APIs; service composition with verified mTLS/OIDC and resource owner adapters; Kafka dispatcher; enrollment/backfill of published resource kinds; multi-tenant/schema migration rollout.
- Next milestone adds independently versioned organization/subtree ownership graph and contextual policy evaluator (IP, MFA, risk, timestamps), plus production AuthZEN routing after security review.

Security acceptance: No permits before Applied reaches Head; overlapping independent assignment grants survive one revoke; a changed tenant policy head during evaluation blocks a stale allow; a changed/missing owner revision denies; two workers cannot write through a stale graph fence.
