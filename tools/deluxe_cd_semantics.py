#!/usr/bin/env python3
"""Recheck the source/linked-code D0 evidence, without executing DOS code.
Requires private A/D, historical source, and optional research dependency capstone.
Emits only addresses/counts/role metadata. Never emits code or asset bytes.
"""
import argparse
import hashlib
import json
import re
import struct
from pathlib import Path
from deluxe_cd_layout import HANDLERS, ROOT, fingerprint, FINGERPRINT

def prove(canonical, data, source):
    from capstone import Cs, CS_ARCH_X86, CS_MODE_16, CS_GRP_RET, CS_GRP_JUMP, CS_GRP_CALL
    from capstone.x86 import X86_OP_IMM, X86_OP_MEM, X86_OP_REG
    assert fingerprint(data)==FINGERPRINT
    game_inventory={r['name']:r for r in json.loads((ROOT/'analysis/game-inventory.json').read_text())}
    for name in ('INTRO.PRG','TABLE4.PRG'):
        assert hashlib.sha256((canonical/name).read_bytes()).hexdigest()==game_inventory[name]['sha256']
    inventory={r['name']:r for r in json.loads((ROOT/'analysis/source-inventory.json').read_text())}
    for name in ['INTRO.ASM','STONES.ASM','FANTASIE.ASM']:
        assert hashlib.sha256((source/name).read_bytes()).hexdigest()==inventory[name]['sha256'],name
    stones=(source/'STONES.ASM').read_text(encoding='latin1').upper()
    fantasy=(source/'FANTASIE.ASM').read_text(encoding='latin1').upper()
    intro=(source/'INTRO.ASM').read_text(encoding='latin1').upper()
    assert 'MOV\tAX,GHOSTAREA[BX].G_ROUTINE' in stones and 'CALL\tAX' in stones
    assert 'CMP\tLASTCHECK,AX' in fantasy and 'MOV\tLASTCHECK,AX' in fantasy
    assert 'UNPKPICS:' in intro and 'MOVE\tES,T21ST1SEG' in intro and 'MOVE\tES,T21ST2SEG' in intro
    content=json.loads(re.search(r'`(.*)`',(ROOT/'internal/stones/content.go').read_text()).group(1))
    native={v:int(k) for k,v in content['area_handlers'].items()}
    for role in HANDLERS.values():
        assert re.search(r'^'+role+r'\s*:',stones,re.M),role
    raw=[(p/'TABLE4.PRG').read_bytes() for p in [canonical,data]]
    bases=[(struct.unpack_from('<H',b,8)[0]+struct.unpack_from('<H',b,22)[0])*16 for b in raw]
    assert bases==[0x300,0x3f4e0]
    cs=Cs(CS_ARCH_X86,CS_MODE_16);cs.detail=True
    def graph(b,base,entry):
        todo=[entry];out={}
        while todo:
            at=todo.pop()
            if at in out:continue
            assert 0<=at<0x8000 and len(out)<4000,'unbounded handler graph'
            x=next(cs.disasm(b[base+at:base+at+15],at),None);assert x
            out[at]=x
            if x.group(CS_GRP_RET):continue
            if (x.group(CS_GRP_JUMP) or x.mnemonic.startswith('loop') or x.mnemonic=='jcxz'):
                if x.operands[0].type==X86_OP_IMM:todo.append(x.operands[0].imm)
                else:continue
                if x.mnemonic=='jmp':continue
            todo.append(at+x.size)
        return out
    def compare(a,d):
        ga,gd=graph(raw[0],bases[0],a),graph(raw[1],bases[1],d)
        assert len(ga)==len(gd)
        for at,x in ga.items():
            y=gd[at+d-a]
            assert (x.mnemonic,x.size,len(x.operands),x.prefix)==(y.mnemonic,y.size,len(y.operands),y.prefix)
            for u,v in zip(x.operands,y.operands):
                assert u.type==v.type and u.size==v.size
                if u.type==X86_OP_REG:assert u.reg==v.reg
                elif u.type==X86_OP_MEM:
                    assert (u.mem.segment,u.mem.base,u.mem.index,u.mem.scale)==(v.mem.segment,v.mem.base,v.mem.index,v.mem.scale)
                    # Source state/data declarations retain identity; state
                    # moved by29, effect/matrix records by80 in the DS contract.
                    assert v.mem.disp-u.mem.disp in (0,29,80)
                elif u.type==X86_OP_IMM:
                    delta=v.imm-u.imm
                    if (x.group(CS_GRP_JUMP) or x.mnemonic.startswith('loop') or x.mnemonic=='jcxz'):assert delta==d-a
                    elif x.group(CS_GRP_CALL):assert delta in (-4,3,131)
                    else:
                        # Only symbol addresses change: source state/effect and
                        # task pointers. Numeric gameplay constants stay exact.
                        assert delta==0 or (x.mnemonic in ('mov','cmp') and delta in (29,80,-4)) or (at==0x230c and x.mnemonic=='add' and u.imm==0x7bc and v.imm==0x7d9), (hex(at),x.mnemonic)
        return len(ga)
    roles=[]
    for raw_id,role in sorted(HANDLERS.items()):
        roles.append(dict(role=role,raw=raw_id,native_id=native[role],instructions=compare(native[role],raw_id)))
    # GROPC's indirect CALL AX is bounded by eight source GHOST_STRUC entries.
    callbacks=[]
    for i in range(8):
        a=struct.unpack_from('<H',raw[0],0x166d0+0x306+7*i)[0]
        d=struct.unpack_from('<H',raw[1],0x9f50+0x323+7*i)[0]
        callbacks.append(dict(slot=i,canonical=a,observed=d,instructions=compare(a,d)))
    # Follow direct helper calls as paired control-flow graphs, rather than
    # inferring identity from entry-address differences. These explicit source
    # boundaries are provided by the common native interpreter/host: IRQ wait,
    # matrix command dispatch, and platform high-score file persistence.
    pending=[(native[role],raw_id) for raw_id,role in HANDLERS.items()]
    pending += [(r['canonical'],r['observed']) for r in callbacks]
    paired={};boundaries=set()
    while pending:
        a,d=pending.pop()
        if a in paired:
            assert paired[a]==d,'inconsistent helper correspondence'
            continue
        paired[a]=d
        assert len(paired)<5000
        if (a,d)==(0x6b38,0x6bad):
            boundaries.add('DOS high-score persistence replaced by native host storage')
            continue
        x=next(cs.disasm(raw[0][bases[0]+a:bases[0]+a+15],a),None)
        y=next(cs.disasm(raw[1][bases[1]+d:bases[1]+d+15],d),None)
        assert x and y and (x.mnemonic,len(x.operands),x.prefix)==(y.mnemonic,len(y.operands),y.prefix),(a,d)
        for u,v in zip(x.operands,y.operands):
            assert u.type==v.type and u.size==v.size
            if u.type==X86_OP_REG:assert u.reg==v.reg
            elif u.type==X86_OP_MEM:
                assert (u.mem.segment,u.mem.base,u.mem.index,u.mem.scale)==(v.mem.segment,v.mem.base,v.mem.index,v.mem.scale)
                assert v.mem.disp-u.mem.disp in (0,29,80,3)
            elif u.type==X86_OP_IMM and not (x.group(CS_GRP_JUMP) or x.group(CS_GRP_CALL) or x.mnemonic.startswith('loop') or x.mnemonic=='jcxz'):
                assert v.imm-u.imm in (0,29,80,-4,3,131,-3187,-3192),(a,d,x.mnemonic)
        if x.group(CS_GRP_RET):continue
        branch=x.group(CS_GRP_JUMP) or x.group(CS_GRP_CALL) or x.mnemonic.startswith('loop') or x.mnemonic=='jcxz'
        if branch:
            if x.operands[0].type==X86_OP_IMM:pending.append((x.operands[0].imm,y.operands[0].imm))
            else:
                assert (a,d) in {(0x41f5,0x41f8),(0x4c87,0x4c8a),(0x268f,0x268b)},(a,d)
                boundaries.add({0x41f5:'IRQ wait',0x4c87:'shared matrix dispatch',0x268f:'eight proved GHOSTAREA callbacks'}[a])
            if x.mnemonic=='jmp':continue
        pending.append((a+x.size,d+y.size))
    # Exact relocated segment operands bind role to FORM, independently of its
    # ordinal position. Initial UNPKPICS uses y0/y123; dead demo UNPKPICS2 is
    # retained in the executable but is outside full-build startup control flow.
    b=(data/'INTRO.PRG').read_bytes();base=struct.unpack_from('<H',b,8)[0]*16
    def placement(start,end,segment_at,target,y):
        ins=list(cs.disasm(b[start:end],start))
        assert base+struct.unpack_from('<H',b,segment_at)[0]*16==target
        assert any(x.mnemonic=='mov' and x.op_str in (f'bx, {hex(y)}',f'bx, {y}') for x in ins)
        assert any(x.mnemonic=='call' and x.operands[0].imm==0x3b170 for x in ins)
    placement(0x38b7e,0x38b93,0x38b8d,0x3b840,0)
    placement(0x38bb7,0x38bcd,0x38bc7,0x3d820,123)
    return dict(roles=roles,ghost_callbacks=callbacks,intro_roles=[dict(role='t21st1seg',offset=0x3b840,width=320,height=123,y=0),dict(role='t21st2seg',offset=0x3d820,width=320,height=117,y=123)],helper_instruction_pairs=len(paired),native_boundaries=sorted(boundaries),scope='source-linked static mapping; no DOS execution')

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for n in ('canonical','data','source'):p.add_argument('--'+n,type=Path,required=True)
    a=p.parse_args();print(json.dumps(prove(a.canonical,a.data,a.source),indent=2))
