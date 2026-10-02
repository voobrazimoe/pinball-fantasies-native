#!/usr/bin/env python3
"""Decode original DATA and DAC transformations independently of native Go."""
import json,hashlib,struct
from pathlib import Path
import prg_content_layout
h=lambda b:hashlib.sha256(bytes(b)).hexdigest()
content=json.loads(Path('.private-cleanup/generated/pf8-content.json').read_text());out={}
for n,c in content.items():
 b=Path('TABLE'+n+'.PRG').read_bytes();frames={}
 for name,a in c['animations'].items():
  dots=bytearray(2560);hashes=[]
  for off in a['offsets']:
   q=off
   for plane in [0,1]:
    count=struct.unpack_from('<H',b,q)[0];q+=2;dest=167
    for v in b[q:q+count]:
     dest+=v//2
     if v==254:continue
     row=dest//168-1;col=dest%84*2+plane
     if 0<=row<16 and col<160:dots[row*160+col]=v%2
    q+=count
   hashes.append(h(dots))
  frames[name]=hashes
 out[n]={'animations':frames,'spring_sha256':h(b[c['spring']:c['spring']+230]),'scorefont_sha256':h(sum([prg_content_layout.score_glyphs(int(n),b)[str(i)] for i in range(48,58)],[]))}
# Griffin upper/lower share the first picture's palette, as in pelle1.
b=Path('INTRO.PRG').read_bytes();o=0x3b810;cm=b.index(b'CMAP',o);pal=bytearray(b[cm+8:cm+8+768]);pal[96:192]=bytes(v>>1 for v in pal[:96]);pal=bytes(((v>>2)<<2)|((v>>2)>>4) for v in pal)
out['griffin']={'palette_sha256':h(pal),'colors':{str(i):list(pal[i*3:i*3+3]) for i in [0,1,5,15,31,32,34,45,63]}}
# Read both original PBM planes directly for the authentic held startup frame.
def pbm(offset):
 q=offset+12;end=offset+8+int.from_bytes(b[offset+4:offset+8],'big');chunks={}
 while q<end:
  tag=b[q:q+4];n=int.from_bytes(b[q+4:q+8],'big');chunks[tag]=b[q+8:q+8+n];q+=8+n+(n&1)
 w,height=struct.unpack_from('>HH',chunks[b'BMHD']);q=0;pixels=bytearray();body=chunks[b'BODY']
 for y in range(height):
  row=bytearray()
  while len(row)<w:
   code=struct.unpack_from('b',body,q)[0];q+=1
   if code>=0:row.extend(body[q:q+code+1]);q+=code+1
   elif code!=-128:row.extend(bytes([body[q]])*(1-code));q+=1
  pixels.extend(row)
 return w,height,pixels
_,_,upper=pbm(0x3b810);_,lowerh,lower=pbm(0x42c50)
indices=upper[:320*240];indices[139*320:240*320]=lower[:101*320]
# fade macro loops CL=20..1; the retained palette is new*19/20, in DAC units.
held=bytes((((v>>2)*19//20)<<2)|(((v>>2)*19//20)>>4) for v in pal)
rgba=b''.join(held[i*3:i*3+3]+bytes([255]) for i in indices)
out['griffin']['hold_rgba_sha256']=h(rgba)
Path('analysis/pf8-visual-fixtures.json').write_text(json.dumps(out,indent=2)+'\n')
print('Source-derived matrix frames, spring, glyph and griffin palette fixtures written')
