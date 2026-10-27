# Resource Manager — Organizations UI (v1)

Implemented in the canonical console under `/resource-manager/organizations/$organizationId`.

## Live screens

- **Overview**: Organization identity, tenant scope, current lifecycle, timestamps and revision-aware counts if data can be read.
- **Org units**: real `OrgStructureService.ListOrgUnits` paging, stable IDs, tree with search, expandable nodes, selection and inspector for management mode/source and structure revision.
- **Resources**: actual organization-scoped `ResourceManagerService.ListResources` response.
- **Ownership, Directory Sync, History**: visible product navigation with an explicit *API not connected* message, never synthetic access, employees or synced providers.

The Organizations list page remains the preexisting list implementation and its `/resource-manager/v1/organizations` route. The detail page uses the new resource and OrgUnit routes from their canonical proto HTTP annotations.

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
