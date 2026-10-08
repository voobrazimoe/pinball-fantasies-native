#!/usr/bin/env python3
"""Authored admission certificate mutations and real fail-closed integration."""
import copy
import os
import unittest
import audit_10min_demo_admission as a
import test_audit_10min_demo_entry as entry_tests


class CallbackVectorMutations(unittest.TestCase):
    def callback(self):
        return dict(root=0x4517,registration_targets=[0x4517],invocations_complete=True,
            invocations=[dict(kind='far',ss=('image',0x200),sp=0x100,
                return_instruction='retf',linkage_complete=True)])

    def binding(self):
        return dict(writers_complete=True,ordering_complete=True,no_later_overwrite=True,
            effect_domain_complete=True,candidate_targets=[['authored',0x400]],
            handler_summaries=[dict(status='BOUNDED',return_instruction='iret',stack_delta=0,
                ss='preserving',sp='preserving',ds='preserving',es='preserving',allocation='preserving',
                object_memory='EXCLUDED',maximum_downward_bytes=4)])

    def fixture(self, **kw):
        entry_tests.EntryHandlerMutations.setUpClass()
        return entry_tests.EntryHandlerMutations().fixture(**kw)

    def test_alternate_callback_registration_target(self):
        c=self.callback();c['registration_targets'].append(0x592b)
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')

    def test_second_invoker_different_ss_retained(self):
        c=self.callback();c['invocations'].append(dict(c['invocations'][0],ss=('image',0x400)))
        r=a.callback_contract(c);self.assertEqual(r['status'],'BOUNDED')
        self.assertEqual([v['entry_ss'] for v in r['contexts']],[('image',0x200),('image',0x400)])

    def test_second_invoker_different_sp_retained(self):
        c=self.callback();c['invocations'].append(dict(c['invocations'][0],sp=0x180))
        r=a.callback_contract(c);self.assertEqual(r['status'],'BOUNDED')
        self.assertEqual([v['entry_sp'] for v in r['contexts']],[0xfc,0x17c])

    def test_near_vs_far_callback_frame(self):
        c=self.callback();c['invocations'][0]['kind']='near'
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')
        self.assertEqual(a.invocation_context('near',1,256,'ret')['frame_bytes'],2)
        self.assertEqual(a.invocation_context('far',1,256,'retf')['frame_bytes'],4)

    def test_interrupt_driver_frame_is_separate(self):
        c=self.callback();c['invocations'][0]['prefix_bytes']=26
        r=a.callback_contract(c);self.assertEqual(r['contexts'][0]['entry_sp'],226)
        self.assertEqual(r['contexts'][0]['restored_source_sp'],230)
        c['invocations'][0].update(kind='interrupt',return_instruction='retf')
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')
        self.assertEqual(a.invocation_context('interrupt',1,256,'iret')['frame_bytes'],6)

    def test_missing_callback_return(self):
        c=self.callback();c['invocations'][0]['return_instruction']=None
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')

    def test_unbalanced_callback_return(self):
        c=self.callback();c['invocations'][0]['return_adjust']=2
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')

    def test_alternate_int66_vector_writer(self):
        b=self.binding();b['candidate_targets'].append(['new',0x600]);b['effect_domain_complete']=False
        self.assertEqual(a.binding_contract(b)['status'],'UNKNOWN')

    def test_int66_overwrite_after_installer(self):
        b=self.binding();b['no_later_overwrite']=False
        r=a.binding_contract(b);self.assertEqual(r['status'],'UNKNOWN');self.assertIsNone(r['installed_target_set'])

    def test_incomplete_installer_before_consumer(self):
        b=self.binding();b['ordering_complete']=False
        self.assertEqual(a.binding_contract(b)['status'],'UNKNOWN')

    def test_admitted_int66_handler_changes_ss(self):
        model,caller,objs=self.fixture(prefix='cd66',handler='8ed0cf')
        self.assertEqual(model.caller(caller,objs,10)['status'],'UNKNOWN')

    def test_admitted_int66_handler_changes_sp(self):
        model,caller,objs=self.fixture(prefix='cd66',handler='89c4cf')
        self.assertEqual(model.caller(caller,objs,10)['status'],'UNKNOWN')

    def test_balanced_int66_iret(self):
        model,caller,objs=self.fixture(prefix='cd66',handler='1e50581fcf')
        self.assertEqual(model.caller(caller,objs,10)['status'],'BOUNDED')

    def test_handler_writes_first_lookup(self):
        model,caller,objs=self.fixture(prefix='cd66',handler='b830008ed8c606000201cf')
        self.assertEqual(model.caller(caller,objs,10)['status'],'UNKNOWN')

    def test_handler_writes_second_lookup(self):
        model,caller,objs=self.fixture(prefix='cd66',handler='b830008ed8c606200201cf')
        self.assertEqual(model.caller(caller,objs,10)['status'],'UNKNOWN')

    def test_handler_disjoint_store(self):
        model,caller,objs=self.fixture(prefix='cd66',handler='1eb830008ed8c6060000011fcf')
        self.assertEqual(model.caller(caller,objs,10)['status'],'BOUNDED')

    def test_alternate_int33_target_missing_contract(self):
        b=self.binding();b['candidate_targets'].append(['mouse2',0x600]);b['effect_domain_complete']=False
        self.assertEqual(a.binding_contract(b)['status'],'UNKNOWN')

    def test_variable_int33_identity_complete_invariant_contract(self):
        b=self.binding();b.update(identity_unknown=True,candidate_targets=None)
        self.assertEqual(a.binding_contract(b)['status'],'BOUNDED')
        # This is certificate consumption, not evidence that real INT33 has it.

    def test_unknown_int33_contract(self):
        b=self.binding();b['handler_summaries']=[]
        self.assertEqual(a.binding_contract(b)['status'],'UNKNOWN')

    def test_matched_return_adds_interrupt_frontier(self):
        model,caller,objs=self.fixture(prefix='e83d00',extras=[(0x340,'cd33c3')])
        row=model.caller(caller,objs,10)
        self.assertEqual(row['status'],'UNKNOWN')
        self.assertIn(0x340,[v['site'] for v in row['admitted_interrupts']])

    def test_unrelated_interrupt_does_not_destroy_closure(self):
        model,caller,objs=self.fixture(extras=[(0x340,'cd33c3')])
        self.assertEqual(model.caller(caller,objs,10)['status'],'BOUNDED')

    def test_new_reachable_interrupt_destroys_closure(self):
        model,caller,objs=self.fixture(prefix='cd33')
        self.assertEqual(model.caller(caller,objs,10)['status'],'UNKNOWN')

    def test_registration_without_invocation_proof(self):
        c=self.callback();c['invocations']=[]
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')

    def test_invocation_candidates_do_not_prove_completeness(self):
        c=self.callback();c['invocations_complete']=False
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')

    def test_unknown_second_source_ss(self):
        c=self.callback();c['invocations'].append(dict(c['invocations'][0],ss=None))
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')

    def test_unknown_second_source_sp(self):
        c=self.callback();c['invocations'].append(dict(c['invocations'][0],sp=None))
        self.assertEqual(a.callback_contract(c)['status'],'UNKNOWN')

    def test_writers_candidate_set_not_complete(self):
        b=self.binding();b['writers_complete']=False
        self.assertEqual(a.binding_contract(b)['status'],'UNKNOWN')


if __name__=='__main__':unittest.main()
