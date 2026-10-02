#!/usr/bin/env python3
"""Correlate real Pulse monitor against continuous native PCM, without SciPy.
Run with a Python providing numpy. Gain and fixed host latency are measured,
not used to alter either stream. Reports alignment drift around visual edges.
"""
import numpy as np,wave,json,sys
from pathlib import Path
root=Path(__file__).resolve().parent.parent;out=Path(sys.argv[1]).resolve() if len(sys.argv)>1 else root/'analysis/pf8-runtime-validation'
with wave.open(str(out/'selector-pulse.wav')) as f:
 assert (f.getnchannels(),f.getframerate(),f.getsampwidth())==(2,48000,2)
 real=np.frombuffer(f.readframes(f.getnframes()),dtype='<i2').reshape(-1,2).astype(float)
expected=np.fromfile(out/'selector-expected.pcm',dtype='<i2').reshape(-1,2).astype(float)
def corr(a,b):
 n=1<<((len(a)+len(b)-2).bit_length())
 conv=np.fft.irfft(np.fft.rfft(a,n)*np.fft.rfft(b[::-1],n),n)
 return conv[len(b)-1:len(a)]
# Locate each expected one-second block independently in the recording.
results=[]
for sec in range(1,21):
 start=sec*48000;target=expected[start:start+48000]
 c=sum(corr(real[:,ch],target[:,ch]) for ch in range(2));sq=np.r_[0,np.cumsum(np.sum(real*real,axis=1))];energy=sq[len(target):]-sq[:-len(target)]
 norm=np.sqrt(energy*np.sum(target*target));normalized=c/np.maximum(norm,1)
 pos=int(normalized.argmax());gain=float(np.sum(real[pos:pos+len(target)]*target)/np.sum(target*target))
 results.append({'expected_second':sec,'recorded_sample':pos,'lag_samples':pos-start,'correlation':float(normalized[pos]),'gain':gain})
lag=results[0]['lag_samples']
start,end=48000,21*48000
exact=bool(np.array_equal(real[start+lag:end+lag],expected[start:end]))
report={'exact_20_second_pcm':exact,'channels':2,'compared_frames':end-start,'recorded_frames':len(real),'sample_rate':48000,'alignments':results,'lag_range_samples':max(r['lag_samples'] for r in results)-min(r['lag_samples'] for r in results),'min_correlation':min(r['correlation'] for r in results)}
(out/'selector-pcm-continuity.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps(report,indent=2))
