#!/usr/bin/env python3
"""C INTRO role binding against the already reviewed D full-build callsites.
Static research only, no DOS execution or payload output. TABLE semantic proof
is reused because the alt layout compiler proves only twelve seed bytes differ.
"""
import argparse
import hashlib
import json
import struct
from pathlib import Path
from deluxe_cd_alt_layout import ALT_FINGERPRINT
from deluxe_cd_layout import FINGERPRINT, fingerprint, ROOT

def prove(data, deluxe, source):
    from capstone import Cs, CS_ARCH_X86, CS_MODE_16
    assert fingerprint(data)==ALT_FINGERPRINT and fingerprint(deluxe)==FINGERPRINT
    inventory={r['name']:r for r in json.loads((ROOT/'analysis/source-inventory.json').read_text())}
    asm=(source/'INTRO.ASM').read_bytes()
    assert hashlib.sha256(asm).hexdigest()==inventory['INTRO.ASM']['sha256']
    assert 'UNPKPICS:' in asm.decode('latin1').upper()
    c,d=(data/'INTRO.PRG').read_bytes(),(deluxe/'INTRO.PRG').read_bytes()
    cs=Cs(CS_ARCH_X86,CS_MODE_16);cs.detail=True
    result=[]
    for role,old_start,old_end,old_operand,at,operand,target,y in [
      ('t21st1seg',0x38b7e,0x38b93,0x38b8d,0x38c9b,0x38caa,0x3b960,0),
      ('t21st2seg',0x38bb7,0x38bcd,0x38bc7,0x38cd4,0x38ce4,0x3d940,123)]:
        ins=list(cs.disasm(c[at:at+old_end-old_start],at))
        old=list(cs.disasm(d[old_start:old_end],old_start))
        # Exact instruction shape/call sequence of the already bound source role.
        assert [(x.mnemonic,[o.type for o in x.operands]) for x in ins]==[(x.mnemonic,[o.type for o in x.operands]) for x in old]
        base=struct.unpack_from('<H',c,8)[0]*16
        assert base+struct.unpack_from('<H',c,operand)[0]*16==target
        assert any(x.mnemonic=='mov' and x.op_str in (f'bx, {hex(y)}',f'bx, {y}') for x in ins)
        assert any(x.mnemonic=='call' and x.operands[0].imm==0x3b290 for x in ins)
        # All operands except reviewed segment and relocated UNPKLBM are exact.
        for x,z in zip(ins,old):
            if x.mnemonic not in ('push','call'): assert x.op_str==z.op_str
        result.append(dict(role=role,segment_operand=operand,form=target,width=320,height=123 if y==0 else 117,x=0,y=y,unpk=0x3b290))
    return result

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for n in ('data','deluxe','source'):p.add_argument('--'+n,type=Path,required=True)
    a=p.parse_args();print(json.dumps(prove(a.data,a.deluxe,a.source),indent=2))
