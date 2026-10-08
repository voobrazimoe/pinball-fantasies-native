#!/usr/bin/env python3
"""Authored semantic alternatives are not physical scheduling witnesses."""
import os
from pathlib import Path
import unittest
import audit_10min_demo_scheduler_confluence as a

class AuthoredConfluence(unittest.TestCase):
    def finals(self,p):return [t['final'] for t in p['traces']]
    def test_l_before_after_zero_budget_converges(self):
        x,y=self.finals(a.pair_a(0));self.assertEqual(x,y)
    def test_l_before_after_crisis_budget_diverges(self):
        x,y=self.finals(a.pair_a(65535));self.assertEqual((x['animation_remaining'],y['animation_remaining']),(1,2))
    def test_last_was_vb_equal_after_l_sync(self):
        for ax in (0,65535):
            x,y=self.finals(a.pair_a(ax));self.assertEqual(x['table'],y['table']);self.assertFalse(x['table']['LAST_WAS_VB'])
    def test_a_counter_equality(self):
        x,y=self.finals(a.pair_a());self.assertEqual(x['electronics'],y['electronics']);self.assertEqual(x['sync'],y['sync'])
    def test_a_busy_equality(self):
        x,y=self.finals(a.pair_a())
        for f in ('INSIDE_BALLHANDLER','INSIDE_RESTOFVBLANK','INSIDE_RASTINT'):
            self.assertFalse(x['table'][f]);self.assertEqual(x['table'][f],y['table'][f])
    def test_a_gameplay_task_program_equal_at_first_difference(self):
        x,y=self.finals(a.pair_a())
        for f in ('gameplay','task_advancements','program_cursor','printtask','ball_updates','flags'):
            self.assertEqual(x[f],y[f])
    def test_next_equal_visit_exposes_frame_phase(self):
        x,y=self.finals(a.pair_a());a.matrix_step(x);a.matrix_step(y)
        self.assertTrue(x['frame_transition_due']);self.assertFalse(y['frame_transition_due'])
    def test_zero_wrap_not_collapsed(self):
        x,y=self.finals(a.pair_a(65535,remaining=0));self.assertEqual((x['animation_remaining'],y['animation_remaining']),(65535,0))
    def test_same_p_before_after_return_zero(self):self.assertTrue(a.pair_b(0)['projected_equal'])
    def test_same_p_before_after_return_crisis(self):self.assertTrue(a.pair_b(65535)['projected_equal'])
    def test_b_same_event_not_extra_p(self):
        for ax in (0,65535):
            x,y=self.finals(a.pair_b(ax));self.assertEqual(x['electronics'],2);self.assertEqual(x['source_events'],['P0']);self.assertEqual(x,y)
    def test_b_latch_counter_busy_matrix_equal(self):
        x,y=self.finals(a.pair_b());self.assertEqual(x,y);self.assertTrue(x['table']['LAST_WAS_VB'])
        self.assertEqual(x['sync'],2);self.assertEqual(x['animation_remaining'],7)
        for f in ('INSIDE_BALLHANDLER','INSIDE_RESTOFVBLANK','INSIDE_RASTINT'):self.assertFalse(x['table'][f])
    def test_b_repeated_finite_word_normalization(self):
        for n in range(1,5):self.assertTrue(a.pair_b(callbacks=n)['projected_equal'])
    def test_tail_native_mutation_breaks_diamond(self):
        self.assertFalse(a.pair_b(tail_native_write=True)['projected_equal'])
    def test_matrix_mutation_persistent_suppression(self):
        # Mutation of the semantic projection to disposable rendering would hide
        # the difference; retaining the checked persistent counter detects it.
        self.assertEqual(*self.finals(a.pair_a(persistent=False)))
        self.assertNotEqual(*self.finals(a.pair_a(persistent=True)))
    def test_p_tail_reorder_in_bounded_projection(self):
        self.assertEqual(*self.finals(a.pair_b()))

class PrivateConfluence(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        data=os.getenv('PF_10MIN_DEMO_DATA')
        if not data:raise unittest.SkipTest('private demo required')
        cls.b=(Path(data)/'TABLE1.PRG').read_bytes();cls.d=a.Decoder();cls.data=data
    def test_decoded_tail_reorder_every_instruction_boundary(self):
        a.tail_certificate(self.d,self.b)
        for cut in range(38):
            with self.subTest(cut=cut):self.assertEqual(*a.tail_reorder(self.d,self.b,cut))
    def test_pinned_dependency(self):self.assertIn('2 -> 1',a.dependency_certificate(self.d,self.b)['first_persistent_difference'])
    def test_all_tail_operations_classified(self):
        r=a.tail_certificate(self.d,self.b);self.assertEqual(len(r['operations']),38);self.assertEqual(r['native_writes'],[])
    def test_tail_instruction_native_memory_mutation(self):
        b=bytearray(self.b);b[0x475a:0x475d]=bytes([0xa3,0x4a,0x36])
        with self.assertRaisesRegex(ValueError,'native state'):a.tail_certificate(self.d,bytes(b))
    def test_persistent_countdown_operand_mutation(self):
        b=bytearray(self.b);b[0x7332]^=1
        with self.assertRaisesRegex(ValueError,'drift'):a.dependency_certificate(self.d,bytes(b))
    def test_matrix_suppression_branch_mutation(self):
        b=bytearray(self.b);b[0x472f]=0x75
        with self.assertRaisesRegex(ValueError,'drift'):a.dependency_certificate(self.d,bytes(b))
    def test_no_full_dos_tail_theorem_from_empty_footprint(self):
        self.assertIn('NOT_PROVED',a.tail_certificate(self.d,self.b)['full_DOS_diamond'])
    def test_artifact_verdict(self):self.assertEqual(a.audit(self.data)['classification'],'REFERENCE_TIMING_REQUIRED')
if __name__=='__main__':unittest.main()
