#!/usr/bin/env python3
"""Read-only per-key audit of one ClickHouse ReplacingMergeTree migration.

Uses the Client from the existing transfer_data.py and its config.json.
Only performs SELECTs against source and target. Never updates the progress journal.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from collections import Counter, defaultdict
from pathlib import Path

from transfer_data import Client, MigrationError, ident

DEFAULT_TABLE = 'zendendb.ПланПродаж_ecom'
DEFAULT_KEYS = ('Подразделение', 'Дата')


def canonical(row):
    return json.dumps(row, sort_keys=True, ensure_ascii=False,
                      separators=(',', ':'), allow_nan=False)


def signature(row):
    return hashlib.sha256(canonical(row).encode('utf-8')).hexdigest()


def build_index(rows, key_columns):
    idx = {}
    for row in rows:
        missing = [k for k in key_columns if k not in row]
        if missing:
            raise MigrationError(f'Expected key column(s) not present: {missing}')
        key = tuple(canonical(row[k]) for k in key_columns)
        if key in idx:
            raise MigrationError(f'FINAL returned duplicate ORDER BY key: {key}')
        idx[key] = row
    return idx


def raw_groups(rows, key_columns):
    groups = defaultdict(list)
    for row in rows:
        groups[tuple(canonical(row[k]) for k in key_columns)].append(signature(row))
    return groups


def compare(source_final, target_final, source_raw, target_raw, key_columns):
    a, b = build_index(source_final, key_columns), build_index(target_final, key_columns)
    ar, br = raw_groups(source_raw, key_columns), raw_groups(target_raw, key_columns)
    missing_target, missing_source, changed = [], [], []
    for key in sorted(set(a) | set(b)):
        display_key = {col: json.loads(val) for col, val in zip(key_columns, key)}
        if key not in b:
            missing_target.append(display_key)
        elif key not in a:
            missing_source.append(display_key)
        elif signature(a[key]) != signature(b[key]):
            sa, sb = signature(a[key]), signature(b[key])
            changed_cols = sorted(c for c in set(a[key]) | set(b[key]) if a[key].get(c) != b[key].get(c))
            changed.append({
                'key': display_key,
                'changed_columns': changed_cols,
                'source_final_sha256': sa,
                'target_final_sha256': sb,
                'source_raw_versions': len(ar.get(key, [])),
                'target_raw_versions': len(br.get(key, [])),
                'target_winner_exists_in_source_raw': sb in ar.get(key, []),
                'source_winner_exists_in_target_raw': sa in br.get(key, []),
            })
    return {
        'source': {
            'physical_rows': len(source_raw), 'logical_rows': len(source_final),
            'duplicate_keys': sum(n > 1 for n in map(len, ar.values())),
            'excess_physical_rows': len(source_raw) - len(ar),
        },
        'target': {
            'physical_rows': len(target_raw), 'logical_rows': len(target_final),
            'duplicate_keys': sum(n > 1 for n in map(len, br.values())),
            'excess_physical_rows': len(target_raw) - len(br),
        },
        'missing_on_target': missing_target,
        'missing_on_source': missing_source,
        'changed_keys': changed,
        'changed_keys_count': len(changed),
        'matching_keys_count': len(set(a) & set(b)) - len(changed),
        'matched': not missing_target and not missing_source and not changed,
    }


def read_rows(client, db, table, normalize_timestamp, final, max_rows):
    # In JSONEachRow, DateTime text depends on server time zone. Normalize to
    # Unix seconds without changing the underlying data or hash input column name.
    modifier = ' REPLACE (toUnixTimestamp(`timestamp`) AS `timestamp`)' if normalize_timestamp else ''
    final_sql = ' FINAL' if final else ''
    sql = f'SELECT *{modifier} FROM {ident(db)}.{ident(table)}{final_sql}'
    result = client.rows(sql)
    if len(result) > max_rows:
        raise MigrationError(f'{client.name}: {db}.{table} has >{max_rows} rows; refusing to hold them in memory')
    return result


def write_private_json(path, payload):
    path.parent.mkdir(parents=True, exist_ok=True)
    handle = os.open(str(path), os.O_CREAT | os.O_TRUNC | os.O_WRONLY, 0o600)
    with os.fdopen(handle, 'w', encoding='utf-8') as stream:
        json.dump(payload, stream, ensure_ascii=False, indent=2)
        stream.write('\n')


def main():
    parser = argparse.ArgumentParser(description='Read-only discrepancy report for ReplacingMergeTree without ver')
    parser.add_argument('--config', default='config.json')
    parser.add_argument('--table', default=DEFAULT_TABLE)
    parser.add_argument('--max-rows', type=int, default=100000)
    parser.add_argument('--report', default='replacing_diff_report.json')
    parser.add_argument('--show', type=int, default=10, help='number of affected keys shown in console')
    args = parser.parse_args()
    if args.max_rows < 1:
        parser.error('--max-rows must be positive')
    db, sep, table = args.table.partition('.')
    if not sep or not db or not table:
        parser.error('--table must be database.table')
    if (db, table) != ('zendendb', 'ПланПродаж_ecom'):
        parser.error('This diagnostic version supports only zendendb.ПланПродаж_ecom; inspect the DDL before extending')

    config_path = Path(args.config).expanduser().resolve()
    config = json.loads(config_path.read_text(encoding='utf-8'))
    source = Client(config['source'], config_path.parent, 'source')
    target = Client(config['target'], config_path.parent, 'target')
    # For the exact source DDL, timestamp is DateTime. Preserve the original
    # values for comparison via toUnixTimestamp to avoid timezone formatting.
    print('Reading source FINAL and raw rows (read only)...', flush=True)
    src_final = read_rows(source, db, table, True, True, args.max_rows)
    src_raw = read_rows(source, db, table, True, False, args.max_rows)
    print('Reading target FINAL and raw rows (read only)...', flush=True)
    dst_final = read_rows(target, db, table, True, True, args.max_rows)
    dst_raw = read_rows(target, db, table, True, False, args.max_rows)
    result = compare(src_final, dst_final, src_raw, dst_raw, DEFAULT_KEYS)
    result['table'] = args.table
    result['note'] = ('Hashes are over JSONEachRow with DateTime normalized to Unix seconds. '
                      'Inspect mismatch keys before any modifications; never retry a partial INSERT blindly.')
    write_private_json(Path(args.report).resolve(), result)
    print(json.dumps({k: result[k] for k in ('source', 'target', 'changed_keys_count',
           'matching_keys_count', 'matched')}, ensure_ascii=False, indent=2))
    if result['missing_on_target']:
        print('Missing on target:', len(result['missing_on_target']))
    if result['missing_on_source']:
        print('Missing on source:', len(result['missing_on_source']))
    print('First differing keys (no cell values shown):')
    for item in result['changed_keys'][:args.show]:
        print(f"  {item['key']}: columns={item['changed_columns']}; "
              f"source_raw_versions={item['source_raw_versions']}; "
              f"target_raw_versions={item['target_raw_versions']}; "
              f"target_winner_in_source_raw={item['target_winner_exists_in_source_raw']}")
    print('Detailed private report:', Path(args.report).resolve())
    print('No data or journal modifications were made.')


if __name__ == '__main__':
    try:
        main()
    except (MigrationError, KeyError, ValueError, OSError) as e:
        print('ERROR:', e)
        raise SystemExit(1)
