#!/usr/bin/env python3
"""Extract original matrix content locations and declarations. No CPU execution.
Static graphics-store records are treated as artwork, as in the PF2 ball reader.
"""
import re,json,struct,hashlib,subprocess
from pathlib import Path
import matrix_operand_schema as schema
allout={}
for table,source,dsbase,font13,animbase,animbound in [(1,'PLAND',0x19d40,0x1ff40,0x20750,0x30000),(2,'SDEV',0x18ee0,0x1f170,0x1f8d0,0x2d000),(3,'SHOW',0x18b60,0x1e640,0x41ac0,0x4cb60),(4,'STONES',0x166d0,0x1cec0,0x1daf0,0x289b0)]:
 b=Path(f'TABLE{table}.PRG').read_bytes()
 s=Path(f'reference/original-dos-source/{source}.ASM').read_bytes().decode('latin1').upper().splitlines()
 # Only retail data; commented examples are excluded.
 filtered=[];comment=False;active=True;stack=[]
 for l in s:
  t=l.strip()
  if t.startswith('COMMENT\\'):comment=True;filtered.append('');continue
  if comment:
   if t=='\\':comment=False
   filtered.append('');continue
  if t=='IF DEMOVER':stack.append(active);active=False;filtered.append('');continue
  if t=='ELSE' and stack:active=stack[-1];filtered.append('');continue
  if t=='ENDIF' and stack:active=stack.pop();filtered.append('');continue
  filtered.append(l if active else '')
 s=filtered
 constants={'SW':336,'TOTCENT':16,'SPEED':4}
 def val(v):
  return schema.resolve_expression(v.strip().upper().replace("OFFSET ", ""),constants)
 for l in s:
  m=re.match(r'\s*(\w+)\s*(?:=|EQU\s+)([^;]+)',l)
  if m:
   try:constants[m[1]]=val(m[2])
   except (KeyError,ValueError,SyntaxError,ZeroDivisionError):pass
 # EMPTYJINGLE comes from a lost include in F1-F3; every table declares the
 # empty jingle record itself, whose position is the recovered equate value.
 assert schema.seed_empty_jingle(constants,s),('EMPTYJINGLE',source)
 # DB text programs, retaining original encoded digits/punctuation and blanks.
 # Each element is one assembler byte expression; character literals combine
 # with arithmetic ('7'+2, 6+'7', '0'+7) exactly as the linker evaluated them.
 texts={};cur=None
 for l in s:
  l=l.split(';')[0].strip();m=re.match(r'(?:(\w+)\s+)?DB\s+(.+)',l)
  if not m:cur=None;continue
  if m[1]:cur=m[1];texts[cur]=[]
  if not cur:continue
  try:texts[cur]+=schema.db_bytes(m[2],constants)
  except (schema.OperandError,KeyError,ValueError,SyntaxError):cur=None
 texts={k:v for k,v in texts.items() if all(0<=x<=255 for x in v)}
 # Matrix commands (native data identifiers, never machine code pointers).
 commands=[];labels={};a,z={1:(1280,2007),2:(1187,1935),3:(1030,1766),4:(1286,2259)}[table]
 for l in s[a:z]:
  l=l.split(';')[0].strip();m=re.match(r'(\w+)\s+(?:LABEL\s+\w+|DW\s+)',l)
  if m:labels[m[1]]=len(commands)
  m=re.search(r'\bDW\s+(.+)',l)
  if m:
   args=[x.strip().replace('OFFSET ','') for x in m[1].split(',')]
   if args[0].startswith('_') or args[0]=='0':commands.append({'op':args[0],'args':args[1:]})
  elif table in (3,4) and l=='CLEARIT':commands.append({'op':'_ANIMATION','args':['_CLEAR']})
  elif re.fullmatch(r'CLEARIT[234]?',l):commands.append({'op':'_CLEAR'+(l[-1] if l[-1].isdigit() else ('4' if table==1 else '1')),'args':[]})
 if table in (3,4):
  for c in commands:
   if c['op']=='_PRINT13_NUMBER' and not c['args']:c['args']=['SPINSCOREPTR','SW*4/4+16*2/4']
 # Retail numeric operands are resolved from the original assembler expression
 # here.  The generated command keeps the raw source text in args for
 # diagnostics and identity lookups and carries the typed integer in nums; the
 # native runtime consumes nums and never evaluates a source expression.
 def attach_numbers(cs):
  for c in cs:
   try:n=schema.numeric_values(c['op'],c['args'],constants)
   except schema.OperandError as e:raise SystemExit(f'{source}.ASM: {e}')
   if n:c['nums']={str(i):v for i,v in n.items()}
 attach_numbers(commands)
 # Expand the reachable frontend ShowHighsTS macro as source data.
 macro=[];inside=False
 for l in s:
  t=l.split(';')[0].strip()
  if t.startswith('SHOWITHI') and 'MACRO' in t:inside=True;continue
  if inside and t=='ENDM':inside=False;continue
  if inside:macro.append(t)
 attract=[];inside=False
 for l in s:
  t=l.split(';')[0].strip()
  if t.startswith('SHOWHIGHSTS') and 'LABEL' in t:inside=True;continue
  if inside and t=='DW 0':break
  if inside:
   for line in (macro if t=='SHOWITHI' else [t]):
    m=re.match(r'DW\s+(.+)',line)
    if m:
     args=[x.strip().replace('OFFSET ','') for x in m[1].split(',')]
     if args[0]=='_JMP':continue
     if args[0].startswith('_'):
      attract.append({'op':args[0],'args':args[1:]})
    elif table in (3,4) and line=='CLEARIT':attract.append({'op':'_ANIMATION','args':['_CLEAR']})
    elif re.fullmatch(r'CLEARIT[234]?',line):attract.append({'op':'_CLEAR'+(line[-1] if line[-1].isdigit() else ('4' if table==1 else '1')),'args':[]})
 attach_numbers(attract)
 # Locate animation tables by headers and ALL durations, in declaration order.
 animations={};headers=[];cur=None;names={};a,z={1:(2155,2495),2:(1940,2210),3:(4197,4456),4:(2261,2479)}[table]
 for l in s[a:z]:
  l=l.split(';')[0].strip();m=re.match(r'(\w+)\s*=\s*(.+)',l)
  if m:
   try:constants[m[1]]=val(m[2])
   except (KeyError,ValueError,SyntaxError):pass
   continue
  m=re.match(r'(?:(\w+)\s+)?DW\s+(.+)',l)
  if not m:
   label=re.match(r'(\w+)\s+LABEL\s+WORD',l)
   if label: pending=label[1]
   continue
  if table in (3,4) and not m[1] and 'pending' in locals() and pending:
   l=pending+' '+l;pending=None;m=re.match(r'(?:(\w+)\s+)?DW\s+(.+)',l)
  args=[x.strip() for x in m[2].split(',')]
  if table in (3,4) and len(args)==2 and not m[1] and not args[0].startswith(('A_','CLEAR')):
   try:headers.extend(val(x) for x in args);cur=None;continue
   except (KeyError,ValueError,SyntaxError):pass
  if len(args)==1:
   try:headers.append(val(args[0]))
   except (KeyError,ValueError,SyntaxError):continue
   cur=None
  elif len(args)==2:
   if m[1]:
    cur=m[1];h=headers[-3:] if len(headers)>=3 else [0]+headers[-2:];animations[cur]={'header':h,'durations':[],'offsets':[]};names[cur]=[];headers=[]
   if cur:
    try:animations[cur]['durations'].append(val(args[1]));names[cur].append(args[0])
    except (KeyError,ValueError,SyntaxError):pass
 cursor=animbase
 for name,a in animations.items():
  candidates=[];n=len(a['durations']);h=a['header']
  for o in range(cursor,min(animbound,len(b))-4*n):
   if b[o:o+4]!=struct.pack('<2H',*h[1:]):continue
   if not all(struct.unpack_from('<H',b,o+6+4*i)[0]==d for i,d in enumerate(a['durations'])):continue
   pointers=[struct.unpack_from('<H',b,o+4+4*i)[0] for i in range(n)]
   # Matrix animation data lives in same DATA2 segment as tables.
   candidates.append((o,pointers))
  assert candidates,(table,name,h)
  o,pointers=candidates[0];cursor=o+4+4*n
  # DATA2 base from static ANIMATION_ROUTINE segment immediate. Table 1=2055.
  base={1:0x20750,2:0x1f8e0,3:0x41ac0,4:0x1daf0}[table]
  a['offsets']=[base+p for p in pointers]
  for off in a['offsets']:
   q=off
   for plane in range(2):
    count=struct.unpack_from('<H',b,q)[0];q+=2;assert count<3000,(table,name,hex(off),count);q+=count
 # Spring segment is recovered from PUTSPRINGINGFX, byte array 10x23.
 p=b.index(bytes.fromhex('be0000bfafd4'));spring=struct.unpack_from('<H',b,p-9)[0]*16+512
 # Exact command positions using assembly SW=336 (memory stride, not display width).
 positions={}
 for c in commands+attract:
  # Packed Tower row selectors retain their original ASM operand identity.
  if c['op']=='_TOWER':positions[c['args'][0]]=val(c['args'][0])
  if (c['op'].startswith('_PRINT') or c['op'].startswith('_RULLGARDIN')) and len(c['args'])>=2:
   try:positions[c['args'][1]]=val(c['args'][1])
   except (KeyError,ValueError,SyntaxError):pass
 lampflash=[]
 for l in s[835:902] if table==1 else (s[705:752] if table==2 else (s[691:724] if table==3 else s[883:920])):
  t=l.split(';')[0].strip();m=re.match(r'(\w+)\s*=\s*(.+)',t)
  if m:constants[m[1]]=val(m[2]);continue
  m=re.match(r'DW\s+(.+)',t)
  if m:
   vs=[val(x) for x in m[1].split(',')]
   if len(vs)==5:lampflash.append(vs[1:])
 # Font artwork is decoded from user-supplied PRGs at runtime, never exported.
 allout[str(table)]={'texts':texts,'animations':animations,'commands':commands,'labels':labels,'positions':positions,'fonts':{'13':font13,'11':font13+0x210,'8':font13+0x3d0,'5':font13+0x510},'spring':spring,'sha256':hashlib.sha256(b).hexdigest(),'attract':attract,'lampflash':lampflash}
import prg_content_layout
allout=prg_content_layout.locate(allout)
Path('internal/presentation/content.go').write_text('// Generated by tools/reference_pf8.py: source DATA declarations/content locations.\npackage presentation\nconst contentJSON = `'+json.dumps(allout,separators=(',',':'))+'`\n')
# Full original-backed dumps belong only in the ignored private archive.
private=Path('.private-cleanup/generated');private.mkdir(parents=True,exist_ok=True)
(private/'pf8-content.json').write_text(json.dumps(prg_content_layout.hydrate(allout,Path('.')),indent=2)+'\n')
print('Extracted original matrix content for all four native tables')
