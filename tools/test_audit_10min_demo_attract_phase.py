#!/usr/bin/env python3
"""Phase arithmetic, linked ordering and fail-closed fresh-prefix mutations."""
import os
from pathlib import Path
import struct
import unittest
import audit_10min_demo_attract_phase as a


class Arithmetic(unittest.TestCase):
    def test_127_128_129_callback_arithmetic_only(self):
        self.assertEqual([a.phase(n) for n in (127,128,129)],[250,0,6])
        self.assertNotEqual(a.phase(128,1),0)

    def test_conditional_wait_contribution(self):
        r=a.task_prefix()
        self.assertEqual((r['installation_visit'],r['body_visit']),(31,111))
        self.assertEqual(1030*111,114330)
        self.assertEqual(a.phase(111),154)
        self.assertIn('CONDITIONAL',r['reachability'])

    def test_phase_equation(self):
        self.assertEqual(a.solve(),[18])
        self.assertEqual(a.phase(18+111),6)
        self.assertEqual(a.phase(1+111,102),6)
        self.assertEqual(a.solve(main=1),[])

    def test_aligned_arithmetic_is_not_reachable_witness(self):
        self.assertEqual(a.phase(1+111,102),a.native_boundary()['first_spring_low_byte'])
        self.assertNotEqual(a.phase(128+111),6)


class PrivatePrefix(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        vals=[os.getenv(n) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
        if not all(vals):raise unittest.SkipTest('private demo/A/source inputs required')
        cls.paths=list(map(Path,vals));demo,_=a.pinned(*cls.paths)
        cls.b=demo['TABLE1.PRG'];cls.d=a.Decoder();cls.report=a.audit(*cls.paths)

    def test_increment_before_demomode(self):
        self.assertTrue(self.report['primary_increments']['before_demomode'])
        xs=a.linear(self.d,self.b,0x4556,0x4566,768)
        self.assertEqual([x.mnemonic for x in xs],['add','cmp','jne','jmp'])

    def test_no_start_recurrence_not_128_dwell(self):
        r=self.report['attract_dwell']
        self.assertEqual(r['initial_pending_start'],0x3b)
        self.assertFalse(r['arbitrary_no_start_dwell_proved'])
        self.assertFalse(r['no_key_is_no_start_request'])
        self.d.expect(self.b,768,0x64ca,'je','0x621b')
        self.d.expect(self.b,768,0x6439,'retf','')
        self.d.expect(self.b,768,0x64c5,'cmp','byte ptr [0x3813], 0')

    def test_start_sampling_is_later_and_main_latches_key(self):
        r=self.report['start_sampling_point']
        self.assertEqual((r['key_scan'],r['acceptance']),(0x36fa,0x64c5))
        self.assertIn('foreground',r['where'])
        self.assertIn('primary increment precedes',r['relative_to_primary'])

    def test_two_real_counter_producers(self):
        rows=self.report['concrete_writers']['concrete_counter_writers']
        self.assertEqual([(r['site'],r['delta']) for r in rows],[(0x39c5,1),(0x4556,1030)])
        self.assertFalse(self.report['phase_formula']['simple_N_only_formula_valid'])

    def test_mutation_moving_increment_after_demomode(self):
        b=bytearray(self.b)
        # Swap ADD and CMP while retaining both instructions and total width.
        b[0x4556:0x4561]=self.b[0x455c:0x4561]+self.b[0x4556:0x455c]
        with self.assertRaises(ValueError):a.linked(b,self.d)

    def test_mutation_main_counter_increment(self):
        b=bytearray(self.b);b[0x39c7]^=1
        with self.assertRaises(ValueError):a.linked(b,self.d)

    def test_mutation_automatic_start_request(self):
        b=bytearray(self.b);b[0x3a78]=0
        with self.assertRaises(ValueError):a.linked(b,self.d)

    def test_reset_writer_mutation(self):
        b=bytearray(self.b)
        # Redirect an existing one-byte new-game store onto counter low byte.
        x=self.d.instruction(b,768,0x3b53)
        at=next(x.address+768 for x in a.linear(self.d,b,0x3b53,0x3bb4,768)
                if x.mnemonic=='mov' and x.op_str.startswith('byte ptr ['))
        struct.pack_into('<H',b,at+2,a.COUNTER)
        with self.assertRaisesRegex(ValueError,'counter|SLUMP'):a.counter_inventory(b,self.d)

    def test_reset_string_writer_mutation(self):
        b=bytearray(self.b);struct.pack_into('<H',b,0x358+1,a.COUNTER)
        with self.assertRaisesRegex(ValueError,'counter'):a.counter_inventory(b,self.d)

    def test_wait_constant_mutation(self):
        for at in (0x6576,0xff5):
            b=bytearray(self.b);b[at]^=1
            with self.assertRaises(ValueError):a.linked(b,self.d)

    def test_carried_state_reset_bindings(self):
        checks=self.report['reset_state_checks']
        self.assertTrue({'score','aggregate totals','XXBALLE','INH_EFF','task list','wait list',
                         'new ball held position','SETBALL position','spring charge','BYGEL light 39'}
                        <= {r['role'] for r in checks})
        for at in (0x331,0x3ae6,0x358,0x3b3,0x3ab3,0x1003,0x1037):
            b=bytearray(self.b);x=self.d.instruction(b,768,at);b[at+x.size-1]^=1
            with self.assertRaises(ValueError):a.reset_checks(b,self.d,{})

    def test_real_prefix_gate_stays_open(self):
        r=self.report
        self.assertEqual(r['verdict'],'ATTRACT_PHASE_ALIGNMENT = NOT_PROVED')
        self.assertEqual(r['FRESH_BYGEL_ENTRY_TRANSFER'],'NOT_PROVED')
        self.assertIsNone(r['chosen_reachable_alignment_witness'])
        self.assertFalse(r['arithmetic_candidate']['reachable'])
        self.assertEqual(r['smallest_unresolved_fact'],a.MISSING)
        self.assertEqual(r['search']['native_trajectory_calculations'],0)
        self.assertEqual(r['search']['BYGEL_scripts_executed'],0)


if __name__=='__main__':unittest.main()
