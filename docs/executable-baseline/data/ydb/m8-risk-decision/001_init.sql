-- M8 YDB schema baseline for m8-risk-decision.
PRAGMA TablePathPrefix("/local/m8");

    -- Common operational tables for m8-risk-decision.
    CREATE TABLE `m8-risk-decision/outbox` (
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

    CREATE TABLE `m8-risk-decision/inbox` (
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

    CREATE TABLE `m8-risk-decision/idempotency_keys` (
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

    CREATE TABLE `m8-risk-decision/operations` (
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


    -- DATA-RISK-001: RiskAssessment; classification=sensitive; retention=180d default.
    CREATE TABLE `m8-risk-decision/risk_assessments` (
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
        INDEX idx_risk_assessments_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_risk_assessments_project GLOBAL ON (project_id, created_at),
        INDEX idx_risk_assessments_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-RISK-002: RiskPolicy; classification=confidential; retention=all published versions.
    CREATE TABLE `m8-risk-decision/risk_policys` (
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
        INDEX idx_risk_policys_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_risk_policys_project GLOBAL ON (project_id, created_at),
        INDEX idx_risk_policys_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-RISK-003: RiskSignal; classification=sensitive; retention=30-180d by type.
    CREATE TABLE `m8-risk-decision/risk_signals` (
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
        INDEX idx_risk_signals_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_risk_signals_project GLOBAL ON (project_id, created_at),
        INDEX idx_risk_signals_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );


    -- DATA-RISK-004: ManualReview; classification=sensitive; retention=7y.
    CREATE TABLE `m8-risk-decision/manual_reviews` (
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
        INDEX idx_manual_reviews_parent GLOBAL ON (parent_id, created_at),
        INDEX idx_manual_reviews_project GLOBAL ON (project_id, created_at),
        INDEX idx_manual_reviews_status GLOBAL ON (status, updated_at)
    ) WITH (
        AUTO_PARTITIONING_BY_SIZE = ENABLED,
        AUTO_PARTITIONING_BY_LOAD = ENABLED
    );
