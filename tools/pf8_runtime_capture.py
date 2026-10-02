#!/usr/bin/env python3
"""Bounded fresh external DOSBox-X capture, with input/time provenance.
Runs only the external oracle; never imported by a native package.
"""
import ctypes as C,sys,time,json,subprocess,contextlib,io,os
from pathlib import Path
from PIL import Image
root=Path(__file__).resolve().parent.parent
# Reuse the existing X11 bindings without targeting any existing window.
ns={'__name__':'pf8_x11'};old=sys.argv;sys.argv=['parity_pf8.py','list']
with contextlib.redirect_stdout(io.StringIO()):exec((root/'tools/parity_pf8.py').read_text(),ns)
sys.argv=old
x,d=ns['x'],ns['d'];helper=['python3',str(root/'tools/parity_pf8.py')]
def windows():return subprocess.check_output(helper+['list'],text=True).splitlines()
def capture(w,matrix):
 rt=C.c_ulong();gx,gy=C.c_int(),C.c_int();a,b,bo,de=C.c_uint(),C.c_uint(),C.c_uint(),C.c_uint()
 assert x.XGetGeometry(d,w,C.byref(rt),C.byref(gx),C.byref(gy),C.byref(a),C.byref(b),C.byref(bo),C.byref(de))
 oy=0
 if matrix:
  if (a.value,b.value)!=(640,350):return None
  oy=317;b.value=33
 im=x.XGetImage(d,w,0,oy,a.value,b.value,0xffffffff,2)
 if not im:return None
 class XImage(C.Structure):
  _fields_=[('width',C.c_int),('height',C.c_int),('xoffset',C.c_int),('format',C.c_int),('data',C.c_void_p),('byte_order',C.c_int),('bitmap_unit',C.c_int),('bitmap_bit_order',C.c_int),('bitmap_pad',C.c_int),('depth',C.c_int),('bytes_per_line',C.c_int),('bits_per_pixel',C.c_int),('red_mask',C.c_ulong),('green_mask',C.c_ulong),('blue_mask',C.c_ulong)]
 info=C.cast(im,C.POINTER(XImage)).contents
 if info.bits_per_pixel==32 and info.byte_order==0 and (info.red_mask,info.green_mask,info.blue_mask)==(0xff0000,0xff00,0xff):
  raw=C.string_at(info.data,info.bytes_per_line*info.height)
  result=Image.frombytes('RGB',(a.value,b.value),raw,'raw','BGRX',info.bytes_per_line,1)
  # Verify raw extraction against the original XGetPixel path on each image.
  for xx,yy in [(0,0),(a.value//2,b.value//2),(a.value-1,b.value-1)]:
   v=x.XGetPixel(im,xx,yy);assert result.getpixel((xx,yy))==((v>>16)&255,(v>>8)&255,v&255)
  x.XDestroyImage(im);return result
 pixels=bytearray()
 for yy in range(b.value):
  for xx in range(a.value):
   v=x.XGetPixel(im,xx,yy);pixels.extend([(v>>16)&255,(v>>8)&255,v&255])
 x.XDestroyImage(im)
 return Image.frombytes('RGB',(a.value,b.value),bytes(pixels))
conf=Path(sys.argv[1]).resolve();out=Path(sys.argv[2]).resolve();mode=sys.argv[3] if len(sys.argv)>3 else 'normal'
out.mkdir(parents=True,exist_ok=True);before=set(windows());events=[];start=time.monotonic()
with open(out/'dosbox.log','w') as log:
 process=subprocess.Popen(['/snap/bin/dosbox-x','-defaultconf','-conf',str(conf)],stdout=log,stderr=subprocess.STDOUT)
 w=None
 try:
  for _ in range(100):
   ids=[int(l.split()[0]) for l in set(windows())-before if 'DOSBox-X' in l]
   if ids:w=max(ids);break
   time.sleep(.1)
  assert w,'fresh external window unavailable'
  def key(k,hold='.03'):
   events.append({'kind':'input','key':k,'seconds':time.monotonic()-start})
   subprocess.run(helper+['key',str(w),k,hold],check=True,stdout=subprocess.DEVNULL)
  def sample(label,seconds,matrix=False,period=.12):
   end=time.monotonic()+seconds;i=0
   while time.monotonic()<end:
    t=time.monotonic();im=capture(w,matrix)
    if im:
     name=f'{label}-{i:04}.png';im.save(out/name);events.append({'kind':'frame','file':name,'seconds':t-start,'size':im.size});i+=1
    time.sleep(max(0,period-(time.monotonic()-t)))
  if mode=='normal':
   sample('startup',48,period=.04)
   sample('selector',4,period=.15)
   key('space');sample('text',3,period=.15)
   key('space');sample('return',2,period=.15)
   key('F1');sample('party-initial-attract',8,True,.04)
   key('Escape');time.sleep(.4);key('y');time.sleep(1.5)
   key('F2');sample('speed-initial-attract',8,True,.04)
  else:
   time.sleep(7);key('space');time.sleep(3);key('F2' if mode in ['bitmap','attract-speed','ending-speed','ending-real'] else 'F1')
   if mode=='ending-real':
    time.sleep(4);key('Return');time.sleep(2)
    def grid(im):return bytes(im.getpixel((xx*4,2+yy*2))[0]>81 for yy in range(16) for xx in range(160))
    target=Image.open(root/'analysis/pf8-checkpoints/table2-high-score.png').convert('RGB')
    targetdots=bytes(target.getpixel((xx*2,319+yy*2))[0]>81 for yy in range(16) for xx in range(160))
    qualified=False
    for ball in range(10):
     key('Down','.60');time.sleep(.75)
     subprocess.run(helper+['burst',str(w),'space','.03','3'],check=True,stdout=subprocess.DEVNULL)
     sample('drain-'+str(ball),7,True,.04)
     im=capture(w,True)
     if im and grid(im)==targetdots:qualified=True;break
    events.append({'kind':'qualification','observed':qualified,'seconds':time.monotonic()-start})
    if qualified:
     for letter in ['a','b','c']:
      key(letter);sample('real-initial-'+letter,.3,True,.02)
     sample('real-after-initials',10,True,.02)
   else:
    sample(mode,12,True,.025)
   if mode.startswith('ending') and mode!='ending-real':
    for letter in ['a','b','c']:
     key(letter);sample('initial-'+letter,.25,True,.02)
    sample('after-initials',8,True,.02)
 finally:
  if w:subprocess.run(helper+['close',str(w)],stdout=subprocess.DEVNULL)
  (out/'timeline.json').write_text(json.dumps({'config':str(conf),'window':w,'mode':mode,'events':events},indent=2)+'\n')
print(out)
