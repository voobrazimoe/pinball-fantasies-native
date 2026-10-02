import os,subprocess,tempfile,time
from pathlib import Path
root=Path.cwd();helper=['python3',str(root/'tools/parity_pf8.py')]
def windows():return subprocess.check_output(helper+['list'],text=True).splitlines()
with tempfile.TemporaryDirectory(prefix='pf5-fixes-scores-') as scores:
 with open('/tmp/pf5-fixes-pulse.log','w') as log:
  game=subprocess.Popen([str(root/'bin/pinballfantasies'),'-data-dir',str(root),'-high-score-dir',scores,'-duration','20s'],env=dict(os.environ,SDL_VIDEODRIVER='x11',SDL_AUDIODRIVER='pulseaudio'),stdout=log,stderr=subprocess.STDOUT)
  try:
   w=None
   for _ in range(100):
    ids=[int(l.split()[0]) for l in windows() if 'Pinball Fantasies' in l and 'pid='+str(game.pid)+' ' in l]
    if ids:w=max(ids);break
    time.sleep(.05)
   assert w
   def key(k):subprocess.run(helper+['key',str(w),k,'.08'],check=True);time.sleep(.7)
   time.sleep(.7)
   key('space');key('F1');key('Return');key('m');key('Shift_L');key('Shift_R');key('space');key('p');key('Escape');key('y')
   key('F2');key('Return');key('m');key('Shift_L');key('Shift_R');key('space')
   game.wait(timeout=12);assert game.returncode==0
  finally:
   if game.poll() is None:game.terminate();game.wait()
text=Path('/tmp/pf5-fixes-pulse.log').read_text();print(text)
assert 'driver: pulseaudio' in text
assert 'PF6: playing (table 1)' in text and 'PF6: playing (table 2)' in text
assert 'queue resets=0 empty-queue observations=0' in text
print('PF5 audio fixes desktop smoke PASS')
