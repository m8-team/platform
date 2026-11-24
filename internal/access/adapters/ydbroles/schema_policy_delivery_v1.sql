-- Access Profile Policy delivery index v1 (additive migration).
-- Primary source event remains profile_policy_events. This index is written
-- in the SAME serializable transaction as the authoritative policy mutation.
-- Do not run while old writers that omit this index can still issue changes.
CREATE TABLE `m8-access/profile_policy_dispatch`
(
  tenant_id Utf8 NOT NULL,
  policy_revision Uint64 NOT NULL,
  assignment_id Utf8 NOT NULL,
  projection_revision Int64 NOT NULL,
  event_id Utf8 NOT NULL,
  published Bool NOT NULL,
  PRIMARY KEY (tenant_id, policy_revision, assignment_id)
);
