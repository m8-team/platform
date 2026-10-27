-- P1: business/legal corporate relationships independent from Group affiliations.
-- Deploy after 003_unified_resources.sql. All data is Tenant-scoped.
PRAGMA TablePathPrefix("/local/m8");

CREATE TABLE `m8-resource-manager/organization_relationships` (
  tenant_id Utf8 NOT NULL,
  id Utf8 NOT NULL,
  from_organization_id Utf8 NOT NULL,
  to_organization_id Utf8 NOT NULL,
  relationship_type Utf8 NOT NULL,
  ownership_basis_points Int32 NOT NULL,
  valid_from_unix Int64 NOT NULL,
  valid_to_unix Int64 NOT NULL,
  evidence_ref Utf8 NOT NULL,
  state Utf8 NOT NULL,
  revision Int64 NOT NULL,
  created_at Timestamp NOT NULL,
  updated_at Timestamp NOT NULL,
  PRIMARY KEY (tenant_id, id),
  INDEX idx_relationships_from GLOBAL ON (tenant_id, from_organization_id, id),
  INDEX idx_relationships_to GLOBAL ON (tenant_id, to_organization_id, id)
);

CREATE TABLE `m8-resource-manager/organization_relationship_requests` (
  tenant_id Utf8 NOT NULL,
  idempotency_key Utf8 NOT NULL,
  request_hash String NOT NULL,
  relationship_id Utf8 NOT NULL,
  PRIMARY KEY (tenant_id, idempotency_key)
);

CREATE TABLE `m8-resource-manager/organization_relationship_history` (
  tenant_id Utf8 NOT NULL,
  relationship_id Utf8 NOT NULL,
  revision Int64 NOT NULL,
  from_organization_id Utf8 NOT NULL,
  to_organization_id Utf8 NOT NULL,
  relationship_type Utf8 NOT NULL,
  ownership_basis_points Int32 NOT NULL,
  valid_from_unix Int64 NOT NULL,
  valid_to_unix Int64 NOT NULL,
  evidence_ref Utf8 NOT NULL,
  state Utf8 NOT NULL,
  updated_at Timestamp NOT NULL,
  PRIMARY KEY (tenant_id, relationship_id, revision)
);

-- Every relationship write touches the same Tenant row, so concurrent
-- insertions cannot independently commit a reciprocal ownership cycle.
CREATE TABLE `m8-resource-manager/organization_relationship_graph_heads` (
  tenant_id Utf8 NOT NULL,
  revision Int64 NOT NULL,
  updated_at Timestamp NOT NULL,
  PRIMARY KEY (tenant_id)
);
