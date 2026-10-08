#!/usr/bin/env python3
"""Private bounded-domain research tests; fixture absent -> clean SKIP."""
import os
from pathlib import Path
import struct
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

import audit_10min_demo_domains as domains
import audit_10min_demo_graph as graph
import audit_10min_demo_control as control
import audit_10min_demo_writers as writers


class PrivateDomains(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        names=('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')
        values=[os.getenv(n) for n in names]
        if not all(values):
            raise unittest.SkipTest('private demo, canonical and historical inputs required')
        cls.paths=list(map(Path,values))
        cls.result=domains.audit(*cls.paths)
        cls.binary=(cls.paths[0]/'TABLE1.PRG').read_bytes()
        cls.by_site={d['source']:d for d in cls.result['domains']}

    def test_fixed_point_retains_every_derived_target(self):
        r=self.result
        self.assertTrue(r['fixed_point'])
        self.assertEqual(r['rounds'][-1]['added_roots'],0)
        roots=set(r['root_files'])
        for domain in r['domains']:
            self.assertLessEqual(set(domain['targets']),roots)
            if domain['status']=='UNKNOWN':
                self.assertTrue(domain['unknown'])
        self.assertEqual(r['unknown_domain_count'],len(r['unknown_sites']))
        self.assertEqual(len(r['unknown_sites']),sum(d['status']=='UNKNOWN' for d in r['domains']))
        for a,b in zip(r['rounds'],r['rounds'][1:]):
            self.assertEqual(a['roots']+a['added_roots'],b['roots'])

    def test_byte_glyph_index_target_sets_are_complete(self):
        for site in (0x7263,0x729a):
            d=self.by_site[site]
            table=d['tables'][0]
            self.assertEqual(table['entries'],256)
            values=struct.unpack_from('<256H',self.binary,table['start'])
            expected={0x300+((v+table['code_addend'])&65535) for v in values}
            self.assertEqual(set(d['targets']),expected)

    def test_task_selector_covers_all_bounded_lookup_outputs(self):
        selector=next(t for t in self.by_site[0x5eab]['tables'] if 'selector_values' in t)
        values=self.binary[selector['selector_base']:selector['selector_base']+selector['selector_count']]
        self.assertEqual(set(values),set(selector['selector_values']))
        targets={0x300+struct.unpack_from('<H',self.binary,p)[0] for p in selector['pointer_sources']}
        self.assertLessEqual(targets,set(self.by_site[0x5eab]['targets']))

    def test_program_predecessors_and_numeric_difference_survive(self):
        programs={p['start']:p for p in self.result['programs']}
        p=programs[0x1b88e]
        self.assertTrue(p['predecessors'])
        self.assertEqual(p['nodes'][1]['node'],'_PRINT5')
        self.assertEqual(p['nodes'][1]['operands'][1],336)
        for p in programs.values():
            self.assertTrue(p['predecessors'])
            self.assertTrue(p['nodes'] or p['unknown'])

    def test_negative_persistence_candidates_do_not_claim_variant_a(self):
        for p in self.result['persistence_candidates']:
            self.assertFalse(p['body_in_direct_or_candidate_cfg'])
            self.assertTrue(p['verdict'].startswith('UNKNOWN'))
        with self.assertRaisesRegex(ValueError,'unresolved'):
            domains.require_closed(self.result)
        # Zero transfer-site UNKNOWNs alone cannot erase the other gates.
        fake=dict(self.result,unknown_domain_count=0)
        with self.assertRaisesRegex(ValueError,'unresolved'):
            domains.require_closed(fake)

    def test_mutation_rejected_before_resolver(self):
        original=Path.read_bytes
        selected=self.paths[0]/'TABLE1.PRG'
        def changed(path):
            b=original(path)
            if path==selected:
                b=b[:0x1d5ec]+bytes([b[0x1d5ec]^1])+b[0x1d5ed:]
            return b
        with patch.object(Path,'read_bytes',changed),patch.object(domains.Resolver,'run') as run:
            with self.assertRaisesRegex(ValueError,'not the pinned demo'):
                domains.audit(*self.paths)
            run.assert_not_called()

    def test_register_save_summary_does_not_guess_clobber_preservation(self):
        # Authored miniature examples, not extracted executable data.
        saved=b'\x53\xbb\x34\x12\x5b\xc3'
        clobbered=b'\x66\xbb\x34\x12\x00\x00\xc3'
        good=domains.Resolver(saved,0,0,[0],len(saved))
        bad=domains.Resolver(clobbered,0,0,[0],len(clobbered))
        self.assertTrue(good.preserves(0,'bx'))
        self.assertFalse(bad.preserves(0,'bx'))

    def resolver(self, binary=None):
        r=domains.Resolver(binary or self.binary,0x300,0x19db0,
                           self.result['root_files'],self.result['main_cs_end'])
        r.prepare()
        return r

    def test_former_unknown_sites_are_not_promoted_without_completeness(self):
        expected={0x66a,0xae9,0x3a01,0x3d13,0x47a7,0x5640,0x5673,0x574d,
                  0x5d0c,0x5eab,0x6095,0x6151}
        self.assertEqual(set(self.result['unknown_sites']),expected)
        self.assertEqual(len(self.result['domains']),30)
        for at in expected:
            f=self.by_site[at]
            self.assertTrue(f['instruction'])
            self.assertTrue(f['operand_class'])
            self.assertTrue(f['domain_category'])
            self.assertTrue(f['unknown'])
            self.assertIsNone(f['completeness_reason'])
        self.assertEqual(self.by_site[0x5eab]['instruction'],'call')
        self.assertIn('not a writer',self.by_site[0x5eab]['domain_category'])

    def test_far_operands_derive_module_and_vector_classification(self):
        for at in (0x66a,0xae9,0x5673,0x574d):
            f=self.by_site[at]
            self.assertEqual(f['module_targets'],[dict(module='CODE2',segment_file_base=0xaed0,
                offset=0,target_file=0xaed0)])
            self.assertTrue(f['tables'][0]['segment_relocation'])
            self.assertEqual(f['segment_sources'],[['module',0x300]])
        for at in (0x3a01,0x3d13,0x5640):
            f=self.by_site[at]
            self.assertEqual(f['binding_slot'],dict(segment=0,offset=0x198))
            self.assertEqual(f['domain_category'],'external/API binding')
            self.assertFalse(f['module_targets'])
            self.assertEqual(len(f['binding_producers']),11)

    def test_api_vector_installers_are_derived_from_supplied_modules(self):
        entries={f['module']:f['target_offset'] for f in self.result['api_vector_installers']}
        self.assertEqual(entries,{'ADLIB.SDR':0x49,'GUS.SDR':0xf7,'INTERNAL.SDR':0x4e,
            'NOSOUND.SDR':0x40,'PAS16.SDR':0x49,'SB16.SDR':0x49,'SB20.SDR':0x49,
            'SBLASTER.SDR':0x49,'SBPRO.SDR':0x49,'SM2.SDR':0x49,'THING.SDR':0x49})
        import audit_10min_demo_sdr as sdr
        original=sdr.unpack
        def missing_store(b):
            raw,count=original(b)
            # Mutate the store's source register in the in-memory module;
            # compressed identity alone cannot substitute for producer checking.
            raw=bytearray(raw)
            raw[0x1b if raw[6]==0xfc else 0x2b] ^= 0x10
            return bytes(raw),count
        with patch.object(sdr,'unpack',missing_store):
            with self.assertRaisesRegex(ValueError,'vector'):
                control.api_installers(self.paths[0])

    def test_area_grammar_includes_all_four_reaching_bases(self):
        f=self.by_site[0x6151]
        self.assertIn(0x60f7,f['definitions'])
        self.assertEqual({t['start'] for t in f['tables']},
                         {0x19db0+v for v in (0xdb7,0xe45,0xebf,0xedf)})
        self.assertEqual(set(f['targets']),{0x19ba,0x1bb3,0x1d12,0x1d70,0x1e2d,0x1fa0,
            0x1fde,0x2296,0x2561,0x2568,0x259d,0x27c0,0x27db,0x27f6,0x287b,0x28c6,
            0x28d9,0x28e0,0x29ce,0x29d2,0x2a2f,0x2a8c,0x2ae9,0x2ceb,0x2cec,0x2ced})
        self.assertEqual(set(self.by_site[0x6095]['targets']),{0x1650,0x16b2,0x1732,0x17b2})
        r=self.resolver()
        self.assertEqual(control.area_operand(r,0x6151)['offsets'],[v-0x300 for v in f['targets']])
        # Same address, changed producer: derived domain changes.
        b=bytearray(self.binary);struct.pack_into('<H',b,0x60f8,0xedf)
        changed=control.area_operand(self.resolver(b),0x6151)
        self.assertLess(len(changed['offsets']),len(f['targets']))
        # Same address, damaged cursor recurrence: fail closed.
        b=bytearray(self.binary);b[0x6127]=8
        broken=control.area_operand(self.resolver(b),0x6151)
        self.assertFalse(broken['offsets'])
        self.assertTrue(any('grammar not proved' in e for e in broken['unknown']))

    def test_related_code2_recursive_candidate_fixed_point(self):
        c=self.result['related_code2']
        self.assertEqual([d['source'] for d in c['domains']],[0xafa6,0xaff7])
        self.assertEqual(c['rounds'][-1]['added_roots'],0)
        self.assertEqual(len(c['root_files']),511)
        self.assertEqual(c['instruction_count'],1723)
        for f in c['domains']:
            self.assertEqual(f['status'],'UNKNOWN')
            self.assertEqual(len(f['targets']),255)
            self.assertLessEqual(set(f['targets']),set(c['root_files']))
            t=f['tables'][0]
            values=struct.unpack_from('<256H',self.binary,t['start'])
            expected={0xaed0+((v+t['code_addend'])&65535) for v in values}
            self.assertEqual(set(f['targets']),expected)

    def test_far_relocation_or_code_role_mutation_does_not_execute_data(self):
        b=bytearray(self.binary)
        source=self.by_site[0x66a]['tables'][0]['pointer_source']
        struct.pack_into('<H',b,source,2)
        f=control.far_operand(self.resolver(b),0x66a)
        self.assertFalse(f['module_targets'])
        self.assertTrue(any('code-entry role' in e for e in f['unknown']))
        b=bytearray(self.binary)
        h=struct.unpack_from('<14H',b)
        for p in range(h[12],h[12]+4*h[3],4):
            off,seg=struct.unpack_from('<HH',b,p)
            if h[4]*16+seg*16+off==source+2:
                struct.pack_into('<HH',b,p,0,0);break
        f=control.far_operand(self.resolver(b),0x66a)
        self.assertFalse(f['module_targets'])
        self.assertTrue(any('no MZ relocation' in e for e in f['unknown']))

    def test_glyph_idiom_is_address_independent_and_table_derived(self):
        # Authored examples only. The two masks/table values are under test.
        code=b'\x83\xe0\x01\xd1\xe0\x83\xc0\x00\x89\xc7\x26\x8b\x15\xb8\xf2\x60\x83\xc2\x30\xff\xd2\xc3'
        for origin in (0,16):
            b=bytearray(b'\x90'*origin+code)
            b.extend(b'\x90'*(64-len(b)));b[48]=0xc3;b[50]=0xc3
            b.extend(struct.pack('<2H',0,2))
            r=domains.Resolver(b,0,64,[origin],64);r.prepare()
            at=next(p for p,x in r.seen.items() if x.mnemonic=='call')
            f=control.glyph_operand(r,at)
            self.assertEqual(f['targets'],[48,50])
            struct.pack_into('<H',b,66,4)
            r=domains.Resolver(b,0,64,[origin],64);r.prepare()
            self.assertEqual(control.glyph_operand(r,at)['targets'],[48,52])
            b[origin+2]=0xff
            r=domains.Resolver(b,0,64,[origin],64);r.prepare()
            self.assertFalse(control.glyph_operand(r,at)['targets'])

    def test_far_construction_is_address_independent_and_rejects_unknown_segment(self):
        # Authored miniature MZ image, not a private executable slice.
        for origin in (0x310,0x350):
            b=bytearray(0x1000)
            struct.pack_into('<14H',b,0,0x5a4d,0,8,1,0x20,0,0xffff,0,0x100,0,
                             origin-0x300,0x10,0x40,0)
            struct.pack_into('<HH',b,0x40,2,0x40)
            struct.pack_into('<HH',b,0x600,0,0x60)
            b[origin:origin+8]=b'\x0e\x07\x26\xff\x1e\x00\x03\xc3'
            r=domains.Resolver(b,0x300,0x1000,[origin],0x700);r.prepare()
            f=control.far_operand(r,origin+2)
            self.assertEqual(f['module_targets'],[dict(module='CODE2',segment_file_base=0x800,
                                                       offset=0,target_file=0x800)])
            self.assertEqual(f['status'],'UNKNOWN')
            b[origin]=0x1e  # PUSH DS instead of the proved PUSH CS
            r=domains.Resolver(b,0x300,0x1000,[origin],0x700);r.prepare()
            f=control.far_operand(r,origin+2)
            self.assertFalse(f['module_targets'])
            self.assertTrue(f['unknown'])

    def test_ambiguous_table_base_rejects_domain(self):
        code=bytes.fromhex('83e001d1e001c89089c7268b15b8f26083c230ffd2c3')
        b=bytearray(code);b.extend(bytes(64-len(b)));b.extend(struct.pack('<2H',0,2))
        r=domains.Resolver(b,0,64,[0],64);r.prepare()
        at=next(p for p,x in r.seen.items() if x.mnemonic=='call')
        f=control.glyph_operand(r,at)
        self.assertEqual(f['status'],'UNKNOWN')
        self.assertFalse(f['targets'])
        self.assertTrue(any('grammar not proved' in v for v in f['unknown']))

    def test_overlapping_offset_relocation_is_not_single_module_pointer(self):
        b=bytearray(self.binary)
        source=self.by_site[0x66a]['tables'][0]['pointer_source']
        h=struct.unpack_from('<14H',b)
        # Repurpose a different existing relocation as an extra offset-word relocation.
        for p in range(h[12],h[12]+4*h[3],4):
            off,seg=struct.unpack_from('<HH',b,p)
            if h[4]*16+16*seg+off not in (source+2,0x566b):
                image=source-h[4]*16
                struct.pack_into('<HH',b,p,image&15,image>>4);break
        f=control.far_operand(self.resolver(b),0x66a)
        self.assertEqual(f['status'],'UNKNOWN')
        self.assertFalse(f['module_targets'])
        self.assertIn('far pointer has additional overlapping loader relocation',f['unknown'])

    def test_callback_initial_values_and_task_writer_inventory_stay_open(self):
        self.assertEqual(self.by_site[0x47a7]['field_evidence']['initial_value'],0)
        self.assertEqual(self.by_site[0x5d0c]['field_evidence']['initial_value'],0)
        w=self.by_site[0x5eab]['writer_evidence']
        self.assertIn(0x5e9c,{p['source'] for p in w['task_api_writer_candidates']})
        self.assertIn(0x3b4b,{p['source'] for p in w['reset_writer_candidates']})
        self.assertEqual(len(self.by_site[0x5eab]['targets']),57)
        with self.assertRaisesRegex(ValueError,'unresolved'):
            domains.require_closed(self.result)

    def test_segment_save_summary_tracks_segment_stack_separately_from_pusha(self):
        saved=b'\x1e\x60\xb8\x00\x00\x8e\xd8\x61\x1f\xc3'
        bad=b'\xb8\x00\x00\x8e\xd8\xc3'
        self.assertTrue(domains.Resolver(saved,0,0,[0],len(saved)).preserves(0,'ds'))
        self.assertFalse(domains.Resolver(bad,0,0,[0],len(bad)).preserves(0,'ds'))
        es_clobber=bytes.fromhex('8ec0c3')
        es_saved=bytes.fromhex('068ec007c3')
        pusha_only=bytes.fromhex('608ec061c3')
        self.assertFalse(domains.Resolver(es_clobber,0,0,[0],len(es_clobber)).preserves(0,'es'))
        self.assertTrue(domains.Resolver(es_saved,0,0,[0],len(es_saved)).preserves(0,'es'))
        self.assertFalse(domains.Resolver(pusha_only,0,0,[0],len(pusha_only)).preserves(0,'es'))

    def test_writer_inventory_retains_every_memory_store_and_open_alias(self):
        r=self.resolver()
        evidence=self.result['object_writer_evidence']
        self.assertEqual({o['start'] for o in evidence[0]['objects']},
                         {0x5669,0x56a8,0x1f8b0,0x1fab0})
        for scan in evidence:
            self.assertFalse(scan['runtime_immutability_proved'])
            self.assertEqual(scan['status'],'UNKNOWN')
            self.assertTrue(scan['implicit_stack_writers'])
            self.assertEqual(len(scan['writers']),scan['explicit_writer_count'])
        expected={(p,j) for p,x in r.seen.items() for j,o in enumerate(x.operands)
                  if o.type==r.x86.X86_OP_MEM and o.access&r.cs.CS_AC_WRITE}
        actual={(w['source'],w['operand']) for w in evidence[0]['writers']}
        self.assertEqual(actual,expected)
        indexed=next(w for w in evidence[0]['writers'] if w['source']==0x5c39)
        self.assertEqual(indexed['segment'],'cs')
        self.assertIsNone(indexed['offsets'])
        self.assertTrue(indexed['range_unknown'])
        self.assertEqual(set(indexed['possible_objects']),
                         {'far-pointer-0x5669','far-pointer-0x56a8'})

    def test_relocation_lifecycle_is_not_runtime_mutability(self):
        rows=self.result['pointer_lifecycle']
        self.assertEqual({v['object_file'] for v in rows},{0x5669,0x56a8})
        for v in rows:
            self.assertEqual(v['loader_relocation_words'],[v['segment_word']])
            self.assertNotEqual(v['offset_word'],v['segment_word'])
            self.assertFalse(v['runtime_offset_immutable'])
            self.assertFalse(v['runtime_segment_immutable'])
            self.assertIn('UNKNOWN',v['executable_initialization'])

    def test_loader_placement_excludes_absolute_writers_without_promoting_family(self):
        c=self.result['placement_contract']
        self.assertEqual(c['load_base_domain']['maximum'],0x7ce0)
        self.assertEqual(c['image_bytes'],0x83066)
        self.assertEqual(c['launcher']['normal_exec_sites'],[0x6c3])
        for scan in self.result['object_writer_evidence']:
            self.assertEqual(scan['status'],'UNKNOWN')
            self.assertTrue(scan['placement_effects']['boundaries'])
            self.assertTrue(scan['placement_effects']['implicit_sites'])
            for w in scan['writers']:
                if w['segment_evidence']['values']==[('literal',0xa000)] and not w['segment_evidence']['unknown']:
                    self.assertEqual(w['possible_objects'],[])
                    self.assertTrue(all(v['status']=='EXCLUDED' for v in w['placement_proofs']))
        code2=self.result['object_writer_evidence'][1]
        self.assertTrue(all(v==0 for v in code2['possible_operand_counts_after'].values()))

    def test_all_typed_code2_candidates_have_independent_segment_summaries(self):
        c=self.result['related_code2']
        for summary in c['callee_segment_summaries']:
            f=next(f for f in c['domains'] if f['source']==summary['source'])
            self.assertEqual({v['entry'] for v in summary['candidates']},set(f['targets']))
            self.assertEqual(summary['candidate_count'],255)
            self.assertTrue(all(v['es_preserved'] for v in summary['candidates']))
            self.assertTrue(all(v['ds_preserved'] for v in summary['candidates']))
        for v in c['target_validity']:
            self.assertEqual(v['raw_entry_count'],256)
            self.assertEqual(v['distinct_target_count'],255)
            self.assertTrue(v['all_at_instruction_boundaries'])
            self.assertEqual(v['instruction_overlap_sources'],[])
            self.assertEqual(v['status'],'VALID TYPED CANDIDATES')

    def test_reachable_code2_es_clobber_reopens_dependency(self):
        b=bytearray(self.binary)
        target=next(t for t in self.result['related_code2']['domains'][0]['targets']
                    if t>0xb09f)
        b[target:target+3]=bytes.fromhex('8ec0c3')  # MOV ES,AX; RET, authored mutation
        r=self.resolver(b)
        result=control.related_code2(r,list(self.by_site.values()))
        first,second=result['domains']
        self.assertEqual(first['status'],'UNKNOWN')
        self.assertTrue(any(not v['es_preserved'] for v in first['callee_segment_summaries']['candidates']))
        self.assertTrue(second['es_alias_evidence']['unknown'])
        self.assertEqual(second['status'],'UNKNOWN')

    def miniature_writer_scan(self, code, origin=0x300):
        # Authored MZ container: two objects in one image segment, no private bytes.
        b=bytearray(0x1000)
        struct.pack_into('<14H',b,0,0x5a4d,0,8,0,0x20,0,0xffff,0,0x100,0,
                         origin-0x300,0x10,0x40,0)
        b[origin:origin+len(code)]=code
        r=domains.Resolver(b,0x300,0x1000,[origin],0x700);r.prepare()
        objects=[dict(name='pointer',start=0x600,end_exclusive=0x604),
                 dict(name='table',start=0x620,end_exclusive=0x630)]
        return writers.scan(r,objects)

    def test_relocated_alias_identity_and_bounded_rep_range(self):
        b=bytearray(0x1000)
        struct.pack_into('<14H',b,0,0x5a4d,0,8,1,0x20,0,0xffff,0,0x100,0,0,0x10,0x40,0)
        struct.pack_into('<HH',b,0x40,0x101,0)
        code=bytes.fromhex('6840001fc606000001c3')
        b[0x300:0x300+len(code)]=code
        r=domains.Resolver(b,0x300,0x1000,[0x300],0x700);r.prepare()
        obj=[dict(name='object',start=0x600,end_exclusive=0x604)]
        scan=writers.scan(r,obj)
        self.assertEqual(scan['writers'][0]['segment_evidence']['values'],[('image',0x600)])
        self.assertEqual(scan['writers'][0]['possible_objects'],['object'])
        # Same numerical immediate without a relocation is an absolute segment.
        struct.pack_into('<H',b,6,0)
        r=domains.Resolver(b,0x300,0x1000,[0x300],0x700);r.prepare()
        scan=writers.scan(r,obj)
        self.assertEqual(scan['writers'][0]['segment_evidence']['values'],[('literal',0x40)])
        self.assertEqual(scan['status'],'UNKNOWN')
        rep=self.miniature_writer_scan(bytes.fromhex('0e07bf1003b90200f3abc3'))
        self.assertEqual(rep['writers'][0]['possible_objects'],[])
        self.assertIsNotNone(rep['writers'][0]['offsets'])
        self.assertEqual(rep['writers'][0]['range_unknown'],[])

    def test_injected_byte_word_indexed_ds_es_writers_fail_closed(self):
        baseline=self.miniature_writer_scan(bytes.fromhex('c3'))
        self.assertEqual(baseline['status'],'BOUNDED')
        self.assertTrue(baseline['runtime_immutability_proved'])
        # Source/consumer addresses do not select the object rules.
        cases=[('2ec606000301','pointer'),       # CS byte offset word
               ('2ec70602030100','pointer'),   # CS word segment word
               ('2ec706ff020100','pointer'),   # word starts one byte before object
               ('0e1fbb0103c60701','pointer'),  # DS=CS; indexed byte
               ('0e07bf020326c7050100','pointer'), # ES=CS; indexed word
               ('0e1fc606200301','table'),      # DS alias, table byte
               ('0e07bf2003b90100f3ab','table')] # ES alias, REP word
        for origin in (0x300,0x340):
            for code,name in cases:
                with self.subTest(origin=origin,code=code):
                    scan=self.miniature_writer_scan(bytes.fromhex(code+'c3'),origin)
                    self.assertEqual(scan['status'],'UNKNOWN')
                    self.assertFalse(scan['runtime_immutability_proved'])
                    self.assertTrue(any(name in v['possible_objects'] for v in scan['writers']))
        # Unknown ES and arbitrary index cannot be discarded.
        scan=self.miniature_writer_scan(bytes.fromhex('26c60701c3'))
        self.assertEqual(set(scan['writers'][0]['possible_objects']),{'pointer','table'})
        self.assertTrue(scan['writers'][0]['segment_evidence']['unknown'])

    def test_overlap_wrap_and_segment_disjointness(self):
        self.assertTrue(writers.overlap(0x300,{0x2ff},2,0x600,0x604))
        self.assertTrue(writers.overlap(0x300,{0xffff},2,0x300,0x301))
        self.assertFalse(writers.overlap(0x1000,None,2,0x600,0x604))
        self.assertTrue(writers.overlap(None,{0},1,0x620,0x630))
        scan=self.miniature_writer_scan(bytes.fromhex('2ec606000401c3'))
        self.assertEqual(scan['writers'][0]['possible_objects'],[])
        # Injecting a new admitted writer changes the overlap evidence.
        changed=self.miniature_writer_scan(bytes.fromhex('2ec6060004012ec606000301c3'))
        self.assertEqual(changed['explicit_writer_count'],2)
        self.assertEqual(changed['writers'][1]['possible_objects'],['pointer'])

    def test_real_code2_closure_requires_every_runtime_obligation(self):
        scan=self.result['object_writer_evidence'][1]
        effects=scan['placement_effects']
        self.assertEqual(len(effects['implicit_sites']),13)
        self.assertEqual(effects['closed_stack_site_count'],0)
        self.assertEqual(effects['unresolved_stack_site_count'],13)
        self.assertTrue(all(v['status']=='UNKNOWN' for v in effects['implicit_sites']))
        self.assertEqual(len(effects['caller_contexts']),4)
        self.assertTrue(all(v['ss'] is None and v['sp'] is None for v in effects['caller_contexts']))
        self.assertEqual(effects['interrupts'],[])
        for boundary in effects['boundaries']:
            self.assertEqual(boundary['classification'],'UNKNOWN')
            self.assertTrue(boundary['candidate_effects_complete'])
            self.assertEqual(len(boundary['callee_summaries']),255)
            self.assertEqual(boundary['maximum_callee_downward_bytes'],0)
            self.assertTrue(all(boundary['effects'][n]=='preserving' for n in ('ds','es','ss','sp','allocation','placement')))
        self.assertFalse(scan['runtime_immutability_proved'])
        self.assertTrue(all(v['status']=='UNKNOWN' for v in self.result['related_code2']['domains']))

    def test_real_code2_entry_projection_retains_unresolved_admission(self):
        e=self.result['object_writer_evidence'][1]['placement_effects']['entry_stack_provenance']
        self.assertEqual(e['status'],'UNKNOWN')
        self.assertEqual({v['site'] for v in e['entry_callers']},{0x66a,0xae9,0x5673,0x574d})
        for row in e['entry_callers']:
            self.assertEqual(row['status'],'UNKNOWN')
            self.assertFalse(row['global_table1_stack_completeness_required'])
            self.assertEqual(row['slice_roots'],[0x4517])
            self.assertEqual(row['local_code2_stack_bound'],10)
            self.assertEqual(row['far_call_frame']['bytes'],4)
            self.assertTrue(row['unresolved_reason'])
        self.assertEqual(e['initial_mz']['ss'],['image',0x200])
        self.assertEqual(e['initial_mz']['sp'],256)
        startup=e['startup_frontier']
        self.assertEqual(startup['site'],0x32ab)
        self.assertEqual(startup['status'],'UNKNOWN')
        self.assertTrue(all(v['status']=='EXCLUDED' for v in startup['immediate_frame']['physical_proofs']))
        caller=next(v for v in e['entry_callers'] if v['site']==0x66a)
        self.assertIn(0x6ec,[h['site'] for h in caller['admitted_interrupts']])

    def test_real_callback_int_admission_stays_fail_closed(self):
        a=self.result['callback_and_interrupt_admission']
        self.assertFalse(a['completeness_flag'])
        self.assertEqual(a['driver_name_admission']['configuration'],'SOUND.CFG')
        self.assertEqual(a['driver_name_admission']['read_bytes'],13)
        self.assertFalse(a['driver_name_admission']['name_allowlist'])
        self.assertEqual(a['driver_name_admission']['exec_site'],0x665a)
        callback=a['callback_root']
        self.assertEqual(callback['root'],0x4517)
        self.assertEqual(callback['registration_interrupt'],0x634f)
        self.assertIsNone(callback['invocation_target_set'])
        self.assertIsNone(callback['entry_ss'])
        self.assertIsNone(callback['entry_sp'])
        self.assertEqual(len(a['drivers']),11)
        self.assertTrue(all(d['invocation_producers'] for d in a['drivers']))
        self.assertTrue(all(not d['invocation_candidates_complete'] for d in a['drivers']))
        self.assertFalse(a['int66']['vector_writers_complete'])
        self.assertIsNone(a['int66']['installed_target_set'])
        self.assertEqual(a['int33']['contract_status'],'UNKNOWN')
        self.assertIsNone(a['int33']['invariant_contract'])
        self.assertFalse(a['startup_int21_required_for_callback_stack'])

    def test_real_driver_name_path_mutation_invalidates_derivation(self):
        import types
        import audit_10min_demo_admission as admission
        dec=domains.graph.Decoder()
        r=types.SimpleNamespace(b=self.binary,base=0x300,ds=0x19db0,dec=dec,cs=dec.cs,seen={})
        at=0x663a
        while at<=0x665a:
            x=dec.instruction(r.b,r.base,at);r.seen[at]=x;at+=x.size
        self.assertFalse(admission.external_driver_name(r)['name_allowlist'])
        changed=bytearray(self.binary);changed[0x6658]^=1;r.b=bytes(changed)
        with self.assertRaisesRegex(ValueError,'consumer drift'):
            admission.external_driver_name(r)

    def test_real_code2_matched_stack_depth_is_conditional_not_a_caller_assumption(self):
        e=self.result['object_writer_evidence'][1]['placement_effects']
        local=e['conditional_module_stack_summary']
        self.assertEqual(local['status'],'BOUNDED')
        self.assertEqual(local['maximum_downward_bytes'],10)
        self.assertEqual(local['register_outputs']['es'],['literal',0xa000])
        self.assertEqual(local['effects']['ds'],'preserving')
        self.assertEqual(local['effects']['ss'],'preserving')
        self.assertEqual(local['effects']['sp'],'preserving')
        self.assertEqual(local['conditional_boundaries'],[0xafa6,0xaff7])
        sites={v['site']:v for v in e['implicit_sites']}
        self.assertEqual(sites[0xaedb]['conditional_incoming_sp_relative'],[-6])
        self.assertEqual(sites[0xafa6]['conditional_incoming_sp_relative'],[-8])
        self.assertEqual(sites[0xaff7]['conditional_incoming_sp_relative'],[-8])
        self.assertTrue(all(v['incoming_domains']['sp'] is None for v in sites.values()))

    def test_real_table1_stack_groups_partition_all_sites(self):
        scan=self.result['object_writer_evidence'][0];e=scan['placement_effects']
        self.assertEqual(len(e['stack_groups']),7)
        self.assertEqual(sum(v['number_of_sites'] for v in e['stack_groups']),1420)
        self.assertEqual({p for v in e['stack_groups'] for p in v['sites']},set(scan['implicit_stack_writers']))
        self.assertEqual(e['closed_stack_site_count'],0)
        self.assertEqual(e['unresolved_stack_site_count'],1420)
        self.assertEqual(e['excluded_architectural_write_count'],1)
        closed=[v for v in e['implicit_sites'] if v['architectural_write_status']=='BOUNDED']
        self.assertEqual([v['site'] for v in closed],[0x32ab])
        self.assertEqual(closed[0]['bytes_written'],6)
        self.assertEqual(closed[0]['status'],'UNKNOWN')
        self.assertTrue(closed[0]['handler_physical_write_domain'])
        self.assertTrue(e['interrupts'])  # Architectural write closure is not handler closure.
        self.assertEqual(len(e['boundaries']),12)
        self.assertTrue(all(v['classification']=='UNKNOWN' for v in e['boundaries']))
        far=[v for v in e['boundaries'] if v.get('target_module')=='CODE2']
        self.assertEqual(len(far),4)
        self.assertTrue(all(v['maximum_callee_downward_bytes']==10 for v in far))

    def test_added_code2_target_with_unknown_ss_or_memory_reopens_effects(self):
        import audit_10min_demo_stack as stack
        c=self.result['related_code2'];contract=self.result['placement_contract']
        objects=self.result['object_writer_evidence'][1]['objects']
        for mutation in ('8ed0c3','36c606000001c3'):
            b=bytearray(self.binary);added=0x19da0;code=bytes.fromhex(mutation)
            b[added:added+len(code)]=code
            sub=domains.Resolver(b,0xaed0,0x19db0,set(c['root_files'])|{added},0x19db0)
            sub.prepare();sub.facts={f['source']:dict(f) for f in c['domains']}
            sub.facts[0xafa6]['targets']=sub.facts[0xafa6]['targets']+[added]
            scan=writers.scan(sub,objects,contract,module='CODE2')
            boundary=next(v for v in scan['placement_effects']['boundaries'] if v['site']==0xafa6)
            self.assertEqual(boundary['classification'],'UNKNOWN')
            self.assertFalse(boundary['candidate_effects_complete'])
            self.assertEqual(set(boundary['candidate_targets']),set(c['domains'][0]['targets'])|{added})

    def test_default_command_fails_and_code2_unknown_cannot_be_hidden_by_counter(self):
        fake=dict(self.result,unknown_domain_count=0,open_scope_gates=[],
                  domains=[dict(f,status='BOUNDED') for f in self.result['domains']])
        with self.assertRaisesRegex(ValueError,'unresolved'):
            domains.require_closed(fake)
        with tempfile.TemporaryDirectory(prefix='pf-dmo0-gate-') as directory:
            output=Path(directory)/'evidence.json'
            cmd=[sys.executable,str(Path(domains.__file__)),
                 '--data',str(self.paths[0]),'--canonical',str(self.paths[1]),
                 '--historical',str(self.paths[2]),'--output',str(output)]
            default=subprocess.run(cmd,capture_output=True,text=True)
            self.assertEqual(default.returncode,2,default.stderr)
            self.assertTrue(output.is_file())
            inspection=subprocess.run(cmd+['--allow-open'],capture_output=True,text=True)
            self.assertEqual(inspection.returncode,0,inspection.stderr)
            self.assertIn('unknown: 12',inspection.stdout)


if __name__=='__main__':
    unittest.main()
