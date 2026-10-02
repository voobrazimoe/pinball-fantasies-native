#!/usr/bin/env python3
import sys, os, glob, struct
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from mod_samples import parse_mod

files = [(p, open(p,'rb').read()) for p in sys.argv[1:]]
for m in sorted(glob.glob('*.MOD')):
    mod = parse_mod(m); d = mod['data']; pats = mod['patterns']
    # 1. pattern table (order list)
    pt = d[952:1080]
    pth = [p for p, b in files if b.find(pt) >= 0]
    # 2. whole pattern block
    block = d[1084:1084+pats*1024]
    bh = [p for p, b in files if b.find(block[:4096]) >= 0]
    # 3. individual patterns
    indiv = []
    for i in range(min(pats, 8)):
        pp = d[1084+i*1024:1084+(i+1)*1024]
        if any(pp.count(0) < 1000 for _ in [0]):
            h = [p for p, b in files if b.find(pp) >= 0]
            if h: indiv.append((i, h[0]))
    print(f'== {m} pats={pats}')
    print(f'   order-table(128B) found in: {pth or "NO"}')
    print(f'   pattern-block first 4K found in: {bh or "NO"}')
    print(f'   individual patterns 0..7 found: {indiv or "none"}')
    # 4. title string
    title = d[:20].rstrip(b'\0')
    th = [p for p, b in files if b.find(title) >= 0]
    print(f'   title {title!r} found in: {th or "NO"}')
