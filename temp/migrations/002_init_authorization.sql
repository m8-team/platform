-- Authorization metadata and read models.

CREATE TABLE IF NOT EXISTS roles (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS resources (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS role_templates (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS binding_operations (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS subject_access_index (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS resource_subject_index (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS access_explain_edges (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS change_impact_index (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS audit_events (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS operations (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);
