# Backend & API Standard

**Status:** ACTIVE (for new/modified code; enforcement phased) · **Applies to:** Go application code, Protobuf source, ConnectRPC/gRPC/HTTP adapters, YDB integrations, events and background processing.

## A. Go implementation

### GO-001 — Formatting and language (MUST)

Use repository toolchain (`go.mod` currently declares Go 1.26.1) and Go idioms. Format Go with `gofmt`; use `goimports` where available. The generic `.editorconfig` spaces setting does **not** override Go's canonical tab-based formatting. Name exported declarations clearly; add GoDoc to public declarations when it provides useful contract semantics. Avoid cryptic names, giant files, cyclic imports and hidden global state.

### GO-002 — Boundaries and interfaces (MUST)

Construct dependencies explicitly. Define narrow interfaces at consumer boundaries when needed for testing/implementation alternatives; do not create one interface per struct by habit. Use Uber Fx in module/composition-root wiring, not inside domain code. Keep constructors deterministic, validate configuration and surface startup errors; no silent fallback from production adapters to in-memory data.

### GO-003 — Errors (MUST)

Return errors, not panics, for normal failure modes. Use `errors.Is/As` and contextual `%w` wrapping; map domain errors to stable gRPC/Connect or HTTP status in the transport adapter. Do not include credentials, internal query text or tenant-sensitive data in client errors. Preserve cancellation and deadline errors, and do not incorrectly map them to internal server faults.

### GO-004 — Context, concurrency and time (MUST)

Pass `context.Context` as first parameter for operations that block, call remote services or handle requests. Propagate cancellation and deadlines to DB, network and async boundaries. Avoid storing request contexts in long-lived structs. Bound goroutines, queues, parallel fan-out and retries; ensure goroutine lifecycle and shutdown. Use injected clocks when business rules depend on time; store canonical timestamps in UTC and format only at presentation boundaries.

### GO-005 — Data structures and correctness (MUST)

Use explicit types for identifiers, revisions and money/decimal values when ambiguity or precision matters. Do not use floating-point values for financial invariants. Validate all boundary input, including nil pointers, UUID formats where contracts require, enum unknown values, masks, list limits and business constraints. Make state transitions explicit and test invalid transitions.

### GO-006 — Configuration and dependencies (SHOULD)

Keep config validated, typed, environment-backed and documented. Pin or intentionally manage dependency upgrades; avoid introducing a second library for an existing concern. Disable insecure/local-only behavior by default in production configuration. Never hard-code secrets or production hostnames.

## B. Protobuf and externally consumed API

### API-001 — Contract-first and naming (MUST)

Define public APIs in `proto/m8/platform/<module>/vN/` with packages like `m8.platform.<module>.vN`. Use existing `go_package`, Buf, HTTP and OpenAPI conventions. Follow Google API Improvement Proposals where they fit, while preserving compatibility with established M8 contract conventions. Don't rename already published RPCs, routes, fields or enum values just for style.

### API-002 — Compatibility discipline (MUST)

Do not renumber or reuse Protobuf fields. Treat changes to request/response semantics, JSON field names, `oneof`s, resource identity, error status, sorting, pagination, event meaning or HTTP routes as possibly breaking even if Buf's wire-level checks pass. Reserve removed numbers and names. Breaking migrations require explicit approval, version transition, consumer impact analysis and tests. `buf breaking` is necessary but not sufficient.

### API-003 — Validation at both layers (MUST)

Mark required/optional/immutable/output-only fields appropriately, and define structural constraints with Buf Protovalidate where useful. Handlers MUST invoke runtime validation (directly or via verified middleware), enforce semantic invariants and sanitize server-managed fields. `field_behavior` annotations and validation tests against messages do not prove runtime checks. Validate full merged state after FieldMask updates.

### API-004 — Mutations and concurrency (MUST)

Create/update/delete/state-transition APIs MUST state their idempotency semantics, response on identical retry, conflict behavior for different payload with same key, and any expected revision/etag. For relevant M8 resource mutations, require an idempotency key and optimistic concurrency as defined by the API contract; implement enforcement in the use case and storage boundary. Do not accept a field then ignore it.

**Recommended request semantics:**

```text
UpdateResourceRequest
  id                 immutable resource identity
  resource           proposed field values
  update_mask        explicit writable paths
  expected_revision  compare-and-swap precondition
  idempotency_key    retry deduplication identifier
```

### API-005 — Read/list/search semantics (MUST)

Specify page size cap, ordering stability, filters, tenant/resource scope, authentication and authorization, cursor integrity and expiry. Do not leak unauthorized entries through results, totals or cursors. Define empty-set behavior and distinguish `NotFound` from `PermissionDenied` according to the resource-disclosure security policy.

### API-009 — Representable resource placement (MUST)

Registration policies MUST NOT advertise owner kinds that the current ownership contract/runtime cannot represent. Validate `allowed_owner_kinds` against supported `OwnerScope` alternatives; reject impossible placements rather than inventing Project-to-Workspace mapping. Existing `project` policy/oneof drift is tracked by [ADR-0001](adr/ADR-0001-resource-owner-scope.md). An actual Protobuf change requires the accepted ADR, compatibility review and explicit migration — it is not authorized by this documentation standard.

### API-006 — Long-running operations (MUST when applicable)

If an operation cannot reasonably complete in a request deadline, use the project's versioned long-running-operation contract/pattern; include operation identity, states, progress, error, cancellation, retry and idempotency behavior. Do not block RPC handlers on unbounded infrastructure work.

### API-007 — Error model (MUST)

Map errors consistently, e.g. invalid input to `InvalidArgument`, missing precondition to `FailedPrecondition`, optimistic conflict to `Aborted`, unauthorized to `PermissionDenied`/`Unauthenticated`, and dependency unavailability to `Unavailable` as appropriate. Preserve actionable metadata without leaking internals. Test error mapping.

### API-008 — Generation and review (MUST)

Use `buf.yaml` and `buf.gen.yaml`. The repository `justfile` currently provides `proto`/`proto-deps`; running `just proto` may update dependency lock/exported files and generated code, so inspect the diff. Never patch `internal/proto` or `docs/openapi` directly. Check generated client/API drift and build all affected consumers.

## C. Data and events

### DATA-001 — Ownership and storage (MUST)

The data owner controls schema, migrations, indexes, retention, transaction boundaries and event contracts. Database operations MUST include tenant/scope constraints where relevant. Service code MUST NOT write another module's tables. YDB access should be isolated in repositories/adapters; domain must not know YDB SDK types.

### DATA-002 — Transactionality and outbox (MUST when publishing follows mutation)

When business state and a domain event must be committed atomically, write them in one supported transactional boundary or use an equivalent durable change log. Outbox publishing is retryable. Consumers must safely handle duplicates and reordering as documented. Do not claim 'exactly once' across network boundaries without a concrete guarantee; design observable idempotent outcomes.

### DATA-003 — Migrations and rollback (MUST)

Schema changes MUST document backward/forward compatibility, online migration strategy, data backfill, old/new consumer overlap, validation, cleanup and rollback/roll-forward strategy. Avoid destructive migrations in the same release that introduces a new reader. Large destructive migrations require explicit approval.

### DATA-004 — Replay, retries, deduplication (MUST)

Event consumers MUST track deduplication identity, version/schema, ordering scope and at-least-once retry behavior. Document poisoned-record treatment, DLQ or quarantine and operator replay. Backfills MUST be independently resumable and must not corrupt live consumer offsets or create double side effects.

### DATA-005 — Reconciliation and eventual consistency (SHOULD)

Cross-system projections and Kubernetes-style controllers SHOULD be level-triggered and idempotent, separate desired from observed state, expose convergence/error conditions, and recover from missing/duplicated/out-of-order events. Reconciliation is not a substitute for source-of-truth ownership.

## D. Endpoint completion checklist

For any new/changed RPC, reviewers MUST be able to identify: (1) owning module, (2) Protobuf contract and HTTP mapping, (3) boundary/runtime validation, (4) authn/authz and tenant resolution, (5) use case/domain invariant, (6) storage transaction/revision behavior, (7) idempotency/retries, (8) error semantics, (9) trace/audit expectations, and (10) positive/negative/regression tests. If an item is not applicable, state why.

**M8-specific known risks:** `internal/resourcemanager/README.md` reports runtime gaps around Protovalidate invocation, FieldMask/merged state, expected revision/idempotency, and generated API drift. Agents MUST verify the relevant implementation rather than assuming these are fixed.
