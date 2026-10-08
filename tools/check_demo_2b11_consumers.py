"""2B11 bounded scored drain and actual zero-aggregate route admission."""
import struct
from pathlib import Path
from check_demo_2b10_consumers import check as predecessor
from audit_10min_demo import require
from audit_10min_demo_graph import Decoder, bonus_program
from audit_10min_demo_new_ball_threshold import ROUTE

OPS = (
 (0x4720,"call","0x2ab5"), (0x2db5,"cmp","byte ptr [0x34ca], 0xff"), (0x2dbb,"jne","0x2ac0"), (0x2dbd,"jmp","0x2ba0"),
 (0x5f1c,"cmp","al, byte ptr [0xca0]"), (0x5f25,"cmp","byte ptr [0x3494], 0xff"),
 (0x5f2f,"cmp","byte ptr [0x34e1], 0xff"), (0x5f3c,"call","0x5cb4"),
 (0x5f41,"mov","word ptr [0x37f5], 0"), (0x5f64,"call","0x6ad3"),
 (0x5f67,"mov","byte ptr [0x34dd], 0xff"), (0x5f74,"call","0x6ad3"),
 (0x5f7c,"jb","0x5ca5"), (0x5f89,"cmp","byte ptr [0x3494], 0xff"), (0x5f93,"cmp","byte ptr [0x34e1], 0xff"),
 (0xd75,"cmp","byte ptr [0xcd], 0"), (0xd7a,"jne","0xa7f"), (0xd7c,"jmp","0xb21"),
 (0x2f93,"mov","cx, 6"), (0x2f96,"xor","ax, ax"), (0x2f98,"cmp","word ptr [bx], ax"),
 (0x2f9a,"jne","0x2ca8"), (0x2f9f,"add","bx, 2"), (0x2fa2,"loop","0x2c98"),
 (0x2fac,"jb","0x2cb7"), (0x2fb1,"mov","bx, word ptr [bx + 4]"), (0x2fce,"jmp","word ptr [bx]"),
 (0x5d2,'cmp','byte ptr [0x34cf], 0xff'), (0x5d8,'je','0x322'),
 (0x5dd,'mov','byte ptr [0x3485], 0'), (0x5e3,'mov','si, 0x6d5'),
 (0x5e6,'call','0x5c14'), (0x5e9,'mov','byte ptr [0x3485], 0'),
 (0x5ef,'mov','byte ptr [0x2405], 0x3e'), (0x5f5,'mov','dx, 0x2fc'),
 (0x5f8,'call','0x5b80'), (0x5fb,'ret',''),
 (0x5fc,'mov','dx, 5'), (0x5ff,'mov','bx, 0x36cb'), (0x602,'call','0x57c7'),
 (0x605,'jb','0x30b'), (0x60a,'ret',''), (0x60b,'mov','bh, 0'),
 (0x60d,'mov','cl, byte ptr [0xc4d]'), (0x611,'mov','bl, byte ptr [0xc4e]'),
 (0x615,'mov','dl, byte ptr [0xc50]'), (0x619,'inc','dl'),
 (0x61b,'mov','al, 0x11'), (0x61d,'int','0x66'), (0x61f,'jmp','0x57ba'),
 (0x622,'mov','si, 0x6f1'), (0x625,'call','0x5c14'),
 (0x73e,'push','bx'), (0x73f,'cmp','byte ptr [0x5b3], 0xff'),
 (0x744,'jne','0x456'), (0x756,'inc','byte ptr [0x238a]'),
 (0x75a,'cmp','byte ptr [0x238a], 0x41'), (0x75f,'jb','0x47f'),
 (0x764,'mov','byte ptr [0x238a], 0x37'), (0x769,'cmp','byte ptr [0x2389], 0x38'),
 (0x76e,'jne','0x47a'), (0x773,'inc','byte ptr [0x2389]'),
 (0x777,'jmp','0x47f'), (0x77a,'mov','byte ptr [0x2389], 0x38'),
 (0x77f,'call','0xe78'), (0x782,'mov','dx, 0xbbb'), (0x785,'call','0x5b80'),
 (0x788,'pop','bx'), (0x789,'jmp','0xb21'),
 (0xebb,'mov','dx, 0x1e'), (0xebe,'mov','bx, 0x36cd'),
 (0xec1,'call','0x57c7'), (0xec4,'jb','0xbca'), (0xeca,'call','0xbce'),
 (0x3af5,'call','0x5831'), (0x3af8,'call','0x3842'), (0x3afb,'call','0x37af'),
 (0x3afe,'cmp','byte ptr [0xd0], 0xff'), (0x3b03,'je','0x3841'),
 (0x3b0d,'call','0x4d85'), (0x3b10,'cmp','byte ptr [0x34f1], 0xff'),
 (0x3b15,'jne','0x3822'), (0x3b1a,'mov','byte ptr [0x34f1], 0'),
 (0x3b3a,'mov','bx, 0x1ade'), (0x3b3d,'call','0x4501'),
 (0xf46,'cmp','byte ptr [0xd1], 0xff'), (0xf4b,'je','0xc5b'),
 (0xf50,'mov','si, 0xc8b'), (0xf53,'call','0x5cb4'),
 (0xf56,'mov','byte ptr [0xd1], 0'), (0xf5b,'mov','byte ptr [0x2405], 0'),
)

def check(a,b):
    predecessor(a,b)
    d=Decoder()
    for at,op,args in OPS: d.expect(b,768,at,op,args)
    program,handlers=bonus_program(b,0x19db0,True)
    for at,op,args,cost in ROUTE:
        require(struct.unpack_from('<'+'H'*(1+len(args)),b,at)==(handlers[op],*args),'zero route operands')
    require(struct.unpack_from('<4H',b,0x1b535)==(handlers['_CLEAR4'],handlers['_WAIT'],32000,0),'bonus tail')
    require(struct.unpack_from('<H24BH',b,0x1a485)==(0xca0,*([0]*24),0x16a9),'LOSTBALL effect')
    require(tuple(b[0x1aa50:0x1aa53])==(6,1,255),'lost jingle')
    require(tuple(b[0x1a9fd:0x1aa01])==(28,18,0,3),'rinner sound')
    require(struct.unpack_from('<8H',b,0x1b88e)==(0x4ebe,0x4c67,0x2379,336,0x4c67,0x2385,1684,0),'reset SHOWPLAYERSTS')
    require(b[0x1bb7e:0x1bb7e+17]==b'GETTING SICK HUH\0','drain text')
    require(b[0x1c135:0x1c135+7]==b'BALL 8\0','mutable source ball text')
    return True

if __name__=='__main__':
    import os
    root=Path(__file__).resolve().parents[1]
    check((root/'TABLE1.PRG').read_bytes(),(Path(os.environ['PF_10MIN_DEMO_DATA'])/'TABLE1.PRG').read_bytes())
