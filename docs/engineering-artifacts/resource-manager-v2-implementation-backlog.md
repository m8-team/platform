# Resource Manager v2 — implementation backlog for Codex

Reference: [Flexible Scopes RFC](../architecture/resource-manager-flexible-scopes-rfc.md).

Do **not** treat the RFC as an already deployed contract. Implement changes incrementally without silently altering v1 protobuf semantics. Follow repository AGENTS.md and the Resource Manager SPDD/requirements traceability conventions.

## P0 — contract alignment

1. **RM2-001: Resolve spec inconsistency.** Align the Requirements Catalog, SPDD, executable baseline and active protobuf APIs. Create a new API/event design review. Explicitly document v1 vs v2 contracts and compatibility.
2. **RM2-002: Scope model.** Expand the small pure-domain placement matrix proof of concept into registered per-resource-type ownership constraints. Add typed identifiers, ID parsing, checks for active parents and bounded ancestor traversal via app ports. Keep Access out of domain.
3. **RM2-003: Group and Organization affiliation.** Implement Group metadata and group_organization relationship; no automatic cross-organization authorization. Group is not a replacement for Company legal profile.
4. **RM2-004: Workspace and Project vertical slices.** Implement both against real storage and add Project direct-under-Organization and under-Workspace rules. Support Get, List, Create, Update, state changes; explicit MoveProject with target scope Access checks. Never move implicitly.
5. **RM2-005: Persist and enforce integrity.** YDB repository, scoped queries, parent/child references and delete blockers, optimistic concurrency, transactionally persisted outbox and idempotency. Avoid purely in-memory production adapters.

Acceptance: direct Organization->Project; Organization->Workspace->Project; existing Organization->Workspace->Service remains valid; no cross-tenant reads; invalid placements rejected; resource events published once logically under duplicate/retry conditions.

## P1 — resource ownership registry

6. **RM2-006: Generic resource metadata registration** (typed descriptor, not domain payload). Choose source-of-truth rules and typed reference naming convention. Implement register/update/unregister/query/path semantics.
7. **RM2-007: Module integration.** Agents, Workflow, Knowledge and Connections register an Organization/Workspace/Project owner without creating dummy containers. Per-kind policy restricts allowed owner types.
8. **RM2-008: Identity and Access integration.** Identity owns users, memberships, employment/departments; RM exposes parent/path resolver only. Access enforces per-resource checks and no implicit Group-role inheritance.
9. **RM2-009: UI/API.** Organization home shows direct resources; optional creation of Workspace/Project; Resource Browser displays canonical owner, breadcrumb and explicit sharing relationships.

Acceptance: create Agent directly at Organization; move Project only through reviewed command; company switcher does not change pinned workflow owner; Group browsing cannot leak member-company resources.

## P2 — service semantics and Shared Services

10. **RM2-010: ServiceRegistration v2** distinct from Business ServiceOffering. Keep v1 Service Workspace ownership and immutable environment in compatibility API until explicit migration.
11. **RM2-011: ServiceAgreement references.** Shared Services module owns agreements/cases; RM provides immutable Organization refs and lifecycle validation; Access owns delegated case-scoped grants.
12. **RM2-012: Migration/rollout.** Backfill typed owner references, dual read/controlled writer cutover, report orphaned resources, contract tests and access regressions; do not silently reparent anything.

Acceptance: Shared IT provider services Company A without becoming Company A admin; revoking agreement/assignment blocks new access; audit shows actor + provider + consumer + case + resource.

## Tests required per PR

* Table-driven domain tests and invalid-kind/parent cases.
* API contract and backwards compatibility tests, buf breaking check when proto changes.
* Persistent database integration tests (idempotency, concurrency, soft-delete, parent state).
* Security regression: Group viewer is not automatically Organization viewer; Company A user cannot retrieve Company B results.
* Event outbox consistency, retry, duplicate delivery and pagination authorization scope.
* No docs-only PR may claim that a phase is implemented. Attach exact command output for every test actually run.

## PR sequence

Architecture RFC -> typed scope model -> Organization persistence and child index -> Workspace + Project APIs -> Group associations -> generic resource metadata -> integrations -> service versioning -> delegated Shared Services.

No automatic deployment or merges as part of this backlog.
