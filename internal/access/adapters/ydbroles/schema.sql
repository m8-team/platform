-- Universal Access profiles and staged assignments, additive migration.
-- Apply only in a reviewed migration runner under a trusted TablePathPrefix.
-- No table here is a PDP source. Only certified policy journal projections can authorize.
CREATE TABLE `m8-access/role_bundle_heads`
(
  profile_name Utf8 NOT NULL,
  head_revision Int64 NOT NULL,
  updated_at Timestamp NOT NULL,
  PRIMARY KEY (profile_name)
);
CREATE TABLE `m8-access/role_bundles`
(
  profile_name Utf8 NOT NULL,
  version Utf8 NOT NULL,
  revision Int64 NOT NULL,
  bundle_json Utf8 NOT NULL,
  bundle_hash Utf8 NOT NULL,
  created_at Timestamp NOT NULL,
  PRIMARY KEY (profile_name, version)
);
CREATE TABLE `m8-access/role_bundle_receipts`
(
  profile_name Utf8 NOT NULL,
  request_key Utf8 NOT NULL,
  request_hash Utf8 NOT NULL,
  version Utf8 NOT NULL,
  revision Int64 NOT NULL,
  PRIMARY KEY (profile_name, request_key)
);
CREATE TABLE `m8-access/role_bundle_read_index`
(
  profile_name Utf8 NOT NULL,
  revision Int64 NOT NULL,
  version Utf8 NOT NULL,
  bundle_json Utf8 NOT NULL,
  PRIMARY KEY (profile_name, revision)
);
CREATE TABLE `m8-access/profile_assignments`
(
  tenant_id Utf8 NOT NULL,
  assignment_id Utf8 NOT NULL,
  revision Int64 NOT NULL,
  status Utf8 NOT NULL,
  assignment_json Utf8 NOT NULL,
  PRIMARY KEY (tenant_id, assignment_id)
);
CREATE TABLE `m8-access/profile_assignment_subjects`
(
  tenant_id Utf8 NOT NULL,
  subject_type Utf8 NOT NULL,
  subject_id Utf8 NOT NULL,
  assignment_id Utf8 NOT NULL,
  revision Int64 NOT NULL,
  status Utf8 NOT NULL,
  assignment_json Utf8 NOT NULL,
  PRIMARY KEY (tenant_id, subject_type, subject_id, assignment_id)
);
CREATE TABLE `m8-access/profile_assignment_receipts`
(
  tenant_id Utf8 NOT NULL,
  assignment_id Utf8 NOT NULL,
  request_key Utf8 NOT NULL,
  request_hash Utf8 NOT NULL,
  revision Int64 NOT NULL,
  PRIMARY KEY (tenant_id, assignment_id, request_key)
);
CREATE TABLE `m8-access/role_management_outbox`
(
  event_stream Utf8 NOT NULL,
  tenant_id Utf8 NOT NULL,
  aggregate_id Utf8 NOT NULL,
  revision Int64 NOT NULL,
  event_id Utf8 NOT NULL,
  event_kind Utf8 NOT NULL,
  payload_json Utf8 NOT NULL,
  published Bool NOT NULL,
  created_at Timestamp NOT NULL,
  PRIMARY KEY (event_stream, tenant_id, aggregate_id, revision)
);

-- Profile graph source is committed with the authoritative state and tenant policy
-- head. Zero capabilities on revoked records express a tombstone for reconciliation.
CREATE TABLE `m8-access/profile_assignment_projection`
(
  tenant_id Utf8 NOT NULL,
  assignment_id Utf8 NOT NULL,
  revision Int64 NOT NULL,
  capabilities_json Utf8 NOT NULL,
  PRIMARY KEY (tenant_id, assignment_id)
);
-- Dispatch to policyjournal.Worker only; this is NOT a general audit outbox.
-- Event may be ACKed only AFTER a fenced SpiceDB checkpoint and certification.
CREATE TABLE `m8-access/profile_policy_events`
(
  tenant_id Utf8 NOT NULL,
  assignment_id Utf8 NOT NULL,
  revision Int64 NOT NULL,
  policy_revision Uint64 NOT NULL,
  event_id Utf8 NOT NULL,
  payload_json Utf8 NOT NULL,
  published Bool NOT NULL,
  created_at Timestamp NOT NULL,
  PRIMARY KEY (tenant_id, assignment_id, revision)
);
