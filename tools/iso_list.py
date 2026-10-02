#!/usr/bin/env python3
"""Minimal ISO9660 (MODE1/2048) directory lister + file extractor."""
import sys, struct

class ISO:
    def __init__(self, path):
        self.f = open(path, 'rb')
        self.f.seek(16 * 2048)
        pvd = self.f.read(2048)
        assert pvd[0] == 1 and pvd[1:6] == b'CD001', 'not ISO9660'
        self.root_extent = struct.unpack('<I', pvd[156+2:156+6])[0]
        self.root_size = struct.unpack('<I', pvd[166:170])[0]
        self.logical_block = struct.unpack('<H', pvd[128:130])[0]
        self.vol_id = pvd[40:72].decode('latin-1').strip()

    def read_sector(self, lba, n=1):
        self.f.seek(lba * self.logical_block)
        return self.f.read(n * self.logical_block)

    def walk(self, extent, size, prefix=''):
        data = self.read_sector(extent, (size + self.logical_block - 1)//self.logical_block)
        off = 0
        entries = []
        while off < size:
            ln = data[off]
            if ln == 0:
                off = (off // self.logical_block + 1) * self.logical_block
                if off >= size: break
                continue
            rec = data[off:off+ln]
            ext = struct.unpack('<I', rec[2:6])[0]
            dsize = struct.unpack('<I', rec[10:14])[0]
            flags = rec[25]
            nlen = rec[32]
            name = rec[33:33+nlen]
            off += ln
            if name in (b'\x00', b'\x01'): continue
            name = name.split(b';')[0].decode('latin-1')
            full = prefix + '/' + name
            if flags & 0x02:
                yield ('DIR', full, ext, dsize)
                if prefix.count('/') < 8:
                    yield from self.walk(ext, dsize, full)
            else:
                yield ('FILE', full, ext, dsize)

    def extract(self, extent, size):
        return self.read_sector(extent, (size + self.logical_block - 1)//self.logical_block)[:size]

if __name__ == '__main__':
    iso = ISO(sys.argv[1])
    print('# volume id:', repr(iso.vol_id), 'block size', iso.logical_block)
    if len(sys.argv) > 2:
        want = sys.argv[2].upper()
        for kind, path, ext, size in iso.walk(iso.root_extent, iso.root_size):
            if kind == 'FILE' and want in path.upper():
                out = sys.argv[3] if len(sys.argv) > 3 else path.lstrip('/').replace('/', '_')
                open(out, 'wb').write(iso.extract(ext, size))
                print('extracted', path, size, '->', out)
    else:
        for kind, path, ext, size in iso.walk(iso.root_extent, iso.root_size):
            print(f'{kind}\t{size:>10}\t{path}')
