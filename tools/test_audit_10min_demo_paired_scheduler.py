#!/usr/bin/env python3
"""Authored semantic tests plus pinned source grammar/mutation checks."""
import os
from pathlib import Path
import unittest
import audit_10min_demo_paired_scheduler as a


class AuthoredTransitions(unittest.TestCase):
    def test_impossible_nesting_without_due(self):
        for f in ('indexed','direct'):
            s = a.Scheduler(f)
            self.assertEqual(s.enter('P', due=False), 'REJECT source not due')
            self.assertIn('UNKNOWN', s.enter('P'))
            self.assertFalse(s.stack)

    def test_priority_inversion_is_not_admitted(self):
        s = a.Scheduler('indexed', ('L','P'))
        self.assertEqual(s.enter('L', True), 'DELIVER')
        self.assertEqual(s.enter('P', True), 'DROP priority')
        self.assertEqual(s.current, 0)
        self.assertEqual(s.active_priority, 200)
        s.leave(); self.assertEqual(s.active_priority, 0)

    def test_phase_inversion_is_not_admitted(self):
        s = a.Scheduler('direct')
        s.enter('P', True, 1)
        self.assertEqual(s.enter('P', True, 1), 'REJECT phase')
        s.enter('L', True, 1); s.leave()
        self.assertEqual(s.phase, 'P')
        self.assertEqual(s.enter('P', True, 1), 'DELIVER')

    def test_source_missing_l_holds_latch(self):
        s = a.Scheduler('indexed', ('L','P'))
        t = a.Table(LAST_WAS_VB=True)
        self.assertEqual(s.enter('L', due=False), 'REJECT source not due')
        self.assertEqual(t.enter('P')['stop'], 'LAST_WAS_VB set')

    def test_dropped_l_does_not_release_latch(self):
        t = a.Table(LAST_WAS_VB=True, INSIDE_BALLHANDLER=True)
        l = t.enter('L'); self.assertEqual(l['stop'], 'INSIDE_BALLHANDLER')
        self.assertTrue(t.LAST_WAS_VB)
        self.assertFalse(t.enter('P')['ball'])

    def test_nested_l_is_eligible_during_rest(self):
        t = a.Table(); t.enter('P'); t.ball_return()
        self.assertEqual(t.enter('L')['stop'], 'L body pending')
        t.later_return(); self.assertFalse(t.LAST_WAS_VB)
        self.assertTrue(t.INSIDE_RESTOFVBLANK)

    def test_nested_second_p_has_ball_but_no_electronics_during_rest(self):
        for f in ('indexed','direct'):
            r = a.conditional_counterexample(f)
            self.assertTrue(r['nested_primary']['ball'])
            self.assertFalse(r['nested_primary_rest']['electronics'])
            self.assertEqual(r['nested_primary_rest']['stop'], 'INSIDE_RESTOFVBLANK')
            self.assertIn('UNKNOWN', r['status'])

    def test_second_p_in_tail_can_reach_electronics(self):
        r = a.conditional_counterexample('indexed', tail=True)
        self.assertTrue(r['nested_primary_rest']['electronics'])
        self.assertIn('UNKNOWN', r['status'])

    def test_time_left_overwrite_before_l_busy_exit(self):
        t = a.Table(LAST_WAS_VB=True, INSIDE_RASTINT=True)
        self.assertEqual(t.enter('L',65535)['stop'], 'INSIDE_RASTINT')
        self.assertFalse(t.matrix_test())
        self.assertTrue(t.LAST_WAS_VB)
        t.enter('P',0)  # even latch-rejected P overwrites budget
        self.assertTrue(t.matrix_test())

    def test_pause_resume_latch(self):
        t = a.Table(LAST_WAS_VB=True, INTERRUPTS_ON=False)
        for k in ('P','L'):
            self.assertEqual(t.enter(k,65535)['stop'], 'INTERRUPTS_ON')
        self.assertTrue(t.TIME_LEFT); self.assertTrue(t.LAST_WAS_VB)
        t.INTERRUPTS_ON=True
        self.assertEqual(t.enter('P')['stop'], 'LAST_WAS_VB set')
        t.enter('L'); t.later_return()
        self.assertTrue(t.enter('P')['ball'])

    def test_ball_busy_prevents_second_physics(self):
        t = a.Table(INSIDE_BALLHANDLER=True)
        self.assertEqual(t.enter('P')['stop'],'INSIDE_BALLHANDLER')
        self.assertFalse(t.enter('L')['ball'])

    def test_indexed_equal_priority_restores_outer_priority(self):
        s=a.Scheduler('indexed',('P','P'))
        s.enter('P',True); s.enter('P',True)
        s.leave(); self.assertEqual(s.active_priority,100)
        s.leave(); self.assertEqual(s.active_priority,0)


class PrivateSourceTransitions(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        data=os.getenv('PF_10MIN_DEMO_DATA')
        if not data:
            raise unittest.SkipTest('private data required before Capstone import')
        cls.data=Path(data); cls.dec=a.Decoder()
        cls.r=a.audit(cls.data)

    def test_all_eleven_prefixes_and_open_edge(self):
        self.assertEqual(len(self.r['drivers']),11)
        self.assertTrue(all(not r['source_due_complete'] for r in self.r['drivers']))
        self.assertEqual(self.r['verdict'],'PAIRED_SCHEDULER_CONTRACT = NOT_PROVED')
        self.assertEqual(self.r['whole_dos_gate']['table1_unknown'],12)

    def test_source_priority_inversion_mutation(self):
        raw,_=a.unpack((self.data/'NOSOUND.SDR').read_bytes())
        changed=bytearray(raw); changed[0x670]=0x73
        with self.assertRaisesRegex(ValueError,'drift'):
            a.indexed_shape(self.dec,bytes(changed),0x687)

    def test_source_phase_inversion_mutation(self):
        raw,_=a.unpack((self.data/'INTERNAL.SDR').read_bytes())
        changed=bytearray(raw); changed[0xb23]=0x74
        with self.assertRaisesRegex(ValueError,'drift'):
            a.direct_shape(self.dec,bytes(changed),0xbdf)

    def test_source_reload_mutation(self):
        raw,_=a.unpack((self.data/'INTERNAL.SDR').read_bytes())
        changed=bytearray(raw); changed[0xbd2]^=1
        with self.assertRaisesRegex(ValueError,'reload'):
            a.direct_shape(self.dec,bytes(changed),0xbdf)

if __name__=='__main__':
    unittest.main()
