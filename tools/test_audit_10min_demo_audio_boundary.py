#!/usr/bin/env python3
"""New-pass tests: local CFG mutations and explicit scheduler proof limits."""
import os
from pathlib import Path
import unittest

import audit_10min_demo_audio_boundary as a


class LocalBoundaryMutations(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        try:
            cls.dec = a.Decoder()
        except ImportError:
            raise unittest.SkipTest('research Capstone required')

    def body(self, code):
        return a.body_counts(self.dec, bytes.fromhex(code), 0, base=0, counted=0x20)

    def test_two_calculations_are_not_collapsed(self):
        # Authored two near calls to 0x20 followed by a far return.
        self.assertEqual(self.body('e81d00e81a00cb')['local_path_counts'], [2])

    def test_conditional_zero_or_one(self):
        self.assertEqual(self.body('7403e81b00cb')['local_path_counts'], [0, 1])

    def test_loop_requires_separate_proof(self):
        with self.assertRaisesRegex(ValueError, 'cycle'):
            self.body('ebfe')

    def test_indirect_branch_is_not_a_return(self):
        with self.assertRaisesRegex(ValueError, 'indirect'):
            self.body('ffe0')

    def test_call_effects_are_exported_as_opaque(self):
        r = self.body('ffd0cb')
        self.assertIsNone(r['opaque_calls'][0]['target'])
        self.assertIn('OPAQUE', r['opaque_calls'][0]['effects'])

    def test_unadjusted_return_required(self):
        with self.assertRaisesRegex(ValueError, 'return changed'):
            self.body('ca0200')

    def test_order_and_multiplicity_change_admission(self):
        self.assertEqual(a.paired_trace(['P','L','P'])['total'], 2)
        self.assertEqual(a.paired_trace(['P','P','L'])['total'], 1)

    def test_paused_later_does_not_release_primary(self):
        r = a.paired_trace(['P','pause','L','resume','P','L','P'])
        self.assertEqual(r['electronics_counts'], [1, 0, 1])

    def test_resume_with_clear_latch_can_admit_first_primary(self):
        self.assertEqual(a.paired_trace(['resume','P'], enabled=False)['total'], 1)


class PrivateAudioBoundary(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        data = os.getenv('PF_10MIN_DEMO_DATA')
        if not data:
            raise unittest.SkipTest('private demo required before Capstone import')
        cls.data = Path(data)
        cls.r = a.audit(cls.data, '/private/tmp/pf-dmo0-admission-domains.json')

    def test_scope_does_not_upgrade_dos_or_native_gate(self):
        self.assertEqual(self.r['verdict'], 'NATIVE_AUDIO_BOUNDARY = NOT_PROVED')
        self.assertEqual(self.r['whole_dos_gate']['table1_unknown'], 12)
        self.assertEqual(self.r['whole_dos_gate']['code2_unknown'], 2)
        self.assertTrue(all(not d['scheduler_complete'] for d in self.r['drivers']))

    def test_two_callbacks_and_resolution(self):
        self.assertEqual(self.r['primary']['local_path_counts'], [0, 1])
        self.assertEqual(self.r['later']['local_path_counts'], [0])
        self.assertEqual(self.r['resolution'], dict(low_cx=264, low_cx_hex='0x108', high_cx=174))

    def test_all_driver_handshakes_and_distinct_families(self):
        rows = {d['driver']: d for d in self.r['drivers']}
        self.assertEqual(set(rows), set(a.SDR))
        self.assertEqual(rows['NOSOUND.SDR']['budget'], 'always AX=0')
        self.assertEqual(rows['GUS.SDR']['budget'], 'always AX=0')
        self.assertIsNone(rows['INTERNAL.SDR']['priority_gate'])
        self.assertIn('audio active', rows['INTERNAL.SDR']['budget'])
        self.assertEqual(sum(d['priority_gate'] is not None for d in rows.values()), 8)

    def test_new_input_mutation_fails_before_semantics(self):
        from unittest.mock import patch
        original = Path.read_bytes
        selected = self.data/'TABLE1.PRG'
        def changed(path):
            b = original(path)
            if path == selected:
                b = b[:0x4537] + bytes([b[0x4537] ^ 1]) + b[0x4538:]
            return b
        with patch.object(Path, 'read_bytes', changed):
            with self.assertRaisesRegex(ValueError, 'not pinned'):
                a.audit(self.data, '/private/tmp/pf-dmo0-admission-domains.json')

    def test_handshake_branch_mutation_rejected(self):
        dec = a.Decoder()
        packed = (self.data/'NOSOUND.SDR').read_bytes()
        raw, unused = a.unpack(packed)
        changed = bytearray(raw)
        changed[0x670] = 0x73  # authored replacement: invert priority branch
        with self.assertRaisesRegex(ValueError, 'drift'):
            a.dispatch_shape(dec, bytes(changed), 0x687, True)


if __name__ == '__main__':
    unittest.main()
