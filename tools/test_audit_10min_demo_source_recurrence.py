#!/usr/bin/env python3
"""Authored source-clock cases; no authored interval is a DOS timing witness."""
import os
from pathlib import Path
import unittest
import audit_10min_demo_source_recurrence as a
from audit_10min_demo_paired_scheduler import Scheduler, Table


class AuthoredSourceRecurrence(unittest.TestCase):
    def test_permission_without_source_ticks_is_not_due(self):
        p=a.PIT(3);p.early_eoi_sti()
        self.assertFalse(p.accept())
        p.tick(2);self.assertFalse(p.accept())

    def test_due_during_outstanding_outer(self):
        p=a.PIT(3);s=Scheduler('indexed')
        s.enter('P',True);p.tick(3)
        self.assertTrue(p.accept());p.early_eoi_sti()
        self.assertEqual(s.enter('L',True),'DELIVER')
        self.assertEqual(s.stack,[0,100])

    def test_l_due_but_following_p_not_due(self):
        p=a.PIT(2);s=Scheduler('indexed')
        s.enter('P',True);p.tick(2);self.assertTrue(p.accept())
        p.early_eoi_sti();p.program(5);s.enter('L',True);s.leave()
        p.tick(4);self.assertFalse(p.accept());self.assertEqual(s.active_priority,100)

    def test_p_due_while_rest_busy(self):
        p=a.PIT(1);t=Table(LAST_WAS_VB=False,INSIDE_RESTOFVBLANK=True)
        p.tick();self.assertTrue(p.accept());p.early_eoi_sti()
        self.assertTrue(t.enter('P')['ball']);self.assertFalse(t.ball_return()['electronics'])

    def test_p_due_after_clear(self):
        p=a.PIT(1);t=Table(INSIDE_RESTOFVBLANK=True)
        t.rest_return();p.tick();self.assertTrue(p.accept())
        p.early_eoi_sti();t.enter('P');self.assertTrue(t.ball_return()['electronics'])
        self.assertEqual(a.temporal_relation(4,3,5),'IN_TAIL')

    def test_mask_keeps_pending_but_prevents_entry(self):
        p=a.PIT(1,masked=True);p.tick();self.assertTrue(p.pending)
        self.assertFalse(p.accept());p.masked=False;self.assertTrue(p.accept())

    def test_stopped_source_cannot_create_due(self):
        p=a.PIT(1);p.running=False;p.tick(100);self.assertFalse(p.accept())

    def test_if_and_isr_exclude_pending_entry(self):
        for field in ('interrupts','in_service'):
            p=a.PIT(1);p.tick();setattr(p,field,field=='in_service')
            self.assertFalse(p.accept());p.early_eoi_sti();self.assertTrue(p.accept())

    def test_indexed_deadline_advancement_and_rejected_record(self):
        p=a.PIT(1);s=Scheduler('indexed',('L','P'))
        s.enter('L',True);p.program(a.indexed_reload(30,20,3))
        p.tick(2);self.assertFalse(p.accept());p.tick();self.assertTrue(p.accept())
        p.early_eoi_sti();self.assertEqual(s.enter('P',True),'DROP priority')
        self.assertEqual(s.current,0)

    def test_indexed_priority_restoration_keeps_source_progress(self):
        s=Scheduler('indexed',('P','L','P'));p=a.PIT(1)
        s.enter('P',True);p.tick();p.accept();p.early_eoi_sti()
        p.program(2);s.enter('L',True);p.tick();s.leave()
        self.assertEqual((s.active_priority,s.current),(100,2))
        self.assertEqual(p.remaining,1);p.tick();self.assertTrue(p.accept())
        self.assertEqual(s.enter('P',True),'DELIVER')

    def test_indexed_zero_word_is_long_not_immediate(self):
        p=a.PIT(a.indexed_reload(20,10,0))
        p.tick(65535);self.assertFalse(p.accept());p.tick();self.assertTrue(p.accept())
        self.assertEqual(a.indexed_reload(10,20,1),65517)

    def test_mode0_does_not_repeat_without_rearm(self):
        p=a.PIT(1);p.tick();p.accept();p.early_eoi_sti();p.tick(100)
        self.assertFalse(p.accept());p.program(1);p.tick();self.assertTrue(p.accept())

    def test_direct_nonzero(self):
        p=a.PIT(1,True);s=a.Direct(2);p.tick()
        self.assertEqual(s.raw_entry(p),'NONZERO');self.assertEqual(s.countdown,1)

    def test_direct_zero_result_selects_phase(self):
        p=a.PIT(1,True);s=a.Direct(1);p.tick()
        self.assertEqual(s.raw_entry(p),'P');self.assertEqual(s.countdown,0)

    def test_direct_loaded_zero_wraps(self):
        p=a.PIT(1,True);s=a.Direct(0);p.tick()
        self.assertEqual(s.raw_entry(p),'NONZERO');self.assertEqual(s.countdown,65535)

    def test_direct_phase_publication_alone_does_not_expire(self):
        p=a.PIT(1,True);s=a.Direct(1);p.tick();self.assertEqual(s.raw_entry(p),'P')
        s.publish_phase();p.tick();self.assertEqual(s.raw_entry(p),'NONZERO')
        self.assertEqual((s.phase,s.countdown),('L',65535))
        s.reload(2);p.tick();self.assertEqual(s.raw_entry(p),'NONZERO')
        p.tick();self.assertEqual(s.raw_entry(p),'L')

    def test_direct_p_l_p_requires_two_expiries(self):
        p=a.PIT(1,True);s=a.Direct(1);p.tick();self.assertEqual(s.raw_entry(p),'P')
        s.publish_phase();s.reload(2)
        p.tick();self.assertEqual(s.raw_entry(p),'NONZERO')
        p.tick();self.assertEqual(s.raw_entry(p),'L')
        s.publish_phase();s.reload(3)
        for _ in range(2):p.tick();self.assertEqual(s.raw_entry(p),'NONZERO')
        p.tick();self.assertEqual(s.raw_entry(p),'P')

    def test_direct_inactive_and_disabled_differ(self):
        p=a.PIT(1,True);s=a.Direct(1,active=False);p.tick()
        self.assertEqual(s.raw_entry(p),'INACTIVE_ZERO_RETAINED')
        self.assertEqual(s.countdown,0)
        p.tick();self.assertEqual(s.raw_entry(p),'NONZERO')
        s=a.Direct(1,enabled=False);p.tick()
        self.assertEqual(s.raw_entry(p),'DISABLED_RELOAD_50');self.assertEqual(s.countdown,50)

    def test_raw_edges_coalesce_while_masked(self):
        p=a.PIT(1,True,masked=True);s=a.Direct(5);p.tick(20)
        self.assertEqual(s.raw_entry(p),'NO_SOURCE_ENTRY');self.assertEqual(s.countdown,5)
        p.masked=False;self.assertEqual(s.raw_entry(p),'NONZERO');self.assertEqual(s.countdown,4)

    def test_reprogram_does_not_discard_pending_irq(self):
        p=a.PIT(1);p.tick();p.program(100);self.assertTrue(p.accept())

    def test_timing_unknown_cannot_be_promoted(self):
        self.assertEqual(a.temporal_relation(None,3,5),'UNKNOWN')
        self.assertEqual(a.temporal_relation(4,None,5),'UNKNOWN')
        self.assertEqual(a.temporal_relation(4,3,None),'UNKNOWN')
        # Authored early/late clocks only exercise classifier, not a derived bound.
        self.assertEqual(a.temporal_relation(2,3,5),'BEFORE_CLEAR')
        self.assertEqual(a.temporal_relation(5,3,5),'NOT_BEFORE_RETURN')


class PrivateSourceProducer(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        data=os.getenv('PF_10MIN_DEMO_DATA')
        if not data:raise unittest.SkipTest('private fixture required before Capstone import')
        cls.data=Path(data);cls.d=a.Decoder();cls.r=a.audit(cls.data)

    def row_raw(self,name):
        q=next(r['paired_premises'] for r in self.r['drivers'] if r['driver']==name)
        raw,_=a.unpack((self.data/name).read_bytes());return q,bytearray(raw)

    def test_all_producers_and_temporal_gate(self):
        self.assertEqual(len(self.r['drivers']),11)
        self.assertEqual(self.r['critical_tail']['paths'],1)
        self.assertTrue(all(r['critical_tail_verdict']=='UNKNOWN' for r in self.r['drivers']))
        self.assertEqual(self.r['real_feasible_traces'],[])
        self.assertEqual(self.r['verdict'],'SOURCE_DUE_RECURRENCE = NOT_PROVED')

    def test_eoi_mutation(self):
        q,b=self.row_raw('NOSOUND.SDR');b[0x5c4]=0x21
        with self.assertRaisesRegex(ValueError,'drift'):a.indexed_source(self.d,bytes(b),q)

    def test_deadline_compensation_mutation(self):
        q,b=self.row_raw('NOSOUND.SDR');b[0x63b]=11
        with self.assertRaisesRegex(ValueError,'drift'):a.indexed_source(self.d,bytes(b),q)

    def test_mode_mutation(self):
        q,b=self.row_raw('INTERNAL.SDR');b[0x1565]=0x30
        with self.assertRaisesRegex(ValueError,'drift'):a.direct_source(self.d,bytes(b),q)

    def test_countdown_target_mutation(self):
        q,b=self.row_raw('INTERNAL.SDR');b[q['countdown_decrement']+3]^=1
        with self.assertRaisesRegex(ValueError,'drift'):a.direct_source(self.d,bytes(b),q)

    def test_disabled_reload_mutation(self):
        q,b=self.row_raw('INTERNAL.SDR');b[0xb01]=49
        with self.assertRaisesRegex(ValueError,'drift'):a.direct_source(self.d,bytes(b),q)

    def test_gus_temporary_mask_mutation(self):
        q,b=self.row_raw('GUS.SDR');b[0x6ea]=0xfc
        with self.assertRaisesRegex(ValueError,'drift'):a.indexed_source(self.d,bytes(b),q)

    def test_tail_call_mutation(self):
        b=bytearray((self.data/'TABLE1.PRG').read_bytes());b[0x475a:0x475d]=b'\xe8\x00\x00'
        with self.assertRaisesRegex(ValueError,'tail control'):a.tail_shape(self.d,bytes(b))

if __name__=='__main__':unittest.main()
