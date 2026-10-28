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

### Resource contract validation

`ResourceRef` contains an exact kind key and a resource UUID. Built-in kind keys
include `organization`, `project` and `workload_service`; module-owned kinds may
be namespaced, such as `provisioning.database`. Neither a reference nor knowledge
of its UUID grants access: resolution must enforce the caller's tenant and
authorization scope. External provider identifiers are not catalog UUIDs.

When an `OwnerScope` is present, exactly one Organization, Workspace or Project
UUID is required. An optional scope filter can be omitted, but `{}` is invalid.
The application still checks the selected owner's existence, tenant, lifecycle
and allowed placement; the oneof cannot establish those facts.

`ResourceMetadata.name` is an optional display name, not a resource path or
identity. Existing name/description limits remain 256/1024 Unicode characters.
Create input may omit the server-generated ID, lifecycle, revision and timestamps.
State must be a defined enum value and revision cannot be negative; zero values
remain valid before creation. Handlers must ignore output-only input fields,
merge field-mask updates into stored state, and validate the resulting value.
Requests carry concurrency preconditions in `expected_revision`, not metadata.

`ResourceDescriptor` requires the canonical module resource reference, one owner,
catalog metadata and a module key. Its derived Organization UUID may be omitted
on input. The catalog entry's metadata ID/revision do not replace the canonical
resource reference or its module-owned version; the schema does not require the
two IDs to be equal. The server must check that the module matches the registered
kind policy's `responsible_module`.

`ResourceKindPolicy.allowed_owner_kinds` is an unordered set containing only
`organization`, `workspace` and `project`. Duplicates and unsupported owner kinds
are invalid. An empty set permits no placement. `environment_binding_allowed`
defaults to false; enabling it does not bypass authorization or same-Organization
checks. Policies apply to module-owned kinds and cannot redefine the built-in
containment matrix.

Example `RegisterResourceKind` payload for the intended registration API:

```json
{
  "policy": {
    "kind": "provisioning.database",
    "allowedOwnerKinds": ["organization", "project"],
    "responsibleModule": "provisioning",
    "environmentBindingAllowed": true
  },
  "idempotencyKey": "register-provisioning-database-v1"
}
```

These constraints preserve field names, numbers and types, while tightening
validation of previously unconstrained values. Contract tests validate generated
messages with Protovalidate. They do not establish runtime enforcement:
ResourceServer does not invoke Protovalidate, descriptor creation is unsupported,
and RegisterResourceKind returns Unimplemented. The current general resource
adapter also drops labels and delete_time and does not merge sparse field masks
correctly. The OpenAPI generator uses unqualified schema names, so the separate
`common.operation.v1.ResourceRef` can collide with this Resource Manager type;
generated OpenAPI references need that separate generator fix before SDK use.

## 3. Environment behavior

Organization environments are registered independently of Workspaces/Projects. Examples: production, staging, testing, development, preview. Each has an identity, classification, isolation profile, lifecycle and optional TTL. Their names need only be unique within the owning Organization.

Resource Definitions/Versions are not cloned just to switch environments. Installations / EnvironmentBindings point to explicit immutable version references, separate configurations and credentials. Access to Testing must not imply access to Production. A resource is bound to an Environment only if its derived owning Organization matches the Environment's owning Organization.

A single deployed M8 Platform control plane can host customer Production and Testing environments, **without** promising physical isolation. ResourceManager records the requested isolation profile; Provisioning / Runtime enforce the actual runtime and data isolation. This is verified against actual topology at promotion time.

### Environment contract validation

`Environment.organization_id` is an immutable Organization UUID. `key` is an
immutable lowercase slug of alphanumeric segments separated by single hyphens
(for example, `preview-123`), unique within that Organization. `metadata.name`
is a display name, not the uniqueness key. `kind` is explicit classification
independent of the key; both `kind` and `isolation` must be defined, nonzero
enum values. `expires_at` is an optional absolute timestamp, not a duration.
Omitting it means no scheduled expiration; expired records remain valid API
representations. The schema does not restrict expiration to preview environments.

An EnvironmentBinding is unique by `(resource, environment_id)`. Both references
are immutable for that binding. `version_ref` identifies an immutable version
in the owning module, but the binding can be updated to point to another
version. `configuration_ref` is optional and contains only an opaque reference.
The server checks the resource kind's `environment_binding_allowed` policy,
matching owning Organizations and authorization; the binding grants no access.
Mutations use the request's `expected_revision`, not the output-only
`binding.revision`, as their concurrency precondition.

The protobuf validation rules apply to complete Environment/EnvironmentBinding
values. For field-mask updates, application handlers must merge the selected
fields with stored state before validating, and separately reject changes to
immutable fields. Metadata IDs and binding IDs may be absent on creation because
the server assigns them; nonempty IDs must be UUIDs. Validation of a complete
value does not establish persistence, authorization, uniqueness, idempotency,
concurrency checks or expiration enforcement.

Example create payload (Organization UUID must refer to an existing authorized
Organization):

```json
{
  "resource": {
    "environment": {
      "metadata": {"name": "Checkout preview"},
      "organizationId": "11111111-1111-4111-8111-111111111111",
      "key": "preview-123",
      "kind": "ENVIRONMENT_KIND_PREVIEW",
      "isolation": "ISOLATION_PROFILE_LOGICAL"
    }
  },
  "idempotencyKey": "checkout-preview-123"
}
```

These rules preserve protobuf field names, numbers and types, but reject
previously unconstrained malformed values when Protovalidate is invoked.
The current ResourceServer does not invoke Protovalidate and does not persist
`kind`, `isolation` or `expires_at`; these remain implementation gaps. Its
`SetEnvironmentBinding` handler also does not yet enforce the request's
`expected_revision` or `idempotency_key`. Contract tests validate the schema,
not those runtime guarantees.

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

## 10. Module catalog and installation

`ModuleRegistry` owns `ModuleDefinition`, immutable `ModuleVersion`, Tenant-scoped `ModuleEntitlement`, and `ModuleInstallation` at an Organization/Workspace/Project scope. The module remains the owner of its business objects. An entitlement means the tenant may install the module, **not** that all users gain rights. Installations require an active entitlement, kind-aware placement and Access permissions. Per-environment rollout is handled through the declared EnvironmentBinding and Delivery workflows, not by duplicating ModuleDefinitions.

Contract: `api/proto/m8/platform/modules/v1/module.proto` and `module_registry_service.proto`.

## 11. Domain events

Resource Manager emits `ResourceChangedEvent`, `GroupAffiliationChangedEvent` and `EnvironmentBindingChangedEvent`; Access emits `AccessGrantChangedEvent`. Event envelopes carry actor, source resource/organization, revision, correlation and idempotency metadata. Events contain no credentials or PII beyond minimal audited identifiers. Outbox, publication and consumer idempotency must be implemented with the corresponding running services.

## P1: Organization Groups vs. legal corporate relationships

Three independent concepts are deliberately modeled separately:

| Entity | Meaning | Access inheritance |
| --- | --- | --- |
| `Organization` | Independent company and its resource boundary | Never implicit |
| `Group` / `GroupOrganizationLink` | Optional business grouping (non-owning) | Never |
| `OrganizationRelationship` | Dated, directed business or legal relation between two Organizations | Never |

Corporate relationships MUST NOT be written to OrgUnit adjacency/closure tables.
A Person may belong to multiple Organizations, but that information belongs to
M8 Identity / Directory; the corporate graph contains no employee assignments.

### Type semantics and temporal constraints

The typed service `m8.platform.resourcemanager.v1.OrganizationRelationshipService` supports:

- `PARENT_SUBSIDIARY`: from = direct parent/shareholder, to = subsidiary;
  optional `ownership_basis_points` (1..10000, 0 = unknown). Multiple direct
  parent organizations are supported. The sum of *known overlapping* direct
  shares for the same subsidiary may not exceed 100%.
- `AFFILIATE`: explicitly recorded related entity, with declared direction
  and no automatic inference of control.
- `PARTNER`: contractual or operational collaboration, without assumed control.
- `SERVICE_PROVIDER`: from organization provides services to the to organization.

`valid_from` is inclusive, `valid_to` is exclusive; absent `valid_to`
means indefinite. A history snapshot is stored at each revision. Only stake,
valid-to, and evidence reference may be amended; changing endpoints, type
or valid-from requires closing and creating a new relationship. Closing never
deletes the companies or prior snapshots.

Duplicate overlapping relationships and temporal cycles in
`PARENT_SUBSIDIARY` are rejected. Cycles use overlapping *time
intervals*, not the union of disjoint historic graphs. Tenant and Organization
existence is verified against authoritative resource records in the same YDB
transaction. Both endpoints must be active Organizations of the same Tenant.
Cross-Tenant federation is a distinct future feature.

### API and implementation

- `POST /resource-manager/v1/organization-relationships`
- `GET /resource-manager/v1/organization-relationships?tenantId=<UUID>`
- `GET /resource-manager/v1/organization-relationships/{id}?tenantId=<UUID>`
- `PATCH /resource-manager/v1/organization-relationships/{id}`
- `POST /resource-manager/v1/organization-relationships/{id}:close`
- `GET /resource-manager/v1/organization-relationships/{id}/versions`

Mutations require an idempotency key; changes require expected revision.
History is immutable, and all YDB mutations write an outbox event in the
same serializable transaction. A shared Tenant graph-head write serializes
concurrent changes that would otherwise independently produce cycles.
The API uses query-scope-bound pagination; it never supports an unscoped
global inventory.

**Deployment:** apply
`docs/executable-baseline/data/ydb/m8-resource-manager/004_organization_relationships.sql`
after migration 003. The Resource Manager intentionally does not auto-migrate
production schema on startup. Integration tests apply both migrations to
disposable YDB.

**Security:** legal or business relationships confer no M8 Access
permissions, role grants, session visibility, or access to employees.
Resource Manager's existing Access production gate remains fail-closed:
until a real M8 Access authorization integration is deployed, production
reads and writes to this service are denied. Loopback development mode is
not a production authentication substitute.

### UI

`/resource-manager/relationships` shows Tenant-scoped corporate links
independently from Group membership. The Group detail view shows relations
involving its member Organizations. Both use the real REST gateway, show
effective dates, direct stakes, state, evidence reference and immutable
version history, and permit create/update/close with explicit review.

**Known constraints:** corporate relations do not compute ultimate-beneficial
ownership, legal consolidation, or voting rights; those interpretations
require a separate governed domain. Read-pagination tokens are currently
scope-bound but not snapshot-pinned. Emptying a previously set valid-to
is not yet an update operation; close and re-create if an unbounded
validity period is needed. No Group membership or corporate relationship
ever changes tenant/resource ownership.
