#!/usr/bin/env python3
"""Conservative ClickHouse to ClickHouse data migration via remoteSecure (v5).

Source is read-only. Target must contain empty, matching Replicated*MergeTree tables.
Copy is server-to-server. Failures are intentionally NOT retried automatically.
Python 3.9+ standard library only.
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


class MigrationError(RuntimeError):
    pass


def ident(s):
    return '`' + s.replace('\\', '\\\\').replace('`', '\\`') + '`'


def lit(s):
    return "'" + str(s).replace('\\', '\\\\').replace("'", "\\'") + "'"


def now():
    return datetime.now(timezone.utc).isoformat()


def safe_error(exc, secrets):
    msg = str(exc)
    for item in secrets:
        if item:
            msg = msg.replace(item, '[REDACTED]')
    return msg[:1800]


class Client:
    def __init__(self, info, base, name):
        self.name = name
        self.host = info['host'].strip()
        self.username = info['username']
        self.password = info['password']
        self.timeout = int(info.get('timeout_seconds', 7200))
        self.url = ('https' if info.get('secure', True) else 'http') + '://' + self.host + ':' + str(info.get('port', 8443)) + '/'
        self.context = None
        if info.get('secure', True):
            cert = info.get('ca_cert')
            if cert:
                p = Path(cert).expanduser()
                if not p.is_absolute():
                    p = base / p
                if not p.is_file():
                    raise MigrationError(f'{name}: CA certificate does not exist: {p}')
                self.context = ssl.create_default_context(cafile=str(p))
            else:
                self.context = ssl.create_default_context()

    def query(self, sql, timeout=None):
        headers = {
            'X-ClickHouse-User': self.username,
            'X-ClickHouse-Key': self.password,
            'Content-Type': 'text/plain; charset=utf-8',
        }
        req = urllib.request.Request(self.url + '?' + urllib.parse.urlencode({'wait_end_of_query': 1}),
                                     data=sql.encode('utf-8'), headers=headers, method='POST')
        try:
            with urllib.request.urlopen(req, timeout=timeout or self.timeout, context=self.context) as response:
                return response.read().decode('utf-8')
        except urllib.error.HTTPError as exc:
            msg = exc.read(2500).decode('utf-8', errors='replace').strip()
            raise MigrationError(f'{self.name}: HTTP {exc.code}: {msg}') from None
        except (urllib.error.URLError, TimeoutError, OSError) as exc:
            raise MigrationError(f'{self.name}: connection error: {exc}') from None

    def rows(self, sql):
        res = self.query(sql.rstrip().rstrip(';') + ' FORMAT JSONEachRow')
        return [json.loads(s) for s in res.splitlines() if s.strip()]

    def scalar(self, sql):
        res = self.query(sql.rstrip().rstrip(';') + ' FORMAT TabSeparated').strip()
        if '\n' in res or '\t' in res:
            raise MigrationError(f'{self.name}: expected one scalar, got multiple values')
        return int(res)


def get_metadata(client, dbs):
    predicate = ', '.join(lit(x) for x in dbs)
    rows = client.rows('SELECT database, name, engine, partition_key, sorting_key, '
                       'primary_key, sampling_key, engine_full, total_bytes, total_rows '
                       f'FROM system.tables WHERE database IN ({predicate}) ORDER BY database, name')
    return {(r['database'], r['name']): r for r in rows}


def get_columns(client, dbs):
    """Bulk compatibility check using stable system.columns fields.

    On ClickHouse 26.8, codec_expression and ttl_expression belong to
    DESCRIBE TABLE output, not to system.columns. The latter uses
    compression_codec; per-column TTL is checked by describe_columns
    immediately before each INSERT.
    """
    predicate = ', '.join(lit(x) for x in dbs)
    rows = client.rows('SELECT database, `table`, name, type, position, '
                       'default_kind, default_expression '
                       f'FROM system.columns WHERE database IN ({predicate}) '
                       'ORDER BY database, `table`, position')
    result = {}
    for r in rows:
        key = (r['database'], r['table'])
        result.setdefault(key, []).append(tuple(str(r.get(k) or '') for k in (
            'name', 'type', 'position', 'default_kind', 'default_expression')))
    return result


def describe_columns(client, key):
    """Compare column CODEC, TTL, comments and defaults before writing.

    DESCRIBE TABLE exposes codec_expression and ttl_expression on CH 26.8.
    This operation is read-only and only runs for an actual copy candidate.
    """
    db, table = key
    rows = client.rows(f'DESCRIBE TABLE {ident(db)}.{ident(table)}')
    fields = ('name', 'type', 'default_type', 'default_expression',
              'comment', 'codec_expression', 'ttl_expression')
    return [tuple(str(row.get(k) or '') for k in fields) for row in rows]


def validate_table(key, source, target, source_columns, target_columns):
    old, new = source[key], target.get(key)
    if not new:
        return 'target table does not exist'
    if new['engine'] != 'Replicated' + old['engine']:
        return f"target engine {new['engine']} is not Replicated{old['engine']}"
    for fld in ('partition_key', 'sorting_key', 'primary_key', 'sampling_key'):
        if (old.get(fld) or '') != (new.get(fld) or ''):
            return f'{fld} differs'
    if source_columns.get(key, []) != target_columns.get(key, []):
        return 'column definitions differ'
    return None


def is_unversioned_replacing(metadata):
    """Read exact engine arguments from system.tables.engine_full.

    ReplacingMergeTree without a `ver` parameter selects the winner using
    part/insertion order, which a cross-cluster INSERT SELECT cannot preserve.
    """
    if metadata.get('engine') != 'ReplacingMergeTree':
        return False
    # engine_full is a FULL engine expression, not just the engine name.
    # Example: "ReplacingMergeTree ORDER BY (`Подразделение`, `Дата`) SETTINGS ...".
    engine_full = str(metadata.get('engine_full') or '').strip()
    match = re.match(r'^ReplacingMergeTree\b', engine_full, flags=re.IGNORECASE)
    if not match:
        raise MigrationError('Cannot determine ReplacingMergeTree arguments from engine_full; refusing unsafe copy')
    tail = engine_full[match.end():].lstrip()
    if tail.startswith('('):
        # No-argument "ReplacingMergeTree()" is also unversioned.
        empty_args = re.match(r'^\(\s*\)', tail)
        if not empty_args:
            return False  # Has version or other explicit engine arguments.
        tail = tail[empty_args.end():].lstrip()
    if tail and not re.match(
        r'^(?:PARTITION\s+BY|ORDER\s+BY|PRIMARY\s+KEY|SAMPLE\s+BY|TTL\b|SETTINGS\b|COMMENT\b)',
        tail, flags=re.IGNORECASE
    ):
        raise MigrationError('Unexpected engine_full syntax after ReplacingMergeTree; refusing unsafe copy')
    return True


def read_journal(path):
    if not path.exists():
        return {'format_version': 1, 'tables': {}}
    journal = json.loads(path.read_text(encoding='utf-8'))
    if journal.get('format_version') != 1:
        raise MigrationError('Unsupported journal version; refusing to continue')
    return journal


def save_journal(path, journal):
    path.parent.mkdir(parents=True, exist_ok=True)
    tmp = path.with_name(path.name + '.tmp')
    fd = os.open(str(tmp), os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, 'w', encoding='utf-8') as file:
        json.dump(journal, file, ensure_ascii=False, indent=2, sort_keys=True)
        file.write('\n')
        file.flush()
        os.fsync(file.fileno())
    os.replace(str(tmp), str(path))


def rowcount(client, key):
    return client.scalar(f'SELECT count() FROM {ident(key[0])}.{ident(key[1])}')


def row_fingerprint(client, key, replacing):
    """Order-independent two-hash fingerprint of full rows.

    For ReplacingMergeTree, FINAL is essential: physical parts can be merged
    independently on source and target, making ordinary counts unreliable.
    This is a strong probabilistic check, not a cryptographic proof.
    """
    table = f'{ident(key[0])}.{ident(key[1])}'
    final = ' FINAL' if replacing else ''
    sql = ('SELECT count() AS rows, '
           'toString(sum(cityHash64(*))) AS checksum_sum, '
           'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
           f'FROM {table}{final}')
    results = client.rows(sql)
    if len(results) != 1:
        raise MigrationError(f'{client.name}: fingerprint expected one row for {key}')
    r = results[0]
    return {'rows': int(r['rows']),
            'checksum_sum': str(r['checksum_sum']),
            'checksum_xor': str(r['checksum_xor'])}


def replica_fingerprints(target, cluster, key, replacing):
    """Run the complete fingerprint query locally on each replica.

    Passing view(SELECT ...) to clusterAllReplicas is intentional: FINAL and
    aggregates execute inside each replica, not over the concatenated data
    from all hosts. Produces one row per host including an empty table.
    """
    final = ' FINAL' if replacing else ''
    inner = (f'SELECT hostName() AS host, count() AS rows, '
             f'toString(sum(cityHash64(*))) AS checksum_sum, '
             f'toString(groupBitXor(sipHash64(*))) AS checksum_xor '
             f'FROM {ident(key[0])}.{ident(key[1])}{final}')
    sql = (f'SELECT host, rows, checksum_sum, checksum_xor '
           f'FROM clusterAllReplicas({lit(cluster)}, view({inner})) ORDER BY host')
    return [{
        'host': str(r['host']), 'rows': int(r['rows']),
        'checksum_sum': str(r['checksum_sum']),
        'checksum_xor': str(r['checksum_xor']),
    } for r in target.rows(sql)]


def fingerprint_equal(a, b):
    return all(str(a.get(k)) == str(b.get(k)) for k in
               ('rows', 'checksum_sum', 'checksum_xor'))


def verify_fingerprint_on_replicas(source, target, cluster, key, engine,
                                   expected, wait_secs, baseline=None):
    """Read-only validation of source, local target and both remote replicas.

    No retries of an INSERT are ever done here. Raise rather than mark DONE
    if the source moved or any host does not match.
    """
    replacing = engine == 'ReplacingMergeTree'
    src = row_fingerprint(source, key, replacing)
    if baseline is not None and not fingerprint_equal(src, baseline):
        raise MigrationError(f'{key[0]}.{key[1]}: source fingerprint changed during copy')
    dst = row_fingerprint(target, key, replacing)
    if not fingerprint_equal(src, dst):
        raise MigrationError(f'{key[0]}.{key[1]}: source and target fingerprints differ: '
                             f'source={src}, target={dst}')
    verified_replicas(target, cluster, key, expected, None, wait_secs)
    status = replica_status(target, cluster, key)
    wanted_hosts = {str(row['host']) for row in status}
    fps = replica_fingerprints(target, cluster, key, replacing)
    observed_hosts = {row['host'] for row in fps}
    if len(fps) != expected or wanted_hosts != observed_hosts:
        raise MigrationError(f'{key[0]}.{key[1]}: replica fingerprint hosts mismatch: '
                             f'expected={sorted(wanted_hosts)}, seen={sorted(observed_hosts)}')
    for r in fps:
        if not fingerprint_equal(src, r):
            raise MigrationError(f'{key[0]}.{key[1]}: replica {r["host"]} does not '
                                 f'match source: expected={src}, actual={r}')
    return src, fps


def replica_status(target, cluster, key):
    db, table = key
    return target.rows(
        'SELECT hostName() AS host, replica_name, is_readonly, is_session_expired, '
        'total_replicas, active_replicas, queue_size, inserts_in_queue, log_max_index, log_pointer '
        f'FROM clusterAllReplicas({lit(cluster)}, system.replicas) '
        f'WHERE database = {lit(db)} AND table = {lit(table)} ORDER BY host')


def replica_counts(target, cluster, key):
    """Get a count for EVERY target host, including empty tables.

    GROUP BY hostName() over a table with no rows returns no groups at all,
    not one row containing zero for each replica. Anchor the result with
    system.one queried on every host, then union actual table counts. The
    sum remains zero for empty tables, and is the true count otherwise.
    """
    db, table = key
    return target.rows(
        'SELECT host, sum(row_count) AS rows FROM ('
        'SELECT hostName() AS host, toUInt64(0) AS row_count '
        f'FROM clusterAllReplicas({lit(cluster)}, system.one) '
        'UNION ALL '
        'SELECT hostName() AS host, count() AS row_count '
        f'FROM clusterAllReplicas({lit(cluster)}, {lit(db)}, {lit(table)}) '
        'GROUP BY host'
        ') GROUP BY host ORDER BY host')


def summarize_replica_states(status, counts):
    """Human-readable diagnostics; no credentials or source data returned."""
    by_host = {str(r.get('host', '')): r.get('rows') for r in counts}
    return [{
        'host': r.get('host'),
        'total_replicas': r.get('total_replicas'),
        'active_replicas': r.get('active_replicas'),
        'is_readonly': r.get('is_readonly'),
        'is_session_expired': r.get('is_session_expired'),
        'queue_size': r.get('queue_size'),
        'inserts_in_queue': r.get('inserts_in_queue'),
        'log_pointer': r.get('log_pointer'),
        'log_max_index': r.get('log_max_index'),
        'rows': by_host.get(str(r.get('host', ''))),
    } for r in status]


def verified_replicas(target, cluster, key, expected, wanted_rows, wait_secs):
    """Verify both replica sessions and physical row counts.

    log_pointer is intentionally NOT a hard gate. It represents the next
    Keeper log entry to queue, not the last part applied; a small transient
    difference against log_max_index is not proof of missing data. Counts on
    both replicas, together with an empty queue and active sessions, are the
    actual success criteria. The indices are still reported for diagnosis.
    """
    deadline = time.monotonic() + wait_secs
    last_status, last_counts = [], []
    while True:
        last_status = replica_status(target, cluster, key)
        healthy = len(last_status) == expected and all(
            int(s['is_readonly']) == 0 and int(s['is_session_expired']) == 0 and
            int(s['total_replicas']) == expected and int(s['active_replicas']) == expected and
            int(s['queue_size']) == 0 and int(s['inserts_in_queue']) == 0
            for s in last_status)
        # Fetch count even if health flags fail: essential for explaining
        # what blocked verification, and for distinguishing lag from data loss.
        last_counts = replica_counts(target, cluster, key)
        status_hosts = {str(s.get('host')) for s in last_status}
        count_hosts = {str(r.get('host')) for r in last_counts}
        if (healthy and len(status_hosts) == expected and
                len(last_counts) == expected and count_hosts == status_hosts and
                (wanted_rows is None or all(int(r['rows']) == wanted_rows for r in last_counts))):
            return last_counts
        if time.monotonic() >= deadline:
            brief = summarize_replica_states(last_status, last_counts)
            raise MigrationError(
                f'replica verification timed out for {key[0]}.{key[1]} '
                f'(expected {expected} replicas each containing {wanted_rows} physical rows): '
                f'{json.dumps(brief, ensure_ascii=False)}')
        time.sleep(5)


def inspect_table(source, target, cluster, key, journal):
    """Read-only snapshot for troubleshooting and recovery."""
    name = '.'.join(key)
    out = {
        'table': name,
        'journal_status': journal.get('tables', {}).get(name, {}).get('status', 'PENDING'),
        'source_rows': rowcount(source, key),
        'target_local_rows': rowcount(target, key),
        'replicas': summarize_replica_states(
            replica_status(target, cluster, key),
            replica_counts(target, cluster, key)),
    }
    print(json.dumps(out, ensure_ascii=False, indent=2))
    return out


def main():
    parser = argparse.ArgumentParser(description='Guarded server-side ClickHouse data copy via remoteSecure')
    parser.add_argument('--config', default='config.json')
    parser.add_argument('--only', help='exact database.table (special names supported)')
    parser.add_argument('--apply', action='store_true', help='actually insert rows into target')
    parser.add_argument('--ack-source-frozen', action='store_true', help='confirm source writes are paused and remain paused')
    parser.add_argument('--resume', action='store_true', help='skip journal entries marked DONE; never retry uncertain copies')
    parser.add_argument('--skip-uncertain', action='store_true', help='with --resume: omit UNCERTAIN/COPYING/VERIFYING tables without modifying their journal records')
    parser.add_argument('--skip-no-version-duplicates', action='store_true', help='skip unversioned ReplacingMergeTree tables with duplicate ORDER BY keys (physical > FINAL rows)')
    parser.add_argument('--limit', type=int, help='maximum number of pending tables to process')
    parser.add_argument('--verify', action='store_true', help='verify previously completed tables (no new inserts)')
    parser.add_argument('--inspect', action='store_true', help='read-only counts, replica status, and journal for selected tables')
    parser.add_argument('--reconcile', action='store_true', help='compare physical/logical full-row fingerprints, without INSERT')
    parser.add_argument('--mark-done', action='store_true', help='with --reconcile and --ack-source-frozen: mark a verified UNCERTAIN copy DONE')
    parser.add_argument('--output', default='./data_transfer_progress')
    parser.add_argument('--replica-wait-seconds', type=int, default=300)
    args = parser.parse_args()
    if args.mark_done and (not args.reconcile or not args.ack_source_frozen or not args.only):
        raise MigrationError('--mark-done requires --reconcile --only db.table --ack-source-frozen')
    if args.skip_uncertain and not args.resume:
        raise MigrationError('--skip-uncertain requires --resume')
    if args.reconcile and args.apply:
        raise MigrationError('--reconcile is read-only and cannot be used with --apply')
    config_path = Path(args.config).expanduser().resolve()
    config = json.loads(config_path.read_text(encoding='utf-8'))
    dbs = config['databases']
    source = Client(config['source'], config_path.parent, 'source')
    target = Client(config['target'], config_path.parent, 'target')
    secrets = [source.password, target.password]
    source_native_port = int(config.get('source_native_tls_port', 9440))
    cluster = config.get('target_cluster', 'default')
    replicas = int(config.get('expected_replicas', 2))
    base = Path(args.output).expanduser().resolve()
    journal_path = base / 'progress.json'
    journal = read_journal(journal_path)
    origin = f'{source.host}:{source_native_port}'
    if journal.get('source') and journal['source'] != origin:
        raise MigrationError('Journal belongs to another source; use a separate --output')
    if journal.get('target') and journal['target'] != target.host:
        raise MigrationError('Journal belongs to another target; use a separate --output')
    journal['source'] = origin
    journal['target'] = target.host
    print('Source:', origin, '| Target:', target.host, '| Cluster:', cluster)
    print('Databases:', ', '.join(dbs))
    source_meta, target_meta = get_metadata(source, dbs), get_metadata(target, dbs)
    src_cols, dst_cols = get_columns(source, dbs), get_columns(target, dbs)
    supported = {'MergeTree', 'ReplacingMergeTree', 'SummingMergeTree',
                 'AggregatingMergeTree', 'CollapsingMergeTree',
                 'VersionedCollapsingMergeTree', 'GraphiteMergeTree'}
    keys = [key for key, row in source_meta.items() if row['engine'] in supported]
    if args.only:
        keys = [key for key in keys if '.'.join(key) == args.only]
        if not keys:
            raise MigrationError('No supported source table found for --only ' + args.only)
    # Start with the smallest tables; helps pilot and isolates failures.
    keys.sort(key=lambda k: (int(source_meta[k].get('total_bytes') or 0), k))
    problems = [(key, validate_table(key, source_meta, target_meta, src_cols, dst_cols)) for key in keys]
    problems = [(key, detail) for key, detail in problems if detail]
    if problems:
        for key, detail in problems[:20]:
            print('BLOCKED', '.'.join(key), detail)
        print('Incompatible / missing tables:', len(problems))
        if args.only:
            raise MigrationError('Selected table has incompatible schema')
    invalid = {k for k, _ in problems}
    keys = [key for key in keys if key not in invalid]
    if args.resume and not args.verify and not args.inspect:
        keys = [key for key in keys if journal['tables'].get('.'.join(key), {}).get('status') != 'DONE']
    skipped_uncertain = 0
    if args.resume and args.skip_uncertain and args.apply:
        ambiguous = {'COPYING', 'VERIFYING', 'UNCERTAIN'}
        suspended = [key for key in keys if journal['tables'].get('.'.join(key), {}).get('status') in ambiguous]
        skipped_uncertain = len(suspended)
        for key in suspended:
            print('SKIP UNCERTAIN', '.'.join(key), '— no write or journal change')
        keys = [key for key in keys if key not in suspended]
    if args.limit is not None:
        if args.limit < 1:
            raise MigrationError('--limit must be >= 1')
        keys = keys[:args.limit]
    print(f'Compatible tables selected: {len(keys)}')
    if args.inspect:
        for key in keys:
            inspect_table(source, target, cluster, key, journal)
        return 0
    if args.reconcile:
        for key in keys:
            name = '.'.join(key)
            prior = journal['tables'].get(name, {}).get('status', 'PENDING')
            if args.mark_done and prior not in ('UNCERTAIN', 'VERIFYING'):
                raise MigrationError(f'{name}: --mark-done only permitted for UNCERTAIN/VERIFYING, got {prior}')
            src_fp, fingerprints = verify_fingerprint_on_replicas(
                source, target, cluster, key, source_meta[key]['engine'],
                replicas, args.replica_wait_seconds)
            print('RECONCILED', name, json.dumps({'source':src_fp, 'replicas':fingerprints},
                                                ensure_ascii=False), flush=True)
            if args.mark_done:
                record = journal['tables'].get(name, {})
                record['status'] = 'DONE'
                record['verified_fingerprint'] = src_fp
                record['replicas'] = fingerprints
                record['reconciled_at'] = now()
                record.pop('error', None)
                journal['tables'][name] = record
                save_journal(journal_path, journal)
                print('MARKED DONE', name, 'without re-inserting any rows', flush=True)
        return 0
    if args.verify:
        for key in keys:
            name = '.'.join(key)
            record = journal['tables'].get(name, {})
            if record.get('status') != 'DONE':
                continue
            fp, _ = verify_fingerprint_on_replicas(
                source, target, cluster, key, source_meta[key]['engine'],
                replicas, args.replica_wait_seconds)
            print('VERIFIED', name, json.dumps(fp), flush=True)
        return 0
    if not args.apply:
        for key in keys:
            r = source_meta[key]
            prior = journal['tables'].get('.'.join(key), {}).get('status', 'PENDING')
            print(f'PLAN {key[0]}.{key[1]} ({r["engine"]}): '
                  f'{r.get("total_rows") or 0} approx rows, {r.get("total_bytes") or 0} compressed bytes; {prior}')
        print('DRY RUN COMPLETE; nothing changed.')
        return 0
    if not args.ack_source_frozen:
        raise MigrationError('Data migration requires --ack-source-frozen: pause ALL source writes and background ingestion')
    if problems and not args.only:
        print('WARN: blocked tables were skipped; migration cannot be called complete.')
    skipped_unsafe = 0
    for index, key in enumerate(keys, 1):
        name = '.'.join(key)
        record = journal['tables'].get(name, {})
        if record.get('status') == 'DONE':
            if args.resume:
                print(f'[{index}/{len(keys)}] SKIP DONE {name}')
                continue
            raise MigrationError(f'{name} already DONE; add --resume to skip it')
        if record.get('status') in ('COPYING', 'VERIFYING', 'UNCERTAIN'):
            raise MigrationError(f'{name} has ambiguous prior state {record["status"]}; '
                                 f'do not retry automatically. Check row counts and repair manually.')
        # DESCRIBE includes CODEC and column TTL; system.columns does not.
        # Compare immediately before transfer so unsupported expressions or
        # mismatched per-column settings cannot be silently ignored.
        if describe_columns(source, key) != describe_columns(target, key):
            raise MigrationError(f'{name}: source and target DESCRIBE TABLE differ '
                                 '(column definitions / codec / TTL / comments). '
                                 'No data was copied for this table.')
        old_rows = rowcount(source, key)
        replacing = source_meta[key]['engine'] == 'ReplacingMergeTree'
        baseline_fp = row_fingerprint(source, key, replacing)
        # With ReplacingMergeTree without a version argument, INSERT SELECT
        # can change the winning row for duplicate ORDER BY keys even when
        # count() FINAL is unchanged. Never silently accept such copies.
        if (replacing and is_unversioned_replacing(source_meta[key])
                and old_rows != baseline_fp['rows']):
            message = (f'{name}: ReplacingMergeTree without ver has {old_rows} physical '
                       f'vs {baseline_fp["rows"]} FINAL rows; normal INSERT SELECT can '
                       'change chosen winners. Requires a separately reviewed FINAL-snapshot '
                       'or part-preserving migration. No data copied.')
            if args.skip_no_version_duplicates:
                skipped_unsafe += 1
                print('SKIP UNSAFE', message, flush=True)
                continue
            raise MigrationError(message + ' Use --skip-no-version-duplicates to continue other tables.')
        new_rows = rowcount(target, key)
        if new_rows != 0:
            raise MigrationError(f'{name}: target already has {new_rows} rows. Refusing to duplicate data.')
        # Refuse if the target has any rows on either replica. A local
        # count() against the first host alone cannot detect stale data on
        # the second host. Never attempt an INSERT in that situation.
        before_counts = replica_counts(target, cluster, key)
        if len(before_counts) != replicas or any(int(r['rows']) != 0 for r in before_counts):
            raise MigrationError(f'{name}: target must be empty on ALL {replicas} replicas; '
                                 f'observed: {json.dumps(before_counts, ensure_ascii=False)}. '
                                 'Use --inspect and reconcile; no data copied.')
        # Check Keeper health BEFORE inserting.
        verified_replicas(target, cluster, key, replicas, 0, args.replica_wait_seconds)
        if not old_rows:
            record = {'status': 'DONE', 'source_rows_before': 0, 'source_rows_after': 0,
                      'target_rows': 0, 'verified_fingerprint': baseline_fp, 'completed_at': now()}
            journal['tables'][name] = record
            save_journal(journal_path, journal)
            print(f'[{index}/{len(keys)}] EMPTY {name}')
            continue
        record = {'status': 'COPYING', 'source_rows_before': old_rows,
                  'source_fingerprint_before': baseline_fp, 'started_at': now()}
        journal['tables'][name] = record
        save_journal(journal_path, journal)
        sql = (f'INSERT INTO {ident(key[0])}.{ident(key[1])} SELECT * FROM '
               f'remoteSecure({lit(origin)}, {lit(key[0])}, {lit(key[1])}, '
               f'{lit(source.username)}, {lit(source.password)})')
        print(f'[{index}/{len(keys)}] COPYING {name}: {old_rows} rows; '
              f'INSERT runs on target; no SQL passwords are printed', flush=True)
        try:
            target.query(sql)
            record['status'] = 'VERIFYING'
            record['insert_returned_at'] = now()
            save_journal(journal_path, journal)
            end_source = rowcount(source, key)
            end_target = rowcount(target, key)
            record['source_rows_after'] = end_source
            record['target_rows'] = end_target
            if not replacing and (old_rows != end_source or old_rows != end_target):
                raise MigrationError(f'{name}: non-ReplacingMergeTree physical row count mismatch: '
                                     f'source before={old_rows}, source after={end_source}, '
                                     f'target={end_target}')
            verified_fp, rep_fps = verify_fingerprint_on_replicas(
                source, target, cluster, key, source_meta[key]['engine'],
                replicas, args.replica_wait_seconds, baseline=baseline_fp)
            record['status'] = 'DONE'
            record['verified_fingerprint'] = verified_fp
            record['replicas'] = rep_fps
            record['completed_at'] = now()
            save_journal(journal_path, journal)
            print(f'[{index}/{len(keys)}] DONE {name}: {verified_fp["rows"]} verified logical rows on both replicas (target physically {end_target})', flush=True)
        except Exception as exc:
            record['status'] = 'UNCERTAIN'
            record['error'] = safe_error(exc, secrets)
            record['failed_at'] = now()
            save_journal(journal_path, journal)
            raise MigrationError(f'{name}: {record["error"]}. State UNCERTAIN: NEVER retry '
                                 'without reconciling the target table.') from None
    print('FINISHED selected tables. Check pending/blocked tables and rerun with --resume if needed.')
    if skipped_unsafe or skipped_uncertain or problems:
        print(f'INCOMPLETE: skipped_unsafe={skipped_unsafe}, skipped_uncertain={skipped_uncertain}, '
              f'missing_or_incompatible={len(problems)}.')
        print('No skipped table was marked DONE. Resolve every skipped table before cutover.')
    return 0


if __name__ == '__main__':
    try:
        sys.exit(main())
    except (MigrationError, KeyError, ValueError, OSError) as exc:
        # If secrets are known they were masked at the point of persistence; errors
        # outside migration SQL should not include credentials.
        print('ERROR:', str(exc), file=sys.stderr)
        sys.exit(1)
