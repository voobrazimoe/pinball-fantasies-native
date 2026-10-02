#!/usr/bin/env python3
"""PF11 fixture: settings/menu symbols and SETSCREENSTART integer checkpoints.
Source files are inputs only. The native implementation never imports this tool.
"""
from pathlib import Path
import hashlib,json,re
root=Path(__file__).resolve().parent.parent
sources={name:(root/'reference/original-dos-source'/name).read_bytes() for name in ['INTRO.ASM','FANTASIE.ASM','BALLCODE.ASM','PLAND.ASM','SDEV.ASM','SHOW.ASM','STONES.ASM']}
text={n:b.decode('latin1') for n,b in sources.items()}
intro=text['INTRO.ASM'];fant=text['FANTASIE.ASM']
assert re.search(r'S_BALLS\s+DB.*S_ANGLE\s+DB.*S_SCROLLING\s+DB.*S_IM\s+DB.*S_RESOLUTION\s+DB.*S_MODE\s+DB',intro,re.S)
assert 'SUB' in fant[fant.index('TABLE_ANGLE Proc'):fant.index('TABLE_ANGLE EndP')].upper()
rows=[]
for height in [240,350]:
 for factor in [20,11,9]:
  values=[]
  for y in [0,100,200,450,570]:
   middle=(height-33)//2;target=max(0,min(576-(height-33),y-(middle-28)))+33
   raster=1600+((target-100)*factor>>2)
   gap=target-(raster>>4)
   if gap<0:
    gap+=middle-28
    if gap<=0:raster+=gap<<4
   else:
    gap-=middle+12
    if gap>=0:raster+=gap<<4
   values.append(raster)
  rows.append(dict(height=height,factor=factor,start_raster=1600,ball_y=[0,100,200,450,570],raster=values))
out=dict(source_sha256={n:hashlib.sha256(b).hexdigest() for n,b in sources.items()},record_order=['S_BALLS','S_ANGLE','S_SCROLLING','S_IM','S_RESOLUTION','S_MODE'],defaults=[0,0,1,0,0,0],installation=list((root/'PINBALL.CFG').read_bytes()),choices=[['3','5'],['HIGH','LOW'],['HARD','MEDIUM','SOFT'],['ON','OFF'],['NORMAL','HIGH'],['COLOR','MONO']],menu=dict(heading='OPTIONS MENU',rows=['BALLS:','ANGLE:','SCROLLING:','INGAME MUSIC:','RESOLUTION:','COLOR MODE:','SAVE AND EXIT'],row_y=[50,68,86,104,122,140,176],arrow_x=175,controls=dict(up=72,down=80,toggle=[28,57],escape=1),help=['  USE THE   ','CURSOR KEYS ','TO SELECT AN','   OPTION   ','            ','TOGGLE WITH ','  ENTER OR  ','  SPACEBAR  ','            ','ESCAPE QUITS']),scroll=rows,video=[dict(logical=[320,240],field=207,matrix=33,line_compare=413),dict(logical=[320,350],field=317,matrix=33,line_compare=316)])
(root/'analysis/pf11-settings-fixtures.json').write_text(json.dumps(out,indent=2)+'\n')
print('PF11 source settings fixture written')
