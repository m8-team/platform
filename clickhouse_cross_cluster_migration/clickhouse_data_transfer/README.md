# ClickHouse cross-cluster migration — daily-partition staging repair v9

This package fixes the repeated ClickHouse error `TOO_MANY_PARTS` / `max_partitions_per_insert_block = 100` caused by a single `INSERT SELECT` touching too many source partitions. Source/target are the previously configured Yandex ClickHouse clusters; target is a **Replicated** database with two **ReplicatedMergeTree** replicas. Does NOT change original data or table names without an explicit manual DBA operation.

## Directly affected tables

The original DDLs for both tables are plain `MergeTree PARTITION BY toYYYYMMDD(\`Период\`)`:

- `zendendb.ЗаказыПокупателей`
- `zendendb.ТоварыВРезервеНаСкладах`

Both have `UNCERTAIN` in the **main** migration journal after failed bulk INSERT, so **do not retry bulk INSERT**. Some rows may already be on both target replicas. The new tool creates independent staging table `<table>__repair` and copies the complete source by **month** (at most 31 day partitions per INSERT). It checks source schema, target schema, source and target timezone, separate Keeper path, replicated state and full table fingerprints. It does NOT touch the original target or main transfer journal, and does not perform EXCHANGE or any destructive SQL.

## Install

Copy these files into your existing `clickhouse_data_transfer` directory, next to `config.json`:

- `repair_daily_stage.py` — **new** reusable repair for daily-partitioned plain MergeTree tables.
- `transfer_monthly.py` — helper for HTTPS + ClickHouse replica checks; same API as prior v6/v7.
- `transfer_data.py` — **optional replacement** of your existing main migrator; all previous v5 features retained plus the `--skip-high-partitions` guard.

**Keep** your current `config.json`, `data_transfer_progress/progress.json`, other stage journals and existing repair scripts. Never overwrite or delete your journal. Python 3.9+ and standard library only.

## Before writing

Stop all source ingestion and writes, and stop target writes to tables being migrated. Confirm adequate target disk space **on each replica** for a complete additional table copy. Confirm your current target is indeed the intended new ClickHouse cluster. Keep source untouched until cutover is verified. The fact a command contains `--ack-source-frozen` does not actually freeze writes: you must enforce it operationally.

## 1. Source and target inspection (read-only)

```sh
python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ЗаказыПокупателей' --inspect

python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ТоварыВРезервеНаСкладах' --inspect
```

Each table gets its own independently named `<table>__repair` stage and own journal in `data_transfer_progress/daily_stage_<hash>.json`. The script refuses to use a nonmatching stage.

## 2. Pilot one month into staging

```sh
python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ЗаказыПокупателей' \
  --prepare --ack-source-frozen --limit-months 1
```

Expected: created staging table and `DONE` for one month on **both** replicas. If any month is PARTIAL, CONFLICT or prior write status UNCERTAIN, the tool stops; no blind retry.

## 3. Finish stage copy

```sh
python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ЗаказыПокупателей' \
  --prepare --ack-source-frozen

python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ТоварыВРезервеНаСкладах' \
  --prepare --ack-source-frozen
```

The second table may also be piloted first using `--limit-months 1`. Previously verified stage months are skipped. When all months match, script prints `STAGING VERIFIED`.

## 4. Verify each stage independently (read-only)

```sh
python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ЗаказыПокупателей' \
  --verify --ack-source-frozen

python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ТоварыВРезервеНаСкладах' \
  --verify --ack-source-frozen
```

This checks physical row count and two content hashes per month, whole-table hashes, replica health, and matching row hashes **on each replica**. `STAGING VERIFIED` is required. Hash checks are strong evidence, not mathematical proof of identical data.

## 5. Cutover — only after `STAGING VERIFIED`

In a scheduled maintenance window, stop all source and target writes, run `--verify` again, ensure no dependent queries or materialized views are writing to original tables, confirm `EXCHANGE` support on the target database, then run the following *manually* against **one target host** (no `ON CLUSTER` with `ENGINE=Replicated` database):

```sql
EXCHANGE TABLES `zendendb`.`ЗаказыПокупателей`
AND `zendendb`.`ЗаказыПокупателей__repair`;

EXCHANGE TABLES `zendendb`.`ТоварыВРезервеНаСкладах`
AND `zendendb`.`ТоварыВРезервеНаСкладах__repair`;
```

Perform each EXCHANGE separately only after the corresponding staging table is verified. The old partial target data is retained under `__repair` for rollback. **Do NOT rerun `repair_daily_stage.py --prepare` / `--verify` after exchange**: the stage name now denotes the old partial table. Verify the new production names on both replicas and use `transfer_data.py --reconcile --mark-done` to close the main journal for each:

```sh
python3 transfer_data.py --config config.json \
  --only 'zendendb.ЗаказыПокупателей' \
  --reconcile --mark-done --ack-source-frozen

python3 transfer_data.py --config config.json \
  --only 'zendendb.ТоварыВРезервеНаСкладах' \
  --reconcile --mark-done --ack-source-frozen
```

`--reconcile` must match the original physical row count and fingerprints on both target replicas before writing `DONE`. Do not mark them `DONE` manually.

## Prevent more failures in remaining pending tables

The packaged `transfer_data.py` adds a non-destructive pre-insert guard for any source table with more than 100 distinct **active** partitions (read from `system.parts`). Those are conservatively skipped *before* `COPYING` / `INSERT`; existing `UNCERTAIN` and unversioned `ReplacingMergeTree` duplicates remain skipped using the prior flags:

```sh
python3 transfer_data.py --config config.json \
  --apply --ack-source-frozen --resume \
  --skip-uncertain --skip-no-version-duplicates \
  --skip-high-partitions
```

This **does not automatically migrate the skipped high-partition tables**. A table with >100 partitions may be safe for some specific block sizes, but the guard chooses safety over throughput. For `PARTITION BY toYYYYMMDD(Период)` + `MergeTree`, reuse `repair_daily_stage.py`. For string-based partitions (`Товары` etc.), use your separate `repair_products_stage.py`. For `ReplacingMergeTree` with ambiguous winners, use a separately reviewed `FINAL` staging approach. For unknown partition expressions, do not force the daily-month script.

The `--skip-high-partitions` heuristic is not a substitute for controlling `max_partitions_per_insert_block`: another query can exceed that limit for a different reason. Keep every skipped or ambiguous table in the final cutover checklist.

## Verification, limitations and operational notes

- No live database was contacted for this release; tests use mocked clients. Test suite: `python3 -m unittest discover -p 'test_*.py' -v`.
- Data hashing is expensive for very large tables and done on two replicas. Streaming server-side `remoteSecure` reads require TCP 9440 reachability.
- `remoteSecure` includes source credentials inside the server-side SQL; ensure query log/audit access is appropriately restricted and use a temporary read-only source account.
- Before mass cutover, separately resolve the long missing table name `zendendb.ДокументыНаОснованииИнтернетЗаказов`, remaining unversioned ReplacingMergeTree duplicates, incomplete stage repairs and unsupported source engines/views.
