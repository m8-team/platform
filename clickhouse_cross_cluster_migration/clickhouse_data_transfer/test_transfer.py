import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import transfer_data as t


class TransferTests(unittest.TestCase):
    def test_escaping(self):
        self.assertEqual(t.ident('А`B'), '`А\\`B`')
        self.assertEqual(t.lit("ab'\\cd"), "'ab\\'\\\\cd'")

    def test_no_mutation_journal_before_copy(self):
        with tempfile.TemporaryDirectory() as tmp:
            p = Path(tmp) / 'journal.json'
            journal = {'format_version': 1, 'tables': {'analytics.stores': {'status': 'COPYING'}}}
            t.save_journal(p, journal)
            self.assertEqual(t.read_journal(p), journal)
            self.assertEqual(p.stat().st_mode & 0o777, 0o600)

    def test_validate_missing_and_mismatch(self):
        key = ('analytics', 'stores')
        src = {key: {'engine': 'ReplacingMergeTree', 'partition_key': '', 'sorting_key': 'id',
                     'primary_key': 'id', 'sampling_key': ''}}
        dst = {key: {'engine': 'ReplicatedReplacingMergeTree', 'partition_key': '', 'sorting_key': 'id',
                     'primary_key': 'id', 'sampling_key': ''}}
        cols = {key: [('id', 'UUID', '1', '', '')]}
        self.assertIsNone(t.validate_table(key, src, dst, cols, cols))
        self.assertEqual(t.validate_table(key, src, {}, cols, cols), 'target table does not exist')
        self.assertEqual(t.validate_table(key, src, dst, cols, {}), 'column definitions differ')

    def test_get_columns_clickhouse_26_8(self):
        class FakeClient:
            def rows(self, sql):
                self.sql = sql
                return [{'database': 'analytics', 'table': 'stores', 'name': 'id',
                         'type': 'UUID', 'position': 1, 'default_kind': '',
                         'default_expression': ''}]
        client = FakeClient()
        result = t.get_columns(client, ['analytics'])
        self.assertEqual(result[('analytics', 'stores')], [('id', 'UUID', '1', '', '')])
        self.assertNotIn('codec_expression', client.sql)
        self.assertNotIn('ttl_expression', client.sql)

    def test_describe_checks_codec_ttl_and_comment(self):
        class FakeClient:
            def rows(self, sql):
                self.sql = sql
                return [{'name': 'a', 'type': 'String', 'default_type': 'DEFAULT',
                         'default_expression': "'abc'", 'comment': 'text',
                         'codec_expression': 'ZSTD(1)', 'ttl_expression': 'ts + INTERVAL 1 DAY'}]
        client = FakeClient()
        descriptions = t.describe_columns(client, ('analytics', 'sales'))
        self.assertEqual(descriptions, [('a', 'String', 'DEFAULT', "'abc'",
                                         'text', 'ZSTD(1)', 'ts + INTERVAL 1 DAY')])
        self.assertIn('DESCRIBE TABLE `analytics`.`sales`', client.sql)

    def test_error_redaction(self):
        value = t.safe_error('remote login password_secret is invalid', ['password_secret'])
        self.assertNotIn('password_secret', value)
        self.assertIn('[REDACTED]', value)

    def test_replica_counts_includes_empty_hosts(self):
        # The previous query GROUPed only data rows and returned [] for an
        # empty table. The corrected query anchors on system.one per host.
        class FakeClient:
            def rows(self, sql):
                self.sql = sql
                return [{'host': 'node1', 'rows': 0}, {'host': 'node2', 'rows': 0}]
        client = FakeClient()
        result = t.replica_counts(client, 'default', ('analytics', 'stores'))
        self.assertEqual(result, [{'host': 'node1', 'rows': 0},
                                  {'host': 'node2', 'rows': 0}])
        self.assertIn("clusterAllReplicas('default', system.one)", client.sql)
        self.assertIn("clusterAllReplicas('default', 'analytics', 'stores')", client.sql)
        self.assertIn('UNION ALL', client.sql)
        self.assertIn('sum(row_count) AS rows', client.sql)

    def test_verified_replicas_checks_unique_hosts(self):
        st = [{'host': name, 'is_readonly': 0, 'is_session_expired': 0,
               'total_replicas': 2, 'active_replicas': 2,
               'queue_size': 0, 'inserts_in_queue': 0}
              for name in ('node1', 'node2')]
        wrong_counts = [{'host': 'node1', 'rows': 0}, {'host': 'node1', 'rows': 0}]
        with patch.object(t, 'replica_status', return_value=st), \
             patch.object(t, 'replica_counts', return_value=wrong_counts):
            with self.assertRaisesRegex(t.MigrationError, 'verification timed out'):
                t.verified_replicas(object(), 'default', ('analytics', 'stores'), 2, 0, 0)

    def test_verified_replicas_zero_rows_on_both_hosts(self):
        st = [{'host': name, 'is_readonly': 0, 'is_session_expired': 0,
               'total_replicas': 2, 'active_replicas': 2,
               'queue_size': 0, 'inserts_in_queue': 0}
              for name in ('node1', 'node2')]
        zero_counts = [{'host': 'node1', 'rows': 0}, {'host': 'node2', 'rows': 0}]
        with patch.object(t, 'replica_status', return_value=st), \
             patch.object(t, 'replica_counts', return_value=zero_counts):
            result = t.verified_replicas(object(), 'default', ('analytics', 'stores'), 2, 0, 0)
        self.assertEqual(result, zero_counts)

    def test_replica_log_pointer_does_not_falsely_block_copy(self):
        # On ClickHouse, log_pointer can temporarily trail log_max_index even
        # if both live replicas report an empty queue and matching data.
        st = [{'host': name, 'is_readonly': 0, 'is_session_expired': 0,
               'total_replicas': 2, 'active_replicas': 2,
               'queue_size': 0, 'inserts_in_queue': 0,
               'log_pointer': 0, 'log_max_index': 1}
              for name in ('node1', 'node2')]
        counts = [{'host': name, 'rows': 0} for name in ('node1', 'node2')]
        with patch.object(t, 'replica_status', return_value=st), \
             patch.object(t, 'replica_counts', return_value=counts):
            self.assertEqual(
                t.verified_replicas(object(), 'default', ('analytics', 'stores'), 2, 0, 0), counts)

    def test_replica_timeout_explains_actual_mismatch(self):
        st = [{'host': name, 'is_readonly': 0, 'is_session_expired': 0,
               'total_replicas': 2, 'active_replicas': 2,
               'queue_size': 0, 'inserts_in_queue': 0,
               'log_pointer': 10, 'log_max_index': 10}
              for name in ('node1', 'node2')]
        counts = [{'host': 'node1', 'rows': 561}, {'host': 'node2', 'rows': 0}]
        with patch.object(t, 'replica_status', return_value=st), \
             patch.object(t, 'replica_counts', return_value=counts):
            with self.assertRaisesRegex(t.MigrationError, '561 physical rows') as cm:
                t.verified_replicas(object(), 'default', ('analytics', 'stores'), 2, 561, 0)
        self.assertIn('"rows": 0', str(cm.exception))
        self.assertIn('"log_pointer": 10', str(cm.exception))

    def test_inspection_is_read_only(self):
        st = [{'host': 'node1', 'is_readonly': 0, 'is_session_expired': 0,
               'total_replicas': 2, 'active_replicas': 2,
               'queue_size': 0, 'inserts_in_queue': 0,
               'log_pointer': 0, 'log_max_index': 0}]
        counts = [{'host': 'node1', 'rows': 0}]
        with patch.object(t, 'rowcount', side_effect=[561, 0]), \
             patch.object(t, 'replica_status', return_value=st), \
             patch.object(t, 'replica_counts', return_value=counts):
            outcome = t.inspect_table(object(), object(), 'default', ('analytics', 'stores'),
                                      {'tables': {}})
        self.assertEqual(outcome['source_rows'], 561)
        self.assertEqual(outcome['target_local_rows'], 0)
        self.assertEqual(outcome['journal_status'], 'PENDING')

    def test_replica_verification(self):
        target = object()
        status = [{'host': 'a', 'is_readonly': 0, 'is_session_expired': 0,
                   'total_replicas': 2, 'active_replicas': 2, 'queue_size': 0,
                   'inserts_in_queue': 0, 'log_pointer': 21, 'log_max_index': 21},
                  {'host': 'b', 'is_readonly': 0, 'is_session_expired': 0,
                   'total_replicas': 2, 'active_replicas': 2, 'queue_size': 0,
                   'inserts_in_queue': 0, 'log_pointer': 21, 'log_max_index': 21}]
        counts = [{'host': 'a', 'rows': 561}, {'host': 'b', 'rows': 561}]
        with patch.object(t, 'replica_status', return_value=status), \
             patch.object(t, 'replica_counts', return_value=counts):
            result = t.verified_replicas(target, 'default', ('analytics', 'stores'), 2, 561, 0)
        self.assertEqual(result, counts)


if __name__ == '__main__':
    unittest.main()
