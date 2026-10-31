-- Preserve ResourceMetadata labels and soft-deletion time in the unified catalog.
-- Columns are nullable so the migration remains compatible with existing rows.
PRAGMA TablePathPrefix("/local/m8");

ALTER TABLE `m8-resource-manager/resource_records`
ADD COLUMN labels_json Utf8;

ALTER TABLE `m8-resource-manager/resource_records`
ADD COLUMN deleted_at Timestamp;
