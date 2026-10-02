#!/usr/bin/env python3
"""Measure named stable logical comparisons; never infer timing from host scaling."""
from PIL import Image
from pathlib import Path
import hashlib,json
root=Path(__file__).resolve().parent.parent
p=root/'analysis/pf8-checkpoints'
cases=[
 ('griffin','01-griffin','verified-x-griffin',(320,240),None),
 ('selector','02-selector','verified-x-selector-settled',(640,240),None),
 ('selector-instructions','02-selector-sidebar','dosbox-x-selector-next',(640,240),(8,95,120,185)),
 ('party-initial','04-table1-initial','dos-party-initial-logical',(320,350),None),
 ('speed-initial','09-table2-initial','dosbox-x-speed-initial-logical',(320,350),None),
 ('party-plunger-charged','table1-plunger-charged','verified-x-party-charged-final',(320,350),(304,297,314,314)),
 ('speed-plunger-charged','table2-plunger-charged','dosbox-x-speed-charged',(320,350),(304,297,314,314)),
 ('party-tilt','07-table1-tilt','verified-x-party-tilt-early',(320,350),(0,317,320,350)),
 ('speed-tilt','12-table2-tilt','verified-x-speed-tilt',(320,350),(0,317,320,350)),
 ('party-source-scroller','05-party-source-scroller','verified-x-party-scroll-long',(320,350),(0,317,320,350)),
 ('speed-attract','08-table2-attract','verified-x-speed-attract',(320,350),(0,317,320,350)),
 ('party-attract-name','table1-attract-name','fast-x-party-scroll',(320,350),(0,317,320,350)),
 ('high-score-entry','table2-high-score','verified-x-ending-gameover',(320,350),(0,317,320,350)),
]
results=[]
for name,native,oracle,size,region in cases:
 n=Image.open(p/(native+'.png')).convert('RGBA');d=Image.open(p/(oracle+'.png')).convert('RGBA')
 assert n.size==size,(name,n.size,size)
 # Verify exact scanline/sample repetition before discarding it.
 assert d.width in [size[0],size[0]*2] and d.height in [size[1],size[1]*2],(name,d.size)
 sx,sy=d.width//size[0],d.height//size[1]
 pix=d.load()
 for y in range(size[1]):
  for x in range(size[0]):
   v=pix[x*sx,y*sy]
   assert all(pix[x*sx+dx,y*sy+dy]==v for dy in range(sy) for dx in range(sx)),(name,'non-identical sample duplication',x,y)
 d=d.resize(size,Image.Resampling.NEAREST)
 if region:n=n.crop(region);d=d.crop(region)
 np,dp=n.load(),d.load();good=sum(np[x,y]==dp[x,y] for y in range(n.height) for x in range(n.width));total=n.width*n.height
 results.append({'checkpoint':name,'native':native+'.png','oracle':oracle+'.png','sample_divisor':[sx,sy],'logical_region':region,'matching_pixels':good,'total_pixels':total,'oracle_rgba_sha256':hashlib.sha256(d.tobytes()).hexdigest()})
 print(name,good,'/',total)
(root/'analysis/pf8-parity-measurements.json').write_text(json.dumps(results,indent=2)+'\n')
