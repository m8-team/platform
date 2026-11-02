#!/usr/bin/env python3
"""Safe staging migration for zendendb.Товары, partitioned by a String column.

Source: MergeTree PARTITION BY `ТипНоменклатуры`.
Target: ReplicatedMergeTree in Replicated DB, two replicas.

Writes ONLY to independently replicated zendendb.Товары__repair.
Does not modify target main table, source, or main migration journal.
"""
from __future__ import annotations

import argparse
import json
import re
import sys
import time
from pathlib import Path

from transfer_monthly import (Client, MigrationError, classify, normalize, qi, qs,
                              require_replicas, save_state, utcnow, redact)

DB = 'zendendb'
TABLE = 'Товары'
STAGE = TABLE + '__repair'
PART_KEY = 'ТипНоменклатуры'
SRC = (DB, TABLE)
DST = (DB, TABLE)
STG = (DB, STAGE)
HEX_RE = re.compile(r'^(?:[0-9A-Fa-f]{2})*$')


def table_sql(tbl):
    return f'{qi(tbl[0])}.{qi(tbl[1])}'


def hex_key(s):
    if not isinstance(s, str) or not HEX_RE.fullmatch(s):
        raise MigrationError('Invalid hex partition key returned by ClickHouse')
    return s.upper()


def key_sql(pk_hex):
    return f'unhex({qs(hex_key(pk_hex))})'


def safe_label(pk_hex):
    # Do not show raw business values or accidentally print control characters.
    key = hex_key(pk_hex)
    return '<empty-string>' if not key else f'hex:{key[:20]}{("..." if len(key) > 20 else "")}'


def metadata(client, table):
    rows = client.rows('SELECT engine, partition_key, sorting_key, primary_key, sampling_key '
                       'FROM system.tables '
                       f'WHERE database={qs(table[0])} AND name={qs(table[1])}')
    if len(rows) > 1:
        raise MigrationError(f'{client.label}: duplicate system.tables rows for {table}')
    return rows[0] if rows else None


def describe(client, table):
    return client.rows(f'DESCRIBE TABLE {table_sql(table)}')


def valid_partition(expr):
    return re.sub(r'[\s`]', '', expr) == PART_KEY


def validate_base(source, target):
    src, dst = metadata(source, SRC), metadata(target, DST)
    if src is None or dst is None:
        raise MigrationError('Source or target base table missing')
    if src['engine'] != 'MergeTree' or dst['engine'] != 'ReplicatedMergeTree':
        raise MigrationError(f'Unexpected source/target engine: {src["engine"]}, {dst["engine"]}')
    if not valid_partition(src['partition_key']):
        raise MigrationError('Unexpected partition expression in source: ' + src['partition_key'])
    for k in ('partition_key', 'sorting_key', 'primary_key', 'sampling_key'):
        if src[k] != dst[k]:
            raise MigrationError(f'Source/target {k} mismatch: {src[k]} vs {dst[k]}')
    sd, td = describe(source, SRC), describe(target, DST)
    if sd != td:
        raise MigrationError('Source and target DESCRIBE TABLE differ')
    cols = [r for r in sd if r['name'] == PART_KEY]
    if len(cols) != 1 or cols[0]['type'] != 'String':
        raise MigrationError(f'Expected {PART_KEY} with exactly String type')


def replica_state(target, cluster, table):
    return target.rows(
        'SELECT hostName() AS host, is_readonly, is_session_expired, '
        'total_replicas, active_replicas, queue_size '
        f'FROM clusterAllReplicas({qs(cluster)}, system.replicas) '
        f'WHERE database={qs(table[0])} AND `table`={qs(table[1])} ORDER BY host')


def check_stage(target, cluster, expected):
    dst, stage = metadata(target, DST), metadata(target, STG)
    if not stage:
        raise MigrationError('Staging table does not exist; run --prepare')
    if stage['engine'] != 'ReplicatedMergeTree':
        raise MigrationError(f'Unexpected staging engine {stage["engine"]}')
    for k in ('partition_key', 'sorting_key', 'primary_key', 'sampling_key'):
        if stage[k] != dst[k]:
            raise MigrationError(f'Staging {k} differs from target')
    if describe(target, STG) != describe(target, DST):
        raise MigrationError('Staging columns differ from target columns')
    paths = target.rows('SELECT `table`, zookeeper_path FROM system.replicas '
                        f'WHERE database={qs(DB)} AND `table` IN ({qs(TABLE)}, {qs(STAGE)})')
    found = {r['table']: r['zookeeper_path'] for r in paths}
    if not found.get(TABLE) or not found.get(STAGE) or found[TABLE] == found[STAGE]:
        raise MigrationError('Cannot prove staging uses a different Keeper path')
    return require_replicas(replica_state(target, cluster, STG), expected)


def partition_summary(client, table):
    sql = (f'SELECT hex({qi(PART_KEY)}) AS pk_hex, count() AS rows, '
           'toString(sum(cityHash64(*))) AS checksum_sum, '
           'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
           f'FROM {table_sql(table)} GROUP BY pk_hex ORDER BY pk_hex')
    rows = client.rows(sql)
    result = {}
    for r in rows:
        pk = hex_key(r['pk_hex'])
        if pk in result:
            raise MigrationError('Duplicate partition group in results')
        result[pk] = normalize(r)
    return result


def one_partition_sql(table, pk_hex, host=False):
    host_expr = 'hostName() AS host, ' if host else ''
    return (f'SELECT {host_expr}count() AS rows, '
            'toString(sum(cityHash64(*))) AS checksum_sum, '
            'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
            f'FROM {table_sql(table)} '
            f'WHERE {qi(PART_KEY)} = {key_sql(pk_hex)}')


def partition_fp(client, table, pk_hex):
    rows = client.rows(one_partition_sql(table, pk_hex))
    if len(rows) != 1:
        raise MigrationError('Cannot read per-partition fingerprint ' + safe_label(pk_hex))
    return normalize(rows[0])


def replica_partition_fps(target, cluster, table, pk_hex, expected_hosts):
    inner = one_partition_sql(table, pk_hex, host=True)
    rows = target.rows('SELECT host, rows, checksum_sum, checksum_xor '
                       f'FROM clusterAllReplicas({qs(cluster)}, view({inner})) ORDER BY host')
    found = {str(r['host']): normalize(r) for r in rows}
    if len(rows) != len(expected_hosts) or set(found) != expected_hosts:
        raise MigrationError(f'Expected 2 explicit replica results, found {len(rows)}: {safe_label(pk_hex)}')
    return found


def full_fp(client, table):
    rows = client.rows('SELECT count() AS rows, '
                       'toString(sum(cityHash64(*))) AS checksum_sum, '
                       'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
                       f'FROM {table_sql(table)}')
    if len(rows) != 1:
        raise MigrationError('Cannot read full table fingerprint')
    return normalize(rows[0])


def full_replicas(target, cluster, table, expected_hosts):
    inner = ('SELECT hostName() AS host, count() AS rows, '
             'toString(sum(cityHash64(*))) AS checksum_sum, '
             'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
             f'FROM {table_sql(table)}')
    rows = target.rows('SELECT host, rows, checksum_sum, checksum_xor '
                       f'FROM clusterAllReplicas({qs(cluster)}, view({inner})) ORDER BY host')
    found = {str(r['host']): normalize(r) for r in rows}
    if len(rows) != len(expected_hosts) or set(found) != expected_hosts:
        raise MigrationError('Cannot read full fingerprints from both replicas')
    return found


def verify_complete(source, target, cluster, hosts, initial):
    now = partition_summary(source, SRC)
    if now != initial:
        raise MigrationError('Source partition fingerprints changed during migration')
    staged = partition_summary(target, STG)
    if staged != now:
        diffs = sorted(set(now) ^ set(staged) |
                       {k for k in (set(now) & set(staged)) if now[k] != staged[k]})
        raise MigrationError(f'Staging partition fingerprints differ for {len(diffs)} keys')
    source_full = full_fp(source, SRC)
    target_full = full_fp(target, STG)
    if source_full != target_full:
        raise MigrationError('Source and staging full fingerprints differ')
    replicated = full_replicas(target, cluster, STG, hosts)
    if any(fp != source_full for fp in replicated.values()):
        raise MigrationError('Staging fingerprint differs on one or more replicas')
    require_replicas(replica_state(target, cluster, STG), len(hosts))
    return source_full, replicated


def main():
    p = argparse.ArgumentParser(description='Staging copy of zendendb.Товары partitioned by a String value')
    p.add_argument('--config', default='config.json')
    mode = p.add_mutually_exclusive_group(required=True)
    mode.add_argument('--inspect', action='store_true', help='read-only summary')
    mode.add_argument('--prepare', action='store_true', help='create independent stage and copy partitions')
    mode.add_argument('--verify', action='store_true', help='read-only full comparison of completed stage')
    p.add_argument('--ack-source-frozen', action='store_true')
    p.add_argument('--limit-partitions', type=int, help='pilot: stage at most N previously empty partitions')
    p.add_argument('--wait-seconds', type=int, default=300)
    p.add_argument('--state', default='data_transfer_progress/products_stage_progress.json')
    args = p.parse_args()
    if args.limit_partitions is not None and (args.limit_partitions < 1 or not args.prepare):
        raise MigrationError('--limit-partitions must be positive and used with --prepare')
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
        raise MigrationError('This procedure supports exactly two replicas')
    try:
        validate_base(source, target)
        require_replicas(replica_state(target, cluster, DST), expected)
        src_summary = partition_summary(source, SRC)
        stage_exists = metadata(target, STG) is not None
        print(f'Table: {DB}.{TABLE}; source_partitions={len(src_summary)}, '
              f'source_rows={sum(int(fp[0]) for fp in src_summary.values())}', flush=True)
        print(f'Current target rows={full_fp(target, DST)[0]}; original target UNCHANGED')
        print(f'Staging table: {DB}.{STAGE}; exists={stage_exists}')
        if args.inspect:
            if stage_exists:
                hosts = check_stage(target, cluster, expected)
                stage_summary = partition_summary(target, STG)
                mismatch = sum(1 for k, fp in stage_summary.items() if src_summary.get(k) != fp)
                print(f'Stage rows={full_fp(target, STG)[0]}, '
                      f'complete={sum(stage_summary.get(k) == fp for k, fp in src_summary.items())}, '
                      f'missing={len(src_summary.keys() - stage_summary.keys())}, '
                      f'mismatched={mismatch}, replicas={len(hosts)}')
            print('INSPECT ONLY: no writes')
            return 0
        if args.prepare and not stage_exists:
            target.query(f'CREATE TABLE {table_sql(STG)} AS {table_sql(DST)}')
            print('Created separately replicated staging table; original target unchanged.', flush=True)
            deadline = time.monotonic() + args.wait_seconds
            while True:
                try:
                    hosts = check_stage(target, cluster, expected)
                    break
                except MigrationError:
                    if time.monotonic() >= deadline:
                        raise
                    time.sleep(3)
        else:
            hosts = check_stage(target, cluster, expected)
        state_path = Path(args.state).expanduser().resolve()
        state = json.loads(state_path.read_text(encoding='utf-8')) if state_path.is_file() else {}
        if state and (state.get('source') != source.host or state.get('target') != target.host or
                      state.get('stage') != STAGE or state.get('table') != TABLE):
            raise MigrationError('Existing stage journal belongs to a different table or cluster')
        state = state or {'source': source.host, 'target': target.host,
                          'table': TABLE, 'stage': STAGE, 'partitions': {}}
        stage_summary = partition_summary(target, STG)
        unexpected = set(stage_summary) - set(src_summary)
        if unexpected:
            raise MigrationError(f'Stage has {len(unexpected)} unexpected partitions; no INSERT')
        ready, done, conflicts = [], [], []
        for pk in sorted(src_summary):
            state_now = state['partitions'].get(pk, {}).get('status')
            if state_now in ('INSERTING', 'UNCERTAIN'):
                raise MigrationError(f'Prior write {state_now} for {safe_label(pk)}; manual review required')
            if stage_summary.get(pk) == src_summary[pk]:
                # The local host may be caught up while the second replica is not.
                fps = replica_partition_fps(target, cluster, STG, pk, hosts)
                if classify(src_summary[pk], fps) == 'COMPLETE':
                    done.append(pk)
                else:
                    conflicts.append(pk)
                continue
            if pk in stage_summary:
                conflicts.append(pk)
                continue
            # A grouped summary on the local target alone is not proof the other replica is empty.
            recs = replica_partition_fps(target, cluster, STG, pk, hosts)
            if classify(src_summary[pk], recs) == 'EMPTY':
                ready.append(pk)
            else:
                conflicts.append(pk)
        print(f'Stage status: complete={len(done)}, empty={len(ready)}, '
              f'conflicted={len(conflicts)}', flush=True)
        if conflicts:
            raise MigrationError(f'Staging contains {len(conflicts)} partial/inconsistent partitions, '
                                 f'first={safe_label(conflicts[0])}. No inserts performed.')
        if args.verify:
            if ready:
                raise MigrationError(f'Stage is incomplete: {len(ready)} partitions missing')
            fp, repl = verify_complete(source, target, cluster, hosts, src_summary)
            print(f'STAGING VERIFIED: {fp[0]} rows, hashes={fp[1:]}')
            for host in sorted(repl):
                print(f'  {host}: MATCH')
            print('NO exchange/rename/drop/truncate or main journal modification')
            return 0
        selected = ready[:args.limit_partitions] if args.limit_partitions else ready
        origin = f'{source.host}:{int(cfg.get("source_native_tls_port", 9440))}'
        for n, pk in enumerate(selected, 1):
            require_replicas(replica_state(target, cluster, STG), expected)
            src_fp = partition_fp(source, SRC, pk)
            if src_fp != src_summary[pk]:
                raise MigrationError('Source changed before transfer of ' + safe_label(pk))
            if classify(src_fp, replica_partition_fps(target, cluster, STG, pk, hosts)) != 'EMPTY':
                raise MigrationError('Stage not empty immediately before insert for ' + safe_label(pk))
            state['partitions'][pk] = {'status': 'INSERTING', 'started_at': utcnow(), 'source': list(src_fp)}
            save_state(state_path, state)
            sql = (f'INSERT INTO {table_sql(STG)} '
                   f'SELECT * FROM remoteSecure({qs(origin)}, {qs(DB)}, {qs(TABLE)}, '
                   f'{qs(source.username)}, {qs(source.password)}) '
                   f'WHERE {qi(PART_KEY)} = {key_sql(pk)}')
            print(f'[{n}/{len(selected)}] COPYING {safe_label(pk)} rows={src_fp[0]} STAGING ONLY', flush=True)
            try:
                target.query(sql)
                deadline = time.monotonic() + args.wait_seconds
                while True:
                    if partition_fp(source, SRC, pk) != src_fp:
                        raise MigrationError('Source changed during transfer of ' + safe_label(pk))
                    current = replica_partition_fps(target, cluster, STG, pk, hosts)
                    healthy = replica_state(target, cluster, STG)
                    if (classify(src_fp, current) == 'COMPLETE' and len(healthy) == expected and
                        all(int(x['queue_size']) == 0 and int(x['active_replicas']) == expected and
                            int(x['is_readonly']) == 0 and int(x['is_session_expired']) == 0
                            for x in healthy)):
                        break
                    if time.monotonic() > deadline:
                        raise MigrationError('Stage replica verification timeout for ' + safe_label(pk))
                    time.sleep(5)
                state['partitions'][pk] = {'status': 'DONE', 'finished_at': utcnow(), 'source': list(src_fp)}
                save_state(state_path, state)
                print(f'[{n}/{len(selected)}] DONE {safe_label(pk)} on both staging replicas', flush=True)
            except Exception as exc:
                state['partitions'][pk] = {'status': 'UNCERTAIN', 'error': redact(str(exc), secrets)[:1000], 'at': utcnow()}
                save_state(state_path, state)
                raise MigrationError(f'Partition {safe_label(pk)} UNCERTAIN; do not retry blindly: '
                                     + redact(str(exc), secrets)[:600]) from None
        if partition_summary(source, SRC) != src_summary:
            raise MigrationError('Source changed during staging process; do not switch tables')
        print(f'BATCH FINISHED: new={len(selected)}, prior_done={len(done)}, '
              f'pending={len(ready) - len(selected)}')
        if len(selected) == len(ready):
            fp, replicated = verify_complete(source, target, cluster, hosts, src_summary)
            print(f'STAGING VERIFIED: {fp[0]} rows, hashes={fp[1:]}')
            for host in sorted(replicated):
                print(f'  {host}: MATCH')
            print('NO EXCHANGE / DROP / TRUNCATE. Run --verify again before manual cutover.')
        return 0
    except MigrationError as exc:
        raise MigrationError(redact(str(exc), secrets)) from None


if __name__ == '__main__':
    try:
        sys.exit(main())
    except (MigrationError, OSError, KeyError, ValueError, json.JSONDecodeError) as exc:
        print('ERROR:', str(exc), file=sys.stderr)
        sys.exit(1)
