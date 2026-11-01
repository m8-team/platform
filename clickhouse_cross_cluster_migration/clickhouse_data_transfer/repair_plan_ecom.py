#!/usr/bin/env python3
"""Safely prepare a staging copy of zendendb.ПланПродаж_ecom from SOURCE FINAL.

The stage table uses the same target DDL, only a different name.
No source mutations. Never drops/truncates/exchanges target tables. Python 3.9+.
Requires transfer_data.py v5 in the same directory.
"""
from __future__ import annotations

import argparse
import json
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

from transfer_data import (Client, MigrationError, ident, lit as sql_lit,
                           get_metadata, get_columns, describe_columns,
                           row_fingerprint, fingerprint_equal,
                           verify_fingerprint_on_replicas, rowcount,
                           is_unversioned_replacing, verified_replicas,
                           replica_counts)

DB = 'zendendb'
TABLE = 'ПланПродаж_ecom'
STAGE = 'ПланПродаж_ecom__repair'


def execute_binary(client: Client, sql: str, data: bytes = b'') -> bytes:
    """Run SQL via verified HTTPS, permitting Native binary response/request bodies.

    Authentication uses HTTP headers, not SQL text or query parameters.
    The caller must restrict SQL to the vetted known table identifiers.
    """
    headers = {
        'X-ClickHouse-User': client.username,
        'X-ClickHouse-Key': client.password,
        'Content-Type': 'application/octet-stream',
    }
    url = client.url + '?' + urllib.parse.urlencode({'wait_end_of_query': 1})
    req = urllib.request.Request(url, data=sql.encode('utf-8') + data,
                                 headers=headers, method='POST')
    try:
        with urllib.request.urlopen(req, timeout=client.timeout, context=client.context) as response:
            return response.read()
    except urllib.error.HTTPError as exc:
        message = exc.read(2000).decode('utf-8', errors='replace')
        for secret in (client.password,):
            if secret:
                message = message.replace(secret, '[REDACTED]')
        raise MigrationError(f'{client.name}: HTTP {exc.code}: {message}') from None
    except (urllib.error.URLError, TimeoutError, OSError) as exc:
        raise MigrationError(f'{client.name}: connection error ({type(exc).__name__})') from None


def meta_check(source, target, stage_exists=False):
    src_meta = get_metadata(source, [DB])
    dst_meta = get_metadata(target, [DB])
    src = src_meta.get((DB, TABLE))
    dst = dst_meta.get((DB, TABLE))
    if not src or not dst:
        raise MigrationError('Source or current target table does not exist')
    if src['engine'] != 'ReplacingMergeTree' or dst['engine'] != 'ReplicatedReplacingMergeTree':
        raise MigrationError('Unexpected source/target engine, expected ReplacingMergeTree / ReplicatedReplacingMergeTree')
    if not is_unversioned_replacing(src):
        raise MigrationError('Source ReplacingMergeTree has explicit version/parameters; this tool requires unversioned ReplacingMergeTree')
    for field in ('partition_key', 'sorting_key', 'primary_key', 'sampling_key'):
        if (src.get(field) or '') != (dst.get(field) or ''):
            raise MigrationError(f'Source and target {field} differ')
    source_cols = get_columns(source, [DB])
    target_cols = get_columns(target, [DB])
    if source_cols.get((DB, TABLE)) != target_cols.get((DB, TABLE)):
        raise MigrationError('Source and current target column definitions differ')
    if describe_columns(source, (DB, TABLE)) != describe_columns(target, (DB, TABLE)):
        raise MigrationError('Source and current target DESCRIBE columns differ')
    if stage_exists:
        stg = dst_meta.get((DB, STAGE))
        if not stg or stg['engine'] != dst['engine']:
            raise MigrationError('Staging engine missing or different')
        for field in ('partition_key', 'sorting_key', 'primary_key', 'sampling_key'):
            if (stg.get(field) or '') != (dst.get(field) or ''):
                raise MigrationError(f'Staging {field} differs from target')
        if target_cols.get((DB, STAGE)) != target_cols.get((DB, TABLE)):
            raise MigrationError('Staging columns differ from target')
        if describe_columns(target, (DB, STAGE)) != describe_columns(target, (DB, TABLE)):
            raise MigrationError('Staging DESCRIBE columns differ from target')
        keeper_rows = target.rows(
            'SELECT `table`, zookeeper_path FROM system.replicas '
            f'WHERE database = {sql_lit(DB)} AND `table` IN ({sql_lit(TABLE)}, {sql_lit(STAGE)})'
        )
        by_table = {r['table']: r['zookeeper_path'] for r in keeper_rows}
        if not all(by_table.get(t) for t in (TABLE, STAGE)):
            raise MigrationError('Cannot verify distinct Keeper paths for target and stage; refusing INSERT')
        if by_table[TABLE] == by_table[STAGE]:
            raise MigrationError('Staging table SHARES Keeper path with original: refusing INSERT')
    return (DB, STAGE) in dst_meta


def run():
    parser = argparse.ArgumentParser(description='Repair a single unversioned ReplacingMergeTree via a staged Native copy of source FINAL')
    parser.add_argument('--config', default='config.json')
    action = parser.add_mutually_exclusive_group(required=True)
    action.add_argument('--inspect', action='store_true', help='no changes: show source and target counts, stage state')
    action.add_argument('--prepare', action='store_true', help='create and populate stage table; never replace original')
    action.add_argument('--verify', action='store_true', help='verify prepared stage versus live source, on both replicas')
    parser.add_argument('--ack-source-frozen', action='store_true', help='confirm source writes are stopped during prepare/verify')
    parser.add_argument('--replica-wait-seconds', type=int, default=300)
    args = parser.parse_args()
    if (args.prepare or args.verify) and not args.ack_source_frozen:
        parser.error('--prepare and --verify require --ack-source-frozen')
    p = Path(args.config).expanduser().resolve()
    cfg = json.loads(p.read_text(encoding='utf-8'))
    source = Client(cfg['source'], p.parent, 'source')
    target = Client(cfg['target'], p.parent, 'target')
    cluster = cfg.get('target_cluster', 'default')
    replicas = int(cfg.get('expected_replicas', 2))
    if replicas != 2:
        raise MigrationError('Expected exactly two target replicas for this prepared procedure')

    target_meta = get_metadata(target, [DB])
    stage_exists = (DB, STAGE) in target_meta
    meta_check(source, target, stage_exists=stage_exists)
    src_key, orig_key, stage_key = (DB, TABLE), (DB, TABLE), (DB, STAGE)
    print('Source physical:', rowcount(source, src_key))
    print('Source FINAL:', row_fingerprint(source, src_key, True))
    print('Current target physical:', rowcount(target, orig_key))
    print('Current target FINAL:', row_fingerprint(target, orig_key, True))
    print('Stage exists:', stage_exists)
    if args.inspect:
        if stage_exists:
            print('Stage physical:', rowcount(target, stage_key))
            print('Stage FINAL:', row_fingerprint(target, stage_key, True))
        return
    if args.prepare:
        src_fp_before = row_fingerprint(source, src_key, True)
        if not stage_exists:
            # In a Replicated database, metadata will automatically propagate to both hosts.
            target.query(f'CREATE TABLE {ident(DB)}.{ident(STAGE)} AS {ident(DB)}.{ident(TABLE)}')
            print('Created staged table; original target untouched.')
        meta_check(source, target, stage_exists=True)
        # Both replicas must be healthy and the staging table empty on BOTH
        # before writing any data; never blindly retry a partially populated stage.
        verified_replicas(target, cluster, stage_key, replicas, None, args.replica_wait_seconds)
        per_replica_counts = replica_counts(target, cluster, stage_key)
        if len(per_replica_counts) != replicas:
            raise MigrationError('Cannot confirm stage row counts on both replicas; refusing INSERT')
        stage_rows = rowcount(target, stage_key)
        if any(int(r['rows']) > 0 for r in per_replica_counts):
            if stage_rows == 0:
                raise MigrationError('Staging is empty locally but contains rows on another replica; refusing INSERT')
            print('Stage already has rows: no INSERT will be retried.')
        elif stage_rows != 0:
            raise MigrationError('Stage local count differs from replica counts; refusing INSERT')
        else:
            # Native preserves DateTime, LowCardinality, Decimals, NULLs precisely.
            native = execute_binary(source, f'SELECT * FROM {ident(DB)}.{ident(TABLE)} FINAL FORMAT Native')
            if not native:
                raise MigrationError('Source FINAL returned empty Native stream unexpectedly')
            print('Reading source FINAL (Native bytes):', len(native))
            execute_binary(target, f'INSERT INTO {ident(DB)}.{ident(STAGE)} FORMAT Native\n', native)
            print('Inserted into staged table only; original target untouched.')
        src_fp_after = row_fingerprint(source, src_key, True)
        if not fingerprint_equal(src_fp_before, src_fp_after):
            raise MigrationError('Source FINAL changed during preparation; stage must be reconciled before switch')
    if not (get_metadata(target, [DB]).get(stage_key)):
        raise MigrationError('Staging table does not exist; run --prepare first')
    if args.verify or args.prepare:
        src_fp = row_fingerprint(source, src_key, True)
        staged_fp = row_fingerprint(target, stage_key, True)
        if not fingerprint_equal(src_fp, staged_fp):
            raise MigrationError(f'Stage FINAL mismatches source FINAL. source={src_fp}, stage={staged_fp}; NO CUTOVER')
        # Verify each replica independently via aggregate query pushed to the local host.
        from transfer_data import replica_fingerprints
        verified_replicas(target, cluster, stage_key, replicas, None, args.replica_wait_seconds)
        per_replica = replica_fingerprints(target, cluster, stage_key, True)
        if len(per_replica) != replicas or any(not fingerprint_equal(src_fp, row) for row in per_replica):
            raise MigrationError(f'Staging replica fingerprints do not match source: {per_replica}')
        print('STAGING VERIFIED:', json.dumps(src_fp, ensure_ascii=False))
        for row in per_replica:
            print('Replica:', row['host'], 'logical_rows:', row['rows'], 'MATCH')
        print('Safe to PLAN a cutover. NO exchange, rename, drop or truncate performed.')
        print('Use EXCHANGE TABLES only during a scheduled window, after a fresh --verify.')


if __name__ == '__main__':
    try:
        run()
    except (MigrationError, OSError, KeyError, ValueError) as error:
        print('ERROR:', error, file=sys.stderr)
        sys.exit(1)
