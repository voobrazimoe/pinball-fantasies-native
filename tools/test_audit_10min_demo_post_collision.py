#!/usr/bin/env python3
"""Concrete continuations and consumed-edge mutations, never long-run inference."""
import copy
import json
import os
from pathlib import Path
import unittest
import audit_10min_demo_post_collision as a

class PostCollision(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not a.OUTPUT.exists():raise unittest.SkipTest('run post-collision auditor')
        cls.r=json.loads(a.OUTPUT.read_text());cls.c=cls.r['continuations'];cls.n=cls.c['no_input'];cls.p=cls.c['periodic_launch'];cls.s=cls.c['single_launch']

    def test_identity(self):self.assertTrue(a.validate(self.r))
    def test_age_asymmetry(self):
        s=self.r['starting_state'];self.assertEqual([s['WAITLIST'][w] for w in ('0x36d1','0x36d3','0x36cf')],[0,1,1])
    def test_task_firing(self):
        fires={e['Label']:row['calculation'] for row in self.n['event_rows'] for e in row['events'] if e['Kind']=='TaskReady'}
        self.assertEqual({k:fires[k] for k in ('SOUNDBRICKUPP','SOUNDNEWBALL','SETBALL')},dict(SOUNDBRICKUPP=36003,SOUNDNEWBALL=36049,SETBALL=36078))
    def test_setball_immediate(self):
        s=next(x for x in self.n['rows'] if x.get('boundary','').startswith('immediately after SETBALL'))
        self.assertEqual((s['ball']['PixelX'],s['ball']['PixelY'],s['ball']['VX'],s['ball']['VY']),(297,530,10,0));self.assertFalse(s['ball']['Hold']);self.assertTrue(s['in_chute']);self.assertTrue(s['expired']);self.assertEqual(s['timer'],36078)
        self.assertEqual([x for x in s['live_task_sites'] if x],['SETBALL']);self.assertFalse(s['matrix_active'])
    def test_expired_survives(self):
        self.assertTrue(all(v['expired'] for v in self.n['rows']))
    def test_show_terminator(self):
        self.assertEqual([(v['calculation'],v['op']) for v in self.n['matrix_progression'][:3]],[(35999,'_CLEAR4'),(36002,'_PRINT5'),(36004,'0')])
        self.assertEqual(36004-35998+1,7)
    def test_no_automatic_restore(self):
        starts=[row['calculation'] for row in self.n['event_rows'] for e in row['events'] if e['Kind']=='MatrixStarted' and e['Label']=='RESEARCH_EXPIRY']
        self.assertEqual(starts,[101534])
    def test_no_input_trajectory_not_seeded(self):
        self.assertEqual(self.n['start']['ball'],self.r['starting_state']['ball']);self.assertEqual(self.n['drain_boundaries'],[])
        self.assertTrue(all(not x['state']['input']['Down'] and not x['state']['input']['Release'] for x in self.n['trajectory']))
    def test_actual_unscored_drain(self):
        d=self.s['drain_boundaries'];self.assertEqual(len(d),1);self.assertEqual(d[0]['calculation'],36288);self.assertFalse(d[0]['SCORECHANGED']);self.assertTrue(d[0]['expired'])
        self.assertTrue(any(e['Label']=='PARTY_ONTS' for row in self.s['event_rows'] for e in row['events']))
        self.assertEqual([row['calculation'] for row in self.s['event_rows'] for e in row['events'] if e['Kind']=='TaskReady' and e['Label']=='PARTY_ON_TASK1'],[36318])
    def test_scored_expired_branch(self):
        d=self.p['drain_boundaries'];self.assertEqual([s['boundary'] for s in d],['drain_entry','RESEARCH_MINUTE5_request','RESEARCH_MINUTE5_result'])
        self.assertTrue(d[0]['SCORECHANGED']);self.assertTrue(d[0]['expired']);self.assertEqual(d[0]['calculation'],37084)
    def test_effect_admission(self):
        req,res=self.p['drain_boundaries'][1:];self.assertFalse(req['INH_EFF']);self.assertFalse(req['SPECIALMODE']);self.assertEqual(req['audio_priority'],255);self.assertEqual(res['audio_priority'],255);self.assertTrue(res['effect_accepted'])
    def test_reinstall_entry(self):
        s=self.p['drain_boundaries'][-1];self.assertEqual(s['matrix_op'],'_CLEAR4');self.assertEqual(s['matrix_remaining'],5);self.assertEqual(s['matrix_pc'],self.r['expiry_native_entry']+1)
    def test_second_expiry_program(self):
        ops=[s['op'] for s in self.p['matrix_progression'] if s['calculation']>=37084];self.assertEqual(ops,list(a.expiry.EXPIRY))
    def test_exact_quit(self):
        self.assertEqual((self.p['final']['calculation'],self.p['final']['QUIT']),(38125,True));self.assertEqual(38125-37084+1,1042)
    def test_no_live_replacement_tasks(self):
        self.assertTrue(all(not any(s['live_task_sites']) for s in self.p['rows'] if s['calculation']>=37084))
        self.assertEqual([e['Label'] for row in self.p['event_rows'] for e in row['events'] if row['calculation']>=37084 and e['Kind']=='MatrixStarted'],['RESEARCH_EXPIRY'])
    def test_wrap_really_relevant(self):
        rows=self.n['rows'];self.assertTrue(any(s['calculation']==65536 and s['timer']==0 for s in rows));self.assertEqual(self.n['final']['calculation'],102575)
        self.assertLess(self.p['final']['calculation'],65536)
    def test_clear_expired_runtime_mutation(self):
        v=self.c['mutation_clear_expired'];self.assertFalse(v['final']['expired']);self.assertFalse(v['final']['QUIT']);self.assertEqual(len(v['drain_boundaries']),1)
    def test_block_admission_runtime_mutation(self):
        v=self.c['mutation_block_effect'];self.assertTrue(v['drain_boundaries'][1]['expired']);self.assertTrue(v['drain_boundaries'][1]['INH_EFF']);self.assertFalse(v['drain_boundaries'][-1]['effect_accepted']);self.assertFalse(v['final']['QUIT'])
    def test_replacement_runtime_mutation(self):
        v=self.c['mutation_replace_program'];self.assertTrue(v['drain_boundaries'][-1]['effect_accepted']);self.assertTrue(v['final']['matrix_active']);self.assertEqual(v['final']['matrix_op'],'_CLEAR4');self.assertNotEqual(v['final']['matrix_pc'],self.r['expiry_native_entry']+1);self.assertFalse(v['final']['QUIT'])
    def test_finite_long_run_not_nontermination(self):
        self.assertEqual(a.classify(finite_run=dict(calculations=10**15,QUIT=False)),'EVENTUAL_TERMINATION_UNKNOWN')
    def test_real_pause_cycle(self):
        c=self.r['pause_cycle'];self.assertEqual(c['input_prefix'],['P','empty','empty']);self.assertTrue(c['gameplay_state_equal']);self.assertEqual(a.classify(cycle=c),'EVENTUAL_QUIT_NOT_GUARANTEED')
    def test_actual_alternate_exit(self):
        e=self.r['pause_cycle']['alternate_exit'];self.assertEqual(e['input'],['P','Esc','Y']);self.assertEqual(e['mode'],'Selector');self.assertTrue(e['session_unloaded'])
    def test_incomplete_cycle_rejected(self):
        for flag in ('closed_transition','gameplay_state_equal','no_exit'):
            c=dict(self.r['pause_cycle']);c[flag]=False;self.assertEqual(a.classify(cycle=c),'EVENTUAL_TERMINATION_UNKNOWN')
    def test_report_mutation_rejected(self):
        for mutation in ('mutation_clear_expired','mutation_block_effect','mutation_replace_program'):
            r=dict(self.r);r['continuations']=dict(self.c);r['continuations']['periodic_launch']=self.c[mutation]
            with self.assertRaises(ValueError):a.validate(r)
    def test_no_contract_promotion(self):self.assertEqual(self.r['expiry_interleaving'],'EXPIRY_INTERLEAVING = NOT_PROVED')

class LinkedPostCollision(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        paths=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(paths):raise unittest.SkipTest('private linked inputs')
        demo,full=a.pinned(*map(Path,paths));cls.b=demo['TABLE1.PRG'];cls.full=full['TABLE1.PRG'];cls.d=a.Decoder();cls.h,_=a.identities(cls.b,Path(paths[2]))
    def test_consumed_linked_slice(self):a.linked(self.b,self.full,self.d,self.h)
    def test_operand_mutations(self):
        for at in (0x5cdd,0x5ce1,0x5d4,0x623,0x525a,0x55aa,0x3cde,0x25ac,0x2735,0x27c1,0x1a4a1,0x1a4a1+26,0x1ba29,0x1ba35):
            with self.subTest(site=hex(at)):
                b=bytearray(self.b);b[at]^=1
                with self.assertRaises(ValueError):a.linked(bytes(b),self.full,self.d,self.h)

if __name__=='__main__':unittest.main()
