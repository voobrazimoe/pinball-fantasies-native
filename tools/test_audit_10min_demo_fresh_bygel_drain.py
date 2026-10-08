#!/usr/bin/env python3
"""Selected record/consumer tests and fail-closed entry-state mutations.

No authored state, arithmetic example, or byte match is a physical replay.
"""
import os
from pathlib import Path
import tempfile
import unittest

import audit_10min_demo_fresh_bygel_drain as a


class PrivateTransfer(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA',
                                    'PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private demo/A/source inputs required')
        cls.paths=list(map(Path,vals))
        demo,full=a.pinned(*cls.paths)
        cls.b=demo['TABLE1.PRG'];cls.c=full['TABLE1.PRG'];cls.d=a.Decoder()
        cls.report=a.audit(*cls.paths)

    def test_selected_physics_records(self):
        records=a.data_records(self.c,self.b)
        self.assertEqual(sum(r['role']=='material' for r in records),8)
        self.assertEqual(sum(r['role']=='SETBALL_lower_collision_sample' for r in records),44)
        self.assertEqual(sum(r['role']=='sine_lookup' for r in records),1)

    def test_geometry_bindings(self):
        records={r['role']:r for r in a.data_records(self.c,self.b)}
        self.assertEqual(records['BYGEL1']['demo_callback_file'],0x27f6)
        self.assertEqual(records['BYGEL2']['demo_callback_file'],0x287b)

    def test_both_score_only_callbacks(self):
        for name in ('BYGEL1_unlit','BYGEL2_unlit'):
            with self.subTest(name=name):
                self.assertEqual(a.compare_block(self.d,self.c,self.b,name)['classification'],
                                 'RELOCATED-IDENTICAL')

    def test_local_launch_state_equivalence(self):
        self.assertEqual(a.compare_block(self.d,self.c,self.b,'SETBALL')['instructions'],26)
        self.assertIsNone(self.report['fresh_initial_state']['actual_reachable_entry'])

    def test_local_release_equivalence(self):
        a.compare_block(self.d,self.c,self.b,'SPRINGUP_velocity_rotation')

    def test_drain_transition_equivalence_bounded(self):
        for name in ('drain_selection','scored_LOSTBALL_request'):
            a.compare_block(self.d,self.c,self.b,name)
        self.d.expect(self.b,768,0x5d2,'cmp','byte ptr [0x34cf], 0xff')

    def test_file_zero_is_not_reachable_entry_zero(self):
        gate=self.report['fresh_initial_state']['launch_seed_gate']
        self.assertEqual(gate['raw_file_initializer'],0)
        self.assertEqual(gate['reachable_new_game_seed'],'UNKNOWN')
        self.assertFalse(gate['permission_to_search'])

    def test_release_depends_on_missing_phase(self):
        # Arithmetic counterexample to an inference, never a gameplay input.
        self.assertNotEqual(-166*32-0,-166*32-6)
        self.assertNotEqual(0&15,6&15)

    def test_no_witness_or_neighbor_upgrade(self):
        self.assertEqual(self.report['membership'],{str(n):'UNKNOWN' for n in (35876,35877,35878)})
        self.assertEqual(self.report['search']['scripts_executed'],0)
        self.assertFalse(self.report['native_witness_authorized'])
        self.assertFalse(self.report['theorem']['proved'])
        self.assertEqual(self.report['deterministic_input_scripts'],[])
        self.assertTrue(all(v is None for v in self.report['witness_states'].values()))

    def test_selected_records_are_not_consumed_witness_records(self):
        self.assertTrue(all(not r['consumed_on_witness'] for r in
                            self.report['selected_physical_region_data_records']))

    def test_one_smallest_unresolved_fact(self):
        self.assertEqual(self.report['smallest_unresolved_fact'],a.MISSING)
        self.assertEqual(self.report['verdict'],'FRESH_BYGEL_DRAIN_PROVENANCE = NOT_PROVED')
        self.assertEqual(self.report['drain_verdict'],'DRAIN_35877_MEMBERSHIP_UNKNOWN')

    def test_mutation_material(self):
        b=bytearray(self.b);b[0x1c1c7]^=1
        with self.assertRaisesRegex(ValueError,'material'):a.data_records(self.c,b)

    def test_mutation_both_geometries(self):
        for at in (0x1abcb,0x1abd5):
            b=bytearray(self.b);b[at]^=1
            with self.assertRaisesRegex(ValueError,'geometry'):a.data_records(self.c,b)

    def test_mutation_collision_sample(self):
        row=next(r for r in a.data_records(self.c,self.b)
                 if r['role']=='SETBALL_lower_collision_sample')
        b=bytearray(self.b);b[row['demo_file']]^=1
        with self.assertRaisesRegex(ValueError,'collision'):a.data_records(self.c,b)

    def test_mutation_sine_record(self):
        b=bytearray(self.b);b[0x1e4b0]^=1
        with self.assertRaisesRegex(ValueError,'sine'):a.data_records(self.c,b)

    def test_mutation_callback_score_destination(self):
        for name,at in [('BYGEL1_unlit',0x2819),('BYGEL2_unlit',0x289b)]:
            b=bytearray(self.b);b[at]^=1
            with self.assertRaises(ValueError):a.compare_block(self.d,self.c,b,name)

    def test_mutation_launch_coordinate(self):
        b=bytearray(self.b);b[0x1007]^=1
        with self.assertRaises(ValueError):a.compare_block(self.d,self.c,b,'SETBALL')

    def test_mutation_release_phase_consumer(self):
        b=bytearray(self.b);b[0x620e]^=1
        with self.assertRaises(ValueError):a.compare_block(self.d,self.c,b,'SPRINGUP_velocity_rotation')

    def test_mutation_drain_scorechanged(self):
        b=bytearray(self.b);b[0x594]^=1
        with self.assertRaises(ValueError):a.compare_block(self.d,self.c,b,'drain_selection')

    def test_mutation_lostball_request(self):
        b=bytearray(self.b);b[0x5e4]^=1
        with self.assertRaises(ValueError):a.compare_block(self.d,self.c,b,'scored_LOSTBALL_request')

    def test_mutation_clock_update(self):
        b=bytearray(self.b);b[0x455a]^=1
        with self.assertRaises(ValueError):a.entry_gate(self.d,self.c,b)

    def test_mutation_native_clock_initialization(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)
            for name in ('game.go','presentation.go','session.go'):
                p=root/'internal/partyland'/name;p.parent.mkdir(parents=True,exist_ok=True)
                text=(a.ROOT/'internal/partyland'/name).read_text()
                if name=='game.go':text=text.replace('g := &Game{matrixTimeLeft:', 'g := &Game{clock: 6, matrixTimeLeft:',1)
                p.write_text(text)
            with self.assertRaisesRegex(ValueError,'initializer'):a.entry_gate(self.d,self.c,self.b,root)


if __name__=='__main__':unittest.main()
