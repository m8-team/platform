# Resource Model v1 — contract rollout and implementation checklist

## Release rule

The old version is deliberately discarded. `v1` now means **Tenant + Organization + optional Group/Workspace/Project + Environment + logical Service**. There is no compatibility obligation to the old `OrganizationService/WorkspaceService/ServiceService` APIs or to `Service.environment`.

## Changes already committed in this branch

- Canonical RM proto messages: ResourceRef/OwnerScope, Tenant, Group, Organization, Workspace, Project, Environment, EnvironmentBinding, WorkloadService, ResourceDescriptor, ResourceKindPolicy.
- New unified ResourceManagerService protobuf; former RPC service proto sources removed.
- Separate module contracts for Identity, Workforce, AccessGrant, SharedServices and Delivery.
- Domain shape-check package with 23 placement scenarios.
- Canonical normative architecture spec and Resource Manager README.

## P0 — block merge until complete

**V1-01 — Contract validation and generated code**
- Run `buf lint` and `buf generate` with the repository's `buf.yaml` / `buf.gen.yaml`.
- Rebuild the separate `github.com/m8-team/go-genproto` module and update dependency references.
- Delete every legacy source-generated Go/Gateway/Connect/OpenAPI artifact for removed RPCs; never mix generations.
- Update contract catalogs, API docs, Postman collections, proto-package mappings and baseline snapshots.
- Treat previous baseline as broken by intent; establish new v1 baseline, then re-enable strict future breaking checks.

**V1-02 — Persistence**
- YDB migration for Tenants, Organizations, Groups, affiliations, Workspaces, Projects, Environments, WorkloadServices, ResourceDescriptors, EnvironmentBindings.
- Validate foreign references and tenant/organization ownership transactionally.
- Introduce parent index, search index, optimistic concurrency, soft-delete/tombstones, outbox + idempotency.
- Import migration or clean reset of old test data; no silent ownership inference from old workspace/service records.

**V1-03 — Resource Manager transport and application**
- Replace old OrganizationService server bindings with ResourceManagerService.
- Implement full CRUD/lifecycle and MoveProject; require authorization on both source and destination.
- Resolve server-side canonical owner organization and ancestry; prevent cycles and unauthorized traversal.
- Server-enforced resource kind placement policy and environment binding checks.
- Implement filtered, stable pagination with scoped list query.

**V1-04 — Identity / Workforce / Access**
- IdentityDirectoryService with verified external identities; prohibit email-based automatic account merges.
- WorkforceService records engagement separately from permissions.
- AccessGrantService: request/approve/revoke/expiry and per-operation real-time checks, including active support case.
- OPA/SpiceDB-backed Authorization; no implied grants from Group, Membership or ServiceAgreement.

**V1-05 — Shared Services and Delivery**
- SharedServicesService owns provider/consumer agreements, offerings and cases.
- Delivery owns deployment reconciliation and runtime state; RM EnvironmentBinding is declaration only.
- Tool Gateway must enforce company + environment + principal + agent + resource grants.

**V1-06 — End-to-end test matrix**
- Single independent Tenant/Organization with Agent directly under Organization (without Group/Workspace/Project).
- Holding with Group affiliation but no implicit cross-Organization permissions.
- OwnerScope Project under Organization and under Workspace; WorkloadService in both allowed scopes.
- Testing vs Production connection / credential isolation; cross-environment access denied.
- Temporary contractor without Organization Membership accessing one Project only.
- External Support provider assigned one Case, can read only that Case's permitted resource, expires/revokes immediately.
- Multi-company human one principal with independent Employment/Engagement and Organization memberships.
- Repeated concurrent mutations, pagination, deleted ancestor, stale revision, moved project, duplicate events.
- Security tests for enumeration across tenant/environment borders.

## Acceptance gates

- Protobuf generated output and descriptor compilation pass.
- Go build/test passes for all affected modules and consumers, not just pure scope package.
- Persistence, outbox, idempotency and authorization integration tests pass.
- Design documents, API catalogs and generated OpenAPI match the new v1.
- No references to removed service RPC names remain in runtime bindings (except explicit historical documentation).
- All breaking SDK/client and event changes are deployed coherently.

## Known limitations of this contract PR

The branch provides protobuf source contracts and a domain placement skeleton; **none of the new RPCs are yet registered as a running gRPC server**. Full regeneration and full Go repository tests require a connected build environment. This branch should remain draft until the gates above pass.

## Additional canonical contracts

- `api/proto/m8/platform/modules/v1/module.proto` / `module_registry_service.proto`: tenant entitlements and scope-aware module installation.
- `api/proto/m8/platform/resourcemanager/v1/resource_events.proto`: Resource, Group affiliation and EnvironmentBinding events.
- `api/proto/m8/platform/access/v1/access_grant_events.proto`: grant lifecycle events.

Before merging, align Module Registry implementation and Kafka event catalogs with these new sources.
