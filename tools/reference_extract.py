#!/usr/bin/env python3
"""Independent PF1 fixture extraction; never imported by the Go program.
Fixed locations from static IFF chunk inspection of supplied TABLE1.PRG.
Produces hashes only by default, optional local PNG for visual inspection.
"""
import argparse, hashlib, json, struct, zlib
from pathlib import Path
OFFSETS = (336944, 366176, 399776, 437168)
def digest(b): return hashlib.sha256(b).hexdigest()
def extract(path, header_heights=None):
    data = Path(path).read_bytes()
    parts, records, palettes = [], [], []
    for offset in OFFSETS:
        assert data[offset:offset+4] == b'FORM'
        assert data[offset+8:offset+12] == b'PBM '
        end = offset + 8 + int.from_bytes(data[offset+4:offset+8], 'big')
        p, chunks = offset + 12, {}
        while p < end:
            tag = data[p:p+4]
            size = int.from_bytes(data[p+4:p+8], 'big')
            chunks[tag] = data[p+8:p+8+size]
            p += 8 + size + (size & 1)
        assert p == end
        w, h = struct.unpack_from('>HH', chunks[b'BMHD'])
        assert (w,h) == (320,(header_heights or {}).get(offset,144)) and chunks[b'BMHD'][8] == 8 and chunks[b'BMHD'][9] in (0,2) and chunks[b'BMHD'][10] == 1
        src, out, i = chunks[b'BODY'], bytearray(), 0
        # ByteRun1 control is signed: 0..127 literal, -1..-127 run, -128 no-op.
        for y in range(h):
            row = bytearray()
            while len(row) < w:
                n = struct.unpack_from('b', src, i)[0]; i += 1
                if n >= 0:
                    row.extend(src[i:i+n+1]); i += n+1
                elif n != -128:
                    row.extend(bytes([src[i]]) * (1-n)); i += 1
                assert len(row) <= w
            out.extend(row)
        assert i == len(src)
        palettes.append(chunks[b'CMAP'])
        parts.append(bytes(out[:320*144]))
        records.append(dict(offset=offset,form_size=end-offset,width=w,height=h,
                            body_size=len(src),indices_sha256=digest(out),
                            palette_sha256=digest(chunks[b'CMAP'])))
    indices = b''.join(parts)
    palette = palettes[-1] # INIT_GFX preserves last stage's palette in PALLE.
    rgba = b''.join(palette[c*3:c*3+3]+b'\xff' for c in indices)
    return dict(file_sha256=digest(data),strips=records,palettes_equal=all(p==palette for p in palettes),
                width=320,height=576,indices_sha256=digest(indices),rgba_sha256=digest(rgba)), rgba

def png(path, rgba):
    def chunk(tag, data): return struct.pack('>I',len(data))+tag+data+struct.pack('>I',zlib.crc32(tag+data))
    scan=b''.join(b'\0'+rgba[y*1280:(y+1)*1280] for y in range(576))
    Path(path).write_bytes(b'\x89PNG\r\n\x1a\n'+chunk(b'IHDR',struct.pack('>IIBBBBB',320,576,8,6,0,0,0))+chunk(b'IDAT',zlib.compress(scan))+chunk(b'IEND',b''))
if __name__ == '__main__':
    ap=argparse.ArgumentParser();ap.add_argument('data',nargs='?',default='TABLE1.PRG');ap.add_argument('--png');a=ap.parse_args()
    info, pixels=extract(a.data)
    print(json.dumps(info,indent=2))
    if a.png: png(a.png,pixels)
