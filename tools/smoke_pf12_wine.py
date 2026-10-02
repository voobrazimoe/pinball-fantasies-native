#!/usr/bin/env python3
"""Additional Wine compatibility smoke, reusing the bounded PF11.2 X11 journey.

Run on an X11 desktop after building bin/pinballfantasies-debug.exe. This is
additional evidence only; real Windows visual/audio acceptance remains required.
Use --fullscreen for the Alt+Enter/geometry/audio continuity journey.
"""
import os
import subprocess
import re
import sys
from pathlib import Path

root = Path(__file__).resolve().parent.parent
os.chdir(root)
source = root / ('tools/smoke_pf112_fullscreen.py' if '--fullscreen' in sys.argv else 'tools/smoke_pf112.py')
def win_geometry():
    output = subprocess.check_output(['wine64','./bin/platform-windows.test.exe','-test.v','-test.run','TestPF12HostDriver'],env=dict(os.environ,WINEPREFIX=os.environ.get('PF12_WINEPREFIX','/tmp/pf12-wine'),WINEDEBUG='-all',PF12_DRIVER_ACTION='geometry'),text=True)
    return tuple(map(int,re.search(r'geometry=(-?\d+),(-?\d+),(-?\d+),(-?\d+)',output).groups()))
def win_client_crop(image):
    output = subprocess.check_output(['wine64','./bin/platform-windows.test.exe','-test.v','-test.run','TestPF12HostDriver'],env=dict(os.environ,WINEPREFIX=os.environ.get('PF12_WINEPREFIX','/tmp/pf12-wine'),WINEDEBUG='-all',PF12_DRIVER_ACTION='client'),text=True)
    left, top, width, height = map(int,re.search(r'client=(-?\d+),(-?\d+),(-?\d+),(-?\d+)',output).groups())
    # Wine/GNOME's X11 bridge includes client-side titlebar and shadows. Locate
    # the real Win32 client in that image rather than measuring decorations.
    px, py = scope['C'].c_int(), scope['C'].c_int()
    child = scope['C'].c_ulong()
    assert scope['x'].XTranslateCoordinates(scope['d'],scope['window'],scope['root'],0,0,scope['C'].byref(px),scope['C'].byref(py),scope['C'].byref(child))
    left, top = left-px.value, top-py.value
    assert 0 <= left and 0 <= top and left+width <= image.width and top+height <= image.height
    return image.crop((left,top,left+width,top+height))
s = source.read_text()
s = s.replace('def position():', 'def position():\n   return win_geometry()[:2]')
s = s.replace("env=dict(os.environ,SDL_VIDEODRIVER='x11',SDL_AUDIODRIVER=os.environ.get('PF11_AUDIO_DRIVER','dummy'))", "env=dict(os.environ,WINEPREFIX=os.environ.get('PF12_WINEPREFIX','/tmp/pf12-wine'),WINEDEBUG='-all')")
s = s.replace("['./bin/pinballfantasies','-data-dir','.','-duration','180s','-high-score-dir',scores,'-config-dir',scores]", "['wine64','./bin/pinballfantasies-debug.exe','-data-dir','Z:'+str(root).replace('/','\\\\'),'-duration','180s','-high-score-dir','Z:'+scores.replace('/','\\\\'),'-config-dir','Z:'+scores.replace('/','\\\\')]")
# Wine's X11 window PID is not necessarily the launcher PID. Only one tested
# Pinball Fantasies process should be running on this desktop during the smoke.
s = s.replace('if own:return w', 'return w')
s = s.replace('x.XResizeWindow(d,window,*requested)', "subprocess.check_call(['wine64','./bin/platform-windows.test.exe','-test.run','TestPF12HostDriver'],env=dict(env,PF12_DRIVER_ACTION='resize'))")
s = s.replace('assert x.XSendEvent(d,target,1,mask,C.byref(event))', "subprocess.check_call(['wine64','./bin/platform-windows.test.exe','-test.run','TestPF12HostDriver'],env=dict(env,PF12_DRIVER_ACTION='key:%s:%d:%d'%(key,event_type,event.key.state)))")
s = s.replace('assert x.XSendEvent(d,window,1,1 if event_type==2 else 2,C.byref(event))', "subprocess.check_call(['wine64','./bin/platform-windows.test.exe','-test.run','TestPF12HostDriver'],env=dict(env,PF12_DRIVER_ACTION='key:%s:%d:%d'%(key,event_type,state)))")
# Wine may recreate its X11 bridge while keeping the application's HWND. Find
# that current bridge for captures; the game's native HWND log proves identity.
s = s.replace('def dimensions():', 'def dimensions():\n   global window\n   window=find(root)\n   assert window, "Wine X11 bridge missing"')
s = s.replace("'PF6 window opened'", "'Win32 window opened'")
# Posted repeats must carry Win32's previous-key-state bit. Independent driver
# processes can change desktop focus; they cannot stand in for physical held keys.
s = s.replace("send('Return',2,8,0);send('Return',2,8,.25)", "send('Return',2,24,0);send('Return',2,24,.25)")
s = s.replace('PF11.2', 'PF12')
s = s.replace('analysis/pf11.2-validation', 'analysis/pf12-validation/wine')
s = s.replace('Persistent SDL window IDs:', 'Persistent Win32 HWNDs:')
s = s.replace('SDL window not found', 'Wine Win32 window not found')
s = s.replace("image.save(dest/(name+'.png'))", "image=win_client_crop(image);image.save(dest/(name+'.png'))")
s = s.replace("assert find(root)==window,'X11 window recreated'", "assert find(root), 'Wine X11 bridge unavailable'")
scope = globals()
exec(compile(s, str(source), 'exec'), scope)
