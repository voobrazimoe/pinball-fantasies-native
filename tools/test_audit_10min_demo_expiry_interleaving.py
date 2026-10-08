#!/usr/bin/env python3
"""Conditional regression tests MUST NOT be reported as reachable witnesses."""
import os
from pathlib import Path
import unittest
import audit_10min_demo_expiry_interleaving as a
from audit_10min_demo_graph import Decoder, pinned


class ConditionalExpirySlice(unittest.TestCase):
    def state(self, task=None, age=0, **kw):
        s=a.Slice(**kw)
        if task: s.add(task);s.waits[task]=age
        return s

    def test_threshold_no_task(self):
        s=self.state();s.electronics()
        self.assertEqual((s.counter,s.expired,s.hold,s.program),(35998,True,True,'expiry'))

    def test_new_ball_one_visit_before_firing(self):
        s=self.state('new_ball',29);s.electronics()
        self.assertEqual((s.program,s.waits['new_ball']),('expiry',30))

    def test_new_ball_fires_on_equality(self):
        s=self.state('new_ball',30);s.electronics()
        self.assertEqual(s.program,'show_player');self.assertTrue(s.expired)

    def test_new_ball_fires_later(self):
        s=self.state('new_ball',29);s.electronics();s.electronics()
        self.assertEqual((s.counter,s.program),(35999,'show_player'))

    def test_setball_same_update(self):
        s=self.state('setball',80,hold=True);s.electronics()
        self.assertEqual((s.program,s.hold,s.expired),('expiry',False,True))

    def test_setball_later(self):
        s=self.state('setball',79,hold=True);s.electronics()
        self.assertTrue(s.hold);s.electronics();self.assertFalse(s.hold)
        self.assertEqual(s.program,'expiry')

    def test_replacement_discards_expiry_cursor(self):
        s=self.state('new_ball',29);s.electronics();s.complete_matrix_operation()
        self.assertEqual(s.cursor,1);s.electronics()
        self.assertEqual((s.program,s.cursor),('show_player',0))

    def test_partyflash_guard(self):
        s=self.state('new_ball',30,party=True);s.electronics()
        self.assertEqual(s.program,'expiry')

    def test_visakeys_guard_consumes_flag(self):
        s=self.state('new_ball',30,visa=True);s.electronics()
        self.assertEqual(s.program,'expiry');self.assertFalse(s.visa)

    def test_party_guard_precedes_visa(self):
        s=self.state('new_ball',30,party=True,visa=True);s.electronics()
        self.assertTrue(s.visa);self.assertEqual(s.program,'expiry')

    def test_unscored_party_task_sets_guard_before_reset(self):
        s=self.state('party_on',30);s.electronics()
        self.assertTrue(s.party);self.assertEqual(s.program,'expiry')

    def test_expired_release_does_not_clear_expired(self):
        s=self.state('setball',80,expired=True,hold=True,counter=36000,program='expiry')
        s.electronics();self.assertFalse(s.hold);self.assertTrue(s.expired)

    def test_expiry_quit_if_every_operation_completes(self):
        s=self.state();s.electronics()
        for unused in a.EXPIRY: s.complete_matrix_operation()
        self.assertTrue(s.quit)
        self.assertEqual([e[1] for e in s.events if e[0]=='completed'],list(a.EXPIRY))

    def test_replacement_trace_has_no_quit_not_infinite_session_proof(self):
        s=self.state('new_ball',30);s.electronics()
        for unused in range(50): s.complete_matrix_operation()
        self.assertEqual((s.program,s.cursor,s.quit),('show_player',3,False))

    def test_conditional_scored_drain_restores_entry_not_cursor(self):
        s=self.state('new_ball',30);s.electronics();s.scored_drain(True)
        self.assertEqual((s.program,s.cursor),('expiry',0))
        for unused in a.EXPIRY: s.complete_matrix_operation()
        self.assertTrue(s.quit)

    def test_rejected_replay_does_not_restore(self):
        s=self.state('new_ball',30);s.electronics();s.scored_drain(False)
        self.assertEqual(s.program,'show_player')

    def test_counter_continues_after_expiry(self):
        s=self.state();s.electronics()
        for unused in range(20):s.electronics()
        self.assertEqual(s.counter,36018);self.assertTrue(s.expired)

    def test_insertion_does_not_zero_shared_age(self):
        s=self.state();s.waits['new_ball']=17;s.add('new_ball')
        self.assertEqual(s.waits['new_ball'],17)

    def test_reset_then_first_free_reuse_scan_order(self):
        s=self.state('new_ball',30);s.electronics()
        # New-ball fires in slot 0; new tasks occupy 0,1,2. Slot 0 is past.
        self.assertNotIn('sound_new',s.waits)
        self.assertEqual((s.waits['setball'],s.waits['sound_brick']),(1,1))

    def test_matrix_producer_visit_31_at_equality_is_conditional(self):
        # Enqueue AFTER E(35967), actual producer predecessor NOT established.
        s=self.state(counter=35967);s.add('new_ball')
        for unused in range(30):s.electronics()
        self.assertEqual((s.counter,s.waits['new_ball']),(35997,30))
        s.electronics();self.assertEqual(s.program,'show_player')

    def test_no_matrix_completion_no_progress(self):
        s=self.state();s.electronics()
        for unused in range(20):s.electronics()
        self.assertEqual(s.cursor,0);self.assertFalse(s.quit)


class PrivateExpiryOperands(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private demo/A/historical source required')
        cls.paths=list(map(Path,vals));demo,unused=pinned(*cls.paths)
        cls.b=demo['TABLE1.PRG'];cls.d=Decoder();cls.r=a.report(*cls.paths)

    def test_linked_producer_wait_and_guard_operands(self):
        self.assertEqual(a.linked(self.b,self.d),len(a.ANCHORS))

    def test_mutated_shared_wait_age_operand_rejected(self):
        b=bytearray(self.b);b[0xebf]^=2
        with self.assertRaises(ValueError):a.linked(bytes(b),self.d)

    def test_mutated_setball_limit_rejected(self):
        b=bytearray(self.b);b[0xff5]=79
        with self.assertRaises(ValueError):a.linked(bytes(b),self.d)

    def test_showplayer_terminal_has_no_quit(self):
        self.assertEqual([x['op'] for x in self.r['programs']['show_player']],['_CLEAR4','_PRINT5','_PRINT5'])

    def test_expiry_graph_has_exit_zero(self):
        nodes=self.r['programs']['expiry']
        self.assertEqual([x['args'] for x in nodes if x['op']=='_WAIT'],[[100],[100]])
        self.assertEqual(nodes[-1]['args'],[0])

    def test_replay_is_guarded_not_guaranteed(self):
        self.assertFalse(self.r['replacement_edges'][1]['unconditional_restore'])

    def test_wrap_simulation_excluded_without_survival_proof(self):
        self.assertEqual(self.r['eventual_exit']['wrap_or_re_equality_reachable'],'UNKNOWN')
        self.assertIn('29538',self.r['eventual_exit']['wrap'])

    def test_re_equality_simulation_excluded_without_survival_proof(self):
        self.assertEqual(self.r['eventual_exit']['indefinite_TABLE1'],'UNKNOWN')
        self.assertIn('65536',self.r['eventual_exit']['re_equality'])

    def test_conditional_traces_cannot_upgrade_verdict(self):
        self.assertEqual(self.r['verdict'],'EXPIRY_INTERLEAVING = NOT_PROVED')
        self.assertFalse(self.r['classes_complete'])
        self.assertTrue(all(x['reachability']=='UNKNOWN_AT_FIRST_EQUALITY' for x in self.r['threshold_state_classes']))
        self.assertEqual(self.r['smallest_remaining_fact'],a.UNKNOWN)

if __name__=='__main__':unittest.main()
