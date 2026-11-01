import unittest
from diagnose_replacing import compare, build_index, signature


class CompareTests(unittest.TestCase):
    def test_same_logical_rows_after_merge(self):
        v1 = {'Подразделение':'A','Дата':'2026-01-01','timestamp':1,'Value':10}
        v2 = {'Подразделение':'A','Дата':'2026-01-01','timestamp':2,'Value':20}
        result = compare([v2], [v2], [v1, v2], [v2], ('Подразделение','Дата'))
        self.assertTrue(result['matched'])
        self.assertEqual(result['source']['physical_rows'], 2)
        self.assertEqual(result['source']['duplicate_keys'], 1)

    def test_diff_winner_originates_from_raw_source(self):
        v1 = {'Подразделение':'A','Дата':'2026-01-01','timestamp':1,'Value':10}
        v2 = {'Подразделение':'A','Дата':'2026-01-01','timestamp':2,'Value':20}
        result = compare([v2], [v1], [v1, v2], [v1, v2], ('Подразделение','Дата'))
        self.assertFalse(result['matched'])
        self.assertEqual(result['changed_keys_count'], 1)
        entry = result['changed_keys'][0]
        self.assertTrue(entry['target_winner_exists_in_source_raw'])
        self.assertTrue(entry['source_winner_exists_in_target_raw'])
        self.assertEqual(entry['changed_columns'], ['Value', 'timestamp'])

    def test_missing_key(self):
        a={'Подразделение':'A','Дата':'2026-01-01','Value':2}
        b={'Подразделение':'B','Дата':'2026-01-01','Value':3}
        result=compare([a], [b], [a], [b], ('Подразделение','Дата'))
        self.assertEqual(len(result['missing_on_target']),1)
        self.assertEqual(len(result['missing_on_source']),1)

    def test_canonical_dict_order(self):
        self.assertEqual(signature({'b':2,'a':1}), signature({'a':1,'b':2}))

    def test_duplicate_final_key_rejected(self):
        a={'Подразделение':'A','Дата':'2026-01-01'}
        with self.assertRaisesRegex(Exception,'FINAL returned duplicate'):
            build_index([a,a],('Подразделение','Дата'))

if __name__ == '__main__':
    unittest.main()
