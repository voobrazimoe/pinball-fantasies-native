#!/usr/bin/env python3
"""Find palette-like tables: OCS 12-bit color words or AGA 24-bit colors."""
import sys, os, glob, struct
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from ipf import IPF
from ipf2adf import block_dat

def ocs_runs(b):
    """Runs of >=16 consecutive BE words with top nibble 0 (0x0RGB form)."""
    n = len(b)//2
    w = struct.unpack('>%dH' % n, b[:n*2])
    runs = []
    i = 0
    while i < n:
        if w[i] & 0xF000 == 0:
            j = i
            while j < n and (w[j] & 0xF000) == 0:
                j += 1
            if j - i >= 16:
                runs.append((i*2, j-i, len(set(w[i:j]))))
            i = j
        else:
            i += 1
    return runs

def aga_runs(b):
    """Runs of >=16 consecutive BE longs representing 24-bit colors."""
    n = len(b)//4
    v = struct.unpack('>%dI' % n, b[:n*4])
    runs = []
    i = 0
    while i < n:
        if (v[i] & 0xFF000000) == 0:
            j = i
            while j < n and (v[j] & 0xFF000000) == 0:
                j += 1
            if j - i >= 16:
                runs.append((i*4, j-i, len(set(v[i:j]))))
            i = j
        else:
            i += 1
    return runs

def report(name, data):
    o = sorted(ocs_runs(data), key=lambda r: -r[1])[:6]
    a = sorted(aga_runs(data), key=lambda r: -r[1])[:6]
    print(f'-- {name} ({len(data)} B)')
    print(f'   longest OCS-word runs (off, len, distinct): {o}')
    print(f'   longest AGA-long runs (off, len, distinct): {a}')

if __name__ == '__main__':
    for f in sys.argv[1:]:
        report(os.path.basename(f), open(f,'rb').read())
