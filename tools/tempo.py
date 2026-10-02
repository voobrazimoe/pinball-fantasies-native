#!/usr/bin/env python3
"""Estimate dominant beat period from the onset envelope (ffmpeg-decoded mono)."""
import subprocess, sys, array, math

RATE, HOP = 8000, 80  # 100 Hz

def env(path):
    raw = subprocess.run(['ffmpeg','-hide_banner','-v','error','-i',path,'-ac','1','-ar',str(RATE),'-f','s16le','-'],
                         capture_output=True, check=True).stdout
    a = array.array('h'); a.frombytes(raw[:len(raw)//2*2])
    return [math.sqrt(sum(v*v for v in a[i:i+HOP])/HOP) for i in range(0,len(a)-HOP,HOP)]

def onset(e):
    o=[max(0.0, e[i]-e[i-1]) for i in range(1,len(e))]
    m=sum(o)/len(o)
    return [x-m for x in o]

def tempo(path):
    o=onset(env(path)); n=len(o)
    res={}
    for lag in range(20, 220):   # 0.2s .. 2.2s
        s=0.0
        for i in range(lag, n):
            s+=o[i]*o[i-lag]
        res[lag]=s/(n-lag)
    top=sorted(res.items(), key=lambda kv:-kv[1])[:6]
    return top

for p in sys.argv[1:]:
    t=tempo(p)
    s=', '.join(f'{60.0*100/lag:.1f}BPM(lag{lag})' for lag,_ in t)
    print(f'{p.split("/")[-1][:52]:<54} {s}')
