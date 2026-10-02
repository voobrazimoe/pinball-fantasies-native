#!/usr/bin/env python3
"""External X11 capture/input helper. Never linked to the native game.
Usage: parity_pf8.py list | capture WINDOW_ID OUTPUT.png | key WINDOW_ID KEYSYM [HOLD_SECONDS] | close WINDOW_ID
Captures the window's host pixels; use native -png for logical-frame comparison.
"""
import ctypes as C
import sys,time
from PIL import Image
x=C.CDLL('libX11.so.6')
x.XOpenDisplay.argtypes=[C.c_char_p];x.XOpenDisplay.restype=C.c_void_p
x.XDefaultRootWindow.argtypes=[C.c_void_p];x.XDefaultRootWindow.restype=C.c_ulong
x.XQueryTree.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_ulong),C.POINTER(C.c_ulong),C.POINTER(C.POINTER(C.c_ulong)),C.POINTER(C.c_uint)]
x.XFetchName.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_char_p)]
x.XFree.argtypes=[C.c_void_p]
x.XGetGeometry.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_ulong),C.POINTER(C.c_int),C.POINTER(C.c_int),C.POINTER(C.c_uint),C.POINTER(C.c_uint),C.POINTER(C.c_uint),C.POINTER(C.c_uint)]
x.XGetImage.argtypes=[C.c_void_p,C.c_ulong,C.c_int,C.c_int,C.c_uint,C.c_uint,C.c_ulong,C.c_int];x.XGetImage.restype=C.c_void_p
x.XGetPixel.argtypes=[C.c_void_p,C.c_int,C.c_int];x.XGetPixel.restype=C.c_ulong
x.XDestroyImage.argtypes=[C.c_void_p]
x.XStringToKeysym.argtypes=[C.c_char_p];x.XStringToKeysym.restype=C.c_ulong
x.XKeysymToKeycode.argtypes=[C.c_void_p,C.c_ulong];x.XKeysymToKeycode.restype=C.c_uint
x.XInternAtom.argtypes=[C.c_void_p,C.c_char_p,C.c_int];x.XInternAtom.restype=C.c_ulong
x.XGetWindowProperty.argtypes=[C.c_void_p,C.c_ulong,C.c_ulong,C.c_long,C.c_long,C.c_int,C.c_ulong,C.POINTER(C.c_ulong),C.POINTER(C.c_int),C.POINTER(C.c_ulong),C.POINTER(C.c_ulong),C.POINTER(C.POINTER(C.c_ubyte))]
x.XRaiseWindow.argtypes=[C.c_void_p,C.c_ulong]
x.XSetInputFocus.argtypes=[C.c_void_p,C.c_ulong,C.c_int,C.c_ulong]
x.XSync.argtypes=[C.c_void_p,C.c_int]
class Key(C.Structure):
 _fields_=[('type',C.c_int),('serial',C.c_ulong),('send_event',C.c_int),('display',C.c_void_p),('window',C.c_ulong),('root',C.c_ulong),('subwindow',C.c_ulong),('time',C.c_ulong),('x',C.c_int),('y',C.c_int),('x_root',C.c_int),('y_root',C.c_int),('state',C.c_uint),('keycode',C.c_uint),('same_screen',C.c_int)]
class ClientData(C.Union):_fields_=[('b',C.c_char*20),('s',C.c_short*10),('l',C.c_long*5)]
class Client(C.Structure):_fields_=[('type',C.c_int),('serial',C.c_ulong),('send_event',C.c_int),('display',C.c_void_p),('window',C.c_ulong),('message_type',C.c_ulong),('format',C.c_int),('data',ClientData)]
class Event(C.Union):_fields_=[('key',Key),('client',Client),('pad',C.c_long*24)]
x.XSendEvent.argtypes=[C.c_void_p,C.c_ulong,C.c_int,C.c_long,C.POINTER(Event)]
d=x.XOpenDisplay(None)
if not d:raise SystemExit('X11 unavailable')
root=x.XDefaultRootWindow(d)
def walk(w):
 name=C.c_char_p();x.XFetchName(d,w,C.byref(name))
 if name.value:
  actual=C.c_ulong();format=C.c_int();count,left=C.c_ulong(),C.c_ulong();value=C.POINTER(C.c_ubyte)()
  x.XGetWindowProperty(d,w,x.XInternAtom(d,b'_NET_WM_PID',0),0,1,0,0,C.byref(actual),C.byref(format),C.byref(count),C.byref(left),C.byref(value))
  pid=C.cast(value,C.POINTER(C.c_ulong))[0] if count.value else 0
  if value:x.XFree(value)
  print(w, 'pid='+str(pid),name.value.decode(errors='replace'))
 if name:x.XFree(name)
 rt,pa=C.c_ulong(),C.c_ulong();children=C.POINTER(C.c_ulong)();n=C.c_uint()
 if x.XQueryTree(d,w,C.byref(rt),C.byref(pa),C.byref(children),C.byref(n)):
  ids=[children[i] for i in range(n.value)]
  if children:x.XFree(children)
  for c in ids:walk(c)
if sys.argv[1]=='list':walk(root)
else:
 w=int(sys.argv[2],0)
 if sys.argv[1]=='close':
  event=Event();event.client.type=33;event.client.display=d;event.client.window=w;event.client.message_type=x.XInternAtom(d,b'_NET_CLOSE_WINDOW',0);event.client.format=32;event.client.data.l[1]=2
  x.XSendEvent(d,root,0,(1<<20)|(1<<19),C.byref(event));x.XSync(d,0)
 elif sys.argv[1] in ['key','keydown','keyup','burst']:
  event=Event();event.client.type=33;event.client.display=d;event.client.window=w;event.client.message_type=x.XInternAtom(d,b'_NET_ACTIVE_WINDOW',0);event.client.format=32;event.client.data.l[0]=2
  x.XSendEvent(d,root,0,(1<<20)|(1<<19),C.byref(event));x.XSync(d,0);time.sleep(.2)
  x.XRaiseWindow(d,w);x.XSetInputFocus(d,w,1,0);x.XSync(d,0)
  code=x.XKeysymToKeycode(d,x.XStringToKeysym(sys.argv[3].encode()))
  events = ([(2,1)] if sys.argv[1]=='keydown' else [(3,2)] if sys.argv[1]=='keyup' else [(2,1),(3,2)])
  if sys.argv[1]=='burst': events *= int(sys.argv[5]) if len(sys.argv)>5 else 4
  for typ,mask in events:
   e=Event();e.key.type=typ;e.key.display=d;e.key.window=w;e.key.root=root;e.key.same_screen=1;e.key.keycode=code
   assert x.XSendEvent(d,w,1,mask,C.byref(e));x.XSync(d,0)
   time.sleep(float(sys.argv[4]) if typ==2 and len(sys.argv)>4 else (float(sys.argv[4]) if sys.argv[1]=='burst' and len(sys.argv)>4 else .05))
 elif sys.argv[1] in ['capture','matrix']:
  rt=C.c_ulong();gx,gy=C.c_int(),C.c_int();a,b,bo,de=C.c_uint(),C.c_uint(),C.c_uint(),C.c_uint()
  assert x.XGetGeometry(d,w,C.byref(rt),C.byref(gx),C.byref(gy),C.byref(a),C.byref(b),C.byref(bo),C.byref(de))
  oy=0
  if sys.argv[1]=='matrix':
   assert (a.value,b.value)==(640,350),'Matrix oracle must be the recovered raw high-mode surface'
   oy=317;b.value=33
  im=x.XGetImage(d,w,0,oy,a.value,b.value,0xffffffff,2);assert im
  pixels=bytearray()
  for yy in range(b.value):
   for xx in range(a.value):
    v=x.XGetPixel(im,xx,yy);pixels.extend([(v>>16)&255,(v>>8)&255,v&255])
  Image.frombytes('RGB',(a.value,b.value),bytes(pixels)).save(sys.argv[3]);x.XDestroyImage(im)
  print(sys.argv[3],a.value,b.value)
