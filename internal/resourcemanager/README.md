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

## Implementation tracking

See `docs/architecture/resource-manager-spec.md` for normative design and `docs/engineering-artifacts/resource-manager-v1-contract-rollout.md` for the required delivery checklist.

Legacy codepath: `cmd/resource-manager` and `internal/resourcemanager/adapter/grpc/organization_server.go` implement old OrganizationService. Local memory repository + placeholder child projector are not production-ready. Remove/replace them while implementing new handlers; no compatibility adapter is necessary.
