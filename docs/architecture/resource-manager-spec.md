# M8 Resource Manager — canonical v1 resource and ownership model

**Status:** Canonical contract design (breaking replacement of the previous protobuf v1).  
**Source of truth:** `api/proto/m8/platform/resourcemanager/v1`.  
**Implementation status:** Proto contract sources added; generated code, service implementation, persistence and migration are **not yet updated**. Do not deploy before those are complete.

## 1. Design decisions

M8 operates for independent companies, optional groups/holdings, external providers and users with temporary grants. The five independent axes are:

1. **Tenancy:** Tenant/Account is the customer/subscription and billing boundary.
2. **Ownership:** Organization is the mandatory security and ownership root; optional Workspace/Project organize work and manage resources.
3. **Execution:** Environment is an Organization-owned registry entity and **NOT** a containment parent. Preview environments may expire.
4. **Deployment:** ResourceManager EnvironmentBinding associates logical resources/versions with environments. Delivery owns rollout/status, module implementations own configurations and secrets.
5. **Authorization:** Identity and Workforce own principals/relationships. Access owns all permissions, grants, risk, delegation and real-time access decisions.

## 2. Ownership and placement matrix

| Resource type | Immediate owner | Notes |
| --- | --- | --- |
| Tenant | platform | root/account; no parent Organization |
| Group | Tenant | optional, not tenant security inheritance |
| Organization | Tenant | Group membership is separate association |
| Workspace | Organization | optional |
| Project | Organization OR Workspace | optional |
| Environment | Organization | not a parent of arbitrary business resources |
| WorkloadService | Workspace OR Project | logical workload independent of environment |
| ResourceDescriptor (Agent/Workflow/Collection/Connection/etc) | Organization, Workspace OR Project **subject to registered policy** | module owns actual data |
| EnvironmentBinding | ResourceRef + Environment | relation only; server verifies matching owning Organization |
| Deployment | ResourceRef + Environment | Delivery owns actual deployment / runtime state |

A Group affiliation does not imply any capability to access member Organizations. An independent Organization has no Group. A user may access a resource without OrganizationMembership through an explicit and authorized grant.

A logical resource has exactly one immediate owner and exactly one derived owning Organization. External relations are not containment edges. All paths are derived from immutable identifiers rather than display names.

## 3. Environment behavior

Organization environments are registered independently of Workspaces/Projects. Examples: production, staging, testing, development, preview. Each has an identity, classification, isolation profile, lifecycle and optional TTL. Their names need only be unique within the owning Organization.

Resource Definitions/Versions are not cloned just to switch environments. Installations / EnvironmentBindings point to explicit immutable version references, separate configurations and credentials. Access to Testing must not imply access to Production. A resource is bound to an Environment only if its derived owning Organization matches the Environment's owning Organization.

A single deployed M8 Platform control plane can host customer Production and Testing environments, **without** promising physical isolation. ResourceManager records the requested isolation profile; Provisioning / Runtime enforce the actual runtime and data isolation. This is verified against actual topology at promotion time.

## 4. Multi-company and workforce

Identity registers global principals (user, service account, agent, workflow, external agent) with separately verified external identities. Do NOT auto-merge identities by email. An identity can be independently linked to multiple Organizations, and may have **no membership** in an Organization that grants case-specific access.

Workforce/HR owns Employment and Engagement records (permanent employee, temporary employee, contractor, freelancer, consultant, auditor, provider). These are not permissions. Engagement validity is independent of AccessGrant validity.

SharedServices owns service providers/consumers, offerings, agreements and cases. Agreements authorize a business relationship but do not automatically grant access. Access grants are separately approved and limited to subject, target resource, organization, environment, actions, purpose and expiry.

## 5. AccessGrant lifecycle

A temporary grant requires start and expiry timestamps. State transitions are Requested -> PendingApproval -> Active or Rejected; Active -> Revoked or Expired. `expires_at` is evaluated **at authorization time**, not only by a scheduled worker. Revocation invalidates newly issued access decisions and external credentials. High-risk actions must reauthorize with strong consistency and recheck all dynamic conditions (case is open, assignment active, agreement valid, MFA, owner approval).

Requester/approver/revoker are server-authenticated, not client-authored values. Enforcement at Tool Gateway and target service may deny independently. AccessGrant does not bypass AuthZEN/PDP.

## 6. Resource Manager API

The new `ResourceManagerService` has typed Create/Get/List/Update/ChangeState plus MoveProject, ResolveResourceScope, Group link operations, Environment bindings and resource-kind registration.

**Invariants:**
- IDs are server-generated and immutable.
- Create mutators are idempotent; expected revisions guard updates and deletes.
- Changing Project owner requires a dedicated MoveProject with source/destination checks.
- KindPolicy constrains where module-owned resource descriptors can live.
- No hierarchy cycles, orphaned resources, hidden default Workspaces or Projects.
- Child creation only under active parents.
- Suspended/deleted parent blocks descendant writes; cleanup requires explicit policy and workflow.
- Server computes owning_organization_id and validates same-Organization EnvironmentBindings.
- List queries MUST scope authorization server-side before pagination.
- Audit/event outbox and idempotency commit atomically with resource changes.
- Group affiliation and SharedServices agreements never create implicit ReBAC inheritance.

The wire-format protobuf contracts are not themselves a full enforcement mechanism. Application handlers must implement all of these invariants.

## 7. Contract files

- Resource Manager: `resource.proto`, `tenant.proto`, `group.proto`, `organization.proto`, `workspace.proto`, `project.proto`, `environment.proto`, `service.proto`, `resource_manager_service.proto`.
- Identity: `identity/v1/principal.proto`, `identity_directory_service.proto`.
- Workforce: `workforce/v1/engagement.proto`, `workforce_service.proto`.
- Access: `access/v1/access_grant.proto`, `access_grant_service.proto` (existing AuthZEN decision API remains).
- Shared Services: `sharedservices/v1/shared_services.proto`, `shared_services_service.proto`.
- Delivery: `delivery/v1/deployment.proto`.

The old `OrganizationService`, `WorkspaceService`, `ServiceService` and immutable `Service.environment` API have been removed from source without a compatibility layer by explicit product decision.

## 8. Required follow-up to make new v1 executable

1. Run `buf lint` and `buf generate` in a connected CI/build environment. Review generated gateway and OpenAPI output; add validation to all newly introduced request fields.
2. Regenerate/replace the separate `github.com/m8-team/go-genproto` module and update M8 Platform imports. Remove stale `api/generate/go` outputs from deleted service RPCs.
3. Replace the current OrganizationService Go implementation with the new ResourceManagerService handlers and implement Workspace/Project/Environment/Tenant/Group/ResourceCatalog.
4. Implement YDB schema changes, lifecycle, parent indexes, server-side tenant/scope filtering, transactions, outbox/idempotency, and Access/Audit adapters.
5. Implement Identity/Workforce/AccessGrant/SharedServices handlers in their own bounded contexts.
6. Execute an explicit **breaking** migration (schema, generated SDKs, clients, docs, events, CI baseline). There is **no promise of old client compatibility**.
7. Add contract/integration/security tests, then enable strict break checks against this new v1 baseline.

## 9. Acceptance criteria

- A single-company Tenant has one Organization and no Group, Workspace or Project.
- Holding Tenant has multiple Organizations; Group is optional; affiliation grants no data access.
- AgentDefinition/Workflow directly owned by Organization requires no dummy Workspace/Project.
- Engineering Organization can create Workspace -> Project -> logical WorkloadService.
- Testing & Production are distinct Environments with separate bindings, credentials and approval.
- Contractor with NO OrganizationMembership can temporarily access one project via approved AccessGrant.
- Support operator employed by Provider can access Consumer resource only for an active case with an unexpired grant, not its sibling resources.
- Revocation or expiry blocks the next access check even if the workflow continues.
- Cross-tenant service agreements are references, never parent-child scope edges.
- All API responses are authorization scoped; no information leaks via enumeration/pagination.
