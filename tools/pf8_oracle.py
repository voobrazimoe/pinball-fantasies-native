#!/usr/bin/env python3
"""Prepare an isolated original-game oracle. Never imported by the native game.
Usage: python3 tools/pf8_oracle.py .tools/pf8-oracle
Launch the emitted config manually with dosbox-x -defaultconf -conf PATH.
"""
import hashlib,json,shutil,sys
from pathlib import Path
root=Path(__file__).resolve().parent.parent
out=Path(sys.argv[1]).resolve()
if out==root or root in out.parents and not out.is_relative_to(root/'.tools'):
 raise SystemExit('Use an isolated directory under .tools or outside the repository.')
if out.exists() and any(out.iterdir()):raise SystemExit('Destination must be empty: preserve existing reference captures/settings.')
out.mkdir(parents=True,exist_ok=True)
for f in json.loads((root/'analysis/game-inventory.json').read_text()):
 name=f['name'];p=root/name
 if not name.endswith('.HI') and hashlib.sha256(p.read_bytes()).hexdigest()!=f['sha256']:
  raise SystemExit('Original file differs from inventory: '+name)
 if name.endswith('.HI'):
  b=p.read_bytes()
  # INTRO.ASM MOVSCORE decrements CX after the first nonzero digit and
  # enters LUDDE unconditionally. Values 0..9 underflow CX and corrupt DS.
  if len(b)!=64 or any(not any(b[i:i+11]) for i in range(0,64,16)):
   raise SystemExit(name+': original INTRO cannot safely format a score below 10; do not modify the user file.')
 shutil.copy2(p,out/name)
conf=out.parent/(out.name+'.conf')
conf.write_text('[sdl]\noutput=surface\nshowmenu=false\n[render]\naspect=false\nscaler=none\n[cpu]\ncycles=30000\n[mixer]\nnosound=false\n[autoexec]\nmount c '+str(out)+'\nc:\npinball.exe\n')
print(conf)
