#!/usr/bin/env python3
"""Check private C/D identity and compile only C INTRO and factory seed metadata.
TABLE mappings and typed Stones selectors are shared with deluxe_cd.json.
No payload is emitted; production does not run this tool.
"""
import argparse
import copy
import hashlib
import json
import struct
from pathlib import Path
from deluxe_cd_layout import ROOT, fingerprint, FINGERPRINT, merge

ALT_FINGERPRINT = '6b680be66b3c53286db0e658d48a5c5475e1fb97f67d06374bb1b743a7e5fe33'
ALT_PROFILE = 'dos-deluxe-cd-alt-linked-v1'

def generate(evidence, data, deluxe):
    assert fingerprint(data) == ALT_FINGERPRINT
    assert fingerprint(deluxe) == FINGERPRINT
    audit = next(x for x in json.loads(evidence.read_text()) if x['fingerprint'] == ALT_FINGERPRINT)
    shared = json.loads((ROOT/'internal/datalayout/deluxe_cd.json').read_text())
    intro = copy.deepcopy(shared['INTRO.PRG'])
    b, d = (data/'INTRO.PRG').read_bytes(), (deluxe/'INTRO.PRG').read_bytes()
    f = audit['files']['INTRO.PRG']
    by_dest = {p['offset']:p for p in f['pictures']}
    canonical = json.loads((ROOT/'internal/datalayout/profiles.json').read_text())['INTRO.PRG']['pictures']
    spans = []
    for dst, old in zip(canonical, intro['source_profile']['pictures']):
        # These source addresses are reviewed full-build segment targets, not ordinals.
        at = {0x3b810:0x3b960, 0x42c50:0x3d940}.get(dst['offset'], by_dest[dst['offset']]['match_offset'])
        size = 8 + struct.unpack_from('>I', b, at+4)[0]
        old_size = 8 + struct.unpack_from('>I', d, old['offset']+4)[0]
        assert size == old_size and b[at:at+size] == d[old['offset']:old['offset']+size]
        old['offset'] = at
        spans.append((dst['offset'], size, at-dst['offset']))
    for r,e in zip(intro['source_profile']['regions'],f['regions']):
        at = e['chosen_offset']
        assert b[at:at+r['size']] == d[r['offset']:r['offset']+r['size']]
        spans.append((e['canonical_offset'], r['size'], at-e['canonical_offset']))
        r['offset'] = at
    intro['copies'] = merge(spans)
    intro['source_profile']['profile'] = ALT_PROFILE
    seeds = {}
    for name,l in shared.items():
        if name == 'INTRO.PRG': continue
        b,d = (data/name).read_bytes(),(deluxe/name).read_bytes()
        r = next(r for r in l['source_profile']['regions'] if r['purpose']=='matrix record HI_SCORE_LIST')
        at = r['offset']; size = r['size']; assert size == 64
        initials = [b[at+16*i+12:at+16*i+15].decode('ascii') for i in range(4)]
        allowed = {at+16*i+12+j for i in range(4) for j in range(3)}
        assert len(b)==len(d)
        differences = {i for i,(x,y) in enumerate(zip(b,d)) if x!=y}
        assert differences == allowed, (name,'unexpected consumed difference')
        seeds[name] = dict(source=at,destination=next(r['offset'] for r in json.loads((ROOT/'internal/datalayout/profiles.json').read_text())[name]['regions'] if r['purpose']=='matrix record HI_SCORE_LIST'),initials=initials,sha256=hashlib.sha256(b[at:at+size]).hexdigest())
        for s in l.get('selectors',[]):
            assert struct.unpack_from('<H',b,s['source'])[0] == s['raw']
        for j in l.get('jingles',[]):
            assert tuple(b[j['source']:j['source']+3]) == (j['position'],j['repeat'],j['priority'])
    return dict(intro=intro,seeds=seeds)

def encode(result): return json.dumps(result,indent=2)+'\n'

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for n in ('evidence','data','deluxe'): p.add_argument('--'+n,type=Path,required=True)
    p.add_argument('--check',action='store_true'); a=p.parse_args()
    out=encode(generate(a.evidence,a.data,a.deluxe)); target=ROOT/'internal/datalayout/deluxe_cd_alt.json'
    if a.check: assert target.read_text()==out,'descriptor drift'
    else: target.write_text(out)
