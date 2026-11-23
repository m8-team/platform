-- Policy Consistency v2: run only after the original m8-access schema.sql.
-- TablePathPrefix is provided by the migration runner, not this DDL.
-- All PolicyMutation writes MUST share the authoritative Access state + Outbox YDB transaction.
CREATE TABLE `m8-access/policy_heads`
(
    tenant_id        Utf8      NOT NULL,
    head_revision    Uint64    NOT NULL,
    applied_revision Uint64    NOT NULL,
    updated_at       Timestamp NOT NULL,
    PRIMARY KEY (tenant_id)
);
CREATE TABLE `m8-access/policy_mutations`
(
    tenant_id       Utf8   NOT NULL,
    policy_revision Uint64 NOT NULL,
    mutation_id     Utf8   NOT NULL,
    total_jobs      Uint32 NOT NULL,
    completed       Bool   NOT NULL,
    PRIMARY KEY (tenant_id, policy_revision)
);
CREATE TABLE `m8-access/policy_jobs`
(
    tenant_id          Utf8   NOT NULL,
    policy_revision    Uint64 NOT NULL,
    projection_kind    Utf8   NOT NULL,
    aggregate_id       Utf8   NOT NULL,
    required_revision  Int64  NOT NULL,
    certified          Bool   NOT NULL,
    evidence_zed_token Utf8   NOT NULL,
    evidence_epoch     Uint64 NOT NULL,
    PRIMARY KEY (tenant_id, policy_revision, projection_kind, aggregate_id)
);
CREATE TABLE `m8-access/policy_request_ids`
(
    tenant_id       Utf8   NOT NULL,
    mutation_id     Utf8   NOT NULL,
    policy_revision Uint64 NOT NULL,
    request_hash    Utf8   NOT NULL,
    PRIMARY KEY (tenant_id, mutation_id)
);
