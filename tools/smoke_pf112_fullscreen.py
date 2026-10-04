#!/usr/bin/env python3
"""Bounded X11/SDL key smoke test; uses a temporary score directory and dummy audio."""
import ctypes as C,os,subprocess,tempfile,time,threading
from pathlib import Path
x=C.CDLL('libX11.so.6')
x.XOpenDisplay.argtypes=[C.c_char_p];x.XOpenDisplay.restype=C.c_void_p
x.XDefaultRootWindow.argtypes=[C.c_void_p];x.XDefaultRootWindow.restype=C.c_ulong
x.XQueryTree.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_ulong),C.POINTER(C.c_ulong),C.POINTER(C.POINTER(C.c_ulong)),C.POINTER(C.c_uint)]
x.XFetchName.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_char_p)]
x.XFree.argtypes=[C.c_void_p]
x.XInternAtom.argtypes=[C.c_void_p,C.c_char_p,C.c_int];x.XInternAtom.restype=C.c_ulong
x.XGetWindowProperty.argtypes=[C.c_void_p,C.c_ulong,C.c_ulong,C.c_long,C.c_long,C.c_int,C.c_ulong,C.POINTER(C.c_ulong),C.POINTER(C.c_int),C.POINTER(C.c_ulong),C.POINTER(C.c_ulong),C.POINTER(C.POINTER(C.c_ubyte))]
x.XGetInputFocus.argtypes=[C.c_void_p,C.POINTER(C.c_ulong),C.POINTER(C.c_int)]
x.XSync.argtypes=[C.c_void_p,C.c_int]
x.XSetInputFocus.argtypes=[C.c_void_p,C.c_ulong,C.c_int,C.c_ulong]
x.XStringToKeysym.argtypes=[C.c_char_p];x.XStringToKeysym.restype=C.c_ulong
x.XKeysymToKeycode.argtypes=[C.c_void_p,C.c_ulong];x.XKeysymToKeycode.restype=C.c_uint
x.XFlush.argtypes=[C.c_void_p]
x.XResizeWindow.argtypes=[C.c_void_p,C.c_ulong,C.c_uint,C.c_uint]
x.XTranslateCoordinates.argtypes=[C.c_void_p,C.c_ulong,C.c_ulong,C.c_int,C.c_int,C.POINTER(C.c_int),C.POINTER(C.c_int),C.POINTER(C.c_ulong)]
x.XGetGeometry.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_ulong),C.POINTER(C.c_int),C.POINTER(C.c_int),C.POINTER(C.c_uint),C.POINTER(C.c_uint),C.POINTER(C.c_uint),C.POINTER(C.c_uint)]
class XImage(C.Structure):
 _fields_=[('width',C.c_int),('height',C.c_int),('xoffset',C.c_int),('format',C.c_int),('data',C.c_void_p),('byte_order',C.c_int),('bitmap_unit',C.c_int),('bitmap_bit_order',C.c_int),('bitmap_pad',C.c_int),('depth',C.c_int),('bytes_per_line',C.c_int),('bits_per_pixel',C.c_int),('red_mask',C.c_ulong),('green_mask',C.c_ulong),('blue_mask',C.c_ulong)]
class KeyEvent(C.Structure):
 _fields_=[('type',C.c_int),('serial',C.c_ulong),('send_event',C.c_int),('display',C.c_void_p),('window',C.c_ulong),('root',C.c_ulong),('subwindow',C.c_ulong),('time',C.c_ulong),('x',C.c_int),('y',C.c_int),('x_root',C.c_int),('y_root',C.c_int),('state',C.c_uint),('keycode',C.c_uint),('same_screen',C.c_int)]
class Event(C.Union):
 _fields_=[('key',KeyEvent),('pad',C.c_long*24)]
x.XGetImage.argtypes=[C.c_void_p,C.c_ulong,C.c_int,C.c_int,C.c_uint,C.c_uint,C.c_ulong,C.c_int];x.XGetImage.restype=C.c_void_p
x.XGetPixel.argtypes=[C.c_void_p,C.c_int,C.c_int];x.XGetPixel.restype=C.c_ulong
x.XDestroyImage.argtypes=[C.c_void_p]
x.XSendEvent.argtypes=[C.c_void_p,C.c_ulong,C.c_int,C.c_long,C.POINTER(Event)]

d=x.XOpenDisplay(None)
assert d,'X11 display unavailable'
lines=[]
smoke_start=time.monotonic()
with tempfile.TemporaryDirectory(prefix='pf6-smoke-') as scores:
 initial_config=bytes([0,1,0,0,0,1])  # DOS MONO imports COLOR, HARD/NORMAL
 Path(scores,'PINBALL.CFG').write_bytes(initial_config)
 env=dict(os.environ,PF_DIAGNOSTICS='1',SDL_VIDEODRIVER='x11',SDL_AUDIODRIVER=os.environ.get('PF11_AUDIO_DRIVER','dummy'))
 p=subprocess.Popen(['./bin/pinballfantasies','-data-dir','.','-duration','180s','-high-score-dir',scores,'-config-dir',scores],env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
 def reader():
  for line in p.stdout:lines.append(line.rstrip());print(line.rstrip(),flush=True)
 thread=threading.Thread(target=reader);thread.start()
 try:
  root=x.XDefaultRootWindow(d)
  def find(w):
   name=C.c_char_p();x.XFetchName(d,w,C.byref(name))
   match=name.value==b'Pinball Fantasies'
   if name:x.XFree(name)
   if match:
    atom=x.XInternAtom(d,b'_NET_WM_PID',0);actual=C.c_ulong();format=C.c_int();count,left=C.c_ulong(),C.c_ulong();value=C.POINTER(C.c_ubyte)()
    x.XGetWindowProperty(d,w,atom,0,1,0,0,C.byref(actual),C.byref(format),C.byref(count),C.byref(left),C.byref(value))
    own=count.value and C.cast(value,C.POINTER(C.c_ulong))[0]==p.pid
    if value:x.XFree(value)
    if own:return w
   rt,pa=C.c_ulong(),C.c_ulong();children=C.POINTER(C.c_ulong)();n=C.c_uint()
   if x.XQueryTree(d,w,C.byref(rt),C.byref(pa),C.byref(children),C.byref(n)):
    ids=[children[i] for i in range(n.value)]
    if children:x.XFree(children)
    for c in ids:
     match=find(c)
     if match:return match
   return None
  for _ in range(100):
   if any('PF6 window opened' in line for line in lines):break
   time.sleep(.05)
  time.sleep(.2)
  window=None
  for _ in range(60):
   window=find(root)
   if window:break
   time.sleep(.05)
  assert window,'SDL window not found'
  x.XSetInputFocus(d,window,1,0);x.XSync(d,0);time.sleep(.2)
  focus=C.c_ulong();revert=C.c_int();x.XGetInputFocus(d,C.byref(focus),C.byref(revert));assert focus.value==window,(focus.value,window)
  def dimensions():
   rt=C.c_ulong();gx,gy=C.c_int(),C.c_int();w,h,border,depth=C.c_uint(),C.c_uint(),C.c_uint(),C.c_uint()
   assert x.XGetGeometry(d,window,C.byref(rt),C.byref(gx),C.byref(gy),C.byref(w),C.byref(h),C.byref(border),C.byref(depth))
   return w.value,h.value
  desktop_before=subprocess.check_output(['xrandr','--current'],text=True)
  initial=dimensions()
  requested=(800,600)
  x.XResizeWindow(d,window,*requested);x.XSync(d,0);time.sleep(.3)
  resized=dimensions();assert resized!=initial,(initial,resized)
  print('Physical window initial=%s manually resized=%s'%(initial,resized),flush=True)
  fullscreen=False
  fullscreen_dimensions=None
  def position():
   px,py=C.c_int(),C.c_int();child=C.c_ulong()
   assert x.XTranslateCoordinates(d,window,root,0,0,C.byref(px),C.byref(py),C.byref(child))
   return px.value,py.value
  windowed_position=position()
  def screenshot(name,aspect=None):
   from PIL import Image
   w,h=dimensions();im=x.XGetImage(d,window,0,0,w,h,0xffffffff,2)
   assert im,'X11 screenshot failed'
   try:
    header=C.cast(im,C.POINTER(XImage)).contents
    assert header.bits_per_pixel==32 and header.byte_order==0,(header.bits_per_pixel,header.byte_order)
    assert (header.red_mask,header.green_mask,header.blue_mask)==(0xff0000,0xff00,0xff)
    pixels=C.string_at(header.data,header.bytes_per_line*h)
    dest=Path('analysis/pf11.2-validation/x11-fullscreen');dest.mkdir(parents=True,exist_ok=True)
    image=Image.frombytes('RGB',(w,h),pixels,'raw','BGRX',header.bytes_per_line,1);image.save(dest/(name+'.png'))
    if aspect:
     box=image.getbbox();ratio=(box[2]-box[0])/(box[3]-box[1]);assert abs(ratio-aspect)<.02,(name,box,ratio,aspect)
     print('Fullscreen content aspect',name,box,ratio,flush=True)
   finally:x.XDestroyImage(im)
  def unchanged():
   assert p.poll() is None,'native process exited'
   assert find(root)==window,'X11 window recreated'
   if fullscreen:
    assert dimensions()==fullscreen_dimensions,('fullscreen size changed',fullscreen_dimensions,dimensions())
   else:
    assert dimensions()==resized,('physical size changed',resized,dimensions())
    assert position()==windowed_position,('window position changed',windowed_position,position())
  def press(key,expect=None,hold=.04):
   print('Smoke input',key,'elapsed',round(time.monotonic()-smoke_start,2),flush=True)
   target=window;assert p.poll() is None,'native process exited'
   start=len(lines)
   x.XGetInputFocus(d,C.byref(focus),C.byref(revert))
   if focus.value!=target:x.XSetInputFocus(d,target,1,0)
   x.XSync(d,0);time.sleep(.15)
   code=x.XKeysymToKeycode(d,x.XStringToKeysym(key.encode()))
   assert code
   for event_type,mask in [(2,1),(3,2)]:
    if event_type==3 and expect and expect.startswith('PF6: quit ('):break
    event=Event();event.key.type=event_type;event.key.display=d;event.key.window=target;event.key.root=root;event.key.same_screen=1;event.key.keycode=code
    assert x.XSendEvent(d,target,1,mask,C.byref(event))
    x.XFlush(d);time.sleep(hold if event_type==2 else .04)
   if expect:
    for _ in range(80):
     if any(expect in line for line in lines[start:]):
      if not expect.startswith('PF6: quit ('):unchanged()
      return
     if p.poll() is not None:break
     time.sleep(.025)
    raise AssertionError('Missing transition '+expect)
   time.sleep(.2)
   unchanged()
  def toggle(alt='Alt_L'):
   global fullscreen,fullscreen_dimensions
   before=Path(scores,'PINBALL.CFG').read_bytes();start=len(lines)
   def send(key,event_type,state=0,delay=.08):
    event=Event();event.key.type=event_type;event.key.display=d;event.key.window=window;event.key.root=root;event.key.same_screen=1
    event.key.keycode=x.XKeysymToKeycode(d,x.XStringToKeysym(key.encode()));event.key.state=state
    assert x.XSendEvent(d,window,1,1 if event_type==2 else 2,C.byref(event));x.XFlush(d);time.sleep(delay)
   send(alt,2);send('Return',2,8,0)
   send('Return',2,8,0);send('Return',2,8,.25) # duplicate physical downs must never toggle twice
   send('Return',3,8);send(alt,3)
   expected=not fullscreen
   for _ in range(80):
    transitions=[l for l in lines[start:] if l.startswith('PF11.2 fullscreen=')]
    if transitions:break
    time.sleep(.025)
   time.sleep(.5)
   transitions=[l for l in lines[start:] if l.startswith('PF11.2 fullscreen=')]
   assert len(transitions)==1,transitions
   assert ('fullscreen='+str(expected).lower()) in transitions[0],transitions
   assert not any(l.startswith('PF6: ') for l in lines[start:]),'shortcut Enter leaked into frontend'
   assert Path(scores,'PINBALL.CFG').read_bytes()==before,'fullscreen modified config'
   continuity=[l for l in lines[start:] if l.startswith('PF11.2 fullscreen audio lifecycle')]
   assert len(continuity)==1,continuity
   for field in ['clears=','starts=']:
    pair=continuity[0].split(field)[1].split()[0].split('/');assert pair[0]==pair[1],continuity
   assert 'producer unchanged=true' in continuity[0],continuity
   fullscreen=expected
   if fullscreen:
    fullscreen_dimensions=dimensions();assert fullscreen_dimensions[0]>=resized[0] and fullscreen_dimensions[1]>resized[1],fullscreen_dimensions
   unchanged()
   print('Alt+Enter',alt,'fullscreen=',fullscreen,'dimensions=',dimensions(),'position=',position(),flush=True)
  toggle();toggle('Alt_R') # startup remains startup: Return must be consumed
  press('space','PF6: selector (table 1)')
  press('F1','PF6: table attract (table 1)') # HARD/NORMAL
  toggle()
  screenshot('normal-fullscreen',320/240)
  press('Escape','PF6: quit question (table 1)');press('y','PF6: selector (table 1)')
  press('F5','PF6: options (table 1)');time.sleep(1)
  toggle('Alt_R');toggle() # consume Enter on BALLS without changing it
  for _ in range(4):press('Down')
  press('Return') # HIGH while fullscreen
  press('Escape','PF6: selector (table 1)');press('F1','PF6: table attract (table 1)')
  assert any('presentation 320x350' in l and 'fullscreen=true' in l for l in lines),'HIGH fullscreen missing'
  screenshot('high-fullscreen',320/350)
  toggle();toggle('Alt_R') # HIGH windowed -> fullscreen
  press('Escape','PF6: quit question (table 1)');press('y','PF6: selector (table 1)')
  press('F5','PF6: options (table 1)');time.sleep(1)
  press('Down');press('Down')
  for _ in range(3):press('Return') # HARD -> MEDIUM -> SOFT -> OFF
  press('Escape','PF6: selector (table 1)');press('F1','PF6: table attract (table 1)')
  assert any('presentation 320x609' in l and 'fullscreen=true' in l for l in lines),'OFF fullscreen missing'
  screenshot('off-fullscreen',320/609)
  press('Return','PF6: playing (table 1)');time.sleep(3)
  toggle();toggle('Alt_R') # gameplay stays playing; no pause/audio producer transition
  screenshot('playing-off-fullscreen',320/609)
  press('p','PF6: paused (table 1)');press('Escape','PF6: quit question (table 1)');press('y','PF6: selector (table 1)')
  press('F5','PF6: options (table 1)');time.sleep(1)
  press('Down');press('Down');press('Return') # OFF -> HARD
  press('Down');press('Down');press('Return') # HIGH -> NORMAL
  press('Escape','PF6: selector (table 1)');press('F1','PF6: table attract (table 1)')
  assert any('presentation 320x240' in l and 'fullscreen=true' in l for l in lines),'NORMAL fullscreen restoration missing'
  screenshot('hard-normal-fullscreen',320/240)
  toggle()
  screenshot('restored-windowed',320/240)
  press('Escape','PF6: quit question (table 1)');press('y','PF6: selector (table 1)');press('Escape','PF6: quit (table 1)')
  assert p.wait(timeout=4)==0
  assert subprocess.check_output(['xrandr','--current'],text=True)==desktop_before,'desktop mode/refresh changed'
  print('Xrandr desktop modes/refresh unchanged')
  thread.join(timeout=2)
  assert os.listdir(scores)==['PINBALL.CFG'],os.listdir(scores)
  assert Path(scores,'PINBALL.CFG').read_bytes()==b'PFNC'+bytes([1,5,0,1,0,0,0]),'shortcut changed BALLS/config'
  ids={line.split('window ID=')[1].split()[0] for line in lines if 'window ID=' in line}
  assert len(ids)==1,ids
  print('Persistent SDL window IDs:',ids)
  print('PF11.2 borderless desktop fullscreen / repeat suppression / Enter consumption / geometry and logical switching smoke PASS')
 finally:
  if p.poll() is None:p.terminate();p.wait(timeout=4)
