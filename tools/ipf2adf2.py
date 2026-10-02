#!/usr/bin/env python3
"""Raw IPF -> ADF: take each block's Data samples as one decoded sector."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from ipf import IPF
from ipf2adf import block_bytes, ados_checksum

def convert(path, outpath):
    ip = IPF(path)
    tracks = {}
    for img in ip.images:
        secs = {}
        for blk in ip.blocks(img):
            raw = block_bytes(ip, img, blk)
            if len(raw) >= 540:
                s = raw[:540]
                secs[s[2]] = s[28:540] if len(secs) < 11 else s[28:540]
        if secs:
            tracks[(img['cyl'], img['head'])] = secs
    ncyl = max(c for c, h in tracks) + 1
    out = bytearray()
    for cyl in range(ncyl):
        for head in (0, 1):
            secs = tracks.get((cyl, head), {})
            for s in range(11):
                out += secs.get(s, b'\x00'*512)
    open(outpath, 'wb').write(out)
    return ncyl, len(tracks)

if __name__ == '__main__':
    ncyl, ntrk = convert(sys.argv[1], sys.argv[2])
    print(f'{os.path.basename(sys.argv[1])}: {ntrk} tracks -> {sys.argv[2]} ({os.path.getsize(sys.argv[2])} B)')
