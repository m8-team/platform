import unittest
from unittest.mock import patch
import transfer_data as t


class ReconciliationTests(unittest.TestCase):
    def test_replacing_fingerprint_uses_final_and_both_hashes(self):
        class Client:
            name = 'source'
            def rows(self, sql):
                self.sql = sql
                return [{'rows': 560, 'checksum_sum': '991', 'checksum_xor': '11'}]
        c = Client()
        fp = t.row_fingerprint(c, ('analytics', 'stores'), True)
        self.assertEqual(fp, {'rows': 560, 'checksum_sum': '991', 'checksum_xor': '11'})
        self.assertIn('FROM `analytics`.`stores` FINAL', c.sql)
        self.assertIn('sum(cityHash64(*))', c.sql)
        self.assertIn('groupBitXor(sipHash64(*))', c.sql)

    def test_plain_merge_tree_does_not_use_final(self):
        class Client:
            name = 'source'
            def rows(self, sql):
                self.sql = sql
                return [{'rows': 12, 'checksum_sum': '3', 'checksum_xor': '4'}]
        c = Client()
        t.row_fingerprint(c, ('analytics', 'stores'), False)
        self.assertNotIn(' FINAL', c.sql)

    def test_replicated_fingerprint_per_host_uses_final(self):
        class Client:
            def rows(self, sql):
                self.sql = sql
                return [{'host': 'h1', 'rows': 560, 'checksum_sum': '1', 'checksum_xor': '2'},
                        {'host': 'h2', 'rows': 560, 'checksum_sum': '1', 'checksum_xor': '2'}]
        c = Client()
        fp = t.replica_fingerprints(c, 'default', ('analytics', 'stores'), True)
        self.assertEqual(len(fp), 2)
        self.assertIn("clusterAllReplicas('default', view(SELECT", c.sql)
        self.assertIn('FROM `analytics`.`stores` FINAL)', c.sql)
        self.assertIn('ORDER BY host', c.sql)

    def test_zero_to_nonzero_physical_count_can_pass_if_final_same(self):
        source = object()
        target = object()
        fp = {'rows': 560, 'checksum_sum': '991', 'checksum_xor': '11'}
        states = [{'host': 'h1'}, {'host': 'h2'}]
        replicas = [dict(fp, host='h1'), dict(fp, host='h2')]
        with patch.object(t, 'row_fingerprint', return_value=fp), \
             patch.object(t, 'verified_replicas', return_value=[{'host':'h1','rows':560},{'host':'h2','rows':560}]) as v, \
             patch.object(t, 'replica_status', return_value=states), \
             patch.object(t, 'replica_fingerprints', return_value=replicas):
            same, per_replica = t.verify_fingerprint_on_replicas(
                source, target, 'default', ('analytics', 'stores'), 'ReplacingMergeTree',
                2, 0)
        self.assertEqual(same, fp)
        self.assertEqual(per_replica, replicas)
        self.assertIsNone(v.call_args.args[4])  # physical counts aren't a gate for Replacing

    def test_detect_checksum_mismatch_even_same_rows(self):
        src = {'rows': 560, 'checksum_sum': '991', 'checksum_xor': '11'}
        dst = {'rows': 560, 'checksum_sum': '992', 'checksum_xor': '11'}
        with patch.object(t, 'row_fingerprint', side_effect=[src, dst]):
            with self.assertRaisesRegex(t.MigrationError, 'fingerprints differ'):
                t.verify_fingerprint_on_replicas(
                    object(), object(), 'default', ('analytics', 'stores'), 'ReplacingMergeTree', 2, 0)

    def test_source_snapshot_changed_cannot_pass(self):
        src = {'rows': 560, 'checksum_sum': '991', 'checksum_xor': '11'}
        prior = {'rows': 561, 'checksum_sum': '991', 'checksum_xor': '11'}
        with patch.object(t, 'row_fingerprint', return_value=src):
            with self.assertRaisesRegex(t.MigrationError, 'changed during copy'):
                t.verify_fingerprint_on_replicas(
                    object(), object(), 'default', ('analytics', 'stores'), 'ReplacingMergeTree',
                    2, 0, baseline=prior)

    def test_mixed_replica_checksums_are_rejected(self):
        fp = {'rows': 560, 'checksum_sum': '991', 'checksum_xor': '11'}
        one = dict(fp, host='h1')
        two = dict(fp, host='h2', checksum_xor='12')
        with patch.object(t, 'row_fingerprint', return_value=fp), \
             patch.object(t, 'verified_replicas'), \
             patch.object(t, 'replica_status', return_value=[{'host':'h1'},{'host':'h2'}]), \
             patch.object(t, 'replica_fingerprints', return_value=[one,two]):
            with self.assertRaisesRegex(t.MigrationError, 'does not match source'):
                t.verify_fingerprint_on_replicas(object(),object(),'default',('analytics','stores'),
                                                  'ReplacingMergeTree',2,0)

    def test_none_wanted_rows_allows_physical_merge_differences(self):
        state = [{'host': name, 'is_readonly':0, 'is_session_expired':0,
                  'total_replicas':2, 'active_replicas':2, 'queue_size':0,'inserts_in_queue':0}
                 for name in ('h1','h2')]
        counts = [{'host':'h1','rows':560},{'host':'h2','rows':561}]
        with patch.object(t, 'replica_status', return_value=state), \
             patch.object(t, 'replica_counts', return_value=counts):
            result = t.verified_replicas(object(), 'default', ('analytics','stores'),2,None,0)
        self.assertEqual(result, counts)

if __name__ == '__main__':
    unittest.main()
