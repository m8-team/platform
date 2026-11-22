# ADR-0001 — Align ResourceKindPolicy owner kinds with OwnerScope

- **Status:** Proposed (not accepted; no Protobuf change in this PR)
- **Date:** 2026-11-22
- **Decision owners:** Resource Manager Module Owner and Architecture Reviewer, to be designated
- **Affected modules:** Resource Manager; resource-kind registration and API consumers
- **Requirement IDs:** ARCH-002, ARCH-005, API-002, API-009, SEC-001
- **Origin:** Post-merge architecture review of PR #55

## Context

`proto/m8/platform/resourcemanager/v1/resource.proto` currently permits strings `organization`, `workspace`, and `project` in `ResourceKindPolicy.allowed_owner_kinds`. The `OwnerScope` oneof only represents `organization_id` and `workspace_id`; field 3 and name `project_id` are reserved. This pre-existing inconsistency is not caused by PR #55.

An owner kind that cannot be represented may allow an impossible policy or a dangerous fallback. No Project ownership semantics should be inferred from the existing string enum.

## Alternatives

| Alternative | Benefit | Cost / risk |
| --- | --- | --- |
| A — Remove `project` from allowed policy kinds | Aligns contract with currently representable owner kinds | Validation and client compatibility change; requires migration |
| B — Introduce a Project owner scope | Explicit hierarchy support | Cannot reuse reserved field 3; demands new field, ownership/authorization model and migration |
| C — Project as non-owner metadata | May retain Project references without hierarchy changes | Must distinguish business link from authoritative owner |

## Proposed decision (not yet accepted)

**Recommend A**: the supported placement owners should be Organization/Workspace. Do **not** change published Protobuf in this documentation PR. Until the decision is approved, newly implemented registration and placement paths MUST reject Project as an unsupported owner; they MUST NOT map it to Workspace or infer tenant/permissions from a Project identifier.

This ADR is an explicit open decision, not authorization to make breaking API changes.

## Compatibility, migration and verification

1. Inventory existing persisted policies and clients that use `project`.
2. Agree whether Project is a real owner or business attribute.
3. Obtain Resource Manager and Architecture Reviewer approval.
4. Update Protobuf (if approved) in a separate contract PR, with Buf checks and client migration plan.
5. Add negative tests for unsupported placement/policy, tenant mismatch, retries and valid owner scopes.
6. Regenerate Go/HTTP/OpenAPI; stage rollout and rollback/roll-forward.
7. Mark the ADR Accepted only after recorded approval, with tracked implementation work.

## Security guarantees

Never derive tenant ownership from client Project strings, silently remap unsupported owner kind, or let `allowed_owner_kinds` claim placement that cannot pass server-side authorization/validation.
