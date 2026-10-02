#!/usr/bin/env python3
import sys, os, glob
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from mod_samples import parse_mod

MAP = {'TABLE1.MOD':'pinballf_amigacd32/pinfilea.dat',
       'TABLE2.MOD':'pinballf_amigacd32/pinfileb.dat',
       'TABLE3.MOD':'pinballf_amigacd32/pinfilec.dat',
       'TABLE4.MOD':'pinballf_amigacd32/pinfiled.dat',
       'INTRO.MOD':'pinballf_amigacd32/pinball1.dat',
       'MOD2.MOD':'pinballf_amigacd32/pinball4.dat'}

for m, target in MAP.items():
    mod = parse_mod(m); d = mod['data']
    t = open(target,'rb').read()
    first = mod['samples'][0]
    base = t.find(first['blob'][:256])
    print(f'== {m} -> {os.path.basename(target)}  size={len(t)}')
    print(f'   sample1 blob at 0x{base:06x}; MOD sample area starts at 0x{mod["sample_data_start"]:06x} (=header+patterns)')
    # verify contiguity of whole sample stream
    ok = True
    for s in mod['samples']:
        if not s['ln']: continue
        exp = base + (s['off'] - mod['sample_data_start'])
        got = t.find(s['blob'][:128], max(0, exp-64))
        if got != exp:
            print(f'   sample #{s["n"]} {s["name"]!r} expected 0x{exp:06x} found 0x{got:06x}  {"MISMATCH" if got>=0 else "MISSING"}')
            ok = False
    print(f'   contiguous verbatim sample stream: {"YES" if ok else "NO"}')
    # bytes before the sample area in the CD32 file: is it the pattern/header area?
    print(f'   CD32 region before samples: 0x000000..0x{base:06x} = {base} bytes')
    print(f'   MOD header+patterns region = {mod["sample_data_start"]} bytes')
    if base > 0:
        head = t[max(0,base-1084):base]
        print(f'   last 1084 bytes before samples (hex head): {head[:48].hex()}')
        print(f'   MOD header first 48 bytes            : {d[:48].hex()}')
