# Access and Resource Manager — DDD / CQRS / Hexagonal migration

Status: **INCREMENTAL / NOT FULLY MIGRATED**. This document is the architecture contract for the scoped refactor; it does not claim that the existing v1 gRPC handler is secure for production.

## Ownership and boundaries

- **Resource Manager** is the only writer for resource identity, hierarchy, tenant, owner, lifecycle and authoritative resource revision. `domain/resource.Snapshot` carries these facts without importing any storage or API types.
- **Access** is the only writer for roles, bindings, grants, policy revision and authorization decisions. SpiceDB relationships are a read projection, never the authoritative grant store.
- Resource Manager application query `AuthoritativeResource.Get/GetActive` sits behind a narrow `ResourceSnapshotRepository` port. `adapter/resourcecatalog` maps the current YDB/v1engine catalog to the domain read model.
- Access depends **only on its own** `consistency.ResourceRevisionReader` and `projection.SnapshotReader` ports. No `internal/access/**` package imports Resource Manager.
- A separate integration adapter, `internal/integration/accessrm`, maps Resource Manager's authoritative CQRS query to those ports; the integration composition root is the only package that imports both bounded contexts.

## Consistency and single writers

```text
Resource Manager commands -> transactional state + outbox (owner)
                                        |
                                        v
                      Resource projection worker (Access)
                              | durable fenced checkpoint
                              v
                             SpiceDB
                              |
RM GetActive(revision) ----> Consistency Barrier <---- Access policy journal head/applied
                              |
                      AuthZEN / permission checks
```

- For an **allow** decision, RM authoritative revision must be covered by an acknowledged Resource Manager watermark; Access policy journal head must equal applied; SpiceDB evaluation must be fully consistent. Recheck both owners after positive decisions.
- A checkpoint's opaque ZedToken cannot be compared with another independent token. The barrier therefore uses fully consistent evaluation when both owner streams matter.
- The graph projection uses a per-aggregate fenced lease and YDB durable checkpoint before event acknowledgement. Reconciliation may run concurrently **only across distinct aggregate keys**. `ReconcileKeys` deduplicates keys, uses bounded worker goroutines and propagates context cancellation. Event feed acknowledgement stays serial and ordered.
- RM's revision must advance on changes to security-relevant owner/tenant/lifecycle facts, including move/transfer. A production write must conditionally verify tenant, owner, expected revision, authorization preconditions and idempotency in its owning transaction.

## CQRS layers

| Module | Command | Query | Projection / adapters |
| --- | --- | --- | --- |
| Resource Manager | `app/orgstructure.Service.Apply` and v1 resource mutations (ongoing migration) | `app/query.AuthoritativeResource`, organization query use cases | YDB catalog to domain snapshot adapter; outbox source |
| Access | role/grant/binding mutation ports and policy journal | `app/usecase.CheckPermissionHandler` via AuthZEN and barrier | fenced `app/projection`; SpiceDB Go SDK |

## Remaining work before claiming full migration

1. Move v1 Resource Manager mutation logic from `adapter/grpc` and `v1engine` into domain aggregates + typed command handlers; enforce expected revision, idempotency, outbox and authorization in the **durable** YDB transaction. The existing transport still contains business logic.
2. Migrate legacy OrganizationService write/read concerns into separate command/query handlers; remove the development-only `AllowAll` configuration from any production deployment.
3. Connect real owner outbox/CDC feeds and reconciliation discovery to both workers, run a two-replica lease-loss/fencing integration test and instrument lag/watermark metrics.
4. Implement permission-aware scoped resource list queries before pagination and replace remaining v1 transport placeholders.
5. Add end-to-end tests for active tenant transfer, concurrent owner changes and grant revocation races, YDB transaction rollback and SpiceDB consistency failures.

This stage changes **no external Protobuf contracts**, no database schema and no deployment. It must not be merged as a claim of complete Resource Manager v1 readiness.

## Dependency rule enforced by tests

```text
internal/access/{domain,app,adapters,wiring}
        |            ports only
        v
  Access-owned interfaces
        ^
        | implements
internal/integration/accessrm
        |
        v
internal/resourcemanager/{app/query,adapter/resourcecatalog}
```

The cross-module adapter is intentionally **not inside** `internal/access`.
The integration package composes an authoritative Resource Manager read implementation
and passes it through Access-owned ports. Production graph projections may still
consume RM-authored updates; Access is decoupled from RM's Go packages and YDB
schema, but the freshness barrier still requires a trusted resource revision
provider. If runtime independence is required, design a separate owner-certified
revision feed/receipt contract and document its revocation guarantees before
removing the authoritative freshness check.

`internal/access/dependency_test.go` rejects any direct imports of
`internal/resourcemanager` or `internal/integration` across Access packages,
including their Go tests.
