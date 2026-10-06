#!/usr/bin/env python3
"""Private bounded-domain research tests; fixture absent -> clean SKIP."""
import os
from pathlib import Path
import struct
import unittest
from unittest.mock import patch

import audit_10min_demo_domains as domains
import audit_10min_demo_graph as graph


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


if __name__=='__main__':
    unittest.main()
