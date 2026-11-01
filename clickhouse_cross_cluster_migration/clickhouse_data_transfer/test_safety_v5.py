import unittest
import transfer_data as t

class SafetyTests(unittest.TestCase):
    def test_detect_unversioned(self):
        self.assertTrue(t.is_unversioned_replacing({'engine':'ReplacingMergeTree','engine_full':'ReplacingMergeTree ORDER BY (`id`)'}))
        self.assertTrue(t.is_unversioned_replacing({'engine':'ReplacingMergeTree','engine_full':'ReplacingMergeTree'}))
        self.assertFalse(t.is_unversioned_replacing({'engine':'ReplacingMergeTree','engine_full':'ReplacingMergeTree(timestamp) ORDER BY id'}))
        self.assertFalse(t.is_unversioned_replacing({'engine':'MergeTree','engine_full':'MergeTree ORDER BY id'}))
    def test_unknown_engine_full_fails_closed(self):
        with self.assertRaisesRegex(t.MigrationError,'Cannot determine'):
            t.is_unversioned_replacing({'engine':'ReplacingMergeTree','engine_full':''})

if __name__=='__main__':
    unittest.main()
