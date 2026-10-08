#!/usr/bin/env python3
"""Authored local tests plus pinned route mutations; no invented reachable prefix."""
import os
from pathlib import Path
import unittest
import audit_10min_demo_new_ball_threshold as a
from audit_10min_demo_graph import Decoder, pinned
from audit_10min_demo_programs import identities

class AuthoredThreshold(unittest.TestCase):
    def test_producer_relative_offset(self):
        self.assertEqual(a.producer_offset(),90)
        self.assertEqual(sum(r[3] for r in a.ROUTE),91)

    def test_producer_after_task_scan(self):
        s=a.expiry.Slice(counter=35966);s.electronics();s.add('new_ball')
        self.assertNotIn('new_ball',s.waits)
        s.electronics();self.assertEqual(s.waits['new_ball'],1)

    def test_exact_threshold_arithmetic(self):
        r=a.conditional_suffix();self.assertEqual(r['fires'],[35998])
        self.assertEqual(r['program'],'show_player');self.assertFalse(r['reachable'])

    def test_producer_35966(self):
        r=a.conditional_suffix(35966)
        self.assertEqual(r['fires'],[35997]);self.assertEqual(r['program'],'expiry')

    def test_producer_35968(self):
        r=a.conditional_suffix(35968)
        self.assertEqual(r['fires'],[]);self.assertEqual(r['wait'],30)
        self.assertEqual(r['program'],'expiry')

    def test_slot_survival_under_explicit_exclusion(self):
        r=a.conditional_suffix(slots=(None,None,None))
        self.assertEqual((r['slot'],r['fires']),(0,[35998]))

    def test_reset_clears_task(self):
        r=a.conditional_suffix(reset_at=35980)
        self.assertEqual(r['fires'],[]);self.assertEqual(r['program'],'expiry')

    def test_shared_wait_interference(self):
        r=a.conditional_suffix(interfere_at=35980)
        self.assertEqual(r['fires'],[35997])

    def test_nonzero_shared_initial_word(self):
        self.assertEqual(a.conditional_suffix(age=1)['fires'],[35997])

    def test_party_writer_changes_conditional_consequence(self):
        self.assertEqual(a.conditional_suffix(party=True)['program'],'expiry')

    def test_visa_writer_changes_conditional_consequence(self):
        self.assertEqual(a.conditional_suffix(visa=True)['program'],'expiry')

    def test_budget_phase_mutation(self):
        self.assertEqual(a.phase_candidate(35877,1)['producer'],35968)
        self.assertEqual(a.phase_candidate(35876,1)['producer'],35967)
        self.assertEqual(a.phase_candidate(35876,1)['reachable'],'UNKNOWN')

    def test_offset_duration_mutation(self):
        route=list(a.ROUTE);row=route[2];route[2]=row[:3]+(79,)
        self.assertEqual(a.producer_offset(route),89)

    def test_slot_reuse_does_not_zero_wait(self):
        s=a.expiry.Slice(tasks=[None]);s.waits['new_ball']=12
        self.assertEqual(s.add('new_ball'),0);self.assertEqual(s.waits['new_ball'],12)

    def test_duplicate_instance_consumes_shared_callsite(self):
        s=a.expiry.Slice();s.add('new_ball');s.add('new_ball');s.electronics()
        self.assertEqual(s.waits['new_ball'],2)

class PinnedThreshold(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(e) for e in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private demo/A/source required')
        cls.paths=list(map(Path,vals));demo,unused=pinned(*cls.paths)
        cls.b=demo['TABLE1.PRG'];cls.d=Decoder();cls.h,unused=identities(cls.b,cls.paths[2])
        cls.r=a.report(*cls.paths)

    def mutation(self,at):
        b=bytearray(self.b);b[at]^=1
        with self.assertRaises(ValueError):a.linked(bytes(b),self.d,self.h)

    def test_linked_path(self):a.linked(self.b,self.d,self.h)
    def test_real_predecessor_path_mutation(self):self.mutation(0x5e4)
    def test_effect_predecessor_mutation(self):self.mutation(0x19db0+0x6d5+26)
    def test_bonus_branch_mutation(self):self.mutation(0x1b467+4)
    def test_relative_wait_mutation(self):self.mutation(0x1b461+2)
    def test_scan_order_anchor_mutation(self):self.mutation(0x5d17)
    def test_slot_survival_scan_mutation(self):self.mutation(0x5eac)
    def test_reset_clearing_mutation(self):self.mutation(0x3af9)
    def test_shared_word_mutation(self):self.mutation(0xebf)
    def test_partyflash_writer_mutation(self):self.mutation(0x2595)
    def test_visakeys_writer_mutation(self):self.mutation(0x3bb2)
    def test_phase_unknown_cannot_upgrade(self):
        self.assertEqual(self.r['reachable_phase_evidence']['verdict'],'UNKNOWN')
        self.assertEqual(self.r['verdict'],'NEW_BALL_THRESHOLD_PROVENANCE = NOT_PROVED')
        self.assertIsNone(self.r['collision']['witness'])
        self.assertIsNone(self.r['collision']['exclusion'])
        self.assertEqual(self.r['smallest_remaining_dependency'],a.UNKNOWN)

if __name__=='__main__':unittest.main()
