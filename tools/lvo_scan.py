#!/usr/bin/env python3
"""Find jsr d16(a6) library calls (opcode 4EAE) and histogram LVOs."""
import sys, os, collections

LVO = {
 -186:'LoadRGB4(OCS)', -288:'SetRGB4(OCS)', -1152:'LoadRGB32(AGA)', -1158:'SetRGB32(AGA)',
 -918:'AllocBitMap(AGA/2.0)', -216:'AllocRaster(OCS)', -210:'FreeRaster',
 -222:'InitBitMap/…', -240:'InitView', -270:'LoadView', -306:'OwnBlitter',
 -294:'WaitBlit', -282:'SetAPen', -276:'SetBPen', -252:'SetDrMd', -246:'SetRast',
 -180:'Draw/…', -96:'FreeChipRaster', -120:'GetGBuffers',
}

def scan(name, b):
    hist = collections.Counter()
    n = len(b) - 4
    for i in range(n):
        if b[i] == 0x4E and b[i+1] == 0xAE:
            off = int.from_bytes(b[i+2:i+4], 'big', signed=True)
            if -2048 <= off < 0:
                hist[off] += 1
    top = hist.most_common(18)
    print(f'== {name}')
    print('   LVO histogram:', {l: c for l, c in top})
    print('   named:', {LVO[l]: c for l, c in hist.items() if l in LVO})
    print('   LoadRGB4=%d  LoadRGB32=%d  AllocRaster=%d  AllocBitMap=%d' % (
        hist.get(-186,0), hist.get(-1152,0), hist.get(-216,0), hist.get(-918,0)))

if __name__ == '__main__':
    for f in sys.argv[1:]:
        scan(os.path.basename(f), open(f,'rb').read())
