#!/usr/bin/env python3
"""Local predicate/transfer regressions; no authored gameplay witness."""
import os
from pathlib import Path
import unittest
import audit_10min_demo_scored_drain as a


class LocalDrain(unittest.TestCase):
    def test_zero_predicate(self):
        self.assertEqual(a.PREDICATE['aggregate_relation'],'==')
        self.assertEqual(a.PREDICATE['aggregate_value'],0)
        self.assertTrue(a.selected([0]*4,False))
        self.assertFalse(a.selected([1]*4,False))

    def test_each_nonzero_aggregate_rejected(self):
        for i in range(4):
            values=[0]*4;values[i]=1
            self.assertFalse(a.selected(values,False))

    def test_xxball_false_path(self):
        self.assertFalse(a.selected([0]*4,True))

    def test_zero_aggregate_preservation(self):
        r=a.unlit_lane_transfer(0,[0]*4,False)
        self.assertEqual(r['score'],50030)
        self.assertTrue(r['score_changed'])
        self.assertTrue(a.selected(r['aggregates'],r['XXBALLE']))
        self.assertEqual(r['reachable'],'UNKNOWN')

    def test_lane_does_not_clear_existing_aggregates(self):
        self.assertEqual(a.unlit_lane_transfer(0,[1,2,3,4],True)['aggregates'],(1,2,3,4))

    def test_scored_unscored(self):
        self.assertEqual(a.drain_branch(False,False),'PARTY_ON')
        self.assertEqual(a.drain_branch(True,False),'LOSTBALL')
        self.assertEqual(a.drain_branch(True,True),'MINUTE5')
        self.assertEqual(a.drain_branch(False,True),'PARTY_ON')

    def test_forced_drain_inhibitor(self):
        self.assertEqual(a.drain_branch(True,False,True),'inhibited')

    def test_lostball_admission(self):
        self.assertTrue(a.lostball_admission())
        for priority in range(256):
            self.assertTrue(a.lostball_admission(priority=priority))

    def test_suppression(self):
        self.assertFalse(a.lostball_admission(inh_eff=True))
        self.assertFalse(a.lostball_admission(special=True))

    def test_priority_reset_is_not_only_protection(self):
        self.assertTrue(a.lostball_admission(priority=255,reset_priority=False))
        with self.assertRaises(ValueError):a.lostball_admission(priority=256)

    def test_charge_saturation(self):
        charge=0
        for unused in range(40):charge=a.charge_step(charge)
        self.assertEqual(charge,32)
        self.assertEqual(a.charge_step(32),32)
        self.assertEqual(a.charge_step(12,down=False),12)

    def test_mutation_breaks_charge_recurrence(self):
        r=a.charge_only_phase_claim([35876,35877,35878],cap=33)
        self.assertFalse(r['charge_stable'])
        self.assertFalse(r['gameplay_lemma'])

    def test_consecutive_jitter_values_do_not_preserve_release(self):
        r=a.charge_only_phase_claim([100,101,102])
        self.assertTrue(r['charge_stable'])
        self.assertFalse(r['same_release'])
        self.assertFalse(r['gameplay_lemma'])

    def test_periodic_jitter_still_not_gameplay_lemma(self):
        r=a.charge_only_phase_claim([100,356,612])
        self.assertTrue(r['same_release'])
        self.assertEqual(r['reachable_indices'],'UNKNOWN')
        self.assertFalse(r['gameplay_lemma'])

    def test_release_jitter(self):
        self.assertEqual(a.release_signature(32,0),(-5312,0))
        self.assertEqual(a.release_signature(32,255),(-5567,15))

    def test_report_typo_removed(self):
        report=(Path(__file__).resolve().parents[1]/'docs/runtime-layout-demo-10min-new-ball-threshold-provenance.md').read_text()
        self.assertIn('with all four aggregates == 0 and',report)
        self.assertNotIn('with all four aggregates above zero and',report)


class PinnedDrain(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(e) for e in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private demo/A/source required')
        cls.paths=list(map(Path,vals));demo,unused=a.pinned(*cls.paths)
        cls.b=demo['TABLE1.PRG'];cls.d=a.Decoder();cls.h,unused=a.identities(cls.b,cls.paths[2])
        cls.r=a.report(*cls.paths)

    def mutation(self,at):
        b=bytearray(self.b);b[at]^=1
        with self.assertRaises(ValueError):a.linked(bytes(b),self.d,self.h)

    def test_linked(self):a.linked(self.b,self.d,self.h)
    def test_91_visits_tied_to_zero(self):
        self.assertEqual(self.r['consistency']['visits'],91)
        self.assertEqual(self.r['corrected_predicate'],a.PREDICATE)
        self.assertEqual(a.threshold.producer_offset(),90)
    def test_branch_mutation(self):self.mutation(0x593)
    def test_aggregate_reset_mutation(self):self.mutation(0x361)
    def test_score_only_lane_mutation(self):self.mutation(0x2819)
    def test_zero_lostball_bonus_mutation(self):self.mutation(0x19db0+0x6d5+14)
    def test_xxball_initializer_mutation(self):self.mutation(0x335)
    def test_effect_inh_suppression_mutation(self):self.mutation(0x5f8a)
    def test_effect_special_suppression_mutation(self):self.mutation(0x5f94)
    def test_effect_carry_mutation(self):self.mutation(0x5f7c)
    def test_effect_priority_mutation(self):self.mutation(0x19db0+0xca0+2)
    def test_phase_cap_mutation(self):self.mutation(0x6174)
    def test_phase_jitter_mutation(self):self.mutation(0x6212)
    def test_drain_detection_mutation(self):self.mutation(0x5d37)
    def test_index_neighbors_remain_unknown(self):
        self.assertEqual(self.r['reachable_calculation_indices']['checks'],dict.fromkeys(('35876','35877','35878'),'UNKNOWN'))
        self.assertEqual(self.r['reachable_calculation_indices']['conditional_producers'],{'35876':35966,'35877':35967,'35878':35968})
    def test_no_local_fact_upgrades_reachability(self):
        self.assertEqual(self.r['verdict'],'SCORED_DRAIN_35877_PROVENANCE = NOT_PROVED')
        self.assertFalse(self.r['phase_adjustment']['period_suffices_for_gameplay'])
        self.assertEqual(self.r['smallest_remaining_predecessor_fact'],a.MISSING)


if __name__=='__main__':unittest.main()
