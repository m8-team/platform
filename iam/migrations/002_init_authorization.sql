-- Authorization metadata and read models.

CREATE TABLE roles (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE resources (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE role_templates (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE binding_operations (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE subject_access_index (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE resource_subject_index (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE access_explain_edges (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE change_impact_index (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE audit_events (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE operations (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);
