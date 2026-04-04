-- Async infrastructure.

CREATE TABLE outbox_events (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE topic_offsets (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE projection_checkpoints (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);

CREATE TABLE workflow_locks (
  id Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  payload JsonDocument,
  created_at Timestamp,
  updated_at Timestamp,
  PRIMARY KEY (tenant_id, id)
);
