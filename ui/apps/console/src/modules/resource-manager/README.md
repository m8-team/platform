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
