# Org Structure v1 — first implementation slice

**Package:** `m8.platform.resourcemanager.v1` (no underscore). Breaking v1 redesign.

## Ownership and invariants

- Resource Manager owns Organization, OrgUnit, adjacency, closure materialization, bitemporal structure versions and succession.
- Identity/Directory owns Person, OrganizationMembership, OrgUnitAssignment, ManagerAssignment and TeamMembership. Do not embed employee edges in closure.
- A stable OrgUnit ID survives renaming and moving. OrgUnit belongs to exactly one Organization; changing legal ownership requires an explicit operation.
- Closure paths are derived from adjacency and organization-scoped. Each node has a depth-0 self-row. No cycles, duplicates or orphan parents.
- Parent changes require optimistic structure revision checks. A no-op does not increment revision. Closed parents cannot receive new units.
- Current and historical structure are separate read models. History uses half-open effective and recorded intervals.
- Resource ownership links reference stable OrgUnit IDs. Closure triggers reassignment review, never cascading deletion.
- M8 Access re-evaluates inherited policies against a sufficiently fresh structure revision.

## Slice boundary

This PR adds protobuf read models and a **pure Go** hierarchy validator/closure calculator, deterministic no-mutation reparenting, and unit tests. It does NOT register a gRPC service, persist YDB tables, ingest 1C:ZUP, or implement approvals. The old `ResourceManagerService` is retained until a coordinated replacement. Generated code and buf lint must be verified by repository CI.

## Next

1. YDB migrations: org_units, org_unit_versions, org_structure_revisions, org_unit_closure, change_sets, outbox, atomic structure revision publishing.
2. OrgUnitChangeSet validate/apply with authz, source-authority, idempotency and impact analysis.
3. gRPC/HTTP adapters with tenant authorization before pagination and revision-bound page tokens.
4. 1C:ZUP Directory Sync: mappings, snapshots, full-snapshot completeness, diff/quarantine and conflict resolution.
5. Identity/Directory temporal membership projections, Access authorization fences and ownership transfer.
6. End-to-end tests for concurrent moves, backdated changes, source outages and dependent permissions.
