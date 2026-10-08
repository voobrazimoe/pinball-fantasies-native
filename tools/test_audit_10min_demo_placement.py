#!/usr/bin/env python3
"""Public authored placement mutations: no private fixtures or Capstone needed."""
import copy
import struct
import unittest
import audit_10min_demo_placement as p


class PlacementMutations(unittest.TestCase):
    def setUp(self):
        # An authored large zero-filled MZ; image size forces a variable low B.
        self.binary=bytearray(0x90200)
        struct.pack_into('<14H',self.binary,0,0x5a4d,0,len(self.binary)//512,
                         0,0x20,0,0xffff,0,0x100,0,0,0x10,0x40,0)
        self.c=p.mz(self.binary)
        self.obj=dict(name='target',start=0x20200,end_exclusive=0x20204)

    def compare(self,segment,offsets=None,width=1,obj=None,contract=None):
        return p.exclusion(contract or self.c,dict(values=[segment],unknown=[]),
                           offsets,width,obj or self.obj)

    def test_variable_legal_load_bases(self):
        d=self.c['load_base_domain']
        self.assertEqual((d['minimum'],d['maximum']),(0x10,0x7000))
        for base in (d['minimum'],0x4321,d['maximum']):
            self.assertLessEqual(base*16+self.c['exec_reserved_image_paragraphs']*16,p.SPACE)
        proof=self.compare(('literal',0xa000))
        self.assertEqual(proof['status'],'EXCLUDED')
        # Inspect concrete endpoints of the universal certificate.
        for base in (0x10,0x4321,0x7000):
            self.assertLess(base*16+0x20004,0xa0000)

    def test_target_near_absolute_segment_boundary(self):
        c=copy.deepcopy(self.c);c['load_base_domain'].update(minimum=0x7000,maximum=0x7000)
        near=dict(name='edge',start=0x30200,end_exclusive=0x30204)
        self.assertEqual(self.compare(('literal',0xa000),{0},obj=near,contract=c)['status'],'UNKNOWN')
        below=dict(name='below',start=0x301fc,end_exclusive=0x30200)
        self.assertEqual(self.compare(('literal',0xa000),{0},obj=below,contract=c)['status'],'EXCLUDED')

    def test_absolute_store_disjoint_has_domains_and_reason(self):
        proof=self.compare(('literal',0xa000),{0xfffe},2)
        self.assertEqual(proof['status'],'EXCLUDED')
        self.assertEqual(proof['destination_domains'][0]['physical_a20_disabled'],[[0xafffe,0xb0000]])
        self.assertIn('every admitted B',proof['reason'])

    def test_smaller_initial_allocation_allows_same_absolute_store(self):
        # Alter the header-declared image requirement, not a segment name.
        b=self.binary[:0x40200]
        struct.pack_into('<H',b,4,len(b)//512)
        changed=p.mz(b)
        self.assertEqual(self.compare(('literal',0xa000))['status'],'EXCLUDED')
        self.assertEqual(self.compare(('literal',0xa000),contract=changed)['status'],'UNKNOWN')

    def test_byte_word_store_across_paragraph_boundary(self):
        c=copy.deepcopy(self.c);c['load_base_domain'].update(minimum=0x10,maximum=0x10)
        obj=dict(name='edge',start=0x210,end_exclusive=0x211)
        self.assertEqual(self.compare(('literal',0x10),{15},1,obj,c)['status'],'EXCLUDED')
        self.assertEqual(self.compare(('literal',0x10),{15},2,obj,c)['status'],'UNKNOWN')

    def test_16_bit_offset_wrap(self):
        self.assertEqual(p.offset_ranges({0xffff},2),[[0xffff,0x10000],[0,1]])
        obj=dict(name='zero',start=0x200,end_exclusive=0x201)
        self.assertEqual(self.compare(('image',0x200),{0xffff},2,obj)['status'],'UNKNOWN')

    def test_physical_20_bit_wrap_both_a20_states(self):
        c=copy.deepcopy(self.c);c['load_base_domain'].update(minimum=0x10,maximum=0x10)
        obj=dict(name='low',start=0x200,end_exclusive=0x201)
        # ffff:0110 reaches 0x100100, wrapping to the target at 0x100.
        proof=self.compare(('literal',0xffff),{0x110},1,obj,c)
        self.assertEqual(proof['status'],'UNKNOWN')
        self.assertEqual(proof['destination_domains'][0]['physical_a20_disabled'],[[0x100,0x101]])
        self.assertEqual(proof['destination_domains'][0]['linear_a20_enabled'],[[0x100100,0x100101]])

    def test_stack_interval_disjoint(self):
        proof=p.stack_exclusion(self.c,('image',0x200),(0,0x100),self.obj)
        self.assertEqual(proof['status'],'EXCLUDED')

    def test_stack_interval_overlapping_target(self):
        obj=dict(name='stack',start=0x280,end_exclusive=0x284)
        self.assertEqual(p.stack_exclusion(self.c,('image',0x200),(0,0x100),obj)['status'],'UNKNOWN')
        self.assertEqual(p.stack_exclusion(self.c,('image',0x200),None,obj)['status'],'UNKNOWN')

    def test_ss_reassignment_invalidates_stack_certificate(self):
        self.assertEqual(p.stack_exclusion(self.c,('image',0x200),(0,0x100),self.obj,
                                         reassignments=[0x444])['status'],'UNKNOWN')

    def test_unknown_transfer_changes_segment_stack_effects(self):
        self.assertEqual(p.stack_exclusion(self.c,('image',0x200),(0,0x100),self.obj,
                                         unknown_transfers=[0x555])['status'],'UNKNOWN')
        for reg in ('ss','ds','es'):
            proof=p.exclusion(self.c,dict(values=[('literal',0xa000)],
                unknown=['unknown transfer changes '+reg]),{0},1,self.obj)
            self.assertEqual(proof['status'],'UNKNOWN')

    def test_loader_relocation_is_not_runtime_mutation(self):
        b=copy.copy(self.binary)
        struct.pack_into('<H',b,6,1);struct.pack_into('<HH',b,0x40,0x10,0)
        c=p.mz(b)
        self.assertEqual(c['relocation_words'],[0x210])
        self.assertIn('not runtime mutation',c['relocation_effect'])
        self.assertEqual(self.compare(('image',0x200),{0},1,contract=c)['status'],'EXCLUDED')

    def test_no_video_value_heuristic(self):
        c=copy.deepcopy(self.c);c['load_base_domain'].update(minimum=0x8000,maximum=0x8000)
        # Deliberately weakened allocation certificate makes target intersect.
        self.assertEqual(self.compare(('literal',0xa000),{0},contract=c)['status'],'UNKNOWN')
        self.assertEqual(self.compare(('literal',0x9000),{0})['status'],'UNKNOWN')

    def test_relocated_segment_wrap_cannot_use_file_relative_exclusion(self):
        c=copy.deepcopy(self.c);c['load_base_domain'].update(minimum=0x10,maximum=0xff00)
        proof=self.compare(('image',0x20200),{0},contract=c)
        self.assertEqual(proof['status'],'UNKNOWN')

    def test_unmodeled_loader_and_target_extents_fail_closed(self):
        b=copy.copy(self.binary);struct.pack_into('<H',b,12,0)
        with self.assertRaisesRegex(ValueError,'load-high'):p.mz(b)
        obj=dict(name='outside',start=len(self.binary),end_exclusive=len(self.binary)+4)
        self.assertEqual(self.compare(('literal',0xa000),obj=obj)['status'],'UNKNOWN')


if __name__=='__main__':unittest.main()
