#!/usr/bin/env python3
"""List Amiga hunk structure (HUNK_HEADER + hunks) of an AmigaOS executable."""
import sys, struct

NAMES = {0x3E9:'CODE',0x3EA:'DATA',0x3EB:'BSS',0x3EC:'RELOC32',0x3ED:'RELOC16',
         0x3EE:'RELOC8',0x3EF:'EXT',0x3F0:'SYMBOL',0x3F1:'DEBUG',0x3F2:'END',
         0x3F3:'HEADER',0x3F4:'OVERLAY',0x3F5:'LIB',0x3F7:'RELOC32SHORT',
         0x3F8:'RELOC24',0x3F9:'RELOC16B',0x3FA:'RELOC8B',0x3FB:'LIB32',
         0x3FC:'NAME',0x3FD:'HUNK_ARCH',0x3FE:'HUNK_DREL',0x3FF:'HUNK_UNIT'}

def parse(path, dump=None):
    d = open(path,'rb').read()
    off = 0
    def u32():
        nonlocal off
        v = struct.unpack('>I', d[off:off+4])[0]; off += 4; return v
    hdr = u32()
    assert hdr == 0x3F3, hex(hdr)
    nblocks = u32(); first = u32(); last = u32(); sizes = [u32() for _ in range(nblocks)]
    print(f'{path}: {nblocks} hunks, first={first} last={last}, header={off}B')
    names = {}
    longsz = any(s >= 0x40000000 for s in sizes)
    sizes = [s & 0x3FFFFFFF for s in sizes]
    total = 0
    for i in range(nblocks):
        if off >= len(d): break
        t = u32()
        tn = NAMES.get(t, hex(t))
        if t in (0x3E9,0x3EA,0x3EB):
            n = u32(); sz = n*4
            total += sz
            print(f'  hunk {first+i:>3} {tn:<6} {n:>7} longs  {sz:>9} B  @0x{off:08x}')
            off += sz
        elif t in (0x3EC,0x3F7):
            n = u32(); off += n*4
            print(f'  hunk {first+i:>3} {tn:<6} {n} relocs')
        elif t in (0x3ED,0x3EE,0x3F8,0x3F9,0x3FA):
            n = u32(); off += n*4
            print(f'  hunk {first+i:>3} {tn:<6} {n} relocs')
        elif t == 0x3F2:
            print(f'  hunk {first+i:>3} END')
            break
        elif t == 0x3F0:
            n = u32(); off += n*4
            print(f'  hunk {first+i:>3} SYMBOL {n} longs')
        elif t == 0x3FC:
            n = u32(); s = d[off:off+n]; off += n; off = (off+3)//4*4
            print(f'  hunk {first+i:>3} NAME {s!r}')
        else:
            print(f'  hunk {first+i:>3} {tn} unhandled @0x{off-4:08x}'); break
    print(f'  total code+data+bss = {total} B, file = {len(d)} B')

if __name__ == '__main__':
    for p in sys.argv[1:]:
        parse(p)
