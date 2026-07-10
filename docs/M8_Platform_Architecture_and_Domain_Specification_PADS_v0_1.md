# M8 Platform Architecture & Domain Specification
_PADS-000 · Version 0.1 · Baseline architecture and domain model · 10 July 2026_

| Field | Value |
| --- | --- |
| Document ID | PADS-000 |
| Version | 0.1 |
| Status | Draft baseline / Architecture seed |
| Owner | Sergey Gorbachev |
| Platform | M8 Platform |
| Scope | Resource Manager, Identity, Authentication, Access, Risk Decision, Provisioning, Audit, Common Operation |
| Architecture style | Domain-Driven Design, Clean Architecture, API First, Event-Driven, Control Plane |
| Core stack | Go, Protobuf, ConnectRPC, buf.validate / Protovalidate, YDB, YDB Topics, Redis, Temporal, SpiceDB, Keycloak, OpenTelemetry |

> **Normative intent:** This document is the first source of truth for platform boundaries, domain language, ownership, requirements distribution and SPDD mapping. Any deviation must be recorded as an ADR.


---

# 0. Document Control

| Version | Date | Status | Description |
| --- | --- | --- | --- |
| 0.1 | 2026-07-10 | Draft baseline | Initial PADS artifact: platform vision, domain model, context map, service boundaries, requirements model and SPDD mapping. |

## 0.1 How to use this document

PADS is a normative engineering specification. It should be read before writing requirements, protobuf contracts, implementation tasks, ADRs, SPDD prompts or generated code. It defines the vocabulary, service boundaries and architectural constraints that all subsequent artifacts must reference.

- Use section identifiers as stable anchors in requirements, ADRs, prompts and tests.
- Use MUST for mandatory constraints, SHOULD for strong defaults and MAY for allowed optional behavior.
- When a product decision conflicts with this document, create or update an ADR before implementation.
- Do not copy external system concepts directly into the domain model; use anti-corruption layers.

## 0.2 Normative language

| Term | Meaning |
| --- | --- |
| MUST | Mandatory rule. Implementation is invalid when the rule is violated. |
| MUST NOT | Mandatory prohibition. The system must not implement the described behavior. |
| SHOULD | Strong recommendation. Deviation requires documented reasoning. |
| MAY | Allowed option. Implementation may choose it when it does not violate mandatory rules. |


---

# Table of Contents

- 1. Purpose and Scope
- 2. Platform Vision
- 3. Design Goals
- 4. Architecture Principles
- 5. Ubiquitous Language
- 6. Business Capability Map
- 7. Domain Model
- 8. Context Map
- 9. Bounded Context Specifications
- 10. Shared Kernel and Common Contracts
- 11. Data Ownership
- 12. API Design Rules
- 13. Event Design Rules
- 14. Integration and Consistency Model
- 15. Security Architecture
- 16. Long Running Operations
- 17. Error Model
- 18. Observability
- 19. Quality Attributes
- 20. Requirements Distribution
- 21. Traceability Model
- 22. SPDD Mapping
- 23. Architecture Governance
- 24. Glossary


---

# 1. Purpose and Scope

## 1.1 Purpose

This specification defines the architecture and domain model of M8 Platform. It establishes stable boundaries between bounded contexts, assigns ownership of domain concepts and data, describes mandatory architectural principles, and provides the traceability path from platform requirements to Structured Prompts and code.

The document is intentionally platform-level. It does not replace service-level specifications, protobuf definitions, ADRs or implementation tasks. Instead, it governs them.

## 1.2 Platform scope

M8 Platform is a modular control-plane and IAM platform. The first baseline includes resource hierarchy management, identity, authentication, access control, risk decisions, provisioning, audit and shared operation management.

| In scope | Description |
| --- | --- |
| Resource hierarchy | Organization → Workspace → Project → Service as the canonical management hierarchy. |
| Identity | User pools, users, groups, memberships and external identities. |
| Authentication | Authentication transactions, challenges, CIBA, OTP, WebAuthn, OIDC/SAML handoff, step-up and re-authentication. |
| Access | Roles, permissions, relationships, permission checks and explanation through SpiceDB-backed implementation. |
| Risk Decision | Risk signals, policy evaluation, decisions and challenge requirements. |
| Provisioning | Managed resource lifecycle, desired/observed state and reconciliation. |
| Audit | Immutable audit events for all significant actions. |
| Operations | Shared long-running operation model based on google.longrunning semantics and universal OperationMetadata. |

## 1.3 Out of scope for this baseline

- Final UI screens and user interface interaction design.
- Billing, pricing and payment automation as first-class bounded contexts.
- Full analytics warehouse model and BI dashboards.
- Exact Kubernetes manifests and infrastructure-as-code implementation.
- Detailed vendor-specific schemas for Keycloak, SpiceDB, Temporal or YDB internals.

## 1.4 Primary architectural thesis

> **PADS-THESIS-001:** M8 Platform should be designed as a set of domain-owned services where each service owns its language, invariants, data and contracts. External tools such as Keycloak, SpiceDB, Temporal and YDB are implementation details behind explicit adapters.

# 2. Platform Vision

M8 Platform provides reusable platform capabilities for building and operating business applications: resource management, identity, authentication, authorization, risk control, provisioning and audit. It should be usable as a control plane for internal services and as a foundation for future platform products.

## 2.1 Core platform qualities

| Quality | Meaning for M8 Platform |
| --- | --- |
| API First | Public contracts are designed before implementation. Protobuf and ConnectRPC are the primary service contract surfaces. |
| Domain First | The domain model and ubiquitous language drive service boundaries, not database tables or vendor APIs. |
| Clean Architecture | Dependencies point inward. Domain and use cases do not import framework, SDK, transport or persistence types. |
| Event Driven | State changes are published as facts through events. Services maintain local consistency and use events for propagation. |
| Control Plane Oriented | The platform manages desired state, operations, policy and lifecycle rather than only synchronous CRUD. |
| Observable by Design | Every significant request, command, event and operation is traceable, measurable and auditable. |
| AI Native | Requirements and architecture constraints are designed to be converted into structured prompts for controlled AI-assisted development. |

## 2.2 Platform resource hierarchy

```text
Organization
└── Workspace
    └── Project
        └── Service
```

Organization is the top-level tenant, security and administrative boundary. Workspace groups related projects. Project is the main isolation and operational scope. Service is a registered platform capability or application service within a project.

## 2.3 Strategic direction

- Start with Resource Manager as the first platform service and canonical owner of Organization, Workspace and Project.
- Separate identity, authentication and access instead of merging them into one large IAM service.
- Use Keycloak for identity-provider and authentication infrastructure where suitable, but keep M8 Authentication domain model independent from Keycloak internals.
- Use SpiceDB as the authorization graph engine, but keep M8 Access domain language independent from SpiceDB tuples.
- Use Temporal for long-running orchestration, but keep workflow engine concepts outside the domain model.
- Use YDB as the application system of record and YDB Topics as the primary event stream. Temporal persistence is separate and must not use YDB as Temporal storage.

# 3. Design Goals

| ID | Goal | Specification |
| --- | --- | --- |
| DG-001 | Single domain owner | Every domain concept has exactly one owning bounded context and one owning service. |
| DG-002 | Independent service evolution | Services should evolve, deploy and scale independently when their contracts remain compatible. |
| DG-003 | Explicit boundaries | All cross-service dependencies must be represented by APIs, events or anti-corruption gateways. |
| DG-004 | Local consistency | A service transaction may only protect data owned by that service. Global consistency is eventual. |
| DG-005 | No distributed transactions | The platform must not rely on distributed database transactions across services. |
| DG-006 | Long-running lifecycle operations | Mutating operations that may outlive a request should return a long-running Operation. |
| DG-007 | Auditability | Every security-relevant and state-changing operation must create an audit trail. |
| DG-008 | Idempotency | Mutating commands must support idempotency where retries are expected. |
| DG-009 | Versioned contracts | APIs, event schemas and error codes must evolve with backward compatibility. |
| DG-010 | Domain isolation from vendors | Keycloak, SpiceDB, Temporal, YDB and Kubernetes must not leak into domain entities. |
| DG-011 | Traceability to code | Every implemented feature should trace back to a requirement, scenario, contract and structured prompt. |
| DG-012 | Security context propagation | Actor, subject, project, client, request and risk context must be explicit. |
| DG-013 | Observable operations | Requests, operations, events and workflows must expose trace and correlation identifiers. |
| DG-014 | Testable architecture | Architecture constraints must be checkable through tests, static rules or review prompts. |
| DG-015 | Small domain aggregates | Aggregates protect invariants but should not become service-wide transaction containers. |

# 4. Architecture Principles

| ID | Principle | Rule |
| --- | --- | --- |
| AP-001 | Database per service | A service MUST NOT read or write another service database. |
| AP-002 | Owned domain model | A service MUST expose its domain language through contracts and MUST NOT expose persistence models. |
| AP-003 | Clean dependency rule | Adapters depend on use cases; use cases depend on domain; domain depends on nothing external. |
| AP-004 | Anti-corruption layer | External systems MUST be wrapped by adapters and translated into M8 concepts. |
| AP-005 | Outbox for event publication | Events that represent committed state changes MUST be stored atomically with the state change. |
| AP-006 | Inbox for event consumption | Consumers SHOULD deduplicate and record processed integration events. |
| AP-007 | Commands are not events | Commands express intent; events express facts that already happened. |
| AP-008 | Optimistic locking | Aggregates SHOULD use version checks for concurrent mutation. |
| AP-009 | Idempotent mutation APIs | Mutation APIs SHOULD accept idempotency keys and return the same result on retry. |
| AP-010 | Explicit state machines | Lifecycle resources MUST have documented allowed state transitions. |
| AP-011 | Long-running operations | Mutations that require orchestration SHOULD return Operation rather than blocking. |
| AP-012 | Backward compatibility | Public fields MUST NOT be repurposed with incompatible meaning. |
| AP-013 | Security before mutation | Authorization and risk checks MUST happen before state mutation when applicable. |
| AP-014 | Audit after decision | Security-relevant decisions and mutations MUST emit audit events. |
| AP-015 | No hidden global state | Request context MUST include explicit tenant/resource scope and correlation identifiers. |
| AP-016 | Stable error taxonomy | Services MUST map domain errors to a shared error model. |
| AP-017 | Generated code is subordinate | AI-generated or generated code MUST obey PADS, ADRs, contracts and tests. |
| AP-018 | YDB Topics as platform event stream | M8 SHOULD use YDB Topics for internal domain/integration event propagation unless an ADR states otherwise. |

# 5. Ubiquitous Language

| Term | Definition | Owner Context |
| --- | --- | --- |
| Organization | Top-level tenant and administrative boundary. | Resource Manager |
| Workspace | Logical grouping of projects within an organization. | Resource Manager |
| Project | Main isolation and operational scope for resources and services. | Resource Manager |
| Service | Registered application or platform capability inside a project. | Resource Manager |
| User Pool | Isolated container of users and identity configuration. | Identity |
| User | Human or system identity represented inside a user pool. | Identity |
| Group | Collection of users used for membership and access assignment. | Identity |
| Membership | Association of a subject with an organization, workspace or project. | Identity / Resource Manager boundary |
| Subject | Entity that can authenticate or receive permissions: user, service account, group or external subject. | Shared language |
| Client | Application allowed to start authentication and receive handoff results. | Authentication |
| Authentication Transaction | Stateful process of verifying a subject for a client. | Authentication |
| Challenge | Concrete authentication step such as OTP, approval, WebAuthn or CIBA. | Authentication |
| Assurance Level | Required or achieved strength of authentication. | Authentication / Risk Decision |
| Permission | Named action allowed against a resource type or instance. | Access |
| Relationship | Graph relation between subject, role and resource. | Access |
| Risk Signal | Input used to evaluate risk: device, IP, velocity, geo, client or behavior. | Risk Decision |
| Decision | Result of policy evaluation: ALLOW, DENY, CHALLENGE or REVIEW. | Risk Decision |
| Managed Resource | External or internal resource whose desired state is controlled by the platform. | Provisioning |
| Operation | Long-running operation resource representing asynchronous mutation progress. | Common Operation |
| Audit Event | Immutable record of a significant action or decision. | Audit |

## 5.1 Naming rules

- Use Resource Manager, not Structure, as the official context and service name.
- Do not use tenant as a primary resource hierarchy term; use Organization.
- Do not collapse identity, authentication and access into one overloaded IAM context.
- Use Service for registered platform/application capability under Project.
- Use Operation for long-running progress and keep resource state separate from operation state.

# 6. Business Capability Map

| Capability ID | Capability | Description |
| --- | --- | --- |
| CAP-RM | Resource Management | Manage Organization, Workspace, Project and Service hierarchy. |
| CAP-ID | Identity Management | Manage user pools, users, groups, memberships and external identities. |
| CAP-AUTHN | Authentication | Verify subject identity through configured providers and challenges. |
| CAP-AUTHZ | Access Management | Manage roles, relationships and permission checks. |
| CAP-RISK | Risk Decision | Evaluate contextual risk and determine required security action. |
| CAP-PROV | Provisioning | Create, reconcile and delete managed resources. |
| CAP-AUD | Audit | Store immutable audit trail and support search/export/retention. |
| CAP-OPS | Operations | Expose standard long-running operation lifecycle. |
| CAP-OBS | Observability | Trace, measure and log platform activity. |
| CAP-GOV | Architecture Governance | Trace requirements through contracts, prompts, code and tests. |

## 6.1 Capability to service mapping

| Capability | Primary Service | Supporting Services |
| --- | --- | --- |
| Resource Management | m8-resource-manager | m8-access, m8-audit |
| Identity Management | m8-identity | m8-access, m8-audit |
| Authentication | m8-authentication | m8-identity, m8-risk-decision, m8-access, m8-audit, Keycloak adapter |
| Access Management | m8-access | SpiceDB adapter, m8-resource-manager, m8-identity |
| Risk Decision | m8-risk-decision | m8-authentication, m8-audit |
| Provisioning | m8-provisioning | Temporal adapter, Kubernetes/cloud adapters, m8-resource-manager |
| Audit | m8-audit | All services |
| Operations | Common operation package + service-specific operation owner | Temporal adapter, m8-audit |

# 7. Domain Model

## 7.1 Platform aggregate overview

```text
Resource Manager
  Organization
  Workspace
  Project
  ServiceRegistration

Identity
  UserPool
  User
  Group
  Membership
  ExternalIdentity

Authentication
  Client
  AuthenticationTransaction
  AuthenticationChallenge
  AuthenticationSession

Access
  AuthorizationModel
  Role
  RoleBinding
  AccessRelationship
  PermissionCheck

Risk Decision
  RiskAssessment
  RiskSignal
  DecisionPolicy
  RiskDecision

Provisioning
  ResourceDefinition
  ResourceRequest
  ManagedResource
  Reconciliation
  Driver

Audit
  AuditEvent
  RetentionPolicy
  ExportJob

Common Operation
  Operation
  OperationMetadata
  OperationProgress
  OperationError
```

## 7.2 Core hierarchy invariants

| Invariant ID | Invariant |
| --- | --- |
| INV-RM-001 | Workspace MUST belong to exactly one Organization. |
| INV-RM-002 | Project MUST belong to exactly one Workspace. |
| INV-RM-003 | ServiceRegistration MUST belong to exactly one Project. |
| INV-RM-004 | Resource identity and parent scope MUST NOT change through ordinary update operations. |
| INV-RM-005 | Parent deletion MUST be blocked while active child resources exist unless a governed cascading operation is explicitly started. |
| INV-RM-006 | Resource lifecycle state MUST be separate from Operation lifecycle state. |

## 7.3 Authentication state model

```text
CREATED
  → CHALLENGE_REQUIRED
  → CHALLENGE_PENDING
  → AUTHENTICATED
  → HANDOFF_CREATED
  → COMPLETED

Terminal alternatives:
  CANCELLED
  EXPIRED
  FAILED
```

## 7.4 Provisioning state model

```text
REQUESTED
  → ACCEPTED
  → PLANNING
  → APPLYING
  → RECONCILING
  → READY

Terminal / exceptional alternatives:
  FAILED
  DELETING
  DELETED
  DRIFT_DETECTED
  SUSPENDED
```

## 7.5 Operation state model

```text
PENDING
  → RUNNING
  → SUCCEEDED

Terminal alternatives:
  FAILED
  CANCELLED
  EXPIRED
```

# 8. Context Map

```text
                         ┌──────────────────────┐
                         │   Resource Manager   │
                         │ Org/Workspace/Project│
                         └──────────┬───────────┘
                                    │
          ┌─────────────────────────┼──────────────────────────┐
          │                         │                          │
          ▼                         ▼                          ▼
┌────────────────────┐    ┌────────────────────┐     ┌────────────────────┐
│      Identity      │    │       Access       │     │    Provisioning    │
│ Users/Pools/Groups │    │ Roles/Relationships│     │ Managed Resources  │
└─────────┬──────────┘    └──────────┬─────────┘     └─────────┬──────────┘
          │                          │                         │
          ▼                          │                         │
┌────────────────────┐               │                         │
│   Authentication   │◄──────────────┘                         │
│ Challenges/Sessions│                                         │
└─────────┬──────────┘                                         │
          │                                                    │
          ▼                                                    │
┌────────────────────┐                                         │
│   Risk Decision    │◄────────────────────────────────────────┘
│ Policy/Risk/StepUp │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│       Audit        │
│ Immutable Events   │
└────────────────────┘
```

## 8.1 Context relationship matrix

| Upstream | Downstream | Relationship | Contract |
| --- | --- | --- | --- |
| Resource Manager | All contexts | Published Language | Resource references, hierarchy events, scope validation API. |
| Identity | Authentication | Customer/Supplier | Subject resolution, user status, external identity mapping. |
| Identity | Access | Published Language | Subject and group references; user lifecycle events. |
| Access | Authentication | Open Host Service | Permission check API and optional access explanation. |
| Risk Decision | Authentication | Customer/Supplier | Risk decision API requiring ALLOW/DENY/CHALLENGE/REVIEW. |
| Resource Manager | Provisioning | Customer/Supplier | Project scope and service registration validation. |
| Provisioning | Risk Decision | Published Language | Provisioning risk signals and resource action context. |
| All contexts | Audit | Conformist to common audit schema | AuditEvent.v1. |

## 8.2 Anti-corruption layer rules

- M8 Authentication must translate Keycloak-specific concepts into M8 AuthenticationTransaction, Challenge and Session concepts.
- M8 Access must translate M8 RoleBinding and AccessRelationship into SpiceDB relationship operations inside the adapter layer.
- M8 Provisioning must translate ResourceRequest and ManagedResource into Kubernetes, cloud or infrastructure-provider commands inside drivers.
- Temporal workflow and activity types must not appear in domain entities or use case request models.
- YDB table rows and YDB SDK models must not appear in domain entities or protobuf contracts.

# 9. Bounded Context Specifications

## 9.1 Resource Manager (CTX-RM)

| Property | Specification |
| --- | --- |
| Service | m8-resource-manager |
| Purpose | Canonical owner of Organization, Workspace, Project and ServiceRegistration. It is the first platform service and the source of truth for resource hierarchy. |
| Owns | Organization, Workspace, Project, ServiceRegistration, resource lifecycle state, labels, version, resource references. |
| Does not own | Users, credentials, roles, permissions, authentication sessions, risk scoring, external infrastructure resources. |
| Primary commands | CreateOrganization, UpdateOrganization, DeleteOrganization, CreateWorkspace, CreateProject, RegisterService, MoveResource through governed operation if allowed. |
| Primary events | OrganizationCreated, WorkspaceCreated, ProjectCreated, ServiceRegistered, ResourceStateChanged, ResourceDeleted. |
| Integrations | Access for authorization; Audit for audit events; YDB for persistence; YDB Topics for event publication; Temporal for long-running lifecycle operations. |

## 9.2 Identity (CTX-ID)

| Property | Specification |
| --- | --- |
| Service | m8-identity |
| Purpose | Manage identities independently from authentication mechanics and authorization decisions. |
| Owns | UserPool, User, Group, Membership, ExternalIdentity, user lifecycle status, subject references. |
| Does not own | Authentication challenges, token issuing, access graph decisions, risk policy decisions. |
| Primary commands | CreateUserPool, CreateUser, DisableUser, LinkExternalIdentity, CreateGroup, AddUserToGroup, AssignMembership. |
| Primary events | UserPoolCreated, UserCreated, UserDisabled, ExternalIdentityLinked, GroupCreated, MembershipChanged. |
| Integrations | Authentication for subject resolution; Access for subject/group publication; Audit for state changes. |

## 9.3 Authentication (CTX-AUTHN)

| Property | Specification |
| --- | --- |
| Service | m8-authentication |
| Purpose | Start, execute, complete, cancel and expire authentication processes for clients and subjects. |
| Owns | Client, AuthenticationTransaction, AuthenticationChallenge, AuthenticationSession, requested/achieved assurance level. |
| Does not own | User profile ownership, role assignment, policy graph, persistent audit storage, direct resource hierarchy mutation. |
| Primary commands | StartAuthentication, SelectChallenge, ResendChallenge, CompleteChallenge, CancelAuthentication, CreateHandoff, RefreshSessionDecision. |
| Primary events | AuthenticationStarted, ChallengeRequired, ChallengeCompleted, AuthenticationCompleted, AuthenticationFailed, AuthenticationCancelled. |
| Integrations | Identity for subject resolution; Risk Decision for step-up/deny/challenge; Access for client/resource permission; Keycloak adapter; Audit. |

## 9.4 Access (CTX-ACC)

| Property | Specification |
| --- | --- |
| Service | m8-access |
| Purpose | Manage authorization language and evaluate whether a subject can perform an action on a resource. |
| Owns | AuthorizationModel, Permission, Role, RoleBinding, AccessRelationship, permission check explanation. |
| Does not own | Authentication proof, user credentials, resource hierarchy ownership, audit storage. |
| Primary commands | CreateRole, UpdateRole, BindRole, RemoveRoleBinding, WriteRelationship, DeleteRelationship, CheckPermission, ExplainPermission. |
| Primary events | RoleCreated, RoleBindingChanged, AccessRelationshipChanged, AuthorizationModelPublished. |
| Integrations | SpiceDB adapter; Resource Manager and Identity published languages; Audit. |

## 9.5 Risk Decision (CTX-RISK)

| Property | Specification |
| --- | --- |
| Service | m8-risk-decision |
| Purpose | Evaluate risk signals and policies to determine allow, deny, challenge or manual review outcomes. |
| Owns | RiskAssessment, RiskSignal, DecisionPolicy, Decision, challenge requirement, risk explanation. |
| Does not own | Executing authentication challenge, managing users, changing access graph, provisioning resources directly. |
| Primary commands | EvaluateAuthenticationRisk, EvaluateAccessRisk, EvaluateProvisioningRisk, CreatePolicy, UpdatePolicy, SimulateDecision. |
| Primary events | RiskAssessmentCreated, RiskDecisionMade, PolicyChanged, RiskSignalObserved. |
| Integrations | Authentication, Provisioning, Access, Audit and optional external signal providers. |

## 9.6 Provisioning (CTX-PROV)

| Property | Specification |
| --- | --- |
| Service | m8-provisioning |
| Purpose | Manage desired state and lifecycle of platform-managed resources through reconciliation. |
| Owns | ResourceDefinition, ResourceRequest, ManagedResource, Reconciliation, Driver, desired and observed state. |
| Does not own | Organization hierarchy ownership, identity ownership, authorization graph ownership. |
| Primary commands | CreateManagedResource, UpdateDesiredState, DeleteManagedResource, ReconcileResource, RegisterDriver, DetectDrift. |
| Primary events | ResourceRequested, ProvisioningStarted, ResourceProvisioned, DriftDetected, ResourceDeleted, ProvisioningFailed. |
| Integrations | Temporal for orchestration; drivers for Kubernetes/cloud/Yandex/self-hosted; Resource Manager; Risk Decision; Audit. |

## 9.7 Audit (CTX-AUD)

| Property | Specification |
| --- | --- |
| Service | m8-audit |
| Purpose | Receive, validate, store, search, export and retain immutable audit events. |
| Owns | AuditEvent, AuditActor, AuditTarget, AuditContext, AuditChangeSet, RetentionPolicy, ExportJob. |
| Does not own | Business-state mutation in source services, authorization decisions, risk scoring. |
| Primary commands | AppendAuditEvent, SearchAuditEvents, CreateExportJob, ConfigureRetention, VerifyIntegrity. |
| Primary events | AuditEventAppended, AuditExportCreated, RetentionPolicyChanged. |
| Integrations | All services as producers; YDB/YDB Topics; object storage for export if needed. |

# 10. Shared Kernel and Common Contracts

The shared kernel must remain small. It contains stable primitives required across services. It must not become a dumping ground for domain logic.

| Package | Allowed contents | Forbidden contents |
| --- | --- | --- |
| m8.platform.common.resource.v1 | ResourceReference, ResourceName, ParentReference, labels, lifecycle enums used in public contracts. | Resource Manager business rules or persistence models. |
| m8.platform.common.operation.v1 | OperationMetadata, OperationProgress, action string, stage, progress percent, timestamps, request correlation. | Service-specific workflow logic. |
| m8.platform.common.audit.v1 | Common audit actor, target, context and change set envelope. | Service-specific event interpretation. |
| m8.platform.common.error.v1 | Error code taxonomy, localized-safe message fields, retryability and violation details. | Transport-specific exception types. |
| m8.platform.common.context.v1 | Request, actor, subject, client, project, correlation and risk context references. | Authentication or authorization business decisions. |

## 10.1 OperationMetadata rule

> **COMMON-OP-001:** OperationMetadata belongs to m8.platform.common.operation.v1. Its action field is a string, not a fixed enum, so every service can define action names without changing the shared contract.

# 11. Data Ownership

| Data / Concept | Owner service | Consumers | Replication method |
| --- | --- | --- | --- |
| Organization, Workspace, Project | m8-resource-manager | All services | API lookups and Resource events. |
| ServiceRegistration | m8-resource-manager | Authentication, Access, Provisioning, Audit | API and ServiceRegistered event. |
| UserPool, User, Group | m8-identity | Authentication, Access, Audit | API and Identity events. |
| Client, AuthenticationTransaction | m8-authentication | Risk Decision, Audit | API and Authentication events. |
| Role, RoleBinding, Relationship | m8-access | Authentication, Resource Manager, UI/BFF | Access API, SpiceDB adapter and Access events. |
| RiskAssessment, DecisionPolicy | m8-risk-decision | Authentication, Provisioning, Audit | Risk API and Risk events. |
| ManagedResource, Reconciliation | m8-provisioning | Resource Manager, Risk, Audit | Provisioning API and events. |
| AuditEvent | m8-audit | Compliance, admin UI, export | Append-only audit API and event stream. |

## 11.1 Data ownership prohibitions

- A service MUST NOT join tables from another service database.
- A service MUST NOT update another service resource state directly.
- A service MAY store a projection of another service data when the source and freshness are explicit.
- A projection MUST be treated as stale unless a contract states otherwise.
- External analytics stores may consume events but must not become the source of truth for operational decisions unless explicitly designed as such.

# 12. API Design Rules

| Rule ID | Rule |
| --- | --- |
| API-001 | Public service contracts SHOULD be Protobuf-first and exposed through ConnectRPC. |
| API-002 | Validation SHOULD be expressed through buf.validate / Protovalidate where possible. |
| API-003 | Mutation APIs SHOULD include request_id or idempotency_key when clients may retry. |
| API-004 | Long-running mutations SHOULD return google.longrunning.Operation or a compatible operation envelope. |
| API-005 | Public API messages MUST NOT expose persistence table names, YDB SDK types or vendor-specific structures. |
| API-006 | List APIs MUST support pagination and stable ordering. |
| API-007 | Filter syntax MUST be documented and validated. |
| API-008 | Errors MUST use shared error taxonomy and include machine-readable codes. |
| API-009 | Security context MUST be derived from trusted authentication/authorization middleware, not arbitrary user-provided fields. |
| API-010 | Breaking changes require a new API version. |

## 12.1 Canonical mutation pattern

```text
Client
  → API Adapter
  → Authentication / AuthGuard
  → Access Check
  → Risk Decision where applicable
  → Application Use Case
  → Domain Aggregate
  → Repository transaction
  → Outbox
  → Operation / Response
  → Audit
```

# 13. Event Design Rules

| Rule ID | Rule |
| --- | --- |
| EVT-001 | Events MUST describe facts in past tense. |
| EVT-002 | Commands MUST NOT be published as events. |
| EVT-003 | Events MUST include event_id, event_type, occurred_at, producer, schema_version, correlation_id and causation_id. |
| EVT-004 | Events MUST include resource references when a resource is affected. |
| EVT-005 | Events SHOULD be published through the Outbox pattern. |
| EVT-006 | Consumers SHOULD be idempotent and use Inbox or equivalent deduplication. |
| EVT-007 | Event schemas MUST be backward compatible within a major version. |
| EVT-008 | Audit events and domain events are related but not identical; domain events describe business facts, audit events describe accountability facts. |

## 13.1 Standard event envelope

```text
event_id: string
schema_version: string
event_type: string
producer_service: string
occurred_at: timestamp
published_at: timestamp
correlation_id: string
causation_id: string
actor: AuditActor | optional
resource: ResourceReference | optional
payload: service_specific_message
```

# 14. Integration and Consistency Model

## 14.1 Consistency classes

| Class | Use when | Allowed mechanism |
| --- | --- | --- |
| Strong local consistency | Single service owns all affected aggregates and invariants. | Single YDB transaction within one service boundary. |
| Read-your-write within service | Client needs immediate confirmation of mutation in same service. | Return updated view or Operation state from owner service. |
| Eventual consistency | Other services need to learn about committed facts. | YDB Topics + Outbox/Inbox + projections. |
| Orchestrated consistency | Multi-step lifecycle across systems. | Temporal workflow + idempotent activities + compensations. |
| External eventual consistency | Vendor or infrastructure API changes asynchronously. | Provisioning reconciliation with desired/observed state. |

## 14.2 Integration patterns

- Synchronous API is used for immediate decisions: permission check, risk decision, subject resolution or operation status.
- Events are used to distribute facts after commit.
- Temporal is used for long-running, retryable, compensatable workflows.
- Redis may be used for cache, leases or rate limiting, but it is not a system of record.
- YDB is the application system of record for service-owned state.

# 15. Security Architecture

M8 Platform follows explicit security context propagation and zero-trust-oriented checks. A request is not trusted only because it came from an internal network. The service must know the actor, subject, client, resource scope, requested action and relevant risk context.

| Security concern | Primary owner | Specification |
| --- | --- | --- |
| Authentication | m8-authentication | Verifies subject identity and achieves required assurance level. |
| Identity lifecycle | m8-identity | Determines whether the subject exists, is active and belongs to a user pool. |
| Authorization | m8-access | Determines whether the subject can perform action on resource. |
| Risk step-up | m8-risk-decision | Determines whether additional challenge or denial is required. |
| Auditability | m8-audit | Records who did what, when, where, why and with which decision context. |

## 15.1 Authentication model

- Primary interactive/login flow uses CIBA where appropriate.
- Refresh uses Keycloak refresh_token when available.
- When refresh fails, a new CIBA authentication is started instead of silently restoring the previous session.
- Step-up starts a new authentication transaction with a higher requested assurance level.
- Authentication supports challenge types such as OTP, approval, Mobile ID, WebAuthn, OIDC, SAML and password when allowed by policy.

## 15.2 Authorization model

- M8 Access owns business authorization language: permissions, roles and relationships.
- SpiceDB is the graph evaluation engine behind the Access adapter.
- Domain logic must not construct SpiceDB tuple strings directly.
- Every mutation must define which permission is required before implementation.

## 15.3 Risk decision model

- Risk Decision can return ALLOW, DENY, CHALLENGE or REVIEW.
- CHALLENGE includes the required assurance level or challenge class.
- Risk Decision must explain the reason in machine-readable terms suitable for audit and debugging.
- Authentication executes the challenge; Risk Decision only decides that challenge is required.

# 16. Long Running Operations

Long-running operations are first-class resources. Mutation APIs that trigger asynchronous work should return an operation immediately and allow clients to get, wait, cancel or observe progress.

## 16.1 Operation contract

```text
operation:
  name: operations/{operation_id}
  done: boolean
  metadata:
    type: m8.platform.common.operation.v1.OperationMetadata
    action: string
    target_resource: ResourceReference
    progress:
      stage: string
      message: string
      percent: int32
    request_id: string
    correlation_id: string
    create_time: timestamp
    update_time: timestamp
  result: service_specific_result | optional
  error: OperationError | optional
```

## 16.2 Operation rules

- Operation state is not the same as resource state.
- Operation cancellation is a request, not a guaranteed immediate rollback.
- Operation metadata must be safe to expose to an authorized caller.
- Each operation must have a stable action string such as resource_manager.projects.create.
- Temporal workflow identifiers may be stored in service persistence but must not leak into public Operation contracts.

# 17. Error Model

| Error category | Examples | HTTP/gRPC mapping intent |
| --- | --- | --- |
| Validation | INVALID_ARGUMENT, FIELD_VIOLATION | Client sent malformed or invalid request. |
| Not found | RESOURCE_NOT_FOUND, USER_NOT_FOUND, CLIENT_NOT_FOUND | Requested resource does not exist or is not visible. |
| Conflict | VERSION_CONFLICT, IDEMPOTENCY_CONFLICT, RESOURCE_ALREADY_EXISTS | Request conflicts with current state. |
| Precondition | RESOURCE_NOT_ACTIVE, CLIENT_DISABLED, DELETE_BLOCKED | Request is valid but cannot be applied in current state. |
| Permission | PERMISSION_DENIED, ACCESS_CHECK_FAILED | Actor is not allowed to perform action. |
| Risk | RISK_DENIED, STEP_UP_REQUIRED, MANUAL_REVIEW_REQUIRED | Risk policy changed the outcome. |
| Dependency | IDENTITY_UNAVAILABLE, RISK_DECISION_UNAVAILABLE, PROVIDER_UNAVAILABLE | Required dependency is unavailable. |

## 17.1 Error response rules

- Every error must include a stable machine-readable code.
- User-facing text must not leak secrets, internal identifiers or provider error details.
- Retryable errors must be explicitly marked as retryable.
- Validation errors should include field-level violations.
- Domain errors must be mapped at the adapter boundary, not thrown as transport exceptions from domain code.

# 18. Observability

| Signal | Required content |
| --- | --- |
| Logs | timestamp, severity, service, operation, request_id, correlation_id, actor/resource references where safe, error code. |
| Traces | incoming request span, use case span, repository span, external dependency span, event publication span. |
| Metrics | request count, latency, error rate, operation duration, event lag, dependency failure rate, workflow retries. |
| Audit | actor, action, target, decision, context, before/after change set when applicable. |

## 18.1 Correlation identifiers

```text
request_id       unique request identifier
correlation_id   end-to-end business flow identifier
causation_id     event or command that caused the current action
operation_id     long-running operation identifier
trace_id         distributed trace identifier
actor_id         authenticated actor where known
subject_id       identity subject where relevant
project_id       resource scope where relevant
```

# 19. Quality Attributes

| Attribute | Baseline target / rule |
| --- | --- |
| Availability | Core authentication, access and resource-read APIs SHOULD target high availability. Service-specific SLOs are defined later. |
| Latency | Permission checks and simple reads SHOULD be optimized for low latency. Long-running work MUST use Operation instead of blocking. |
| Scalability | Event consumers, API services and workers SHOULD scale horizontally. |
| Reliability | Retries must be idempotent. External dependency failures must be bounded with timeouts and circuit breakers where appropriate. |
| Maintainability | Clean Architecture, small bounded contexts, explicit contracts and ADRs are mandatory governance mechanisms. |
| Security | Authentication, authorization, risk and audit are first-class architecture concerns. |
| Recoverability | Services should support replaying events or rebuilding projections where feasible. |
| Testability | Domain and use cases must be testable without running external systems. |

# 20. Requirements Distribution

Requirements are distributed by ownership. A requirement belongs to the service that owns the affected domain invariant. Supporting services may be listed as dependencies but must not become hidden owners.

## 20.1 Requirement identifier families

| Prefix | Owner | Examples |
| --- | --- | --- |
| PLT-* | Platform-level governance | PLT-NFR-001, PLT-SEC-002 |
| RM-* | Resource Manager | RM-FR-001 Create Organization, RM-FR-020 Register Service |
| ID-* | Identity | ID-FR-001 Create User Pool, ID-FR-010 Link External Identity |
| AUTH-* | Authentication | AUTH-FR-001 Start Authentication, AUTH-FR-017 Re-authentication after refresh failure |
| ACC-* | Access | ACC-FR-001 Check Permission, ACC-FR-012 Explain Access Decision |
| RISK-* | Risk Decision | RISK-FR-001 Evaluate Authentication Risk, RISK-FR-011 Simulate Policy |
| PROV-* | Provisioning | PROV-FR-001 Create Managed Resource, PROV-FR-020 Detect Drift |
| AUD-* | Audit | AUD-FR-001 Append Audit Event, AUD-FR-012 Export Audit Events |
| OPS-* | Operations | OPS-FR-001 Get Operation, OPS-FR-004 Cancel Operation |

## 20.2 Initial service requirement allocation

| Requirement ID | Requirement | Owner | Notes |
| --- | --- | --- | --- |
| RM-FR-001 | Create Organization | m8-resource-manager | Organization aggregate, Operation, audit event. |
| RM-FR-002 | Create Workspace | m8-resource-manager | Workspace under Organization, version checks. |
| RM-FR-003 | Create Project | m8-resource-manager | Project under Workspace, Operation return. |
| RM-FR-004 | Register Service | m8-resource-manager | ServiceRegistration inside Project. |
| ID-FR-001 | Create User Pool | m8-identity | UserPool aggregate and audit. |
| ID-FR-002 | Create User | m8-identity | User aggregate, external identity optional. |
| ID-FR-003 | Disable User | m8-identity | User lifecycle event consumed by Authentication and Access. |
| AUTH-FR-001 | Start Authentication | m8-authentication | AuthenticationTransaction creation with risk and identity dependencies. |
| AUTH-FR-017 | Re-authentication after refresh failure | m8-authentication | Start new CIBA flow; do not restore previous session silently. |
| ACC-FR-001 | Check Permission | m8-access | SpiceDB-backed decision through M8 Access language. |
| ACC-FR-002 | Bind Role | m8-access | RoleBinding and AccessRelationship write. |
| RISK-FR-001 | Evaluate Authentication Risk | m8-risk-decision | ALLOW/DENY/CHALLENGE/REVIEW decision. |
| PROV-FR-001 | Create Managed Resource | m8-provisioning | Desired state + Temporal orchestration + Operation. |
| PROV-FR-002 | Reconcile Managed Resource | m8-provisioning | Desired/observed state reconciliation. |
| AUD-FR-001 | Append Audit Event | m8-audit | Immutable storage and integrity metadata. |
| OPS-FR-001 | Get Operation | Service-specific owner + common operation contract | Read operation by authorized caller. |

# 21. Traceability Model

```text
Business Capability
  → Platform Requirement
    → Context Requirement
      → Service Requirement
        → Use Case
          → API / Event Contract
            → Structured Prompt
              → Code Change
                → Unit Test
                  → Contract Test
                    → Acceptance Test
                      → Release Evidence
```

## 21.1 Traceability record

```text
traceability:
  capability: CAP-AUTHN
  platform_requirement: PLT-AUTHN-001
  context_requirement: AUTH-FR-017
  service: m8-authentication
  use_case: UC-AUTH-009
  contracts:
    - auth.v1.AuthenticationService.StartAuthentication
    - m8.authentication.AuthenticationStarted.v1
  structured_prompts:
    - SP-AUTH-017-01
  tests:
    - UT-AUTH-017-01
    - CT-AUTH-017-01
    - AT-AUTH-017-01
  adr:
    - ADR-0012-keycloak-ciba
    - ADR-0021-operation-model
```

# 22. SPDD Mapping

SPDD means Structured-Prompt-Driven Development for M8 Platform. Structured Prompts are not casual chat prompts. They are versioned engineering artifacts that bind requirements, domain language, contracts, constraints, tests and review rules.

## 22.1 SPDD artifact hierarchy

```text
/docs/07-spdd
├── constitution
│   └── M8-SPDD-CONSTITUTION.md
├── contexts
│   ├── resource-manager.prompt.md
│   ├── identity.prompt.md
│   ├── authentication.prompt.md
│   ├── access.prompt.md
│   ├── risk-decision.prompt.md
│   ├── provisioning.prompt.md
│   └── audit.prompt.md
├── features
├── tasks
├── reviews
└── templates
```

## 22.2 Prompt levels

| Level | Artifact | Purpose |
| --- | --- | --- |
| L1 | Constitution Prompt | Global architecture, stack, dependency rules, security rules, testing and prohibited shortcuts. |
| L2 | Context Prompt | Bounded-context language, responsibilities, aggregates, dependencies and forbidden ownership. |
| L3 | Feature Prompt | One business feature with requirements, scenarios, contracts and acceptance criteria. |
| L4 | Task Prompt | Small implementation unit with allowed files, constraints and tests. |
| L5 | Review Prompt | Independent review against PADS, ADRs, contracts, tests and architecture rules. |

## 22.3 Structured Prompt template

```text
spdd_version: "1.0"
metadata:
  id: SP-AUTH-017-01
  title: Implement re-authentication after refresh token failure
  context: Authentication
  service: m8-authentication
  type: implementation
  status: draft
traceability:
  requirements:
    - AUTH-FR-017
  use_cases:
    - UC-AUTH-009
  contracts:
    - auth.v1.AuthenticationService.StartAuthentication
    - m8.authentication.AuthenticationStarted.v1
objective: >
  Implement StartAuthentication behavior for creating a new CIBA authentication
  transaction when refresh token cannot be used.
scope:
  include:
    - application use case
    - domain aggregate behavior
    - repository transaction
    - outbox event
    - unit tests
  exclude:
    - protobuf contract changes
    - Keycloak adapter implementation
    - Risk Decision service implementation
constraints:
  architecture:
    - Clean Architecture dependency rule
    - domain must not import Keycloak, YDB, Temporal, ConnectRPC or SpiceDB types
  data:
    - use service-owned repository only
    - store outbox atomically with aggregate state
  security:
    - check client state
    - resolve subject through IdentityGateway
    - call RiskDecisionGateway before mutation where required
acceptance:
  - AUTH-FR-017-AC-01
  - AUTH-FR-017-AC-02
tests:
  unit:
    - duplicate idempotency key returns existing transaction
    - disabled client is rejected
    - risk denial prevents transaction creation
  integration:
    - aggregate and outbox are committed atomically
review:
  must_check:
    - no forbidden imports
    - no direct Keycloak calls from use case
    - no event publish before transaction commit
    - error codes mapped through shared model
```

## 22.4 SPDD constitution rules

- A prompt MUST reference requirement IDs and acceptance criteria IDs.
- A prompt MUST declare allowed and forbidden areas of change.
- A prompt MUST include architecture constraints inherited from PADS.
- A prompt MUST define tests before implementation.
- A generated change MUST be reviewed by a Review Prompt before merge.
- A prompt MUST NOT ask an agent to invent service boundaries or bypass ADRs.

# 23. Architecture Governance

## 23.1 ADR policy

Every significant architectural decision must be captured as an ADR. PADS defines the current baseline; ADRs explain why a decision was made and when it supersedes or refines a baseline rule.

| ADR trigger | Examples |
| --- | --- |
| Technology choice | Choosing YDB Topics over Kafka for internal event streaming. |
| Boundary change | Moving Membership ownership between Identity and Resource Manager. |
| Contract change | Introducing a new public API version. |
| Consistency change | Moving from synchronous decision to event-driven projection. |
| Security change | Changing primary authentication flow or assurance-level rules. |

## 23.2 Architecture checks

- Static import checks: domain must not import infrastructure, SDK or transport packages.
- Contract checks: protobuf breaking-change detection through buf.
- Architecture tests: repository and adapter dependencies must point inward.
- Event schema checks: event envelopes must include required metadata.
- SPDD checks: prompts must include traceability, scope and tests.
- Review checks: every feature branch must include requirement coverage evidence.

# 24. Glossary

| Term | Definition |
| --- | --- |
| ACL | Anti-Corruption Layer. Translation layer isolating domain model from external systems. |
| ADR | Architecture Decision Record. A short document explaining a significant architectural decision. |
| Bounded Context | A DDD boundary where a model has a specific meaning and owner. |
| CIBA | Client Initiated Backchannel Authentication. Used as a primary authentication flow where applicable. |
| Clean Architecture | Architecture style where dependencies point inward toward use cases and domain. |
| ConnectRPC | HTTP/gRPC-compatible RPC framework used as the primary API transport. |
| LRO | Long Running Operation. Standard asynchronous operation resource. |
| Outbox | Pattern for atomically storing state change and event to publish after commit. |
| Projection | Local copy/read model derived from another owner service events. |
| SPDD | Structured-Prompt-Driven Development. Versioned prompts tied to requirements and architecture constraints. |
| YDB Topics | M8 baseline event stream mechanism for domain and integration events. |


---

# Appendix A. Initial Repository Layout

```text
/cmd
  /resourcemanager-api
  /identity-api
  /authentication-api
  /access-api
  /risk-decision-api
  /provisioning-api
  /audit-api

/internal
  /modules
    /resourcemanager
      /domain
      /application
      /adapter
      /infrastructure
    /identity
    /authentication
    /access
    /riskdecision
    /provisioning
    /audit
  /platform
    /config
    /logger
    /metrics
    /tracing
    /module

/api
  /proto
    /m8/platform/common
    /m8/platform/resourcemanager
    /m8/platform/identity
    /m8/platform/authentication
    /m8/platform/access
    /m8/platform/riskdecision
    /m8/platform/provisioning
    /m8/platform/audit

/docs
  /01-domain
  /02-architecture
  /03-requirements
  /04-contracts
  /05-decisions
  /07-spdd
  /08-validation
```

# Appendix B. Minimal Definition of Done

- Requirement ID and acceptance criteria are linked.
- Contract is defined or explicitly unchanged.
- Domain model change is documented.
- Service owner is clear.
- No forbidden dependency is introduced.
- Outbox/audit behavior is specified for mutations.
- Operation behavior is specified for long-running work.
- Unit, contract and acceptance tests are listed.
- Structured Prompt is created or updated.
- Review Prompt confirms compliance with PADS and ADRs.
