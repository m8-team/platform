# Canonical Resource Manager v1 — runtime status

**Status:** Executable **development-only** gRPC and grpc-gateway API, **NOT production ready**.

## Build & verification

The repository workflow `.github/workflows/resource-manager-v1.yml` runs:
```sh
buf lint
buf generate
go test ./internal/resourcemanager/v1engine
go test ./...
```
Generated protobuf/gRPC/Connect/gateway/OpenAPI sources are committed to `api/generate/` by the workflow on the integration branch. The root Go module resolves `github.com/m8-team/go-genproto` to `./api/generate/go` via `go.mod`.

## Run locally only

```sh
export M8_RM_ALLOW_UNAUTHENTICATED=true
export M8_HTTP_ADDR=127.0.0.1:8080
export M8_GRPC_ADDR=127.0.0.1:9090
export M8_HEALTH_HTTP_ADDR=127.0.0.1:8081
go run ./cmd/resource-manager
```

The default mode **denies all RPC access** until a production authorization port is configured. The temporary development bypass is rejected at startup unless both API listeners are bound to loopback addresses.

Example through canonical grpc-gateway:

```sh
curl -fsS -X POST http://127.0.0.1:8080/resource-manager/v1/resources \
  -H 'Content-Type: application/json' \
  --data '{"resource":{"tenant":{"metadata":{"name":"Demo account"}}},"idempotencyKey":"demo-tenant-1"}'
```

The response is a typed protobuf JSON ResourceRecord with a server-generated UUID. The returned `tenant.metadata.id` can then be used:

```sh
curl -fsS -X POST http://127.0.0.1:8080/resource-manager/v1/resources \
  -H 'Content-Type: application/json' \
  --data '{"resource":{"organization":{"metadata":{"name":"Retail LLC"},"tenantId":"<TENANT_UUID>"}},"idempotencyKey":"demo-org-1"}'
```

To run a separate, loopback-only standard-library prototype (independent of grpc-gateway), set `M8_RM_DEV_TOKEN` to a secret of at least 32 characters and `go run ./cmd/resource-manager-v1-local` (port 18080, bearer token required).

## Implemented on development memory adapter

- Create/Get/Update/ChangeState for Tenant, Group, Organization, Workspace, Project, Environment and WorkloadService.
- Scoped listing of Organization-owned objects (simple query mode).
- Explicit Project move within one Organization and canonical ancestor resolution.
- Group-to-Organization affiliation, association removal and listing.
- EnvironmentBinding creation/update, lookup by resource/environment, revision-aware removal.
- Blocking deletion of resources with active children.
- Server-side computation of each object's owning Organization and rejection of cross-Organization environment bindings.
- Unauthenticated mode disabled by default; development bypass restricted to local interfaces.

## Explicitly unsupported or incomplete

- The `ResourceDescriptor` and `RegisterResourceKind` domain-specific registration flow is not yet implemented. Unsupported operations return `Unimplemented`, not success.
- Full paginated/filtered listing, field-mask semantics beyond basic fields, durable idempotency and event publication require the production storage and query model.
- Persisting requested Environment.kind/isolation independently from Environment.key requires a complete Environment aggregate. Do not treat the dev default isolation response as actual Runtime enforcement.
- `SetGroupOrganizationLink` and environment binding mutations are in-memory only.
- Authentication, authorization, risk, approvals, audit, YDB transactional outbox/inbox and deployment reconciliation are **not connected**.
- Identity, Workforce, AccessGrant, ModuleRegistry, SharedServices and Delivery have protobuf source contracts and generated bindings, but are **not yet implemented as running services**.
- Automatic code-generation workflow runs in CI, but production release requires immutable SDK artifacts and end-to-end deployment tests.

## Mandatory production acceptance gates

1. YDB repositories with parent indexes, optimistic concurrency, transactional outbox/idempotency and event deduplication.
2. Access Service / AuthZEN integration with relationship checks on **every** resource action and prefiltered resource lists.
3. Correct user and service principal propagation, cross-tenant denial, short-lived scoped support grants, environment access policies and complete audit trail.
4. Complete type-safe ResourceDescriptor registration, kind policy and environment data models.
5. Backward-compatibility is not required for the replaced old model, but the newly established **canonical v1** contract must be stable after release.
6. All new domain API servers and affected clients compiled, deployed, and subjected to integration/security tests.
