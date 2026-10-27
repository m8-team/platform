-- P0.7/P0.8: canonical resource catalog and atomic organization provision.
-- Deploy after 001_init.sql and 002_org_structure.sql.
PRAGMA TablePathPrefix("/local/m8");

CREATE TABLE `m8-resource-manager/resource_records` (
  kind Utf8 NOT NULL,
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  owner_kind Utf8 NOT NULL,
  owner_id Utf8 NOT NULL,
  organization_id Utf8 NOT NULL,
  display_name Utf8 NOT NULL,
  description Utf8 NOT NULL,
  state Utf8 NOT NULL,
  revision Int64 NOT NULL,
  environment_key Utf8 NOT NULL,
  created_at Timestamp NOT NULL,
  updated_at Timestamp NOT NULL,
  PRIMARY KEY (kind, id),
  INDEX idx_resources_organization GLOBAL ON (organization_id, kind, id),
  INDEX idx_resources_owner GLOBAL ON (owner_kind, owner_id, state)
);

CREATE TABLE `m8-resource-manager/resource_create_requests` (
  scope Utf8 NOT NULL,
  idempotency_key Utf8 NOT NULL,
  request_hash String NOT NULL,
  resource_kind Utf8 NOT NULL,
  resource_id Utf8 NOT NULL,
  created_at Timestamp NOT NULL,
  PRIMARY KEY (scope, idempotency_key)
);

-- One environment key per organization. Tombstones remain for historical
-- reference, and releasing a name requires an explicit lifecycle workflow.
CREATE TABLE `m8-resource-manager/resource_environment_keys` (
  organization_id Utf8 NOT NULL,
  environment_key Utf8 NOT NULL,
  environment_id Utf8 NOT NULL,
  PRIMARY KEY (organization_id, environment_key)
);

CREATE TABLE `m8-resource-manager/resource_group_affiliations` (
  group_id Utf8 NOT NULL,
  organization_id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  created_at Timestamp NOT NULL,
  PRIMARY KEY (group_id, organization_id)
);

CREATE TABLE `m8-resource-manager/resource_environment_bindings` (
  id Utf8 NOT NULL,
  resource_kind Utf8 NOT NULL,
  resource_id Utf8 NOT NULL,
  organization_id Utf8 NOT NULL,
  environment_id Utf8 NOT NULL,
  version_ref Utf8 NOT NULL,
  configuration_ref Utf8 NOT NULL,
  revision Int64 NOT NULL,
  created_at Timestamp NOT NULL,
  updated_at Timestamp NOT NULL,
  PRIMARY KEY (id),
  INDEX idx_binding_resource GLOBAL ON (resource_kind, resource_id, environment_id)
);
