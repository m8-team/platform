#!/usr/bin/env python3
"""Safe cross-cluster copy of zendendb.ЦеныНоменклатуры in monthly batches.

Source: MergeTree PARTITION BY toYYYYMMDD(`Период`).
Target: ReplicatedMergeTree with identical schema, one shard, two replicas, Replicated DB.
Uses only Python standard library. Does not modify source or main migration journal.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import ssl
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

TABLE = ('zendendb', 'ЦеныНоменклатуры')
EXPECTED_SOURCE_ENGINE = 'MergeTree'
EXPECTED_TARGET_ENGINE = 'ReplicatedMergeTree'

class MigrationError(RuntimeError):
    pass

def qi(s):
    return '`' + s.replace('\\', '\\\\').replace('`', '\\`') + '`'

def qs(s):
    return "'" + str(s).replace('\\', '\\\\').replace("'", "\\'") + "'"

class Client:
    def __init__(self, cfg, basedir, label):
        self.label = label
        self.username = cfg['username']
        self.password = cfg['password']
        self.host = cfg['host'].strip()
        scheme = 'https' if cfg.get('secure', True) else 'http'
        self.url = f"{scheme}://{self.host}:{cfg.get('port', 8443)}/"
        self.timeout = int(cfg.get('timeout_seconds', 7200))
        self.ctx = None
        if cfg.get('secure', True):
            ca = cfg.get('ca_cert')
            if ca:
                ca = Path(ca).expanduser()
                if not ca.is_absolute():
                    ca = basedir / ca
                if not ca.is_file():
                    raise MigrationError(f'{label}: missing CA file {ca}')
                self.ctx = ssl.create_default_context(cafile=str(ca))
            else:
                self.ctx = ssl.create_default_context()

    def query(self, sql):
        req = urllib.request.Request(
            self.url + '?' + urllib.parse.urlencode({'wait_end_of_query': '1'}),
            data=sql.encode('utf-8'), method='POST', headers={
                'X-ClickHouse-User': self.username,
                'X-ClickHouse-Key': self.password,
                'Content-Type': 'text/plain; charset=utf-8'})
        try:
            with urllib.request.urlopen(req, timeout=self.timeout, context=self.ctx) as r:
                return r.read().decode('utf-8')
        except urllib.error.HTTPError as exc:
            msg = exc.read(2000).decode('utf-8', 'replace')
            raise MigrationError(f'{self.label}: HTTP {exc.code}: {msg}') from None
        except (TimeoutError, OSError, urllib.error.URLError) as exc:
            raise MigrationError(f'{self.label}: connection error: {exc}') from None

    def rows(self, sql):
        raw = self.query(sql.strip().rstrip(';') + ' FORMAT JSONEachRow')
        return [json.loads(line) for line in raw.splitlines() if line.strip()]

    def scalar_text(self, sql):
        return self.query(sql.strip().rstrip(';') + ' FORMAT TabSeparated').strip()


def metadata(c):
    r = c.rows(f"SELECT engine, partition_key, sorting_key FROM system.tables WHERE database={qs(TABLE[0])} AND name={qs(TABLE[1])}")
    if len(r) != 1:
        raise MigrationError(f'{c.label}: missing table {TABLE[0]}.{TABLE[1]}')
    return r[0]

def ensure_schema(source, target):
    src, dst = metadata(source), metadata(target)
    if src['engine'] != EXPECTED_SOURCE_ENGINE or dst['engine'] != EXPECTED_TARGET_ENGINE:
        raise MigrationError(f'Engine mismatch: source={src["engine"]}, target={dst["engine"]}')
    pat = re.compile(r'^toYYYYMMDD\(\s*`?Период`?\s*\)$', re.I)
    if not pat.fullmatch(src['partition_key']) or src['partition_key'] != dst['partition_key']:
        raise MigrationError(f'Unexpected partition expression: source={src["partition_key"]}, target={dst["partition_key"]}')
    if src['sorting_key'] != dst['sorting_key']:
        raise MigrationError('Different sorting keys; migration cancelled')
    table = f'{qi(TABLE[0])}.{qi(TABLE[1])}'
    src_cols = source.rows(f'DESCRIBE TABLE {table}')
    dst_cols = target.rows(f'DESCRIBE TABLE {table}')
    if src_cols != dst_cols:
        raise MigrationError('Source/target DESCRIBE TABLE differ; migration cancelled')
    src_tz, dst_tz = source.scalar_text('SELECT timezone()'), target.scalar_text('SELECT timezone()')
    if src_tz != dst_tz:
        raise MigrationError(f'Default timezones differ ({src_tz} vs {dst_tz}); avoid shifting DateTime daily partitions')


def status(target, cluster):
    sql = ('SELECT hostName() AS host, is_readonly, is_session_expired, '
           'total_replicas, active_replicas, queue_size '
           f'FROM clusterAllReplicas({qs(cluster)}, system.replicas) '
           f'WHERE database={qs(TABLE[0])} AND table={qs(TABLE[1])} ORDER BY host')
    return target.rows(sql)

def require_replicas(rows, expected):
    hosts = {str(x['host']) for x in rows}
    if len(rows) != expected or len(hosts) != expected or not all(
        int(x['total_replicas']) == expected and int(x['active_replicas']) == expected
        and int(x['is_readonly']) == 0 and int(x['is_session_expired']) == 0
        and int(x['queue_size']) == 0 for x in rows):
        raise MigrationError(f'Replicas are not healthy: {json.dumps(rows, ensure_ascii=False)}')
    return hosts


def aggregate_select(month=None, local_host=False):
    table = f'{qi(TABLE[0])}.{qi(TABLE[1])}'
    where = f' WHERE toYYYYMM(`Период`) = {int(month)}' if month is not None else ''
    if month is None:
        return (f'SELECT toYYYYMM(`Период`) AS month, count() AS rows, '
                'toString(sum(cityHash64(*))) AS checksum_sum, '
                'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
                f'FROM {table} GROUP BY month')
    host_field = 'hostName() AS host, ' if local_host else ''
    return (f'SELECT {host_field}count() AS rows, '
            'toString(sum(cityHash64(*))) AS checksum_sum, '
            'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
            f'FROM {table}{where}')


def normalize(r):
    return tuple(str(r[k]) for k in ('rows', 'checksum_sum', 'checksum_xor'))

def month_summary(c):
    return {int(r['month']): normalize(r) for r in c.rows(aggregate_select())}

def one_source_month(source, month):
    r = source.rows(aggregate_select(month))
    if len(r) != 1:
        raise MigrationError('Unable to read source fingerprint')
    return normalize(r[0])

def replicas_month(target, cluster, month, expected_hosts):
    inner = aggregate_select(month, local_host=True)
    records = target.rows(
        f'SELECT host, rows, checksum_sum, checksum_xor FROM '
        f'clusterAllReplicas({qs(cluster)}, view({inner})) ORDER BY host')
    by_host = {str(r['host']): normalize(r) for r in records}
    if set(by_host) != set(expected_hosts) or len(records) != len(expected_hosts):
        raise MigrationError(f'Month {month}: fingerprints missing replica hosts: {records}')
    return by_host

def classify(src_fp, target_by_host):
    values = list(target_by_host.values())
    if not values:
        return 'CONFLICT'
    if all(fp == src_fp for fp in values):
        return 'COMPLETE'
    if all(int(fp[0]) == 0 for fp in values):
        return 'EMPTY'
    return 'CONFLICT'

def save_state(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    tmp = path.with_suffix('.tmp')
    fd = os.open(str(tmp), os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, 'w', encoding='utf-8') as fh:
        json.dump(data, fh, ensure_ascii=False, indent=2)
        fh.flush()
        os.fsync(fh.fileno())
    os.replace(tmp, path)

def utcnow():
    return datetime.now(timezone.utc).isoformat()

def redact(text, secrets):
    for s in secrets:
        if s:
            text = text.replace(s, '[REDACTED]')
    return text


def main():
    p = argparse.ArgumentParser(description='Transfer daily-partitioned ClickHouse pricing table in monthly INSERTs')
    p.add_argument('--config', default='config.json')
    p.add_argument('--table', default='zendendb.ЦеныНоменклатуры')
    p.add_argument('--apply', action='store_true')
    p.add_argument('--ack-source-frozen', action='store_true')
    p.add_argument('--limit-months', type=int)
    p.add_argument('--state', default='data_transfer_progress/prices_monthly_progress.json')
    p.add_argument('--replica-wait-seconds', type=int, default=300)
    args = p.parse_args()
    if args.table != '.'.join(TABLE):
        raise MigrationError('This safety-checked tool supports only zendendb.ЦеныНоменклатуры')
    if args.apply and not args.ack_source_frozen:
        raise MigrationError('--apply requires --ack-source-frozen (all source writes must actually be paused)')
    if args.limit_months is not None and args.limit_months < 1:
        raise MigrationError('--limit-months must be positive')
    cfg_path = Path(args.config).expanduser().resolve()
    cfg = json.loads(cfg_path.read_text(encoding='utf-8'))
    source = Client(cfg['source'], cfg_path.parent, 'source')
    target = Client(cfg['target'], cfg_path.parent, 'target')
    secrets = [source.password, target.password]
    cluster = cfg.get('target_cluster', 'default')
    expected = int(cfg.get('expected_replicas', 2))
    try:
        ensure_schema(source, target)
        hosts = require_replicas(status(target, cluster), expected)
        src_summary = month_summary(source)
        # A global target summary also spots unexpected target-only months.
        dst_summary = month_summary(target)
        unexpected = set(dst_summary) - set(src_summary)
        if unexpected:
            raise MigrationError('Target has months absent from source: ' + ', '.join(map(str, sorted(unexpected))))
        months = sorted(src_summary)
        state_path = Path(args.state).expanduser().resolve()
        existing = json.loads(state_path.read_text(encoding='utf-8')) if state_path.is_file() else {}
        if existing:
            if existing.get('source') != source.host or existing.get('target') != target.host:
                raise MigrationError('Monthly state file belongs to a different source or target')
        state = existing or {'source': source.host, 'target': target.host, 'table': args.table, 'months': {}}
        print(f'Table: {args.table}; months with source rows: {len(months)}, replicas: {len(hosts)}')
        ready, complete, conflicts = [], [], []
        # Get every month from both replicas before ANY INSERT to rule out partial previous attempt.
        for month in months:
            fp = one_source_month(source, month)
            if fp != src_summary[month]:
                raise MigrationError(f'Source changed during scan for {month}; writes must remain paused')
            dst = replicas_month(target, cluster, month, hosts)
            category = classify(fp, dst)
            print(f'{month} | {category:8} | source_rows={fp[0]} | target_rows={list(x[0] for x in dst.values())}', flush=True)
            if category == 'COMPLETE':
                complete.append(month)
            elif category == 'EMPTY':
                ready.append(month)
            else:
                conflicts.append(month)
        if conflicts:
            raise MigrationError('Partial or inconsistent months on target: ' + ', '.join(map(str, conflicts)) +
                                 '. DO NOT retry or DROP data automatically; inspect specific months first')
        if not args.apply:
            print(f'DRY RUN: complete={len(complete)} empty={len(ready)} conflicted=0; no changes made')
            return 0
        if not months:
            print('Source table empty; nothing to copy')
            return 0
        chosen = ready[:args.limit_months] if args.limit_months else ready
        origin = f"{source.host}:{int(cfg.get('source_native_tls_port', 9440))}"
        table = f'{qi(TABLE[0])}.{qi(TABLE[1])}'
        for i, month in enumerate(chosen, 1):
            prior_status = state['months'].get(str(month), {}).get('status')
            if prior_status in ('INSERTING', 'UNCERTAIN'):
                raise MigrationError(f'Month {month} has ambiguous previous state {prior_status}. '
                                     'Review query logs and journal before trying again')
            require_replicas(status(target, cluster), expected)
            before = one_source_month(source, month)
            if before != src_summary[month]:
                raise MigrationError(f'Source changed before month {month}; stopping')
            # Re-check emptiness immediately before insertion.
            if classify(before, replicas_month(target, cluster, month, hosts)) != 'EMPTY':
                raise MigrationError(f'Month {month} no longer empty on both replicas; stopping')
            state['months'][str(month)] = {'status': 'INSERTING', 'source_fingerprint': before, 'started_at': utcnow()}
            save_state(state_path, state)
            sql = (f'INSERT INTO {table} SELECT * FROM '
                   f'remoteSecure({qs(origin)}, {qs(TABLE[0])}, {qs(TABLE[1])}, '
                   f'{qs(source.username)}, {qs(source.password)}) '
                   f'WHERE toYYYYMM(`Период`) = {month}')
            print(f'[{i}/{len(chosen)}] COPYING month={month}, rows={before[0]}', flush=True)
            try:
                target.query(sql)
                deadline = time.monotonic() + args.replica_wait_seconds
                while True:
                    after = one_source_month(source, month)
                    if after != before:
                        raise MigrationError('Source changed during copy')
                    dst = replicas_month(target, cluster, month, hosts)
                    current = status(target, cluster)
                    if classify(before, dst) == 'COMPLETE' and all(
                        int(x['queue_size']) == 0 and int(x['is_readonly']) == 0 and int(x['active_replicas']) == expected
                        for x in current) and len(current) == expected:
                        break
                    if time.monotonic() >= deadline:
                        raise MigrationError(f'Month {month} replica verification timeout: {dst}')
                    time.sleep(5)
                state['months'][str(month)] = {'status': 'DONE', 'source_fingerprint': before, 'completed_at': utcnow()}
                save_state(state_path, state)
                print(f'[{i}/{len(chosen)}] DONE month={month} rows={before[0]} both replicas', flush=True)
            except Exception as exc:
                state['months'][str(month)] = {'status': 'UNCERTAIN', 'source_fingerprint': before,
                                              'error': redact(str(exc), secrets)[:1600], 'failed_at': utcnow()}
                save_state(state_path, state)
                raise MigrationError(f'Month {month}: uncertain transfer; inspect target before retry. '
                                     + redact(str(exc), secrets)[:1000]) from None
        # Check all months again, even if --limit-months was specified.
        src_end = month_summary(source)
        if src_end != src_summary:
            raise MigrationError('Source changed during operation; no global completion guarantee')
        target_end = month_summary(target)
        all_equal = src_end == target_end
        print(f'BATCH FINISHED: new months={len(chosen)}, previously complete={len(complete)}, '
              f'pending={len(ready)-len(chosen)}; first-target full comparison={"MATCH" if all_equal else "INCOMPLETE"}')
        if all_equal:
            print('Run the original transfer_data.py --only zendendb.ЦеныНоменклатуры --reconcile --mark-done --ack-source-frozen to update the MAIN journal.')
        return 0
    except MigrationError as exc:
        raise MigrationError(redact(str(exc), secrets)) from None

if __name__ == '__main__':
    try:
        sys.exit(main())
    except (MigrationError, OSError, ValueError, KeyError, json.JSONDecodeError) as exc:
        print('ERROR:', str(exc), file=sys.stderr)
        sys.exit(1)
