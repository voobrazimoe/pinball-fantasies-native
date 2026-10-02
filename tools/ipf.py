#!/usr/bin/env python3
"""Minimal SPS/CAPS IPF reader: records, image descriptors, block descriptors,
data/gap stream elements. Enough to reconstruct raw MFM tracks."""
import struct, sys

class IPF:
    def __init__(self, path):
        self.d = open(path, 'rb').read()
        assert self.d[:4] == b'CAPS', 'not an IPF'
        self.records = []
        off = 0
        # file header: 'CAPS' + u32 + u32 crc  = 12 bytes (per libdisk ipf_header)
        off = 12
        while off + 12 <= len(self.d):
            rid = self.d[off:off+4]
            if rid not in (b'INFO', b'IMGE', b'DATA', b'CTEI', b'CTEX', b'DUMP', b'TRCK'):
                break
            rlen, rcrc = struct.unpack('>II', self.d[off+4:off+12])
            payload = self.d[off+12:off+rlen]
            skip = rlen
            if rid == b'DATA':
                # DATA record is header-only (28 B); its data area follows it.
                dasize = (struct.unpack('>I', payload[4:8])[0] + 7)//8
                payload = self.d[off+12:off+12+16+dasize]
                skip = rlen + dasize
            self.records.append((rid.decode(), off, rlen, rcrc, payload))
            off += skip
        self.info = None
        self.images = []
        self.datas = {}
        for rid, o, l, c, p in self.records:
            if rid == 'INFO':
                f = struct.unpack('>20I', p[:80])
                self.info = dict(type=f[0], encoder=f[1], encrev=f[2], release=f[3],
                                 revision=f[4], origin=f[5], mincyl=f[6], maxcyl=f[7],
                                 minhead=f[8], maxhead=f[9], date=f[10], time=f[11],
                                 platform=f[12:16], disknum=f[16], userid=f[17])
            elif rid == 'IMGE':
                f = struct.unpack('>17I', p[:68])
                img = dict(cyl=f[0], head=f[1], dentype=f[2], sigtype=f[3], trksize=f[4],
                           startpos=f[5], startbit=f[6], databits=f[7], gapbits=f[8],
                           trkbits=f[9], blkcnt=f[10], process=f[11], flags=f[12],
                           dat_chunk=f[13], offset=o, payload=p)
                self.images.append(img)
            elif rid == 'DATA':
                size, bsize, dcrc, chunk = struct.unpack('>4I', p[:16])
                self.datas[chunk] = dict(size=size, bsize=bsize, dcrc=dcrc,
                                         chunk=chunk, payload=p, offset=o)

    def blocks(self, img):
        """Yield block descriptors for an image (32 bytes each)."""
        p = self.datas[img['dat_chunk']]['payload']
        sps = self.info['encoder'] == 2
        out = []
        for i in range(img['blkcnt']):
            o = 16 + i*32
            f = struct.unpack('>8I', p[o:o+32])
            b = dict(dataBits=f[0], gapBits=f[1])
            if sps:
                b['gapOffset'], b['cellType'] = f[2], f[3]
            else:
                b['dataBytes'], b['gapBytes'] = f[2], f[3]
            b['encoderType'], b['flags'], b['gapDefault'], b['dataOffset'] = f[4], f[5], f[6], f[7]
            out.append(b)
        return out

    def data_chunks(self, img, blk):
        """Parse the data stream element list -> list of (type, sample_bytes, bitlen)."""
        p = self.datas[img['dat_chunk']]['payload']
        o = blk['dataOffset']
        inbits = bool(blk['flags'] & 4)
        elems = []
        while o < len(p):
            head = p[o]; o += 1
            if head == 0: break
            w = head >> 5; t = head & 0x1f
            size = int.from_bytes(p[o:o+w], 'big') if w else 0
            o += w
            nbytes = size if not inbits else (size+7)//8
            sample = p[o:o+nbytes]; o += nbytes
            elems.append((t, sample, size, inbits))
        return elems

if __name__ == '__main__':
    ip = IPF(sys.argv[1])
    print('INFO:', ip.info)
    print(f'{len(ip.records)} records, {len(ip.images)} images, {len(ip.datas)} data records')
    from collections import Counter
    print('record types:', Counter(r[0] for r in ip.records))
    for img in ip.images[:4]:
        print(f"  IMGE cyl={img['cyl']} head={img['head']} blkcnt={img['blkcnt']} "
              f"databits={img['databits']} gapbits={img['gapbits']} trkbits={img['trkbits']} "
              f"dat_chunk={img['dat_chunk']} startbit={img['startbit']}")
        for i, b in enumerate(ip.blocks(img)[:3]):
            print(f'     blk{i}: dataBits={b["dataBits"]} gapBits={b["gapBits"]} '
                  f'enc={b["encoderType"]} flags={b["flags"]} dataOff={b["dataOffset"]}')
            for t, s, size, ib in ip.data_chunks(img, b)[:8]:
                print(f'        elem type={t} size={size} inbits={ib} sample={s[:16].hex()}')
