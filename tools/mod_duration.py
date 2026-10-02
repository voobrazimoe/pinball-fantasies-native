#!/usr/bin/env python3
import sys, os, glob, struct
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from mod_samples import parse_mod

def period_to_note(p):
    return p

def analyse(path):
    mod = parse_mod(path); d = mod['data']
    songlen = d[950]
    orders = list(d[952:952+128])
    npat = mod['patterns']
    total_rows = 0
    t = 0.0
    bpm = 125.0; speed = 6
    used = set()
    for pi in range(songlen if songlen else 128):
        pat = orders[pi]
        if pat >= npat: continue
        used.add(pat)
        base = 1084 + pat*1024
        rows = 64
        for r in range(64):
            cell = d[base+r*4: base+r*4+4]
            total_rows += 1
            t += speed * (2.5/bpm)
            eff = cell[2] & 0x0F; par = cell[3]
            if eff == 0x0F:
                if par == 0: speed = 0 if False else 1
                elif par < 32: speed = par
                else: bpm = par
    return dict(songlen=songlen, npat=npat, total_rows=total_rows, seconds=t,
                patterns_used=len(used), orders=orders[:songlen])

for m in sorted(glob.glob('*.MOD')):
    a = analyse(m)
    mm, ss = divmod(int(a['seconds']), 60)
    print(f"{m:<12} songlen={a['songlen']:>3} patterns={a['npat']:>3} used={a['patterns_used']:>3} rows={a['total_rows']:>5} dur={mm}:{ss:02d}  orders={a['orders']}")
