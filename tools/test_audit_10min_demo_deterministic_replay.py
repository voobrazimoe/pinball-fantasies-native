#!/usr/bin/env python3
"""Assertions on the real replay plus fail-closed consumed-path mutations."""
import json
import os
from pathlib import Path
import unittest
import audit_10min_demo_deterministic_replay as a


class DeterministicWitness(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        p=Path('/private/tmp/pf-dmo0-deterministic-bygel-drain.json')
        if not p.exists():raise unittest.SkipTest('run owner-local deterministic auditor first')
        cls.report=json.loads(p.read_text());cls.r=cls.report['deterministic_replay'];cls.w=cls.r['witness']
        cls.cp={c['boundary']:c for c in cls.w['Checkpoints']}

    def test_exact_replayed_index_and_complete_trace(self):
        self.assertTrue(a.validate_replay(self.r))

    def test_first_simple_witness_preserved(self):
        r=self.report['first_simple_witness']['witness']
        self.assertTrue(r['Valid']);self.assertEqual((r['Release'],r['Bygel'],r['Drain']),(61,255,349))
        cp={c['boundary']:c for c in r['Checkpoints']}
        self.assertFalse(cp['BYGEL2']['light39']);self.assertEqual(cp['drain_entry']['totals'],[0,0,0,0])

    def test_unlit_bygel_score_only(self):
        self.assertFalse(self.cp['BYGEL1']['light39'])
        awards=[e for row in self.w['Rows'] for e in row['state']['events'] or [] if e['Kind']=='ScoreAwarded']
        self.assertEqual([(e['Label'],e['Value']) for e in awards],[('BCD50030',50030),('LOSTBALL',0)])

    def test_zero_aggregates_through_drain(self):
        self.assertTrue(all(r['state']['totals']==[0,0,0,0] for r in self.w['Rows']))

    def test_false_xxballe_through_drain(self):
        self.assertTrue(all(not r['state']['XXBALLE'] for r in self.w['Rows']))

    def test_inh_eff_and_special_at_admission(self):
        for label in ('drain_entry','LOSTBALL_request','LOSTBALL_result'):
            self.assertFalse(self.cp[label]['INH_EFF']);self.assertFalse(self.cp[label]['SPECIALMODE'])

    def test_lostball_admitted_and_bonus_identity(self):
        self.assertTrue(self.cp['LOSTBALL_result']['effect_accepted'])
        self.assertEqual(self.report['LOSTBALL']['normal_demo_bonus_program'],0x1b459)
        self.assertEqual(self.w['Rows'][-1]['state']['matrix_pc'],212)

    def test_real_release_charge_and_clock(self):
        cp=self.cp['spring_release'];self.assertEqual(cp['calculation'],35460)
        self.assertEqual(cp['spring'],22);self.assertEqual(cp['clock_low8'],(1030*35460)&255)

    def test_one_input_mutation(self):
        self.assertEqual(self.r['mutation']['Script']['Charge'],21)
        self.assertNotEqual((self.r['mutation']['Bygel'],self.r['mutation']['Drain']),(35790,35877))

    def test_expired_false_before_pre_electronics_check(self):
        self.assertFalse(self.report['expiration']['at_drain'])
        self.assertLess(self.w['Drain'],self.report['expiration']['first_equality'])

    def test_long_down_hold_checked_dynamically(self):
        hold=self.report['release_delay_tests']['long_down_hold']
        self.assertEqual(hold['electronics'],35516);self.assertTrue(hold['all_selected_assertions'])
        self.assertEqual(hold['checkpoints'][-1]['spring'],32)

    def test_release_variation_is_actual_data(self):
        rs=self.report['release_delay_tests']['release_variation']
        self.assertEqual([(r['Release'],r['Bygel'],r['Drain']) for r in rs[:5]],
          [(132,462,549),(260,590,677),(388,718,805),(1156,1486,1573),(16516,16846,16933)])
        self.assertEqual([(r['Release'],r['Drain']) for r in rs[-5:]],
          [(35458,35645),(35459,36458),(35460,35877),(35461,35740),(35462,36249)])

    def test_neighbor_unknown_is_not_impossible(self):
        self.assertTrue(self.report['neighbors']['35876'].startswith('UNKNOWN'))
        self.assertTrue(self.report['neighbors']['35878'].startswith('UNKNOWN'))

    def test_suffix_reused_and_collision_deferred(self):
        r=self.report['reused_suffix'];self.assertEqual((r['visits'],r['producer_after'],r['conditional_NEW_BALL_TASK']),(91,35967,35998))
        self.assertEqual(r['NEW_BALL_TASK_collision_reachability'],'NOT_PROVED')

    def test_trace_index_mutation_rejected(self):
        w=self.w.copy();w['Drain']=35876
        with self.assertRaisesRegex(ValueError,'indices'):a.validate_replay({'witness':w,'mutation':self.r['mutation']})


class PrivateConsumedPath(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private A/demo/historical inputs required')
        demo,full=a.pinned(*map(Path,vals));cls.A=full['TABLE1.PRG'];cls.B=demo['TABLE1.PRG'];cls.d=a.Decoder()
        cls.r=json.loads(Path('/private/tmp/pf-dmo0-deterministic-bygel-drain.json').read_text())['deterministic_replay']

    def test_consumed_path_correspondence(self):
        rows=a.consumed_projection(self.d,self.A,self.B,self.r)
        self.assertFalse(any(r['classification']=='TRANSFER-BLOCKING-DIFFERENCE' for r in rows))

    def test_changed_wall_response_rejected(self):
        b=bytearray(self.B);b[0x8b3b]=0x90
        with self.assertRaises(ValueError):a.compare_instructions(self.d,self.A,b)

    def test_changed_ramp_parameter_rejected(self):
        b=bytearray(self.B);b[0x19db0+0x74]^=1
        with self.assertRaisesRegex(ValueError,'ramp_records'):a.consumed_projection(self.d,self.A,b,self.r)

    def test_changed_flipper_frame_rejected(self):
        b=bytearray(self.B);b[0x4c8a0]^=1
        with self.assertRaisesRegex(ValueError,'flipper_collision_frames'):a.consumed_projection(self.d,self.A,b,self.r)

    def test_changed_sine_rejected(self):
        b=bytearray(self.B);b[0x1e4b0]^=1
        with self.assertRaisesRegex(ValueError,'sine_lookup'):a.consumed_projection(self.d,self.A,b,self.r)

    def test_changed_mask_rejected(self):
        b=bytearray(self.B);b[0x3baa0]^=1
        with self.assertRaisesRegex(ValueError,'mask12'):a.consumed_projection(self.d,self.A,b,self.r)

    def test_changed_callback_binding_rejected(self):
        b=bytearray(self.B);b[0x1abcb+8]^=1
        with self.assertRaisesRegex(ValueError,'callback binding'):a.consumed_projection(self.d,self.A,b,self.r)

    def test_changed_material_rejected(self):
        b=bytearray(self.B);b[0x1c1c7+16*2]^=1
        with self.assertRaisesRegex(ValueError,'material'):a.consumed_projection(self.d,self.A,b,self.r)

    def test_changed_admission_record_rejected(self):
        b=bytearray(self.B);b[0x19db0+0xca0+2]^=1
        with self.assertRaisesRegex(ValueError,'admission'):a.consumed_projection(self.d,self.A,b,self.r)


if __name__=='__main__':unittest.main()
