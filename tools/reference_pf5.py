#!/usr/bin/env python3
"""Independent content/sample oracle. Never executes DOS code."""
from pathlib import Path
import hashlib,struct,json
b=Path('TABLE1.MOD').read_bytes()
sha=lambda x:hashlib.sha256(x).hexdigest()
orders=list(b[952:952+b[950]]);count=max(orders)+1
effects={};off=1084+1024*count;samples=[]
for i in range(31):
 h=b[20+30*i:50+30*i];n=struct.unpack_from('>H',h,22)[0]*2
 samples.append(dict(number=i+1,name=h[:22].rstrip(b'\0').decode('ascii'),length=n,fine=h[24],volume=h[25],loop_start=struct.unpack_from('>H',h,26)[0]*2,loop_length=struct.unpack_from('>H',h,28)[0]*2,offset=off,sha256=sha(b[off:off+n])))
 off+=n
for i in range(1084,1084+1024*count,4):
 q=b[i:i+4];e=q[2]&15;effects.setdefault(str(e),set()).add(q[3])
# SBUMPER1 uses sample24 / note25 (period202), voice3 (centered runtime effect), volume64.
# Independently evaluate fixed-point linear interpolation of 900 frames within the first tick.
s=samples[23];pcm=b[s['offset']:s['offset']+s['length']];step=int(3546895/202*2**32/48000)
render=bytearray()
for frame in range(900):
 phase=frame*step;index=phase>>32;v=pcm[index];v=v-256 if v>=128 else v
 adjacent=pcm[index+1];adjacent=adjacent-256 if adjacent>=128 else adjacent
 render+=struct.pack('<hh',v*64+((adjacent-v)*(phase&0xffffffff)*64>>32),v*64+((adjacent-v)*(phase&0xffffffff)*64>>32))
v=dict(module_sha256=sha(b),orders=orders,patterns=count,effects={k:sorted(v) for k,v in effects.items()},samples=samples,bumper900_sha256=sha(render))
Path('analysis/pf5-audio-fixtures.json').write_text(json.dumps(v,indent=2)+'\n')
print('PF5 content fixtures:',sha(b),'samples',len(samples),'effects',list(effects))
