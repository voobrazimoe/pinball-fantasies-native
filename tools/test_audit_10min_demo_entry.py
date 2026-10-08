#!/usr/bin/env python3
"""Authored entry/handler mutations; no private payload or root allowlist."""
import copy
import struct
import unittest
import audit_10min_demo_domains as d
import audit_10min_demo_stack as s
import audit_10min_demo_placement as p
import audit_10min_demo_writers as w
import audit_10min_demo_entry as e


class EntryHandlerMutations(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        try:d.graph.Decoder()
        except ImportError:raise unittest.SkipTest('research Capstone required')

    def fixture(self, prefix='90', handler=None, extras=(), objects=None):
        b=bytearray(0x90200)
        struct.pack_into('<14H',b,0,0x5a4d,0,len(b)//512,0,0x20,0,0xffff,0,0x100,0,0,0x10,0x40,0)
        raw=bytes.fromhex(prefix+'ff1e0000c3');b[0x300:0x300+len(raw)]=raw
        roots={0x300};caller=0x300+len(bytes.fromhex(prefix))
        for origin,code in extras:
            raw=bytes.fromhex(code);b[origin:origin+len(raw)]=raw;roots.add(origin)
        if handler is not None:
            raw=bytes.fromhex(handler);b[0x400:0x400+len(raw)]=raw;roots.add(0x400)
        r=d.Resolver(b,0x300,0,roots,0x700);r.prepare()
        r.facts={caller:dict(status='UNKNOWN',module_targets=[dict(module='CODE2',target_file=0x600)])}
        c=p.mz(b);objs=objects or [dict(name='glyph-table-a',start=0x500,end_exclusive=0x510),dict(name='glyph-table-b',start=0x520,end_exclusive=0x530)]
        scan=w.scan(r,objs,c,module='authored')
        summaries=s.Summaries(r,scan['writers'],'authored')
        handlers=None if handler is None else {0x300:dict(complete=True,targets=[dict(entry=0x400,summaries=summaries)])}
        return e.Admission(r,c,summaries,handlers=handlers),caller,objs

    def proof(self, **kwargs):
        a,caller,objs=self.fixture(**kwargs)
        return a.caller(caller,objs,10)

    def test_unknown_ss_alternate_caller(self):
        a,caller,objs=self.fixture(extras=[(0x340,'90ff1e0000c3')])
        a.roots[0x340]={'sp':[('literal',0x100)]}
        self.assertEqual(a.caller(0x341,objs,10)['status'],'UNKNOWN')
        self.assertEqual(a.caller(caller,objs,10)['status'],'BOUNDED')

    def test_unknown_sp_alternate_caller(self):
        a,caller,objs=self.fixture(extras=[(0x340,'90ff1e0000c3')])
        a.roots[0x340]={'ss':[('image',0x200)]}
        row=a.caller(0x341,objs,10)
        self.assertIsNone(row['sp']);self.assertEqual(row['status'],'UNKNOWN')
        self.assertEqual(a.caller(caller,objs,10)['status'],'BOUNDED')

    def test_ss_reassignment_reaches_caller(self):
        self.assertEqual(self.proof(prefix='8ed0')['status'],'UNKNOWN')
        row=self.proof(prefix='b800a08ed0')
        self.assertEqual(row['ss'],[('literal',0xa000)])
        self.assertEqual(row['status'],'BOUNDED')

    def test_sp_reassignment_reaches_caller(self):
        self.assertEqual(self.proof(prefix='89c4')['status'],'UNKNOWN')
        row=self.proof(prefix='bc0002')
        self.assertEqual(row['sp'],[0x200]);self.assertEqual(row['code2_entry_sp'],[0x1fc])

    def test_handler_changes_ss(self):
        self.assertEqual(self.proof(prefix='cd21',handler='8ed0cf')['status'],'UNKNOWN')

    def test_handler_changes_sp(self):
        self.assertEqual(self.proof(prefix='cd21',handler='89c4cf')['status'],'UNKNOWN')

    def test_balanced_handler_iret(self):
        row=self.proof(prefix='cd21',handler='1e50581fcf')
        self.assertEqual(row['status'],'BOUNDED')
        h=row['admitted_interrupts'][0]
        self.assertEqual(h['maximum_downward_bytes'],4)
        self.assertEqual(h['effects']['ds'],'preserving')
        self.assertEqual(row['sp'],[0x100])
        self.assertTrue(h['physical_stack_complete'])

    def test_unbalanced_handler_iret(self):
        self.assertEqual(self.proof(prefix='cd21',handler='50cf')['status'],'UNKNOWN')

    def test_handler_write_first_table(self):
        self.assertEqual(self.proof(prefix='cd21',handler='2ec606000201cf')['status'],'UNKNOWN')

    def test_handler_write_second_table(self):
        self.assertEqual(self.proof(prefix='cd21',handler='2ec606200201cf')['status'],'UNKNOWN')

    def test_handler_write_disjoint_memory(self):
        row=self.proof(prefix='cd21',handler='2ec606000101cf')
        self.assertEqual(row['status'],'BOUNDED')
        self.assertEqual(row['admitted_interrupts'][0]['effects']['memory'],'bounded finite effect')

    def test_alternate_handler_binding(self):
        a,caller,objs=self.fixture(prefix='cd21',handler='cf',extras=[(0x440,'8ed0cf')])
        a.handlers[0x300]['targets'].append(dict(entry=0x440,summaries=a.summaries))
        self.assertEqual(a.caller(caller,objs,10)['status'],'UNKNOWN')

    def test_unknown_vector_target(self):
        self.assertEqual(self.proof(prefix='cd21')['status'],'UNKNOWN')
        a,caller,objs=self.fixture(prefix='cd21',handler='cf')
        a.handlers[0x300]['complete']=False
        self.assertEqual(a.caller(caller,objs,10)['status'],'UNKNOWN')

    def test_disjoint_interrupt_frame_is_insufficient(self):
        row=self.proof(prefix='cd21')
        h=row['admitted_interrupts'][0]
        self.assertTrue(all(v['status']=='EXCLUDED' for v in h['immediate_frame']['physical_proofs']))
        self.assertEqual(h['status'],'UNKNOWN');self.assertEqual(row['status'],'UNKNOWN')

    def test_bounded_handler_unrelated_stack_unknown(self):
        a,caller,objs=self.fixture(prefix='cd21',handler='cf',extras=[(0x340,'8ed050c3')])
        unknown=s.stack_record(a.r,0x342,a.contract,objs,(None,None),'authored')
        self.assertEqual(unknown['status'],'UNKNOWN')
        self.assertEqual(a.caller(caller,objs,10)['status'],'BOUNDED')

    def test_unrelated_stack_mutation_does_not_destroy_entry_proof(self):
        baseline=self.proof()
        changed=self.proof(extras=[(0x340,'89c450c3')])
        self.assertEqual(baseline['status'],changed['status'])
        self.assertEqual(baseline['backward_provenance_slice'],changed['backward_provenance_slice'])
        self.assertEqual(changed['status'],'BOUNDED')

    def test_new_stack_path_reaching_caller_destroys_proof(self):
        # external root changes SP then branches into the original far caller.
        changed=self.proof(extras=[(0x340,'89c4e9bcff')])
        self.assertEqual(changed['status'],'UNKNOWN')
        self.assertIn(0x340,changed['backward_provenance_slice'])

    def test_all_legal_load_bases(self):
        a,caller,objs=self.fixture()
        for base in (a.contract['load_base_domain']['minimum'],0x4321,a.contract['load_base_domain']['maximum']):
            c=copy.deepcopy(a.contract);c['load_base_domain'].update(minimum=base,maximum=base)
            self.assertEqual(e.Admission(a.r,c,a.summaries).caller(caller,objs,10)['status'],'BOUNDED')

    def test_stack_table_overlap(self):
        objs=[dict(name='glyph-table-a',start=0x2f8,end_exclusive=0x300)]
        self.assertEqual(self.proof(objects=objs)['status'],'UNKNOWN')
        self.assertEqual(self.proof()['status'],'BOUNDED')

    def test_int66_on_required_slice_stays_unknown(self):
        row=self.proof(prefix='cd66')
        self.assertEqual(row['status'],'UNKNOWN')
        self.assertTrue(any('INT66' in reason for reason in row['admitted_interrupts'][0]['unresolved_reason']))

    def test_balanced_loop_and_unbalanced_cycle(self):
        # Both branches admitted, with zero versus negative loop stack delta.
        self.assertEqual(self.proof(prefix='505875fc')['status'],'BOUNDED')
        self.assertEqual(self.proof(prefix='5075fd')['status'],'UNKNOWN')

    def test_stack_only_contract_cannot_authorize_handler_memory(self):
        a,caller,objs=self.fixture(prefix='cd21',handler='cf')
        a.handlers[0x300]['targets'][0]['summaries']=s.Summaries(a.r,[],'authored',stack_only=True)
        self.assertEqual(a.caller(caller,objs,10)['status'],'UNKNOWN')

    def test_handler_placement_mutation_fails_closed(self):
        self.assertEqual(self.proof(prefix='cd21',handler='e692cf')['status'],'UNKNOWN')

    def test_nested_handler_placement_mutation_fails_closed(self):
        self.assertEqual(self.proof(prefix='cd21',handler='e83d00cf',extras=[(0x440,'e692c3')])['status'],'UNKNOWN')

    def test_handler_stack_switch_requires_physical_frame_proof(self):
        self.assertEqual(self.proof(prefix='cd21',handler='168c d3 b800a0 8ed0 50 58 8ed3 17 cf'.replace(' ',''))['status'],'UNKNOWN')

    def test_multiple_bounded_incoming_contexts_are_retained(self):
        a,caller,objs=self.fixture(extras=[(0x340,'e9beff')])
        a.roots[0x340]={'ss':[('image',0x200)],'sp':[('literal',0x180)]}
        row=a.caller(caller,objs,10)
        self.assertEqual(row['status'],'BOUNDED')
        self.assertEqual(row['sp'],[0x100,0x180])
        self.assertEqual(row['code2_entry_sp'],[0xfc,0x17c])

    def test_stack_return_proof_does_not_depend_on_unrelated_callee_store(self):
        a,caller,objs=self.fixture(prefix='e83d00',extras=[(0x340,'c606000001c3')])
        # Direct near callee has unknown DS memory but a matched stack return.
        full=a.caller(caller,objs,10)
        self.assertEqual(full['status'],'UNKNOWN')
        stack=e.Admission(a.r,a.contract,s.Summaries(a.r,[],'authored',stack_only=True))
        self.assertEqual(stack.caller(caller,objs,10)['status'],'BOUNDED')
        self.assertEqual(s.transfer(a.r,0x300,stack.summaries)['status'],'UNKNOWN')

    def test_complete_frame_and_bound_are_shared_not_site_allowlist(self):
        row=self.proof(prefix='50')
        self.assertEqual(row['sp'],[254]);self.assertEqual(row['code2_entry_sp'],[250])
        self.assertEqual(row['far_call_frame']['bytes'],4)
        self.assertEqual(row['physical_stack_ranges'][0]['offset_ranges'],[[v,v+1] for v in range(240,254)])
        self.assertFalse(row['global_table1_stack_completeness_required'])


if __name__=='__main__':unittest.main()
