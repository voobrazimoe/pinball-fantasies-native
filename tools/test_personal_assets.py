"""Personal inputs must be independent of optional mutable DOS scores."""
import tempfile
import unittest
from pathlib import Path
from personal_assets import ROOT, PERSONAL_INPUTS, inventory


class PersonalAssetsTest(unittest.TestCase):
    @unittest.skipUnless(all((ROOT/name).is_file() for name in PERSONAL_INPUTS),
                         "requires 12 legally obtained DOS runtime files; personal integration only")
    def test_inventory_ignores_missing_and_modified_scores(self):
        with tempfile.TemporaryDirectory() as folder:
            data = Path(folder)
            names = ['INTRO.PRG', 'INTRO.MOD', 'MOD2.MOD', 'PINBALL.CFG']
            names += [f'TABLE{table}.{ext}' for table in range(1, 5) for ext in ['PRG', 'MOD']]
            for name in names:
                (data/name).symlink_to(ROOT/name)
            _, clean = inventory(data)
            self.assertEqual(set(names), {record['name'] for record in clean})
            self.assertEqual(12, len(clean))
            for table in range(1, 5):
                (data/f'TABLE{table}.HI').write_bytes(b'not even a valid score file')
            _, modified = inventory(data)
            self.assertEqual(clean, modified)
            (data/'TABLE1.HI').write_bytes(b'another local record')
            self.assertEqual(clean, inventory(data)[1])


if __name__ == '__main__':
    unittest.main()
