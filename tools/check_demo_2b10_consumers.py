"""Bounded CHECKHIGHSCORE operands and native factory seed admission, no DOS IO."""
import hashlib
import json
from pathlib import Path
from check_demo_2b9_consumers import check as predecessor
from audit_10min_demo_graph import Decoder
from audit_10min_demo import require

# File offsets, CS-relative near destinations. No executable bytes stored.
OPS = (
 (0x56ac,'cmp','byte ptr cs:[0x3476], 0'),
 (0x56b2,'jne','0x53b7'), (0x56b4,'jmp','0x5452'),
 (0x56b7,'cmp','word ptr [0x37f5], 0x2d0'), (0x56bd,'je','0x5406'),
 (0x56c2,'cmp','byte ptr [0x382b], 0xff'), (0x56c7,'je','0x53f8'),
 (0x56cc,'cmp','byte ptr [0x3812], 0xff'), (0x56d1,'je','0x542a'),
 (0x56d6,'cmp','byte ptr [0x34df], 0xff'), (0x56db,'je','0x543e'),
 (0x56e0,'cmp','byte ptr [0x34e0], 0xff'), (0x56e5,'jne','0x543e'),
 (0x56ea,'mov','byte ptr [0x34e0], 0'),
 (0x56ef,'mov','bx, 0x1ade'), (0x56f2,'call','0x4514'), (0x56f5,'jmp','0x543e'),
 (0x573e,'call','0x5456'), (0x5741,'inc','word ptr [0x37f5]'),
 (0x5745,'mov','si, 0x46b5'), (0x5748,'mov','bx, 0xc8'),
 (0x574b,'push','cs'), (0x574c,'pop','es'), (0x574d,'lcall','es:[0x53a8]'),
 (0x5752,'mov','si, 0'), (0x5755,'ret',''),
 (0x5756,'pushaw',''),
 (0x5757,'cmp','byte ptr [0x34df], 0xff'), (0x575c,'je','0x5492'),
 (0x5761,'cmp','byte ptr [0x34e1], 0xff'), (0x5766,'je','0x5492'),
 (0x576b,'cmp','byte ptr [0x34f0], 0xff'), (0x5770,'je','0x5492'),
 (0x5775,'mov','di, 0x16'), (0x5778,'mov','si, 0x46b5'),
 (0x577b,'mov','cx, 0xc'), (0x577e,'mov','bx, 0'),
 (0x5781,'mov','al, byte ptr [bx + di]'), (0x5783,'cmp','byte ptr [bx + si], al'),
 (0x5785,'jb','0x5492'), (0x578a,'ja','0x5494'),
 (0x578f,'inc','bx'), (0x5790,'loop','0x5481'),
 (0x5792,'popaw',''), (0x5793,'ret',''),
 (0x5794,'mov','byte ptr [0x34f0], 0xff'), (0x5799,'mov','bx, 0x1413'),
 (0x579c,'call','0x4501'), (0x579f,'mov','byte ptr [0x34e0], 0xff'),
 (0x57a4,'popaw',''), (0x57a5,'ret',''),
)

def check(a, b):
    predecessor(a, b)
    d = Decoder()
    for at, op, args in OPS:
        d.expect(b,768,at,op,args)
    # Reject modified padding as well: each gap in this bounded path is NOP.
    for left, right in ((0x5756,0x57a6),(0x573e,0x5756)):
        cursor = left
        for at, op, args in OPS:
            if not left <= at < right: continue
            require(all(v == 0x90 for v in b[cursor:at]), 'branch padding drift')
            cursor = at + d.instruction(b,768,at).size
        require(cursor == right, 'bounded end')
    # Same native Defaults(1) values; inventory identity proves the factory seed.
    seed = b''.join(bytes(map(int,f'{v:012d}'))+name+b'\0' for v,name in
        zip((50000000,25000000,10000000,5000000),(b'TSP',b'ICE',b'ANY',b'J L')))
    inventory = json.loads((Path(__file__).resolve().parents[1]/'analysis/game-inventory.json').read_text())
    want = next(r['sha256'] for r in inventory if r['name']=='TABLE1.HI')
    require(hashlib.sha256(seed).hexdigest() == want, 'native factory seed identity')
    require(b[0x19db0+0x16:0x19db0+0x16+64] == seed, 'linked compiled factory seed')
    require(int.from_bytes(b[0x19db0+0x1413:0x19db0+0x1415],"little") == 0x2f83, "BEATENTS _DOBEATEN producer")
    d.expect(b,768,0x328f,"inc","byte ptr [0xce]")
    return True

if __name__ == '__main__':
    import os
    check((Path(__file__).resolve().parents[1]/'TABLE1.PRG').read_bytes(),
          (Path(os.environ['PF_10MIN_DEMO_DATA'])/'TABLE1.PRG').read_bytes())
