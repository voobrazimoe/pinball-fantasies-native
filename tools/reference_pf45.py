#!/usr/bin/env python3
"""Extract Party Land timing CONTENT tables, not executable instructions.
Independent oracle uses only ASM data directives and MOD row control effects.
No samples, renderer, CPU emulation or instruction translation.
"""
import re,json,hashlib
from pathlib import Path
import prg_content_layout
import os
import matrix_operand_schema as schema
s=Path('reference/original-dos-source/PLAND.ASM').read_bytes().decode('latin1')
lines=s.splitlines()
# Select the retail non-demo data branches; preserve indices for provenance.
stack=[]; active=True; filtered=[]
for l in lines:
 t=l.strip().upper()
 if t=='IF DEMOVER':stack.append((active,False));active=False;filtered.append('');continue
 if t=='ELSE' and stack:
  parent,branch=stack[-1];active=parent and not branch;stack[-1]=(parent,not branch);filtered.append('');continue
 if t=='ENDIF' and stack:active=stack.pop()[0];filtered.append('');continue
 filtered.append(l if active else '')
lines=filtered
const={'SW':336,'TOTCENT':16}
def val(x):
 return schema.resolve_expression(x.strip().upper().replace("OFFSET ", ""),const)
for line in lines:
 m=re.match(r'\s*(\w+)\s*(?:=|EQU\s+)([^;]+)',line,re.I)
 if m:
  try: const[m[1].upper()]=val(m[2])
  except (KeyError,ValueError,SyntaxError):pass
# EMPTYJINGLE is referenced by the matrix stream but declared in a lost include.
# Recover it from the table's own S_Empty jingle record.
assert schema.seed_empty_jingle(const,lines),'EMPTYJINGLE'
# Metadata for effects already independently validated by PF4.
effects={}
for i,l in enumerate(lines):
 m=re.match(r'^(\w+)\s+DW\s+(\w+)\s*(?:;.*)?$',l,re.I)
 if not m:continue
 rows=[r.split(';')[0].strip() for r in lines[i+1:i+6]]
 priority=0
 if re.fullmatch(r'DB\s+\d+',rows[0],re.I):priority=int(rows.pop(0).split()[1])
 if len(rows)>2 and all(len(re.findall(r'\d+',r))==12 for r in rows[:2]):
  ts=re.match(r'DW\s+(\w+)',rows[2],re.I)
  if ts:effects[m[1].upper()]={'jingle':m[2].upper(),'priority':priority,'matrix':ts[1].upper()}
jingles={}
for l in lines:
 m=re.match(r'^(S_\w+|SJINGLE\w+)\s+DB\s+(\d+),(\d+),(\d+)',l,re.I)
 if m:jingles[m[1].upper()]={'position':int(m[2]),'repeat':int(m[3]),'priority':int(m[4])}
# Original module control data: count tracker ticks up to each Bxx cue.
# All Fxx in this module select speed (<32), no BPM changes or E6/EE loops.
b=Path('TABLE1.MOD').read_bytes(); assert b[1080:1084]==b'M.K.'
orders=list(b[952:952+b[950]]); cues={}
for pos in range(len(orders)):
 speed=6; ticks=0; order=pos; row=0;segments=[{'position':pos,'start':0}]
 for _ in range(8192):
  p=orders[order]; jump=None; br=None
  for c in range(4):
   q=b[1084+p*1024+row*16+c*4:1084+p*1024+row*16+c*4+4]; e=q[2]&15; v=q[3]
   if e==15:
    assert 0<v<32
    speed=v
   if e==11:jump=v
   if e==13:br=(v>>4)*10+(v&15)
   assert not(e==14 and v>>4 in (6,14))
  ticks+=speed
  if jump is not None:
   cues[str(pos)]={'tracker_ticks':ticks,'next':jump,'syncs':(ticks*71+49)//50,'segments':segments};break
  if br is not None:
   order=(order+1)%len(orders);row=br;segments.append({'position':order,'start':ticks})
  else:
   row+=1
   if row==64:
    row=0;order=(order+1)%len(orders);segments.append({'position':order,'start':ticks})
 else:raise ValueError(('no cue',pos))
# Matrix data directives. Keep commands and label offsets, including branch targets.
commands=[]; labels={}
for l in lines[1280:2007]:
 l=l.split(';')[0].strip().upper()
 m=re.match(r'(\w+)\s+(?:LABEL\s+\w+|DW\s+)',l)
 if m:labels[m[1]]=len(commands)
 m=re.search(r'\bDW\s+(.+)',l)
 if m:
  args=[a.strip().replace('OFFSET ','') for a in m[1].split(',')]
  if args[0].startswith('_') or args[0]=='0':
   cmd={'op':args[0],'args':args[1:]}
   commands.append(cmd)
 elif re.fullmatch(r'CLEARIT[234]?',l):commands.append({'op':'_CLEAR'+(l[-1] if l[-1].isdigit() else '4'),'args':[]})
# Resolve every numeric operand from the original assembler expression.  The
# raw text stays in args for diagnostics; the runtime and the trace oracle
# consume nums (ticks is the retained _WAIT convenience alias).
for c in commands:
 try:n=schema.numeric_values(c['op'],c['args'],const)
 except schema.OperandError as e:raise SystemExit(f'PLAND.ASM: {e}')
 if n:
  c['nums']={str(i):v for i,v in n.items()}
  if c['op']=='_WAIT':c['ticks']=n[0]
# Animation timings: source headers precede frame pairs, all frame durations explicit.
anims={}; speed=4; headers=[]; current=None
for l in lines[2155:2495]:
 l=l.split(';')[0].strip().upper()
 m=re.match(r'SPEED\s*=\s*(\d+)',l)
 if m:speed=int(m[1]);continue
 const['SPEED']=speed
 m=re.match(r'(?:(\w+)\s+)?DW\s+(.+)',l)
 if not m:continue
 a=[x.strip() for x in m[2].split(',')]
 if len(a)==1:
  try:headers.append(val(a[0]))
  except (KeyError,ValueError,SyntaxError):continue
  current=None
 elif len(a)==2:
  if m[1]:
   current=m[1];anims[current]={'header':headers[-3:],'frames':[]};headers=[]
  if current:
   try:anims[current]['frames'].append(val(a[1]))
   except (KeyError,ValueError,SyntaxError):pass
# SPEED is forward referenced for _CYCLONE in this revision; supplied linked
# Party Land data fixes it to4. Cross-check every frame timing sequence against
# the linked DATA2 table (do not copy any graphics).
linked=Path('TABLE1.PRG').read_bytes()
import struct
for name,a in anims.items():
 if len(a['header'])==2:a['header'].insert(0,0)
 found=False
 for off in range(0x20750,0x21000,2):
  if struct.unpack_from('<HH',linked,off+2)!=tuple(a['header'][1:]):continue
  ds=[struct.unpack_from('<H',linked,off+8+i*4)[0] for i in range(len(a['frames']))]
  if ds==a['frames']:found=True;break
 assert found,('animation timing differs from linked data',name)
# Scroll length: DB content only; no byte contents copied into generated code.
scrolls={}; current=None
for l in lines[2005:2155]:
 l=l.split(';')[0].strip()
 m=re.match(r'(?:(\w+)\s+)?DB\s+(.+)',l,re.I)
 if not m:continue
 if m[1]:current=m[1].upper();scrolls[current]=0
 if current:
  # Each comma-separated element is one assembler byte (or one string).
  scrolls[current]+=schema.db_byte_count(m[2],const)
# Some scrolls occur before text block (e.g. BEATENSCROLL).
scrolls['BEATENSCROLL']=21+len('YOU HAVE BEATEN THE HIGHSCORE')+21+1
for a in anims.values():
 if len(a['header'])==2:a['header'].insert(0,0)
out={'effects':effects,'jingles':jingles,'cues':cues,'commands':commands,'labels':labels,'animations':anims,'scrolls':scrolls,'module_sha256':hashlib.sha256(b).hexdigest()}
# Independent command traces from FANTASIE's SISA/one-dispatch cadence.
# Direct rule entry at tick zero; tasks are scanned before each matrix call.
def animation_calls(a):
 index=0;loops=a['header'][1];calls=1
 for _ in range(65536):
  old=index
  if old==a['header'][2]:
   loops-=1
   if loops==0:return calls
   index=a['header'][0]
  index+=4;calls+=a['frames'][old//4]
 raise ValueError('animation never completes')
def matrix_trace(name):
 now=0;index=labels[name];trace=[];phase=8
 for _ in range(1000):
  c=commands[index];index+=1;op=c['op'];trace.append({'tick':now,'op':op})
  if op=='0' or op=='_COUNTDOWN':break
  if op=='_WAIT':d=c['ticks']
  elif op=='_CLEAR4':d=5
  elif op=='_ANIMATION':d=animation_calls(anims[c['args'][0]])
  elif op=='_SCROLL':
   left=scrolls[c['args'][0]]-21;d=0
   while True:
    d+=1
    for _ in range(2):
     if left==0:break
     phase-=1
     if phase==0:phase=8;left-=1
    else:continue
    break
  elif op=='_WAITJINGLE2':
   # Mystery B30 cue can complete before animation ends; polls on next call.
   complete=cues['30']['syncs'];d=max(1,complete-now)
  else:d=1
  now+=d
 return trace
traces={name:matrix_trace(name) for name in ['HIDDENTS','MYSTERYTS','HAPPYHOURTS','MEGALAUGHTS','CRAZYLETTERTS'] if name!='CRAZYLETTERTS'}
out['traces']=traces
# Restrict native content to programs reachable by Party Land gameplay, no menu,
# attract/high-score programs or tasks for neighboring tables.
roots={v['matrix'] for v in effects.values() if v['matrix']!='0'}|{'TILTTS','BEATENTS','BEATEN_BH_TS','PARTY_ONTS','PARTY_OFFTS','SHOOT_AGAIN_ONTS','OUT_OF_BALLSTS','JACKPOTTS','JACKPOT_SPECIAL_HH_TS','JACKPOT_SPECIAL_ML_TS'}
keep=set();queue=[labels[n] for n in roots]
while queue:
 n=queue.pop()
 if n in keep:continue
 keep.add(n);c=commands[n];op=c['op']
 if op in ['_JMP','_JBONUSX1']:queue.append(labels[c['args'][0]])
 if op=='_JBCDZ':queue.append(labels[c['args'][1]])
 if op not in ['0','_JMP','_CHECK_XXBALLS']:queue.append(n+1)
ordered=sorted(keep);remap={n:i for i,n in enumerate(ordered)}
out['commands']=[commands[n] for n in ordered]
out['labels']={k:remap[n] for k,n in labels.items() if n in remap}
used_anims={c['args'][0] for c in out['commands'] if c['op']=='_ANIMATION'}
out['animations']={k:v for k,v in anims.items() if k in used_anims}
used_scrolls={c['args'][0] for c in out['commands'] if c['op']=='_SCROLL'}
out['scrolls']={k:v for k,v in scrolls.items() if k in used_scrolls}
used_jingles={v['jingle'] for v in effects.values()}|{c['args'][0] for c in out['commands'] if c['op']=='_JINGLE'}|{'S_EMPTY','S_TILT','S_DANGER','SJINGLE4','S_ENDFIG','S_KNACKET','S_MAIN','S_SPRING','S_GAMEOVER','S_GAMEOVER2'}
out['jingles']={k:v for k,v in jingles.items() if k in used_jingles}
# Preserve NEXT_A termination for commands preceding an original zero word.
# Native uses explicit zero entries, rather than interpreting machine addresses.
# Hand-derived control-flow checkpoints, separate from native execution:
# ball_lostTS; bonus1000, multiplier2, cyclones2, no mode totals, no extra ball.
out['ball_trace']=[
 {'tick':0,'kind':'BallLost','label':'LOOSE_BALL'},
 {'tick':160,'kind':'BonusMultiplied','label':'_BONUS_X_CALCS'},
 {'tick':229,'kind':'BonusAdded','label':'_CALC_CYCLO'},
 {'tick':432,'kind':'ScoreAwarded','label':'DO_FLORPA'},
 {'tick':436,'kind':'ScoreAwarded','label':'DO_FLORPA'},
 {'tick':440,'kind':'ScoreAwarded','label':'DO_FLORPA'},
 {'tick':444,'kind':'ScoreAwarded','label':'DO_FLORPA'},
 {'tick':520,'kind':'NewBall','label':'NEW_BALL'},
 {'tick':525,'kind':'Sound','label':'SBRICKUPP'},
 {'tick':571,'kind':'Sound','label':'SNEWBALL'},
 {'tick':600,'kind':'TaskReady','label':'SETBALL'},
]
Path('.private-cleanup/generated').mkdir(parents=True,exist_ok=True)
Path(os.environ.get('PF_REFERENCE_OUTPUT', '.private-cleanup/generated/pf45-timing-fixtures.json')).write_text(json.dumps(out,indent=2)+'\n')
if not os.environ.get('PF_REFERENCE_NO_NATIVE'):
 Path('internal/partyland/timing_data.go').write_text('// Code generated by tools/reference_pf45.py from original timing data; DO NOT EDIT.\npackage partyland\n\nconst originalTimingData = `'+json.dumps(prg_content_layout.locate_table_records(1,out),separators=(',',':'))+'`\n')
print('Retained',len(out['commands']),'Party Land matrix commands and',len(traces),'independent command traces')
