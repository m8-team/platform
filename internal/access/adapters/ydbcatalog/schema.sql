-- Universal Access Catalog v1. Apply after creating the m8-access directory.
-- Execute under an explicitly configured TablePathPrefix; NEVER migrate at server startup.
-- These tables are independent of the tenant policy/grant journal. Catalog publication
-- does not change entitlements, SpiceDB relationship state, or applied policy barriers.
CREATE TABLE `m8-access/catalog_heads`
(
  module_id     Utf8      NOT NULL,
  head_revision Int64     NOT NULL,
  updated_at    Timestamp NOT NULL,
  PRIMARY KEY (module_id)
);
CREATE TABLE `m8-access/catalog_releases`
(
  module_id     Utf8      NOT NULL,
  version       Utf8      NOT NULL,
  revision      Int64     NOT NULL,
  manifest_json Utf8      NOT NULL,
  manifest_hash Utf8      NOT NULL,
  created_at    Timestamp NOT NULL,
  PRIMARY KEY (module_id, version)
);
CREATE TABLE `m8-access/catalog_receipts`
(
  module_id    Utf8  NOT NULL,
  request_key  Utf8  NOT NULL,
  request_hash Utf8  NOT NULL,
  version      Utf8  NOT NULL,
  revision     Int64 NOT NULL,
  PRIMARY KEY (module_id, request_key)
);
-- Query-side materialization: strong, same-transaction projection.
CREATE TABLE `m8-access/catalog_read_index`
(
  module_id     Utf8  NOT NULL,
  revision      Int64 NOT NULL,
  version       Utf8  NOT NULL,
  manifest_json Utf8  NOT NULL,
  PRIMARY KEY (module_id, revision)
);
-- Durable at-least-once event delivery. Consumers MUST deduplicate event_id.
CREATE TABLE `m8-access/catalog_outbox`
(
  module_id    Utf8      NOT NULL,
  revision     Int64     NOT NULL,
  event_id     Utf8      NOT NULL,
  event_kind   Utf8      NOT NULL,
  payload_json Utf8      NOT NULL,
  published    Bool      NOT NULL,
  created_at   Timestamp NOT NULL,
  PRIMARY KEY (module_id, revision)
);
