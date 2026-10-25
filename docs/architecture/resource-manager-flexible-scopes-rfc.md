# RFC: Resource Manager v2 — flexible scopes and company groups

**Status:** Proposed for review (not yet the production API)  
**Context:** M8 Resource Manager  
**Date:** 2026-10-25

## 1. Problem in the current repository

The active protobuf contracts and `docs/architecture/resource-manager-spec.md` require `Organization -> Workspace -> Service`, with immutable `workspace_id` and `environment` for a Service. The Requirements Catalog (`RM-FR-020` onward), SPDD context and executable YDB baseline also define a `Project` entity, which is not present in the active protobuf v1 API. Only OrganizationService is currently implemented end-to-end and uses local in-memory adapters.

The fixed hierarchy creates artificial Workspaces and Services for HR records, AI agents, documents, workflows and shared business services. At the same time, the developer-platform use case genuinely needs Workspace, Project and deployment-aware Service.

## 2. Decision

Keep **Organization** as the mandatory isolation/ownership root for tenant-owned resources. Make **Workspace** and **Project** optional management scopes. Reserve **ServiceRegistration / WorkloadService** for technical runtime services; do not reuse this entity for HR service offerings, Agent definitions, modules or Temporal workflow executions.

Adopt a **typed placement matrix**, not an arbitrary parent graph:

| Child kind | Allowed immediate parent |
| --- | --- |
| Group | None; group is a grouping root |
| Organization | None; association to Group is a separate edge |
| Workspace | Organization |
| Project | Organization **or** Workspace |
| Workload Service / existing Service | Workspace (legacy), Project (new v2) |
| Catalog Resource: Agent, Workflow, Connection, Knowledge | Organization, Workspace or Project as allowed by a per-kind policy |

A resource has exactly one **owner scope**. A resource may have multiple explicit *references*, *bindings* or *grants* elsewhere; these are not parentage. Types with a business owner other than the above (Employment, SupportCase, ServiceAgreement) keep their canonical payload in their own domain and register only a typed metadata reference when needed.

Do not automatically create hidden default Workspaces, Projects or Services. A company can use M8 Agents directly within its Organization.

## 3. Distinct graphs and domain ownership

* **Resource Manager**: Group metadata (holding), Group-to-Organization association, Organization, Workspace, Project, ServiceRegistration metadata, ResourceDescriptor/ScopeRef, lifecycle, canonical ownership path.
* **Identity**: Person, user, IdP identities, Employee, Employment, OrganizationMembership, Department, Team. Neither users nor departments are Resource Manager aggregates.
* **Access**: roles, permissions, ReBAC edges, delegation, permission decisions and policy inheritance. Being in the same Group does not imply cross-company access.
* **Shared Services**: provider/consumer Agreement, Offering, Subscription, Case; these refer to Organization IDs and do not create ownership parentage between companies.
* **Modules / Agents / Runtime**: module installation, agent installation, application logic, workflow state, credentials and deployment status are module-owned, and use ScopeRef for ownership.

A Group may contain multiple Organizations by an explicit affiliation edge; Group is not a super-admin tenant security boundary. Whether a given company is its own Organization is a policy/default of organization onboarding, not a requirement of the identity model.

## 4. Scenarios

**Simple company HR request:** Organization A owns AgentInstallation, WorkflowDefinition and HR API connection directly. No Workspace/Project is needed.

**Team collaboration:** Organization A -> Workspace Engineering -> Project OMS -> Service API. Service has deployment bindings to dev/stage/prod in a future release. Existing v1 per-environment Service records remain valid during migration.

**Multi-company employee:** one Identity principal may have distinct Employment and Membership records for A and B; every tool call evaluates the selected target resource under the corresponding Organization. No merging of entitlements.

**Shared IT Service Desk:** Organization Shared-IT provides services to Organizations A/B through ServiceAgreements. Support cases are owned by the consumer service domain; technician/agent access is granted for specific actions/cases with expiry. Provider and consumer relations never mutate Resource Manager parentage.

**Group analytics:** aggregate from explicitly authorized Organization scopes, with per-Organization approval/policy checks. A Group role alone does not grant read access to its companies.

## 5. Scope API v2 proposal

Use a typed reference instead of adding nullable `organization_id`, `workspace_id`, `project_id` to every resource:

~~~proto
message ResourceRef {
  string type = 1;  // registered resource kind
  string id = 2;    // stable UUID
}

message ResourceScope {
  ResourceRef owner = 1;  // organization | workspace | project
}

message ResourceDescriptor {
  ResourceRef resource = 1;
  ResourceScope scope = 2;
  string owning_organization_id = 3;  // server derived
  string display_name = 4;
  int64 revision = 5;
}
~~~

Precise protobuf contracts, validation annotations, names and resource-name patterns require contract review before publishing. Resource IDs must stay immutable. Paths should be **derived from IDs**; mutable display names/slugs are not primary keys. API must offer GetResource, ListChildren, ResolveScope/Ancestors and ListByScope including descendants with explicit authorization.

For published module resources, maintain a per-resource-type **AllowedOwnerKinds** policy. A ServiceRegistration should not be the universal parent for module-owned resources.

## 6. Lifecycle and safety invariants

1. Each resource has exactly one owning Organization (computed by traversing typed containment, not supplied by client).
2. No containment cycles; hierarchy depth and maximum fan-out are bounded.
3. An active child requires an existing, active parent.
4. Delete/suspend/archive behavior checks descendants and references; no silent recursive destructive deletion.
5. Scope changes / project moves require explicit API, authorization on source and destination, optimistic concurrency, and audit.
6. Access grants are independent from Group affiliation and do not automatically cross Organization boundaries.
7. Per-resource Access checks still occur after containment resolution; paths are not authorization decisions.
8. Change events use stable IDs, aggregate revisions, outbox and idempotency; follow existing domain event conventions.
9. A service's `environment` in v1 remains immutable until a deliberate v2 migration to logical service vs deployment/installation is approved.
10. Retrieval and listing must enforce scope filtering server-side before pagination to avoid data leakage.

## 7. Migration strategy (nonbreaking first)

**Phase A — documentation + domain primitives:** agree on typed scope and placement matrix; include pure domain tests. Preserve v1 APIs as-is.

**Phase B — production Resource Manager:** persistent YDB repositories, child indexes, idempotency, transactional outbox, Access and Audit ports. Implement Workspace and Project CRUD and Group associations with dedicated v2 proto/API and compatibility review. Existing Service's `workspace_id` remains supported.

**Phase C — resource catalog:** implement typed ResourceDescriptor and per-kind allowed scope policies. Agents/Workflows/Connections can then register directly at Organization or optional Workspace/Project level.

**Phase D — Service refactor:** separate runtime ServiceRegistration from business ServiceOffering and module installation. Introduce optional Project ownership and environment deployment bindings with explicit migration from existing Service v1 semantics.

**Phase E — Shared Services:** publish relationships and scope-aware delegation APIs through Shared Services + Access, not through RM containment.

No existing proto fields are removed or repurposed in place. Build new contracts, test data migration on a snapshot, roll out readers before writers, emit versioned events, deprecate only after consumers migrate.

## 8. Out of scope for this RFC

Implementing Employment, Department, Team, permission decisions, billing, Agent Runtime, a universal business object database, or an entire generic polymorphic CRUD API in Resource Manager.

## 9. Acceptance criteria

- Organization-owned Agent/Workflow can exist without Workspace or Project.
- Project can be owned by Organization or Workspace, never Service.
- Workspace always belongs to one Organization.
- Existing Workspace-owned Service remains a valid placement.
- Group membership does not grant permissions on member Organizations.
- One principal can access A and B under independent membership/policy evaluations.
- Shared IT technician can be authorized for Case A but denied for unrelated A/B resources.
- All resource listings are authorization filtered and have deterministic pagination.
- All mutations have optimistic concurrency, idempotency, outbox event and audit trace.
- v1 Organization/Workspace/Service contracts remain compatible until an explicit v2 transition.
