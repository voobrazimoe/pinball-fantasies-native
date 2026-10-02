#!/usr/bin/env python3
"""Run a PF8 smoke/audio command on an owned X display, without desktop input."""
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parent.parent
server = subprocess.Popen(
    [str(root / '.tools/xvfb/usr/bin/Xvfb'), '-displayfd', '1', '-screen', '0',
     '1600x1800x24', '-ac', '-nolisten', 'tcp'],
    stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
try:
    display = server.stdout.readline().strip()
    assert display.isdigit(), server.stderr.read()
    env = dict(os.environ, DISPLAY=':' + display, SDL_VIDEODRIVER='x11')
    env.pop('XAUTHORITY', None)
    print('Owned isolated X display', env['DISPLAY'], flush=True)
    result = subprocess.run(sys.argv[1:], env=env)
    sys.exit(result.returncode)
finally:
    server.terminate()
    server.wait(timeout=5)
