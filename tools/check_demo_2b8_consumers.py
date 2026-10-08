"""Bounded 2B8 admission using existing correspondences; no replay or search."""
import json
import struct
from pathlib import Path
from audit_10min_demo import require
from audit_10min_demo_graph import Decoder
from audit_10min_demo_deterministic_replay import compare_instructions
from audit_10min_demo_fresh_bygel_drain import compare_block


def check(a, b):
    d = Decoder()
    compare_instructions(d, a, b)
    for role in ('BYGEL1_unlit', 'BYGEL2_unlit'):
        compare_block(d, a, b, role)
    require(struct.unpack_from('<7H', b, 0x1b259) ==
            (0x4ebe, 0x2c47, 1, 0x4d3c, 1, 0, 0x4ebe), 'PARTY_OFFTS')
    for at, mnemonic, operand in (
        (0x2f47, 'mov', 'word ptr [0x34e6], 0x52d6'),
        (0x2f4d, 'mov', 'word ptr [0x34e8], 1'),
        (0x2f53, 'jmp', '0x5395'),
        (0x503c, 'push', '1'),
        (0x503e, 'pop', 'word ptr [0x34e8]'),
        (0x5042, 'mov', 'word ptr [0x34e6], 0x52d6'),
        (0x5048, 'call', '0x4d85'),
        (0x504b, 'jmp', '0x5395'),
        (0x2596, 'mov', 'byte ptr [0xd1], 0'),
    ):
        d.expect(b, 768, at, mnemonic, operand)
    require(struct.unpack_from('<8H', b, 0x1e2b3) ==
            (0x4ebe,0x4c67,0x2379,340,0x4c67,0x1e8b,1684,0), 'SHOWPLAYERSTS')
    # Original sound structures, channel=3 encodes the fourth voice.
    for offset, record in ((0xc61, (7,10,0,3)), (0xc65, (7,12,0,3)),
                           (0xc49, (25,22,0,3)), (0xc59, (30,18,0,3))):
        require(tuple(b[0x19db0+offset:0x19db0+offset+4]) == record, 'sound operands')
    require(tuple(b[0x19db0+0xc8e:0x19db0+0xc8e+3]) == (1,0,1), 'S_MAIN')
    # BYGEL3/4 call the same matrix-free DOEFFECT, followed by voice four.
    for start in (0x27c0, 0x27db):
        for delta, op, args in ((0,'mov','si, 0xba8'), (3,'call','0x5c14'),
            (6,'mov','bh, 0'), (8,'mov','cl, byte ptr [0xc65]'),
            (12,'mov','bl, byte ptr [0xc66]'), (16,'mov','dl, byte ptr [0xc68]'),
            (20,'inc','dl'), (22,'mov','al, 0x11'), (24,'int','0x66'), (26,'ret','')):
            d.expect(b,768,start+delta,op,args)
    record = struct.unpack_from('<HB12B12BH', b, 0x19db0+0xba8)
    require(record[:2] == (0,1) and record[-1] == 0, 'BYGELSETB no jingle/matrix')
    number = lambda digits: int(''.join(str(v) for v in digits))
    require(number(record[2:14]) == 10040 and number(record[14:26]) == 1000,
            'BYGELSETB arithmetic')
    for at, want in ((0x1abb7,(25,435,35,445,0x24c0)),
                     (0x1abc1,(263,435,273,445,0x24db))):
        require(struct.unpack_from('<5H',b,at) == want, 'BYGEL3/4 binding')
    return True


if __name__ == '__main__':
    import os
    import hashlib
    from audit_10min_demo import FILES, digest
    root = Path(os.environ['PF_10MIN_DEMO_DATA'])
    for name, (size, identity) in FILES.items():
        data = (root/name).read_bytes()
        require(len(data) == size and digest(data) == identity, 'demo identity')
    workspace = Path(__file__).resolve().parents[1]
    a = (workspace/'TABLE1.PRG').read_bytes()
    inventory = json.loads((workspace/'analysis/game-inventory.json').read_text())
    require(hashlib.sha256(a).hexdigest() == next(r['sha256'] for r in inventory if r['name']=='TABLE1.PRG'), 'canonical A identity')
    check(a, (root/'TABLE1.PRG').read_bytes())
