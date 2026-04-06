# Architecture

## Modules

- `identity`: tenants, users, memberships, groups, service accounts, OAuth clients
- `authz`: role catalog, binding APIs, `CheckAccess`, `ExplainAccess`
- `graph`: reverse access lookups and impact simulation
- `support`: temporary support grants with TTL and workflow orchestration
- `audit`: mutation audit log
- `ops`: long-running operation status

## Runtime Split

- `Keycloak` handles authentication, brokering, clients, and service accounts.
- `SpiceDB` handles runtime permission checks and relationship graph evaluation.
- `YDB` stores business metadata, audit history, operations, read models and outbox state.
- `YDB Topics` fan out domain events for projections and authorization sync.
- `Temporal` orchestrates long-running flows and timers.
- `Redis` caches `CheckAccess` and related hot-path computations.

## Current Implementation Shape

- gRPC contracts are generated from protobuf definitions.
- Domain services persist protobuf documents through a shared store interface.
- Adapter packages are wired through `internal/app`.
- Unit tests exercise fallback authorization, role expansion, cache keys and support grant lifecycle.

## Next Hardening Steps

1. Replace conservative YDB store stubs with table-session read/write code and outbox transactions.
2. Replace conservative SpiceDB adapter methods with Authzed/SpiceDB RPC calls.
3. Implement YDB Topics producer and consumer over the real SDK.
4. Add integration suites against live YDB, Redis, Keycloak, SpiceDB and Temporal instances.
