#!/usr/bin/env python3
import sys, os, re
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from ipf import IPF
from ipf2adf import block_dat

for ipf in sys.argv[1:]:
    ip = IPF(ipf)
    data = bytearray()
    for img in ip.images:
        for blk in ip.blocks(img):
            data += block_dat(ip, img, blk)
    runs = re.findall(rb'[ -~]{8,}', bytes(data))
    runs = [r for r in runs if not re.fullmatch(rb'[0-9A-Fa-f ]+', r)]
    print(f'== {os.path.basename(ipf)}: {len(data)} B, {len(runs)} text runs')
    for r in runs[:14]:
        print('    ', r[:70])
