#!/usr/bin/env python3
import sys, os, collections, glob
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from ipf import IPF
from ipf2adf import block_bytes

for p in sys.argv[1:]:
    ip = IPF(p)
    c = collections.Counter(i['blkcnt'] for i in ip.images)
    img = ip.images[0]
    s = block_bytes(ip, img, ip.blocks(img)[0])
    print(os.path.basename(p), '| images', len(ip.images), '| blkcnt', dict(c))
    print('   first track blk0 len', len(s), 'head', s[:24].hex())
    for im in ip.images:
        if im['blkcnt'] == 11:
            b = block_bytes(ip, im, ip.blocks(im)[0])
            print('   11-block track cyl%d.%d len=%d head=%s' % (im['cyl'], im['head'], len(b), b[:24].hex()))
            break
