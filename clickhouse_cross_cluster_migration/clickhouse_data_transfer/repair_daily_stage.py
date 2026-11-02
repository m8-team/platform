#!/usr/bin/env python3
"""Read-only source -> independent staging for MergeTree daily partitions.

Generic across tables with the validated schema/engine/partitioning pattern.
NEVER touches current target table, source, or main migration journal.
Copies one month per INSERT; audits both replicas, rows and checksums.
"""
from __future__ import annotations

import argparse
import hashlib
import re
import json
import sys
import time
from pathlib import Path

from transfer_monthly import (Client, MigrationError, classify, normalize, qi, qs,
                              require_replicas, save_state, utcnow, redact)

# Bound once at CLI startup; functions take no table input from untrusted SQL.
DB = NAME = STAGE = PART_COL = None
SRC = DST = STG = None


def select_table(full_name):
    """Select a target table, safely quoting identifiers in every SQL statement."""
    global DB, NAME, STAGE, PART_COL, SRC, DST, STG
    # Accept Cyrillic, spaces, etc., but require one database/table separator.
    if not isinstance(full_name, str) or full_name.count('.') != 1:
        raise MigrationError('--table must have database.table syntax')
    db, name = full_name.split('.', 1)
    if not db or not name or any(c in db + name for c in ('\x00', '\n', '\r')):
        raise MigrationError('Invalid database/table name')
    if name.endswith('__repair'):
        raise MigrationError('Do not select staging names as source table')
    DB, NAME = db, name
    STAGE = name + '__repair'
    SRC = DST = (DB, NAME)
    STG = (DB, STAGE)
    PART_COL = None


def default_state_path():
    token = hashlib.sha256(f'{DB}.{NAME}'.encode('utf-8')).hexdigest()[:16]
    return f'data_transfer_progress/daily_stage_{token}.json'


def table_sql(tbl):
    return f'{qi(tbl[0])}.{qi(tbl[1])}'


def metadata(client, tbl):
    r = client.rows('SELECT engine, partition_key, sorting_key, primary_key, sampling_key '
                    'FROM system.tables '
                    f'WHERE database={qs(tbl[0])} AND name={qs(tbl[1])}')
    if len(r) > 1:
        raise MigrationError(f'{client.label}: ambiguous table {tbl}')
    return r[0] if r else None


def describe(client, tbl):
    return client.rows(f'DESCRIBE TABLE {table_sql(tbl)}')


def daily_partition_column(expression):
    """Only simple toYYYYMMDD(column) is supported, with quoted Unicode columns.

    Do not guess a date field from the table name: mismatched partition and
    filter columns can silently duplicate or omit rows in a stage.
    """
    compact = re.sub(r'[\s`]', '', str(expression))
    match = re.fullmatch(r'toYYYYMMDD\(([^(),]+)\)', compact)
    if not match:
        raise MigrationError(f'Unexpected source partition key: {expression}')
    return match.group(1)


def validate_base(source, target):
    global PART_COL
    sm, tm = metadata(source, SRC), metadata(target, DST)
    if not sm or not tm:
        raise MigrationError('Source/target table not found')
    if sm['engine'] != 'MergeTree' or tm['engine'] != 'ReplicatedMergeTree':
        raise MigrationError(f'Wrong source/target engine: {sm["engine"]}, {tm["engine"]}')
    # Day-partition validation guards the monthly batching assumption.
    partition_column = daily_partition_column(sm['partition_key'])
    for k in ('partition_key', 'sorting_key', 'primary_key', 'sampling_key'):
        if sm.get(k, '') != tm.get(k, ''):
            raise MigrationError(f'Source/target {k} mismatch')
    if describe(source, SRC) != describe(target, DST):
        raise MigrationError('Source/target DESCRIBE differs')
    period = [c for c in describe(source, SRC) if c['name'] == partition_column]
    if len(period) != 1 or not re.fullmatch(r'(Date|Date32|DateTime|DateTime64\(\d+(?:,\s*[^)]+)?\))', period[0]['type']):
        raise MigrationError(f'Unsupported date type for partition column {partition_column}: {period}')
    sz = source.scalar_text('SELECT timezone()')
    tz = target.scalar_text('SELECT timezone()')
    if sz != tz:
        raise MigrationError(f'Source/target server timezones differ ({sz} vs {tz}); cannot safely partition by day')
    PART_COL = partition_column


def check_stage(source, target, cluster, expected):
    tm, st = metadata(target, DST), metadata(target, STG)
    if not st:
        raise MigrationError('Stage does not exist; run --prepare first')
    if st['engine'] != 'ReplicatedMergeTree':
        raise MigrationError(f'Stage has unexpected engine {st["engine"]}')
    if not tm:
        raise MigrationError('Original target table missing')
    for k in ('partition_key', 'sorting_key', 'primary_key', 'sampling_key'):
        if st.get(k, '') != tm.get(k, ''):
            raise MigrationError(f'Staging {k} differs from target')
    if describe(target, STG) != describe(target, DST):
        raise MigrationError('Stage/target column definitions differ')
    sql = ('SELECT `table`, zookeeper_path FROM system.replicas '
           f'WHERE database={qs(DB)} AND `table` IN ({qs(NAME)}, {qs(STAGE)})')
    paths = {r['table']: r['zookeeper_path'] for r in target.rows(sql)}
    if not paths.get(NAME) or not paths.get(STAGE) or paths[NAME] == paths[STAGE]:
        raise MigrationError('Cannot prove staging Keeper path is separate from original')
    # The clone must be independently replicated to both hosts.
    return require_replicas(replica_state(target, cluster, STG), expected)


def replica_state(target, cluster, tbl):
    return target.rows(
        'SELECT hostName() AS host, is_readonly, is_session_expired, '
        'total_replicas, active_replicas, queue_size '
        f'FROM clusterAllReplicas({qs(cluster)}, system.replicas) '
        f'WHERE database={qs(tbl[0])} AND `table`={qs(tbl[1])} ORDER BY host'
    )


def month_query(table, month=None, local_host=False):
    if PART_COL is None:
        raise MigrationError('Partition column not validated; run validate_base first')
    date_col = qi(PART_COL)
    host = 'hostName() AS host, ' if local_host else ''
    where = f' WHERE toYYYYMM({date_col})={int(month)}' if month is not None else ''
    cols = (f'toYYYYMM({date_col}) AS month, ' if month is None else '')
    query = (f'SELECT {host}{cols}count() AS rows, '
             'toString(sum(cityHash64(*))) AS checksum_sum, '
             'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
             f'FROM {table_sql(table)}{where}')
    if month is None:
        query += ' GROUP BY month'
    return query


def month_summary(client, table):
    return {int(r['month']): normalize(r) for r in client.rows(month_query(table))}


def one_month(client, tbl, month):
    rows = client.rows(month_query(tbl, month))
    if len(rows) != 1:
        raise MigrationError(f'Expected exactly one monthly fingerprint for {tbl} {month}')
    return normalize(rows[0])


def replica_months(target, cluster, tbl, month, hosts):
    inner = month_query(tbl, month, local_host=True)
    records = target.rows(f'SELECT host, rows, checksum_sum, checksum_xor '
                          f'FROM clusterAllReplicas({qs(cluster)}, view({inner})) ORDER BY host')
    found = {r['host']: normalize(r) for r in records}
    if set(found) != hosts or len(records) != len(hosts):
        raise MigrationError(f'Missing or duplicate replica monthly fingerprints: {records}')
    return found


def full_fingerprint(client, tbl):
    data = client.rows('SELECT count() AS rows, '
                       'toString(sum(cityHash64(*))) AS checksum_sum, '
                       'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
                       f'FROM {table_sql(tbl)}')
    if len(data) != 1:
        raise MigrationError('Cannot read full table fingerprint')
    return normalize(data[0])


def full_replica_fingerprints(target, cluster, tbl, hosts):
    inner = ('SELECT hostName() AS host, count() AS rows, '
             'toString(sum(cityHash64(*))) AS checksum_sum, '
             'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
             f'FROM {table_sql(tbl)}')
    records = target.rows(f'SELECT host, rows, checksum_sum, checksum_xor '
                          f'FROM clusterAllReplicas({qs(cluster)}, view({inner})) ORDER BY host')
    found = {r['host']: normalize(r) for r in records}
    if len(records) != len(hosts) or set(found) != hosts:
        raise MigrationError('Cannot fetch full fingerprints from all replicas')
    return found


def verify_all(source, target, cluster, hosts, src_summary):
    now = month_summary(source, SRC)
    if now != src_summary:
        raise MigrationError('Source changed during stage transfer')
    staged_summary = month_summary(target, STG)
    if now != staged_summary:
        remaining = sorted(set(now) ^ set(staged_summary) |
                           {m for m in set(now) & set(staged_summary) if now[m] != staged_summary[m]})
        raise MigrationError(f'Stage incomplete or inconsistent months: {remaining}')
    source_full = full_fingerprint(source, SRC)
    staged_full = full_fingerprint(target, STG)
    if source_full != staged_full:
        raise MigrationError(f'Full source/stage fingerprints differ {source_full} vs {staged_full}')
    replicas = full_replica_fingerprints(target, cluster, STG, hosts)
    if any(r != source_full for r in replicas.values()):
        raise MigrationError(f'Stage replicas differ from source: {replicas}')
    require_replicas(replica_state(target, cluster, STG), len(hosts))
    return source_full, replicas


def main():
    p = argparse.ArgumentParser(description='Build verified staging copy of MergeTree PARTITION BY toYYYYMMDD(date_column) table')
    p.add_argument('--config', default='config.json')
    p.add_argument('--table', required=True, help='exact database.table; must have simple daily PARTITION BY toYYYYMMDD(Дата/Период/etc)')
    group = p.add_mutually_exclusive_group(required=True)
    group.add_argument('--inspect', action='store_true', help='read-only state, does not create staging')
    group.add_argument('--prepare', action='store_true', help='create stage and copy months, never touches original target')
    group.add_argument('--verify', action='store_true', help='read-only exhaustive verification of complete staging copy')
    p.add_argument('--ack-source-frozen', action='store_true', help='assert no writes to source or target')
    p.add_argument('--limit-months', type=int, help='pilot: stage at most this many new months')
    p.add_argument('--wait-seconds', type=int, default=300)
    p.add_argument('--state', default=None, help='optional journal path; default isolated per table')
    args = p.parse_args()
    select_table(args.table)
    if args.state is None:
        args.state = default_state_path()
    if args.limit_months is not None and (args.limit_months < 1 or not args.prepare):
        raise MigrationError('--limit-months must be positive and used only with --prepare')
    if (args.prepare or args.verify) and not args.ack_source_frozen:
        raise MigrationError('--prepare and --verify require --ack-source-frozen')
    cfg_path = Path(args.config).expanduser().resolve()
    cfg = json.loads(cfg_path.read_text(encoding='utf-8'))
    source = Client(cfg['source'], cfg_path.parent, 'source')
    target = Client(cfg['target'], cfg_path.parent, 'target')
    secrets = [source.password, target.password]
    cluster = cfg.get('target_cluster', 'default')
    expected = int(cfg.get('expected_replicas', 2))
    if expected != 2:
        raise MigrationError('This validated procedure expects exactly two target replicas')
    try:
        validate_base(source, target)
        require_replicas(replica_state(target, cluster, DST), expected)
        src_summary = month_summary(source, SRC)
        existing = metadata(target, STG) is not None
        print(f'Partition: toYYYYMMDD({PART_COL}); batching using toYYYYMM({PART_COL})')
        print(f'Source: months={len(src_summary)}, physical_rows={sum(int(v[0]) for v in src_summary.values())}')
        print(f'Original target: physical_rows={full_fingerprint(target, DST)[0]} (UNCHANGED by this script)')
        print(f'Stage exists: {existing}; stage name: {DB}.{STAGE}')
        if args.inspect:
            if existing:
                hosts = check_stage(source, target, cluster, expected)
                print(f'Stage physical_rows={full_fingerprint(target, STG)[0]}, healthy_replicas={len(hosts)}')
                current = month_summary(target, STG)
                missing = sorted(set(src_summary) - set(current))
                mismatched = sorted(m for m in src_summary.keys() & current.keys() if src_summary[m] != current[m])
                print(f'Missing stage months={missing}; mismatched stage months={mismatched}')
            print('INSPECT ONLY: no changes')
            return 0
        if args.prepare and not existing:
            # ENGINE=Replicated database propagates CREATE. The source table is NOT modified.
            target.query(f'CREATE TABLE {table_sql(STG)} AS {table_sql(DST)}')
            print('Created independent staging table; original target untouched.', flush=True)
            deadline = time.monotonic() + args.wait_seconds
            while True:
                try:
                    hosts = check_stage(source, target, cluster, expected)
                    break
                except MigrationError:
                    if time.monotonic() >= deadline:
                        raise
                    time.sleep(3)
        else:
            hosts = check_stage(source, target, cluster, expected)
        existing_state_path = Path(args.state).expanduser().resolve()
        state = (json.loads(existing_state_path.read_text(encoding='utf-8'))
                 if existing_state_path.is_file() else {})
        if state and (state.get('source') != source.host or state.get('target') != target.host or state.get('stage') != STAGE or state.get('table') != args.table):
            raise MigrationError('Staging journal belongs to another cluster or stage')
        state = state or {'source': source.host, 'target': target.host, 'table': args.table, 'stage': STAGE, 'months': {}}
        stage_summary = month_summary(target, STG)
        if set(stage_summary) - set(src_summary):
            raise MigrationError('Stage contains months absent from source: ' + str(sorted(set(stage_summary) - set(src_summary))))
        empty, done, conflicts = [], [], []
        for month in sorted(src_summary):
            before = one_month(source, SRC, month)
            if before != src_summary[month]:
                raise MigrationError(f'Source changed before scan month {month}')
            stage_fps = replica_months(target, cluster, STG, month, hosts)
            state_for_month = classify(before, stage_fps)
            print(f'{month}: {state_for_month:8} source={before[0]}, stage={list(x[0] for x in stage_fps.values())}', flush=True)
            if state_for_month == 'EMPTY': empty.append(month)
            elif state_for_month == 'COMPLETE': done.append(month)
            else: conflicts.append(month)
        if conflicts:
            raise MigrationError('Stage has partially loaded or inconsistent months: ' + str(conflicts) + '. No inserts made in this run.')
        if args.verify:
            if empty:
                raise MigrationError(f'Staging is incomplete: {len(empty)} source months not yet copied')
            full, repl = verify_all(source, target, cluster, hosts, src_summary)
            print(f'STAGING VERIFIED: rows={full[0]}, fingerprints={full[1:]}')
            for host in sorted(repl):
                print(f'  replica {host}: MATCH')
            print('NO EXCHANGE, DROP, TRUNCATE, RENAME or changes to original target occurred.')
            return 0
        selected = empty[:args.limit_months] if args.limit_months else empty
        origin = f'{source.host}:{int(cfg.get("source_native_tls_port", 9440))}'
        for n, month in enumerate(selected, 1):
            prior = state['months'].get(str(month), {}).get('status')
            if prior in ('INSERTING', 'UNCERTAIN'):
                raise MigrationError(f'Stage month {month} has {prior} in journal; review before a retry')
            require_replicas(replica_state(target, cluster, STG), expected)
            src_fp = one_month(source, SRC, month)
            if src_fp != src_summary[month]:
                raise MigrationError(f'Source changed before INSERT month {month}')
            if classify(src_fp, replica_months(target, cluster, STG, month, hosts)) != 'EMPTY':
                raise MigrationError(f'Stage month {month} not empty just before INSERT')
            state['months'][str(month)] = {'status': 'INSERTING', 'started_at': utcnow(), 'source': list(src_fp)}
            save_state(existing_state_path, state)
            sql = (f'INSERT INTO {table_sql(STG)} SELECT * FROM '
                   f'remoteSecure({qs(origin)}, {qs(DB)}, {qs(NAME)}, '
                   f'{qs(source.username)}, {qs(source.password)}) '
                   f'WHERE toYYYYMM({qi(PART_COL)})={month}')
            print(f'[{n}/{len(selected)}] COPYING {month}, rows={src_fp[0]} into STAGE only', flush=True)
            try:
                target.query(sql)
                deadline = time.monotonic() + args.wait_seconds
                while True:
                    if one_month(source, SRC, month) != src_fp:
                        raise MigrationError(f'Source changed during month {month}')
                    current = replica_months(target, cluster, STG, month, hosts)
                    health = replica_state(target, cluster, STG)
                    if classify(src_fp, current) == 'COMPLETE' and all(
                        int(x['queue_size']) == 0 and int(x['active_replicas']) == expected and int(x['is_readonly']) == 0
                        for x in health) and len(health) == expected:
                        break
                    if time.monotonic() > deadline:
                        raise MigrationError(f'Stage replica verification timeout for {month}')
                    time.sleep(5)
                state['months'][str(month)] = {'status': 'DONE', 'finished_at': utcnow(), 'source': list(src_fp)}
                save_state(existing_state_path, state)
                print(f'[{n}/{len(selected)}] DONE {month} on both staging replicas', flush=True)
            except Exception as exc:
                state['months'][str(month)] = {'status': 'UNCERTAIN', 'error': redact(str(exc), secrets)[:1000], 'at': utcnow()}
                save_state(existing_state_path, state)
                raise MigrationError(f'Stage month {month} uncertain; NO AUTOMATIC RETRY: {redact(str(exc), secrets)[:600]}') from None
        if month_summary(source, SRC) != src_summary:
            raise MigrationError('Source changed during stage loading')
        pending = len(empty) - len(selected)
        print(f'BATCH FINISHED: loaded_new_months={len(selected)}, prior_complete={len(done)}, pending={pending}')
        if pending == 0:
            full, replicas = verify_all(source, target, cluster, hosts, src_summary)
            print(f'STAGING VERIFIED: rows={full[0]}, fingerprints={full[1:]}')
            print('No EXCHANGE / RENAME / DROP / TRUNCATE was executed. Run --verify again before any manual cutover.')
        return 0
    except MigrationError as exc:
        raise MigrationError(redact(str(exc), secrets)) from None


if __name__ == '__main__':
    try:
        sys.exit(main())
    except (MigrationError, OSError, ValueError, KeyError, json.JSONDecodeError) as exc:
        print('ERROR:', str(exc), file=sys.stderr)
        sys.exit(1)
