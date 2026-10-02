#!/usr/bin/env python3
"""Named temporal content comparisons. No timing inferred by arbitrary resizing."""
import json,hashlib,collections
from pathlib import Path
from PIL import Image
root=Path(__file__).resolve().parent.parent;p=root/'analysis/pf8-runtime-validation';n=p/'native'
def h(im):return hashlib.sha256(im.convert('RGBA').tobytes()).hexdigest()
frames=json.loads((n/'states.json').read_text());known=collections.defaultdict(list)
for e in frames:
 im=Image.open(n/(e['name']+'.png'));known[im.size,h(im)].append(e['name'])
original=p/'normal-isolated';timeline=json.loads((original/'timeline.json').read_text());results=[]
for e in timeline['events']:
 if e['kind']!='frame' or not e['file'].startswith('startup'):continue
 im=Image.open(original/e['file']);names=known.get((im.size,h(im)),[])
 if not names and im.size==(640,480):
  small=im.resize((320,240),Image.Resampling.NEAREST)
  if small.resize(im.size,Image.Resampling.NEAREST).tobytes()==im.tobytes():names=known.get((small.size,h(small)),[])
 results.append({'file':e['file'],'seconds':e['seconds'],'native_matches':names})
traces=json.loads((n/'attract-states.json').read_text());matrices=[]
for folder,table,prefix in [('normal-isolated',1,'party-initial-attract'),('speed-attract-isolated',2,'attract-speed')]:
 known=collections.defaultdict(list)
 for e in traces:
  if e['table']==table:known[e['rgba']].append(e['tick'])
 path=p/folder
 if not (path/'timeline.json').exists():continue
 tl=json.loads((path/'timeline.json').read_text());matches=[]
 for e in tl['events']:
  if e['kind']!='frame' or not e['file'].startswith(prefix):continue
  im=Image.open(path/e['file']);small=im.resize((320,33),Image.Resampling.NEAREST)
  assert small.resize(im.size,Image.Resampling.NEAREST).tobytes()==im.tobytes(),'non-identical X samples'
  ticks=known.get(h(small),[])
  matches.append({'file':e['file'],'seconds':e['seconds'],'native_ticks':ticks})
 matrices.append({'folder':folder,'table':table,'frames':matches})
text_original=Image.open(original/'text-0000.png').convert('RGB')
text_logical=text_original.resize((640,240),Image.Resampling.NEAREST)
assert text_logical.resize(text_original.size,Image.Resampling.NEAREST).tobytes()==text_original.tobytes()
region=(128,0,640,240)
text_hash=h(text_logical.crop(region))
text_matches=[]
for f in sorted(n.glob('text-*.png')):
 if h(Image.open(f).crop(region))==text_hash:text_matches.append(f.name)
out={'startup':results,'attract':matrices,'selector_text':{'oracle':'normal-isolated/text-0000.png','logical_region':list(region),'pixels':512*240,'rgba_sha256':text_hash,'native_matches':text_matches,'scope':'Content/palette match at fade level16/20. Sidebar and absolute callback alignment excluded.'}}
(p/'temporal-comparisons.json').write_text(json.dumps(out,indent=2)+'\n')
print('startup captures',len(results),'matching',sum(bool(x['native_matches']) for x in results),'states',sorted(set(k for x in results for k in x['native_matches'])))
for m in matrices:print('attract table',m['table'],'matched',sum(bool(x['native_ticks']) for x in m['frames']),'of',len(m['frames']))
print('selector text exact region',text_matches)
