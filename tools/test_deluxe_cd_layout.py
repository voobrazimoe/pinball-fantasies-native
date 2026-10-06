import os
from pathlib import Path
import unittest
from deluxe_cd_layout import merge, generate, encode, ROOT

class LayoutCompilerTest(unittest.TestCase):
    def test_incompatible_overlap_rejects(self):
        with self.assertRaises(AssertionError): merge([(100,20,1),(110,20,2)])
    def test_adjacent_same_translation_merges(self):
        self.assertEqual(merge([(100,20,1),(120,5,1)]),[dict(destination=100,source=101,size=25)])
    def test_gap_not_copied(self):
        self.assertEqual(len(merge([(100,8,1),(110,8,1)])),2)
    def test_private_reproducibility(self):
        values=[os.getenv(n) for n in ['PF_RUNTIME_DATA','PF_DELUXE_CD_DATA','PF_LAYOUT_AUDIT_EVIDENCE']]
        if not all(values):self.skipTest('private A/D/audit evidence not supplied')
        a,d,e=map(Path,values)
        self.assertEqual((ROOT/'internal/datalayout/deluxe_cd.json').read_text(),encode(generate(e,a,d)))

if __name__=='__main__':unittest.main()
