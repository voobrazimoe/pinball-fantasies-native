#!/usr/bin/env python3
"""Linked producer operands, fresh replay and adversarial drop/scan mutations."""
import json,os,unittest
from pathlib import Path
import audit_10min_demo_droptask2 as a

class Arithmetic(unittest.TestCase):
    def test_equation(self):self.assertEqual(a.firing(35810+130),35998)
    def test_compare_before_increment(self):
        r=a.visit_trace(35971,27);self.assertEqual(len(r),28);self.assertEqual(r[0],dict(calculation=35971,before=0,after=1,fire=False));self.assertEqual(r[-2]['after'],27);self.assertEqual(r[-1],dict(calculation=35998,before=27,after=0,fire=True))
    def test_first_visit_mutation(self):self.assertEqual(a.firing(35940,first_next=True),35999)
    def test_child_slot_mutation(self):self.assertEqual(a.firing(35940,child_next=False),35997)
    def test_limit_mutation(self):self.assertEqual(a.firing(35940,l2=28),35999)

class Concrete(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not a.OUTPUT.exists():raise unittest.SkipTest('run DROPTASK2 auditor first')
        cls.r=json.loads(a.OUTPUT.read_text());cls.w=cls.r['fresh_input_replay']['witness'];cls.ix=a.index(cls.w)
    def b(self,role,n=35998):return self.ix[(role,n)][0]
    def test_full_validate(self):self.assertTrue(a.validate(self.r))
    def test_deterministic_fresh_explicit_script(self):
        self.assertTrue(self.r['fresh_input_replay']['deterministic_equal']);self.assertEqual(self.w['script'],a.TARGET);self.assertEqual(self.b('area_before:GROPD',35810)['reset_count'],1)
    def test_real_capture_family(self):
        w=self.r['any_capture_replay'];self.assertEqual((w['capture'],w['START_DROP'],w['DROPTASK1_fire'],w['DROPTASK2_fire']),(331,461,491,519));self.assertFalse(a.boundary(w,'body_after:DROPTASK2',519)['ball']['Hold'])
    def test_capture_producer_and_lock(self):
        self.assertEqual(self.b('area_after:GROPD',35810)['live_task_sites'][:3],['cameraDrop.func1','WAIT_FOR_TUNNEL_EFFECT','LOCK_BALL_IN_TUNNEL']);self.assertTrue(self.b('body_after:LOCK_BALL_IN_TUNNEL',35812)['ball']['Hold']);self.assertFalse(self.b('body_after:LOCK_BALL_IN_TUNNEL',35812)['linked_HOLDSTILL'])
    def test_camera_child_same_scan(self):
        self.assertTrue(self.b('camera_parent_after',35810)['live_task_sites'][3].startswith('cameraDrop.func1.'));self.assertTrue(self.b('camera_child_before',35810));self.assertFalse(any(self.b('tasks_after',35810)['live_task_sites'][i] for i in (0,3)))
    def test_start_drop_slots_and_wait_origin(self):
        self.assertEqual(self.b('start_drop_after',35940)['live_task_sites'][:3],['cameraDrop.func1','WAIT_FOR_TUNNEL_EFFECT','DROPTASK1']);self.assertEqual(self.b('wait_before:DROPTASK1',35940)['waits'].get('DROPTASK1',0),0);self.assertEqual(self.b('tasks_after',35940)['waits']['DROPTASK1'],1)
    def test_drop1_fire_and_earlier_child_next_scan(self):
        self.assertEqual(self.b('wait_before:DROPTASK1',35970)['waits']['DROPTASK1'],30);self.assertEqual(self.b('body_after:DROPTASK1',35970)['live_task_sites'][:3],['DROPTASK2','','DROPTASK1']);self.assertNotIn(('wait_before:DROPTASK2',35970),self.ix);self.assertEqual(self.b('wait_after:DROPTASK2',35971)['waits']['DROPTASK2'],1)
    def test_slot_id_survival_and_no_duplicate(self):
        ids=[]
        for n in range(35971,35999):
            x=self.b('wait_before:DROPTASK2',n);self.assertEqual(x['live_task_sites'],['DROPTASK2']+['']*49);self.assertEqual(x['reset_count'],1);self.assertFalse(x['QUIT']);ids.append(x['task_ids'][0])
        self.assertEqual(len(set(ids)),1)
    def test_wait_26_27_boundary(self):self.assertEqual(self.b('wait_after:DROPTASK2',35996)['waits']['DROPTASK2'],26);self.assertEqual(self.b('wait_after:DROPTASK2',35997)['waits']['DROPTASK2'],27);self.assertEqual(self.b('body_after:DROPTASK2')['waits']['DROPTASK2'],0)
    def test_expiry_before_firing(self):self.assertEqual(self.b('electronics_before')['timer'],35997);self.assertFalse(self.b('electronics_before')['expired']);self.assertTrue(self.b('wait_before:DROPTASK2')['expired']);self.assertTrue(self.b('wait_before:DROPTASK2')['linked_HOLDSTILL'])
    def test_coordinates_velocity_high_and_source_hold(self):
        x=self.b('body_after:DROPTASK2');b=x['ball'];self.assertEqual((b['PixelX'],b['PixelY'],b['X'],b['Y'],b['VX'],b['VY'],b['High']),(15,47,15360,48128,0,52,True));self.assertTrue(b['Hold']);self.assertTrue(x['linked_HOLDSTILL']);self.assertEqual(x['SCREENFORCE'],-1)
    def test_no_HOLDSTILL_clear_in_linked_body(self):self.assertFalse(self.r['linked_evidence']['HOLDSTILL_write']);self.assertNotIn(0x3026,{x['DS'] for x in self.r['linked_evidence']['DROPTASK2_direct_writes']})
    def test_expired_matrix_cursor_preserved(self):
        for role in ('expiry_installed','body_after:DROPTASK2','tasks_after','matrix_before'):self.assertEqual(self.b(role)['linked_matrix'],dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=5));self.assertTrue(self.b(role)['expired'])
    def test_suicide_and_remaining_scan(self):self.assertFalse(any(self.b('tasks_after')['live_task_sites']));self.assertEqual(self.r['DROPTASK2_result']['remaining_tasks'],0)
    def test_first_matrix_visit(self):self.assertEqual(self.b('matrix_after')['linked_matrix'],dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=4))
    def test_same_calculation_early_and_late_physics(self):
        self.assertEqual(len(self.ix[('step_before',35998)]),3);self.assertEqual(self.b('late_before')['ball'],self.b('late_after')['ball']);self.assertFalse(any(x[0] in ('ring','response','sine') for x in self.w['equality_physics_reads']));self.assertIn(['flipper_copy',2,0],self.w['equality_physics_reads'])
    def test_concrete_no_input_continuation(self):
        x=self.r['concrete_no_input_continuation'];self.assertEqual((x['QUIT'],x['calculation'],x['new_captures'],x['new_drains'],x['expiry_restarts']),(True,37039,0,0,0));self.assertFalse(x['high_ball_enters_play']);self.assertEqual(x['final']['ball']['VY'],52)
    def test_actual_wait_mutation(self):self.assertEqual(self.r['fresh_input_replay']['wait_plus_one']['DROPTASK2_fire'],35999)
    def test_actual_slot_mutation(self):
        w=self.r['fresh_input_replay']['child_later_slot'];self.assertEqual(w['DROPTASK2_fire'],35997);self.assertEqual(a.boundary(w,'body_after:DROPTASK1',35970)['live_task_sites'][3],'DROPTASK2');self.assertTrue(a.boundary(w,'wait_before:DROPTASK2',35970))
    def test_actual_drop_state_mutation(self):
        w=self.r['fresh_input_replay']['drop_state'];x=a.boundary(w,'body_after:DROPTASK2',35998)['ball'];self.assertEqual(x['PixelX'],16);self.assertFalse(x['High'])
    def test_original_native_clear_hold_is_detected(self):
        w=self.r['fresh_input_replay']['clear_hold'];self.assertFalse(a.boundary(w,'body_after:DROPTASK2',35998)['ball']['Hold']);self.assertEqual(a.boundary(w,'late_after',35998)['ball']['Y'],48180)
    def test_tampered_wait_hold_expired_matrix_late_rejected(self):
        for role,key,val in [('wait_before:DROPTASK2','waits',{'DROPTASK2':26}),('body_after:DROPTASK2','expired',False),('body_after:DROPTASK2','linked_HOLDSTILL',False),('body_after:DROPTASK2','linked_matrix',{}),('matrix_after','linked_matrix',{}),('late_after','ball',self.b('electronics_before')['ball'])]:
            x=self.b(role);old=x[key];x[key]=val
            try:
                with self.assertRaises(ValueError):a.validate(self.r)
            finally:x[key]=old
    def test_bounded_inventory_reconsidered(self):
        self.assertEqual(len(self.r['bounded_inventory_review']['entries']),9);self.assertEqual(self.r['expiry_interleaving'],'EXPIRY_INTERLEAVING = NOT_PROVED');self.assertEqual(self.r['exactly_one_remaining_dependency']['task'],'DURINGFLASH');self.assertEqual(self.r['exactly_one_remaining_dependency']['HOLDSTILL_clear'],0x2778)

class Linked(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        ps=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(ps):raise unittest.SkipTest('private fixtures unavailable')
        b,f=a.pinned(*map(Path,ps));cls.b=b['TABLE1.PRG'];cls.full=f['TABLE1.PRG'];cls.d=a.Decoder();cls.h,_=a.identities(cls.b,Path(ps[2]))
    def test_producer_chain_correspondence(self):self.assertTrue(a.linked(self.b,self.full,self.d,self.h)['drop_chain'])
    def test_wait_word_limit_order_drop_and_suicide_mutations(self):
        for at in (0x147a,0x147d,0x149b,0x14a4,0x14a7,0x14b3,0x14b9,0x14bd,0x14c3,0x14c9,0x14cf,0x14f1,0x14f5,0x1519,0x1465,0x146b,0x151c,0x1ebd,0x1f23,0x1f3e,0x1ab85,0x5abf,0x5ad3,0x2751,0x2754,0x2779,0x27ad):
            with self.subTest(site=at):
                b=bytearray(self.b);b[at]^=1
                with self.assertRaises(ValueError):a.linked(bytes(b),self.full,self.d,self.h)
if __name__=='__main__':unittest.main()
