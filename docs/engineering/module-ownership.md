# M8 Module Ownership Registry (v1)

**Status:** ACTIVE as a reference inventory; entries labelled planned or contracted do **not** authorize a new production implementation.  
**Principle:** One authoritative write owner per state transition. A Protobuf package or Console fixture does not prove a working backend.

| Capability | Authoritative state / expected owner | Must not own | Evidence / maturity |
| --- | --- | --- | --- |
| Resource Manager | Tenant metadata, Organization/Workspace hierarchy, resource descriptors/owner scopes, environment bindings | Principal lifecycle, authorization decisions, business records, actual rollout | `internal/resourcemanager/` and `proto/m8/platform/resourcemanager/v1/` exist; replacement v1 runtime transition **incomplete** |
| Access | Roles/grants, relationships, authorization model/policy decisions | Identity lifecycle, org hierarchy, credential sessions | `internal/access/`, `proto/m8/platform/access/v1/`; production completeness requires inspection |
| Identity | Principals, linked/external identities, identity lifecycle | Access grants/policy, owner hierarchy, authentication session details | `proto/m8/platform/identity/v1/` exists; implementation maturity **unverified** |
| IAM / Authentication | Authentication/session/credential and assurance workflows (pending approved split) | Access decisions, business data, Resource Manager hierarchy | `proto/m8/platform/iam/v1/`; authority split with Identity **requires ADR** |
| Delivery / Runtime | Workload rollout, desired/observed state, reconciliation | Resource Manager logical owner metadata | `proto/m8/platform/delivery/`; **contracted/planned**, verify runtime |
| Risk Decision | Risk evaluations, signals/policy decisions, explainability | Identity/session authority, Access grants, owner hierarchy | `proto/m8/platform/riskdecision/`; **contracted/planned**, source authority pending |
| Gateway / API Access | Gateway route/policy configurations, rate limits, reconciliation | Per-object business authorization and source-of-truth ownership | `ui/console/src/modules/gateway/`; **UI evidence only**, backend owner unverified |
| Quota Manager | Quota rules/reservations/usage reconciliation (if implemented) | Provider capacity facts, billing ledger, owner hierarchy | `ui/console/src/modules/quota-manager/`; **UI evidence only**, backend unverified |
| Audit | Audit event retention/export store (if implemented) | Other modules' business state or auth policy | `ui/console/src/modules/audit/`; **UI evidence only**, backend unverified |
| Workforce / HR | Worker, employment and contract lifecycle | Access decisions, identity sessions, owner hierarchy | `proto/m8/platform/workforce/`; **contracted/planned** |
| Shared Services | Shared service requests/case workflow | Resource Manager hierarchy, credentials, grants | `proto/m8/platform/sharedservices/`; **contracted/planned** |
| Console | Presentation state and bounded client caches | Authoritative authorization, server persistence | `ui/console/` exists; demo data is not authoritative |

## Registering a new owner

Before introducing authoritative storage or state transitions, a module README and reviewed ADR MUST define: entities/resources it writes; schema/API/events and consumers; tenancy/owner scope; allowed permissions; write transaction and concurrency model; projection consistency; idempotency/replay; operational maturity (`planned`, `contract-only`, `dev/fixture`, `production-ready`); human approver and verification evidence.

## Explicit unresolved boundaries

- **Project owner:** `ResourceKindPolicy.allowed_owner_kinds` includes `project` while `OwnerScope` has only Organization/Workspace. See [ADR-0001](adr/ADR-0001-resource-owner-scope.md). No new Project owner placements until an accepted contract migration.
- **Identity vs IAM/Authentication:** credential/session authority needs a separate ADR; do not create duplicate principal/session sources of truth.
- **Gateway, Quota, Audit, Risk Decision:** verify actual backend ownership before claiming production readiness; Console views and API namespace do not establish a source of truth.

The existing `internal/resourcemanager/README.md`, `internal/access/README.md` and reviewed contracts remain authoritative within approved module boundaries, subject to explicit recorded implementation gaps.
