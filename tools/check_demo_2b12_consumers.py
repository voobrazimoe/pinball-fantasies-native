"""2B12 complete linked ShowInfoTS admission; private text stays in owned inputs."""
import struct
from pathlib import Path
from check_demo_2b11_consumers import check as predecessor
from audit_10min_demo_graph import Decoder, generated
from audit_10min_demo import require
OPS = (
    (0x316, 'mov', 'si, 0x124'),
    (0x319, 'mov', 'di, 0x10c'),
    (0x31e, 'mov', 'cx, 0xc'),
    (0x321, 'rep movsb', 'byte ptr es:[di], byte ptr [si]'),
    (0x5706, 'cmp', 'byte ptr [0x3486], 0xff'),
    (0x570b, 'je', '0x543e'),
    (0x5710, 'mov', 'word ptr [0x37f7], 0xff'),
    (0x5716, 'mov', 'al, 0x38'),
    (0x5718, 'mov', 'byte ptr [0x2341], al'),
    (0x571b, 'mov', 'al, byte ptr [0x238a]'),
    (0x571e, 'mov', 'byte ptr [0x234b], al'),
    (0x5721, 'mov', 'bx, 0x18fc'),
    (0x5724, 'call', '0x4501'),
    (0x5727, 'jmp', '0x543e'),
    (0x4801, 'call', '0x4d85'),
    (0x4804, 'mov', 'byte ptr [0x34f8], 0'),
    (0x4809, 'mov', 'word ptr [0x37f5], 0'),
    (0x480f, 'mov', 'byte ptr [0x34e0], 0xff'),
    (0x482c, 'jmp', 'word ptr [bx]'),
    (0x4be3, 'mov', 'word ptr [0x34e6], 0x491d'),
    (0x4bf7, 'mov', 'word ptr [0x3809], 0x10'),
    (0x4c00, 'mov', 'word ptr [0x34e6], 0x4973'),
    (0x4c14, 'mov', 'word ptr [0x3809], 0xfff3'),
    (0x4c1d, 'dec', 'word ptr [0x3809]'),
    (0x4c63, 'mov', 'ax, word ptr [0x380b]'),
    (0x4c66, 'cmp', 'word ptr [0x3809], ax'),
    (0x4c6a, 'je', '0x49c9'),
    (0x4c73, 'inc', 'word ptr [0x3809]'),
    (0x4cbc, 'cmp', 'word ptr [0x3809], ax'),
    (0x4cc0, 'je', '0x49c9'),
    (0x4cc9, 'mov', 'si, 0'),
    (0x4ccc, 'mov', 'word ptr [0x457a], 0'),
    (0x4cd2, 'mov', 'word ptr [0x457c], 0'),
    (0x4cd8, 'ret', ''),
)
PROGRAM = [['_CLEAR4', []],
 ['_CLEAR4', []],
 ['_RULLGARDIN_NED', ['PLAY_TEXT', 1]],
 ['_WAIT', [120]],
 ['_MATRIXLGT', [0]],
 ['_CLEAR4', []],
 ['_PRINT13', ['JACK_TEXT', 336]],
 ['_PRINT13_NUMBER', ['JACKVALUE', 368]],
 ['_FLASHON', [1]],
 ['_WAIT', [120]],
 ['_FLASHOFF', [1]],
 ['_CLEAR4', []],
 ['_RULLGARDIN_UPP', ['BONUS_TEXT', 1]],
 ['_WAIT', [120]],
 ['_CLEAR4', []],
 ['_PRINT13', ['ALLTIME_TEXT', 338]],
 ['_WAIT', [40]],
 ['_CLEAR4', []],
 ['_FLASHOFF', [1]],
 ['_MATRIXLGT', [0]],
 ['_PRINT13', ['HI_1', 336]],
 ['_PRINT13_NUMBER', ['DEMO_HI_SCORE_0', 368]],
 ['_PRINT13', ['DEMO_HI_NAME_0', 344]],
 ['_FLASHON', [1]],
 ['_WAIT', [140]],
 ['_CLEAR4', []],
 ['_FLASHOFF', [1]],
 ['_MATRIXLGT', [0]],
 ['_PRINT13', ['HI_2', 336]],
 ['_PRINT13_NUMBER', ['DEMO_HI_SCORE_1', 368]],
 ['_PRINT13', ['DEMO_HI_NAME_1', 344]],
 ['_FLASHON', [1]],
 ['_WAIT', [140]],
 ['_CLEAR4', []],
 ['_FLASHOFF', [1]],
 ['_MATRIXLGT', [0]],
 ['_PRINT13', ['HI_3', 336]],
 ['_PRINT13_NUMBER', ['DEMO_HI_SCORE_2', 368]],
 ['_PRINT13', ['DEMO_HI_NAME_2', 344]],
 ['_FLASHON', [1]],
 ['_WAIT', [140]],
 ['_CLEAR4', []],
 ['_FLASHOFF', [1]],
 ['_MATRIXLGT', [0]],
 ['_PRINT13', ['HI_4', 336]],
 ['_PRINT13_NUMBER', ['DEMO_HI_SCORE_3', 368]],
 ['_PRINT13', ['DEMO_HI_NAME_3', 344]],
 ['_FLASHON', [1]],
 ['_WAIT', [140]],
 ['_CLEAR4', []],
 ['_FLASHOFF', [1]],
 ['_CLEAR4', []],
 ['_PRINT5', ['PLAYERSTEXT', 336]],
 ['_PRINT5', ['DEMO_BALLSTEXT', 1684]],
 ['0', []]]
HANDLERS = {'_CLEAR4': 20158,
 '_RULLGARDIN_NED': 18688,
 '_WAIT': 19875,
 '_MATRIXLGT': 19790,
 '_PRINT13': 19364,
 '_PRINT13_NUMBER': 18027,
 '_FLASHON': 17819,
 '_FLASHOFF': 19772,
 '_RULLGARDIN_UPP': 18659,
 '_PRINT5': 19559,
 '0': 0}
REFS = {'PLAY_TEXT': 9017,
 'JACK_TEXT': 8986,
 'JACKVALUE': 268,
 'BONUS_TEXT': 8997,
 'ALLTIME_TEXT': 8678,
 'HI_1': 14131,
 'DEMO_HI_SCORE_0': 22,
 'DEMO_HI_NAME_0': 34,
 'HI_2': 14139,
 'DEMO_HI_SCORE_1': 38,
 'DEMO_HI_NAME_1': 50,
 'HI_3': 14147,
 'DEMO_HI_SCORE_2': 54,
 'DEMO_HI_NAME_2': 66,
 'HI_4': 14155,
 'DEMO_HI_SCORE_3': 70,
 'DEMO_HI_NAME_3': 82,
 'PLAYERSTEXT': 9081,
 'DEMO_BALLSTEXT': 9093}

def check(a,b):
    predecessor(a,b)
    d=Decoder()
    for at,op,args in OPS: d.expect(b,768,at,op,args)
    at=0x1b6ac
    content=generated('internal/presentation/content.go')['1']
    for op,args in PROGRAM:
        want=(HANDLERS[op],*(v if isinstance(v,int) else REFS[v] for v in args))
        require(struct.unpack_from('<'+'H'*len(want),b,at)==want,'ShowInfoTS linked command '+hex(at))
        at+=len(want)*2
    require(at==0x1b796,'ShowInfoTS end')
    # All four 16-byte records are already verified by the predecessor factory
    # policy. Compare presentation glyph strings without embedding private text.
    for label in ('PLAY_TEXT','JACK_TEXT','BONUS_TEXT','ALLTIME_TEXT','HI_1','HI_2','HI_3','HI_4'):
        at=0x19db0+REFS[label]; ref,n=content['text_refs'][label]; data=a[ref:ref+n]
        require(b[at:at+len(data)]==data,'ShowInfoTS text correspondence '+label)
    require(b[0x19db0+0x124:0x19db0+0x130]==bytes(map(int,'000010000000')),'reset JACKINIT')
    return True

if __name__=='__main__':
    import os
    root=Path(__file__).resolve().parents[1]
    check((root/'TABLE1.PRG').read_bytes(),(Path(os.environ['PF_10MIN_DEMO_DATA'])/'TABLE1.PRG').read_bytes())
