#!/usr/bin/env python3
"""Decode the PC TABLE*.PRG FORM/PBM playfield strips (mirrors internal/assets)."""
import struct

def byterun1(src, width, height):
    out = bytearray(); p = 0
    for y in range(height):
        rowend = len(out) + width
        while len(out) < rowend:
            n = struct.unpack('b', src[p:p+1])[0]; p += 1
            if n == -128: continue
            if n >= 0:
                cnt = n + 1
                out += src[p:p+cnt]; p += cnt
            else:
                cnt = 1 - n
                out += src[p:p+1] * cnt; p += 1
    return bytes(out)

def decode_form(data, offset):
    if data[offset:offset+4] != b'FORM' or data[offset+8:offset+12] != b'PBM ':
        return None
    size = struct.unpack('>I', data[offset+4:offset+8])[0]
    end = offset + 8 + size
    hdr = cmap = body = None
    pos = offset + 12
    while pos < end:
        n = struct.unpack('>I', data[pos+4:pos+8])[0]
        nxt = pos + 8 + n + (n & 1)
        kind = data[pos:pos+4]; payload = data[pos+8:pos+8+n]
        if kind == b'BMHD': hdr = payload
        elif kind == b'CMAP': cmap = payload
        elif kind == b'BODY': body = payload
        pos = nxt
    w = struct.unpack('>H', hdr[0:2])[0]
    h = struct.unpack('>H', hdr[2:4])[0]
    return dict(w=w, h=h, depth=hdr[8], cmap=cmap,
                indices=byterun1(body, w, h), body=body)

def planar(indices, w, h, plane):
    """Extract one bitplane: list of bytearrays, 40 bytes per 320px row."""
    stride = (w + 15)//16*2
    out = bytearray()
    for y in range(h):
        row = indices[y*w:(y+1)*w]
        acc = 0; bits = 0
        for x in range(w):
            acc = (acc << 1) | ((row[x] >> plane) & 1)
            bits += 1
        while bits % 8:
            acc <<= 1; bits += 1
        out += acc.to_bytes(bits//8, 'big')
        while len(out) % stride:
            out.append(0)
    return bytes(out)
