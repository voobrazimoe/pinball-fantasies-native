#!/usr/bin/env python3
"""ProTracker M.K. parser: sequential sample layout, no pointer field."""
import struct

def parse_mod(path):
    d = open(path, 'rb').read()
    tag = d[1080:1084]
    pats = max(struct.unpack('>128B', d[952:1080])) + 1
    sdata = 1084 + pats * 1024
    off = sdata
    out = []
    for i in range(31):
        o = 20 + i * 30
        nm = d[o:o+22].split(b'\0')[0].decode('latin-1')
        ln = struct.unpack('>H', d[o+22:o+24])[0] * 2
        ft = d[o+24] & 0x0F
        vol = d[o+25]
        ls = struct.unpack('>H', d[o+26:o+28])[0] * 2
        ll = struct.unpack('>H', d[o+28:o+30])[0] * 2
        blob = d[off:off+ln] if ln else b''
        out.append(dict(n=i+1, name=nm, ln=ln, off=off, finetune=ft, vol=vol,
                        loop=(ls, ll), blob=blob))
        off += ln + (ln & 1)
    return dict(data=d, tag=tag, patterns=pats, sample_data_start=sdata,
                samples=out, ptable=d[952:1080])
