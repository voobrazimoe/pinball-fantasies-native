#!/usr/bin/env python3
"""Static expansion of supplied SP packets for the callback dependency cone.
This decodes literal/repeat DATA packets; no instruction execution or emulation.
Output is disposable disassembler input, never loaded by the native application.
"""
from pathlib import Path
import struct,hashlib,json,subprocess
result={}
for name in ['SBLASTER','SB16','NOSOUND']:
 original=Path(name+'.SDR').read_bytes();b=original[512:]
 cs=struct.unpack_from('<H',original,22)[0]*16
 header=struct.unpack_from('<8H',b,cs)
 # Verified stub: STD, scan backwards from paragraph immediately before stub;
 # destination immediately before relocated stub. See SBLASTER image2600+12.
 src=cs-header[7]*16+15;dest=header[6]*16-header[7]*16+15
 while b[src]==255:src-=1
 out=bytearray(dest+1);packets=0
 while True:
  marker=b[src];n=struct.unpack_from('<H',b,src-2)[0];src-=3
  assert n<=dest+1
  if marker&254==0xb0:out[dest-n+1:dest+1]=bytes([b[src]])*n;src-=1
  elif marker&254==0xb2:out[dest-n+1:dest+1]=b[src-n+1:src+1];src-=n
  else:raise ValueError((name,hex(src+3),hex(marker)))
  dest-=n;packets+=1
  if marker&1:break
 assert src==dest
 out[:dest+1]=b[:src+1]
 target=Path('/tmp/pf5-'+name+'-static.bin');target.write_bytes(out)
 result[name]=dict(original_sha256=hashlib.sha256(original).hexdigest(),static_image_sha256=hashlib.sha256(out).hexdigest(),packets=packets,expanded_bytes=len(out),unchanged_prefix=dest+1)
 if name=='SBLASTER':
  table=list(struct.unpack_from('<16H',out,0x1e20+0x4ca));assert table[11]==0x12e4
  result[name]['effect_dispatch']=table
  dis=[]
  for start,end in [(0x19e,0x21c),(0xcaf,0xd48),(0xd73,0xe2b),(0x109f,0x10de),(0x12e4,0x130a),(0x1cc9,0x1d2b)]:
   dis.append(subprocess.check_output(['objdump','-D','-b','binary','-m','i8086','-Mintel',f'--start-address={start}',f'--stop-address={end}',str(target)],text=True))
  Path('analysis/pf5-sdr-callback-disassembly.txt').write_text('\n'.join(dis))
Path('analysis/pf5-sdr-evidence.json').write_text(json.dumps(result,indent=2)+'\n')
print('Static callback evidence written; no DOS execution.')
