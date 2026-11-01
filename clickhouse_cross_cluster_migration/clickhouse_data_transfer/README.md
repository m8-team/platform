# ClickHouse: safe recovery of partial months in `zendendb.ЦеныНоменклатуры`

## Problem

The first bulk `INSERT SELECT remoteSecure(...)` failed with `TOO_MANY_PARTS` because the source table partitions by **day** (`toYYYYMMDD(Период)`). It left partial **August 2026** (6,658 of 13,197 rows) and **September 2026** (7,119 of 7,121 rows) on the target cluster. `transfer_monthly.py` correctly refuses to insert on top of them. **Do not retry the same `INSERT`, raise the partition limit, truncate or drop target data.**

## Strategy

Build a NEW, separate replicated table `zendendb.ЦеныНоменклатуры__repair` via `CREATE TABLE ... AS existing_target`. Copy the complete source (45 calendar months) using INSERTs limited to ONE calendar month (max. 31 daily partitions per query). Verify *every month* and *full-table fingerprints* on **BOTH** replicas. Then, in a separate planned cutover, atomically exchange the table names (provided this form is supported on your server; it worked with the preceding `ПланПродаж_ecom` migration). The original partial target is retained under the `__repair` name for rollback.

This script **never modifies the existing target table** and **never changes the main migration journal**. It creates/writes only the staging table; does not execute `EXCHANGE`, `DROP`, `TRUNCATE`, `RENAME`, `REPLACE PARTITION`, or `ALTER`.

**Before using**: stop source writes and ALL target writes. Check there is sufficient free disk on **both** target replicas for another full copy of this table. This is a `MergeTree`, not `ReplacingMergeTree`: copying all physical rows is the intended behavior. Keep original source data until cutover is checked.

## Setup

Place these two Python files beside your current `config.json`:

- `repair_prices_stage.py`
- `transfer_monthly.py` (the compatible helper already supplied in v6)

Python >=3.9; no third-party packages. Do NOT replace your currently working `transfer_data.py` or remove `data_transfer_progress/progress.json`.

## 1. Inspect (read-only)

```sh
python3 repair_prices_stage.py --config config.json --inspect
```

Expected: stage does not yet exist; original target has partial data from the earlier failed insert.

## 2. Prepare one pilot month (writes only to new staging table)

```sh
python3 repair_prices_stage.py \
  --config config.json \
  --prepare \
  --ack-source-frozen \
  --limit-months 1
```

If no stage exists, the script creates it in the replicated database, confirms a Keeper path distinct from the original and healthy replicas, and inserts the first available month. It checks per-month row counts and two content checksums on each replica.

## 3. Copy all remaining months

```sh
python3 repair_prices_stage.py \
  --config config.json \
  --prepare \
  --ack-source-frozen
```

All complete stage months are skipped. If a month is partially loaded, differs, or has ambiguous journal status, the script refuses a blind retry. Stage journal: `data_transfer_progress/prices_stage_progress.json`.

## 4. Independently verify stage (read-only)

```sh
python3 repair_prices_stage.py \
  --config config.json \
  --verify \
  --ack-source-frozen
```

Requires all source months to match staging on **both replicas** and full-table source/stage fingerprint to match on both replicas. Expected `STAGING VERIFIED`. No production data mutation.

## 5. Planned manual cutover (ONLY after STAGING VERIFIED)

Freeze all writes to source and target again, re-run Step 4, confirm staging health/free disk, and run on **one** target host (no `ON CLUSTER` because database is `ENGINE=Replicated`):

```sql
EXCHANGE TABLES
    `zendendb`.`ЦеныНоменклатуры`
AND
    `zendendb`.`ЦеныНоменклатуры__repair`;
```

**Caution**: Confirm EXCHANGE support and the maintenance-window procedure on your exact Yandex ClickHouse version. This statement is NOT executed by the script.

After EXCHANGE, `ЦеныНоменклатуры` points to the fully verified copy. The old partial target is retained as `ЦеныНоменклатуры__repair` for rollback. **Do not re-run `repair_prices_stage.py --prepare` after exchange**, since the stage name will then refer to the old partial data.

## 6. Reconcile the main migration journal

```sh
python3 transfer_data.py \
  --config config.json \
  --only 'zendendb.ЦеныНоменклатуры' \
  --reconcile \
  --mark-done \
  --ack-source-frozen
```

This should move the existing `UNCERTAIN` main-journal state to `DONE` only after a full-table fingerprint match. Do NOT re-INSERT over an existing table. Keep the old `__repair` copy until after final audit.

## Why not copy only the missing 6,539 and 2 rows?

Counts alone do not identify missing physical records. Row duplicates may be valid in a plain `MergeTree`, and blindly doing an anti-join can silently lose or double rows. An exact multiset diff could be implemented, but rebuilding in a staging table is more deterministic and leaves the production target untouched until an explicitly reviewed switch.

## Security and limitations

- `remoteSecure(...)` credentials are sent inside SQL to the target and might be recorded in query/audit logs. Use least-privileged short-lived source credentials and rotate as appropriate.
- Exact types, server timezone, primary/sorting/partitioning keys, stage Keeper isolation and replica health are checked prior to writing.
- Counts + 2 hashes offer a **probabilistic** comparison; not a formal proof of all bytes.
- Each month scan re-reads source data. This is a read-intensive migration.
- Only supports the hardcoded `zendendb.ЦеныНоменклатуры` schema on this 1-shard 2-replica target. No online write synchronization.
- Test suite is local and mock-based. No actual Yandex ClickHouse cluster was contacted here.
