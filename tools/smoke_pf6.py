#!/usr/bin/env python3
"""Bounded X11/SDL key smoke test; uses a temporary score directory and dummy audio."""
import ctypes as C,os,subprocess,tempfile,time,threading
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
x.XGetGeometry.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_ulong),C.POINTER(C.c_int),C.POINTER(C.c_int),C.POINTER(C.c_uint),C.POINTER(C.c_uint),C.POINTER(C.c_uint),C.POINTER(C.c_uint)]
class KeyEvent(C.Structure):
 _fields_=[('type',C.c_int),('serial',C.c_ulong),('send_event',C.c_int),('display',C.c_void_p),('window',C.c_ulong),('root',C.c_ulong),('subwindow',C.c_ulong),('time',C.c_ulong),('x',C.c_int),('y',C.c_int),('x_root',C.c_int),('y_root',C.c_int),('state',C.c_uint),('keycode',C.c_uint),('same_screen',C.c_int)]
class Event(C.Union):
 _fields_=[('key',KeyEvent),('pad',C.c_long*24)]
x.XSendEvent.argtypes=[C.c_void_p,C.c_ulong,C.c_int,C.c_long,C.POINTER(Event)]

d=x.XOpenDisplay(None)
assert d,'X11 display unavailable'
lines=[]
with tempfile.TemporaryDirectory(prefix='pf6-smoke-') as scores:
 env=dict(os.environ,SDL_VIDEODRIVER='x11',SDL_AUDIODRIVER='dummy')
 p=subprocess.Popen(['./bin/pinballfantasies','-data-dir','.','-duration','60s','-high-score-dir',scores],env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
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
  initial=dimensions()
  requested=(max(400,initial[0]-73),max(350,initial[1]-47))
  x.XResizeWindow(d,window,*requested);x.XSync(d,0);time.sleep(.3)
  resized=dimensions();assert resized!=initial,(initial,resized)
  print('Physical window initial=%s manually resized=%s'%(initial,resized),flush=True)
  def unchanged():
   assert find(root)==window,'SDL window replaced'
   assert dimensions()==resized,('physical size changed',resized,dimensions())
  def press(key,expect=None):
   target=find(root);assert target
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
    x.XSync(d,0);time.sleep(.04)
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
  press('space','PF6: selector (table 1)')
  press('space','PF6: selector text (table 1)')
  press('space','PF6: selector (table 1)')
  press('F3','PF6: selector (table 3)');press('F4','PF6: selector (table 4)')
  press('F1','PF6: table attract (table 1)');press('Return','PF6: playing (table 1)')
  press('space');press('p','PF6: paused (table 1)');press('q','PF6: playing (table 1)')
  time.sleep(.6)
  press('p','PF6: paused (table 1)');press('Escape','PF6: quit question (table 1)');press('n','PF6: playing (table 1)')
  time.sleep(.6)
  press('p','PF6: paused (table 1)');press('Escape','PF6: quit question (table 1)');press('y','PF6: selector (table 1)')
  press('F2','PF6: table attract (table 2)');press('Return','PF6: playing (table 2)')
  press('space');press('p','PF6: paused (table 2)');press('q','PF6: playing (table 2)')
  time.sleep(.6)
  press('p','PF6: paused (table 2)');press('Escape','PF6: quit question (table 2)');press('y','PF6: selector (table 2)')
  press('Escape','PF6: quit (table 2)')
  assert p.wait(timeout=4)==0
  thread.join(timeout=2)
  assert not os.listdir(scores),'aborted game wrote scores'
  print('PF6 X11 input/session smoke PASS')
 finally:
  if p.poll() is None:p.terminate();p.wait(timeout=4)
