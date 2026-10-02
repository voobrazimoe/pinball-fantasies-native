#!/usr/bin/env python3
"""Run the external oracle on an owned, private X server, avoiding user input."""
import os,subprocess,sys
from pathlib import Path
root=Path(__file__).resolve().parent.parent
server=subprocess.Popen([str(root/'.tools/xvfb/usr/bin/Xvfb'),'-displayfd','1','-screen','0','1024x768x24','-ac','-nolisten','tcp'],stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
try:
 display=server.stdout.readline().strip();assert display.isdigit(),server.stderr.read()
 env=dict(os.environ,DISPLAY=':'+display,SDL_VIDEODRIVER='x11');env.pop('XAUTHORITY',None)
 print('Owned isolated X display',env['DISPLAY'],flush=True)
 subprocess.run(['python3',str(root/'tools/pf8_runtime_capture.py'),*sys.argv[1:]],env=env,check=True)
finally:
 server.terminate();server.wait(timeout=5)
