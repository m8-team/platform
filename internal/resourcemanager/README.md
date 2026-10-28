# M8 Resource Manager — canonical v1 contract transition

## Status

The new protobuf contracts for the **replacement v1** are in `api/proto/m8/platform/resourcemanager/v1`. Breaking change explicitly approved. The old protobuf RPC service files were deleted.

**This is contract-first work, not an operational API migration.** The existing OrganizationService code still implements the former contract and must be replaced. Generated clients/OpenAPI are stale until regeneration.

## What Resource Manager owns

- Tenant (account metadata), independent Organization, optional Group affiliation
- Optional Workspace and Project management scopes
- Organization Environment registry (prod/staging/dev/test/preview)
- Logical WorkloadService catalog records
- Typed ResourceDescriptor ownership metadata for module-owned resources
- EnvironmentBinding records (logical resource version -> environment)

## What it does not own

- Principal, external identities, memberships: Identity
- Employment, contracts, freelancers: Workforce / HR
- Role bindings, grants, authorization decisions: Access
- HR/Support services, case management, service agreements: Shared Services
- Runtime rollout state, actual deployments: Delivery and Runtime
- Secrets, module business data, providers, audit storage

## Required invariants

Exactly one owner scope for each tenant-owned resource and one server-derived owning Organization. No mandatory Workspace/Project. Group links are business affiliations and confer no permissions. Environment is not a resource-container parent. All mutation operations need idempotency, expected revisions, authorization, audit and transactionally persisted domain events. Listing applies per-resource authorization before paging.

The new pure Go scope matrix in `internal/resourcemanager/domain/scope` validates *only* parent types and presence of an ID. It is **not** a full Resource Manager implementation.

## Resource contract

`resource.proto` validates supplied OwnerScopes, required descriptor fields,
UUID references and policy owner kinds. An absent optional scope filter is
allowed; a supplied empty scope is invalid. Empty policy owner sets permit no
placement, and environment bindings are disabled by default. Metadata keeps
optional display fields and zero server-managed values for creation; partial
updates must validate merged state after ignoring output-only input fields.

See the [Resource contract and registration example](../../docs/architecture/resource-manager-spec.md#resource-contract-validation).
Run `go test ./internal/resourcemanager/adapter/grpc -run 'TestResource.*Contract'`
for schema validation tests. Descriptor/kind-policy registration, runtime
validation, complete metadata persistence and correct field-mask merging remain
implementation gaps. OpenAPI also needs qualified schema names to avoid the
existing ResourceRef collision with the common operation package.

## Environment contract

`environment.proto` defines validation for complete Environment and
EnvironmentBinding values: UUID references, required messages, lowercase slugs,
explicit enum values and nonempty version references. Creation may omit
server-generated IDs. Field-mask updates must validate the merged resource;
immutability, ownership, authorization and concurrency checks remain application
responsibilities. Isolation is requested intent; expiration is an optional
absolute timestamp, including for already expired records.

See the [Environment contract and create example](../../docs/architecture/resource-manager-spec.md#environment-contract-validation).
Schema validation is covered by `go test ./internal/resourcemanager/adapter/grpc -run 'TestEnvironment.*Contract'`.
The current ResourceServer does not invoke Protovalidate or preserve
`kind`/`isolation`/`expires_at`, and SetEnvironmentBinding does not enforce
`expected_revision`/`idempotency_key`. These runtime gaps are not resolved by
protobuf annotations.

## Implementation tracking

See `docs/architecture/resource-manager-spec.md` for normative design and `docs/engineering-artifacts/resource-manager-v1-contract-rollout.md` for the required delivery checklist.

Legacy codepath: `cmd/resource-manager` and `internal/resourcemanager/adapter/grpc/organization_server.go` implement old OrganizationService. Local memory repository + placeholder child projector are not production-ready. Remove/replace them while implementing new handlers; no compatibility adapter is necessary.
