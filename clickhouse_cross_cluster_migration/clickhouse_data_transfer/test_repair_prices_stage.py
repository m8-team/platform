import unittest
from unittest.mock import Mock
from repair_prices_stage import (DB, NAME, STAGE, SRC, DST, STG, month_query,
                                 metadata, validate_base, check_stage,
                                 replica_months, verify_all, full_fingerprint)
from transfer_monthly import MigrationError, classify


class TestSafety(unittest.TestCase):
    def test_table_names_distinct(self):
        self.assertEqual(SRC, DST)
        self.assertNotEqual(SRC, STG)
        self.assertTrue(STAGE.endswith('__repair'))

    def test_stage_query_does_not_select_original(self):
        q = month_query(STG, 202608)
        self.assertIn('`ЦеныНоменклатуры__repair`', q)
        self.assertNotIn('`ЦеныНоменклатуры`', q)
        self.assertIn('toYYYYMM(`Период`)=202608', q)

    def test_source_query_uses_original(self):
        q = month_query(SRC, 202609)
        self.assertIn('`ЦеныНоменклатуры`', q)
        self.assertNotIn('__repair', q)

    def test_monthly_query_aggregates_checksum(self):
        q = month_query(STG, 202609, local_host=True)
        self.assertIn('hostName() AS host', q)
        self.assertIn('cityHash64(*)', q)
        self.assertIn('sipHash64(*)', q)

    def test_daily_partitions_below_one_hundred_when_monthly(self):
        import calendar
        for year in (2023, 2024, 2025, 2026):
            for month in range(1, 13):
                self.assertLess(calendar.monthrange(year, month)[1], 100)

    def test_partial_month_conflict(self):
        src = ('13197', '123', '456')
        dst = {'one': ('6658', '4', '5'), 'two': ('6658', '4', '5')}
        self.assertEqual('CONFLICT', classify(src, dst))

    def test_two_empty_replicas_are_empty(self):
        src = ('13197', '123', '456')
        dst = {'one': ('0', '0', '0'), 'two': ('0', '0', '0')}
        self.assertEqual('EMPTY', classify(src, dst))

    def test_equal_fingerprints_complete(self):
        src = ('13197', '123', '456')
        dst = {'one': src, 'two': src}
        self.assertEqual('COMPLETE', classify(src, dst))

    def test_duplicate_month_nonzero_different_checksum_is_conflict(self):
        src = ('100', '1', '2')
        dst = {'one': ('100', '1', '3'), 'two': ('100', '1', '3')}
        self.assertEqual('CONFLICT', classify(src, dst))

    def test_meta_none(self):
        c = Mock()
        c.rows.return_value = []
        self.assertIsNone(metadata(c, STG))

    def test_validate_base_rejects_wrong_engine(self):
        source, target = Mock(), Mock()
        src = {'engine':'ReplacingMergeTree','partition_key':'toYYYYMMDD(`Период`)'}
        dst = {'engine':'ReplicatedMergeTree','partition_key':'toYYYYMMDD(`Период`)'}
        source.rows.return_value = [src]
        target.rows.return_value = [dst]
        with self.assertRaises(MigrationError):
            validate_base(source, target)

    def test_stage_guard_rejects_shared_keeper_path(self):
        target = Mock()
        meta = {'engine':'ReplicatedMergeTree',
                'partition_key':'toYYYYMMDD(`Период`)',
                'sorting_key':'(`Период`, `ТоварГуид`, `ТипЦенГуид`)',
                'primary_key':'(`Период`, `ТоварГуид`, `ТипЦенГуид`)',
                'sampling_key':''}
        # First two metadata calls and DESCRIBE then Keeper paths
        target.rows.side_effect = [[meta], [meta], [], [],
                                   [{'table':NAME,'zookeeper_path':'/keeper/shared'},
                                    {'table':STAGE,'zookeeper_path':'/keeper/shared'}]]
        with self.assertRaisesRegex(MigrationError, 'Keeper path'):
            check_stage(Mock(), target, 'default', 2)

    def test_verify_all_fails_if_stage_incomplete(self):
        source, target = Mock(), Mock()
        src = {202608:('13197','1','2'), 202609:('7121','3','4')}
        source.rows.return_value = [{'month':m, 'rows':int(v[0]),
                                      'checksum_sum':v[1], 'checksum_xor':v[2]} for m,v in src.items()]
        target.rows.return_value = [{'month':202608, 'rows':13197,
                                     'checksum_sum':'1', 'checksum_xor':'2'}]
        with self.assertRaisesRegex(MigrationError, 'incomplete'):
            verify_all(source, target, 'default', {'a', 'b'}, src)


if __name__ == '__main__':
    unittest.main()
