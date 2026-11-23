# Universal Access Runtime Integration v1

**Status: opt-in runtime composition proposed. Not enabled in a production deployment.**

This increment follows Policy Consistency v3 and routes versioned module-defined
resource kinds through the real AuthZEN interface, without changing the existing
legacy fixed-kind access checks or weakening tenant Policy/Resource barriers.

## Runtime routing

`wiring.NewProductionUniversalPDP` composes:

1. **Verified AuthZEN PrincipalResolver** from OIDC/mTLS middleware. The request's
   subject must match the authenticated principal; delegation remains denied.
2. **Legacy PolicyBarrier** (resource owner, YDB policy journal, SpiceDB), unchanged.
3. **UniversalPermissionReader** with strong tenant installation lookup and an
   immutable pinned catalog release. A global catalog publication alone never
   installs a module for any tenant.
4. **UniversalBarrier** requiring current tenant PolicyHead == Applied, module
   permission, authoritative owner existence/revision, fully consistent SpiceDB
   capability check, and post-check policy/owner freshness.
5. **RestrictedUniversalBarrier** additionally requiring an exhaustive strong
   restriction inventory for the exact tenant/type/instance/action and its
   matching PolicyHead; trusted IP, MFA and resource attributes with the owner
   revision; enforcement of every relevant REQUIRE/DENY_WHEN condition; and
   post-check PolicyHead, owner and module rechecks.

`authzen.NewWithUniversal` dispatches namespaced resource kinds such as
`finance.invoice` to the restricted universal PDP. Fixed legacy kinds like
`workspace` remain on the original path. A dotted resource kind **never**
falls back to the legacy PDP if the optional universal route is absent. Client
supplied AuthZEN context and properties continue to be rejected.

An unknown module action, uninstalled or disabled module, missing tenant module
version, incomplete restriction inventory, stale source revision, unverifiable
MFA/client IP or resource attributes, or graph conditional result cannot ALLOW.

## Tenant module installation schema

Migration `internal/access/adapters/ydbcatalog/schema_tenant_modules_v1.sql`
adds `catalog_tenant_modules(tenant_id,module_id,release_version,enabled,revision)`.
Its reader is implemented by `ydbcatalog.Store` and rejects missing, disabled,
unversioned and corrupt rows. **No public installation writer is included.**
Only a separately authorized module lifecycle owner is permitted to populate
this table via CAS and a reviewed operational migration. Activation cannot
be considered production-ready until module installation ownership is wired.

## Trusted context

`adapters/trustedcontext.Source` requires three independent, authenticated
provider ports: VerifiedClientIP (gateway-established peer after proxy trust),
VerifiedMFA (IdP/session assurance), VerifiedResourceAttributes (owner revision
and bounded attributes). It uses a server clock and refuses malformed or
missing evidence. No `X-Forwarded-For`, request-body MFA or user-defined
resource properties are trusted.

The **RestrictionSource** must return a *complete* policy snapshot with
`PolicyHead` equal to the durable Access journal. It must also guarantee
restrictions are written as authorized, versioned changes in the same
authoritative policy revision lifecycle. This PR defines and tests its
read-contract; it **does not** introduce a public policy authoring API or
durable restriction-snapshot writer. Wiring requires an external trusted
implementation. Only INSTANCE-scoped mandatory policies are evaluated; broad
or unverifiable scopes fail closed, never silently skipped.

## Durable profile-policy event delivery

An additive `ydbroles/schema_policy_delivery_v1.sql` migration creates
`profile_policy_dispatch`, ordered per tenant by PolicyRevision, populated
in the SAME YDB transaction as assignment/PolicyHead/source/outbox mutation.

`ydbroles.PolicyDispatcher` reads bounded pending events, calls an injected
authenticated `PolicyEventPublisher` and marks both durable outbox/index rows
published **only after broker ACK**. Broker failure leaves the event pending,
and ambiguity may cause at-least-once duplicates. Consumer MUST dedupe
EventID/PolicyRevision and acknowledge only after a certified projection
checkpoint and the PolicyJournal applied barrier. Parallel dispatchers must
use a separately configured tenant lease; no delivery exactly-once promise.

This PR intentionally does not hardcode or install a Kafka client SDK, topic,
consumer group, credentials or broker connection. The publisher is a required
integration port; actual production broker wiring remains deployment work.

## Rollout and rollback

Deployment gate:

- Apply existing Universal Catalog/Role/Policy v1–v3 schema, then tenant-module
  and policy-delivery additive migrations with old profile policy writers
  stopped. Populate tenant installation state only through authorized owners.
- Deploy verified identity, tenant module lifecycle, owner resolver, a complete
  policy snapshot provider, trusted IP/MFA/attribute resolvers, SpiceDB and
  YDB checkpoints. All required dependencies are constructor-validated.
- Connect a real broker publisher; configure tenant-dispatcher lease/retry,
  dead-letter telemetry, Kafka consumer and crash/reconcile management.
- Enable the new route only for approved tenants after schema verification,
  audit and security review. No automatic migration or fail-open fallback.
- Roll back by disabling universal route and dispatcher; leave PolicyBarrier
  active, retain YDB outbox and policy journal, and reconcile pending revisions
  rather than dropping events.

## Tests

- Catalog exact module version/action/type lookup and invalid namespace regex.
- Restricted universal decision with verified IP/MFA/attribute and full policy
  snapshot; missing/wrong context, incomplete inventory or stale head denies.
- AuthZEN dispatch isolation, authenticated principal, rejected untrusted
  attributes and no legacy fallback.
- Real disposable YDB: tenant module pin, missing installation and disable;
  profile policy journal graph activation plus transactional broker failure
  and idempotent dispatch.
- Existing real SpiceDB tests for independently revocable profile grants and
  graph fencing remain in the CI workflow.

## Remaining milestones

A production-authoritative module installation command/CRD owner, versioned
mandatory policy CRUD and restriction query projection, broker SDK adapter,
authenticated runtime deployment config, and multiple-replica failover/load
tests are still required. Hierarchical ORGANIZATION/SUBTREE/TENANT scopes and
contextual caveat compilation are **not** implicitly enabled by this PR.
