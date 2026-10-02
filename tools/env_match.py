#!/usr/bin/env python3
"""Cross-correlate RMS envelopes of two audio files (any format ffmpeg reads)."""
import subprocess, sys, array, math

RATE = 8000
HOP  = 160   # 50 Hz envelope

def env(path):
    raw = subprocess.run(['ffmpeg','-hide_banner','-v','error','-i',path,
                          '-ac','1','-ar',str(RATE),'-f','s16le','-'],
                         capture_output=True, check=True).stdout
    a = array.array('h'); a.frombytes(raw[:len(raw)//2*2])
    out = []
    for i in range(0, len(a)-HOP, HOP):
        s = 0
        for v in a[i:i+HOP]:
            s += v*v
        out.append(math.sqrt(s/HOP))
    return out

def norm(e):
    m = sum(e)/len(e)
    d = math.sqrt(sum((x-m)**2 for x in e)/len(e)) or 1.0
    return [(x-m)/d for x in e]

def best_lag(a, b, maxlag):
    # correlate a against b, normalize by overlap length
    res = []
    n = len(b)
    for lag in range(0, maxlag):
        s = 0.0
        for i in range(lag, min(len(a), n)):
            s += a[i]*b[i-lag]
        res.append(s/(min(len(a),n)-lag))
    return res

if __name__ == '__main__':
    a = norm(env(sys.argv[1]))
    b = norm(env(sys.argv[2]))
    maxlag = int(sys.argv[3]) if len(sys.argv) > 3 else 50*120
    r = best_lag(a, b, maxlag)
    top = sorted(range(len(r)), key=lambda i: -r[i])[:5]
    print(f'{sys.argv[1]} vs {sys.argv[2]}')
    print(f'  lengths: {len(a)} : {len(b)} frames @50Hz')
    for i in top:
        print(f'  lag {i:>5} ({i/50:7.2f}s)  r={r[i]:+.4f}')
