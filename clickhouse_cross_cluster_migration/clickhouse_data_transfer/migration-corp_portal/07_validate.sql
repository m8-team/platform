-- Check data and replication. Part counts can vary for ReplacingMergeTree due to merges.
-- Review and execute only when the previous stage succeeded.

SELECT name, engine FROM system.databases WHERE name IN ('corp_portal_import', 'corp_portal');

SELECT database, engine, count() AS objects FROM system.tables WHERE database IN ('corp_portal_import', 'corp_portal') GROUP BY database, engine ORDER BY database, engine;

SELECT database, table, sum(rows) AS active_part_rows, formatReadableSize(sum(bytes_on_disk)) AS size FROM system.parts WHERE database IN ('corp_portal_import', 'corp_portal') AND active GROUP BY database, table ORDER BY table, database;

SELECT database, table, is_readonly, is_session_expired, queue_size, absolute_delay FROM system.replicas WHERE database = 'corp_portal' ORDER BY table;

