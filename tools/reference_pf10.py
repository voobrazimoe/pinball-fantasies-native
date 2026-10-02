#!/usr/bin/env python3
"""STONES declaration/content fixtures. No native output and no CPU execution.
Run reference_pf8.py first to locate original matrix data for all four tables.
"""
import json,re,struct,hashlib
from pathlib import Path
import prg_content_layout
import reference_extract as ex
b=Path('TABLE4.PRG').read_bytes();source=Path('reference/original-dos-source/STONES.ASM').read_bytes().decode('latin1').upper().splitlines()
p=json.loads(Path('.private-cleanup/generated/pf8-content.json').read_text())['4']
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
 name=m[1];v=p['texts'][name];v=v[:2+3*v[1]];off=b.find(bytes(v),0x166d0,0x1da30);assert off>=0,name
 lamps[name[3:]]={'start':v[0],'rgb':v[2:]};order.append(int(name[3:]))
gates=[]
# GATE2..5 descriptors are literal linked DATA records.
for name,x,y,w,h in [('GATE2',30,5,2,20),('GATE3',20,5,2,24),('GATE4A',4,191,4,14),('GATE4B',4,247,4,14),('GATE4C',4,303,4,14),('GATE5',0,511,4,13)]:
 sig=struct.pack('<HHH',y*40+x,w,h);loc=b.index(sig,0x166d0,0x1da30)-4
 pos,neg=struct.unpack_from('<HH',b,loc)
 def rows(o):return list(b''.join(b[0x166d0+o+w+3*w*r:0x166d0+o+2*w+3*w*r] for r in range(h)))
 gates.append(dict(name=name,x=x,y=y,width=w,height=h,high=name.startswith('GATE4'),opened=rows(neg),closed=rows(pos)))
# Tracker cue boundaries: independent order/row walk over DATA, not audio code.
mod=Path('TABLE4.MOD').read_bytes();orders=list(mod[952:952+66]);assert mod[950]==66 and max(orders)==63;cues={};fx=set()
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
areas={}
for name in ['AREALISTA_L','AREALISTA_U','AREALISTA_L_T','AREALISTA_U_T']:
 start=next(i for i,l in enumerate(lines) if l.startswith(name+'\t') or l.startswith(name+' '));rs=[]
 for l in lines[start+1:]:
  if l=='DW 0' or re.fullmatch(r'DW\s+0',l):break
  m=re.fullmatch(r'DW\s+(\d+),(\d+),(\d+),(\d+)',l)
  if m:rs.append(dict(rect=list(map(int,m.groups())),handler=''))
  elif rs:
   m=re.fullmatch(r'DW\s+(\w+)',l)
   if m:rs[-1]['handler']=m[1]
 areas[name]=rs
out=dict(effects=effects,jingles=jingles,cues=cues,lamps=lamps,gates=gates,areas=areas)
# Preserve source order independently of tracker walk variable.
out['lamp_order']=[int(re.match(r'LON(\d+)\s+DB',l)[1]) for l in lines if re.match(r'LON(\d+)\s+DB',l)]
Path('internal/stones').mkdir(exist_ok=True)
Path('internal/stones/content.go').write_text('// Generated from STONES declarations by tools/reference_pf10.py.\npackage stones\nconst contentJSON = `'+json.dumps(prg_content_layout.locate_table_records(4,out),separators=(',',':'))+'`\n')
ex.OFFSETS=(0x4bc10,0x54a00,0x5da70,0x66e20);art,rgba=ex.extract('TABLE4.PRG', {0x66e20:1219})
locations=json.loads(Path('analysis/pf2-ball-locations.json').read_text())['offsets'];ball=bytes(b[o+0x7a0] if o else 0 for o in locations);mask=b[0x28f30:0x28f30+23040];palette=b[b.index(b'CMAP',0x66e20)+8:][:768];full=bytearray(rgba)
for i,c in enumerate(ball):
 x,y=297+i%16,530+i//16
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
art.update(sample_start=sample_start,sample_hashes=sample_hashes,header_orders=66,referenced_orders=66,initial_rgba_sha256=ex.digest(full[259*1280:]),ball_sha256=ex.digest(ball),foreground_sha256=ex.digest(mask),start=[297,530],hidden_start=[282,530],tasks=20,flashes=64,tracker_effects=sorted(fx),animations=animations,mod_sha256=ex.digest(mod))
Path('analysis/pf10-assets.json').write_text(json.dumps(art,indent=2)+'\n')
print('STONES:',len(effects),'effects',len(jingles),'jingles',len(p['animations']),'animations',len(lamps),'lamps; tracker effects',sorted(fx))
# Source BALLCODE first sync: two passes at gravity8, then RampTable_hi[0]
# establishes gravity10 before the late pass. No contact exists in SETBALL chute.
art['first_sync']={'x':297*1024+30,'y':530*1024+24,'vx':10,'vy':26,'gx':0,'gy':10}
art['geometry']={'stage_offsets':list(ex.OFFSETS),'viewport':[0,259,320,576],
 'flipper_descriptor':0x1da30,'flipper_stride':30,'flipper_graphics':0x7f670,
 'flipper_frames':[0x455b0,0x47bd0], 'flipper_count':2,
 'gravities':[[0,10],[-10,5],[0,-10],[5,0],[5,15],[-10,12],[2,15],[-8,12],[3,10],[4,13],[7,10]],
 'masks':{name:{'offset':o,'length':23040,'sha256':ex.digest(b[o:o+23040])} for name,o in [('hid1',0x28f30),('hid2',0x2edb0),('mask12',0x347b0),('mask11',0x3a1b0),('mask22',0x3fbb0),('mask13',0x6e870),('mask21',0x74270),('mask23',0x79c70)]}}
art['spring']={'offset':p['spring'],'size':[10,23],'dest':[304,556],'sha256':ex.digest(b[p['spring']:p['spring']+230]),'release32':-166*32}
art['hi_sha256']=ex.digest(Path('TABLE4.HI').read_bytes())
art['hi_size']=len(Path('TABLE4.HI').read_bytes())
art['prg_size']=len(b);art['mod_size']=len(mod);art['sample_lengths']=sum(struct.unpack_from('>H',mod,42+30*i)[0]*2 for i in range(31))
art['tower_windows']={str(row):ex.digest(bytes(bool(b[0x4a1f0+(row+y)*40+x//4]&(1<<(7-2*(x%4)))) for y in range(16) for x in range(160))) for row in [151,73,47,20]}
# Independent FLIPPRA source stream decoding, starting with original indexed PBM.
field=bytearray()
for strip in ex.OFFSETS:
 end=strip+8+struct.unpack_from('>I',b,strip+4)[0];q=strip+12
 while q<end:
  n=struct.unpack_from('>I',b,q+4)[0]
  if b[q:q+4]==b'BODY':
   src=b[q+8:q+8+n];v=bytearray();k=0
   while k<len(src):
    c=struct.unpack_from('b',src,k)[0];k+=1
    if c>=0:v.extend(src[k:k+c+1]);k+=c+1
    elif c!=-128:v.extend(bytes([src[k]])*(1-c));k+=1
   field.extend(v[:320*144])
  q+=8+n+(n&1)
# First held-left flipper pass: speed=-68, angle68, frame=floor(68/55)=1.
record=0xca80;count=struct.unpack_from('<H',b,record)[0]
for i in range(count):
 dst,src=struct.unpack_from('<HH',b,record+18+4*i);src-=0xd4f4
 for plane in range(4):field[(510+dst//84)*320+80+4*(dst%84)+plane]=b[0x7f670+30*plane+src]
art['first_left_flipper_indices_sha256']=ex.digest(field)
# Area DATA is private/native migration work, not a public hash fixture.
Path('analysis/pf10-assets.json').write_text(json.dumps(art,indent=2)+'\n')
