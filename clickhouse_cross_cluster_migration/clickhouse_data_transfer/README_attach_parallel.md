# Parallel ATTACH runner for ClickHouse

`run_attach_partitions_parallel.py` runs multiple `ALTER TABLE ... ATTACH PARTITION ID ... FROM ...` operations concurrently, with **at most one operation per destination table**. It is compatible with the existing v3 SQL file, table-name mapping, and journal; no SQL changes or progress reset are needed.

## Contents

- `run_attach_partitions_parallel.py` — parallel executor (Python 3.10+, macOS/Linux)
- `run_attach_partitions_v3.py` — required parsing/journaling helpers and previous sequential executor
- `clickhouse_table_name_map.json` — target -> source table mapping

## Safety checklist

1. **Stop the old v3 runner.** This script uses the same `.progress.jsonl.lock` to prevent concurrent executions on the same computer.
2. **Make a copy** of `migration-sql/02_attach_partitions.sql.progress.jsonl` before starting. Do NOT modify `02_attach_partitions.sql`.
3. Run on the **same ClickHouse host, user, target and source** as in the existing journal. Do not load data independently on a second replica.
4. Keep the source database static until completion. Do not write to target tables from other applications during copying.
5. Do not delete the original `_import` database until all verification is complete.
6. A successful preflight verifies metadata and key compatibility, not all constraints (columns, projections, storage policies).
7. Network timeout or ClickHouse error **after `STARTED`** is marked `UNCERTAIN`, not automatically retried. Reconcile with `system.processes`, `system.query_log` and `system.parts`, then resolve with v3.
8. This uses multiple TLS connections and S3 is unrelated. Permissions are the same as for the existing v3 executor.

## Commands

Set secret without putting it in shell history:

```bash
read -rs CH_PASSWORD; echo; export CH_PASSWORD
```

Run a read-only preflight (no mutations):

```bash
python3 -u run_attach_partitions_parallel.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1a-mh2j1dseb1qsgb68.mdb.yandexcloud.net:8443 \
  --user analytics \
  --source zendendb_import \
  --target zendendb \
  --table-map ./clickhouse_table_name_map.json \
  --ca-file ./certs/RootCA.pem \
  --preflight
```

Try 40 operations across multiple tables, 4 concurrent workers:

```bash
python3 -u run_attach_partitions_parallel.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1a-mh2j1dseb1qsgb68.mdb.yandexcloud.net:8443 \
  --user analytics \
  --source zendendb_import \
  --target zendendb \
  --table-map ./clickhouse_table_name_map.json \
  --ca-file ./certs/RootCA.pem \
  --workers 4 --execute --limit 40
```

Then continue the remaining operations, without the limit:

```bash
python3 -u run_attach_partitions_parallel.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1a-mh2j1dseb1qsgb68.mdb.yandexcloud.net:8443 \
  --user analytics \
  --source zendendb_import \
  --target zendendb \
  --table-map ./clickhouse_table_name_map.json \
  --ca-file ./certs/RootCA.pem \
  --workers 4 --progress-every 50 --execute
```

When healthy, use `--workers 6` or `--workers 8` on a later restart. Do not change SQL or journal, and do not run two executors at once. Do not raise concurrency unless disk/CPU and replication queue are stable. For loaded clusters, `--workers 2` may be better.

### Status

When the runner **is stopped**:

```bash
python3 run_attach_partitions_parallel.py --sql ./migration-sql/02_attach_partitions.sql --status
```

When it is **running**, monitor its stdout or last lines of journal instead (status mode uses the exclusive lock):

```bash
tail -n 6 ./migration-sql/02_attach_partitions.sql.progress.jsonl
```

### Replication monitoring

```sql
SELECT table, total_replicas, active_replicas,
       queue_size, absolute_delay, is_readonly, is_session_expired
FROM system.replicas
WHERE database = 'zendendb'
ORDER BY queue_size DESC LIMIT 20;
```

```sql
SELECT count() AS running_attaches
FROM system.processes
WHERE startsWith(query_id, 'migrate-attach-');
```

### On an error

The runner stops scheduling new ATTACH operations; already-running requests finish and write journal records. After that, read the journal and investigate the problem. If there is `UNCERTAIN` or `STARTED`, do **not** immediately rerun or delete it. Resolve with the v3 runner, using its `--resolve done` or `--resolve retry` procedure after verifying ClickHouse state. A mere presence of an active target partition does not prove full success if partial data already existed.

If no unresolved operations remain, rerun the same command. Completed operations will be skipped.

### Why it may not speed up linearly

Copying 53k partitions also taxes filesystem metadata, Keeper and replication. More parallelism can become slower after saturation, and 53k partition-level actions may reflect an overly granular PARTITION BY design. Avoid changing the partition key within this ATTACH migration: different partition keys are incompatible with ATTACH PARTITION FROM.

## Testing

The parallel runner was tested locally with a fake ClickHouse client for concurrency, one-at-a-time per table, journal compatibility, resuming, and uncertain-error handling. It has not been executed against your production Yandex Managed ClickHouse cluster.
