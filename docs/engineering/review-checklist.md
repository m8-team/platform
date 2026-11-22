# M8 Engineering Review & Acceptance Checklist

**Status:** PROPOSED. Use as PR template or AI final verification report. Mark N/A with reason; do not check items without evidence.

## Context and scope

- [ ] User story/bug and acceptance criteria are clear.
- [ ] Owning module, authoritative state and affected resources are identified (`ARCH-002`).
- [ ] Diff is minimal and does not overwrite unrelated work (`AI-005`).
- [ ] Existing README/ADR/contract/runtime discrepancies are recorded, not silently normalized (`AI-003`).

## Architecture & security

- [ ] Module dependency direction and public API boundaries are preserved (`ARCH-003`).
- [ ] No cross-module writes or duplicate authoritative owners (`ARCH-002`).
- [ ] Tenant scope and object-level authorization are enforced server-side (`SEC-001`, `SEC-002`).
- [ ] Denied/missing identity, cross-tenant, revocation and failure paths are covered (`SEC-003`).
- [ ] No production bypass, sensitive logs or exposed secrets (`SEC-002`, `SEC-005`).
- [ ] Audit trail and mutation integrity are addressed (`SEC-006`).

## Backend & API

- [ ] Contract-first API and published event semantics preserved (`API-001`, `API-002`).
- [ ] Runtime validation and merged FieldMask semantics implemented (`API-003`).
- [ ] Idempotency and revision handling implemented and tested (`API-004`).
- [ ] Pagination/filtering and resource isolation correct (`API-005`).
- [ ] Errors, deadlines, cancellation and config validation are correct (`GO-003`, `GO-004`).
- [ ] YDB schema/transactions and event/outbox consistency addressed (`DATA-001`, `DATA-002`).
- [ ] Migration, replay, rollback plan present if needed (`DATA-003`, `DATA-004`).
- [ ] Generated code was regenerated from source, not edited (`API-008`).

## Quality & testing

- [ ] New behavior has relevant unit/integration/contract tests (`TEST-002`).
- [ ] Bug fixes include regression coverage where practicable (`TEST-003`).
- [ ] Security and tenant-isolation negative cases covered (`TEST-005`).
- [ ] Retry/stale revision/partial update cases covered (`TEST-006`).
- [ ] Go formatter/vet/tests completed or explicitly marked blocked (`TEST-001`).
- [ ] Buf lint/breaking/generation checks completed if `.proto` changed (`TEST-009`).
- [ ] Frontend lint/build completed if Console changed (`TEST-001`).
- [ ] Final results list exact checks, failures and unverified items (`AI-007`, `AI-012`).

## Frontend & UX

- [ ] Existing Gravity UI/React patterns and feature boundaries are followed (`FE-001`, `FE-002`).
- [ ] API types, tenant-scoped query keys and stale-cache behavior are correct (`FE-003`, `FE-004`).
- [ ] Loading, empty, error, forbidden, success and conflict states are handled (`UX-001`).
- [ ] Keyboard, labels, focus, i18n and responsive behavior checked (`UX-002`, `UX-003`).
- [ ] Demo/fixture data cannot masquerade as live data (`FE-007`).

## Human review and release

- [ ] Breaking/public/security/ownership/data changes have required human review (`AI-010`).
- [ ] ADR/exception exists for deliberately bypassed MUST rule (`architecture-security.md`).
- [ ] Staged rollout/rollback is clear for risky change.
- [ ] No merge, deployment, destructive migration or other protected action occurred without approval.

## Outcome template

```text
Verdict: APPROVE / CHANGES REQUESTED / BLOCKED
Highest severity: BLOCKER / HIGH / MEDIUM / LOW / NONE
Required standard IDs:
Evidence (tests and changed files):
Residual risks:
Unverified assumptions:
Human approvals required:
```
