-- Execute within the configured YDB database (absolute TablePathPrefix is supplied by deployment).
-- Do not run during a rollout that still has a single-host file checkpoint writer.
CREATE TABLE `m8-access/projection_checkpoints` (
    tenant_id Utf8 NOT NULL,
    projection_kind Utf8 NOT NULL,
    aggregate_id Utf8 NOT NULL,
    revision Int64 NOT NULL,
    event_id Utf8 NOT NULL,
    zed_token Utf8 NOT NULL,
    fence_epoch Uint64 NOT NULL,
    edges_json Utf8 NOT NULL,
    updated_at Timestamp NOT NULL,
    lease_epoch Uint64 NOT NULL,
    lease_holder Utf8 NOT NULL,
    lease_until Timestamp NOT NULL,
    PRIMARY KEY (tenant_id, projection_kind, aggregate_id)
);
CREATE TABLE `m8-access/resource_projection_watermarks` (
    tenant_id Utf8 NOT NULL,
    resource_kind Utf8 NOT NULL,
    resource_id Utf8 NOT NULL,
    source_revision Uint64 NOT NULL,
    zed_token Utf8 NOT NULL,
    updated_at Timestamp NOT NULL,
    PRIMARY KEY (tenant_id, resource_kind, resource_id)
);
