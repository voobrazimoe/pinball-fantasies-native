#!/usr/bin/env python3
"""PF4 oracle: parse original PLAND decimal constants, effects and lamp data.
No Go import, DOS execution, assembler translation, or copyrighted asset copies.
State expectations below are separately derived from named PLAND routines.
"""
import hashlib,json,re
from pathlib import Path
import os
src=Path('reference/original-dos-source/PLAND.ASM').read_bytes().decode('latin1')
data=Path('TABLE1.PRG').read_bytes()
constants={}
effects={}
lines=src.splitlines()
for i,line in enumerate(lines):
 m=re.match(r'^(\w+)\s+DB\s+([0-9, ]+)',line,re.I)
 if m:
  values=[int(x) for x in m[2].replace(' ','').split(',') if x]
  if len(values)==12:constants[m[1].upper()]=int(''.join(map(str,values)))
 m=re.match(r'^(\w+)\s+DW\s+(\w+)\s*(?:;.*)?$',line,re.I)
 if m:
  following=lines[i+1:i+4]
  if following and re.match(r'\s+DB\s+\d+\s*(?:;.*)?$',following[0],re.I):following=following[1:]+lines[i+4:i+5]
  values=[]
  for row in following[:2]:
   v=re.findall(r'\d+',row.split(';')[0])
   if len(v)==12:values.append(int(''.join(v)))
  if len(values)==2:effects[m[1].upper()]=dict(score=values[0],bonus=values[1],audio=m[2])
# Sequential external palette records cross-checked against source's literal LON1.
assert data[0x1adb9:0x1adc1]==bytes([128,2,95,0,0,64,0,0])
p=0x1adb9;lamps=[]
for n in range(1,57):
 start,count=data[p:p+2];p+=2
 rgb=data[p:p+count*3];p+=count*3
 lamps.append(dict(number=n,start=start,count=count,sha256=hashlib.sha256(rgb).hexdigest()))
fixtures={
 'bumper':1000,'kicker':500,'duck_bank_score':3*constants['BCD7510'],
 'duck_bank_bonus':3*constants['BCD750'],
 'puke_first_score':4*constants['BCD20070']+effects['PARTYSCORE5']['score'],
 'puke_first_bonus':4*constants['BCD1000']+effects['PARTYSCORE5']['bonus'],
 'skyride_bank_score':sum(effects['RIDE'+str(n)]['score'] for n in range(1,4)),
 'skyride_bank_bonus':sum(effects['RIDE'+str(n)]['bonus'] for n in range(1,4)),
 # three skyride awards -> first reverse + M2 -> ADDBONUS lane1000*2
 'multiplier_path_score':850000+effects['PARTYSCORE4']['score']+effects['RSCORE1']['score']+effects['M2']['score']+effects['BYGELSETB']['score'],
 'multiplier_path_bonus':85000+effects['PARTYSCORE4']['bonus']+effects['RSCORE1']['bonus']+effects['M2']['bonus']+effects['BYGELSETB']['bonus'],
 'bonus_before_cyclones':2000,'bonus_after_cyclones':202000,
}
out=dict(constants=constants,effects=effects,lamps=lamps,fixtures=fixtures,
 content={k:dict(offset=o,length=l,sha256=hashlib.sha256(data[o:o+l]).hexdigest()) for k,o,l in [('font13',0x1ff40,36*13),('font5',0x20450,36*5),('duck1up',0x205f0,30),('duck1down',0x20610,30),('duck2up',0x20630,30),('duck2down',0x20650,30),('duck3up',0x20680,15),('duck3down',0x20670,15)]})
Path('.private-cleanup/generated').mkdir(parents=True,exist_ok=True)
Path(os.environ.get('PF_REFERENCE_OUTPUT', '.private-cleanup/generated/pf4-rule-fixtures.json')).write_text(json.dumps(out,indent=2)+'\n')
print('Wrote source-derived PF4 fixtures:',len(effects),'effects,',len(constants),'constants')
