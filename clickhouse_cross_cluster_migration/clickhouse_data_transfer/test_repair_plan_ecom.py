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
