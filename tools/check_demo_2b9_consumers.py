"""2B9 bounded source admission; reuses reviewed correspondence, no search."""
import struct
from check_demo_2b8_consumers import check as predecessor
from audit_10min_demo_graph import Decoder
from audit_10min_demo import require


def check(a, b):
    predecessor(a, b)  # includes complete TOUCHER/ENABLETOUCHER correspondence
    d = Decoder()
    require(struct.unpack_from('<5H', b, 0x1ab3d) == (130,196,146,204,0x1350), 'TOUCHER target binding')
    require(tuple(b[0x19db0+0xc79:0x19db0+0xc7d]) == (7,10,0,3), 'S_TOUCH2 sample/note/voice4')
    for at, op, args in (
        (0x1650,'cmp','byte ptr [0x94], 0xff'),
        (0x166e,'mov','byte ptr [0x94], 0xff'),
        (0x1673,'mov','dx, 0x139b'),
        (0x169b,'mov','dx, 0x14'),
        (0x169e,'mov','bx, 0x36db'),
        (0x16a1,'call','0x57c7'),
        (0x16aa,'mov','byte ptr [0x94], 0'),
        (0x16af,'jmp','0x57ba'),
        (0x5ac7,'cmp','word ptr [bx], dx'),
        (0x5ac9,'je','0x57d2'),
        (0x5ace,'inc','word ptr [bx]'),
        (0x5ad0,'clc',''), (0x5ad1,'ret',''),
        (0x5ad2,'mov','word ptr [bx], 0'),
        (0x5ad6,'stc',''), (0x5ad7,'ret',''),
    ): d.expect(b,768,at,op,args)
    return True


if __name__ == '__main__':
    import os
    from pathlib import Path
    # Identity admission is performed by the predecessor entry before this tool.
    check((Path(__file__).resolve().parents[1]/'TABLE1.PRG').read_bytes(),
          (Path(os.environ['PF_10MIN_DEMO_DATA'])/'TABLE1.PRG').read_bytes())
