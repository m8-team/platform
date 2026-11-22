# Quality & Testing Standard

**Status:** ACTIVE (for new/modified code; enforcement phased) · **Scope:** all changed production, test, generated and configuration files.

## A. Definition of Done

### TEST-001 — Evidence-based completion (MUST)

A task is complete only when acceptance criteria are met and the applicable checks were run and reported. Tests MUST exercise behavioral risk rather than simply line coverage. A passing build alone is not a test of authorization, tenant isolation, idempotency, or business invariants. Never say “all tests pass” if only scoped tests ran, and never conceal failures or skips.

### TEST-002 — Test levels (MUST where applicable)

| Level | Purpose | Representative examples |
| --- | --- | --- |
| Domain unit | Business invariants without infrastructure | Resource owner constraints, state transitions, decimal math |
| Application unit | Use case orchestration, errors, ports | Authorization before write, retries, expected revision |
| Adapter integration | Actual storage/protocol behavior | YDB transactions, ConnectRPC validation, generated codecs |
| Contract | External compatibility | Buf lint/breaking, HTTP mappings, field behaviors |
| Security negative | Denial and isolation | Cross-tenant Get/List/Update, PDP timeout/fail-closed |
| Reliability/race | Concurrency and recovery | Duplicate event, concurrent update, cancellation, worker restart |
| Frontend behavior | User-visible interactions | Loading/empty/error/forbidden, forms, tenant switch |
| E2E/smoke | Main path across components | Authenticated resource create/read/update flow |

Where a test type is not supported by infrastructure, note the gap and the proposed follow-up; do not substitute a fabricated passing result.

### TEST-003 — Bug fix regression (MUST)

For bug fixes, demonstrate the failure with a regression test where practical; then fix implementation and re-run the test. Include negative and boundary conditions. Do not weaken expectations or delete tests to mask defects.

### TEST-004 — Determinism (MUST)

Tests must not depend on uncontrolled wall clock, random IDs, global mutable state, execution order, internet access or real credentials. Prefer injected clocks/ID generators, isolated test tenants, deterministic event order and reusable fixtures. Integration tests must have an explicit opt-in environment or isolated CI service.

### TEST-005 — Multi-tenant and security coverage (MUST for protected paths)

For any protected resource path, test allowed and denied subject, missing context, cross-tenant ID, unauthorized list/pagination, service-to-service identity, revocation or PDP failure where relevant. A UI conditional alone does not satisfy this requirement.

### TEST-006 — Stateful mutation coverage (MUST when relevant)

Test first mutation, identical idempotent retry, conflicting replay, stale revision, invalid FieldMask, output-only update, partial merged-state validation, persistence error, and event/audit delivery. Verify consistency across retry/restart and failure windows.

### TEST-007 — Code quality and scope (SHOULD)

Keep functions, packages and components small enough to review; avoid incidental refactors. Prefer table-driven Go tests with explicit cases; use `t.Parallel()` only when state is safely isolated. Add race checks where concurrency changes. Mock only at a meaningful port boundary; prefer integration test for actual SQL/YDB query and transaction correctness.

## B. Current available commands (checked against existing files)

Run from repository root unless indicated:

```bash
# Go formatting check (read-only; nonzero exit if unformatted):
unformatted="$(find cmd internal -name '*.go' -type f -print0 | xargs -0 -r gofmt -l)"
if [ -n "$unformatted" ]; then
  printf 'Unformatted Go files:\n%s\n' "$unformatted" >&2
  exit 1
fi

# Go analysis and tests:
go vet ./...
go test ./...
go test -race ./internal/...  # when environment/toolchain supports race

# Buf — use pinned .bin/buf if installed via justfile:
just install-buf
.bin/buf lint
.bin/buf breaking --against '.git#branch=main'

# Frontend:
cd ui/console
pnpm install --frozen-lockfile
pnpm lint
pnpm build
```

Caveats: The Buf breaking command needs a local Git baseline and dependencies; it may fail during an approved contract transition. Installing dependencies may require network access. `pnpm install` changes local installation state but should not rewrite the lockfile with `--frozen-lockfile`. Do not treat commands as executed merely because they appear in this document.

## C. Phased CI adoption (limited gates implemented)

### TEST-008 — Continuous integration baseline (SHOULD, adoption milestone)

Phase 1 adds `.github/workflows/engineering-standards.yml` for policy documentation references/IDs and **changed Go files only** formatting (so unrelated legacy code is not made a new blocking failure). The workflow does not become a required merge gate until administrators configure branch protection. Next phases SHOULD add independent jobs:

1. **Go format, vet, compile and tests**; optional scoped race run based on time/resources.
2. **Buf lint, breaking and generated-code drift** against a known `main` baseline.
3. **Console lint, TypeScript compile, Vite build** with frozen dependency resolution.
4. **Security/dependency scan**, secrets check, license/policy review when configured.
5. **Architecture checks** for forbidden imports, module-level dependency direction and generated-code edits.

Do not introduce blanket required checks until they pass on an agreed baseline, or explicitly track grandfathered failures. Do not change Buf's current lint exceptions or breaking policy without their own decision and migration.

### TEST-009 — Generated artifacts (MUST)

If any `.proto` changes, generate affected Go/Connect/gRPC/gateway/OpenAPI artifacts using repository configuration, compare diffs, and explain why generated outputs changed. If generation cannot complete, mark the PR as not fully verified. Generated code MUST NOT be manually patched.

### TEST-010 — Contract and migration review (MUST)

A public API or data schema change requires a compatibility matrix of old/new producer and consumer; rollback/roll-forward path; effects on versioning, permissions, event replay and existing data. Test backwards compatibility when possible. Approved breaking changes MUST have a recorded release plan.

### TEST-011 — Quality gates are not self-certification (MUST)

Only a human reviewer or CI/release policy may approve exceptions. An AI agent cannot mark unverified test results green, edit thresholds to sidestep failures, or self-merge a protected change.

## D. PR quality report template

```text
Scope / goal:
Files changed:
Contracts, modules and resource owner affected:
Security / tenancy implications:
Data migrations / rollout / rollback:
Tests run and results (exact commands):
Tests not run and why:
Known residual risks or follow-up:
Required human approvals:
```

## E. Severity and review escalation

- **Blocker:** cross-tenant access, authorization bypass, leaked credentials, destructive data loss, critical contract break without approval, unsafe production defaults.
- **High:** unhandled failed idempotency, missing concurrency enforcement, revocation-freshness violations or ownership TOCTOU, incompatible migrations, test suite regressions, missing audit for critical mutations.
- **Medium:** unbounded retries, missing cancellation, significant accessibility regressions, weak error handling.
- **Low:** style, small refactors, documentation clarity when no correctness/security impact.

Do not mislabel high severity findings as style. Every blocker must have an explicit fix or approved exception before release.
