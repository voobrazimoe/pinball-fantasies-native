"""Exercise the no-history exporter against a synthetic approval of one blob."""
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parent.parent

class SnapshotTest(unittest.TestCase):
    @unittest.skipUnless((ROOT/".git").exists(), "requires Git checkout metadata; a source-only export has no .git")
    def test_only_approved_blob_is_exported_and_hashes_are_enforced(self):
        commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()
        names = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', commit], cwd=ROOT, text=True).splitlines()
        raw = subprocess.check_output(['git', 'show', commit+':README.md'], cwd=ROOT)
        review = {'commit': commit, 'legal_decisions': 'SYNTHETIC TEST ONLY; no public approval',
                  'files': {'README.md': hashlib.sha256(raw).hexdigest()},
                  'excluded': {name: 'synthetic fixture exclusion' for name in names if name != 'README.md'}}
        with tempfile.TemporaryDirectory() as folder:
            base = Path(folder)
            approval = base/'review.json'
            approval.write_text(json.dumps(review))
            cmd = ['python3', str(ROOT/'tools/export_public_snapshot.py'), '--approval', str(approval), '--output', str(base/'tree')]
            subprocess.run(cmd, check=True, capture_output=True)
            self.assertEqual([p.name for p in (base/'tree').iterdir()], ['README.md'])
            self.assertEqual((base/'tree/README.md').read_bytes(), raw)
            review['files']['README.md'] = '0'*64
            approval.write_text(json.dumps(review))
            cmd[-1] = str(base/'bad')
            self.assertNotEqual(subprocess.run(cmd, capture_output=True).returncode, 0)
            self.assertFalse((base/'bad').exists())

if __name__ == '__main__':
    unittest.main()
