# AGENTS.md

## Project

This repository contains **M8 Platform** — a Go-based modular SaaS / Internal Developer Platform / Control Plane.

M8 Platform is designed as a modular monorepo with many bounded contexts and many independently deployable Go services.

Primary language: **Go**  
Primary IDE: **GoLand**

The codebase must be optimized for:

- long-term maintainability
- clear module ownership
- many services and workers
- explicit architectural boundaries
- testability
- observability
- secure-by-default behavior
- predictable Codex-assisted development

Codex must not optimize the repository for a small single-service application.

---

## Core M8 Domains

M8 Platform contains multiple business modules.

Expected bounded contexts:

```text
M8 Resource Manager   owns organizations, workspaces, projects and resource catalog
M8 Identity           owns users, identities and user pools
M8 Authentication     owns authentication sessions, challenges and authenticators
M8 Access             owns permissions, relationships and authorization checks
M8 Provisioning       owns resource descriptors, provider drivers and reconciliation
M8 Runtime            owns clusters, namespaces, placements and runtime profiles
M8 Delivery           owns deployments, releases, rollouts and rollbacks
M8 Audit              owns audit events and compliance history
M8 Operations         owns long-running operations and operational visibility
```

Each module must have a clear owner, clear responsibility and explicit boundaries.

---

## Modules vs Services

A **module** is a bounded context.

A **service** is a deployment unit.

One module may have many services.

Example:

```text
internal/provisioning/
├── domain/
├── app/
├── adapter/
├── workflow/
├── reconciler/
├── migration/
└── module.go

cmd/
├── m8-provisioning-api/
├── m8-provisioning-worker/
└── m8-provisioning-reconciler/
```

Do not assume one module equals one service.

Do not assume one service owns the entire platform.

---

## Expected Service Types

The platform may contain many independently deployable services:

```text
m8-resource-manager-api
m8-identity-api
m8-authentication-api
m8-authentication-worker
m8-access-api
m8-provisioning-api
m8-provisioning-worker
m8-provisioning-reconciler
m8-runtime-api
m8-delivery-api
m8-audit-consumer
m8-cli
```

Avoid generic services:

```text
m8-core
m8-backend
m8-platform-service
m8-main-service
m8-manager
```

Every service must have a clear responsibility.

---

## Repository Layout

Use a modular monorepo layout that scales to many services.

```text
m8-platform/
├── cmd/
│   ├── m8-resource-manager-api/
│   ├── m8-identity-api/
│   ├── m8-authentication-api/
│   ├── m8-authentication-worker/
│   ├── m8-access-api/
│   ├── m8-provisioning-api/
│   ├── m8-provisioning-worker/
│   ├── m8-provisioning-reconciler/
│   ├── m8-runtime-api/
│   ├── m8-delivery-api/
│   ├── m8-audit-consumer/
│   └── m8-cli/
│
├── internal/
│   ├── platform/
│   ├── sharedkernel/
│   ├── resourcemanager/
│   ├── identity/
│   ├── authentication/
│   ├── access/
│   ├── provisioning/
│   ├── runtime/
│   ├── delivery/
│   └── audit/
│
├── api/
│   └── proto/
│       └── m8/
│           ├── resourcemanager/
│           ├── identity/
│           ├── authentication/
│           ├── access/
│           ├── provisioning/
│           ├── runtime/
│           ├── delivery/
│           └── audit/
│
├── deploy/
│   ├── helm/
│   ├── kustomize/
│   └── local/
│
├── docs/
│   ├── adr/
│   ├── hld/
│   └── rfc/
│
├── tools/
├── Makefile
├── Taskfile.yml
├── go.mod
├── go.sum
└── AGENTS.md
```

Do not create `pkg/common`, `pkg/utils`, `internal/common` or similar catch-all packages.

---

## Module Layout

Each business module should follow the same general structure.

```text
internal/<module>/
├── domain/
│   ├── model/
│   ├── event/
│   ├── policy/
│   ├── service/
│   └── errors.go
│
├── app/
│   ├── command/
│   ├── query/
│   ├── usecase/
│   └── ports/
│
├── adapter/
│   ├── grpc/
│   ├── http/
│   ├── postgres/
│   ├── ydb/
│   ├── temporal/
│   ├── events/
│   └── external/
│
├── workflow/
│   └── temporal workflows if the module owns long-running processes
│
├── reconciler/
│   └── reconciliation loops if the module owns desired/actual state
│
├── migration/
│   └── database migrations owned by this module
│
├── config/
│   └── module-specific configuration
│
├── module.go
└── README.md
```

Not every module needs every folder.

Do not create empty folders only for symmetry.

---

## Clean Architecture Dependency Rule

Dependency direction must be:

```text
adapter -> app -> domain
infra   -> app -> domain
cmd     -> module composition
```

Domain must not import:

```text
database clients
HTTP frameworks
gRPC transport code
Temporal SDK
Keycloak SDK
Kubernetes SDK
SpiceDB SDK
logging frameworks
protobuf-generated transport models
```

Application layer may define ports/interfaces.

Adapters implement those ports.

---

## Domain Layer Rules

The domain layer contains business behavior.

Domain code should be deterministic and independent from infrastructure.

Domain may contain:

```text
entities
aggregates
value objects
domain services
domain policies
domain events
domain errors
```

Domain must not contain:

```text
SQL queries
HTTP handlers
gRPC handlers
Temporal workflow code
Kafka/YDB topic consumers
OpenTelemetry code
logger setup
configuration loading
environment variable reads
```

Prefer rich domain behavior over anemic DTOs.

Good:

```go
func (p *Project) Suspend(reason SuspendReason, now time.Time) error {
    if p.State == ProjectStateDeleted {
        return ErrProjectAlreadyDeleted
    }

    p.State = ProjectStateSuspended
    p.StateReason = reason.String()
    p.UpdatedAt = now

    return nil
}
```

Avoid:

```go
project.State = "suspended"
project.StateReason = "manual"
```

---

## Application Layer Rules

The application layer coordinates use cases.

It may contain:

```text
commands
queries
use case handlers
transaction boundaries
authorization checks
idempotency checks
ports
DTOs for application-level responses
```

Application layer owns orchestration of business actions.

Transport adapters must call application use cases.

Do not put use case logic directly into HTTP/gRPC handlers.

---

## Adapter Layer Rules

Adapters connect the application to the outside world.

Adapters may contain:

```text
gRPC handlers
HTTP handlers
database repositories
event publishers
event consumers
Temporal activities
Temporal workflow adapters
Keycloak clients
SpiceDB clients
Kubernetes clients
cloud provider clients
```

Adapters must not own business rules.

Adapters translate:

```text
transport request -> application command/query
application result -> transport response
infrastructure error -> application/domain error where appropriate
```

---

## Platform Foundation Layer

Because M8 Platform will contain many services, shared technical infrastructure must live in `internal/platform`.

```text
internal/platform/
├── bootstrap/
├── config/
├── logging/
├── telemetry/
├── database/
├── transaction/
├── server/
├── grpc/
├── http/
├── temporal/
├── events/
├── health/
├── authn/
├── authz/
├── idgen/
├── clock/
└── errors/
```

Rules:

- `internal/platform` contains technical foundation only.
- It must not contain M8 business logic.
- It must not know about concrete business modules.
- Business modules may depend on `internal/platform`.
- `internal/platform` must not depend on business modules.

Good:

```text
internal/authentication/app -> internal/platform/transaction
internal/provisioning/app   -> internal/platform/idgen
internal/access/adapter     -> internal/platform/telemetry
```

Bad:

```text
internal/platform/database -> internal/authentication/domain
internal/platform/server   -> internal/resourcemanager/app
```

---

## Shared Kernel Rule

Do not create one large shared domain package.

Shared domain concepts are allowed only when they are stable and truly cross-cutting.

Allowed examples:

```text
internal/sharedkernel/
├── tenant/
│   ├── organization_id.go
│   ├── workspace_id.go
│   └── project_id.go
├── operation/
│   └── operation_id.go
└── resource/
    └── resource_name.go
```

Shared kernel may contain:

```text
IDs
simple value objects
immutable primitive types
cross-module naming constraints
stable enums
```

Shared kernel must not contain:

```text
repositories
use cases
business workflows
provider logic
database logic
Temporal logic
authorization logic
```

---

## Inter-module Communication

Modules must not freely import each other's application services.

Avoid:

```go
authentication/app imports resourcemanager/app
provisioning/app imports runtime/app
delivery/app imports provisioning/app
```

Prefer one of these patterns.

### 1. API call through explicit client port

```go
type ProjectLookup interface {
    GetProject(ctx context.Context, id ProjectID) (*ProjectRef, error)
}
```

The adapter may call Resource Manager API.

### 2. Events

One module publishes events, another module reacts.

Example:

```text
ResourceManager publishes ProjectCreated
Provisioning reacts and prepares default resources
Audit records the event
```

### 3. Projection / read model

A module may maintain a local projection of external data.

Example:

```text
Access keeps a projection of organization/workspace/project hierarchy for authorization checks.
```

### 4. Shared identifiers only

Modules can share stable IDs without sharing behavior.

Example:

```go
type ProjectID string
```

---

## Service Bootstrap Rules

Every service should have a small `cmd/<service>/main.go`.

`main.go` should only:

1. load config
2. initialize logger
3. initialize telemetry
4. initialize infrastructure
5. compose modules
6. start servers/workers
7. handle graceful shutdown

Do not put business logic in `main.go`.

Recommended structure:

```go
func main() {
    ctx := context.Background()

    cfg := config.MustLoad()
    logger := logging.New(cfg.Logging)

    shutdownTelemetry, err := telemetry.Init(ctx, cfg.Telemetry)
    if err != nil {
        logger.Fatal("telemetry init failed", "error", err)
    }
    defer shutdownTelemetry(ctx)

    app, cleanup, err := bootstrap.NewApp(ctx, cfg, logger)
    if err != nil {
        logger.Fatal("bootstrap failed", "error", err)
    }
    defer cleanup()

    if err := app.Run(ctx); err != nil {
        logger.Fatal("service failed", "error", err)
    }
}
```

Do not duplicate bootstrap logic across services.

Move reusable technical bootstrap into `internal/platform/bootstrap`.

---

## Service Ownership

Each module must have a `README.md`.

Template:

```md
# M8 Authentication

## Responsibility

Owns authentication sessions, challenges, authenticators and authentication lifecycle.

## Owns

- Authentication
- AuthenticationChallenge
- AuthenticationSession
- AuthenticationOperation

## Does Not Own

- Users
- Organizations
- Projects
- Permissions

## Main APIs

- StartAuthentication
- GetAuthentication
- CancelAuthentication
- SelectAuthenticationChallenge
- ResendAuthenticationChallenge

## Events Published

- AuthenticationStarted
- AuthenticationSucceeded
- AuthenticationFailed
- AuthenticationCancelled

## Events Consumed

- UserDisabled
- ProjectSuspended
```

Codex must update module README files when introducing important module-level changes.

---

## Go Package Naming Rules

Use small packages with clear names.

Good package names:

```text
domain
model
event
policy
command
query
usecase
ports
postgres
ydb
grpcadapter
temporaladapter
```

Avoid:

```text
common
utils
helpers
manager
serviceimpl
misc
stuff
core
base
```

A package should have one reason to change.

---

## Dependency Injection

Use explicit constructor-based dependency injection.

Do not use global mutable state.

Good:

```go
type Service struct {
    repo  Repository
    clock Clock
    tx    TransactionManager
}

func NewService(repo Repository, clock Clock, tx TransactionManager) *Service {
    return &Service{
        repo:  repo,
        clock: clock,
        tx:    tx,
    }
}
```

Avoid:

```go
var globalRepo Repository

func SetRepository(repo Repository) {
    globalRepo = repo
}
```

Composition root lives in:

```text
cmd/<service>/main.go
internal/<module>/module.go
internal/platform/bootstrap
```

---

## Context Usage

Use `context.Context` in:

```text
application layer
adapter layer
infrastructure layer
transport layer
```

Avoid `context.Context` in pure domain entities and value objects.

Good:

```go
func (h *CreateProjectHandler) Handle(ctx context.Context, cmd CreateProjectCommand) (*ProjectDTO, error)
```

Avoid:

```go
func (p *Project) Activate(ctx context.Context) error
```

Domain behavior should be deterministic and independent from IO.

---

## ID Rules

Use explicit typed identifiers inside domain code.

```go
type OrganizationID string
type WorkspaceID string
type ProjectID string
type UserID string
type ResourceID string
type OperationID string
```

Avoid passing raw strings across domain boundaries when the value has business meaning.

External API IDs should be stable, opaque and safe to expose.

Recommended public ID prefixes:

```text
org_
ws_
prj_
usr_
idn_
authn_
authc_
acc_
res_
op_
```

---

## State Model Rules

For stateful resources, separate:

```text
desired state
actual state
operation progress
conditions
state reason
timestamps
version
etag
correlation id
retry count
placement metadata
```

Do not store only one generic `status`.

Preferred model:

```go
type Resource struct {
    ID            ResourceID
    DesiredState  DesiredState
    ActualState   ActualState
    Conditions    []Condition
    StateReason   string
    Version       int64
    ETag          string
    CorrelationID string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

Use optimistic locking for concurrent updates.

---

## Operations and Async Workflows

Long-running actions must be represented as operations.

Examples:

```text
create project
provision Kafka topic
create database
suspend workspace
migrate tenant
rotate credentials
delete resource
```

Use operation records for user-visible progress.

Use Temporal for orchestration when the process has:

```text
retries
timers
external side effects
compensation
human-visible progress
multiple steps
long execution time
```

Do not put business rules only inside Temporal workflow code.

Temporal workflows should orchestrate application services, not replace the application layer.

---

## Idempotency

All external mutation APIs should support idempotency where applicable.

Use idempotency keys for create and mutate commands.

Example:

```go
type CreateProjectCommand struct {
    OrganizationID OrganizationID
    WorkspaceID    WorkspaceID
    Name           string
    IdempotencyKey string
}
```

Do not rely on client retries being safe by accident.

---

## Events and Outbox

For state changes that publish events, use transactional outbox.

Do not publish messages directly inside a database transaction unless the failure model is explicitly documented.

Events must include:

```text
event_id
event_type
aggregate_id
aggregate_type
occurred_at
correlation_id
causation_id
version
payload
```

Example event names:

```text
m8.resourcemanager.project.created.v1
m8.authentication.authentication.succeeded.v1
m8.provisioning.resource.provisioned.v1
m8.access.permission.changed.v1
```

Events should be module-owned and versioned.

---

## API Design

Prefer protobuf-first APIs.

Use:

```text
protobuf
buf
buf validate
google.api.field_behavior
google.longrunning.Operation for async operations where appropriate
```

API packages must be module-owned and versioned.

Good:

```proto
package m8.authentication.v1;
package m8.resourcemanager.v1;
package m8.access.v1;
package m8.provisioning.v1;
```

Avoid:

```proto
package m8.v1;
```

Do not expose internal database models directly through API.

Transport DTOs and domain models are separate.

---

## API Compatibility

Public API changes must be deliberate.

When changing protobuf contracts:

- update `.proto` files
- run generation
- update examples
- update docs
- do not manually edit generated code
- consider backward compatibility
- use new fields instead of changing existing field meaning unless compatibility is intentionally broken

Generated code must not be manually edited.

---

## Authentication and Authorization

Authentication answers:

```text
Who is the subject?
```

Authorization answers:

```text
What can this subject do on this resource?
```

Keep these concerns separate.

Module ownership:

```text
M8 Identity         owns users and identities
M8 Authentication  owns authentication sessions and challenges
M8 Access          owns permissions, policies and authorization checks
M8 ResourceManager owns organization/workspace/project/resource hierarchy
```

Access checks should be explicit in application use cases.

Do not hide authorization only in HTTP middleware.

---

## Resource Manager Principles

Resource Manager is the source of truth for tenant hierarchy and resource catalog.

It owns:

```text
organizations
workspaces
projects
resource metadata
desired state
actual state summary
placement metadata
lifecycle state
suspend/delete semantics
```

Provisioning systems may reconcile actual infrastructure, but Resource Manager owns platform-level intent.

---

## Provisioning Principles

M8 Provisioning is operator-like, but not only Kubernetes-native.

It should support:

```text
cloud resources
self-hosted resources
provider drivers
resource descriptors
reconciliation
operation tracking
drift detection
retries
rollback/compensation where possible
```

Provisioning API should not expose provider-specific details too early.

Use provider drivers behind interfaces.

---

## Database Ownership

Each module owns its own database schema or logical namespace.

Good:

```text
resourcemanager.projects
authentication.authentications
access.relationships
provisioning.resources
audit.events
```

Bad:

```text
public.users
public.projects
public.statuses
public.settings
```

Avoid cross-module joins in application logic.

If a module needs external data, prefer:

```text
API call
event projection
read model
cached reference
denormalized snapshot
```

---

## Repository Rules

Repositories should express business-oriented persistence operations.

Good:

```go
type ProjectRepository interface {
    GetByID(ctx context.Context, id ProjectID) (*Project, error)
    Save(ctx context.Context, project *Project) error
}
```

Avoid:

```go
type ProjectRepository interface {
    Select(ctx context.Context, query string, args ...any) (*Project, error)
}
```

Do not leak SQL or database-specific logic into domain.

Use migrations.

Use optimistic locking for aggregates that can be updated concurrently.

---

## Configuration

Configuration must be explicit and typed.

Avoid reading environment variables deep inside business logic.

Good:

```go
type Config struct {
    HTTP     HTTPConfig
    GRPC     GRPCConfig
    Database DatabaseConfig
    Temporal TemporalConfig
    Auth     AuthConfig
}
```

Load config once at startup.

Validate config before starting the service.

---

## Logging

Use structured logging.

Every log related to a request or operation should include where available:

```text
correlation_id
request_id
operation_id
organization_id
workspace_id
project_id
user_id
resource_id
```

Do not log:

```text
passwords
private keys
tokens
refresh tokens
authorization codes
OTP values
session cookies
client secrets
personal documents
```

---

## Observability

Every service should expose:

```text
metrics
structured logs
traces
health checks
readiness checks
```

Prefer OpenTelemetry-compatible instrumentation.

Important metrics:

```text
request count
request latency
error count
workflow duration
operation duration
queue lag
reconciliation lag
retry count
failed operations
external provider latency
```

---

## Security Principles

Security must be designed in, not added later.

Rules:

```text
least privilege by default
deny by default
explicit authorization checks
no secrets in logs
no secrets in git
no raw tokens in database unless explicitly required
encrypt sensitive values where appropriate
audit important actions
validate all external input
avoid unsafe reflection or dynamic code execution
```

Every externally visible mutation must consider:

```text
authentication
authorization
idempotency
audit
input validation
rate limiting where applicable
```

---

## Error Handling

Use explicit domain and application errors.

Do not compare error strings.

Good:

```go
var ErrProjectNotFound = errors.New("project not found")
```

or typed errors:

```go
type ValidationError struct {
    Field   string
    Message string
}
```

Wrap infrastructure errors at boundaries:

```go
return fmt.Errorf("load project %s: %w", id, err)
```

Transport adapters are responsible for mapping internal errors to gRPC / HTTP errors.

---

## Go Coding Style

Use idiomatic Go.

Rules:

```text
small interfaces, defined by consumers
explicit errors
no panic for normal control flow
no package-level mutable state
no unnecessary abstractions
no premature generics
keep functions short
prefer clarity over cleverness
use gofmt
```

Run when relevant:

```bash
go test ./...
go test -race ./...
go vet ./...
golangci-lint run
```

---

## Testing Strategy

Use several test levels:

```text
domain tests       - pure business rules
use case tests     - application behavior with fake ports
adapter tests      - database / external integration tests
contract tests     - API compatibility
workflow tests     - Temporal workflow behavior
end-to-end tests   - critical flows only
```

Domain tests should be fast and not require infrastructure.

Use table-driven tests for business rules.

Do not overmock domain models.

---

## Local Development

The project should support local development of many services.

Preferred local development files:

```text
docker-compose.yml
devcontainer/
Makefile
Taskfile.yml
.env.example
```

Useful commands:

```bash
make test
make lint
make proto
make run-resource-manager
make run-authentication
make run-provisioning
make run-local
```

GoLand should be able to run individual services from `cmd/<service-name>`.

Do not assume that all services must be started for every development task.

---

## GoLand Rules

The developer uses GoLand as the primary IDE.

Do not modify:

```text
.idea/**
```

unless explicitly asked.

It is acceptable to create or update:

```text
.run/
```

only when the task explicitly asks for GoLand run configurations.

Prefer commands that work both in terminal and GoLand:

```bash
go test ./...
go test -race ./...
go vet ./...
golangci-lint run
buf lint
buf breaking
```

Do not assume VS Code-specific configuration.

Do not add VS Code settings unless explicitly requested.

---

## Generated Code

Generated code must not be manually edited.

Generated folders may include:

```text
api/gen/
internal/gen/
```

If protobuf changes are needed:

1. edit `.proto` files
2. run generation
3. use generated code from the output directory

Do not patch generated files manually.

---

## Documentation Rules

For architectural changes, add or update documentation.

Use:

```text
docs/adr/     - architectural decisions
docs/hld/     - high-level design
docs/rfc/     - proposals
```

ADR format:

```md
# ADR-0001: Title

## Status

Proposed | Accepted | Deprecated | Superseded

## Context

What problem are we solving?

## Decision

What did we decide?

## Consequences

What becomes easier?
What becomes harder?
What risks remain?
```

---

## Codex Workflow For Implementing Features

When implementing a feature:

1. read the relevant module first
2. identify the bounded context
3. check existing domain models and use cases
4. add or update domain behavior first
5. add application command/query/use case
6. add or update ports
7. implement adapters
8. add tests
9. run formatting and tests
10. keep changes minimal and focused

When unsure between two designs, prefer:

```text
explicit over implicit
domain model over anemic DTO
small interface over large interface
composition over inheritance-like structures
clear package boundary over convenience import
operation tracking over hidden async side effect
idempotency over unsafe retry behavior
event/outbox over unsafe direct publish
projection over cross-module database join
```

---

## Codex Rules For Adding New Modules

When Codex adds a new module, it must:

1. create the module under `internal/<module>`
2. create clear domain/app/adapter boundaries
3. add a module `README.md`
4. add API contracts under `api/proto/m8/<module>/v1`
5. add a service under `cmd/` only if it is independently deployable
6. avoid importing another module's app/domain directly unless explicitly allowed
7. add tests for domain and use cases
8. update documentation if the module changes architecture
9. keep the first implementation small but structurally correct

---

## Codex Rules For Adding New Services

When Codex adds a new service, it must:

1. create `cmd/<service-name>/main.go`
2. use existing platform bootstrap/config/logging/telemetry packages
3. compose only required modules
4. expose only the service-owned API
5. support graceful shutdown
6. add a local run command if the repository uses Makefile/Taskfile
7. not duplicate bootstrap logic from other services
8. not create business logic inside `main.go`

---

## What Codex Must Avoid

Do not:

```text
create utils/common/helper packages without strong reason
put business logic in HTTP/gRPC handlers
put business logic only in Temporal workflows
let adapters depend on each other directly
expose database models through API
introduce global service locators
use panic for expected errors
swallow errors
log secrets
create large god services
create circular dependencies
bypass authorization checks
introduce framework-specific code into domain
modify generated files manually
rewrite large parts of the repository without explicit instruction
```

---

## Preferred Use Case Style

Use this style for application use cases:

```go
type CreateProjectCommand struct {
    OrganizationID OrganizationID
    WorkspaceID    WorkspaceID
    Name           string
    IdempotencyKey string
}

type CreateProjectHandler struct {
    projects ProjectRepository
    orgs     OrganizationRepository
    tx       TransactionManager
    clock    Clock
}

func NewCreateProjectHandler(
    projects ProjectRepository,
    orgs OrganizationRepository,
    tx TransactionManager,
    clock Clock,
) *CreateProjectHandler {
    return &CreateProjectHandler{
        projects: projects,
        orgs:     orgs,
        tx:       tx,
        clock:    clock,
    }
}

func (h *CreateProjectHandler) Handle(ctx context.Context, cmd CreateProjectCommand) (*Project, error) {
    if err := cmd.Validate(); err != nil {
        return nil, err
    }

    var result *Project

    err := h.tx.WithTx(ctx, func(ctx context.Context) error {
        org, err := h.orgs.GetByID(ctx, cmd.OrganizationID)
        if err != nil {
            return err
        }

        project, err := NewProject(
            cmd.OrganizationID,
            cmd.WorkspaceID,
            cmd.Name,
            h.clock.Now(),
        )
        if err != nil {
            return err
        }

        if err := org.CanCreateProject(project); err != nil {
            return err
        }

        if err := h.projects.Save(ctx, project); err != nil {
            return err
        }

        result = project
        return nil
    })
    if err != nil {
        return nil, err
    }

    return result, nil
}
```

---

## Definition of Done

A task is done only when:

```text
code compiles
tests are added or updated
relevant docs are updated
public API changes are reflected in protobuf/docs
errors are handled explicitly
observability is considered
security impact is considered
authorization is considered
idempotency is considered for mutations
no generated files are manually edited
gofmt is applied
go test ./... passes unless repository state prevents it
```

If tests or checks cannot be run, Codex must clearly state why.
