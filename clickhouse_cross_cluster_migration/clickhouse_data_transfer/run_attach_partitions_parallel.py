#!/usr/bin/env python3
"""Parallel, journal-compatible executor for ATTACH PARTITION FROM (v3 SQL/journal).

Runs at most one ATTACH at a time per destination table. Uses a single durable
journal, and stops scheduling after any error; in-flight work drains and records
its own outcome. Requires run_attach_partitions_v3.py in the same directory.
Python >=3.10; macOS/Linux; no external dependencies.
"""
from __future__ import annotations

import argparse
from collections import defaultdict, deque
from concurrent.futures import ThreadPoolExecutor, wait, FIRST_COMPLETED
import datetime as dt
import fcntl
import getpass
import hashlib
import http.client
import json
import os
from pathlib import Path
import ssl
import sys
import threading
import time
from urllib.parse import urlsplit, urlencode
import uuid

import run_attach_partitions_v3 as v3


def make_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    p.add_argument('--sql', default='./migration-sql/02_attach_partitions.sql')
    p.add_argument('--state', help='Progress journal, default <sql>.progress.jsonl')
    p.add_argument('--url', help='ClickHouse HTTPS endpoint')
    p.add_argument('--user', default='analytics')
    p.add_argument('--ca-file')
    p.add_argument('--source', default='zendendb_import')
    p.add_argument('--target', default='zendendb')
    p.add_argument('--table-map', help='JSON mapping {"target_table": "source_table"}')
    p.add_argument('--workers', type=int, default=4, help='Concurrent tables/ATTACH commands, 1..16')
    p.add_argument('--limit', type=int, default=0, help='Number of additional operations, 0 means all')
    p.add_argument('--timeout', type=int, default=7200, help='HTTP request timeout in seconds')
    p.add_argument('--sleep', type=float, default=0.0, help='Per-worker pause after each successful ATTACH')
    p.add_argument('--progress-every', type=int, default=20)
    mode = p.add_mutually_exclusive_group()
    mode.add_argument('--execute', action='store_true', help='Actually execute SQL')
    mode.add_argument('--preflight', action='store_true', help='Only check source and target metadata')
    mode.add_argument('--status', action='store_true', help='Only inspect the local journal')
    return p


class PersistentCH:
    """One HTTPS keep-alive connection per worker. Never retry mutations."""
    def __init__(self, url: str, user: str, password: str, ca: str | None, timeout: int):
        from base64 import b64encode
        parsed = urlsplit(url)
        if parsed.scheme != 'https' or not parsed.hostname or parsed.username or parsed.password or parsed.query or parsed.fragment:
            raise ValueError('Expected HTTPS URL without credentials, query or fragment')
        self.host = parsed.hostname
        self.port = parsed.port or 443
        self.path = parsed.path or '/'
        self.timeout = timeout
        self.context = ssl.create_default_context(cafile=ca)
        self.headers = {
            'Authorization': 'Basic ' + b64encode((user + ':' + password).encode()).decode(),
            'Content-Type': 'text/plain; charset=utf-8',
        }
        self.conn: http.client.HTTPSConnection | None = None

    def _connection(self):
        if self.conn is None:
            self.conn = http.client.HTTPSConnection(self.host, self.port, timeout=self.timeout, context=self.context)
        return self.conn

    def post(self, sql: str, query_id: str | None = None) -> str:
        params = {'wait_end_of_query': '1'}
        if query_id:
            params['query_id'] = query_id
        path = self.path + ('&' if '?' in self.path else '?') + urlencode(params)
        try:
            conn = self._connection()
            conn.request('POST', path, body=sql.encode('utf-8'), headers=self.headers)
            response = conn.getresponse()
            body = response.read().decode('utf-8', errors='replace')
            if response.status >= 400:
                raise RuntimeError(f'ClickHouse HTTP {response.status}: {body[:2000]}')
            if body.lstrip().startswith(('Code: ', 'DB::Exception:')):
                raise RuntimeError(f'ClickHouse response error: {body[:2000]}')
            return body
        except Exception:
            self.close()
            raise

    def rows(self, sql: str) -> list[dict]:
        return json.loads(self.post(sql.rstrip().rstrip(';') + ' FORMAT JSON'))['data']

    def close(self):
        if self.conn is not None:
            try:
                self.conn.close()
            finally:
                self.conn = None


class Executor:
    def __init__(self, args, statepath: Path, password: str):
        self.args = args
        self.statepath = statepath
        self.password = password
        self.tlocal = threading.local()
        self.journal_lock = threading.Lock()
        self.stats_lock = threading.Lock()
        self.ok_run = 0
        self.start_time = time.monotonic()
        self.total = 0

    def client(self) -> PersistentCH:
        if not hasattr(self.tlocal, 'ch'):
            self.tlocal.ch = PersistentCH(self.args.url, self.args.user, self.password, self.args.ca_file, self.args.timeout)
        return self.tlocal.ch

    def record(self, event: dict):
        # Only one process holds the file lock; this mutex serializes threads.
        with self.journal_lock:
            v3.journal_append(self.statepath, event)

    def check_partition(self, step: dict) -> None:
        a = self.args
        src_tbl, dst_tbl = step['source_table'], step['table']
        partition = step['partition']
        ch = self.client()
        # A single metadata read handles both tables; no scanning table data.
        rows = ch.rows(
            'SELECT database, name, engine, partition_key, sorting_key, primary_key '
            'FROM system.tables '
            f'WHERE (database={v3.sql_string(a.source)} AND name={v3.sql_string(src_tbl)}) '
            f'OR (database={v3.sql_string(a.target)} AND name={v3.sql_string(dst_tbl)})'
        ) if not hasattr(self.tlocal, 'table_checked') or (src_tbl, dst_tbl) not in self.tlocal.table_checked else None
        if rows is not None:
            m = {(r['database'], r['name']): r for r in rows}
            src = m.get((a.source, src_tbl))
            dst = m.get((a.target, dst_tbl))
            target_engine = {'MergeTree': 'ReplicatedMergeTree', 'ReplacingMergeTree': 'ReplicatedReplacingMergeTree'}.get(src['engine'] if src else None)
            if not target_engine or not dst or dst['engine'] != target_engine:
                raise RuntimeError(f'Incompatible/missing table: {src_tbl} -> {dst_tbl}; expected={target_engine}, actual={dst["engine"] if dst else None}')
            if any(src[k] != dst[k] for k in ('partition_key', 'sorting_key', 'primary_key')):
                raise RuntimeError(f'Partition, sorting or primary key mismatch: {src_tbl} -> {dst_tbl}')
            if not hasattr(self.tlocal, 'table_checked'):
                self.tlocal.table_checked = set()
            self.tlocal.table_checked.add((src_tbl, dst_tbl))
        pairs = ch.rows(
            'SELECT database, table, count() AS parts, sum(rows) AS rows '
            'FROM system.parts '
            'WHERE active AND '
            f'partition_id={v3.sql_string(partition)} AND ('
            f'(database={v3.sql_string(a.source)} AND table={v3.sql_string(src_tbl)}) '
            'OR '
            f'(database={v3.sql_string(a.target)} AND table={v3.sql_string(dst_tbl)})) '
            'GROUP BY database, table'
        )
        stats = {(r['database'], r['table']): r for r in pairs}
        src = stats.get((a.source, src_tbl))
        dst = stats.get((a.target, dst_tbl))
        if src is None or int(src['parts']) == 0:
            raise RuntimeError(f'Source has no active parts: {src_tbl} / {partition}')
        if dst is not None and int(dst['parts']) > 0:
            raise RuntimeError(f'Target already has active parts: {dst_tbl} / {partition}; manual reconciliation required')
        return int(src['parts']), int(src['rows'])

    def execute_one(self, step: dict):
        idx, tbl, src_tbl, pid = step['index'], step['table'], step['source_table'], step['partition']
        # All read-only errors here are before STARTED and are safe to diagnose.
        src_parts, src_rows = self.check_partition(step)
        query_id = 'migrate-attach-' + str(uuid.uuid4())
        self.record({
            'event': 'STARTED', 'time': v3.iso(), 'index': idx, 'sha': step['sha'],
            'query_id': query_id, 'table': tbl, 'source_table': src_tbl, 'partition': pid,
            'src_parts_before': src_parts, 'src_rows_before': src_rows,
        })
        t0 = time.monotonic()
        try:
            self.client().post(step['effective_sql'], query_id)
        except Exception as exc:
            self.record({
                'event': 'UNCERTAIN', 'time': v3.iso(), 'index': idx, 'sha': step['sha'],
                'query_id': query_id, 'table': tbl, 'source_table': src_tbl, 'partition': pid,
                'error': str(exc)[:1000],
            })
            raise RuntimeError(f'UNCERTAIN #{idx} {tbl} / {pid}, query_id={query_id}: {exc}') from exc
        self.record({
            'event': 'OK', 'time': v3.iso(), 'index': idx, 'sha': step['sha'],
            'query_id': query_id, 'table': tbl, 'source_table': src_tbl, 'partition': pid,
        })
        if self.args.sleep:
            time.sleep(self.args.sleep)
        with self.stats_lock:
            self.ok_run += 1
            done = self.ok_run
            if done == 1 or done % self.args.progress_every == 0:
                rate = done / max(0.001, time.monotonic() - self.start_time)
                print(f'OK run={done}, global={self.total + done}, last=#{idx} ({tbl}/{pid}), '
                      f'last_latency={time.monotonic()-t0:.2f}s, rate={rate:.2f}/s', flush=True)
        return True


def main():
    args = make_parser().parse_args()
    if not (1 <= args.workers <= 16) or args.limit < 0 or args.timeout <= 0 or args.sleep < 0 or args.progress_every < 1:
        raise ValueError('Required: 1 <= workers <= 16, limit >= 0, timeout > 0, sleep >= 0, progress-every >= 1')
    table_map = v3.load_table_map(args.table_map)
    sqlpath = Path(args.sql).expanduser().resolve()
    if not sqlpath.is_file():
        raise RuntimeError('SQL file not found: ' + str(sqlpath))
    statepath = Path(args.state).expanduser().resolve() if args.state else Path(str(sqlpath) + '.progress.jsonl')
    with Path(str(statepath) + '.lock').open('a+') as lockfile:
        try:
            fcntl.flock(lockfile.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            raise RuntimeError('Another attach executor currently holds progress lock; stop it first')
        digest = v3.sha_file(sqlpath)
        all_steps = list(v3.statements(sqlpath, args.source, args.target, table_map))
        if not all_steps:
            raise RuntimeError('No ATTACH statements in SQL file')
        meta, progress = v3.read_events(statepath)
        if meta and (meta.get('file_sha256'), meta.get('source'), meta.get('target')) != (digest, args.source, args.target):
            raise RuntimeError('SQL SHA256/source/target differs from existing progress journal. STOP.')
        for event_idx, event in progress.items():
            if event_idx < 1 or event_idx > len(all_steps):
                raise RuntimeError('Journal references non-existent SQL step')
            step = all_steps[event_idx-1]
            if event.get('sha') != step['sha']:
                raise RuntimeError(f'Journal SQL hash mismatch at #{event_idx}')
            if step['source_table'] != step['table'] and event.get('event') in ('OK', 'RESOLVED_DONE'):
                if event.get('source_table', step['table']) != step['source_table']:
                    raise RuntimeError(f'Remapping already completed step #{event_idx} is forbidden')
        completed_statuses = {'OK', 'RESOLVED_DONE'}
        completed = sum(event['event'] in completed_statuses for event in progress.values())
        unresolved = [(ix, r) for ix, r in sorted(progress.items())
                      if r['event'] not in (*completed_statuses, 'RESOLVED_RETRY')]
        remaining = [step for step in all_steps if progress.get(step['index'], {}).get('event') not in completed_statuses]
        print(f'File={sqlpath} statements={len(all_steps)} completed={completed} remaining={len(remaining)} '
              f'unresolved={len(unresolved)} journal={statepath}', flush=True)
        if unresolved:
            ix, event = unresolved[0]
            print(f'UNRESOLVED #{ix}: {event.get("event")}, query_id={event.get("query_id")}')
        if args.status:
            if meta:
                print(f'Journal server: {meta.get("url")} host={meta.get("host")} user={meta.get("user")}')
            return
        if unresolved:
            raise RuntimeError('Unresolved STARTED/UNCERTAIN entries. Use v3 --resolve after manual reconciliation')
        if not args.execute and not args.preflight:
            print('DRY RUN: distinct remaining tables:', len({s['table'] for s in remaining}))
            print('Use --preflight or --execute. No changes performed.')
            return
        if not args.url or not args.url.startswith('https://'):
            raise ValueError('HTTPS --url is required')
        password = os.environ.get('CH_PASSWORD')
        if password is None:
            password = getpass.getpass('ClickHouse password: ')
        ch = v3.CH(args.url, args.user, password, args.ca_file, args.timeout)
        identity = ch.rows('SELECT hostName() AS host, currentUser() AS username')[0]
        dbrows = ch.rows('SELECT name, engine FROM system.databases '
                         f'WHERE name IN ({v3.sql_string(args.source)}, {v3.sql_string(args.target)})')
        dbs = {r['name']: r['engine'] for r in dbrows}
        if args.source not in dbs or dbs.get(args.target) != 'Replicated':
            raise RuntimeError(f'Source must exist, target must have Replicated database engine: {dbs}')
        if meta and (meta['url'] != args.url.rstrip('/') or meta['host'] != identity['host'] or meta['user'] != identity['username']):
            raise RuntimeError('Host/URL/user differs from progress journal META; refusing to run')
        print(f'Connected {identity["host"]} as {identity["username"]}; target engine Replicated', flush=True)
        # Always do read-only preflight, including under --execute.
        good = v3.preflight_tables(ch, sqlpath, args.source, args.target, progress, table_map)
        if not good:
            raise RuntimeError('Preflight failed; no mutations attempted')
        if args.preflight:
            return
        if not remaining:
            print('All steps already complete.')
            return
        if not meta:
            v3.journal_append(statepath, {
                'event': 'META', 'file_sha256': digest, 'source': args.source,
                'target': args.target, 'url': args.url.rstrip('/'),
                'user': identity['username'], 'host': identity['host'], 'time': v3.iso(),
            })
        # For a small smoke-test limit, select round-robin over tables rather
        # than taking only the first table's partitions from the SQL file.
        if args.limit and args.limit < len(remaining):
            per_table = defaultdict(deque)
            for step in remaining:
                per_table[step['table']].append(step)
            table_round_robin = deque(per_table.keys())
            sampled = []
            while table_round_robin and len(sampled) < args.limit:
                tbl = table_round_robin.popleft()
                sampled.append(per_table[tbl].popleft())
                if per_table[tbl]:
                    table_round_robin.append(tbl)
            remaining = sampled
        grouped = defaultdict(deque)
        for step in remaining:
            grouped[step['table']].append(step)
        ready = deque(grouped.keys())
        executor = Executor(args, statepath, password)
        executor.total = completed
        print(f'Starting {len(remaining)} steps, {len(grouped)} tables, {args.workers} workers. '
              'Only one in-flight ATTACH per target table.', flush=True)
        failures = []
        with ThreadPoolExecutor(max_workers=args.workers, thread_name_prefix='ch-attach') as pool:
            inflight = {}
            while ready or inflight:
                while ready and len(inflight) < args.workers and not failures:
                    tbl = ready.popleft()
                    step = grouped[tbl].popleft()
                    fut = pool.submit(executor.execute_one, step)
                    inflight[fut] = (tbl, step)
                if not inflight:
                    break
                finished, _ = wait(inflight, return_when=FIRST_COMPLETED)
                for fut in finished:
                    tbl, step = inflight.pop(fut)
                    try:
                        fut.result()
                    except Exception as exc:
                        failures.append((step['index'], str(exc)))
                        print(f'STOP scheduling new work: #{step["index"]}: {exc}', file=sys.stderr, flush=True)
                    else:
                        if grouped[tbl]:
                            ready.append(tbl)
                # On failure, drain all already-running tasks. Each completed task
                # appends its own OK or UNCERTAIN event before returning.
        if failures:
            print(f'Finished draining workers; successes this run: {executor.ok_run}; errors={len(failures)}. '
                  'Do not restart until journal/status and errors are checked.', file=sys.stderr, flush=True)
            raise RuntimeError('Stopped on error; see the first failure and journal above')
        print(f'Done: {executor.ok_run} additional ATTACH operations completed in '
              f'{time.monotonic()-executor.start_time:.1f}s. '
              'Rerun same command to continue; original journal preserved.', flush=True)


if __name__ == '__main__':
    try:
        main()
    except (RuntimeError, ValueError, OSError, json.JSONDecodeError) as exc:
        print('ERROR:', exc, file=sys.stderr, flush=True)
        sys.exit(2)
    except KeyboardInterrupt:
        print('INTERRUPTED: in-flight ATTACH may still be running. Check journal for STARTED and '
              'reconcile it with ClickHouse before any restart.', file=sys.stderr, flush=True)
        sys.exit(130)
