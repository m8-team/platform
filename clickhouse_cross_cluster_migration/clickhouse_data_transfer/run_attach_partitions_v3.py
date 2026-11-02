#!/usr/bin/env python3
"""Execute only ALTER TABLE ... ATTACH PARTITION ID ... FROM ... statements safely.

Version 3: supports a target-name -> source-name map without modifying the
original SQL or the progress journal. Python 3.10+, standard library.
"""
from __future__ import annotations

import argparse
import base64
import datetime as dt
import fcntl
import getpass
import difflib
import hashlib
import json
import os
from pathlib import Path
import re
import ssl
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

QIDENT = r'`(?:``|[^`])*`'
QUALIFIED = rf'(?P<db>{QIDENT})\.(?P<table>{QIDENT})'
PARTITION = r"(?P<part>'(?:\\.|''|[^'])*')"
ATTACH_RE = re.compile(
    rf'^\s*ALTER\s+TABLE\s+(?P<dst>{QIDENT}\.{QIDENT})\s+'
    rf'ATTACH\s+PARTITION\s+ID\s+{PARTITION}\s+'
    rf'FROM\s+(?P<src>{QIDENT}\.{QIDENT})\s*;\s*$',
    flags=re.IGNORECASE,
)


def iso():
    return dt.datetime.now(dt.timezone.utc).isoformat(timespec='seconds')


def ident(s: str) -> str:
    return '`' + s.replace('`', '``') + '`'


def sql_string(s: str) -> str:
    return "'" + s.replace('\\', '\\\\').replace("'", "\\'") + "'"


def parse_ident(s: str) -> str:
    return s[1:-1].replace('``', '`')


def parse_qualified(s: str) -> tuple[str, str]:
    m = re.fullmatch(rf'({QIDENT})\.({QIDENT})', s)
    if not m:
        raise ValueError('Invalid backtick-qualified identifier')
    return parse_ident(m.group(1)), parse_ident(m.group(2))


def parse_sql_literal(s: str) -> str:
    # Generator uses simple ClickHouse single-quoted string syntax.
    raw = s[1:-1]
    result = []
    i = 0
    escapes = {'n': '\n', 'r': '\r', 't': '\t', '0': '\0', "'": "'", '\\': '\\'}
    while i < len(raw):
        if raw[i] == "'" and i + 1 < len(raw) and raw[i + 1] == "'":
            result.append("'")
            i += 2
        elif raw[i] == '\\' and i + 1 < len(raw):
            c = raw[i + 1]
            if c not in escapes:
                raise ValueError(f'Unsupported partition ID escape: {c}')
            result.append(escapes[c])
            i += 2
        else:
            result.append(raw[i])
            i += 1
    return ''.join(result)


def sha_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()


def load_table_map(path: str | None) -> dict[str, str]:
    if not path:
        return {}
    with Path(path).expanduser().open(encoding='utf-8') as f:
        mapping = json.load(f)
    if not isinstance(mapping, dict) or not all(
        isinstance(k, str) and k and isinstance(v, str) and v
        for k, v in mapping.items()
    ):
        raise ValueError('--table-map must be a JSON object {"target_table": "source_table"}')
    if len(set(mapping.values())) != len(mapping):
        raise ValueError('Two target table names map to the same source table')
    return mapping


def statements(path: Path, source: str, target: str, table_map: dict[str, str] | None = None):
    # Strict parser. Preserve *original* SQL SHA for already-completed journal entries.
    table_map = table_map or {}
    ix = 0
    with path.open('r', encoding='utf-8-sig') as f:
        for line_number, line in enumerate(f, 1):
            val = line.strip()
            if not val or val.startswith('--'):
                continue
            m = ATTACH_RE.fullmatch(val)
            if not m:
                raise ValueError(f'Unsupported/invalid SQL on line {line_number}: {val[:180]}')
            sdb, stbl = parse_qualified(m.group('src'))
            ddb, dtbl = parse_qualified(m.group('dst'))
            part = parse_sql_literal(m.group('part'))
            src_table = table_map.get(dtbl, dtbl)
            if sdb != source or ddb != target or stbl not in (dtbl, src_table):
                raise ValueError(f'Unexpected source/target on line {line_number}: {val[:150]}')
            # For unmapped tables, source & target names must match.
            if dtbl not in table_map and stbl != dtbl:
                raise ValueError(f'Unexpected different table names on line {line_number}')
            ix += 1
            effective_sql = val
            if stbl != src_table:
                effective_sql = (
                    f'ALTER TABLE {ident(ddb)}.{ident(dtbl)} '
                    f'ATTACH PARTITION ID {sql_string(part)} '
                    f'FROM {ident(sdb)}.{ident(src_table)};'
                )
            yield {
                'index': ix, 'line': line_number, 'table': dtbl,
                'source_table': src_table, 'partition': part,
                'sql': val, 'effective_sql': effective_sql,
                'sha': hashlib.sha256(val.encode('utf-8')).hexdigest(),
            }


class CH:
    def __init__(self, url: str, user: str, password: str, ca: str | None, timeout: int):
        self.url = url.rstrip('/') + '/'
        self.timeout = timeout
        self.auth = 'Basic ' + base64.b64encode(f'{user}:{password}'.encode()).decode('ascii')
        self.ssl_context = ssl.create_default_context(cafile=ca)

    def post(self, sql: str, query_id: str | None = None) -> str:
        query = {'wait_end_of_query': '1'}
        if query_id:
            query['query_id'] = query_id
        url = self.url + '?' + urllib.parse.urlencode(query)
        req = urllib.request.Request(
            url, method='POST', data=sql.encode('utf-8'),
            headers={'Authorization': self.auth, 'Content-Type': 'text/plain; charset=utf-8'})
        try:
            with urllib.request.urlopen(req, context=self.ssl_context, timeout=self.timeout) as f:
                body = f.read().decode('utf-8', errors='replace')
        except urllib.error.HTTPError as e:
            body = e.read(2500).decode('utf-8', errors='replace')
            raise RuntimeError(f'ClickHouse HTTP {e.code}: {body[:2300]}') from e
        if body.lstrip().startswith(('Code: ', 'DB::Exception:')):
            raise RuntimeError(f'ClickHouse response error: {body[:2300]}')
        return body

    def rows(self, sql: str) -> list[dict]:
        result = self.post(sql.rstrip().rstrip(';') + ' FORMAT JSON')
        return json.loads(result)['data']

    def partition(self, db: str, table: str, partition: str) -> dict:
        rows = self.rows(
            'SELECT count() AS parts, sum(rows) AS rows, sum(bytes_on_disk) AS bytes '
            'FROM system.parts '
            f'WHERE active AND database = {sql_string(db)} '
            f'AND table = {sql_string(table)} AND partition_id = {sql_string(partition)}'
        )
        return rows[0]

    def running(self, query_id: str) -> int:
        result = self.rows(
            'SELECT count() AS cnt FROM system.processes '
            f'WHERE query_id = {sql_string(query_id)}')
        return int(result[0]['cnt'])


def read_events(path: Path) -> tuple[dict | None, dict[int, dict]]:
    if not path.exists():
        return None, {}
    meta = None
    state = {}
    with path.open('r', encoding='utf-8') as f:
        for line_num, line in enumerate(f, 1):
            if not line.strip():
                continue
            try:
                event = json.loads(line)
            except json.JSONDecodeError as e:
                raise RuntimeError(f'Corrupt journal at line {line_num}; do not rerun: {e}') from e
            if event['event'] == 'META':
                if meta is not None:
                    raise RuntimeError('Journal contains duplicate META')
                meta = event
            else:
                if meta is None:
                    raise RuntimeError('Journal events before META')
                state[int(event['index'])] = event
    return meta, state


def journal_append(path: Path, event: dict):
    with path.open('a', encoding='utf-8') as f:
        f.write(json.dumps(event, ensure_ascii=False, separators=(',', ':')) + '\n')
        f.flush()
        os.fsync(f.fileno())


def build_parser():
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    p.add_argument('--sql', default='./migration-sql/02_attach_partitions.sql')
    p.add_argument('--state', help='Progress journal; default <sql>.progress.jsonl')
    p.add_argument('--url', help='ClickHouse HTTPS endpoint: https://host:8443')
    p.add_argument('--user', default='analytics')
    p.add_argument('--ca-file', help='CA certificate')
    p.add_argument('--source', default='zendendb_import')
    p.add_argument('--target', default='zendendb')
    p.add_argument('--table-map', help='JSON mapping: target table name -> source table name. Original SQL is unchanged.')
    p.add_argument('--execute', action='store_true', help='Actually run SQL; without this, inspect only')
    p.add_argument('--status', action='store_true', help='Show local journal progress (no connection needed)')
    p.add_argument('--preflight', action='store_true', help='Read-only audit of source/target tables for all remaining ATTACH steps')
    p.add_argument('--limit', type=int, default=0, help='Run only N more commands; 0 = all remaining')
    p.add_argument('--sleep', type=float, default=0.0, help='Pause in seconds after each success')
    p.add_argument('--timeout', type=int, default=7200, help='Network timeout per SQL command, seconds')
    p.add_argument('--resolve', choices=['done', 'retry'],
                   help='Manually resolve UNCERTAIN command after investigating on server')
    p.add_argument('--index', type=int, help='1-based index for --resolve')
    p.add_argument('--ack', help='Required for --resolve: I_VERIFIED_CLICKHOUSE')
    return p



def preflight_tables(ch: CH, sqlpath: Path, source: str, target: str, progress: dict[int, dict], table_map: dict[str, str]) -> bool:
    """Check pending source->target table pairs with one read-only metadata query."""
    expected: dict[tuple[str, str], list[int]] = {}
    for step in statements(sqlpath, source, target, table_map):
        if progress.get(step['index'], {}).get('event') in ('OK', 'RESOLVED_DONE'):
            continue
        expected.setdefault((step['source_table'], step['table']), []).append(step['index'])

    tbl_rows = ch.rows(
        'SELECT database, name, engine, partition_key, sorting_key, primary_key '
        'FROM system.tables '
        f'WHERE database IN ({sql_string(source)}, {sql_string(target)})'
    )
    tbls = {(r['database'], r['name']): r for r in tbl_rows}
    source_names = [name for db, name in tbls if db == source]
    mismatches = []
    good_count = 0
    for (src_table, dst_table), indices in expected.items():
        src = tbls.get((source, src_table))
        dst = tbls.get((target, dst_table))
        need = {'MergeTree': 'ReplicatedMergeTree',
                'ReplacingMergeTree': 'ReplicatedReplacingMergeTree'}.get(src['engine'] if src else None)
        keys_match = bool(src and dst and all(
            src[key] == dst[key] for key in ('partition_key', 'sorting_key', 'primary_key')
        ))
        if not need or not dst or dst['engine'] != need or not keys_match:
            candidates = difflib.get_close_matches(dst_table, source_names, n=3, cutoff=0.65) if not src else []
            mismatches.append((src_table, dst_table, indices[0], len(indices),
                               src['engine'] if src else None, dst['engine'] if dst else None,
                               need, keys_match, candidates))
        else:
            good_count += 1

    print(f'PREFLIGHT (read-only): {len(expected)} source->target pairs with pending ATTACH steps; '
          f'compatible={good_count}, incompatible/missing={len(mismatches)}', flush=True)
    for src_table, dst_table, ix, count, src, dst, need, keys_match, suggestions in mismatches[:100]:
        print(f'  #{ix} {src_table!r} -> {dst_table!r}: {count} pending partitions, '
              f'source_engine={src!r}, target_engine={dst!r}, expected={need!r}, '
              f'keys_match={keys_match}', flush=True)
        if suggestions:
            print('     Similar source names: ' + ', '.join(repr(x) for x in suggestions), flush=True)
    if len(mismatches) > 100:
        print(f'  ... {len(mismatches)-100} additional mismatches.', flush=True)
    print('No SQL mutations performed. Matching engine/key metadata alone does not prove '
          'column and storage-policy compatibility.', flush=True)
    return not mismatches


def main():
    args = build_parser().parse_args()
    table_map = load_table_map(args.table_map)
    sqlpath = Path(args.sql).expanduser().resolve()
    if not sqlpath.is_file():
        raise RuntimeError(f'SQL file not found: {sqlpath}')
    statepath = Path(args.state).expanduser().resolve() if args.state else Path(str(sqlpath) + '.progress.jsonl')
    if args.limit < 0 or args.sleep < 0 or args.timeout <= 0:
        raise ValueError('Invalid --limit / --sleep / --timeout')

    # Exclusive lock prevents two copies of this executor on one machine.
    with Path(str(statepath) + '.lock').open('a+') as lockfile:
        try:
            fcntl.flock(lockfile.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            raise RuntimeError(f'Another executor is running (lock {statepath}.lock)')

        digest = sha_file(sqlpath)
        total = 0
        selected = None
        for step in statements(sqlpath, args.source, args.target, table_map):
            total += 1
            if args.resolve and step['index'] == args.index:
                selected = step
        if total == 0:
            raise RuntimeError('No ATTACH statements found')
        meta, progress = read_events(statepath)
        if meta:
            expected = (digest, args.source, args.target)
            actual = (meta['file_sha256'], meta['source'], meta['target'])
            if actual != expected:
                raise RuntimeError('SQL SHA256/source/target differs from journal. Never reuse old journal for modified SQL')
        # Do not allow remapping an already completed step. We only have the
        # original statement hash in older journals, not a historical mapping.
        for step in statements(sqlpath, args.source, args.target, table_map):
            if step['source_table'] != step['table'] and progress.get(step['index'], {}).get('event') in ('OK', 'RESOLVED_DONE'):
                prev = progress[step['index']]
                if prev.get('source_table', step['table']) != step['source_table']:
                    raise RuntimeError(f'Cannot remap already completed step #{step["index"]}: '
                                       'its source table is not confirmed in old journal')
        good = sum(v['event'] in ('OK', 'RESOLVED_DONE') for v in progress.values())
        pending = [(i, x) for i, x in sorted(progress.items())
                   if x['event'] not in ('OK', 'RESOLVED_DONE', 'RESOLVED_RETRY')]
        print(f'File: {sqlpath}  |  statements: {total}  |  completed: {good}  |  journal: {statepath}', flush=True)
        if pending:
            i, x = pending[0]
            print(f'UNRESOLVED #{i}: {x["event"]}, table={x.get("table")}, partition={x.get("partition")}, '
                  f'query_id={x.get("query_id", "unknown")}', flush=True)

        if args.status:
            if meta:
                print(f'Journal server: {meta.get("url")} host={meta.get("host")} user={meta.get("user")}')
            print(f'Remaining: {total - good}, unresolved entries: {len(pending)}')
            return
        if not (args.execute or args.resolve or args.preflight):
            print('DRY RUN: no SQL executed. First pending statements:')
            shown = 0
            for item in statements(sqlpath, args.source, args.target, table_map):
                if progress.get(item['index'], {}).get('event') in ('OK', 'RESOLVED_DONE'):
                    continue
                print(f'  #{item["index"]}: {item["source_table"]} -> {item["table"]} / {item["partition"]}')
                shown += 1
                if shown == 10:
                    break
            print('Add --execute to run, e.g. --execute --limit 10')
            return
        if not args.url:
            raise ValueError('Provide --url when connecting to ClickHouse')
        if not args.url.startswith(('https://', 'http://')):
            raise ValueError('ClickHouse URL must start with https:// or http://')
        if args.url.startswith('http://'):
            raise ValueError('Refusing unencrypted HTTP for password-based authentication')
        password = os.environ.get('CH_PASSWORD')
        if password is None:
            password = getpass.getpass('ClickHouse password: ')
        ch = CH(args.url, args.user, password, args.ca_file, args.timeout)
        identity = ch.rows('SELECT hostName() AS host, currentUser() AS username')[0]
        host = identity['host']
        actualuser = identity['username']
        dbrows = ch.rows('SELECT name, engine FROM system.databases '
                         f'WHERE name IN ({sql_string(args.source)}, {sql_string(args.target)})')
        dbs = {r['name']: r['engine'] for r in dbrows}
        if args.source not in dbs or dbs.get(args.target) != 'Replicated':
            raise RuntimeError(f'Wrong databases: source must exist and target must be Replicated; got {dbs}')
        if meta and (meta['url'] != args.url.rstrip('/') or meta['host'] != host or meta['user'] != actualuser):
            raise RuntimeError('Server host / URL / user has changed since journal creation; stop to avoid copying data on wrong replica')
        print(f'Connected: {host} as {actualuser}; target database engine: Replicated', flush=True)
        if args.preflight:
            if args.execute or args.resolve:
                raise ValueError('--preflight must be run alone, without --execute / --resolve')
            if not preflight_tables(ch, sqlpath, args.source, args.target, progress, table_map):
                raise RuntimeError('Preflight found incompatible/missing tables. No SQL commands executed.')
            return

        if args.resolve:
            if args.ack != 'I_VERIFIED_CLICKHOUSE' or not selected:
                raise ValueError('To resolve provide --index N --ack I_VERIFIED_CLICKHOUSE')
            previous = progress.get(args.index, {})
            if previous.get('event') not in ('STARTED', 'UNCERTAIN'):
                raise RuntimeError('Only STARTED/UNCERTAIN commands can be manually resolved')
            if previous.get('sha') != selected['sha']:
                raise RuntimeError('SQL hash does not match journal')
            if ch.running(previous['query_id']):
                raise RuntimeError('Query still running on this host. Do not resolve yet')
            stats = ch.partition(args.target, selected['table'], selected['partition'])
            found = int(stats['parts'])
            if args.resolve == 'done' and found <= 0:
                raise RuntimeError('Target partition has no active parts, cannot mark done')
            if args.resolve == 'retry' and found != 0:
                raise RuntimeError('Target partition has active parts: refusing retry (duplication risk)')
            journal_append(statepath, {
                'event': 'RESOLVED_DONE' if args.resolve == 'done' else 'RESOLVED_RETRY',
                'index': args.index, 'sha': selected['sha'], 'time': iso(),
                'table': selected['table'], 'partition': selected['partition'],
                'query_id': previous['query_id'], 'parts': found,
            })
            print(f'Manually resolved #{args.index} as {args.resolve}. Check source and target rows/parts yourself.')
            return

        if pending:
            raise RuntimeError('Uncertain prior statement: no automatic retry. Check system.processes, '
                               'system.query_log and source/target system.parts; then use --resolve done/retry')
        if not meta:
            journal_append(statepath, {
                'event': 'META', 'file_sha256': digest, 'source': args.source,
                'target': args.target, 'url': args.url.rstrip('/'),
                'user': actualuser, 'host': host, 'time': iso(),
            })
        executed = 0
        started = time.monotonic()
        for step in statements(sqlpath, args.source, args.target, table_map):
            previous = progress.get(step['index'])
            if previous and previous['event'] in ('OK', 'RESOLVED_DONE'):
                if previous.get('sha') != step['sha']:
                    raise RuntimeError(f'Journal hash mismatch at #{step["index"]}')
                continue
            if previous and previous['event'] != 'RESOLVED_RETRY':
                raise RuntimeError(f'Unresolved statement #{step["index"]}')
            if args.limit and executed >= args.limit:
                break
            tbl = step['table']
            pid = step['partition']
            src_tbl = step['source_table']
            print(f'[{step["index"]}/{total}] {src_tbl} -> {tbl}  partition={pid}', flush=True)
            table_rows = ch.rows(
                'SELECT database, name, engine, partition_key, sorting_key, primary_key FROM system.tables '
                f'WHERE (database = {sql_string(args.source)} AND name = {sql_string(src_tbl)}) '
                f'OR (database = {sql_string(args.target)} AND name = {sql_string(tbl)})')
            metadata = {(r['database'], r['name']): r for r in table_rows}
            src_meta = metadata.get((args.source, src_tbl))
            dst_meta = metadata.get((args.target, tbl))
            engines = {
                args.source: src_meta['engine'] if src_meta else None,
                args.target: dst_meta['engine'] if dst_meta else None,
            }
            src_engine = engines.get(args.source)
            expected_engine = {'MergeTree': 'ReplicatedMergeTree',
                               'ReplacingMergeTree': 'ReplicatedReplacingMergeTree'}.get(src_engine)
            if not expected_engine or engines.get(args.target) != expected_engine:
                raise RuntimeError(f'Incompatible or missing source/target table {tbl}: {engines}, '
                                   f'expected target={expected_engine}')
            if any(src_meta[key] != dst_meta[key] for key in ('partition_key', 'sorting_key', 'primary_key')):
                raise RuntimeError(f'Incompatible partition/sorting/primary keys: {src_tbl} -> {tbl}')
            src = ch.partition(args.source, src_tbl, pid)
            dst = ch.partition(args.target, tbl, pid)
            if int(src['parts']) == 0:
                raise RuntimeError(f'Source partition {tbl}/{pid} has zero active parts; refuse ATTACH')
            if int(dst['parts']) != 0:
                raise RuntimeError(f'Target partition {tbl}/{pid} already has {dst["parts"]} active parts; refuse duplicate ATTACH')

            query_id = 'migrate-attach-' + str(uuid.uuid4())
            journal_append(statepath, {
                'event': 'STARTED', 'time': iso(), 'index': step['index'],
                'sha': step['sha'], 'query_id': query_id,
                'table': tbl, 'source_table': src_tbl, 'partition': pid,
                'src_parts_before': int(src['parts']),
                'src_rows_before': int(src['rows'] or 0),
            })
            t0 = time.monotonic()
            try:
                ch.post(step['effective_sql'], query_id)
            except Exception as e:
                journal_append(statepath, {
                    'event': 'UNCERTAIN', 'time': iso(), 'index': step['index'],
                    'sha': step['sha'], 'query_id': query_id,
                    'table': tbl, 'source_table': src_tbl, 'partition': pid, 'error': str(e)[:1000],
                })
                raise RuntimeError(f'STOP: ATTACH result uncertain for #{step["index"]} '
                                   f'({tbl}/{pid}): {e}. DO NOT RETRY until reconciliation') from e
            journal_append(statepath, {
                'event': 'OK', 'time': iso(), 'index': step['index'],
                'sha': step['sha'], 'query_id': query_id,
                'table': tbl, 'source_table': src_tbl, 'partition': pid,
            })
            executed += 1
            progress[step['index']] = {'event': 'OK', 'sha': step['sha'], 'source_table': src_tbl}
            print(f'  OK  {time.monotonic() - t0:.1f}s; total succeeded this run: {executed}', flush=True)
            if args.sleep:
                time.sleep(args.sleep)
        print(f'Done for this run: {executed} ATTACH statements, elapsed {time.monotonic() - started:.1f}s', flush=True)
        print('Rerun the same --execute command to continue; completed statements will be skipped.')


if __name__ == '__main__':
    try:
        main()
    except (RuntimeError, ValueError, urllib.error.URLError, OSError, json.JSONDecodeError) as exc:
        print('ERROR:', exc, file=sys.stderr, flush=True)
        sys.exit(2)
    except KeyboardInterrupt:
        print('\nINTERRUPTED. If an ATTACH was in flight, result is UNKNOWN. '
              'Do not retry blindly; reconcile using the journal and ClickHouse.', file=sys.stderr)
        sys.exit(130)
