-- M8 YDB schema baseline for m8-provisioning.
PRAGMA TablePathPrefix("/local/m8");

    -- Common operational tables for m8-provisioning.
    CREATE TABLE `m8-provisioning/outbox` (
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

    CREATE TABLE `m8-provisioning/inbox` (
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

    CREATE TABLE `m8-provisioning/idempotency_keys` (
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

    CREATE TABLE `m8-provisioning/operations` (
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


    -- DATA-PROV-001: ResourceDefinition; classification=internal; retention=all published versions.
    CREATE TABLE `m8-provisioning/resource_definitions` (
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
        INDEX idx_resource_definitions_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_resource_definitions_project GLOBAL ON (project_id, created_at),
        INDEX idx_resource_definitions_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-PROV-002: ManagedResource; classification=internal/confidential; retention=lifetime + 7y.
    CREATE TABLE `m8-provisioning/managed_resources` (
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
        INDEX idx_managed_resources_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_managed_resources_project GLOBAL ON (project_id, created_at),
        INDEX idx_managed_resources_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-PROV-003: DesiredState; classification=confidential; retention=current + revision history.
    CREATE TABLE `m8-provisioning/desired_states` (
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
        INDEX idx_desired_states_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_desired_states_project GLOBAL ON (project_id, created_at),
        INDEX idx_desired_states_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-PROV-004: ObservedState; classification=confidential; retention=rolling 90d + snapshots.
    CREATE TABLE `m8-provisioning/observed_states` (
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
        INDEX idx_observed_states_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_observed_states_project GLOBAL ON (project_id, created_at),
        INDEX idx_observed_states_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-PROV-005: Placement; classification=internal; retention=lifetime + 3y.
    CREATE TABLE `m8-provisioning/placements` (
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
        INDEX idx_placements_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_placements_project GLOBAL ON (project_id, created_at),
        INDEX idx_placements_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-PROV-006: Driver; classification=confidential; retention=active + 3y.
    CREATE TABLE `m8-provisioning/drivers` (
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
        INDEX idx_drivers_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_drivers_project GLOBAL ON (project_id, created_at),
        INDEX idx_drivers_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );
