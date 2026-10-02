#!/usr/bin/env python3
"""Independent Party Land numeric reference, derived from BALLCODE/FANTASIE.
Reads original data; never calls Go, executes x86, or emulates DOS hardware.
Python integers explicitly narrowed at the original word/product boundaries.
"""
import struct, json, hashlib
from pathlib import Path
import os
from reference_extract import extract
ORIGINAL_RGBA=extract('TABLE1.PRG')[1]

D=Path('TABLE1.PRG').read_bytes()
assert hashlib.sha256(D).hexdigest()=='4d7a69e7dc95260ad2541c6981a11ab842e2f1f20e45447e5613688b86e38414'
CMAP=D[D.index(b'CMAP',437168)+8:D.index(b'CMAP',437168)+8+768]
BALL_OFFSETS=json.loads(Path('analysis/pf2-ball-locations.json').read_text())['offsets']
BALL=bytes(D[o] if o else 0 for o in BALL_OFFSETS)
def s16(n): return (n+32768)%65536-32768
def s32(n): return (n+2147483648)%4294967296-2147483648
def div(n,d):
    assert d
    q=abs(n)//abs(d)*(-1 if (n<0)!=(d<0) else 1)
    assert -32768<=q<=32767
    return q
def clamp(n): return max(-4100,min(4100,s16(n)))
def words(o,n): return struct.unpack_from('<'+'h'*n,D,o)
def region(o,n): return bytearray(D[o:o+n])
SIN=struct.unpack_from('>2560h',D,0x1e340)
MAT=[words(0x1c05d+16*i,5) for i in range(8)]
XY=[(16,8),(16,9),(16,10),(15,11),(15,12),(14,13),(13,14),(12,15),(11,15),(10,16),(9,16),(8,16),(7,16),(6,16),(5,15),(4,15),(3,14),(2,13),(1,12),(1,11),(0,10),(0,9),(0,8),(0,7),(0,6),(1,5),(1,4),(2,3),(3,2),(4,1),(5,1),(6,0),(7,0),(8,0),(9,0),(10,0),(11,1),(12,1),(13,2),(14,3),(15,4),(15,5),(16,6),(16,7)]
# Explicit checkpoint records: row, byte displacement, rotated bit, angle,
# quadrant, XY index. This oracle samples the original rotated 16-bit words;
# production Go samples direct ring pixels instead.
CP=[(0,1,6,1616,8,35),(0,1,7,1577,8,34),(0,0,0,1536,8,33),(0,0,1,1495,4,32),(0,0,2,1456,4,31),
(1,1,4,1705,8,37),(1,1,5,1668,8,36),(1,0,3,1404,4,30),(1,0,4,1367,4,29),
(2,1,3,1762,8,38),(2,0,5,1310,4,28),(3,1,2,1822,8,39),(3,0,6,1250,4,27),
(4,1,1,1879,8,40),(4,0,7,1193,4,26),(5,1,1,1916,8,41),(5,0,7,1156,4,25),
(6,1,0,1968,8,42),(6,-1,0,1104,4,24),(7,1,0,2007,8,43),(7,-1,0,1065,4,23),
(8,1,0,0,1,0),(8,-1,0,1024,2,22),(9,1,0,41,1,1),(9,-1,0,983,2,21),
(10,1,0,80,1,2),(10,-1,0,944,2,20),(11,1,1,132,1,3),(11,0,7,892,2,19),
(12,1,1,169,1,4),(12,0,7,855,2,18),(13,1,2,226,1,5),(13,0,6,798,2,17),
(14,1,3,286,1,6),(14,0,5,738,2,16),(15,1,4,343,1,7),(15,1,5,380,1,8),
(15,0,3,644,2,14),(15,0,4,681,2,15),(16,1,6,432,1,9),(16,1,7,471,1,10),
(16,0,0,512,2,11),(16,0,1,553,2,12),(16,0,2,592,2,13)]
BUMPS=[(211,252,235,276),(268,263,292,287),(185,283,209,307)]
KICKS=[(50,415,80,470),(219,415,249,470)]
LEVELS=[[(215,160,235,190),(50,165,85,185),(300,210,320,240),(3,245,22,270),(300,360,320,400)],[(10,100,40,200),(300,140,320,160),(275,165,300,220),(60,185,100,220),(195,175,215,200),(300,240,320,340),(20,240,50,400),(300,400,320,500),(260,450,277,470),(0,450,50,470)]]
def inside(r,x,y): return r[0]<=x%65536<=r[2] and r[1]<=y%65536<=r[3]
def bit(m,x,y):
 i=y*40+x//8
 return 0<=i<len(m) and bool(m[i]&(128>>(x%8)))
FIELDS=['X','Y','VX','VY','GX','GY','PixelX','PixelY','Rotation','High','Hold','Lost','HitX','HitY','CollisionAngle','ContactCount','Material']
class Model:
 def __init__(self,x=297,y=530,vx=10,vy=0,high=False):
  self.b=dict(zip(FIELDS,[x*1024,y*1024,vx,vy,0,8,x,y,0,high,False,False,0,0,0,0,0]));self.raster=4672;self.stopped=False;self.pending=None;self.events=[];self.graphics=bytearray(ORIGINAL_RGBA);self.oldframes=[0,0,0]
  self.m=[region(0x41330,23040),region(0x3b930,23040),region(0x46d30,23040),region(0x71b30,23040),region(0x77530,20400),region(0x7cf30,20400)]
  self.flips=[]
  for k,off in enumerate([0x4c730,0x4fe10,0x4ed50]):
   o=0x20690+60*k; w=words(o+2,29); f={'kind':D[o],'w':w,'speed':w[12],'angle':w[13],'frame':w[14]};stride=w[3]*w[2]*2;f['stride']=stride
   masks=region(off,stride*(w[15]+1))
   for frame in range(w[15]+1):
    for yy in range(w[3]):
     for xx in range(w[2]*2): masks[frame*stride+yy*w[2]*2+xx] |= self.m[1][(w[1]+yy)*40+w[0]//8+xx]
   f['masks']=masks;self.flips.append(f);self.copyflip(f)
 def copyflip(self,f):
  w=f['w'];stride=f['stride'];start=f['frame']*stride
  for yy in range(w[3]):
   dst=(w[1]+yy)*40+w[0]//8;src=start+yy*w[2]*2
   self.m[1][dst:dst+w[2]*2]=f['masks'][src:src+w[2]*2]
 def collision(self):
  b=self.b;x,y=b['PixelX'],b['PixelY'];m=self.m[2 if b['High'] else 1];total=down=count=quad=0;last=0
  origin=(y-1)*40+((x-1)%65536>>3);shift=((x-1)&7)+1
  for row,byte,k,angle,q,p in CP:
   off=origin+row*40+byte
   if off<0 or off+1>=len(m):continue
   word=m[off]+256*m[off+1];rot=((word<<shift)|(word>>(16-shift)))&65535
   if rot&(1<<k):total=(total+angle)&65535;count+=1;last=p;quad|=q;down+=q in (1,2)
  if not count:return
  if quad in (11,9,13):total=(total+(down<<11))&65535
  angle=total//count&2047;b['CollisionAngle']=angle;b['ContactCount']=count
  mx,my=x-1+XY[last][0],y-1+XY[last][1]
  if my>=576:return
  masks=[self.m[4],self.m[2],self.m[5]] if b['High'] else [self.m[0],self.m[1],self.m[3]]
  mat=sum(1<<i for i,m in enumerate(masks) if bit(m,mx,my));b['Material']=mat
  index=(1408*angle+32768)>>16;px,py=XY[index] if index<44 else (angle,count);hx,hy=s16(x+px),s16(y+py);b['HitX']=hx;b['HitY']=hy
  xa=ya=0
  if mat in (3,7):
   for i,r in enumerate(KICKS if mat==3 else BUMPS):
    if hx%65536<r[0]:continue
    if hy%65536<r[1]:break
    if inside(r,hx,hy):self.pending=[2 if mat==3 else 1,i,hx,hy];break
  if mat==2:
   for i,f in enumerate(self.flips):
    w=f['w'];r=(w[4],w[6],w[5],w[7])
    if not inside(r,hx,hy):continue
    self.events.append([3,i,hx,hy]);dx=s16(hx-w[8]);dy=s16(hy-w[9])
    if (f['kind']==1 and dx>=0) or (f['kind']!=1 and dx<0):break
    if w[10]:dx,dy=dy,dx;a=abs(dy>>1)
    else:
     if f['kind']==1:dx=s16(-dx);dy=s16(-dy)
     a=abs(dy)>>2
    dx=s16(-(dx+a));xa=s16(f['speed']*dy);ya=s16(f['speed']*dx);break
  return angle,count,mat,xa,ya
 def response(self,c):
  b=self.b;angle,count,mat,xa,ya=c;wf,bf,bounce,minimum,maxangle=MAT[mat]
  x=clamp(b['VX']+xa);y=clamp(b['VY']+ya);inv=(2048-angle)&2047;s= SIN[inv];co=SIN[inv+512]
  n=s16((s32(x*co-y*s)<<3)>>16);t=s16((s32(x*s+y*co)<<3)>>16)
  if n<=0:self.pending=None;return
  n=s16(-n);boost=False
  if n<minimum:
   a=abs(div(16*t,n))
   if a%65536<maxangle%65536:
    if self.pending and (mat!=3 or n<=-300):n=s16(n+(-2000 if mat==3 else -7000));boost=True
   else:n=0
  else:n=0
  if not boost:self.pending=None
  n=s16(n-div(n*256,bounce))
  if n>=-1023:factor=(-n>>6)+1;wf=s16(wf*factor);bf=s16(bf*factor)
  slip=s16(b['Rotation']-t);a=div(slip*256,wf) if wf else s16(slip*256);z=div(slip*256,bf) if bf else s16(slip*256)
  t=s16(t+a);b['Rotation']=s16(b['Rotation']-z);t=div(t*2048,2049);s=SIN[angle];co=SIN[angle+512]
  b['VX']=clamp(s16((s32(n*co-t*s)<<1)>>16)-xa);b['VY']=clamp(s16((s32(n*s+t*co)<<1)>>16)-ya)
  if count>=6:b['X']=s32(b['X']+(-1024*co>>16));b['Y']=s32(b['Y']+(-1024*s>>16))
 def step(self,left=False,right=False):
  b=self.b
  if not b['Hold']:
   c=self.collision()
   if c:self.response(c)
  for f in self.flips:
   w=f['w'];up=left if f['kind']==2 else right
   f['speed']=s16(f['speed']+(w[17] if up else w[18]))
   if up:f['speed']=min(f['speed'],w[19])
   f['angle']=s16(f['angle']-f['speed']);frame=max(0,f['angle']//55)
   if frame==0:f['speed']=f['angle']=0
   if frame>=w[15]:f['speed']=0;f['angle']=w[16];frame=w[15]
   f['frame']=frame
  if not b['Hold']:
   b['Y']=s32(b['Y']+b['VY']);b['PixelY']=div(b['Y'],1024);b['Lost']|=b['PixelY']>=576
   b['X']=s32(b['X']+b['VX']);b['PixelX']=div(b['X'],1024);b['VY']=s16(b['VY']+b['GY']);b['VX']=s16(b['VX']+b['GX'])
   b['Rotation']=max(0,b['Rotation']-2) if b['Rotation']>0 else min(0,b['Rotation']+2)
  for f in self.flips:
   w=f['w']
   if inside((w[4],w[6],w[5],w[7]),b['PixelX'],b['PixelY']):self.copyflip(f)
 def sync(self,left=False,right=False):
  self.events=[]
  if self.stopped:return
  self.step(left,right);self.step(left,right)
  if self.pending:self.events.append(self.pending);self.pending=None
  b=self.b;i=(b['PixelY']+8)%65536*40+((b['PixelX']+8)%65536>>3)-1
  for off in range(i,i+3):
   if not 0<=off<23040:continue
   if b['High']:
    if off>=20400:continue
    occ=self.m[4][off]|self.m[2][off];r=self.m[5][off]&15
   else:occ=self.m[0][off]|self.m[1][off];r=self.m[3][off]&15
   if occ==0:
    if r<4:b['GX'],b['GY']=[(0,7),(2,11),(-2,11),(-4,13)][r]
    break
  for r in LEVELS[int(b['High'])]:
   if inside(r,b['PixelX']+8,b['PixelY']+8):b['High']=not b['High'];break
  if b['Lost']:self.events.append([5,0,b['PixelX'],b['PixelY']]);self.stopped=True;return
  hx,hy=b['HitX'],b['HitY'];b['HitX']=b['HitY']=0
  if not b['High'] and (hx or hy):
   for i,r in enumerate([(130,196,146,204),(147,277,156,293),(152,293,163,311),(158,311,168,329)]):
    if inside(r,hx,hy):self.events.append([4,i,hx,hy]);break
  target=min(259,max(0,b['PixelY']-130))+33;delta=s16((target-(self.raster>>4))*9);self.raster=s16(self.raster+(delta>>2));gap=target-(self.raster>>4)
  if gap<0:
   gap+=130
   if gap<=0:self.raster=s16(self.raster+(gap<<4))
  else:
   gap-=170
   if gap>=0:self.raster=s16(self.raster+(gap<<4))
  self.step(left,right)
 def release(self,charge=32,jitter=0):self.b['VX']=0;self.b['VY']=-166*charge-jitter;self.b['Rotation']=jitter&15
 def framehash(self):
  for i,f in enumerate(self.flips):
   old=self.oldframes[i];now=f['frame'];self.oldframes[i]=now
   if old==now:continue
   o=0x20690+60*i;w=f['w'];step=words(o+56,1)[0];base=words(o+54,1)[0]%65536+8*i;maximum=words(o+48,1)[0]%65536+8*i
   record=base+old*step if now>old else maximum-old*step;distance=abs(now-old)
   while distance:
    chunk=min(9,distance)
    for countdown in range(chunk,0,-1):
     count=struct.unpack_from('<H',D,0xc2e0+record+2*(countdown-1))[0]
     for k in range(count):
      dst,src=struct.unpack_from('<HH',D,0xc2e0+record+18+4*k);row,q=divmod(dst,84)
      for plane in range(4):
       color=D[0x82930+437*plane+src-0xd4f4];x=w[0]+4*q+plane;y=w[1]+row
       if x<320 and y<576:self.graphics[(y*320+x)*4:(y*320+x)*4+3]=CMAP[color*3:color*3+3]
     record+=step
    distance-=chunk
  frame=bytearray(self.graphics);b=self.b;foreground=D[0x35f30:0x35f30+23040] if b['High'] else D[0x300b0:0x300b0+23040]
  for i,color in enumerate(BALL):
   x=b['PixelX']+i%16;y=b['PixelY']+i//16
   if color and 0<=x<320 and 0<=y<576 and not (foreground[y*40+x//8]&(128>>(x%8))):frame[(y*320+x)*4:(y*320+x)*4+3]=CMAP[color*3:color*3+3]
  origin=(self.raster>>4)-33
  assert 0<=origin<=259
  return hashlib.sha256(frame[origin*1280:(origin+317)*1280]).hexdigest()
 def snapshot(self):return {'ball':self.b.copy(),'raster':self.raster,'flippers':[[f['speed'],f['angle'],f['frame']] for f in self.flips],'events':[e[:] for e in self.events],'rgba_sha256':self.framehash()}

def fixtures():
 cases=[('setball',[297,530,10,0,False],False,False,120,100),('wall',[302,100,1200,-600,True],False,False,30,-1),('bumper',[211,239,0,1200,False],False,False,45,-1),('flipper',[115,510,0,1500,False],True,False,40,-1),('drain',[150,575,0,1024,False],False,False,5,-1),('ramp',[307,375,0,-1500,False],False,False,30,-1),('slope',[56,135,0,0,False],False,False,25,-1),('slingshot',[56,410,-2000,1500,False],False,False,35,-1)]
 result=[]
 for name,start,l,r,n,release in cases:
  m=Model(*start);frames=[m.snapshot()]
  for tick in range(n):
   if tick==release:m.release()
   m.sync(l,r);frames.append(m.snapshot())
  result.append(dict(name=name,start=[int(v) for v in start],left=l,right=r,release_at=release,frames=frames))
 return result
if __name__=='__main__':
 body=json.dumps(fixtures(),separators=(',',':'))+'\n'
 if os.environ.get('PF_REFERENCE_OUTPUT'):Path(os.environ['PF_REFERENCE_OUTPUT']).write_text(body)
 else:print(body,end='')
