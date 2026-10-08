#!/usr/bin/env python3
"""Fresh DURINGFLASH proof and linked/runtime adversarial mutations."""
import json,os,unittest
from pathlib import Path
import audit_10min_demo_duringflash as a

class Arithmetic(unittest.TestCase):
    def test_equation(self):self.assertEqual(a.firing(35886),35998)
    def test_capture_neighbors(self):self.assertEqual([a.firing(n) for n in (35885,35887)],[35997,35999])
    def test_next_first_visit(self):self.assertEqual(a.firing(35886,producer_next=True),35999)
    def test_next_child_visit(self):self.assertEqual(a.firing(35886,child_next=True),35999)
    def test_wait27_visit28(self):self.assertEqual(a.drop.visit_trace(35971,27)[-1],dict(calculation=35998,before=27,after=0,fire=True))
    def test_limit_mutation(self):self.assertEqual(a.firing(35886,wait=28),35999)

class Concrete(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not a.OUTPUT.exists():raise unittest.SkipTest('run DURINGFLASH audit first')
        cls.r=json.loads(a.OUTPUT.read_text());cls.w=cls.r['fresh_input_replay']['witness'];cls.ix=a.index(cls.w)
    def b(self,role,n=35998):return self.ix[(role,n)][0]
    def test_full_validation(self):self.assertTrue(a.validate(self.r))
    def test_fresh_fixed_script_twice(self):self.assertEqual(self.w['script'],a.TARGET);self.assertTrue(self.r['fresh_input_replay']['deterministic_equal'])
    def test_first_genuine_dragon(self):
        w=self.r['fresh_input_replay']['any_capture'];self.assertEqual((w['capture'],w['BEFOREFLASH_fire'],w['DURINGFLASH_fire']),(798,883,910))
    def test_producer_guards(self):
        x=self.b('area_before:GROPE',35886);self.assertFalse(any(x[k] for k in ('Dragon','FiveMillion','BallFeature','JackpotNormal','JackpotTimed','SPECIALMODE')))
    def test_source_hold_capture(self):self.assertTrue(self.b('area_after:GROPE',35886)['source_HOLDSTILL'])
    def test_parent_child_same_scan(self):self.assertEqual(self.b('body_after:BEFOREFLASH',35971)['live_task_sites'][:2],['BEFOREFLASH','DURINGFLASH']);self.assertEqual(self.b('wait_before:DURINGFLASH',35971)['waits'].get('DURINGFLASH',0),0)
    def test_shared_wait_ages(self):
        self.assertEqual(self.b('wait_before:DURINGFLASH',35997)['waits']['DURINGFLASH'],26);self.assertEqual(self.b('wait_after:DURINGFLASH',35997)['waits']['DURINGFLASH'],27);self.assertEqual(self.b('wait_before:DURINGFLASH')['waits']['DURINGFLASH'],27);self.assertEqual(self.b('body_after:DURINGFLASH')['waits']['DURINGFLASH'],0)
    def test_no_duplicate_or_replacement(self):
        xs=[self.b('wait_before:DURINGFLASH',n) for n in range(35971,35999)];self.assertEqual(len({x['task_ids'][1] for x in xs}),1);self.assertTrue(all(x['live_task_sites'].count('DURINGFLASH')==1 and x['reset_count']==1 for x in xs))
    def test_pre_firing_checkpoint(self):
        x=self.r['checkpoints']['pre_firing_35997'];self.assertEqual(x['waits']['DURINGFLASH'],27);self.assertFalse(x['expired']);self.assertTrue(x['source_HOLDSTILL'])
    def test_expiry_before_body(self):self.assertEqual(self.b('electronics_before')['timer'],35997);self.assertEqual(self.b('expiry_installed')['audio_priority'],255);self.assertTrue(self.b('wait_before:DURINGFLASH')['expired'])
    def test_release_all_fields(self):
        x=self.b('body_after:DURINGFLASH');b=x['ball'];self.assertEqual((b['X'],b['Y'],b['VX'],b['VY'],b['High'],b['Hold']),(263168,317440,-575,1575,False,False));self.assertFalse(x['source_HOLDSTILL']);self.assertFalse(x['native_capture_Hold'])
    def test_expiry_cursor_preservation(self):self.assertEqual(self.b('body_after:DURINGFLASH')['linked_matrix'],dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=5))
    def test_remaining_scan(self):self.assertFalse(any(self.b('tasks_after')['live_task_sites']));self.assertEqual(len(self.r['task_scan35998']),50)
    def test_clear5_to4(self):self.assertEqual(self.b('matrix_after')['matrix_remaining'],4)
    def test_same_calculation_movement(self):self.assertEqual((self.b('late_after')['ball']['X'],self.b('late_after')['ball']['Y'],self.b('late_after')['ball']['VY']),(262593,319015,1582))
    def test_consumed_late_ring_no_material(self):
        xs=self.r['late_physics']['consumed_reads'];self.assertEqual(sum(x[0]=='ring' for x in xs),44);self.assertFalse(any(x[0]=='response' for x in xs))
    def test_actual_collision_trajectory(self):
        xs=self.r['concrete_continuation']['collision_responses'];self.assertEqual(xs[0],['response',36027,3,80,1,0,0]);self.assertEqual(xs[-1],['response',36166,3,855,1,0,0]);self.assertEqual(len(xs),11)
    def test_actual_drain_effect_restart(self):
        x=self.b('RESEARCH_MINUTE5_result',36173);self.assertTrue(x['effect_accepted']);self.assertTrue(x['source_HOLDSTILL']);self.assertEqual(x['matrix_remaining'],5)
    def test_real_quit(self):self.assertTrue(self.w['final']['QUIT']);self.assertEqual(self.w['final']['calculation'],37212)
    def test_input_cutoff(self):self.assertFalse(any(any(x['input'].values()) for x in self.w['rows'] if x['calculation']>35886))
    def test_wait_mutation(self):self.assertEqual(self.r['fresh_input_replay']['wait_plus_one']['DURINGFLASH_fire'],35999)
    def test_slot_mutation(self):self.assertEqual(self.r['fresh_input_replay']['child_earlier_slot']['DURINGFLASH_fire'],35999)
    def test_remove_source_clear_mutation(self):
        w=self.r['fresh_input_replay']['remove_hold_clear'];x=a.boundary(w,'body_after:DURINGFLASH',35998);self.assertFalse(x['native_capture_Hold']);self.assertTrue(x['source_HOLDSTILL']);self.assertTrue(x['ball']['Hold']);self.assertEqual(a.boundary(w,'late_after',35998)['ball']['Y'],317440)
    def test_release_mutation(self):self.assertTrue(a.boundary(self.r['fresh_input_replay']['release_physics'],'body_after:DURINGFLASH',35998)['ball']['High'])
    def test_corrupt_certificate_rejected(self):
        for role,key,value in [('wait_before:DURINGFLASH','waits',{}),('body_after:DURINGFLASH','source_HOLDSTILL',True),('body_after:DURINGFLASH','expired',False),('body_after:DURINGFLASH','audio_priority',1),('matrix_after','linked_matrix',{}),('late_after','ball',self.b('late_before')['ball'])]:
            x=self.b(role);old=x[key];x[key]=value
            try:
                with self.assertRaises(ValueError):a.validate(self.r)
            finally:x[key]=old
    def test_independent_inventory_blocker(self):self.assertEqual(self.r['exactly_one_remaining_dependency']['task'],'WAIT_FOR_SPIN_TASK');self.assertFalse(self.r['bounded_inventory_review']['classes_complete'])

class Linked(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        ps=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(ps):raise unittest.SkipTest('private data not available')
        d,f=a.pinned(*map(Path,ps));cls.demo=d['TABLE1.PRG'];cls.full=f['TABLE1.PRG'];cls.dec=a.Decoder();cls.h,_=a.identities(cls.demo,Path(ps[2]))
    def test_producer_and_prefix_correspondence(self):self.assertTrue(a.linked(self.demo,self.full,self.dec,self.h)['fresh_prefix_correspondence'])
    def test_linked_mutations(self):
        for at in (0x259e,0x25e7,0x2622,0x2628,0x3ab4,0x3ab7,0x2733,0x2748,0x2751,0x2754,0x2779,0x277f,0x2785,0x278b,0x27ad,0x27b3,0x27b9,0x1abad,0x20d4,0x516):
            with self.subTest(site=hex(at)):
                b=bytearray(self.demo);b[at]^=1
                with self.assertRaises(ValueError):a.linked(bytes(b),self.full,self.dec,self.h)
    def test_real_independent_arcade_producer(self):
        iv=json.loads(a.OUTPUT.read_text())['independent_inventory_producer']['found'];self.assertEqual(iv['arcade_pending_producer'],1001);self.assertIn('arcade.func1',a.boundary(iv,'area_after:GROPB',1001)['live_task_sites'])
    def test_source_hold_store_is_present(self):self.assertIn(0x3026,{x['DS'] for x in a.linked(self.demo,self.full,self.dec,self.h)['DURINGFLASH_direct_writes']})
if __name__=='__main__':unittest.main()
