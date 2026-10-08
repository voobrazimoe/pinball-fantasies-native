#!/usr/bin/env python3
"""Reference schedule inheritance, never physical IRQ uniqueness."""
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import audit_10min_demo_canonical_a_timing as a

class ReferenceSchedule(unittest.TestCase):
    def test_case_a_native_order_selects_one_visit(self):
        x,y=[t['final'] for t in a.pair_a()['traces']]
        self.assertEqual((x['animation_remaining'],y['animation_remaining']),(1,2))
    def test_countdown_difference_persists(self):
        x,y=[t['final'] for t in a.pair_a()['traces']]
        a.matrix_step(x);a.matrix_step(y)
        self.assertTrue(x['frame_transition_due']);self.assertFalse(y['frame_transition_due'])
    def test_zero_budget_is_shared_convention(self):
        self.assertEqual(*[t['final'] for t in a.pair_a(0)['traces']])
    def test_preexpiry_does_not_replace_schedule(self):
        for counter in (0,1,35996,35998,65535):
            value,expired,order=a.timer_preexpiry(counter)
            self.assertFalse(expired);self.assertEqual(order,['electronics','tasks','matrix'])
    def test_exact_expiry_equality(self):
        self.assertEqual(a.timer_preexpiry(35997)[:2],(35998,True))
    def test_native_extracted_order(self):
        r=a.native_order();self.assertIn('budget-gated matrix/animation (BeforeLate)',r['ordinary'])
        self.assertIn('A0',r['case_A_choice'])
    def mutated_native(self,path,old,new):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)
            for f in ('internal/partyland/game.go','internal/physics/ball.go','internal/partyland/timing.go'):
                p=root/f;p.parent.mkdir(parents=True,exist_ok=True);p.write_text((a.ROOT/f).read_text())
            p=root/path;s=p.read_text();self.assertIn(old,s);p.write_text(s.replace(old,new,1))
            with self.assertRaisesRegex(ValueError,'native'):a.native_order(root)
    def test_mutation_native_matrix_before_electronics(self):
        self.mutated_native('internal/physics/ball.go','g.AfterTargets(input)','g.BeforeLate()')
    def test_mutation_native_budget_capture(self):
        self.mutated_native('internal/partyland/game.go','g.matrixTimeLeft = timeLeft','g.matrixTimeLeft = true')
    def test_mutation_native_default_budget(self):
        self.mutated_native('internal/partyland/game.go','return g.SyncWithMatrixBudget(input, true)','return g.SyncWithMatrixBudget(input, false)')

class PrivateInheritance(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private demo/A/source inputs required')
        cls.paths=list(map(Path,vals));demo,full=a.pinned(*cls.paths)
        cls.b=demo['TABLE1.PRG'];cls.c=full['TABLE1.PRG'];cls.d=a.Decoder()
        cls.result=a.audit(*cls.paths)
    def mutate(self,at,value):
        b=bytearray(self.b);b[at:at+len(value)]=value
        with self.assertRaises(ValueError):a.match(self.d,self.c,bytes(b))
    def test_all_callback_blocks(self):
        self.assertEqual(len(a.match(self.d,self.c,self.b)),8)
    def test_time_left_before_guards(self):
        for name,at,op in [('primary',0x4537,'or'),('later',0x5956,'or')]:
            self.assertEqual(self.d.instruction(self.b,768,at).mnemonic,op)
        self.mutate(0x4537,b'\x09\xdb') # OR BX,BX changes budget source.
    def test_latch_after_ball_before_rest(self):
        self.d.expect(self.b,768,0x4714,'mov','byte ptr cs:[0x449c], 0xff')
        self.mutate(0x4719,b'\x00')
    def test_later_latch_clear(self):
        self.d.expect(self.b,768,0x5a69,'mov','byte ptr cs:[0x449c], 0')
        self.mutate(0x5a6e,b'\xff')
    def test_matrix_budget_placement(self):
        self.d.expect(self.b,768,0x472a,'cmp','byte ptr [0x37f1], 0')
        self.mutate(0x472f,b'\x75')
    def test_animation_countdown_operand(self):
        self.d.expect(self.c,768,0x72c2,'dec','word ptr [0x42a]')
        self.d.expect(self.b,768,0x7330,'dec','word ptr [0x42a]')
        self.mutate(0x7332,b'\x2b')
    def test_mutation_moving_matrix_visit(self):
        b=bytearray(self.b)
        b[0x4723:0x4726],b[0x4746:0x4749]=b[0x4746:0x4749],b[0x4723:0x4726]
        with self.assertRaises(ValueError):a.match(self.d,self.c,bytes(b))
    def test_mutation_budget_constant(self):self.mutate(0x4547,b'\xff')
    def test_timer_at_electronics(self):self.assertEqual(len(a.interference(self.d,self.b)),6)
    def test_mutation_timer_insertion(self):
        b=bytearray(self.b);b[0x5cdb]=0x90
        with self.assertRaises(ValueError):a.interference(self.d,bytes(b))
    def test_expiry_branch_skips_all_replacement(self):
        self.d.expect(self.b,768,0x5ce5,'jne','0x5a00')
        b=bytearray(self.b);b[0x5ce5]=0x74
        with self.assertRaises(ValueError):a.interference(self.d,bytes(b))
    def test_sdr_shared_and_relocated(self):
        rows=self.result['scheduler_ABI']
        self.assertEqual(sum(r['packed_bytes_equal'] for r in rows),9)
        self.assertEqual({r['driver'] for r in rows if not r['packed_bytes_equal']},{'PAS16.SDR','SB16.SDR'})
        self.assertTrue(all(r['classification'] in ('IDENTICAL','RELOCATED-IDENTICAL') for r in rows))
    def test_registration_priorities(self):
        self.d.expect(self.b,768,0x6348,'mov','bl, 0x64')
        self.d.expect(self.b,768,0x6364,'mov','bl, 0xc8')
    def test_native_audio_reference_only(self):
        self.assertEqual(self.result['native_audio_boundary'],'NATIVE_AUDIO_BOUNDARY = PROVED')
        self.assertEqual(self.result['whole_DOS_gate']['exit'],2)
    def test_inheritance_and_historical_boundary(self):
        self.assertEqual(self.result['verdict'],'CANONICAL_A_TIMING_INHERITANCE = PROVED')
        self.assertEqual(set(self.result['historical_unknowns'].values()),{'BELOW-NATIVE-REFERENCE-BOUNDARY'})
if __name__=='__main__':unittest.main()
