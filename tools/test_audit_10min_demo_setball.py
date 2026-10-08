#!/usr/bin/env python3
"""First-equality SETBALL concrete replay, operand and adversarial checks."""
import copy
import json
import os
from pathlib import Path
import unittest
import audit_10min_demo_setball as a

class Arithmetic(unittest.TestCase):
    def test_firing_arithmetic(self):self.assertEqual(a.firing_from_drain(35888),(35918,35998))
    def test_drain_target_off_by_one(self):self.assertEqual(a.target_drain(),35888);self.assertEqual(a.target_drain(same_scan=False),35887)
    def test_mutation_drain_minus_one(self):self.assertEqual(a.firing_from_drain(35887),(35917,35997))
    def test_mutation_drain_plus_one(self):self.assertEqual(a.firing_from_drain(35889),(35919,35999))
    def test_mutation_limit(self):self.assertEqual(a.firing_from_drain(35888,setball_limit=81),(35918,35999))

class Concrete(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not a.OUTPUT.exists():raise unittest.SkipTest('run SETBALL auditor first')
        cls.r=json.loads(a.OUTPUT.read_text());cls.w=cls.r['fresh_input_replay']['witness']
    def b(self,role,n=35998):return a.boundary(self.w,role,n)
    def test_validate(self):self.assertTrue(a.validate(self.r))
    def test_deterministic_input_only_target(self):
        self.assertTrue(self.r['fresh_input_replay']['deterministic_equal']);self.assertEqual((self.w['release'],self.w['charge'],self.w['right_duration']),(35764,0,0));self.assertTrue(self.w['prefix_reset_one']);self.assertTrue(self.w['prefix_score_false'])
    def test_unscored_drain_predicate(self):
        x=self.b('drain_entry',35888);self.assertEqual(x['timer'],35887);self.assertFalse(x['expired']);self.assertFalse(x['SCORECHANGED']);self.assertTrue(x['BALL_DOWN']);self.assertFalse(any(x['live_task_sites']))
    def test_party_date_and_age(self):
        self.assertEqual(self.b('wait_before:PARTY_ON_TASK1',35918)['waits']['PARTY_ON_TASK1'],30);self.assertTrue(self.b('party_guard',35918)['partyflash'])
    def test_reset_and_queued_slots(self):
        self.assertEqual(self.b('after_reset',35918)['waits'],{});self.assertFalse(any(self.b('after_reset',35918)['live_task_sites']));self.assertEqual(self.b('body_after:PARTY_ON_TASK1',35918)['live_task_sites'][:3],['SOUNDNEWBALL','SETBALL','SOUNDBRICKUPP'])
    def test_same_scan_initial_word(self):
        self.assertEqual(self.b('wait_before:SETBALL',35918)['waits'].get('SETBALL',0),0);self.assertEqual(self.b('wait_after:SETBALL',35918)['waits']['SETBALL'],1);self.assertEqual(self.b('tasks_after',35918)['waits'].get('SOUNDNEWBALL',0),0)
    def test_actual_slot_survives_single_visits(self):
        for n in range(35918,35999):
            x=self.b('wait_before:SETBALL',n);self.assertEqual(x['live_task_sites'][1],'SETBALL');self.assertEqual(x['live_task_sites'].count('SETBALL'),1);self.assertEqual(x['reset_count'],2)
            self.assertEqual(sum(y['boundary']=='wait_before:SETBALL' and y['calculation']==n for y in self.w['boundaries']),1)
    def test_79_80_fire(self):
        self.assertEqual(self.b('wait_after:SETBALL',35996)['waits']['SETBALL'],79);self.assertEqual(self.b('wait_after:SETBALL',35997)['waits']['SETBALL'],80);self.assertEqual(self.b('body_after:SETBALL')['waits']['SETBALL'],0)
    def test_expiry_installs_before_task(self):
        self.assertEqual(self.b('electronics_before')['timer'],35997);self.assertEqual(self.b('expiry_installed')['timer'],35998);self.assertTrue(self.b('wait_before:SETBALL')['expired']);self.assertTrue(self.b('wait_before:SETBALL')['ball']['Hold'])
    def test_SETBALL_stores_and_expired(self):
        x=self.b('body_after:SETBALL');self.assertFalse(x['ball']['Hold']);self.assertFalse(x['ball']['High']);self.assertTrue(x['expired']);self.assertTrue(x['in_chute']);self.assertEqual((x['ball']['PixelX'],x['ball']['PixelY'],x['ball']['VX'],x['ball']['VY']),(297,530,10,0))
    def test_expiry_cursor_preserved(self):
        for role in ('expiry_installed','body_after:SETBALL','tasks_after','matrix_before'):self.assertEqual(self.b(role)['linked_matrix'],dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=5))
    def test_first_matrix_visit_after_setball(self):self.assertEqual(self.b('matrix_after')['linked_matrix'],dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=4))
    def test_same_calculation_late_physics(self):
        x=self.b('late_after')['ball'];self.assertEqual((x['X'],x['Y'],x['VX'],x['VY']),(304138,542720,10,7));self.assertEqual([x for x in self.w['equality_physics_reads'] if x[0]=='response'],[['response',7,1024,13,0,0]])
    def test_no_same_calculation_effect_or_drain(self):
        roles=[x['boundary'] for x in self.w['boundaries'] if x['calculation']==35998];self.assertNotIn('drain_entry',roles);self.assertNotIn('target_callback',roles)
        events=next(x['events'] for x in self.w['event_rows'] if x['calculation']==35998);self.assertEqual([e['Label'] for e in events if e['Kind']=='MatrixStarted'],['RESEARCH_EXPIRY'])
    def test_concrete_no_input_quit(self):
        x=self.r['concrete_no_input_continuation'];self.assertEqual((x['QUIT'],x['calculation'],x['additional_drains']),(True,37039,0));self.assertTrue(x['in_chute_through_QUIT']);self.assertLess(x['ball_settled_from'],36100);self.assertEqual(x['matrix_starts'],[{'calculation':35998,'label':'RESEARCH_EXPIRY'}])
    def test_real_drain_minus_one(self):
        w=self.r['fresh_input_replay']['neighbors'][0];self.assertEqual(w['drain'],35887);self.assertTrue(a.boundary(w,'body_after:SETBALL',35997));self.assertIn('UNKNOWN',self.r['neighbor_search']['D_plus_one'])
    def test_limit_mutation_actual_replay(self):
        w=self.r['fresh_input_replay']['limit_plus_one'];self.assertFalse(any(x['boundary']=='body_after:SETBALL' and x['calculation']==35998 for x in w['boundaries']));self.assertEqual(a.boundary(w,'wait_before:SETBALL',35999)['waits']['SETBALL'],81)
    def test_added_matrix_replacement_actual_replay(self):self.assertNotEqual(a.boundary(self.r['fresh_input_replay']['matrix_replace'],'body_after:SETBALL',35998)['matrix_pc'],self.b('expiry_installed')['matrix_pc'])
    def test_tampered_age_hold_expired_cursor_late_rejected(self):
        for role,key,val in [('wait_before:SETBALL','waits',{'SETBALL':79}),('body_after:SETBALL','expired',False),('body_after:SETBALL','linked_matrix',{}),('matrix_after','matrix_remaining',5),('late_after','ball',self.b('late_before')['ball'])]:
            old=self.b(role)[key];self.b(role)[key]=val
            try:
                with self.assertRaises(ValueError):a.validate(self.r)
            finally:self.b(role)[key]=old
    def test_inventory_not_promoted(self):self.assertEqual(self.r['expiry_interleaving'],'EXPIRY_INTERLEAVING = NOT_PROVED');self.assertFalse(self.r['bounded_inventory_review']['classes_complete']);self.assertIn('DROPTASK2',self.r['exactly_one_remaining_dependency'])

class Linked(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        ps=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(ps):raise unittest.SkipTest('private fixtures unavailable')
        b,f=a.pinned(*map(Path,ps));cls.b=b['TABLE1.PRG'];cls.full=f['TABLE1.PRG'];cls.d=a.Decoder();cls.h,_=a.identities(cls.b,Path(ps[2]))
    def test_consumed_linked_operands_data(self):a.linked(self.b,self.full,self.d,self.h)
    def test_operand_limit_wait_store_mutations(self):
        for at in (0xff5,0xff8,0x1004,0x100a,0x1010,0x1032,0x1038,0x103e,0x1044,0x104a,0x3baa0+20000):
            with self.subTest(site=at):
                b=bytearray(self.b);b[at]^=1
                with self.assertRaises(ValueError):a.linked(bytes(b),self.full,self.d,self.h)
if __name__=='__main__':unittest.main()
