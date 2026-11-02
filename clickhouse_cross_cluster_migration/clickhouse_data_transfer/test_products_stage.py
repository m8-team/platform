#!/usr/bin/env python3
import unittest
from unittest.mock import patch
import repair_products_stage as p

class Tests(unittest.TestCase):
    def test_key_is_hex_and_safe(self):
        self.assertEqual(p.key_sql(''), "unhex('')")
        self.assertEqual(p.key_sql('d0a2'), "unhex('D0A2')")
        for bad in ["'); DROP TABLE x;--", 'A', 'GG', 'A0G0', None, 32, '漢字']:
            with self.subTest(bad=bad):
                with self.assertRaises(p.MigrationError):
                    p.key_sql(bad)

    def test_valid_partition_name(self):
        self.assertTrue(p.valid_partition('`ТипНоменклатуры`'))
        self.assertTrue(p.valid_partition('ТипНоменклатуры'))
        self.assertFalse(p.valid_partition('toYYYYMM(`Период`)'))
        self.assertFalse(p.valid_partition('`ТипНоменклатуры`, `ТоварГуид`'))

    def test_private_labels(self):
        self.assertEqual(p.safe_label(''), '<empty-string>')
        self.assertEqual(p.safe_label('414243'), 'hex:414243')
        self.assertTrue(p.safe_label('FF'*40).endswith('...'))

    def test_source_select_is_one_partition(self):
        sql=p.one_partition_sql(p.SRC, 'D0A2', host=True)
        self.assertIn('WHERE `ТипНоменклатуры` = unhex(\'D0A2\')', sql)
        self.assertIn('hostName() AS host', sql)
        self.assertNotIn('GROUP BY', sql)
        self.assertIn('count() AS rows', sql)

    def test_classification_strict(self):
        expected=('4', '17', '16')
        self.assertEqual(p.classify(expected, {'a': expected, 'b': expected}), 'COMPLETE')
        self.assertEqual(p.classify(expected, {'a': ('0','0','0'), 'b': ('0','0','0')}), 'EMPTY')
        self.assertEqual(p.classify(expected, {'a': ('2','12','2'), 'b': expected}), 'CONFLICT')

    def test_stage_not_same_keeper_path(self):
        class Dummy:
            def rows(self, query):
                if 'zookeeper_path' in query:
                    return [{'table':p.TABLE, 'zookeeper_path':'/p'},
                            {'table':p.STAGE, 'zookeeper_path':'/p'}]
                raise AssertionError(query)
        with patch.object(p, 'metadata', side_effect=[{'engine':'ReplicatedMergeTree','partition_key':'x','sorting_key':'y','primary_key':'z','sampling_key':''},
                                                      {'engine':'ReplicatedMergeTree','partition_key':'x','sorting_key':'y','primary_key':'z','sampling_key':''}]), \
             patch.object(p, 'describe', side_effect=[[{'name':'a'}],[{'name':'a'}]]):
            with self.assertRaisesRegex(p.MigrationError, 'different Keeper'):
                p.check_stage(Dummy(), 'default', 2)

    def test_base_rejects_wrong_partition(self):
        class Dummy: label='dummy'
        metadata={'engine':'MergeTree','partition_key':'toYYYYMM(date)','sorting_key':'x','primary_key':'x','sampling_key':''}
        clone={**metadata, 'engine':'ReplicatedMergeTree'}
        with patch.object(p, 'metadata', side_effect=[metadata,clone]):
            with self.assertRaisesRegex(p.MigrationError, 'Unexpected partition'):
                p.validate_base(Dummy(), Dummy())

    def test_base_rejects_unexpected_engine(self):
        class Dummy: label='dummy'
        metadata={'engine':'ReplacingMergeTree','partition_key':'ТипНоменклатуры','sorting_key':'x','primary_key':'x','sampling_key':''}
        clone={**metadata, 'engine':'ReplicatedReplacingMergeTree'}
        with patch.object(p, 'metadata', side_effect=[metadata,clone]):
            with self.assertRaisesRegex(p.MigrationError, 'Unexpected source/target engine'):
                p.validate_base(Dummy(), Dummy())

    def test_schema_requires_string(self):
        class Dummy: label='dummy'
        metadata={'engine':'MergeTree','partition_key':'ТипНоменклатуры','sorting_key':'x','primary_key':'x','sampling_key':''}
        clone={**metadata, 'engine':'ReplicatedMergeTree'}
        with patch.object(p, 'metadata', side_effect=[metadata,clone]), \
             patch.object(p, 'describe', side_effect=[[{'name':p.PART_KEY,'type':'Nullable(String)'}],
                                                        [{'name':p.PART_KEY,'type':'Nullable(String)'}]]):
            with self.assertRaisesRegex(p.MigrationError, 'exactly String'):
                p.validate_base(Dummy(), Dummy())

    def test_partition_summary_preserves_raw_bytes(self):
        class Dummy:
            def rows(self, sql):
                self.sql=sql
                return [{'pk_hex':'D0A2','rows':3,'checksum_sum':'22','checksum_xor':'33'},
                        {'pk_hex':'','rows':1,'checksum_sum':'4','checksum_xor':'4'}]
        d=Dummy()
        actual=p.partition_summary(d, p.SRC)
        self.assertEqual(actual['D0A2'], ('3','22','33'))
        self.assertEqual(actual[''], ('1','4','4'))
        self.assertIn('GROUP BY pk_hex', d.sql)

    def test_missing_source_target(self):
        class Dummy: label='dummy'
        with patch.object(p, 'metadata', side_effect=[None, None]):
            with self.assertRaisesRegex(p.MigrationError, 'missing'):
                p.validate_base(Dummy(),Dummy())

    def test_table_name(self):
        self.assertEqual(p.table_sql(p.STG),'`zendendb`.`Товары__repair`')

if __name__ == '__main__':
    unittest.main()
