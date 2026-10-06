#!/usr/bin/env python3
"""Compile reviewed CD-family address/typed metadata from private A/D evidence.
No discovery, payload output or production scanning. C reuses this TABLE map
through the separate alt metadata compiler.
"""
import argparse
import hashlib
import json
import re
import struct
from pathlib import Path
from powerpack_layout import ROOT

FINGERPRINT = '633f1acc51ecc2d6cd313ead2a486bfb4c4ab30b284a22ea16590593d6ee94b8'
PROFILE = 'dos-deluxe-cd-linked-v1'
NAMES = ['INTRO.PRG', 'INTRO.MOD', 'MOD2.MOD'] + [f'TABLE{i}.{ext}' for i in range(1,5) for ext in ('PRG','MOD')]
# Source labels, not an arithmetic selector adjustment. Values select native
# semantic tokens registered in stones/content.go; they are never executed.
HANDLERS = {
 5940:'PARTYOFFAREA',5988:'CLOSE1',6015:'BYGEL15',6016:'BYGEL16',6026:'OPENBUMPERS',
 6089:'BYGEL5',6358:'BYGEL6',6627:'BYGEL7',7268:'GROPA',8379:'GROPA_T',8638:'GROPB',
 9145:'GROPB_T',9157:'BYGEL19',9158:'BYGEL3',9249:'BYGEL4',9340:'BYGEL1',9407:'BYGEL2',
 9474:'GROPC',10738:'CLOSE4',10739:'BYGEL28',10746:'BYGELSI',10753:'BYGEL8',11094:'BYGEL12',
 11285:'BYGEL13',11476:'BYGEL14',11667:'BYGEL8B',11674:'BYGEL11',12275:'BYGEL9',12529:'BYGEL10',
 12640:'OPEN6',12647:'CLOSE6',12654:'OPEN7',12661:'CLOSE7',12668:'BYGEL18'}

def fingerprint(data):
    return hashlib.sha256(''.join(n+'\0'+hashlib.sha256((data/n).read_bytes()).hexdigest()+'\n' for n in NAMES).encode()).hexdigest()

def merge(spans):
    result=[]
    for shift in sorted({s[2] for s in spans}):
        group=[]
        for at,size,delta in sorted(spans):
            if shift != delta: continue
            if group and at<=group[-1][1]: group[-1][1]=max(group[-1][1],at+size)
            else: group.append([at,at+size])
        result += [dict(destination=a,source=a+shift,size=z-a) for a,z in group]
    result.sort(key=lambda r:r['destination'])
    for a,b in zip(result,result[1:]):
        assert a['destination']+a['size']<=b['destination'], ('overlapping translations',a,b)
    return result

def generate(evidence, canonical, data):
    assert fingerprint(data)==FINGERPRINT, 'not audited D (use alt metadata compiler for C)'
    audit=next(s for s in json.loads(evidence.read_text()) if s['fingerprint']==FINGERPRINT)
    profiles=json.loads((ROOT/'internal/datalayout/profiles.json').read_text())
    inventory={r['name']:r for r in json.loads((ROOT/'analysis/game-inventory.json').read_text())}
    content=json.loads(re.search(r'`(.*)`',(ROOT/'internal/stones/content.go').read_text()).group(1))
    native={label:int(token) for token,label in content['area_handlers'].items()}
    assert set(native)==set(HANDLERS.values())
    output={}
    for name,p in profiles.items():
        src,ref=(data/name).read_bytes(),(canonical/name).read_bytes()
        assert hashlib.sha256(ref).hexdigest()==inventory[name]['sha256']
        f=audit['files'][name]
        evidence_regions={(r['canonical_offset'],r['size'],r['purpose']):r for r in f['regions']}
        source_regions=[];spans=[];selectors=[]
        for r in p['regions']:
            e=evidence_regions[(r['offset'],r['size'],r['purpose'])]
            dest,at,size=r['offset'],e['chosen_offset'],r['size']
            if r['purpose']=='tower packed bitmap':size=6680
            observed,expected=src[at:at+size],ref[dest:dest+size]
            assert len(observed)==size
            if r['purpose']=='area handlers/rectangles/terminator':
                for list_name,(start,count) in content['area_refs'].items():
                    for i in range(count):
                        dst=start+10*i;origin=at+dst-dest
                        assert src[origin:origin+8]==ref[dst:dst+8]
                        raw=struct.unpack_from('<H',src,origin+8)[0]
                        label=HANDLERS[raw];token=native[label]
                        assert token==struct.unpack_from('<H',ref,dst+8)[0]
                        selectors.append(dict(destination=dst+8,source=origin+8,raw=raw,native_id=token,role=label))
                        spans.append((dst,8,origin-dst))
                    dst=start+10*count;origin=at+dst-dest
                    assert src[origin:origin+2]==ref[dst:dst+2]==b'\0\0'
                    spans.append((dst,2,origin-dst))
                assert len(selectors)==44
            else:
                if r['purpose']=='matrix record S_EMPTY' and name=='TABLE1.PRG':
                    assert expected==bytes((62,0,1)) and observed==bytes((62,0,0))
                else: assert observed==expected,(name,r)
                spans.append((dest,size,at-dest))
            q=dict(r,offset=at,size=size)
            if 'sha256' in q or r['purpose']=='matrix record HI_SCORE_LIST':q['sha256']=hashlib.sha256(observed).hexdigest()
            source_regions.append(q)
        pictures=[]
        for pic,e in zip(p['pictures'],f['pictures']):
            dest=pic['offset']
            if dest in [0x3b810,0x42c50]:
                at,height={0x3b810:(0x3b840,123),0x42c50:(0x3d820,117)}[dest]
                q=dict(pic,offset=at,height=height)
            else:
                at=e['match_offset'];q=dict(pic,offset=at)
            size=8+struct.unpack_from('>I',src,at+4)[0]
            assert at+size<=len(src)
            if dest not in [0x3b810,0x42c50]:assert src[at:at+size]==ref[dest:dest+size]
            pictures.append(q);spans.append((dest,size,at-dest))
        layout=dict(decoded_size=len(ref),copies=merge(spans),source_profile=dict(profile=PROFILE,regions=source_regions,pictures=pictures))
        if selectors:layout['selectors']=selectors
        if name=='TABLE1.PRG':layout['jingles']=[dict(role='S_EMPTY',destination=0x1a9ac,source=0xe9e9,position=62,repeat=0,priority=0)]
        output[name]=layout
    return output

def encode(result):
    lines=['{']
    for index,(name,l) in enumerate(result.items()):
        lines += ['  '+json.dumps(name)+': {', '    "decoded_size": '+str(l['decoded_size'])+',', '    "copies": [']
        lines += ['      '+json.dumps(r,separators=(',',':'))+(',' if i<len(l['copies'])-1 else '') for i,r in enumerate(l['copies'])]
        lines += ['    ],', '    "source_profile": {', '      "profile": '+json.dumps(PROFILE)+',', '      "regions": [']
        regions=l['source_profile']['regions']
        lines += ['        '+json.dumps(r,separators=(',',':'))+(',' if i<len(regions)-1 else '') for i,r in enumerate(regions)]
        lines += ['      ],', '      "pictures": '+json.dumps(l['source_profile']['pictures'],separators=(',',':')), '    }'+(',' if 'jingles' in l or 'selectors' in l else '')]
        for key in ('jingles','selectors'):
            if key in l:
                lines += ['    '+json.dumps(key)+': '+json.dumps(l[key],separators=(',',':'))]
        lines += ['  }'+(',' if index<len(result)-1 else '')]
    return '\n'.join(lines+['}'])+'\n'

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for n in ('evidence','canonical','data'):p.add_argument('--'+n,type=Path,required=True)
    p.add_argument('--check',action='store_true')
    a=p.parse_args();result=encode(generate(a.evidence,a.canonical,a.data))
    target=ROOT/'internal/datalayout/deluxe_cd.json'
    if a.check:assert target.read_text()==result,'descriptor drift'
    else:target.write_text(result)
