-- Org structure v1 migration. Requires 001_init.sql: outbox, inbox, idempotency.
PRAGMA TablePathPrefix("/local/m8");

-- Readers first load this pointer then only immutable rows for that snapshot.
CREATE TABLE `m8-resource-manager/org_structure_heads` (
    tenant_id Utf8 NOT NULL,
    organization_id Utf8 NOT NULL,
    active_revision Uint64 NOT NULL,
    active_snapshot_id Utf8 NOT NULL,
    updated_at Timestamp NOT NULL,
    PRIMARY KEY (tenant_id, organization_id)
);

-- Stable identity record; never cascaded on OrgUnit reparent or close.
CREATE TABLE `m8-resource-manager/org_units` (
    tenant_id Utf8 NOT NULL,
    organization_id Utf8 NOT NULL,
    org_unit_id Utf8 NOT NULL,
    authoritative_source_id Utf8,
    create_time Timestamp NOT NULL,
    PRIMARY KEY (tenant_id, organization_id, org_unit_id)
);

-- Every builder uses its own snapshot ID. Concurrent writers must NOT share
-- a revision+1 snapshot namespace before the head's optimistic CAS.
CREATE TABLE `m8-resource-manager/org_unit_snapshots` (
    tenant_id Utf8 NOT NULL,
    organization_id Utf8 NOT NULL,
    snapshot_id Utf8 NOT NULL,
    org_unit_id Utf8 NOT NULL,
    parent_org_unit_id Utf8,
    display_name Utf8 NOT NULL,
    lifecycle_state Utf8 NOT NULL,
    authoritative_source_id Utf8,
    PRIMARY KEY (tenant_id, organization_id, snapshot_id, org_unit_id)
);

CREATE TABLE `m8-resource-manager/org_unit_closure` (
    tenant_id Utf8 NOT NULL,
    organization_id Utf8 NOT NULL,
    snapshot_id Utf8 NOT NULL,
    ancestor_id Utf8 NOT NULL,
    descendant_id Utf8 NOT NULL,
    depth Uint32 NOT NULL,
    PRIMARY KEY (tenant_id, organization_id, snapshot_id, ancestor_id, descendant_id),
    INDEX idx_closure_descendant GLOBAL ON (
        tenant_id, organization_id, snapshot_id, descendant_id, depth
    )
);

-- Future bitemporal corrections query by both effective and recorded intervals.
-- These rows cannot be changed by the immediate-only Go writer in this slice.
CREATE TABLE `m8-resource-manager/org_unit_versions` (
    tenant_id Utf8 NOT NULL,
    organization_id Utf8 NOT NULL,
    org_unit_id Utf8 NOT NULL,
    recorded_from Timestamp NOT NULL,
    valid_from Timestamp NOT NULL,
    valid_to Timestamp,
    recorded_to Timestamp,
    parent_org_unit_id Utf8,
    display_name Utf8 NOT NULL,
    lifecycle_state Utf8 NOT NULL,
    source_revision Utf8,
    published_revision Uint64 NOT NULL,
    PRIMARY KEY (tenant_id, organization_id, org_unit_id, recorded_from, valid_from)
);

CREATE TABLE `m8-resource-manager/org_unit_change_sets` (
    tenant_id Utf8 NOT NULL,
    organization_id Utf8 NOT NULL,
    change_set_id Utf8 NOT NULL,
    source_id Utf8,
    base_revision Uint64 NOT NULL,
    target_revision Uint64,
    staged_snapshot_id Utf8,
    request_hash String NOT NULL,
    status Utf8 NOT NULL,
    effective_at Timestamp NOT NULL,
    recorded_at Timestamp NOT NULL,
    changes JsonDocument NOT NULL,
    impact_evidence JsonDocument,
    created_at Timestamp NOT NULL,
    updated_at Timestamp NOT NULL,
    PRIMARY KEY (tenant_id, organization_id, change_set_id),
    INDEX idx_changesets_state GLOBAL ON (tenant_id, organization_id, status, updated_at)
);
