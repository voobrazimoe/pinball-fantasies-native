#!/usr/bin/env python3
"""SHOW declaration/content fixtures. No native output and no CPU execution.
Run reference_pf8.py first to locate original matrix data for all three tables.
"""
import json,re,struct,hashlib
from pathlib import Path
import prg_content_layout
import reference_extract as ex
b=Path('TABLE3.PRG').read_bytes();source=Path('reference/original-dos-source/SHOW.ASM').read_bytes().decode('latin1').upper().splitlines()
p=json.loads(Path('.private-cleanup/generated/pf8-content.json').read_text())['3']
lines=[];comment=False;demo=False
for l in source:
 t=l.strip()
 if t.startswith('COMMENT\\'):comment=True;continue
 if comment:
  if t=='\\':comment=False
  continue
 if t=='IF DEMOVER':demo=True;continue
 if demo:
  if t=='ENDIF':demo=False
  continue
 lines.append(l.split(';')[0].strip())
effects={};jingles={}
# Mutated score records (JACKVALUE/CASHPOTVAL/RAISING_M) are named DB rows.
for i,l in enumerate(lines):
 m=re.fullmatch(r'(\w+)\s+DW\s+(\w+|0)',l)
 if not m:continue
 rows=[r for r in lines[i+1:i+10] if r];priority=0
 if rows and re.fullmatch(r'DB\s+\d+',rows[0]):priority=int(rows.pop(0).split()[1])
 if len(rows)<3:continue
 scores=[]
 for row in rows[:2]:
  q=re.search(r'\bDB\s+([\d, ]+)$',row)
  if not q:break
  ns=re.findall(r'\d+',q[1])
  if len(ns)!=12:break
  scores.append(int(''.join(ns)))
 tail=re.fullmatch(r'DW\s+(\w+|0)',rows[2])
 if len(scores)==2 and tail:effects[m[1]]={'jingle':m[2],'priority':priority,'matrix':tail[1],'score':scores[0],'bonus':scores[1]}
for name,vs in p['texts'].items():
 if re.match(r'S_|SJINGLE',name) and len(vs)==3:jingles[name]={'position':vs[0],'repeat':vs[1],'priority':vs[2]}
lamps={};order=[]
for l in lines:
 m=re.match(r'(LON\d+)\s+DB',l)
 if not m:continue
 name=m[1];v=p['texts'][name];v=v[:2+3*v[1]];off=b.find(bytes(v),0x18b60,0x1f230);assert off>=0,name
 lamps[name[3:]]={'start':v[0],'rgb':v[2:]};order.append(int(name[3:]))
gates=[]
# Literal data addresses recovered from opengate/closegate static references.
for n,(opened,closed,x,y,w,h,high) in enumerate([(0x6190,0x61e0,0,96,3,25,True),(0x6230,0x6280,35,305,2,34,False),(0x62d0,0x6370,17,227,2,16,False),(0x62f0,0x6360,17,247,1,16,False),(0x6300,0x6350,4,266,1,16,False),(0x6310,0x6330,3,286,2,16,False),(0x6390,0x6480,22,163,4,20,True),(0x6570,0x6620,0,511,4,14,False)],1):
 def rows(o):
  if n<7:return list(b[0x18b60+o:0x18b60+o+w*h])
  # MOVE_MASK_DATA_B copies width then skips two widths; caller adds width.
  return list(b''.join(b[0x18b60+o+w+3*w*r:0x18b60+o+2*w+3*w*r] for r in range(h)))
 gates.append(dict(number=n,x=x,y=y,width=w,height=h,high=high,opened=rows(opened),closed=rows(closed)))
# Tracker cue boundaries: independent order/row walk over DATA, not audio code.
mod=Path('TABLE3.MOD').read_bytes();orders=list(mod[952:952+63]);assert mod[950]==61 and orders[61:]==[62,63];cues={};fx=set()
for pos in range(len(orders)):
 speed=6;ticks=0;order=pos;row=0;segments=[{'position':pos,'start':0}]
 for _ in range(8192):
  jump=br=None
  for c in range(4):
   q=mod[1084+orders[order]*1024+row*16+c*4:1088+orders[order]*1024+row*16+c*4];e=q[2]&15;v=q[3];fx.add(e)
   if e==15:assert 0<v<32;speed=v
   if e==11:jump=v
   if e==13:br=(v>>4)*10+(v&15)
   assert not(e==14 and v>>4 in (6,14))
  ticks+=speed
  if jump is not None:cues[str(pos)]={'tracker_ticks':ticks,'next':jump,'segments':segments};break
  if br is not None:order=(order+1)%len(orders);row=br;segments.append({'position':order,'start':ticks})
  else:
   row+=1
   if row==64:row=0;order=(order+1)%len(orders);segments.append({'position':order,'start':ticks})
 else:raise ValueError(('cue',pos))
out=dict(effects=effects,jingles=jingles,cues=cues,lamps=lamps,gates=gates)
# Preserve source order independently of tracker walk variable.
out['lamp_order']=[int(re.match(r'LON(\d+)\s+DB',l)[1]) for l in lines if re.match(r'LON(\d+)\s+DB',l)]
Path('internal/gameshow').mkdir(exist_ok=True)
Path('internal/gameshow/content.go').write_text('// Generated from SHOW declarations by tools/reference_pf9.py.\npackage gameshow\nconst contentJSON = `'+json.dumps(prg_content_layout.locate_table_records(3,out),separators=(',',':'))+'`\n')
ex.OFFSETS=(0x4cb60,0x52410,0x5a6b0,0x634d0);art,rgba=ex.extract('TABLE3.PRG')
locations=json.loads(Path('analysis/pf2-ball-locations.json').read_text())['offsets'];ball=bytes(b[o-0xd70] if o else 0 for o in locations);mask=b[0x1f870:0x1f870+23040];palette=b[b.index(b'CMAP',0x634d0)+8:][:768];full=bytearray(rgba)
for i,c in enumerate(ball):
 x,y=299+i%16,530+i//16
 if c and not mask[y*40+x//8]&(128>>(x%8)):full[(y*320+x)*4:(y*320+x)*4+3]=palette[c*3:c*3+3]
animations={}
for name,a in p['animations'].items():
 dots=[False]*2560;hashes=[]
 for off in a['offsets']:
  q=off
  for plane in range(2):
   count=struct.unpack_from('<H',b,q)[0];q+=2;di=167
   for v in b[q:q+count]:
    di+=v>>1
    if v==254:continue
    x=(di%84)*2+plane;y=di//168-1
    if 0<=x<160 and 0<=y<16:dots[y*160+x]=bool(v&1)
   q+=count
  hashes.append(ex.digest(bytes(dots)))
 animations[name]=hashes
sample_start=1084+(max(mod[952:1080])+1)*1024
sample_hashes=[];q=sample_start
for i in range(31):
 n=struct.unpack_from('>H',mod,42+30*i)[0]*2
 sample_hashes.append(ex.digest(mod[q:q+n]));q+=n
assert q==len(mod)
art.update(sample_start=sample_start,sample_hashes=sample_hashes,header_orders=61,referenced_orders=63,initial_rgba_sha256=ex.digest(full[259*1280:]),ball_sha256=ex.digest(ball),foreground_sha256=ex.digest(mask),start=[299,530],hidden_start=[284,530],tasks=20,flashes=30,tracker_effects=sorted(fx),animations=animations,mod_sha256=ex.digest(mod))
Path('analysis/pf9-assets.json').write_text(json.dumps(art,indent=2)+'\n')
print('SHOW:',len(effects),'effects',len(jingles),'jingles',len(p['animations']),'animations',len(lamps),'lamps; tracker effects',sorted(fx))
