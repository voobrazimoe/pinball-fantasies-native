#!/usr/bin/env python3
"""Decode all tracks of an IPF and count hits of every PC module instrument."""
import sys, os, struct
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from ipf import IPF
from ipf2adf import block_dat
from mod_samples import parse_mod

MODS = ['TABLE1.MOD', 'TABLE2.MOD', 'TABLE3.MOD', 'TABLE4.MOD', 'INTRO.MOD', 'MOD2.MOD']

def decode(ipf):
    ip = IPF(ipf)
    out = bytearray()
    for img in ip.images:
        for blk in ip.blocks(img):
            out += block_dat(ip, img, blk)
    return bytes(out), ip

def main(ipf):
    data, ip = decode(ipf)
    print(f'{os.path.basename(ipf)}: {len(data)} decoded bytes, {len(ip.images)} tracks')
    for m in MODS:
        mod = parse_mod(m)
        tot = hit = 0
        for s in mod['samples']:
            if s['ln'] < 64: continue
            tot += 1
            b = s['blob']
            probe = b[len(b)//3: len(b)//3+256]
            if data.find(probe) >= 0: hit += 1
        order = data.find(bytes(mod['data'][952:952+64]))
        periods = b''.join(struct.pack('>H', p) for p in
            [856,808,762,720,678,640,604,570,538,508,480,453,428,404,381,360,339,320,302,285,269,254,240,226,214,202,190,180,170,160,151,143,135,127,120,113])
        print(f'   {m:<11} instruments found {hit:>2}/{tot:<2}  order-list {"YES" if order>=0 else "no"}  period-table x{data.count(periods)}')

if __name__ == '__main__':
    for f in sys.argv[1:]:
        main(f)
