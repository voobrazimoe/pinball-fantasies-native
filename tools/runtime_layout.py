#!/usr/bin/env python3
"""Generate address/hash-only runtime profiles from exact private oracle inputs.
No payload is exported. Run explicitly with --data; never regenerate in tests.
The read map mirrors assets, physics/data.go and presentation/table record readers.
"""
import argparse, hashlib, json, pathlib, re, struct
ROOT = pathlib.Path(__file__).resolve().parent.parent

def embedded(path):
    return json.loads(re.search(r'`(.*)`', (ROOT/path).read_text()).group(1))

def generate(data):
    inventory = {r['name']: r for r in json.loads((ROOT/'analysis/game-inventory.json').read_text())}
    out = {}
    matrix = embedded('internal/presentation/content.go')
    for table in range(5):
        name = 'INTRO.PRG' if table == 0 else f'TABLE{table}.PRG'
        b = (data/name).read_bytes()
        assert hashlib.sha256(b).hexdigest() == inventory[name]['sha256'], name
        regions, pictures = [], []
        def region(at, size, why, exact=False):
            assert at >= 0 and at+size <= len(b), (name,at,size)
            if size:
                r = dict(offset=at, size=size, purpose=why)
                if exact: r['sha256'] = hashlib.sha256(b[at:at+size]).hexdigest()
                regions.append(r)
        def picture(at):
            assert b[at:at+4] == b'FORM'
            end = at+8+struct.unpack_from('>I',b,at+4)[0]
            pos = at+12
            while pos<end:
                n=struct.unpack_from('>I',b,pos+4)[0]
                if b[pos:pos+4]==b'BMHD':
                    h=b[pos+8:pos+8+n]
                    pictures.append(dict(offset=at, kind=b[at+8:at+12].decode(), width=struct.unpack_from('>H',h)[0], height=struct.unpack_from('>H',h,2)[0], planes=h[8]))
                pos+=8+n+(n&1)
        if table == 0:
            for at in [0x1cb10,0x6b70,0x8b00,0x20430,0x24a50,0x29410,0x2dd40,0x3b810,0x42c50,0x46930,0x4daa0,0x12020,0x17f20,0x10e60,0xa450,0x34d90,0x33410]: picture(at)
            region(233806,240,'sidebar/options text')
        else:
            i=table-1
            for at in [[336944,366176,399776,437168],[0x50730,0x583f0,0x60030,0x67b00],[0x4cb60,0x52410,0x5a6b0,0x634d0],[0x4bc10,0x54a00,0x5da70,0x66e20]][i]: picture(at)
            # Opaque foreground artwork: no pointers or native behavioral assumptions.
            for at in [[0x300b0,0x35f30],[0x2d4f0,0x33370],[0x1f870,0x256f0],[0x28f30,0x2edb0]][i]: region(at,23040,'foreground bitplane')
            masks=[[0x41330,0x3b930,0x46d30,0x71b30,0x77530,0x7cf30],[0x3e770,0x38d70,0x44170,0x6d7c0,0x731c0,0x78bc0],[0x30af0,0x2b0f0,0x364f0,0x6a3f0,0x6fdf0,0x757f0],[0x3a1b0,0x347b0,0x3fbb0,0x6e870,0x74270,0x79c70]][i]
            for j,at in enumerate(masks): region(at,20400 if table<=2 and j>=4 else 23040,'collision mask tied to native table regions',True)
            at,n=[(0xc2e0,55888),(0xbad0,54288),(0xb570,0x18b60-0xb570),(0xca80,0x166d0-0xca80)][i]
            region(at,n,'BALLCODE delta lookup',True)
            region([0x1e340,0x1d570,0x1ca40,0x1b2c0][i],5120,'sine lookup',True)
            at=[0x1c05d,0x1b00b,0x1a93d,0x18e77][i]
            for j in range(8): region(at+16*j,10,'material parameters',True)
            desc=[0x20690,0x1f820,0x1f230,0x1da30][i]
            frames=[[0x4c730,0x4fe10,0x4ed50],[0x49b70,0x4e110,0x4c190],[0x3bef0,0x3f4a0,0x3e510],[0x455b0,0x47bd0,0]][i]
            for j in range(3):
                at=desc+60*j
                region(at,1,'flipper kind',True)
                if not b[at]: break
                for start,n in [(2,22),(26,16),(48,2),(54,4)]: region(at+start,n,'flipper descriptor',True)
                words,h,fc=struct.unpack_from('<H',b,at+6)[0],struct.unpack_from('<H',b,at+8)[0],struct.unpack_from('<H',b,at+32)[0]
                region(frames[j],h*words*2*(fc+1),'flipper collision frames',True)
            at,n=[(0x82930,1748),(0x7e5c0,168),(0x7b1f0,220),(0x7f670,120)][i]
            region(at,n,'flipper graphics')
            palette=re.findall(r'0x[0-9a-f]+',(ROOT/'internal/assets/ball_locations.go').read_text())
            for at in sorted(set(int(x,16)+[0,-0x810,-0xd70,0x7a0][i] for x in palette)): region(at,1,'PUTBALL literal palette')
            c=matrix[str(table)]
            scrolls={cmd['args'][0] for cmd in c['commands']+c.get('attract',[]) if cmd['op']=='_SCROLL'}
            for label,(at,n) in c['text_refs'].items():
                region(at,n,'matrix record '+label,(n==3 and (label.startswith('S_') or label.startswith('SJINGLE') or 'JINGLE' in label)) or label in ('MATRIXON','MATRIXOFF'))
                if n==4 and label.startswith('S'): region(at,2,'sound sample/note '+label,True)
                if label in scrolls and n: regions[-1]['kind']='scroll'

            for label,(at,count,base,prefix) in c['animation_refs'].items():
                region(at-2*prefix,4+4*count+2*prefix,'animation control '+label,True)
                for j in range(count):
                    pos=base+struct.unpack_from('<H',b,at+4+j*4)[0]
                    for plane in range(2):
                        n=struct.unpack_from('<H',b,pos)[0]
                        region(pos,2,'bitmap count',True);region(pos+2,n,'matrix bitmap');pos+=2+n
            at,count=c['lamp_flash_ref'];region(at,count*10,'lamp scheduling',True)
            for height,at in c['Fonts'].items() if 'Fonts' in c else c['fonts'].items(): region(at,(42 if int(height)==13 else 36)*int(height),'matrix font')
            region(c.get('Spring',c.get('spring')),230,'spring artwork')
            # Literal glyph address tables and store records; store syntax remains data.
            ds=[0x19d40,0x18ee0,0x18b60,0x166d0][i];shift=[0,0x90,-0x720,0x5f0][i];code=[0xae60,0xa650,0xa0f0,0xb600][i]
            tabs=[(0x5a00+shift,code+0x1a0,48,57),(0x5c00+shift,code+0xc40,48,57),(0x6000+shift,0x300+[0x7090,0x6880,0x6320,0x7830][i],32,96),(0x5e00+shift,0x300+[0x7c00,0x73f0,0x6e90,0x83a0][i],32,96)]
            for tab,base,first,last in tabs:
                chars=list(range(first,last+1))+([1,255] if first==32 else [])
                for char in chars:
                    region(ds+tab+2*char,2,'glyph pointers',True)
                for char in chars:
                    at=base+struct.unpack_from('<H',b,ds+tab+char*2)[0];end=at
                    while b[end]!=0xc3:
                        assert b[end]==0x88 and b[end+1] in (0x87,0xa7)
                        end+=4
                    region(at,end-at+1,'literal glyph stores');regions[-1]['kind']='glyph'
            if table==1:
                at=0x1adb9
                for j in range(56):
                    n=b[at+1]*3; region(at,2,'lamp packet header',True);region(at+2,n,'lamp RGB');at+=2+n
                for at,n in [(0x205f0,30),(0x20630,30),(0x20680,15),(0x20610,30),(0x20650,30),(0x20670,15)]: region(at,n,'duck collision gate',True)
            else:
                c=embedded(f'internal/{["","","speeddevils","gameshow","stones"][table]}/content.go')
                for lamp in c['lamps'].values():
                    at=lamp.get('offset',lamp.get('ref',[0])[0]);n=2+3*b[at+1]
                    region(at,2,'lamp packet header',True);region(at+2,n-2,'lamp RGB')
                for gate in c.get('gates',[]):
                    for key in ('opened_ref','closed_ref'):
                        at,w,count,stride=gate[key]
                        for j in range(count): region(at+j*stride,w,'gate collision record',True)
                for at,count in c.get('area_refs',{}).values(): region(at,count*10+2,'area handlers/rectangles/terminator',True)
            if table==4: region(0x4a1f0,(152+16)*40,'tower packed bitmap')
        # Deduplicate identical reads without including any gaps/unconsumed bytes.
        unique={(r['offset'],r['size'],r.get('sha256',''),r.get('kind','')):r for r in regions}
        groups={}
        for r in unique.values():
            key=(r['purpose'], bool(r.get('sha256')),r.get('kind',''))
            groups.setdefault(key,[]).append(r)
        compact=[]
        for (purpose,exact,kind),group in groups.items():
            merged=[]
            for r in sorted(group,key=lambda r:(r['offset'],-r['size'])):
                if merged and kind not in ('glyph','scroll') and r['offset']<=merged[-1]['offset']+merged[-1]['size']:
                    prev=merged[-1];prev['size']=max(prev['offset']+prev['size'],r['offset']+r['size'])-prev['offset']
                else: merged.append(dict(r))
            for r in merged:
                if exact: r['sha256']=hashlib.sha256(b[r['offset']:r['offset']+r['size']]).hexdigest()
            compact+=merged
        out[name]=dict(profile='dos-retail-linked-v1',regions=sorted(compact,key=lambda r:r['offset']),pictures=pictures)
    return out

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--data',required=True,type=pathlib.Path);a=p.parse_args()
    target=ROOT/'internal/datalayout/profiles.json';target.parent.mkdir(exist_ok=True)
    profile=generate(a.data)
    # One record per line makes the private read map reviewable without 30k lines.
    lines=['{']
    for index,(name,value) in enumerate(profile.items()):
        lines+=['  '+json.dumps(name)+': {', '    "profile": '+json.dumps(value['profile'])+',', '    "regions": [']
        lines += ['      '+json.dumps(r,separators=(',', ':'))+(',' if j<len(value['regions'])-1 else '') for j,r in enumerate(value['regions'])]
        lines += ['    ],', '    "pictures": '+json.dumps(value['pictures'],separators=(',', ':')), '  }'+(',' if index<len(profile)-1 else '')]
    lines+=['}']
    target.write_text('\n'.join(lines)+'\n')
