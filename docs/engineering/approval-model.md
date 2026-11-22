# Engineering Approval, Exceptions & Adoption Model

**Status:** ACTIVE (Engineering Standard v1 policy; CI enforcement is phased).  
**Scope:** Changes to code, API, ownership, access rules, deployment, CI and engineering standards.

## Accountable roles

- **Change Author:** responsible for scoped diff, evidence, regression coverage, and PR description.
- **Module Owner:** human maintainer accountable for the state/contract in [module-ownership.md](module-ownership.md). When unassigned, repository maintainers designate an eligible person in the PR; agents MUST NOT assume the role.
- **Architecture Reviewer:** human explicitly designated by repository administrators for cross-module, ownership, consistency or public-contract decisions.
- **Security Reviewer:** human designated for authentication, authorization, tenant isolation, secrets and encryption.
- **Release/Operations Approver:** human authorized for production changes, data migrations, releases and rollback decisions.
- **Repository Administrator:** configures CODEOWNERS, branch rules, reviewer requirements, status checks and team membership.

The initial `.github/CODEOWNERS` routes standards-related reviews to the verified repository maintainer `@sergey-a-gorbachev`. This does **not** prove an independent architecture/security review, and does not require approval without GitHub rulesets. If the author is the only available code owner, administrators MUST arrange another eligible human reviewer for risk-sensitive work or explicitly record why independent review is unavailable and how the risk is handled. No AI agent can act as a human approver.

## Approval matrix

| Risk | Examples | Required review / action |
| --- | --- | --- |
| R0 | Docs, comments, isolated nonbehavioral tooling | Author evidence + applicable repository merge rules |
| R1 | Single-module behavior or UI change | Module owner review where designated, relevant tests |
| R2 | Public contracts, owner boundaries, tenancy, authorization, YDB schema, CI/protection | Module owner plus architecture/security reviewer as relevant; ADR, compatibility and tests where applicable |
| R3 | Production deployment, destructive migration, secret rotation, force push, bulk deletion | Specific informed authorization from operations/security/repository authority, rollback/roll-forward plan and audit |

A person may hold more than one role only if repository governance explicitly allows it. **No eligible reviewer or documented authorization = no presumed approval.** Review approval MUST be attributable to a human and associated with a PR or accepted ADR. AI cannot waive requirements.

## Exceptions to a MUST requirement

Exceptions require (1) stable requirement ID, business reason, impact/threat and owner; (2) time-bound expiry or measurable migration milestone; (3) compensating controls and evidence; (4) human approval under the risk matrix, linked in an accepted ADR or tracked issue; (5) follow-up owner and due date. Expired exceptions require explicit reapproval. A failing test, unresolved note or self-approved agent report is not an exception.

## Adoption and enforcement

- **Phase 0 — normative policy:** root/scoped AGENTS and engineering standards apply to touched code; existing legacy gaps are logged, not silently expanded.
- **Phase 1 — limited CI:** `.github/workflows/engineering-standards.yml` checks engineering links and rule-ID uniqueness, plus Go formatting of changed files only. These checks are not guaranteed merge requirements until administrators configure branch protections.
- **Phase 2 — baseline validation:** add independently green Go vet/tests, Buf lint/breaking/generated-code drift, frontend lint/build, vulnerability and architecture checks. Track existing baseline failures explicitly.
- **Phase 3 — required gates:** administrators activate required status checks, suitable code-owner reviews and bypass restrictions; verify via test PR.

**Repository administrator action still required:** configure GitHub branch protection/rulesets and designate eligible Architecture/Security reviewers. CODEOWNERS and GitHub Actions YAML alone cannot enforce them. Do not mark enforcement complete until checked in GitHub.
