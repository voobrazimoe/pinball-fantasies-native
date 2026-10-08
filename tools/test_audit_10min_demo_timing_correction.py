"""Corrected fixed witnesses, separate from archived shortened-buffer results."""
import json
import os
from pathlib import Path
import unittest


class CorrectedWitnesses(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        p = Path(os.getenv('PF_DMO_TIMING_CORRECTION', '/private/tmp/pf-dmo-impl2b7-final/summary.json'))
        if not p.exists():
            raise unittest.SkipTest('corrected private fixed replay is NOT AVAILABLE')
        cls.r = json.loads(p.read_text())
        if set(cls.r['runs']) != {'post', 'party', 'setball', 'drop', 'flash'}:
            raise unittest.SkipTest('corrected fixed replay is incomplete')

    def test_basis(self):
        self.assertEqual(self.r['timing_oracle_basis'], 'FF-inclusive-v1')
        self.assertEqual(self.r['configuration']['TimingBasis'], 'FF-inclusive-v1')
        self.assertEqual(self.r['configuration']['ScrollLengths'], [98, 88])

    def test_concrete_dates_and_observed_visits(self):
        for name, quit_at, visits in [('party', 37047, 1050), ('setball', 37047, 1050),
                                      ('drop', 37047, 1050), ('flash', 37220, 1048)]:
            w = self.r['runs'][name]['witness']
            self.assertEqual((w['QUIT'], w['admitted_visits']), (quit_at, visits))
            self.assertGreater(w['prefix_equality_matched']['rows'], 0)
            self.assertGreater(w['prefix_equality_matched']['boundaries'], 0)
        for name, quit_at in [('no_input', 102583), ('single_launch', 102583), ('periodic_launch', 38133)]:
            w = self.r['runs']['post'][name]
            self.assertEqual((w['QUIT'], w['admitted_visits']), (quit_at, 1050))
            self.assertGreater(w['prefix_equality_matched']['rows'], 0)

    def test_every_command(self):
        ops = ['_CLEAR4', '_SCROLL', '_FLASHON', '_PRINT13_NUMBER', '_WAIT',
               '_FLASHOFF', '_SCROLL', '_FADE', '_WAIT', 'QUIT']
        cases = [('party', 'witness', 35998), ('setball', 'witness', 35998),
                 ('drop', 'witness', 35998), ('flash', 'witness', 36173),
                 ('post', 'no_input', 101534), ('post', 'single_launch', 101534),
                 ('post', 'periodic_launch', 37084)]
        for group, name, start in cases:
            rows = [x for x in self.r['runs'][group][name]['transitions'] if x['calculation'] >= start]
            self.assertEqual([x['matrix_op'] for x in rows], ops, (group, name))
            self.assertEqual(rows[0]['expiry_visits'], 1)
            self.assertEqual(rows[1]['matrix_text_left'], 78)
            self.assertEqual(rows[6]['matrix_text_left'], 68)
            self.assertEqual(rows[-1]['expiry_visits'], 1048 if group == 'flash' else 1050)
            phase = 4 if group == 'flash' else 8
            self.assertEqual(rows[0]['scroll_phase'], phase)
            self.assertEqual(rows[1]['scroll_phase'], phase)
            offsets = [0, 4, 315, 316, 317, 417, 418, 691, 947, 1047] if group == 'flash' else [0, 4, 317, 318, 319, 419, 420, 693, 949, 1049]
            self.assertEqual([x['calculation']-start for x in rows], offsets)

    def test_pause_counterexample_preserved(self):
        c = self.r['pause_cycle']
        self.assertTrue(c['closed_transition'])
        self.assertTrue(c['gameplay_state_equal'])
        self.assertTrue(c['no_exit'])


class BasisSeparation(unittest.TestCase):
    def test_archived_date_cannot_claim_corrected_basis(self):
        import audit_10min_demo_party_on as party
        if not party.OUTPUT.exists():
            self.skipTest('historical private artifact is NOT AVAILABLE')
        report = json.loads(party.OUTPUT.read_text())
        self.assertTrue(party.validate(report))
        report['timing_oracle_basis'] = 'FF-inclusive-v1'
        with self.assertRaises((AssertionError, ValueError)):
            party.validate(report)


if __name__ == '__main__':
    unittest.main()
