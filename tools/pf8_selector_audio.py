#!/usr/bin/env python3
"""Real PulseAudio + normal SDL frontend transition probe. Isolated score store."""
import array,json,os,subprocess,tempfile,time,sys,wave
from pathlib import Path
root=Path(__file__).resolve().parent.parent
out=Path(sys.argv[2]).resolve() if len(sys.argv)>2 else root/'analysis/pf8-runtime-validation';out.mkdir(exist_ok=True)
binary=Path(sys.argv[1]).resolve() if len(sys.argv)>1 else root/'bin/pinballfantasies'
helper=["python3",str(root/'tools/parity_pf8.py')]
def windows():
 return subprocess.check_output(helper+['list'],text=True).splitlines()
before=set(windows())
with tempfile.TemporaryDirectory(prefix='pf8-audio-scores-') as scores:
 with open(out/'selector-pulse.log','w') as log,open(out/'selector-monitor.log','w') as capturelog:
  rec=subprocess.Popen(['ffmpeg','-nostdin','-hide_banner','-y','-f','pulse','-sample_rate','48000','-channels','2','-i','@DEFAULT_MONITOR@','-t','24',str(out/'selector-pulse.wav')],stdout=capturelog,stderr=subprocess.STDOUT)
  env=dict(os.environ,SDL_VIDEODRIVER='x11',SDL_AUDIODRIVER='pulseaudio')
  game=subprocess.Popen([str(binary),'-data-dir',str(root),'-high-score-dir',scores,'-duration','22s'],stdout=log,stderr=subprocess.STDOUT,env=env)
  timeline=[];start=time.monotonic();w=None
  try:
   for _ in range(80):
    ids=[int(l.split()[0]) for l in set(windows())-before if 'Pinball Fantasies' in l and 'pid='+str(game.pid)+' ' in l]
    if ids:w=max(ids);break
    time.sleep(.05)
   assert w,'own native window unavailable'
   time.sleep(.5)
   for i in range(13):
    timeline.append({'edge':'space','wall_seconds':time.monotonic()-start})
    subprocess.run(helper+['key',str(w),'space','.035'],check=True)
    time.sleep(1.05)
   game.wait(timeout=12);rec.wait(timeout=12)
   assert game.returncode==0 and rec.returncode==0
  finally:
   if game.poll() is None:game.terminate();game.wait()
   if rec.poll() is None:rec.terminate();rec.wait()
 (out/'selector-audio-inputs.json').write_text(json.dumps(timeline,indent=2)+'\n')
log=(out/'selector-pulse.log').read_text()
assert '48000 Hz stereo signed 16-bit; driver: pulseaudio' in log,'real stereo PulseAudio was not opened'
with wave.open(str(out/'selector-pulse.wav')) as capture:
 assert (capture.getnchannels(),capture.getframerate(),capture.getsampwidth())==(2,48000,2)
 samples=array.array('h',capture.readframes(capture.getnframes()))
 if sys.byteorder!='little':samples.byteswap()
 left,right=samples[::2],samples[1::2]
 assert any(left) and any(right) and left!=right,'missing or duplicated stereo channel'
assert log.count('PF6: selector text')>=6 and log.count('PF6: selector (')>=6
if binary==root/'bin/pinballfantasies':
 assert 'lifecycle clears=0 starts=1' in log and 'queue resets=0 empty-queue observations=0' in log
print(out)
