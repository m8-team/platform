# M8 Platform — Engineering Standard v1

**Status:** ACTIVE — normative for new/modified code; automated enforcement is phased.  
**Version:** 1.0 (post-merge governance clarification)  
**Date:** 2026-11-22  
**Scope:** `m8-team/platform`: Go backend, Protobuf APIs, React console, and AI-agent activity.  
**Language:** English is normative for both human and AI contributors.

## Documents

| Document | Scope |
| --- | --- |
| `../../AGENTS.md` | Global agent instructions; canonical policy entrypoint |
| `../../CLAUDE.md` | Claude Code import of common policy |
| `architecture-security.md` | Boundaries, owners, tenancy, authorization, security |
| `module-ownership.md` | Authority registry: implemented, contracted, and planned modules |
| `approval-model.md` | Review owners, exceptions, evidence, and adoption phases |
| `adr/ADR-0001-resource-owner-scope.md` | Proposed resolution of existing owner-kind mismatch |
| `backend-api.md` | Go, ports/adapters, persistence, Protobuf/ConnectRPC, events |
| `quality-testing.md` | Test strategy, evidence, CI gates, rollout |
| `ai-agent-governance.md` | Agent workflow, autonomy limits, approvals |
| `frontend-ux.md` | React/TypeScript/Gravity UI, state, UX, accessibility |
| `review-checklist.md` | PR/agent acceptance checklist |
| `adr-template.md` | Template for recorded architecture exceptions and decisions |

## Normative language

- **MUST / MUST NOT**: required for changed code; deviations must be explicitly approved and recorded.
- **SHOULD / SHOULD NOT**: default expectation; deviations need an explained rationale.
- **MAY**: optional choice within other constraints.
- Requirements use stable IDs (`ARCH-001`, `SEC-001`, `GO-001`, `API-001`, `DATA-001`, `TEST-001`, `AI-001`, `FE-001`, `UX-001`) to identify review findings.

## Applicability and adoption

New and modified code MUST comply unless an explicitly approved migration plan/exception states otherwise. Existing code remains legacy until touched; do not make false claims that older implementations comply. Do not opportunistically rewrite legacy behavior outside task scope. Legacy gap should be documented with file/behavior/risk/follow-up.

Normative requirements apply as policy from adoption in `main`, not as a claim that all legacy code or CI already conforms. Phase 1 CI checks engineering-document consistency and formatting of **changed** Go files; runtime/contract tests remain human or agent evidence until independent gates are introduced. See [approval model](approval-model.md) for accountable roles and enforcement phases and [ownership registry](module-ownership.md) for owner definitions.

Human-approved ADRs and published interface contracts are authoritative architectural decisions. If they conflict with these proposed general guidelines, flag the discrepancy and seek reconciliation; do not silently change either. Human instructions never authorize credential leaks, insecure behavior, or violations of higher-priority safety constraints.

## Project baseline and phased enforcement

- Go module `github.com/m8-team/platform`, toolchain directive `go 1.26.1`.
- `buf.yaml` uses Buf v2, `STANDARD` lint with explicit exceptions, `FILE` breaking policy; `buf.gen.yaml` generates Go/Connect/gRPC/gateway/OpenAPI.
- `internal/access`, `internal/resourcemanager`, `internal/platform`; `cmd/resource-manager` composition root; YDB SDK and Uber Fx are dependencies.
- Console uses React 19, TypeScript 6, Gravity UI, TanStack Router/Query, Vite, and pnpm 10.12.1.
- The Resource Manager README and handlers identify transition/runtime gaps; contract annotations alone are not enough. The `justfile` currently provides Buf-related commands. The new CI workflow checks Markdown policy links, normative IDs and changed-Go formatting only; it does **not** attest to implementation correctness or enforce branch protection without repository admin settings.

## Review and change control

- Changes to this standard require an explicit PR describing motivations, impacted modules, migration, and enforcement.
- Security, tenancy, public contract, resource ownership, or destructive data changes require human architecture/security review under [approval-model.md](approval-model.md). CODEOWNERS is review routing; branch protection and mandatory checks require explicit repository admin configuration.
- Prefer mechanically enforcing objective rules in CI and reserving agent guidance for architectural judgment.
- Do not add `.claude/rules` or automation solely to repeat all content; scope context to changed paths.
