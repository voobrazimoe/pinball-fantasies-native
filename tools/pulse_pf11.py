#!/usr/bin/env python3
"""Requested real PulseAudio validation, using only isolated native stores.
Records the monitor at48kHz/s16/stereo and executes the own-window X11 smoke.
"""
import array,json,os,subprocess,sys,time,wave
from pathlib import Path
root=Path(__file__).resolve().parent.parent
out=root/'analysis/pf11-validation';out.mkdir(exist_ok=True)
mode=sys.argv[1] if len(sys.argv)>1 else 'on'
env=dict(os.environ,PF11_AUDIO_DRIVER='pulseaudio')
if mode=='off':env['PF11_CONFIG']='010102000100' # menu toggles music OFF
script=root/'tools/smoke_pf11.py'
if mode=='options':
 source=script.read_text();a=source.index('  for table in range(1,5):');b=source.index("  press('Escape','PF6: quit (table 4)')",a)
 source=source[:a]+'''  for _ in range(4):
   press('F5','PF6: options (table 1)');time.sleep(.9)
   press('Escape','PF6: selector (table 1)')
'''+source[b:];source=source.replace('PF6: quit (table 4)','PF6: quit (table 1)')
 script=out/'options-pulse-harness.py';script.write_text(source)
with open(out/f'pulse-{mode}-monitor.log','w') as monitor:
 rec=subprocess.Popen(['ffmpeg','-nostdin','-hide_banner','-y','-f','pulse','-sample_rate','48000','-channels','2','-i','@DEFAULT_MONITOR@','-t','45' if mode!='options' else '20',str(out/f'pulse-{mode}.wav')],stdout=monitor,stderr=subprocess.STDOUT)
 try:
  result=subprocess.run(['python3','-u',str(script)],cwd=root,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True,timeout=100)
  (out/f'pulse-{mode}-smoke.log').write_text(result.stdout);print(result.stdout,flush=True)
  assert result.returncode==0
  rec.wait(timeout=55)
  assert rec.returncode==0
 finally:
  if rec.poll() is None:rec.terminate();rec.wait()
log=result.stdout
assert '48000 Hz stereo signed 16-bit; driver: pulseaudio' in log
assert 'queue resets=0' in log
if mode=='options':assert 'lifecycle clears=0 starts=1' in log
with wave.open(str(out/f'pulse-{mode}.wav')) as f:
 assert (f.getnchannels(),f.getframerate(),f.getsampwidth())==(2,48000,2)
 values=array.array('h',f.readframes(f.getnframes()))
 if sys.byteorder!='little':values.byteswap()
 left,right=values[::2],values[1::2]
 assert any(left) and any(right) and left!=right
 metrics=dict(mode=mode,channels=2,rate=48000,frames=len(left),left_rms=(sum(v*v for v in left)/len(left))**.5,right_rms=(sum(v*v for v in right)/len(right))**.5,different_frames=sum(a!=b for a,b in zip(left,right)))
 (out/f'pulse-{mode}-metrics.json').write_text(json.dumps(metrics,indent=2)+'\n')
print('PF11 real PulseAudio',mode,'PASS',flush=True)
