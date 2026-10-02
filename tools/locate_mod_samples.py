#!/usr/bin/env python3
"""Search each instrument blob of the PC .MOD files inside another file set."""
import sys, os, glob
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from mod_samples import parse_mod

paths = sys.argv[1:]
files = [(p, open(p,'rb').read()) for p in paths]
for m in sorted(glob.glob('*.MOD')):
    mod = parse_mod(m)
    tot = 0; found = []
    for s in mod['samples']:
        if not s['ln']: continue
        tot += 1
        b = s['blob']
        probe = b[len(b)//3: len(b)//3 + 512]
        if len(probe) < 32: probe = b[:min(512, s['ln'])]
        hits = [p for p, d in files if d.find(probe) >= 0]
        if hits: found.append((s['n'], s['name'], s['ln'], hits))
    print(f'== {m}: {len(found)}/{tot} instruments located')
    for n, nm, ln, hits in found:
        print(f'    #{n:<2} {nm!r:<22} {ln:>6}B  in: {", ".join(hits[:3])}{" ..." if len(hits)>3 else ""}')
