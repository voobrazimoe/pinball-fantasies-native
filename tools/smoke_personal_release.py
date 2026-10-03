#!/usr/bin/env python3
"""Local cold-directory personal artifact smoke (Linux X11 or Windows/Wine).

Only the artifact is copied. Logs/metadata remain in ignored .build-personal;
this is automated evidence, not real-Windows Explorer/listening acceptance.
"""
import argparse
import array
import signal
import wave
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import time

repo = Path(__file__).resolve().parent.parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--backend', choices=['linux', 'wine'], required=True)
parser.add_argument('--capture-audio', action='store_true', help='record local PulseAudio monitor and check each table output')
args = parser.parse_args()
# Reuse the existing Xlib bindings, without running its test journey.
exec((repo/'tools/smoke_pf112.py').read_text().split('d=x.XOpenDisplay(None)')[0])
d = x.XOpenDisplay(None)
assert d, 'X11 desktop required'
root = x.XDefaultRootWindow(d)
backend = args.backend
artifact = repo/('release/personal/linux/PinballFantasies-x86_64.AppImage' if backend == 'linux' else 'release/personal/windows/pinballfantasies.exe')
sha = hashlib.sha256(artifact.read_bytes()).hexdigest()
winpath = lambda p: 'Z:'+str(p).replace('/', '\\')
out = repo/'.build-personal'
out.mkdir(exist_ok=True)
with tempfile.TemporaryDirectory(prefix='Personal Ж 日本 empty ') as tmp:
 folder = Path(tmp)/'Игра 日本 with spaces'; folder.mkdir()
 game = folder/artifact.name; shutil.copy2(artifact, game)
 assert list(folder.iterdir()) == [game]
 env = dict(os.environ, SDL_VIDEODRIVER='x11', SDL_AUDIODRIVER='pulseaudio', WINEPREFIX=os.environ.get('PF12_WINEPREFIX','/tmp/pf12-wine'), WINEDEBUG='-all')
 trace=out/(backend+'-transitions.jsonl');trace.unlink(missing_ok=True)
 if backend=='wine':env['PF12_TRANSITION_LOG']=winpath(trace)
 command = [str(game)] if backend == 'linux' else ['wine64', str(game)]
 log = folder/'userdata/native.log'
 def content():
  return log.read_text(errors='replace') if log.exists() else ''
 def find(w):
  name = C.c_char_p(); x.XFetchName(d,w,C.byref(name)); match = name.value == b'Pinball Fantasies'
  if name: x.XFree(name)
  if match:
   atom=x.XInternAtom(d,b'_NET_WM_PID',0);actual=C.c_ulong();fmt=C.c_int();count,left=C.c_ulong(),C.c_ulong();value=C.POINTER(C.c_ubyte)()
   x.XGetWindowProperty(d,w,atom,0,1,0,0,C.byref(actual),C.byref(fmt),C.byref(count),C.byref(left),C.byref(value))
   if value:x.XFree(value)
   if count.value:return w
  rt,pa=C.c_ulong(),C.c_ulong();children=C.POINTER(C.c_ulong)();n=C.c_uint()
  if x.XQueryTree(d,w,C.byref(rt),C.byref(pa),C.byref(children),C.byref(n)):
   ids=[children[i] for i in range(n.value)]
   if children:x.XFree(children)
   for child in ids:
    found=find(child)
    if found:return found
  return None
 def wait_for(text, start=0):
  deadline=time.monotonic()+10
  while time.monotonic()<deadline:
   if text in content()[start:]:return
   if p.poll() is not None:break
   time.sleep(.05)
  raise AssertionError('Missing '+text+'\n'+content()[-1500:])
 def send(key,state=0,hold=.05):
  window=find(root); assert window
  x.XSetInputFocus(d,window,1,0);x.XSync(d,0);time.sleep(.2)
  focus=C.c_ulong();revert=C.c_int();x.XGetInputFocus(d,C.byref(focus),C.byref(revert));assert focus.value==window,(focus.value,window)
  if backend=='linux':
   # Activate through the window manager, which may reject focus-only requests.
   x.XTranslateCoordinates.argtypes=[C.c_void_p,C.c_ulong,C.c_ulong,C.c_int,C.c_int,C.POINTER(C.c_int),C.POINTER(C.c_int),C.POINTER(C.c_ulong)]
   px,py=C.c_int(),C.c_int();child=C.c_ulong()
   assert x.XTranslateCoordinates(d,window,root,20,20,C.byref(px),C.byref(py),C.byref(child))
   xt=C.CDLL('libXtst.so.6');xt.XTestFakeMotionEvent.argtypes=[C.c_void_p,C.c_int,C.c_int,C.c_int,C.c_ulong];xt.XTestFakeButtonEvent.argtypes=[C.c_void_p,C.c_uint,C.c_int,C.c_ulong]
   xt.XTestFakeMotionEvent(d,-1,px.value,py.value,0);xt.XTestFakeButtonEvent(d,1,1,0);xt.XTestFakeButtonEvent(d,1,0,0);x.XFlush(d);time.sleep(.15)
  print('Input',key,'window',window,flush=True)
  for kind in [2,3]:
   if backend == 'wine':
    subprocess.run(['wine64',str(repo/'bin/platform-windows.test.exe'),'-test.run','TestPF12HostDriver'],env=dict(env,PF12_DRIVER_ACTION=f'key:{key}:{kind}:{state}'),check=True,stdout=subprocess.DEVNULL)
   else:
    # XTest emits ordinary server key events; bundled SDL may use XInput2.
    xt=C.CDLL('libXtst.so.6');xt.XTestFakeKeyEvent.argtypes=[C.c_void_p,C.c_uint,C.c_int,C.c_ulong]
    if state&8 and kind==2:
     alt=x.XKeysymToKeycode(d,x.XStringToKeysym(b'Alt_L'));xt.XTestFakeKeyEvent(d,alt,1,0)
    code=x.XKeysymToKeycode(d,x.XStringToKeysym(key.encode()))
    assert xt.XTestFakeKeyEvent(d,code,int(kind==2),0)
    if state&8 and kind==3:xt.XTestFakeKeyEvent(d,alt,0,0)
    x.XFlush(d)
   time.sleep(hold if kind==2 else .1)
 def press(key,expect=None,**kw):
  start=len(content());send(key,**kw)
  if expect:wait_for(expect,start)
 def launch(extra=[]):
  return subprocess.Popen(command+extra,cwd='/tmp',env=env,stdout=subprocess.DEVNULL,stderr=(out/(backend+'-host-stderr.log')).open('a'))
 capture=None;intervals=[];record_start=time.monotonic()
 if args.capture_audio:
  capture=subprocess.Popen(['ffmpeg','-nostdin','-hide_banner','-y','-f','pulse','-sample_rate','48000','-channels','2','-i','@DEFAULT_MONITOR@',str(out/(backend+'-audio.wav'))],stdout=subprocess.DEVNULL,stderr=(out/(backend+'-capture.log')).open('w'))
  time.sleep(1);record_start=time.monotonic()-1
 p=launch(['-duration','120s'])
 try:
  wait_for('window opened');time.sleep(2)
  press('space','PF6: selector (table 1)')
  # Supplied DOS seed has music OFF; enable normally through the runtime control.
  press('F5','PF6: options (table 1)');time.sleep(1)
  for _ in range(3):press('Down')
  press('Return');press('Escape','PF6: selector (table 1)')
  for _ in range(4):press('Return',state=8);time.sleep(.5)
  for table in range(1,5):
   press(f'F{table}',f'PF6: table attract (table {table})')
   press('Return',f'PF6: playing (table {table})');audio_start=time.monotonic()-record_start;time.sleep(3)
   press('Down',hold=.4);press('Shift_L');press('Shift_R');time.sleep(.6)
   intervals.append((table,audio_start,time.monotonic()-record_start))
   press('p',f'PF6: paused (table {table})');press('Escape',f'PF6: quit question (table {table})');press('y',f'PF6: selector (table {table})')
  press('F5','PF6: options (table 4)');time.sleep(1)
  press('Return') # save a changed balls setting
  press('Escape','PF6: selector (table 4)')
  # Quit key-down only: the window disappears before a release can be posted.
  if backend=='wine':
   subprocess.run(['wine64',str(repo/'bin/platform-windows.test.exe'),'-test.run','TestPF12HostDriver'],env=dict(env,PF12_DRIVER_ACTION='key:Escape:2:0'),check=True,stdout=subprocess.DEVNULL)
  else:send('Escape')
  assert p.wait(timeout=10)==0
 finally:
  if capture:
   capture.send_signal(signal.SIGINT);capture.wait(timeout=10)
  (out/(backend+'-latest.log')).write_text(content())
  if p.poll() is None:p.terminate();p.wait(timeout=10)
 cfg=folder/'userdata/PINBALL.CFG';saved=cfg.read_bytes();assert saved.startswith(b'PFNC')
 first_log=content();(out/(backend+'-cold-live.log')).write_text(first_log)
 assert 'Audio output unavailable' not in first_log
 assert set(p.name for p in folder.iterdir())=={game.name,'userdata'},list(folder.iterdir())
 # Restart through bundled defaults and ensure saved native settings survive.
 png=Path(tmp)/'restart.png'
 p=launch(['-png',str(png) if backend=='linux' else winpath(png)])
 assert p.wait(timeout=20)==0 and png.is_file()
 assert cfg.read_bytes()==saved
 # Explicit original-data override remains available on both platforms.
 p=launch(['-data-dir',str(repo) if backend=='linux' else winpath(repo),'-png',str(png) if backend=='linux' else winpath(png)])
 assert p.wait(timeout=20)==0
 assert hashlib.sha256(game.read_bytes()).hexdigest()==sha
 result={'backend':backend,'sha256':sha,'cold_directory_one_file':True,'all_four_tables':True,'music_enabled_and_flipper_spring_inputs':True,'four_alt_enter_edges':True,'external_userdata':True,'settings_restart_preserved':True,'explicit_data_override':True,'artifact_unchanged':True,'unicode_spaces':True}
 if capture:
  with wave.open(str(out/(backend+'-audio.wav')),'rb') as audio:
   rate=audio.getframerate();assert audio.getnchannels()==2 and audio.getsampwidth()==2
   output=[]
   for table,start,end in intervals:
    audio.setpos(int(start*rate));samples=array.array('h',audio.readframes(int((end-start)*rate)))
    nonzero=sum(v!=0 for v in samples);assert nonzero>rate,('silent table',table,nonzero)
    output.append({'table':table,'nonzero_samples':nonzero,'peak':max(abs(v) for v in samples)})
   result['captured_table_audio']=output
 (out/(backend+'-cold-smoke.json')).write_text(json.dumps(result,indent=2)+'\n')
 print(json.dumps(result))
