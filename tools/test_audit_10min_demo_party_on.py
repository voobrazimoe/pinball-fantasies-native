#!/usr/bin/env python3
"""Concrete fresh-party replay invariants and relevant adversarial mutations."""
import copy
import json
import os
from pathlib import Path
import unittest
import audit_10min_demo_party_on as a

class PartyReplay(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not a.OUTPUT.exists():raise unittest.SkipTest('run party-on auditor')
        cls.r=json.loads(a.OUTPUT.read_text());cls.w=cls.r['replay']['witness']
    def b(self,role,n=35998):return a.boundary(self.w,role,n)
    def test_validate(self):self.assertTrue(a.validate(self.r))
    def test_same_update_first_visit(self):
        self.assertEqual(self.b('wait_before:PARTY_ON_TASK1',35968)['waits'].get('PARTY_ON_TASK1',0),0)
        self.assertEqual(self.b('wait_after:PARTY_ON_TASK1',35968)['waits']['PARTY_ON_TASK1'],1)
    def test_target_arithmetic(self):
        self.assertEqual(a.target_drain(),35968);self.assertEqual(a.target_drain(False),35967)
    def test_real_neighbor_drains(self):
        for w,d,f in zip(self.r['replay']['neighbors'],(35967,35969),(35997,35999)):
            self.assertEqual(w['drain'],d);self.assertFalse(w['scored']);self.assertTrue(a.boundary(w,'party_guard',f)['partyflash'])
    def test_exact_input_determinism(self):
        self.assertTrue(self.r['replay']['deterministic_equal']);self.assertEqual((self.w['release'],self.w['charge'],self.w['right_duration']),(35842,0,0))
        self.assertTrue(all(not x['input']['Left'] and not x['input']['Right'] for x in self.w['rows']))
    def test_fresh_prefix(self):
        self.assertTrue(self.w['prefix_score_false']);self.assertTrue(self.w['prefix_reset_one']);self.assertFalse(self.w['scored'])
        self.assertEqual(self.b('drain_entry',35968)['reset_count'],1)
    def test_old_expired_and_timer(self):
        d=self.b('drain_entry',35968);self.assertFalse(d['expired']);self.assertFalse(d['SCORECHANGED']);self.assertEqual(d['timer'],35967)
        self.assertEqual(self.b('wait_before:PARTY_ON_TASK1',35968)['timer'],35968)
    def test_direct_program_admission(self):
        d=self.b('party_admitted',35968);self.assertEqual(d['linked_matrix'],dict(program=0x1b243,cursor=0x1b245,op='_CLEAR4',remaining=5))
        self.assertTrue(any(e['Kind']=='MatrixStarted' and e['Label']=='PARTY_ONTS' for e in d['events']))
    def test_actual_slot_and_initial_word(self):
        d=self.b('party_admitted',35968);self.assertEqual(d['live_task_sites'],['PARTY_ON_TASK1']+['']*49);self.assertEqual(d['waits'].get('PARTY_ON_TASK1',0),0)
    def test_29_30_fire(self):
        self.assertEqual(self.b('wait_after:PARTY_ON_TASK1',35996)['waits']['PARTY_ON_TASK1'],29)
        self.assertEqual(self.b('wait_after:PARTY_ON_TASK1',35997)['waits']['PARTY_ON_TASK1'],30)
        self.assertEqual(self.b('wait_before:PARTY_ON_TASK1')['waits']['PARTY_ON_TASK1'],30)
        self.assertEqual(self.b('party_guard')['waits']['PARTY_ON_TASK1'],0)
    def test_single_visit_survival(self):
        for n in range(35968,35998):
            xs=[x for x in self.w['boundaries'] if x['calculation']==n and x['boundary']=='wait_before:PARTY_ON_TASK1']
            self.assertEqual(len(xs),1);self.assertEqual(xs[0]['live_task_sites'],['PARTY_ON_TASK1']+['']*49);self.assertEqual(xs[0]['reset_count'],1)
    def test_matrix_sets_party_earlier(self):
        rows={x['calculation']:x for x in self.w['rows']}
        self.assertFalse(rows[35972]['partyflash']);self.assertTrue(rows[35973]['partyflash']);self.assertEqual(rows[35973]['matrix_op'],'_PARTYONN')
        self.assertTrue(self.b('wait_before:PARTY_ON_TASK1')['partyflash'])
    def test_expiry_before_body(self):
        self.assertTrue(self.b('expiry_installed')['expired']);self.assertEqual(self.b('expiry_installed')['timer'],35998)
        roles=[x['boundary'] for x in self.w['boundaries'] if x['calculation']==35998]
        self.assertLess(roles.index('expiry_installed'),roles.index('party_guard'))
    def test_flag_before_reset(self):self.assertTrue(self.b('party_guard')['partyflash']);self.assertTrue(self.b('before_reset')['partyflash'])
    def test_preserved_cursor(self):
        for role in ('expiry_installed','before_reset','after_reset','body_after:PARTY_ON_TASK1'):
            self.assertEqual(self.b(role)['linked_matrix'],dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=5))
    def test_scan_continuation(self):
        s=next(x for x in self.w['rows'] if x['calculation']==35998)
        self.assertEqual(s['live_task_sites'][:3],['SOUNDNEWBALL','SETBALL','SOUNDBRICKUPP']);self.assertEqual(s['waits'],{'SETBALL':1,'SOUNDBRICKUPP':1});self.assertEqual(s['matrix_remaining'],4)
    def test_queued_task_fires(self):
        fires={e['Label']:x['calculation'] for x in self.w['event_rows'] for e in x['events'] if e['Kind']=='TaskReady'}
        self.assertEqual(fires,{'PARTY_ON_TASK1':35998,'SOUNDBRICKUPP':36003,'SOUNDNEWBALL':36049,'SETBALL':36078})
    def test_actual_quit_no_replacement(self):
        self.assertEqual((self.w['final']['QUIT'],self.w['final']['calculation']),(True,37039))
        programs=[(x['calculation'],e['Label']) for x in self.w['event_rows'] for e in x['events'] if e['Kind']=='MatrixStarted' and x['calculation']>=35998]
        self.assertEqual(programs,[(35998,'RESEARCH_EXPIRY')]);self.assertEqual(37039-35998+1,1042)
    def test_clear_flag_changes_guard_branch(self):
        m=self.r['replay']['clear_party'];self.assertFalse(a.boundary(m,'before_reset',35998)['partyflash']);self.assertNotEqual(a.boundary(m,'body_after:PARTY_ON_TASK1',35998)['matrix_pc'],self.b('expiry_installed')['matrix_pc'])
        self.assertTrue(any(e['Label']=='SHOWPLAYERSTS' and e['Kind']=='MatrixStarted' for x in m['event_rows'] for e in x['events']))
    def test_delayed_store_redundant_on_actual_suffix(self):
        m=self.r['replay']['delay_party'];self.assertTrue(a.boundary(m,'before_reset',35998)['partyflash']);self.assertEqual(a.boundary(m,'body_after:PARTY_ON_TASK1',35998)['matrix_pc'],self.b('expiry_installed')['matrix_pc'])
    def test_body_restores_prebody_clear(self):
        m=self.r['replay']['clear_before_body'];self.assertFalse(a.boundary(m,'wait_before:PARTY_ON_TASK1',35998)['partyflash']);self.assertTrue(a.boundary(m,'party_guard',35998)['partyflash'])
    def test_input_shift_one_calculation(self):
        self.assertEqual(self.r['replay']['neighbors'][1]['release'],35841);self.assertEqual(self.r['replay']['neighbors'][1]['drain'],35969)
        self.assertEqual(self.r['replay']['release_plus_one']['drain'],35977)
    def test_trace_tampering_fails_closed(self):
        for role,key,value in [('wait_before:PARTY_ON_TASK1','waits',{'PARTY_ON_TASK1':29}),('after_reset','matrix_remaining',4),('party_guard','partyflash',False)]:
            r=copy.deepcopy(self.r);a.boundary(r['replay']['witness'],role,35998)[key]=value
            with self.assertRaises(ValueError):a.validate(r)
    def test_no_completeness_promotion(self):
        self.assertEqual(self.r['expiry_interleaving'],'EXPIRY_INTERLEAVING = NOT_PROVED');self.assertIn('SETBALL',self.r['exactly_one_remaining_dependency'])

class LinkedParty(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        ps=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(ps):raise unittest.SkipTest('private inputs unavailable')
        b,f=a.pinned(*map(Path,ps));cls.b=b['TABLE1.PRG'];cls.full=f['TABLE1.PRG'];cls.d=a.Decoder();cls.h,_=a.identities(cls.b,Path(ps[2]))
    def test_linked_order_guards_program(self):a.linked(self.b,self.full,self.d,self.h)
    def test_mutated_linked_claims_rejected(self):
        for at in (0x46b9,0x4721,0x4724,0x5b4,0x5bd,0x5c9,0x5e84,0x5ea3,0x3b04,0x2eff,0x1b244,0x1b24b):
            with self.subTest(site=at):
                b=bytearray(self.b);b[at]^=1
                with self.assertRaises(ValueError):a.linked(bytes(b),self.full,self.d,self.h)
if __name__=='__main__':unittest.main()
