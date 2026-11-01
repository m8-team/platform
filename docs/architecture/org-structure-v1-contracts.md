# OrgUnit v1 — kind registry, external references and temporal history

**Package:** `m8.platform.resourcemanager.v1` in `api/proto/m8/platform/resourcemanager/v1`.

This change **fixes Protobuf contracts only**, not persistence, generated bindings or live
handlers. In particular, current `OrgStructureServer` supports immediate ChangeSets and
`ListOrgUnitVersions` returns UNIMPLEMENTED. New history/catalog RPCs must not be
advertised as available until their implementation is registered.

## Canonical invariants

1. `OrgUnit.metadata.id` is a permanent identity across moves, renames, source
   migration, code changes and kind changes. An OrgUnit belongs to exactly one
   `tenant_id` and `organization_id`. No cross-organization parent links.
2. `OrgUnit.kind_id` references `OrgUnitKind.metadata.id`, scoped SYSTEM, TENANT
   or ORGANIZATION. The free-text `OrgUnit.kind` field #6 is deprecated for a
   staged migration; never reuse tag #6. A disabled kind remains resolvable for
   existing units. The kind `code` is unique within its scope, stable on rename.
3. External IDs reside in `OrgUnitExternalReference` (separate collection), never
   serve as OrgUnit IDs and never get written into `authoritative_source_id`.
   A reference is unique by `(tenant_id, source_id, external_type, external_id)`
   for intersecting effective intervals. Its `[valid_from, valid_to)` is
   half-open. Retiring an integration does not delete old references.
4. `OrgUnit.structure_revision` is the published **Organization-wide** snapshot
   revision, not `ResourceMetadata.revision` or a local OrgUnit version.
   All reads of one snapshot share a revision; a no-op does not increment it.
   Publish CAS/snapshot/outbox in one transaction, preserving revision-bound
   pagination and authorization checks.
5. `OrgUnitVersion` has independent effective (`valid_from/valid_to`) and
   knowledge (`recorded_from/recorded_to`) half-open intervals. The new
   `structure_revision_from/to` record the Organization revision span for the
   registered view. Missing upper bounds mean infinity. Within a recorded
   revision, versions of the same unit must not overlap in effective time.
   Historical reads select both `as_of` and `at_revision` (0 = latest).
   A backdated correction must create new knowledge, not rewrite prior history.
6. `OrgUnitMove` is immutable move provenance (parents, actor, reason,
   effective time, recorded time, change set, Organization revision).
   `OrgUnitVersion` is authoritative for reconstruction; the move log is audit.
   No parent may be its own descendant at any scheduled effective boundary.

## Persistence / access follow-up

- Extend existing YDB `org_unit_snapshots`, `org_unit_versions` and
  `org_structure_heads` without discarding their established snapshot IDs;
  add kinds, external references/active-key claims and move audit tables.
  Existing `002_org_structure.sql` is **not** sufficient to persist new fields.
- Preserve source-authority, Access decision gates and approved ChangeSets.
  Current `ApplyImmediate` rejects scheduled/backdated times; temporal writers
  must be separate and atomic before `GetOrganizationStructureAt` works.
- Generate committed Go, ConnectRPC, gateway and OpenAPI artifacts with
  `buf generate` and check `buf lint` / breaking policy before merge.
- Implement tenant-scoped authz for kind/ref/history APIs and require revision
  fences for rights derived from the org tree.
- Backfill existing free-text `kind` values into Kind records before requiring
  `kind_id`, and migrate active foreign IDs without changing OrgUnit UUIDs.

### Example

```text
Organization rev 41: IT -> Development
Organization rev 42: Engineering -> Development
OrgUnit.id remains identical
1C/ZUP: (source=hr-zup, type=department, external_id=81) -> OrgUnit.id
SAP:    (source=sap,    type=department, external_id=SF-42) -> same OrgUnit.id
```
