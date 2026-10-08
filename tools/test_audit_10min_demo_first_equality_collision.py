#!/usr/bin/env python3
"""Concrete saved handoff tests and adversarial reachable suffix mutations."""
import copy
import json
import os
from pathlib import Path
import unittest
import audit_10min_demo_first_equality_collision as a

class ConcreteCollision(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not a.OUTPUT.exists():raise unittest.SkipTest('run concrete collision auditor')
        cls.r=json.loads(a.OUTPUT.read_text());cls.h=cls.r['concrete_drain_handoff'];cls.s=cls.r['suffix']

    def test_replay_handoff(self):
        self.assertIn('identical',self.h['handoff_validation'])
        self.assertEqual(len(self.h['TASKLIST']),50)
        self.assertEqual(self.h['installation']['calculation'],35877)
        self.assertEqual(self.h['installation']['score'],50030)
        self.assertTrue(self.h['installation']['effect_accepted'])

    def test_exact_producer(self):
        self.assertTrue(a.validate_suffix(self.s))
        p=self.s['producer'];self.assertEqual(p['calculation'],35967)
        self.assertFalse(any(e['kind'] in ('NEW_BALL','reset_tasks_waits','expiry_install') and e['calculation']<35967 for e in self.s['events']))

    def test_actual_first_free(self):
        p=self.s['producer'];self.assertEqual(p['TASKLIST_before'],['']*50)
        self.assertEqual((p['slot'],p['DS_slot']),(0,0x3417))

    def test_initial_shared_word(self):
        self.assertEqual(self.h['DS_0x36cd'],0)
        self.assertEqual(self.s['producer']['DS_0x36cd'],0)

    def test_no_duplicate(self):
        self.assertEqual(self.s['producer']['duplicates'],1)
        for r in self.s['rows']:
            self.assertLessEqual(r['tasklist_before'].count('NEW_BALL_TASK'),1)

    def test_ages_one_29_30_boundary(self):
        for calc,age in [(35968,1),(35996,29),(35997,30)]:
            r=self.s['rows'][calc-35877];self.assertEqual(r['wait_after'],age)
            self.assertFalse(r['visits'][0]['fire'])
        r=self.s['rows'][-1];self.assertEqual(r['visits'][0]['before'],30)
        self.assertTrue(r['visits'][0]['fire']);self.assertEqual(r['visits'][0]['after'],0)

    def test_slot_survival_mutation(self):
        with self.assertRaisesRegex(ValueError,'slot survival'):a.validate_suffix(a.concrete_suffix(self.h,mutation='slot'))

    def test_shared_wait_mutation(self):
        with self.assertRaises(ValueError):a.validate_suffix(a.concrete_suffix(self.h,mutation='wait'))

    def test_party_writer_mutation(self):
        with self.assertRaisesRegex(ValueError,'guard writer'):a.validate_suffix(a.concrete_suffix(self.h,mutation='party'))

    def test_visa_writer_mutation(self):
        with self.assertRaisesRegex(ValueError,'guard writer'):a.validate_suffix(a.concrete_suffix(self.h,mutation='visa'))

    def test_equality_installation_before_scan_and_fire(self):
        r=self.s['rows'][-1];k=[e['kind'] for e in r['events']]
        self.assertLess(k.index('expiry_install'),k.index('KEYTASK'))
        self.assertLess(k.index('KEYTASK'),k.index('task_fire'))
        self.assertEqual(next(e for e in r['events'] if e['kind']=='expiry_install')['timer'],35998)

    def test_show_replacement(self):
        r=self.s['rows'][-1]
        e=next(e for e in r['events'] if e['kind']=='SHOWPLAYERSTS')
        self.assertEqual((e['program_before'],e['program_after']),(0x1ba17,0x1b88e))
        self.assertTrue(r['expired']);self.assertTrue(r['HOLDSTILL'])
        self.assertFalse(r['BALL_DOWN']);self.assertFalse(r['LOOSING']);self.assertTrue(r['I_UTSKJUT'])

    def test_old_cursor_not_retained(self):
        e=next(e for e in self.s['events'] if e['kind']=='SHOWPLAYERSTS')
        self.assertIsNone(e['saved_expiry_cursor'])
        self.assertFalse(self.s['old_expiry_cursor_retained'])
        self.assertEqual(self.s['post_collision']['cursor'],0x1b890)

    def test_producer_shift_moves_firing(self):
        for producer,fire in [(35966,35997),(35968,35999)]:
            s=a.concrete_suffix(self.h,producer)
            self.assertEqual([e['calculation'] for e in s['events'] if e['kind']=='task_fire' and e['task']=='NEW_BALL_TASK'],[fire])
            self.assertFalse(s['collision'])

    def test_guards_timeline(self):
        self.assertTrue(all(not r['PARTYFLASH_after'] and not r['VISAKEYS_after'] for r in self.s['rows']))
        self.assertIn('MUSICOK',self.r['guard_correction'])

    def test_actual_native_reset_correspondence(self):
        ref=self.r['native_task_reference']
        for n in (35967,35968,35996,35997,35998):
            native=next(r for r in ref if r.get('calculation')==n)
            projected=self.s['rows'][n-35877]
            self.assertEqual(native['live_task_sites'],projected['tasklist_after'])
            self.assertEqual(native['waits'].get('NEW_BALL_TASK',0),projected['wait_after'])
        final=self.r['final_post_collision_state']
        self.assertEqual(final['reset_count'],2)
        self.assertFalse(final['SCORECHANGED'])
        self.assertEqual((final['ball']['PixelX'],final['ball']['PixelY']),(282,530))
        self.assertEqual(final['demo_program'],0x1b88e)
        self.assertEqual(final['matrix_remaining'],4)

    def test_no_eventual_exit_upgrade(self):
        self.assertEqual(self.r['expiry_interleaving'],'EXPIRY_INTERLEAVING = NOT_PROVED')
        self.assertEqual(self.r['eventual_termination'],'NOT INVESTIGATED')

class LinkedCollision(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private linked inputs required')
        demo,_=a.pinned(*map(Path,vals));cls.b=demo['TABLE1.PRG'];cls.d=a.Decoder();cls.h,_=a.identities(cls.b,Path(vals[2]))

    def test_linked_consumers(self):a.linked(self.b,self.d,self.h)

    def test_linked_mutations(self):
        for at in (0xebf,0x5eac,0x5ead,0x2595,0x259a,0x3b14,0x3b1e,0x5cf0,0x5cfe,0x785,0x1b533,0x3ab7,0x1b539,0x1b88e+2):
            with self.subTest(site=hex(at)):
                b=bytearray(self.b);b[at]^=1
                with self.assertRaises(ValueError):a.linked(bytes(b),self.d,self.h)

if __name__=='__main__':unittest.main()
