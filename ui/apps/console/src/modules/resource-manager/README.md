# Resource Manager — Organizations UI (v1)

Implemented in the canonical console under `/resource-manager/organizations/$organizationId`.

## Live screens

- **Overview**: Organization identity, tenant scope, current lifecycle, timestamps and revision-aware counts if data can be read.
- **Org units**: real `OrgStructureService.ListOrgUnits` paging, stable IDs, tree with search, expandable nodes, selection and inspector for management mode/source and structure revision.
- **Resources**: actual organization-scoped `ResourceManagerService.ListResources` response.
- **Ownership, Directory Sync, History**: visible product navigation with an explicit *API not connected* message, never synthetic access, employees or synced providers.

The Organizations list now uses canonical `GET /resource-manager/v1/resources?kind=organization&tenantId=<UUID>`, with an explicitly selected Tenant (or `VITE_RESOURCE_MANAGER_TENANT_ID` in development). The retired `/resource-manager/v1/organizations` route is not registered by ResourceManagerService v1. Details use the resource and OrgUnit routes from canonical proto HTTP annotations.

## Data integrity and security

- Queries use the console's `loggedFetch`, same-origin credentials and TanStack Query cancellation.
- Invalid data, revision changes between OrgUnit pages, repeated tokens, duplicate units and limit overflow produce errors rather than a misleading partial hierarchy.
- The UI doesn't invent source health, ownership, memberships or access grants. Unreadable OrgUnits and parents remain governed by M8 Access server-side policies; the front end does not infer permissions from the tree.
- Mutations are not exposed before the real M8 Access/Impact Approval integration. An HTTP 403 from the fail-closed gRPC backend is surfaced as an error.
- The current backend limits advanced ResourceManagerService list pagination, and historical versions are UNIMPLEMENTED. Those are intentionally separate follow-ups.

## Validation

```sh
cd ui/apps/console
pnpm install --frozen-lockfile
pnpm lint
pnpm build
pnpm dev
```

Local dev proxies `/resource-manager/v1` to `http://127.0.0.1:8080`. A running REST gateway is required; gRPC-only registration is not sufficient to serve browser fetches.

## Follow-up

Connect verified browser authentication and REST gateway; implement Directory Sync, ownership and version-history backend, scoped/tree pagination for organizations with over 20k visible OrgUnits, and create/change-set workflows when backend policies permit changes.

## Phase 2: Organization management and reviewable ChangeSets

A second, stacked PR adds:
- `Create organization` action on the Organizations list with explicit Tenant UUID, name, description, field validation and a stable idempotency key per submission payload. A successful response navigates directly to the created Organization.
- `Settings` tab: editable name/description through `UpdateResource` (FieldMask, expected revision and idempotency key), immutable Tenant/ID, explicit suspend confirmation and resume via `ChangeResourceState`.
- `ChangeSets` tab: stage CREATE, RENAME, MOVE, CLOSE for manual OrgUnits, preview before/after name, parent, status, validate with `OrgUnitChangeSetService.ValidateOrgUnitChangeSet`, then explicitly confirm `ApplyOrgUnitChangeSet`. The apply action stays disabled until server validation succeeds for the exact current draft and base revision. The backend is authoritative for all constraints and impact-review/Access checks. No backdated or scheduled mutations are offered.
- `Ownership` tab: resource cards show the owning Organization and immediate structural parent from existing API. Business owner, technical owner, operating team and accountable OrgUnit display **API unavailable**, not **unassigned**; assignments cannot be changed before a dedicated Directory/Ownership API.

### Important deployment and API limitations

- All mutations go to real REST annotation paths, never local browser mocks. A configured REST gateway must expose `CreateResource`, `UpdateResource`, `ChangeResourceState`, `ValidateOrgUnitChangeSet`, and `ApplyOrgUnitChangeSet`.
- Production authentication and M8 Access (P0.9) and impact review (P0.10) are still required for writes. HTTP 401/403/409 errors are displayed. UI controls **do not** bypass authorization.
- Organization creation requires an actual Tenant ID; a Tenant picker and directory of authorized tenants do not exist yet.
- Company legal profiles, ownership assignments and 1C:ЗУП sync must be exposed by separate APIs; form fields are not invented to simulate persistence.
- The diff describes the hierarchy only; it does not claim to compute changes to external access, projects or employee assignments. Server-side approval/impact analysis is authoritative.
- Merge the base Organization details UI PR before this stacked PR, then update the base to `main`.

### Acceptance checks

```sh
cd ui/apps/console
pnpm install --frozen-lockfile
pnpm lint
pnpm build
```

## Organization creation: HTTP 404 / gRPC code 5

The `POST /resource-manager/v1/resources` route is registered and reached. If
the parent `tenantId` does not refer to a Tenant resource in this *same*
Resource Manager backend, `CreateResource(Organization)` returns gRPC
`NOT_FOUND` (code 5), translated by the REST gateway to HTTP 404
`resource not found`. A syntactically valid UUID does **not** imply that a
Tenant exists. An ID displayed by a different legacy `/organizations`
service may not be backed by a Tenant record in the canonical v1 store.

The Create Organization form supports two explicit flows:

1. **Existing Tenant:** enter the actual Tenant ID. The form first calls
   `GET /resource-manager/v1/resources/tenant/{id}` and requires a matching,
   ACTIVE Tenant record; 404 becomes a clear validation error.
2. **New Tenant (platform administration only):** first create a Tenant with
   `POST /resource-manager/v1/resources` and a `tenant` resource payload,
   then create the Organization using the server-issued Tenant ID. These are
   two separate requests with independent, stable idempotency keys. The UI
   displays the created Tenant ID even if Organization creation fails.

This does **not** weaken server-side authorization. Production M8 Access may
forbid Tenant creation or even reading tenants; 401/403 are not treated as
evidence of non-existence. Provisioning a Tenant must be authorized and should
not happen silently if a parent lookup fails.

Development smoke test (against a configured, authorized Resource Manager
gateway; never use a production bypass):

```http
POST /resource-manager/v1/resources
Content-Type: application/json

{"resource":{"tenant":{"metadata":{"name":"Test Holding"}}},"idempotencyKey":"tenant-test-001"}
```

Copy `tenant.metadata.id` from the response and use it for the Organization:

```http
POST /resource-manager/v1/resources
Content-Type: application/json

{"resource":{"organization":{"tenantId":"<id returned above>","metadata":{"name":"Test Retail"}}},"idempotencyKey":"organization-test-001"}
```

REST gateway, CORS and active environment must point to the same RM service
for both operations. Verify with
`GET /resource-manager/v1/resources/tenant/<id>`.

### Diagnosing an older Tenant ID

If `M8_RM_YDB_DSN` is unset, the current Resource Manager falls back to a
development-only in-memory backend. Restarting that backend discards its
Tenant records, so previously issued IDs are no longer valid. Production
requires migrated YDB storage and M8 Access. Never respond to a missing
Tenant by implicitly creating a different one under the caller's authority.

## Tenant-scoped Organization inventory (route 404 repair)

The old `GET /resource-manager/v1/organizations` returns **HTTP 404** because it is not part of the canonical v1 ResourceManagerService. The new list uses `GET /resource-manager/v1/resources?kind=organization&tenantId=<id>`. The Protobuf `ListResourcesRequest.tenant_id` field is additive and only valid with `kind=organization`, empty `organization_id`, and no additional owner scope. An unknown Tenant returns 404; no Tenant scope never results in a global list.

The UI requires an explicit Tenant UUID and stores the last selected Tenant only in `sessionStorage`; `VITE_RESOURCE_MANAGER_TENANT_ID` can supply a development default. Sorting, filtering and paging run client-side over the returned Tenant list in this phase. This is not a security boundary and is not suitable for very large Tenant inventories until indexed, signed server cursors are implemented. M8 Access authorization must be enforced on the server before enabling production read access; the current service guard remains fail-closed.
