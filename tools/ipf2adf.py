#!/usr/bin/env python3
"""SPS/CAPS IPF -> ADF for AmigaDOS tracks.

The IPF data stream stores each MFM-encoded 32-bit value in libdisk's
bc_mfm_even_odd written order: 16 odd-indexed bits (MSB first), then the 16
even-indexed bits. Undo that to recover the logical sector bytes.
"""
import sys, os, struct
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from ipf import IPF

SEC = 512
SECTORS = 11

def unshuffle(b):
    out = bytearray()
    for i in range(0, len(b) - 3, 4):
        w0 = (b[i] << 8) | b[i+1]
        w1 = (b[i+2] << 8) | b[i+3]
        x = 0
        for k in range(16):
            x |= ((w0 >> k) & 1) << (2*k + 1)
            x |= ((w1 >> k) & 1) << (2*k)
        out += x.to_bytes(4, 'big')
    return bytes(out)

def ados_checksum(b):
    csum = 0
    for i in range(0, len(b)//4*4, 4):
        csum ^= struct.unpack('>I', b[i:i+4])[0]
    csum ^= csum >> 1
    return csum & 0x55555555

def block_dat(ip, img, blk):
    p = ip.datas[img['dat_chunk']]['payload']
    o = 16 + blk['dataOffset']
    inbits = bool(blk['flags'] & 4)
    out = bytearray()
    while o < len(p):
        head = p[o]; o += 1
        if head == 0: break
        w = head >> 5; t = head & 0x1f
        size = int.from_bytes(p[o:o+w], 'big') if w else 0
        o += w
        nbytes = size if not inbits else (size + 7)//8
        sample = p[o:o+nbytes]; o += nbytes
        if t == 2:
            out += sample
    return unshuffle(bytes(out))

def track_sectors(ip, img):
    secs = {}
    for blk in ip.blocks(img):
        raw = block_dat(ip, img, blk)
        for off in range(0, len(raw) - 539, 540):
            s = raw[off:off+540]
            hc = struct.unpack('>I', s[20:24])[0]
            dc = struct.unpack('>I', s[24:28])[0]
            if ados_checksum(s[:20]) != hc or ados_checksum(s[28:540]) != dc:
                continue
            if s[0] != 0xff:
                continue
            secs[s[2]] = s[28:540]
    return secs

def convert(path, outpath):
    ip = IPF(path)
    tracks, bad = {}, []
    for img in ip.images:
        secs = track_sectors(ip, img)
        if secs:
            tracks[(img['cyl'], img['head'])] = secs
    ncyl = max((c for c, h in tracks), default=-1) + 1
    out = bytearray()
    for cyl in range(ncyl):
        for head in (0, 1):
            secs = tracks.get((cyl, head))
            if not secs or len(secs) != SECTORS:
                bad.append((cyl, head, len(secs) if secs else 0))
                out += b'\x00' * (SECTORS*SEC)
            else:
                for s in range(SECTORS):
                    out += secs[s]
    open(outpath, 'wb').write(out)
    return ncyl, len(tracks), bad

if __name__ == '__main__':
    ncyl, ntrk, bad = convert(sys.argv[1], sys.argv[2])
    print(f'{os.path.basename(sys.argv[1])}: {ntrk} tracks, {ncyl} cyl, {os.path.getsize(sys.argv[2])} B')
    if bad:
        print('   incomplete tracks:', len(bad), bad[:8])
