BEGIN;
CREATE SCHEMA IF NOT EXISTS platform;
CREATE SCHEMA IF NOT EXISTS product;
CREATE TABLE platform.idempotency (
 tenant text NOT NULL, operation text NOT NULL, key text NOT NULL,
 request_hash bytea NOT NULL CHECK (octet_length(request_hash)=32),
 status text NOT NULL CHECK (status IN ('PROCESSING','SUCCEEDED')),
 result bytea, created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 updated_at timestamptz NOT NULL DEFAULT clock_timestamp(), expires_at timestamptz NOT NULL,
 PRIMARY KEY (tenant,operation,key)
);
CREATE INDEX idempotency_expiry ON platform.idempotency(expires_at);
CREATE TABLE platform.inbox (
 consumer text NOT NULL, message_id text NOT NULL,
 processed_at timestamptz NOT NULL DEFAULT clock_timestamp(), PRIMARY KEY(consumer,message_id)
);
CREATE INDEX inbox_retention ON platform.inbox(processed_at);
CREATE TABLE platform.outbox (
 id text PRIMARY KEY, topic text NOT NULL, aggregate_type text NOT NULL,
 aggregate_id text NOT NULL, event_type text NOT NULL, payload bytea NOT NULL,
 headers jsonb NOT NULL DEFAULT '{}', created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 published_at timestamptz, attempts integer NOT NULL DEFAULT 0,
 next_attempt_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 lease_token text, lease_until timestamptz, dead_at timestamptz
);
CREATE INDEX outbox_pending ON platform.outbox(next_attempt_at,created_at,id) WHERE published_at IS NULL AND dead_at IS NULL;
CREATE TABLE product.products (
 id text PRIMARY KEY, tenant text NOT NULL, name text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX products_tenant ON product.products(tenant,id);
COMMIT;
