#!/usr/bin/env python3
"""Authored stack/transfer mutations, never private executable byte slices."""
import copy
import struct
import unittest

import audit_10min_demo_domains as d
import audit_10min_demo_placement as p
import audit_10min_demo_stack as s
import audit_10min_demo_writers as w


class StackTransferMutations(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        try:d.graph.Decoder()
        except ImportError:raise unittest.SkipTest('research Capstone required for authored x86 mutations')

    def fixture(self,code='5058c3',origin=0x300):
        b=bytearray(0x90200)
        struct.pack_into('<14H',b,0,0x5a4d,0,len(b)//512,0,0x20,0,0xffff,0,0x100,0,origin-0x300,0x10,0x40,0)
        b[0x310]=b[0x320]=0xc3
        raw=bytes.fromhex(code);b[origin:origin+len(raw)]=raw
        r=d.Resolver(b,0x300,0x19db0,[origin],0x700);r.prepare()
        c=p.mz(b)
        objects=[dict(name='table-a',start=0x20200,end_exclusive=0x20400),
                 dict(name='table-b',start=0x20400,end_exclusive=0x20600)]
        return r,c,objects

    def summaries(self,r,objects=None,c=None):
        scan=w.scan(r,objects or [],c)
        return s.Summaries(r,scan['writers'],'authored')

    def record(self,ss=('image',0x200),sp=0x100,objects=None,c=None):
        r,contract,objs=self.fixture()
        state=(None if ss is None else frozenset({ss}),None if sp is None else frozenset({sp}))
        return s.stack_record(r,0x300,c or contract,objects or objs,state,'authored')

    def boundary(self,callee='c3',targets=(0x310,),unknown=False):
        r,c,objs=self.fixture('ffd2c3')
        b=bytearray(r.b);raw=bytes.fromhex(callee);b[0x310:0x310+len(raw)]=raw
        b[0x330:0x333]=bytes.fromhex('8ed0c3')
        r=d.Resolver(b,0x300,0x19db0,{0x300,*targets},0x700);r.prepare()
        r.facts={0x300:dict(targets=list(targets),status='UNKNOWN' if unknown else 'BOUNDED')}
        summaries=self.summaries(r,objs,c)
        return s.transfer(r,0x300,summaries),r,c,objs

    def test_code_stack_disjoint_from_both_lookup_tables(self):
        proof=self.record()
        self.assertEqual(proof['status'],'BOUNDED')
        self.assertEqual(proof['possible_objects'],[])
        self.assertEqual(len(proof['physical_write_domain']),2)

    def test_modified_ss_possible_table_overlap_is_unknown(self):
        self.assertEqual(self.record(('image',0x20200),2)['status'],'UNKNOWN')
        self.assertEqual(self.record(None)['status'],'UNKNOWN')

    def test_modified_sp_possible_overlap_is_unknown(self):
        self.assertEqual(self.record(('image',0x20200),0x500)['status'],'BOUNDED')
        self.assertEqual(self.record(('image',0x20200),2)['status'],'UNKNOWN')
        obj=[dict(name='near',start=0x280,end_exclusive=0x284)]
        self.assertEqual(self.record(sp=0x100,objects=obj)['status'],'BOUNDED')
        self.assertEqual(self.record(sp=0x82,objects=obj)['status'],'UNKNOWN')

    def test_increased_acyclic_call_depth_can_overlap(self):
        r,c,objs=self.fixture('e80d00c3')
        b=bytearray(r.b);b[0x310:0x312]=bytes.fromhex('50c3')
        good=d.Resolver(b,0x300,0,[0x300],0x700);good.prepare()
        # Unmatched PUSH invalidates the depth proof instead of assigning a budget.
        self.assertEqual(self.summaries(good).routine(0x300)['status'],'UNKNOWN')
        b[0x310:0x313]=bytes.fromhex('5058c3')
        good=d.Resolver(b,0x300,0,[0x300],0x700);good.prepare()
        shallow=self.summaries(good).routine(0x300)
        self.assertEqual(shallow['maximum_downward_bytes'],4)
        b[0x310:0x314]=bytes.fromhex('e80d00c3');b[0x320:0x323]=bytes.fromhex('5058c3')
        deep=d.Resolver(b,0x300,0,[0x300],0x700);deep.prepare()
        deeper=self.summaries(deep).routine(0x300)
        self.assertEqual(deeper['maximum_downward_bytes'],6)
        obj=dict(name='depth-edge',start=0x2fa,end_exclusive=0x2fc)
        self.assertEqual(p.stack_exclusion(c,('image',0x200),(0x100-shallow['maximum_downward_bytes'],0x100),obj)['status'],'EXCLUDED')
        self.assertEqual(p.stack_exclusion(c,('image',0x200),(0x100-deeper['maximum_downward_bytes'],0x100),obj)['status'],'UNKNOWN')

    def test_far_call_width_and_matched_retf(self):
        r,c,objs=self.fixture('ff1e0003c3')
        x=r.seen[0x300];op=s.operation(r,x)
        self.assertEqual((op['delta'],op['width']),(-4,4))
        proof=s.stack_record(r,0x300,c,[dict(name='fourth-byte',start=0x2fc,end_exclusive=0x2fd)],
            (frozenset({('image',0x200)}),frozenset({0x100})),'authored')
        self.assertEqual(proof['status'],'UNKNOWN')
        r,_,_=self.fixture('cb')
        summary=self.summaries(r).routine(0x300,far=True)
        self.assertEqual(summary['status'],'BOUNDED')
        self.assertEqual(s.operation(r,r.seen[0x300])['delta'],4)

    def test_near_call_width_and_matched_ret(self):
        r,c,objs=self.fixture('e80d00c3')
        self.assertEqual(s.operation(r,r.seen[0x300])['width'],2)
        row=s.stack_record(r,0x300,c,[dict(name='fourth-byte',start=0x2fc,end_exclusive=0x2fd)],
            (frozenset({('image',0x200)}),frozenset({0x100})),'authored')
        self.assertEqual(row['status'],'BOUNDED')
        r,_,_=self.fixture('c3')
        self.assertEqual(s.operation(r,r.seen[0x300])['delta'],2)

    def test_pushf_popf_effects(self):
        r,c,objs=self.fixture('9c9dc3')
        summary=self.summaries(r).routine(0x300)
        self.assertEqual(summary['status'],'BOUNDED')
        self.assertEqual(summary['maximum_downward_bytes'],2)
        self.assertEqual(s.operation(r,r.seen[0x301])['width'],0)

    def test_segment_register_push_pop(self):
        r,_,_=self.fixture('1e06071fc3')
        summary=self.summaries(r).routine(0x300)
        self.assertEqual(summary['maximum_downward_bytes'],4)
        self.assertEqual(summary['effects']['ds'],'preserving')
        self.assertEqual(summary['effects']['es'],'preserving')
        r,_,_=self.fixture('1617c3')
        self.assertEqual(self.summaries(r).routine(0x300)['effects']['ss'],'preserving')

    def test_unmatched_stack_adjustment_is_unknown(self):
        for code in ('83ec02c3','83c402c3','5cc3','c20200','c8000000c9c3'):
            r,_,_=self.fixture(code)
            self.assertEqual(self.summaries(r).routine(0x300)['status'],'UNKNOWN')

    def test_transfer_preserves_ss_sp(self):
        boundary,_,_,_=self.boundary('1617c3')
        self.assertEqual(boundary['classification'],'preserving')
        self.assertEqual(boundary['effects']['ss'],'preserving')
        self.assertEqual(boundary['effects']['sp'],'preserving')

    def test_transfer_bounded_disjoint_ss_domain(self):
        boundary,r,c,objs=self.boundary('6800a017c3')
        self.assertEqual(boundary['classification'],'bounded finite effect')
        self.assertEqual(boundary['callee_summaries'][0]['ss_output'],[['literal',0xa000]])
        self.assertEqual(self.record(('literal',0xa000),objects=objs,c=c)['status'],'BOUNDED')

    def test_transfer_unknown_ss_domain_is_unknown(self):
        boundary,_,_,_=self.boundary('8ed0c3')
        self.assertEqual(boundary['classification'],'UNKNOWN')

    def test_transfer_possible_target_memory_write_is_unknown(self):
        boundary,_,_,_=self.boundary('c606000001c3')
        self.assertEqual(boundary['classification'],'UNKNOWN')
        self.assertTrue(boundary['callee_summaries'][0]['possible_objects'])

    def test_added_admitted_callee_unknown_effects_reopens_transfer(self):
        good,_,_,_=self.boundary()
        self.assertEqual(good['status'],'BOUNDED')
        changed,_,_,_=self.boundary(targets=(0x310,0x330))
        self.assertEqual(changed['status'],'UNKNOWN')
        self.assertEqual(changed['candidate_targets'],[0x310,0x330])

    def test_stack_proof_is_independent_of_consumer_addresses(self):
        for origin in (0x300,0x340,0x380):
            r,c,objs=self.fixture(origin=origin)
            row=s.stack_record(r,origin,c,objs,(frozenset({('image',0x200)}),frozenset({256})),'authored')
            self.assertEqual(row['status'],'BOUNDED')

    def test_all_legal_loader_bases(self):
        r,c,objs=self.fixture()
        self.assertEqual(self.record(c=c)['status'],'BOUNDED')
        for base in (c['load_base_domain']['minimum'],0x4321,c['load_base_domain']['maximum']):
            changed=copy.deepcopy(c);changed['load_base_domain'].update(minimum=base,maximum=base)
            self.assertEqual(self.record(c=changed)['status'],'BOUNDED')

    def test_physical_address_wrap_is_conservative(self):
        r,c,objs=self.fixture()
        c['load_base_domain'].update(minimum=16,maximum=16)
        obj=[dict(name='wrapped',start=0x200,end_exclusive=0x201)]
        row=self.record(('literal',0xffff),0x112,obj,c)
        self.assertEqual(row['status'],'UNKNOWN')
        self.assertEqual(row['physical_write_domain'][0]['destination_domains'][0]['physical_a20_disabled'],[[0x100,0x102]])

    def test_runtime_admission_is_separate_from_candidate_effect_summary(self):
        boundary,_,_,_=self.boundary(unknown=True)
        self.assertTrue(boundary['candidate_effects_complete'])
        self.assertEqual(boundary['classification'],'UNKNOWN')
        self.assertIn('runtime target admission/immutability not proved',boundary['unresolved_reason'])

    def test_recursive_unbounded_stack_growth_is_unknown(self):
        r,_,_=self.fixture('e8fdffc3')
        self.assertEqual(self.summaries(r).routine(0x300)['status'],'UNKNOWN')

    def test_interrupt_is_not_invented_and_admitted_int_stays_unknown(self):
        r,c,objs=self.fixture('cd21c3')
        summary=self.summaries(r).routine(0x300)
        self.assertEqual(summary['status'],'UNKNOWN')
        self.assertEqual(s.operation(r,r.seen[0x300])['width'],6)
        clean,_,_=self.fixture('c3')
        self.assertFalse(any(x.mnemonic=='int' for x in clean.seen.values()))


if __name__=='__main__':unittest.main()
