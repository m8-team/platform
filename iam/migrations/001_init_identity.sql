-- Identity and OAuth facade tables.

CREATE TABLE users (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE tenants (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE memberships (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE groups (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE group_members (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE service_accounts (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE federated_links (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE oauth_clients (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE client_secret_refs (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);
