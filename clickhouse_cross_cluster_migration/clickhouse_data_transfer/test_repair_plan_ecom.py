import unittest
from unittest.mock import patch, MagicMock
import repair_plan_ecom as m

class TestRepair(unittest.TestCase):
    def test_constants(self):
        self.assertEqual(m.DB, 'zendendb')
        self.assertEqual(m.TABLE, 'ПланПродаж_ecom')
        self.assertNotEqual(m.TABLE, m.STAGE)

    def test_native_bytes_transport(self):
        cli = MagicMock()
        cli.username='user';cli.password='password';cli.timeout=5
        cli.url='https://example.com:8443/'
        cli.context=None
        with patch('repair_plan_ecom.urllib.request.urlopen') as opener:
            opener.return_value.__enter__.return_value.read.return_value = b'\x00\xff\xfe'
            payload = m.execute_binary(cli, 'SELECT 1 FORMAT Native')
            self.assertEqual(payload,b'\x00\xff\xfe')
            sent=opener.call_args.args[0]
            self.assertIn(b'FORMAT Native',sent.data)
            self.assertEqual(sent.get_header('X-clickhouse-key'), 'password')

    def test_quoted_identifiers(self):
        self.assertEqual(m.ident('ПланПродаж_ecom'), '`ПланПродаж_ecom`')

    def test_meta_rejects_wrong_engine(self):
        a=MagicMock();b=MagicMock()
        with patch('repair_plan_ecom.get_metadata') as metadata:
            metadata.side_effect=[{(m.DB,m.TABLE): {'engine':'MergeTree'}},
                                  {(m.DB,m.TABLE): {'engine':'ReplicatedReplacingMergeTree'}}]
            with self.assertRaises(m.MigrationError):
                m.meta_check(a,b)

if __name__=='__main__': unittest.main()

class TestRealEngineExpressions(unittest.TestCase):
    def test_unversioned_with_order_by_and_settings(self):
        from transfer_data import is_unversioned_replacing
        for full in (
            'ReplacingMergeTree ORDER BY (`Подразделение`, `Дата`) SETTINGS index_granularity = 8192',
            'ReplacingMergeTree() ORDER BY (`Подразделение`, `Дата`)',
            'ReplacingMergeTree SETTINGS index_granularity = 8192',
            'ReplacingMergeTree PARTITION BY toYYYYMM(timestamp) ORDER BY id',
            'ReplacingMergeTree',
        ):
            with self.subTest(full=full):
                self.assertTrue(is_unversioned_replacing({'engine':'ReplacingMergeTree','engine_full':full}))

    def test_versioned_must_be_rejected(self):
        from transfer_data import is_unversioned_replacing
        for full in ('ReplacingMergeTree(timestamp) ORDER BY id',
                     'ReplacingMergeTree (CHANGED_DATE) ORDER BY id',
                     'ReplacingMergeTree(timestamp, is_deleted) ORDER BY id'):
            with self.subTest(full=full):
                self.assertFalse(is_unversioned_replacing({'engine':'ReplacingMergeTree','engine_full':full}))

    def test_unexpected_engine_full_fails_closed(self):
        from transfer_data import is_unversioned_replacing
        for full in ('ReplacingMergeTreeUnexpected', 'ReplacingMergeTree unknown syntax'):
            with self.subTest(full=full):
                with self.assertRaises(m.MigrationError):
                    is_unversioned_replacing({'engine':'ReplacingMergeTree','engine_full':full})

    def test_meta_check_accepts_real_source_ddl(self):
        source = MagicMock()
        target = MagicMock()
        src = {'engine':'ReplacingMergeTree',
               'engine_full':'ReplacingMergeTree ORDER BY (`Подразделение`, `Дата`) SETTINGS index_granularity = 8192',
               'sorting_key':'Подразделение, Дата'}
        dst = {'engine':'ReplicatedReplacingMergeTree','sorting_key':'Подразделение, Дата'}
        with patch('repair_plan_ecom.get_metadata', side_effect=[{(m.DB,m.TABLE):src},{(m.DB,m.TABLE):dst}]), \
             patch('repair_plan_ecom.get_columns', side_effect=[{(m.DB,m.TABLE): [('a','Int64')]},{(m.DB,m.TABLE): [('a','Int64')]}]), \
             patch('repair_plan_ecom.describe_columns', return_value=[('a','Int64')]):
            self.assertFalse(m.meta_check(source,target))

    def test_meta_check_rejects_shared_keeper_path(self):
        source = MagicMock()
        target = MagicMock()
        src = {'engine':'ReplacingMergeTree','engine_full':'ReplacingMergeTree ORDER BY id', 'sorting_key':'id'}
        dst = {'engine':'ReplicatedReplacingMergeTree','sorting_key':'id'}
        stage = {'engine':'ReplicatedReplacingMergeTree','sorting_key':'id'}
        target.rows.return_value = [
            {'table':m.TABLE, 'zookeeper_path':'/same'},
            {'table':m.STAGE, 'zookeeper_path':'/same'},
        ]
        with patch('repair_plan_ecom.get_metadata', side_effect=[{(m.DB,m.TABLE):src},{(m.DB,m.TABLE):dst,(m.DB,m.STAGE):stage}]), \
             patch('repair_plan_ecom.get_columns', side_effect=[{(m.DB,m.TABLE): [('a','Int64')]},{(m.DB,m.TABLE): [('a','Int64')], (m.DB,m.STAGE): [('a','Int64')]}]), \
             patch('repair_plan_ecom.describe_columns', return_value=[('a','Int64')]):
            with self.assertRaisesRegex(m.MigrationError, 'SHARES Keeper path'):
                m.meta_check(source,target,stage_exists=True)
