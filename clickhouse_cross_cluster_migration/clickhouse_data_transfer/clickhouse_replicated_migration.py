#!/usr/bin/env python3
"""Generate a reviewable migration plan from a restored ClickHouse database.

Source: unreplicated database (typically restored from BACKUP).
Target: existing, EMPTY Replicated database on the SAME ClickHouse cluster.
MySQL engines are retained as external tables, with password placeholders.

Generates SQL; never executes CREATE, ATTACH, INSERT, or DROP statements.
Requires Python 3.10+. Uses only the standard library.
"""

import argparse
import base64
import csv
import json
import os
from pathlib import Path
import re
import ssl
import sys
import urllib.error
import urllib.request
from collections import Counter


def ident(value: str) -> str:
    return '`' + value.replace('`', '``') + '`'


def sql_string(value: str) -> str:
    return "'" + value.replace('\\', '\\\\').replace("'", "\\'") + "'"


def qualified(db: str, table: str) -> str:
    return ident(db) + '.' + ident(table)


class CH:
    def __init__(self, url: str, username: str, password: str, ca_file: str | None):
        self.url = url.rstrip('/') + '/'
        cred = base64.b64encode((username + ':' + password).encode()).decode('ascii')
        self.auth = 'Basic ' + cred
        self.ssl_context = ssl.create_default_context(cafile=ca_file)

    def rows(self, sql: str) -> list[dict]:
        query = sql.rstrip().rstrip(';') + ' FORMAT JSON'
        req = urllib.request.Request(
            self.url,
            data=query.encode('utf-8'),
            headers={'Authorization': self.auth, 'Content-Type': 'text/plain; charset=utf-8'},
            method='POST',
        )
        try:
            with urllib.request.urlopen(req, context=self.ssl_context, timeout=180) as reply:
                doc = json.load(reply)
        except urllib.error.HTTPError as exc:
            body = exc.read().decode('utf-8', errors='replace')
            raise RuntimeError(f'ClickHouse HTTP {exc.code}: {body[:1500]}') from exc
        return doc['data']


IDENTIFIER = r'(?:`(?:``|[^`])*`|[^\s.(]+)'
HEADER = re.compile(
    rf'^(\s*CREATE\s+(?:TABLE|VIEW)\s+(?:IF\s+NOT\s+EXISTS\s+)?)'
    rf'((?:{IDENTIFIER}\.)?{IDENTIFIER})',
    flags=re.I | re.S,
)
ENGINE = re.compile(r'\bENGINE\s*=\s*(ReplacingMergeTree|MergeTree)\b', re.I)


def log_to_replicated(ddl: str, src_db: str, name: str, dst_db: str) -> str:
    """Convert Log to ReplicatedMergeTree, retaining columns and table COMMENT.

    Plain Log has no ORDER BY or partitioning. We preserve any trailing COMMENT
    while refusing unexpected clauses rather than generating corrupt SQL.
    """
    ddl = retarget_header(ddl, src_db, name, dst_db)
    match = re.search(r'\bENGINE\s*=\s*Log\b(?:\s*\(\s*\))?', ddl, flags=re.I)
    if not match:
        raise ValueError(f'Expected Log engine for {name}')
    suffix = ddl[match.end():].strip()
    if suffix and not re.match(r'^COMMENT\s+\'', suffix, flags=re.I):
        raise ValueError(f'Unexpected Log engine suffix for {name}; review manually')
    return (ddl[:match.start()].rstrip() +
            ' ENGINE = ReplicatedMergeTree() ORDER BY tuple()' +
            (' ' + suffix if suffix else ''))


def mysql_preserve_engine(ddl: str, src_db: str, name: str, dst_db: str) -> str:
    """Retarget the original MySQL CREATE, replacing credentials with a placeholder.

    Keeps the remote host, database, table, user, columns, and MySQL ENGINE.
    The credential is intentionally NOT carried to the generated SQL file.
    """
    ddl = retarget_header(ddl, src_db, name, dst_db)
    match = re.search(r'\bENGINE\s*=\s*MySQL\s*\(', ddl, flags=re.I)
    if not match:
        raise ValueError(f'Expected MySQL engine for {name}')
    opening = match.end() - 1
    pos = opening + 1
    args = []
    arg_start = pos
    depth = 1
    quote = None
    while pos < len(ddl) and depth:
        ch = ddl[pos]
        if quote:
            if ch == '\\' and pos + 1 < len(ddl):
                pos += 2
                continue
            if ch == quote:
                if pos + 1 < len(ddl) and ddl[pos + 1] == quote:
                    pos += 2
                    continue
                quote = None
        else:
            if ch in ("'", '"', '`'):
                quote = ch
            elif ch == '(':
                depth += 1
            elif ch == ')':
                depth -= 1
                if depth == 0:
                    args.append(ddl[arg_start:pos].strip())
                    break
            elif ch == ',' and depth == 1:
                args.append(ddl[arg_start:pos].strip())
                arg_start = pos + 1
        pos += 1
    if depth:
        raise ValueError(f'Unbalanced MySQL ENGINE argument list for {name}')
    if len(args) < 5:
        # Do not accidentally export credentials in non-standard / named-collection
        # MySQL DDL. This variant needs manual review instead of a blind copy.
        raise ValueError(f'Non-positional MySQL engine arguments for {name}; review manually')
    args[4] = sql_string('__REPLACE_MYSQL_PASSWORD__')
    return ddl[:opening + 1] + ', '.join(args) + ddl[pos:]

def retarget_header(ddl: str, src_db: str, name: str, dest_db: str) -> str:
    m = HEADER.match(ddl)
    if not m:
        raise ValueError(f'Unable to parse CREATE header for {name}: {ddl[:150]!r}')
    old = m.group(2)
    # The canonical statement should refer to the exact source table.
    identifiers = re.fullmatch(rf'(?:(?P<db>{IDENTIFIER})\.)?(?P<leaf>{IDENTIFIER})', old)
    if not identifiers:
        raise ValueError(f'Cannot identify CREATE table name: {old!r}')
    leaf = identifiers.group('leaf')
    old_leaf = leaf[1:-1].replace('``', '`') if leaf.startswith('`') else leaf
    if old_leaf != name:
        raise ValueError(f'Unexpected table identifier for {name}: {old!r}')
    new_ddl = ddl[:m.start(2)] + qualified(dest_db, name) + ddl[m.end(2):]
    # Restoring the source database retains source UUID. New table needs a NEW UUID.
    new_ddl, _ = re.subn(
        r"\s+UUID\s+'[0-9a-fA-F-]{36}'(?=\s|\()", '', new_ddl, count=1, flags=re.I
    )
    return new_ddl.rstrip().rstrip(';')


def new_engine(ddl: str, expected_engine: str) -> str:
    m = ENGINE.search(ddl)
    if not m or m.group(1).lower() != expected_engine.lower():
        raise ValueError(f'ENGINE mismatch; expected {expected_engine}: {ddl[:300]!r}')
    # Explicit Keeper parameters are necessary for ReplacingMergeTree(version[, is_deleted]);
    # replacing only engine name could mistakenly parse version as Keeper path.
    original_params = ''
    end = m.end()
    rest = ddl[end:]
    opening = re.match(r'\s*\(', rest)
    if opening:
        pos = end + opening.end()  # first char after '('
        start = pos
        depth = 1
        quote = ''
        while pos < len(ddl) and depth:
            ch = ddl[pos]
            if quote:
                if ch == quote:
                    if pos + 1 < len(ddl) and ddl[pos + 1] == quote:
                        pos += 2
                        continue
                    quote = ''
                elif ch == '\\' and pos + 1 < len(ddl):
                    pos += 2
                    continue
            else:
                if ch in ("'", '"', '`'):
                    quote = ch
                elif ch == '(':
                    depth += 1
                elif ch == ')':
                    depth -= 1
            pos += 1
        if depth:
            raise ValueError('Unbalanced engine argument list')
        original_params = ddl[start:pos - 1].strip()
        end = pos
    # A database with ENGINE = Replicated uses server-side default replica path/name.
    # Supplying explicit paths may be rejected on ClickHouse 26.8+.
    rep_engine = 'Replicated' + expected_engine + '(' + original_params + ')'
    return ddl[:m.start()] + 'ENGINE = ' + rep_engine + ddl[end:]


def rewrite_views(ddl: str, src_db: str, name: str, dest_db: str) -> str:
    ddl = retarget_header(ddl, src_db, name, dest_db)
    # Replace DB-qualified identifiers in the SELECT body; inspect generated views
    # because string literals/explicit references to other DBs may need manual repair.
    for prefix in (ident(src_db) + '.', src_db + '.'):
        ddl = ddl.replace(prefix, ident(dest_db) + '.')
    return ddl


def write_sql(path: Path, header: str, statements: list[str]) -> None:
    with path.open('w', encoding='utf-8') as f:
        f.write('-- ' + header + '\n')
        f.write('-- Review and execute only when the previous stage succeeded.\n\n')
        for stmt in statements:
            f.write(stmt.rstrip().rstrip(';') + ';\n\n')


def run(args):
    if args.source == args.target:
        raise ValueError('Source and target databases must differ')
    ch = CH(args.url, args.user, os.environ.get('CH_PASSWORD', ''), args.ca_file)
    dbs = {r['name']: r['engine'] for r in ch.rows(
        f'SELECT name, engine FROM system.databases '
        f'WHERE name IN ({sql_string(args.source)}, {sql_string(args.target)})'
    )}
    if args.source not in dbs:
        raise RuntimeError(f'Source database {args.source!r} does not exist')
    if dbs.get(args.target) != 'Replicated':
        raise RuntimeError(
            f'Target {args.target!r} must exist with ENGINE = Replicated; '
            f'actual engine is {dbs.get(args.target)!r}. Stop: do NOT DROP any database automatically.'
        )

    target_tables = ch.rows(
        f'SELECT name, engine FROM system.tables WHERE database = {sql_string(args.target)}'
    )
    if target_tables:
        raise RuntimeError(
            f'Target is not empty ({len(target_tables)} tables/views). '
            f'This generator requires an empty destination to avoid duplicate ATTACH/INSERT. '
            f'Existing objects: {[r["name"] for r in target_tables[:10]]}'
        )

    source_tables = ch.rows(
        f'SELECT name, engine, create_table_query FROM system.tables '
        f'WHERE database = {sql_string(args.source)} ORDER BY name'
    )
    if not source_tables:
        raise RuntimeError('Source has no objects')

    if args.csv:
        with open(args.csv, encoding='utf-8-sig', newline='') as f:
            expected = {r['name']: r['engine'] for r in csv.DictReader(f)}
        actual = {r['name']: r['engine'] for r in source_tables}
        if expected != actual:
            missing = sorted(set(expected) - set(actual))[:10]
            added = sorted(set(actual) - set(expected))[:10]
            changed = [(k, expected[k], actual[k]) for k in expected.keys() & actual.keys()
                       if expected[k] != actual[k]][:10]
            raise RuntimeError(f'CSV/source disagree. Missing: {missing}, added: {added}, changed: {changed}')

    partitions = ch.rows(
        f'SELECT table, partition_id FROM system.parts '
        f'WHERE database = {sql_string(args.source)} AND active = 1 '
        'GROUP BY table, partition_id ORDER BY table, partition_id'
    )
    by_table = {}
    for rec in partitions:
        by_table.setdefault(rec['table'], []).append(rec['partition_id'])

    create_merge, attach, create_log, copy_log = [], [], [], []
    create_mysql_tables, views, notes = [], [], []
    for rec in source_tables:
        name, engine, ddl = rec['name'], rec['engine'], rec['create_table_query']
        src, dst = qualified(args.source, name), qualified(args.target, name)
        if engine in ('MergeTree', 'ReplacingMergeTree'):
            updated = retarget_header(ddl, args.source, name, args.target)
            create_merge.append(new_engine(updated, engine))
            for part_id in by_table.get(name, []):
                attach.append(f'ALTER TABLE {dst} ATTACH PARTITION ID {sql_string(part_id)} FROM {src}')
        elif engine == 'Log':
            create_log.append(log_to_replicated(ddl, args.source, name, args.target))
            copy_log.append(f'INSERT INTO {dst} SELECT * FROM {src}')
        elif engine == 'MySQL':
            # Preserve MySQL external engine (no local snapshot / INSERT).
            create_mysql_tables.append(mysql_preserve_engine(ddl, args.source, name, args.target))
            notes.append(f'{name}: MySQL external engine preserved; manually provide its password, and test network access on every replica')
        elif engine == 'View':
            views.append(rewrite_views(ddl, args.source, name, args.target))
        else:
            notes.append(f'{name}: unsupported engine {engine}; requires manual treatment')

    out = Path(args.out).expanduser().resolve()
    out.mkdir(parents=True, exist_ok=True)
    write_sql(out / '01_create_replicated_tables.sql', 'Create replicated table definitions for MergeTree-family tables.', create_merge)
    write_sql(out / '02_attach_partitions.sql', 'Attach source partitions ONCE on a single replica, no ON CLUSTER. NOT retry-safe.', attach)
    write_sql(out / '03_create_log_tables.sql', 'Convert Log tables into ReplicatedMergeTree with ORDER BY tuple().', create_log)
    write_sql(out / '04_copy_log_data.sql', 'Copy Log data ONCE. Repeating creates duplicate records.', copy_log)
    write_sql(out / '05_create_mysql_tables.sql', 'Keep remote MySQL ENGINE; replace password placeholders before execution.', create_mysql_tables)
    write_sql(out / '06_create_views.sql', 'Create views AFTER all referenced tables/views exist; inspect DB references and dependency order.', views)
    checks = [
        f'SELECT name, engine FROM system.databases WHERE name IN ({sql_string(args.source)}, {sql_string(args.target)})',
        f'SELECT database, engine, count() AS objects FROM system.tables WHERE database IN '
        f'({sql_string(args.source)}, {sql_string(args.target)}) GROUP BY database, engine ORDER BY database, engine',
        f'SELECT database, table, sum(rows) AS active_part_rows, formatReadableSize(sum(bytes_on_disk)) AS size '
        f'FROM system.parts WHERE database IN ({sql_string(args.source)}, {sql_string(args.target)}) '
        f'AND active GROUP BY database, table ORDER BY table, database',
        f'SELECT database, table, is_readonly, is_session_expired, queue_size, absolute_delay '
        f'FROM system.replicas WHERE database = {sql_string(args.target)} ORDER BY table',
    ]
    write_sql(out / '07_validate.sql', 'Check data and replication. Part counts can vary for ReplacingMergeTree due to merges.', checks)

    counts = Counter(r['engine'] for r in source_tables)
    summary = {
        'source_database': args.source,
        'target_database': args.target,
        'database_engines': dbs,
        'objects': dict(counts),
        'merge_table_definitions': len(create_merge),
        'partition_attach_statements': len(attach),
        'log_tables': len(create_log),
        'mysql_external_tables': len(create_mysql_tables),
        'views': len(views),
        'source_data_kept': True,
        'ddl_executed': False,
        'manual_checks': notes + [
            'Pause writes to source during migration or perform a controlled final delta/cutover.',
            'Only a single host/shard is inspected. On multishard clusters, migrate every shard separately.',
            'Review generated view dependencies and source database references.',
            'Never execute generated MySQL DDL with __REPLACE_MYSQL_PASSWORD__ unresolved; password is not copied.',
            'Do not rerun ATTACH PARTITION or INSERT after uncertain outcomes; reconcile counts/parts first.',
            'Check TTL, projections, quotas, Keeper access, storage space, and replica health.',
        ],
    }
    (out / 'manifest.json').write_text(json.dumps(summary, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    print(f'Generated {out} (SQL not executed)')
    print(json.dumps({k: v for k, v in summary.items() if k not in ('manual_checks', 'database_engines')}, ensure_ascii=False, indent=2))
    print('Review all generated statements before running; read manifest.json.')


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--url', required=True, help='ClickHouse HTTPS URL, e.g. https://ch-host:8443')
    ap.add_argument('--user', default='admin', help='ClickHouse login; set password using CH_PASSWORD env variable')
    ap.add_argument('--ca-file', help='Custom Yandex Cloud CA file, if needed; TLS validation remains enabled')
    ap.add_argument('--source', default='zendendb_import')
    ap.add_argument('--target', default='zendendb')
    ap.add_argument('--csv', help='Optional CSV with name,engine to verify the live inventory matches the attachment')
    ap.add_argument('--out', default='./clickhouse-migration-sql')
    args = ap.parse_args()
    try:
        run(args)
    except (ValueError, RuntimeError, urllib.error.URLError) as exc:
        print(f'ERROR: {exc}', file=sys.stderr)
        sys.exit(2)


if __name__ == '__main__':
    main()
