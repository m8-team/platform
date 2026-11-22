# Architecture & Security Standard

**Status:** ACTIVE (for new/modified code; enforcement phased) · **Scope:** backend modules, integration boundaries, every API and Console surface handling tenant resources. See [index](README.md).

## A. Architecture and ownership

### ARCH-001 — Contract-first design (MUST)

External API/event behavior is defined in versioned, reviewed Protobuf contracts before implementations depend on it. Generated interfaces derive from these contracts; transport implementation does not invent incompatible semantics. Existing documented contract/runtime drift is treated as an implementation gap, not a reason to weaken contracts without review.

**Verification:** `.proto` diff, Buf lint/breaking report, handler-to-contract test and generated artifact diff.

### ARCH-002 — One authoritative writer per state (MUST)

Each business resource/state has one owning module responsible for authoritative writes, lifecycle and invariant validation. Other modules interact through published APIs or events; they MUST NOT modify the owning module's tables or hidden in-process state. Materialized read projections MAY exist, but are explicitly read-only replicas with owner-defined freshness/repair semantics.

**Examples of current intended boundaries:**

| Capability | Owner | Other modules must not own |
| --- | --- | --- |
| Organization/workspace hierarchy, resource descriptors, owner metadata | Resource Manager | Authorization decisions; module-specific business state |
| Permission definitions, grants, relationships, policy/model revisions, PDP evaluations | Access | External identities; org/workspace hierarchy |
| Principal and external identity lifecycle | Identity | Access policy and resource ownership |
| Actual workload rollout/provider reconciliation | Delivery/Runtime (where implemented) | Logical owner hierarchy |
| UI screens | Console | Authoritative authorization or persistence state |

These are *intended responsibility contracts*, not assertions that every module has a production implementation today. New owners MUST be documented in their module README or ADR. See [module-ownership.md](module-ownership.md) for the current authority registry with explicit planned vs implemented labels.

### ARCH-003 — Inward dependency direction (MUST)

Allowed conceptual dependency direction: `transport adapters -> application/use cases -> domain`; application-defined ports are implemented by infrastructure adapters. Domain MUST NOT import transport/gRPC/Protobuf, YDB/Kafka/Envoy/Kubernetes SDKs, React, or composition-root libraries. Keep each module's invariants local; put shared pure utilities in `internal/platform` only when genuinely cross-cutting.

**Recommended layout (adapt to current module conventions, not a mandatory rename):**

```text
internal/<module>/
  domain/          # entities, value objects, invariants
  app/             # use cases, commands, queries, ports
  adapter/         # gRPC/HTTP, persistence, external services
  module.go        # dependency wiring
cmd/<service>/     # configuration, lifecycle, Fx composition
```

### ARCH-004 — No implicit cross-module ownership (MUST)

New methods that mutate another module's resource MUST route to the owner. If a proposed change requires a different owner or shared database writes, create an ADR and migration strategy first. Do not create duplicate domain services merely to avoid an API call.

### ARCH-005 — Resource scope invariants (MUST for relevant Resource Manager paths)

Tenant-owned resources have exactly one valid owner scope according to the owning resource-kind contract (currently Organization or optional Workspace where defined), and one server-derived owning Organization where applicable. A Workspace is not always required. OrganizationGroup affiliation DOES NOT imply permissions. An Environment is not a hierarchy parent. Reject invalid owner combinations; verify the relevant scope against persisted authoritative state, not only client input.

**Existing drift — interim safety rule:** current `ResourceKindPolicy.allowed_owner_kinds` accepts `project`, but `OwnerScope` only supports Organization/Workspace and reserves `project_id`. Until [ADR-0001](adr/ADR-0001-resource-owner-scope.md) is accepted and its contract/runtime migration is implemented, **MUST NOT** introduce Project as an owner scope or register a policy that permits a placement the runtime cannot represent. Reject invalid Project placements without silently mapping Project to Workspace or weakening the public Protobuf contract.

### ARCH-006 — Consistency and asynchronous boundaries (MUST)

For cross-module workflows, identify the consistency model, event-driven completion, replay behavior, failure handling and owner of any projection. Do not imply a distributed transaction without evidence. Use transactional outbox/inbox or equivalent durable mechanisms where state and event publication need atomicity. Reconciliation MUST be idempotent and robust to retries and duplicate delivery.

### ARCH-007 — New module definition (MUST)

A new module proposal MUST state: business responsibility; owned entities and write boundaries; public APIs/events; consuming dependencies; tenant/resource scope; consistency model; authorization actions; migrations; observability; tests; failure behavior. Add a module README with these sections before treating the module as production-ready.

## B. Security architecture

### SEC-001 — Trusted tenant context and authorization (MUST)

Obtain tenant identity and principal context from verified authentication/authorization middleware and trusted service-to-service identity, not from untrusted request fields alone. Validate resource ownership, tenant scope, and action before returning or mutating protected state. Treat missing/ambiguous tenant or identity context as denial for protected operations.

Use request tenant IDs for *selection* only after checking that the authenticated principal may act on the chosen tenant. Never trust a tenant header forwarded directly by an external caller without an authenticated binding.

### SEC-002 — Defense in depth for PDP/PEP (MUST)

M8 Access provides policy decisions (AuthZEN-compatible APIs). Gateway MAY enforce coarse edge checks, rate limits and route policy; an application service remains responsible for its object-specific permission, business-state and ownership constraints. UI-based hiding is not authorization. Internal service-to-service calls require explicit identity and permissions appropriate to the operation.

**Deny-by-default policy:** Authorization failures or PDP unavailability MUST fail closed for critical reads/writes. A documented non-critical fail-open exception MAY be used only if its policy explicitly permits it, has a decision/audit reference, and does not expose sensitive data. No production `AllowAll`, `AllowUnauthenticated`, or equivalent bypass.

### SEC-003 — No cross-tenant leakage (MUST)

Enforce tenant scope at query, cache, event consumption and response mapping boundaries. Pagination, filtering and counts must not expose unauthorized objects. Never perform global list then paginate and filter authorization; apply eligible scope before pagination, or use an authorization-aware query plan with equivalent semantics.

Cache keys MUST include tenant and relevant authorization/model scope; invalidate or isolate on tenant switch, permission change and sensitive revocation. Avoid secrets or PII in telemetry or caches without approved storage controls. See `SEC-009` for decision consistency/freshness and `SEC-010` for mutation races.

### SEC-004 — Authentication and transport (MUST)

Verify token signature, audience, issuer, expiration and sender constraints as applicable; use supported OAuth/OIDC patterns (for example PKCE/DPoP when the client/flow requires them). Treat external headers and metadata as untrusted. Protect management and internal endpoints via authenticated mTLS or approved service identity when deployed. Never interpret a successfully authenticated caller as blanket authorization.

### SEC-005 — Secrets, logs, error responses (MUST)

No secrets, credentials, tokens, raw authorization context or unnecessary personal data in code, VCS, prompts, fixtures, public errors or logs. Logs should use structured, redacted fields (`trace_id`, `tenant_id` where safe, `resource_kind`, `outcome`) with appropriate access controls and retention. Return stable sanitized public errors; expose diagnostic details only to authorized observability paths.

### SEC-006 — Mutation integrity (MUST)

Protect mutations with input validation, relevant authorization, optimistic concurrency where required, idempotency, an auditable change record and consistent publication of domain events when applicable. An API field or annotation does not substitute for runtime enforcement. Update masks MUST be checked against allowed/immutable fields; validate the merged object after applying the mask.

### SEC-007 — Least privilege and change control (MUST)

Service identities and deployment credentials are scoped minimally. Actions changing authorization policy, tenant ownership, token verification, storage encryption, public network exposure, or secrets MUST receive human review. Changes to defaults toward permissive behavior require explicit security approval and tests.

### SEC-008 — Abuse resistance (SHOULD)

Set per-route/per-tenant resource limits, bounded payloads, pagination caps, request deadlines, and abuse-aware rate limits. Design retries with budgets and jitter; protect against retry amplification and unbounded fan-out. Add explicit threat analysis for security-sensitive public interfaces.

### SEC-009 — Authorization consistency and revocation (MUST for protected paths)

Every authorization check/cache MUST declare the decision inputs (principal, tenant, action, resource, context and policy/model revision where supported), consistency strategy, and maximum acceptable staleness/revocation window. There is **no universally safe cache TTL**: protected-data criticality determines whether decisions require fresh evaluation or validated revision-based caching. For sensitive reads, privilege/ownership changes and critical mutations, stale positive decisions MUST NOT outlive the approved revocation window. If the PDP cannot satisfy the required freshness, fail closed and surface a safe retryable error. Cache keys MUST distinguish principal, tenant, resource/action and applicable policy revision. Document tests for revocation, stale cache, PDP timeout and policy evolution.

### SEC-010 — Authorization check/use races (MUST for protected mutations)

An initial Gateway/application PDP result does not guarantee owner/tenant state is unchanged at mutation time. The owning service MUST bind authoritative tenant, owner, resource revision and security-relevant state preconditions to its write/transaction, rejecting or retrying on mismatch. When a remote PDP decision cannot join the database transaction, document the accepted decision lifetime, re-evaluation trigger and conditional-write guarantees. Do not assume a remote authorization check and later storage write are atomic. Test concurrent move/transfer, revocation and stale-revision scenarios.

## C. Concrete security decision flow

1. Authenticate caller and verify trusted principal/tenant context.
2. Resolve authoritative resource owner/tenant and business state (no client-derived ownership).
3. Ask Access/PDP for the relevant decision, including context/model revision as needed.
4. Apply domain and business invariants locally in the owning module.
5. Execute read/write with scope restrictions and concurrency/idempotency protections, including current authoritative owner/tenant/revision preconditions (`SEC-010`).
6. Record an appropriate audit event; return only authorized fields.

No step may be skipped because a gateway or frontend already made a similar check.

## D. Exception and verification policy

An exception to `ARCH-*` or `SEC-*` MUST follow [the approval model](approval-model.md) and identify: affected requirement ID, owner, business rationale, threat/impact, alternative controls, expiration or migration target, approving reviewer, and an ADR link. AI agents MUST NOT self-approve exceptions.

**High-value review tests:** unauthorized `Get/List/Update`; cross-tenant ID in request; scope changed after lookup; stale expected revision; replay same idempotency key with same/different payload; revocation during caching; PDP timeout; event replay; multi-tenant pagination; failure to write outbox/audit record.

**References in this repository:** `internal/access/README.md`, `internal/resourcemanager/README.md`, `proto/m8/platform/access/v1/`, `proto/m8/platform/resourcemanager/v1/`. For Resource Manager v1, inspect the actual handlers; a contract and README may describe unresolved runtime gaps.
