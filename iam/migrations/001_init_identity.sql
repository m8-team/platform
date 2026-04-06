-- Identity and OAuth facade tables.

CREATE TABLE IF NOT EXISTS users (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS tenants (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS memberships (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS groups (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS group_members (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS service_accounts (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS federated_links (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS oauth_clients (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS client_secret_refs (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);
