# ClickHouse `ATTACH PARTITION` runner v3 — renamed table support

This release fixes source and target table names that differ **without modifying** `02_attach_partitions.sql` or its existing `.progress.jsonl` journal. No additional Python dependencies; Python 3.10+.

## Identified mapping

- Source: `zendendb_import.\`ДокументыНаОснованииИнтернетЗаказов\`` (`MergeTree`)
- Destination: `zendendb.\`ДокументыНаОснованииИнтернет\`` (`ReplicatedMergeTree`)
- Requested partition from the failed step: `20200930`.

`clickhouse_table_name_map.json` uses `{"target_table": "source_table"}`. You can add other known renames, but **do not change a mapping for an already completed step**.

## Safety

- Keep the original SQL and journal unchanged, including their filenames.
- Stop any previous runner process before launching v3.
- Take a copy of your progress journal before attempting the migration.
- Use the exact same URL and user as the journal's `META` entry.
- The script validates SQL file SHA-256, target database engine, server identity, source and destination engines, partition/sorting/primary keys and presence/absence of the source/destination partition before each ATTACH.
- For renamed tables only, it replaces the `FROM` source table name in the SQL sent to ClickHouse. The original SQL and original per-statement SHA remain unchanged.
- If a command times out, its status is **uncertain**. Never retry blindly; reconcile with ClickHouse `system.parts`, `system.processes`, and `system.query_log`.
- Never run this simultaneously on multiple replicas; the target table handles data replication.
- Complete structure compatibility, index/projection and storage policies are still checked by ClickHouse itself during ATTACH. A successful preflight does not guarantee the ALTER succeeds.
- Ensure the source remains static throughout the migration; data changes/merges can affect row counts.

## First: inspect the mapping (no DB connection)

```bash
python3 run_attach_partitions_v3.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --source zendendb_import --target zendendb \
  --table-map ./clickhouse_table_name_map.json
```

## Next: preflight (read only)

```bash
python3 -u run_attach_partitions_v3.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1a-mh2j1dseb1qsgb68.mdb.yandexcloud.net:8443 \
  --user analytics \
  --source zendendb_import \
  --target zendendb \
  --table-map ./clickhouse_table_name_map.json \
  --ca-file ./certs/RootCA.pem \
  --preflight
```

## Resume: first one operation, then everything

```bash
python3 -u run_attach_partitions_v3.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1a-mh2j1dseb1qsgb68.mdb.yandexcloud.net:8443 \
  --user analytics \
  --source zendendb_import \
  --target zendendb \
  --table-map ./clickhouse_table_name_map.json \
  --ca-file ./certs/RootCA.pem \
  --execute --limit 1
```

Then the **same command** without `--limit 1` to run all remaining operations. Previously completed steps remain tracked.

## After that: validation for the renamed table

```sql
SELECT
    database,
    table,
    partition_id,
    sum(rows) AS rows,
    count() AS active_parts
FROM system.parts
WHERE active
  AND (
    (database = 'zendendb_import' AND table = 'ДокументыНаОснованииИнтернетЗаказов')
    OR
    (database = 'zendendb' AND table = 'ДокументыНаОснованииИнтернет')
  )
  AND partition_id = '20200930'
GROUP BY database, table, partition_id
ORDER BY database;
```

The source and destination row counts should normally match after the partition attaches if the source is static and no relevant merges/TTL have modified the target.

## Test status

This version passed local unit tests for mapping and resumable execution with an existing journal. It was **not** executed against the live ClickHouse server.
