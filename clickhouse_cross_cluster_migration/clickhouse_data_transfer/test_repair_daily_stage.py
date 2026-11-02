import unittest
import calendar
from unittest.mock import Mock

import repair_daily_stage as d
import transfer_data as t
from transfer_monthly import MigrationError, classify


class DailyStageTests(unittest.TestCase):
    def setUp(self):
        d.select_table('zendendb.ЗаказыПокупателей')

    def test_table_names_and_state(self):
        self.assertEqual(d.SRC, ('zendendb', 'ЗаказыПокупателей'))
        self.assertEqual(d.STG, ('zendendb', 'ЗаказыПокупателей__repair'))
        self.assertNotEqual(d.SRC, d.STG)
        a = d.default_state_path()
        d.select_table('zendendb.ТоварыВРезервеНаСкладах')
        self.assertNotEqual(d.default_state_path(), a)
        self.assertEqual(d.STG[1], 'ТоварыВРезервеНаСкладах__repair')

    def test_user_supplied_table_escaping(self):
        d.select_table('zendendb.Моя таблица')
        self.assertEqual(d.table_sql(d.STG), '`zendendb`.`Моя таблица__repair`')
        d.select_table('zendendb.Номенклатура`A')
        self.assertEqual(d.table_sql(d.SRC), '`zendendb`.`Номенклатура\\`A`')

    def test_invalid_names(self):
        for bad in ('abc', 'a.b.c', '.x', 'a.', 'a.b__repair', 'a.b\nDROP'):
            with self.subTest(name=bad):
                with self.assertRaises(MigrationError):
                    d.select_table(bad)

    def test_month_sql(self):
        d.select_table('zendendb.ЗаказыПокупателей')
        sql = d.month_query(d.STG, 202608, local_host=True)
        self.assertIn('`ЗаказыПокупателей__repair`', sql)
        self.assertIn('toYYYYMM(`Период`)=202608', sql)
        self.assertIn('hostName() AS host', sql)
        self.assertIn('cityHash64(*)', sql)

    def test_month_max_partitions(self):
        for y in range(2020, 2029):
            for m in range(1, 13):
                self.assertLess(calendar.monthrange(y, m)[1], 100)

    def test_conflict(self):
        self.assertEqual('CONFLICT', classify(('100', '1', '2'), {'node1': ('99', '1', '2'), 'node2': ('99', '1', '2')}))
        self.assertEqual('EMPTY', classify(('100', '1', '2'), {'node1': ('0', '0', '0'), 'node2': ('0', '0', '0')}))
        self.assertEqual('COMPLETE', classify(('100', '1', '2'), {'node1': ('100', '1', '2'), 'node2': ('100', '1', '2')}))

    def test_reject_non_daily_partition(self):
        source, target = Mock(), Mock()
        src = {'engine': 'MergeTree', 'partition_key': 'toYear(`Период`)', 'sorting_key': '`Период`', 'primary_key': '`Период`', 'sampling_key': ''}
        dst = dict(src, engine='ReplicatedMergeTree')
        source.rows.return_value = [src]
        target.rows.return_value = [dst]
        with self.assertRaisesRegex(MigrationError, 'Unexpected source partition key'):
            d.validate_base(source, target)

    def test_reject_replacing_engine(self):
        source, target = Mock(), Mock()
        src = {'engine': 'ReplacingMergeTree', 'partition_key': 'toYYYYMMDD(`Период`)' }
        dst = {'engine': 'ReplicatedReplacingMergeTree', 'partition_key': 'toYYYYMMDD(`Период`)' }
        source.rows.return_value = [src]
        target.rows.return_value = [dst]
        with self.assertRaisesRegex(MigrationError, 'Wrong source/target engine'):
            d.validate_base(source, target)

    def test_keep_separate_keeper(self):
        target = Mock()
        dst = {'engine': 'ReplicatedMergeTree', 'partition_key': 'toYYYYMMDD(`Период`)', 'sorting_key': 'x', 'primary_key': 'x', 'sampling_key': ''}
        target.rows.side_effect = [[dst], [dst], [], [], [{'table': d.NAME, 'zookeeper_path': '/same'}, {'table': d.STAGE, 'zookeeper_path': '/same'}]]
        with self.assertRaisesRegex(MigrationError, 'Keeper path'):
            d.check_stage(Mock(), target, 'default', 2)

    def test_missing_month_is_not_verified(self):
        source, target = Mock(), Mock()
        src = {202608: ('100', '1', '2'), 202609: ('200', '2', '3')}
        source.rows.return_value = [{'month': m, 'rows': int(v[0]), 'checksum_sum': v[1], 'checksum_xor': v[2]} for m, v in src.items()]
        target.rows.return_value = [{'month': 202608, 'rows': 100, 'checksum_sum': '1', 'checksum_xor': '2'}]
        with self.assertRaisesRegex(MigrationError, 'incomplete'):
            d.verify_all(source, target, 'default', {'one', 'two'}, src)


class BulkSafetyTests(unittest.TestCase):
    def test_partition_precheck(self):
        c = Mock()
        c.rows.return_value = [{'partitions': 427}]
        self.assertEqual(427, t.source_partition_count(c, ('zendendb', 'ТоварыВРезервеНаСкладах')))
        self.assertIn('system.parts', c.rows.call_args[0][0])
        self.assertIn('active = 1', c.rows.call_args[0][0])

    def test_bulk_prestore_guard(self):
        from pathlib import Path
        s = Path(t.__file__).read_text(encoding='utf-8')
        self.assertLess(s.index('partition_count = source_partition_count(source, key)'), s.index("record = {'status': 'COPYING'"))
        self.assertIn('SKIP HIGH PARTITIONS', s)
        self.assertIn('--skip-high-partitions', s)


if __name__ == '__main__':
    unittest.main()
