#!/usr/bin/env python3
"""Read-only SDEV content archaeology. Extract data declarations and tracker
control metadata, never ASM instructions or CPU execution. Generates native
content metadata and independent original indexed-art hashes."""
import re,json,struct,hashlib
from pathlib import Path
import prg_content_layout
import matrix_operand_schema as schema
s=Path('reference/original-dos-source/SDEV.ASM').read_bytes().decode('latin1')
lines=s.upper().splitlines()
# MASM COMMENT\ blocks and disabled demo content are not data.
comment=False;demo=False
for i,l in enumerate(lines):
 t=l.strip()
 if t.startswith('COMMENT\\'):comment=True;lines[i]='';continue
 if comment:
  lines[i]=''
  if t=='\\':comment=False
  continue
 if t=='IF DEMOVER':demo=True
 if demo:
  lines[i]=''
  if t=='ENDIF':demo=False
const={'SW':336,'TOTCENT':16,'SPEED':4}
def val(x):
 return schema.resolve_expression(x.strip().upper().replace("OFFSET ", ""),const)
for l in lines:
 m=re.match(r'\s*(\w+)\s*(?:=|EQU\s+)([^;]+)',l)
 if m:
  try:const[m[1]]=val(m[2])
  except (KeyError,ValueError,SyntaxError):pass
# EMPTYJINGLE is referenced by the matrix stream but declared in a lost include.
# Recover it from the table's own S_Empty jingle record.
assert schema.seed_empty_jingle(const,lines),'EMPTYJINGLE'
E={};J={}
for i,l in enumerate(lines):
 m=re.match(r'^(\w+)\s+DW\s+(\w+)\s*(?:;.*)?$',l)
 if m:
  rows=[r.split(';')[0].strip() for r in lines[i+1:i+15] if r.split(';')[0].strip()][:6];priority=0
  if len(rows)<3:continue
  if re.fullmatch(r'DB\s+\d+',rows[0]):priority=int(rows.pop(0).split()[1])
  if all(len(re.findall(r'\d+',r))==12 for r in rows[:2]):
   tail=re.match(r'DW\s+(\w+)',rows[2])
   if tail:
    nums=lambda row:int(''.join(re.findall(r'\d+',row)))
    E[m[1]]={'jingle':m[2],'priority':priority,'matrix':tail[1],'score':nums(rows[0]),'bonus':nums(rows[1])}
 m=re.match(r'^(S_\w+|SJINGLE\w+)\s+DB\s+(\d+),(\d+),(\d+)',l)
 if m:J[m[1]]={'position':int(m[2]),'repeat':int(m[3]),'priority':int(m[4])}
commands=[];labels={}
for l in lines[1187:1935]:
 l=l.split(';')[0].strip()
 m=re.match(r'(\w+)\s+(?:LABEL\s+\w+|DW\s+)',l)
 if m:labels[m[1]]=len(commands)
 m=re.search(r'\bDW\s+(.+)',l)
 if m:
  a=[x.strip().replace('OFFSET ','') for x in m[1].split(',')]
  if a[0].startswith('_') or a[0]=='0':
   c={'op':a[0],'args':a[1:]}
   commands.append(c)
 elif re.fullmatch(r'CLEARIT[234]?',l):commands.append({'op':'_CLEAR'+(l[-1] if l[-1].isdigit() else '1'),'args':[]})
# Resolve every numeric operand from the original assembler expression.  The
# raw text stays in args for diagnostics; the runtime consumes nums.
for c in commands:
 try:n=schema.numeric_values(c['op'],c['args'],const)
 except schema.OperandError as e:raise SystemExit(f'SDEV.ASM: {e}')
 if n:
  c['nums']={str(i):v for i,v in n.items()}
  if c['op']=='_WAIT':c['ticks']=n[0]
anims={};headers=[];cur=None;const['SPEED']=4
for l in lines[1940:2210]:
 l=l.split(';')[0].strip();m=re.match(r'(\w+)\s*=\s*(.+)',l)
 if m:
  try:const[m[1]]=val(m[2])
  except (KeyError,ValueError,SyntaxError):pass
  continue
 m=re.match(r'(?:(\w+)\s+)?DW\s+(.+)',l)
 if not m:continue
 a=[x.strip() for x in m[2].split(',')]
 if len(a)==1:
  try:headers.append(val(a[0]))
  except (KeyError,ValueError,SyntaxError):continue
  cur=None
 elif len(a)==2:
  if m[1] and len(headers)==2:headers.insert(0,0)
  if m[1]:cur=m[1];anims[cur]={'header':headers[-3:],'frames':[]};headers=[]
  if cur:
   try:anims[cur]['frames'].append(val(a[1]))
   except (KeyError,ValueError,SyntaxError):pass
scrolls={}
for l in lines:
 l=l.split(';')[0].strip();m=re.match(r'(\w+)\s+DB\s+(.+)',l)
 if not m:continue
 try:scrolls[m[1]]=schema.db_byte_count(m[2],const)
 except schema.OperandError:
  if any(c['op']=='_SCROLL' and c['args'][0]==m[1] for c in commands):raise
  # Non-text storage may use SIZE structures from missing includes.
  continue
b=Path('TABLE2.MOD').read_bytes();orders=list(b[952:952+b[950]]);cues={};effects=set()
for p in range(max(orders)+1):
 for r in range(64):
  for c in range(4):
   q=b[1084+p*1024+r*16+c*4:1088+p*1024+r*16+c*4];effects.add((q[2]&15,q[3]))
for pos in range(len(orders)):
 speed=6;ticks=0;order=pos;row=0;segments=[{'position':pos,'start':0}]
 for _ in range(8192):
  jump=br=None
  for c in range(4):
   q=b[1084+orders[order]*1024+row*16+c*4:1088+orders[order]*1024+row*16+c*4];e=q[2]&15;v=q[3]
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
 else:raise ValueError(('missing cue',pos))
# LON palette records occur in source order, which is not lamp-number order.
lamps={};blocks=[];name=None
for l in lines[758:1100]:
 l=l.split(';')[0].strip();m=re.match(r'(?:(LON\d+)\s+)?DB\s+(.+)',l)
 if not m:continue
 if m[1]:name=m[1];blocks.append((name,[]))
 if name:
  try:blocks[-1][1].extend(val(x) for x in m[2].split(','))
  except (KeyError,ValueError,SyntaxError):break
linked=Path('TABLE2.PRG').read_bytes()
# Verify source timing declarations against linked DATA2. Non-looping anims
# have only two header words; the normalized loop origin is unused for them.
for name,a in anims.items():
 header=struct.pack('<2H',*a['header'][1:])
 assert any(linked[o:o+4]==header and all(
  struct.unpack_from('<H',linked,o+6+4*i)[0]==v
  for i,v in enumerate(a['frames']))
  for o in range(0x1f8d4,0x22000)),('linked animation timing',name)
for name,values in blocks:
 values=values[:2+3*values[1]]
 raw=bytes(values);off=linked.find(raw,0x18ee0)
 assert off!=-1,(name,raw)
 lamps[name[3:]]={'start':values[0],'count':values[1],'offset':off+2}
attract=[]
for l in lines[705:752]:
 l=l.split(';')[0].strip();m=re.match(r'(\w+)\s*=\s*(.+)',l)
 if m:const[m[1]]=val(m[2]);continue
 m=re.match(r'DW\s+(.+)',l)
 if m:
  a=[val(x) for x in m[1].split(',')]
  if len(a)==5:attract.append(a[1:])
out={'effects':E,'jingles':J,'commands':commands,'labels':labels,'animations':anims,'scrolls':scrolls,'cues':cues,'lamps':lamps,'attract':attract}
Path('internal/speeddevils/content.go').write_text('// Generated by tools/reference_pf7.py from SDEV data declarations.\npackage speeddevils\nconst contentJSON = `'+json.dumps(prg_content_layout.locate_table_records(2,out),separators=(',',':'))+'`\n')
# Independent strip decoder from PF1; no production Go dependency.
import reference_extract as ex
ex.OFFSETS=(0x50730,0x583f0,0x60030,0x67b00)
art,rgba=ex.extract('TABLE2.PRG');d=Path('TABLE2.PRG').read_bytes()
locations=json.loads(Path('analysis/pf2-ball-locations.json').read_text())['offsets']
ball=bytes(d[o-0x810] if o else 0 for o in locations)
mask=d[0x2d4f0:0x2d4f0+23040];palette=d[d.index(b'CMAP',0x67b00)+8:][:768];full=bytearray(rgba)
for i,c in enumerate(ball):
 x,y=300+i%16,530+i//16
 if c and not mask[y*40+x//8]&(128>>(x%8)):full[(y*320+x)*4:(y*320+x)*4+3]=palette[c*3:c*3+3]
art.update(initial_rgba_sha256=ex.digest(full[259*1280:]),ball_sha256=ex.digest(ball),foreground_sha256=ex.digest(mask),tracker_effects=sorted(set(e for e,v in effects)),start=[300,530],hidden_start=[285,530])
Path('analysis/pf7-assets.json').write_text(json.dumps(art,indent=2)+'\n')
print('Effects',len(E),'jingles',len(J),'matrix commands',len(commands),'animations',len(anims),'tracker effects',art['tracker_effects'])
print('Missing animations',sorted({c['args'][0] for c in commands if c['op']=='_ANIMATION'}-anims.keys()))
