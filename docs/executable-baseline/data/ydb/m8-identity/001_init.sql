-- M8 YDB schema baseline for m8-identity.
PRAGMA TablePathPrefix("/local/m8");

    -- Common operational tables for m8-identity.
    CREATE TABLE `m8-identity/outbox` (
        event_id Utf8 NOT NULL,
        aggregate_type Utf8 NOT NULL,
        aggregate_id Utf8 NOT NULL,
        aggregate_revision Uint64 NOT NULL,
        event_type Utf8 NOT NULL,
        topic Utf8 NOT NULL,
        partition_key Utf8 NOT NULL,
        payload JsonDocument NOT NULL,
        headers JsonDocument,
        status Utf8 NOT NULL,
        attempt_count Uint32 NOT NULL,
        next_attempt_at Timestamp,
        created_at Timestamp NOT NULL,
        published_at Timestamp,
        PRIMARY KEY (event_id),
        INDEX idx_outbox_dispatch GLOBAL ON (status, next_attempt_at, created_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );

    CREATE TABLE `m8-identity/inbox` (
        consumer_name Utf8 NOT NULL,
        event_id Utf8 NOT NULL,
        event_type Utf8 NOT NULL,
        received_at Timestamp NOT NULL,
        processed_at Timestamp,
        status Utf8 NOT NULL,
        payload_hash String NOT NULL,
        error_code Utf8,
        PRIMARY KEY (consumer_name, event_id),
        INDEX idx_inbox_status GLOBAL ON (status, received_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );

    CREATE TABLE `m8-identity/idempotency_keys` (
        scope Utf8 NOT NULL,
        idempotency_key Utf8 NOT NULL,
        request_hash String NOT NULL,
        operation_name Utf8,
        response_payload JsonDocument,
        state Utf8 NOT NULL,
        created_at Timestamp NOT NULL,
        expires_at Timestamp NOT NULL,
        PRIMARY KEY (scope, idempotency_key),
        INDEX idx_idempotency_expiry GLOBAL ON (expires_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED
    );

    CREATE TABLE `m8-identity/operations` (
        operation_id Utf8 NOT NULL,
        operation_type Utf8 NOT NULL,
        state Utf8 NOT NULL,
        progress_percent Uint32 NOT NULL,
        stage Utf8,
        message Utf8,
        workflow_id Utf8,
        workflow_run_id Utf8,
        result JsonDocument,
        error JsonDocument,
        revision Uint64 NOT NULL,
        created_at Timestamp NOT NULL,
        updated_at Timestamp NOT NULL,
        completed_at Timestamp,
        PRIMARY KEY (operation_id),
        INDEX idx_operations_state GLOBAL ON (state, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-ID-001: UserPool; classification=internal; retention=active + 7y tombstone.
    CREATE TABLE `m8-identity/user_pools` (
        id Utf8 NOT NULL,
        parent_id Utf8,
        project_id Utf8,
        status Utf8 NOT NULL,
        display_name Utf8,
        attributes JsonDocument,
        labels JsonDocument,
        classification Utf8 NOT NULL,
        revision Uint64 NOT NULL,
        etag Utf8 NOT NULL,
        created_at Timestamp NOT NULL,
        updated_at Timestamp NOT NULL,
        deleted_at Timestamp,
        PRIMARY KEY (id),
        INDEX idx_user_pools_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_user_pools_project GLOBAL ON (project_id, created_at),
        INDEX idx_user_pools_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-ID-002: User; classification=personal; retention=purpose + legal retention.
    CREATE TABLE `m8-identity/users` (
        id Utf8 NOT NULL,
        parent_id Utf8,
        project_id Utf8,
        status Utf8 NOT NULL,
        display_name Utf8,
        attributes JsonDocument,
        labels JsonDocument,
        classification Utf8 NOT NULL,
        revision Uint64 NOT NULL,
        etag Utf8 NOT NULL,
        created_at Timestamp NOT NULL,
        updated_at Timestamp NOT NULL,
        deleted_at Timestamp,
        PRIMARY KEY (id),
        INDEX idx_users_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_users_project GLOBAL ON (project_id, created_at),
        INDEX idx_users_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-ID-003: ProfileAttribute; classification=personal/sensitive; retention=attribute policy.
    CREATE TABLE `m8-identity/profile_attributes` (
        id Utf8 NOT NULL,
        parent_id Utf8,
        project_id Utf8,
        status Utf8 NOT NULL,
        display_name Utf8,
        attributes JsonDocument,
        labels JsonDocument,
        classification Utf8 NOT NULL,
        revision Uint64 NOT NULL,
        etag Utf8 NOT NULL,
        created_at Timestamp NOT NULL,
        updated_at Timestamp NOT NULL,
        deleted_at Timestamp,
        PRIMARY KEY (id),
        INDEX idx_profile_attributes_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_profile_attributes_project GLOBAL ON (project_id, created_at),
        INDEX idx_profile_attributes_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-ID-004: ExternalIdentity; classification=personal; retention=active + audit reference.
    CREATE TABLE `m8-identity/external_identitys` (
        id Utf8 NOT NULL,
        parent_id Utf8,
        project_id Utf8,
        status Utf8 NOT NULL,
        display_name Utf8,
        attributes JsonDocument,
        labels JsonDocument,
        classification Utf8 NOT NULL,
        revision Uint64 NOT NULL,
        etag Utf8 NOT NULL,
        created_at Timestamp NOT NULL,
        updated_at Timestamp NOT NULL,
        deleted_at Timestamp,
        PRIMARY KEY (id),
        INDEX idx_external_identitys_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_external_identitys_project GLOBAL ON (project_id, created_at),
        INDEX idx_external_identitys_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-ID-005: Group; classification=internal; retention=active + 3y.
    CREATE TABLE `m8-identity/groups` (
        id Utf8 NOT NULL,
        parent_id Utf8,
        project_id Utf8,
        status Utf8 NOT NULL,
        display_name Utf8,
        attributes JsonDocument,
        labels JsonDocument,
        classification Utf8 NOT NULL,
        revision Uint64 NOT NULL,
        etag Utf8 NOT NULL,
        created_at Timestamp NOT NULL,
        updated_at Timestamp NOT NULL,
        deleted_at Timestamp,
        PRIMARY KEY (id),
        INDEX idx_groups_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_groups_project GLOBAL ON (project_id, created_at),
        INDEX idx_groups_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-ID-006: Membership; classification=internal; retention=active + 3y.
    CREATE TABLE `m8-identity/memberships` (
        id Utf8 NOT NULL,
        parent_id Utf8,
        project_id Utf8,
        status Utf8 NOT NULL,
        display_name Utf8,
        attributes JsonDocument,
        labels JsonDocument,
        classification Utf8 NOT NULL,
        revision Uint64 NOT NULL,
        etag Utf8 NOT NULL,
        created_at Timestamp NOT NULL,
        updated_at Timestamp NOT NULL,
        deleted_at Timestamp,
        PRIMARY KEY (id),
        INDEX idx_memberships_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_memberships_project GLOBAL ON (project_id, created_at),
        INDEX idx_memberships_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );
