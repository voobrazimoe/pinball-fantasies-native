#!/usr/bin/env python3
"""Regressions for optional mutable CFG in the immutable-original gate."""
import hashlib
import pathlib
import tempfile
import unittest

from validate_macos_originals import IMMUTABLE_NAMES, check_inventory, input_hashes, stage_inputs


class OriginalGateTests(unittest.TestCase):
    def test_optional_and_modified_valid_cfg(self):
        for cfg in (None, bytes([1, 1, 2, 1, 1, 0])):
            with self.subTest(cfg=cfg), tempfile.TemporaryDirectory() as directory:
                root = pathlib.Path(directory)
                data = root/'originals'
                data.mkdir()
                expected = {}
                for name in IMMUTABLE_NAMES:
                    raw = name.encode()
                    (data/name).write_bytes(raw)
                    expected[name] = hashlib.sha256(raw).hexdigest()
                expected['PINBALL.CFG'] = hashlib.sha256(bytes(6)).hexdigest()
                if cfg is not None:
                    (data/'PINBALL.CFG').write_bytes(cfg)
                before = input_hashes(data)
                check_inventory(before, expected)
                staged = root/'tests'
                stage_inputs(data, staged)
                self.assertFalse((staged/'PINBALL.CFG').is_symlink())
                self.assertEqual((staged/'PINBALL.CFG').read_bytes(), cfg or bytes([0, 0, 1, 0, 0, 0]))
                for name in IMMUTABLE_NAMES + ['PINBALL.CFG']:
                    self.assertFalse((staged/name).is_symlink())
                    (staged/name).write_bytes(b'test writes')
                self.assertEqual(before, input_hashes(data))
                self.assertEqual((data/'PINBALL.CFG').exists(), cfg is not None)
                for name in IMMUTABLE_NAMES:
                    bad = dict(before, **{name: 'wrong hash'})
                    with self.assertRaisesRegex(ValueError, name):
                        check_inventory(bad, expected)
                    missing = dict(before)
                    del missing[name]
                    with self.assertRaises(KeyError):
                        check_inventory(missing, expected)


if __name__ == '__main__':
    unittest.main()
