#!/usr/bin/env python3
"""Private static demo obligations. No fixture -> clean SKIP; no DOS execution."""
import os
from pathlib import Path
import unittest
from unittest.mock import patch

import audit_10min_demo_graph as graph


class PrivateDemoGraph(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        names = ('PF_10MIN_DEMO_DATA', 'PF_RUNTIME_DATA', 'PF_DMO0_HISTORICAL_SOURCE')
        values = [os.getenv(n) for n in names]
        if not all(values):
            raise unittest.SkipTest('set '+', '.join(names)+' for private static evidence')
        cls.paths = list(map(Path, values))
        cls.result = graph.audit(*cls.paths)

    def test_bonus_consumer_and_removed_full_continuation(self):
        r = self.result['bonus_program']
        self.assertEqual(r['start'], 0x1b459)
        tail = [n['node'] for n in r['nodes'][51:]]
        self.assertEqual(tail, ['_KOLLA_XXBALL', '_DEMOVER_CHANGE_PLAYER', '_CLEAR4', '_WAIT', '0'])
        self.assertEqual(r['nodes'][52]['source'], 0x1b533)
        self.assertEqual(r['nodes'][52]['handler_file'], 0x73e)
        full = [n['node'] for n in self.result['canonical_bonus_program']['nodes'][51:]]
        self.assertIn('_CHANGE_PLAYER', full)
        self.assertIn('_WAITIFMULTI', full)

    def test_cards_use_consumer_crop_and_demo_menu_records(self):
        p = self.result['presentation']
        self.assertEqual(p['cards']['source_crop'], [0, 0, 440, 95])
        self.assertEqual(p['cards']['destinations'], [[160, 10], [160, 135]])
        self.assertEqual(p['cards']['copied_rows'], 95)
        self.assertEqual(p['cards']['palette_banks'], [16, 0])
        self.assertEqual(p['cards']['raster_packets'], 18)
        self.assertEqual(p['sidebar_source'], 0x396d3)
        self.assertEqual(p['options_source'], 0x3974b)
        self.assertFalse(any('F2' in r or 'F3' in r or 'F4' in r for r in p['sidebar_rows']))
        self.assertEqual(p['pages'][0]['rows'][3], '10 MINUTE DEMO')
        self.assertEqual(p['pages'][0]['rows'][9], 'F1 - PLAY PINBALL')
        self.assertEqual(p['pages'][1]['rows'][2], 'IS AVAILABLE NOW')
        self.assertTrue(self.result['intro_text_consumer_reached'])

    def test_absent_direct_references_do_not_close_indirect_proof(self):
        for p in self.result['persistence']:
            self.assertFalse(p['raw_rel16_incoming'])
            self.assertFalse(p['entry_pointer_in_main_cs'])
            self.assertFalse(p['entry_pointer_in_ds'])
            self.assertTrue(p['indirect_verdict'].startswith('OPEN'))
        self.assertTrue(self.result['table_direct_cfg']['indirect_boundaries'])
        self.assertEqual(self.result['status'], 'DMO0 NOT CLOSED. DMO1 NOT STARTED.')
        self.assertTrue(self.result['expiry_interleaving']['counter_then_tasks'])
        self.assertTrue(self.result['expiry_interleaving']['reachability_verdict'].startswith('OPEN'))

    def test_input_mutation_fails_before_analysis(self):
        original = Path.read_bytes
        selected = self.paths[0]/'TABLE1.PRG'
        def changed(path):
            b = original(path)
            if path == selected:
                b = b[:0x73e]+bytes([b[0x73e] ^ 1])+b[0x73f:]
            return b
        with patch.object(Path, 'read_bytes', changed):
            with self.assertRaisesRegex(ValueError, 'not the pinned demo: TABLE1.PRG'):
                graph.audit(*self.paths)

    def test_sdr_container_and_callback_families(self):
        s = self.result['sdr_dispatch']
        self.assertEqual(len(s['inputs']), 11)
        families = {r['driver']: r for r in s['reviewed_dispatch']}
        self.assertEqual(families['NOSOUND.SDR']['callback'], 0x687)
        self.assertEqual(families['ADLIB.SDR']['vblank_callback'], 0xbdd)
        self.assertTrue(s['status'].startswith('OPEN'))
        import audit_10min_demo_sdr as sdr
        self.assertEqual({r['driver'] for r in s['api_domains']}, set(sdr.SDR))
        for r in s['api_domains']:
            self.assertEqual({p['api'] for p in r['callback_registration']}, {11, 12})
            self.assertIn('UNKNOWN', r['status'])

    def test_sdr_mutation_is_rejected(self):
        import audit_10min_demo_sdr as sdr
        original = Path.read_bytes
        selected = self.paths[0]/'NOSOUND.SDR'
        def changed(path):
            b = original(path)
            return b[:-1]+bytes([b[-1] ^ 1]) if path == selected else b
        with patch.object(Path, 'read_bytes', changed):
            with self.assertRaisesRegex(ValueError, 'not the pinned SDR: NOSOUND.SDR'):
                sdr.audit(self.paths[0])


if __name__ == '__main__':
    unittest.main()
