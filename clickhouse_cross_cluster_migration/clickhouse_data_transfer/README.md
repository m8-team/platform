# ClickHouse daily-partition staging transfer — v10

## Why v9 failed on `zendendb.ИнтернетЗаказПокупателя`

The original source DDL is plain `MergeTree PARTITION BY toYYYYMMDD(`Дата`)`. A single bulk `INSERT ... remoteSecure` can exceed `max_partitions_per_insert_block=100`. If it fails, main transfer journal remains `UNCERTAIN`; any rows already inserted into the original target table must be preserved pending reconciliation.

v10:
- `repair_daily_stage.py` works with any simple `toYYYYMMDD(column)` key, including **`Дата`** and **`Период`**. It discovers the column from `system.tables.partition_key`, validates it against `DESCRIBE`, compares source and destination DDL and timezones, then does one `INSERT` per calendar month (at most 31 daily partitions).
- `transfer_data.py --skip-high-partitions` now **always skips simple daily-partitioned tables before INSERT**, even when `system.parts` claims 100 or fewer existing partitions. It still skips other tables whose `system.parts` count is greater than 100. This remains a conservative heuristic for non-daily partition schemes.
- Neither script retries ambiguous inserts or changes the main journal unless a user explicitly runs verified `--reconcile --mark-done`.

## Installation

Replace `repair_daily_stage.py` and optionally `transfer_data.py` with the v10 copies. Keep existing `config.json`, all `progress.json` journals, other scripts and staging tables unchanged. Copy the bundled compatible `transfer_monthly.py` if not already present. Do not reset or delete any journal.

## `zendendb.ИнтернетЗаказПокупателя` — safely stage from the beginning

First inspect (**read only**):

```sh
python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ИнтернетЗаказПокупателя' --inspect
```

At this point the ORIGINAL target is in an `UNCERTAIN` state; this stage process never touches it. You must actually stop writes to source and target; `--ack-source-frozen` alone does not stop anything. Ensure disk space on both target replicas, and check there are no attached MVs that would be triggered by staging insertion.

Pilot one month:

```sh
python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ИнтернетЗаказПокупателя' \
  --prepare --ack-source-frozen --limit-months 1
```

Finish remaining months (already verified months are skipped):

```sh
python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ИнтернетЗаказПокупателя' \
  --prepare --ack-source-frozen
```

Fresh independent verification:

```sh
python3 repair_daily_stage.py --config config.json \
  --table 'zendendb.ИнтернетЗаказПокупателя' \
  --verify --ack-source-frozen
```

Only `STAGING VERIFIED` means data has matched for each source month, entire table and both target replicas. The staging name is `zendendb.ИнтернетЗаказПокупателя__repair` and its Keeper path is verified separate. The original partial target remains unchanged.

## Manual cutover (not performed by this script)

After a fresh `--verify`, stop all readers/writers for the short cutover, confirm your Yandex ClickHouse Replicated database supports `EXCHANGE`, and use one target host, **without `ON CLUSTER`**:

```sql
EXCHANGE TABLES
  `zendendb`.`ИнтернетЗаказПокупателя`
AND
  `zendendb`.`ИнтернетЗаказПокупателя__repair`;
```

The old partial target table remains under the `__repair` name. Do not run stage `--verify`/`--prepare` again after exchange because the names have swapped.

Then validate and mark the already `UNCERTAIN` main journal entry `DONE`, WITHOUT a fresh INSERT:

```sh
python3 transfer_data.py --config config.json \
  --only 'zendendb.ИнтернетЗаказПокупателя' \
  --reconcile --mark-done --ack-source-frozen
```

If and only if fingerprints match on both replicas, the migrator will mark it `DONE`. Keep rollback backup until overall migration is verified.

## Prevent failures in remaining tables

```sh
python3 transfer_data.py --config config.json \
  --apply --ack-source-frozen --resume \
  --skip-uncertain --skip-no-version-duplicates --skip-high-partitions
```

**Important:** This command intentionally skips ALL daily-partitioned tables and those with >100 observed partitions, leaving them for separate stage transfer. It also skips ambiguous ReplacingMergeTree cases. Skips are not a completed migration. This v10 safeguard does not make bulk transfer universally safe for arbitrary custom partition keys; carefully review any further exceptions and cut over only when all required tables are reconciled.

## Limitations

No actual source or destination connection was made. Tests are local/mocked. The script uses an exact complete set of month fingerprints with row counts and two aggregate hashes, and enforces two healthy replicas. A two-hash comparison is strong evidence, not absolute proof. `remoteSecure` places source credentials within server-side query SQL; use least-privilege credentials and audit restrictions.
