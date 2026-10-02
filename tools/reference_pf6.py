#!/usr/bin/env python3
"""Independent IFF content oracle. PF8 uses six-bit DAC bit replication; Never reads/executes machine instructions."""
from pathlib import Path
import hashlib,json,struct
b=Path('INTRO.PRG').read_bytes()
assert hashlib.sha256(b).hexdigest()=='f6b5590f88949174b8f530d5b2d5b59f329d9fde24bfa98a0154d74f19f06881'
out=[];offset=0
while True:
 offset=b.find(b'FORM',offset)
 if offset<0:break
 start=offset;offset+=4
 size=struct.unpack_from('>I',b,start+4)[0];end=start+8+size;pos=start+12;chunks={}
 while pos<end:
  n=struct.unpack_from('>I',b,pos+4)[0];chunks[b[pos:pos+4]]=b[pos+8:pos+8+n];pos+=8+n+(n&1)
 assert pos==end
 h=chunks[b'BMHD'];w,height=struct.unpack_from('>HH',h);planes=h[8]
 chunky=b[start+8:start+12]==b'PBM '
 stride=(w+1)&~1 if chunky else (w+15)//16*2
 src=chunks[b'BODY'];q=0;rows=[]
 for _ in range(height if chunky else height*planes):
  row=bytearray()
  while len(row)<stride:
   n=struct.unpack_from('b',src,q)[0];q+=1
   if n>=0:row+=src[q:q+n+1];q+=n+1
   elif n!=-128:row+=bytes([src[q]])*(1-n);q+=1
  assert len(row)==stride
  rows.append(row)
 assert q==len(src)
 if chunky:indices=bytes(v for row in rows for v in row[:w])
 else:
  indices=bytes(sum(((rows[y*planes+p][x//8]>>(7-x%8))&1)<<p for p in range(planes)) for y in range(height) for x in range(w))
 palette=bytes(((v>>2)<<2)|((v>>2)>>4) for v in chunks[b'CMAP'])
 rgba=bytes(v for index in indices for v in (*palette[index*3:index*3+3],255))
 out.append(dict(offset=start,width=w,height=height,indices=hashlib.sha256(indices).hexdigest(),rgba=hashlib.sha256(rgba).hexdigest()))
Path('analysis/pf6-art-fixtures.json').write_text(json.dumps(out,indent=2)+'\n')
print('Validated',len(out),'original frontend IFF images')
