import unittest
from transfer_monthly import TABLE, aggregate_select, classify, normalize, qi, qs


class PartitionMigrationTests(unittest.TestCase):
    def test_identifiers(self):
        self.assertEqual(qi('x`y'), '`x\\`y`')
        self.assertEqual(qi(TABLE[1]), '`ЦеныНоменклатуры`')

    def test_literal_quotes(self):
        self.assertEqual(qs("it's\\ok"), "'it\\'s\\\\ok'")

    def test_month_filter_limits_partitions(self):
        sql = aggregate_select(202503)
        self.assertIn('WHERE toYYYYMM(`Период`) = 202503', sql)
        self.assertIn('FROM `zendendb`.`ЦеныНоменклатуры`', sql)

    def test_bulk_month_summary(self):
        sql = aggregate_select()
        self.assertIn('GROUP BY month', sql)
        self.assertIn('cityHash64(*)', sql)

    def test_empty_if_both_replicas_zero(self):
        src = ('15', '123', '456')
        dst = {'r1': ('0','0','0'), 'r2': ('0','0','0')}
        self.assertEqual(classify(src, dst), 'EMPTY')

    def test_complete_only_if_both_replicas_match(self):
        src = ('15', '123', '456')
        dst = {'r1': src, 'r2': src}
        self.assertEqual(classify(src, dst), 'COMPLETE')

    def test_partial_target_is_conflict(self):
        src = ('15','123','456')
        self.assertEqual(classify(src, {'r1': ('12','10','10'), 'r2': ('12','10','10')}), 'CONFLICT')
        self.assertEqual(classify(src, {'r1': src, 'r2': ('0','0','0')}), 'CONFLICT')

    def test_hash_mismatch_is_conflict(self):
        src = ('15','123','456')
        self.assertEqual(classify(src, {'r1': ('15','123','999'), 'r2': ('15','123','999')}), 'CONFLICT')

    def test_missing_replica_is_conflict(self):
        self.assertEqual(classify(('1','2','3'), {}), 'CONFLICT')

    def test_fingerprint_normalization(self):
        self.assertEqual(normalize({'rows': 10, 'checksum_sum': '2','checksum_xor':'3'}), ('10','2','3'))


if __name__ == '__main__':
    unittest.main()
