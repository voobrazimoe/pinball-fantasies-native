#!/usr/bin/env python3
"""Bounded real-output fullscreen pacing audit (Linux SDL + optional Win32/Wine).

The no-present control retains logical frame generation and all source syncs,
input, mixer PCM and device submissions; only host framebuffer transfer is off.
No queue measurement is fed back into gameplay scheduling.
"""
import argparse,json,os,statistics,subprocess,tempfile
from pathlib import Path

root=Path(__file__).resolve().parent.parent
parser=argparse.ArgumentParser()
parser.add_argument('--label',default='after')
parser.add_argument('--backend',choices=['linux','wine','both'],default='both')
parser.add_argument('--duration',type=int,default=12)
args=parser.parse_args()
out=root/'analysis/pf12-validation/pacing';out.mkdir(parents=True,exist_ok=True)
summary=[]
for backend in (['linux','wine'] if args.backend=='both' else [args.backend]):
 for no_present in [False,True]:
  name=f'{backend}-{args.label}-'+('no-present' if no_present else 'present')
  trace=out/(name+'.jsonl')
  env=dict(os.environ,PF12_AUDIT_FULLSCREEN='1',PF12_NO_PRESENT='1' if no_present else '0')
  with tempfile.TemporaryDirectory(prefix='pf12-pacing-') as storage:
   if backend=='linux':
    env.update(SDL_VIDEODRIVER='x11',SDL_AUDIODRIVER='pulseaudio',PF12_PACING_LOG=str(trace))
    cmd=[str(root/'bin/pinballfantasies'),'-data-dir',str(root),'-config-dir',storage,'-high-score-dir',storage,'-duration',f'{args.duration}s']
   else:
    winpath=lambda p:'Z:'+str(p).replace('/','\\')
    env.update(WINEPREFIX=os.environ.get('PF12_WINEPREFIX','/tmp/pf12-wine'),WINEDEBUG='-all',PF12_PACING_LOG=winpath(trace))
    cmd=['wine64',str(root/'bin/pinballfantasies-debug.exe'),'-data-dir',winpath(root),'-config-dir',winpath(storage),'-high-score-dir',winpath(storage),'-duration',f'{args.duration}s']
   with open(out/(name+'.log'),'w') as log:
    subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=args.duration+15,check=True)
  rows=[json.loads(line) for line in trace.read_text().splitlines()]
  stages={}
  for stage in ['update','frame','present']:
   group=[r for r in rows if r['stage']==stage]
   values=sorted(r['Duration']/1e6 for r in group)
   if values:stages[stage]={'count':len(values),'median_ms':statistics.median(values),'p95_ms':values[int(len(values)*.95)],'max_ms':max(values)}
  steady=[r['QueuedBytes'] for r in rows if r['stage']=='update' and r['Elapsed']>1_000_000_000]
  item={'name':name,'stages':stages,'queue_bytes_min':min(steady),'queue_bytes_median':statistics.median(steady),'queue_bytes_max':max(steady),'empty_observations':max(r['Empty'] for r in rows),'queue_resets':max(r['Resets'] for r in rows)}
  summary.append(item);print(json.dumps(item),flush=True)
(out/(args.label+'-summary.json')).write_text(json.dumps(summary,indent=2)+'\n')
