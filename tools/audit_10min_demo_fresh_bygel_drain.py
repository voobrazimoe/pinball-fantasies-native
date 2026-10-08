#!/usr/bin/env python3
"""Fresh BYGEL transfer gate. No simulation is permitted by a partial map.

Exports semantic records only. A relocated block comparison is not a proof of
its opaque callees, a fresh entry state, or a consumed collision path.
"""
import argparse
import json
import os
from pathlib import Path
import struct

from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_source_recurrence import linear

ROOT = Path(__file__).resolve().parents[1]
MISSING = ('Establish a reachable fresh demo NEW_BALL/SETBALL entry whose '
           'SLUMP_COUNTERN low byte equals the native fresh-session clock phase, '
           'through the real initialization prefix; neither raw PRG zero nor '
           'an authored counter seed establishes this correspondence.')

# Explicit operand correspondences reviewed against the two bounded consumers.
# No range-wide inferred address shift, nearest match, or candidate-count gate.
MEMORY = {0x35ba:0x36b7, 0xc45:0xc61, 0xc46:0xc62, 0xc48:0xc64,
          0x33e0:0x34dd, 0x36f8:0x37f5, 0x36fa:0x37f7,
          0x35d6:0x36d3, 0x2ee2:0x2fdc, 0x2ee4:0x2fde,
          0x331c:0x3416, 0x2ee6:0x2fe0, 0x2ee8:0x2fe2,
          0x2eea:0x2fe4, 0x2eec:0x2fe6, 0x2ef0:0x2fea,
          0x2eee:0x2fe8, 0x2f2c:0x3026, 0x3387:0x3481,
          0x23a7:0x24a1, 0x338c:0x3486, 0x3556:0x3653,
          0x33ef:0x34ec, 0x2ee0:0x2fda,
          0x33e4:0x34e1, 0x33d0:0x34ca, 0x338b:0x3485,
          0x230b:0x2405}
BLOCKS = {
    'BYGEL1_unlit': (0x27ed,0x283b,0x27f6,0x2844,
        {0x253b:0x2544,0x253a:0x2543,0x45b8:0x46b5,
         0x6a59:0x6ad3,0x1acc:0x1ade,0x44ab:0x4501}),
    'BYGEL2_unlit': (0x2872,0x28bd,0x287b,0x28c6,
        {0x253b:0x2544,0x25bc:0x25c5,0x45b8:0x46b5,
         0x6a59:0x6ad3,0x1acc:0x1ade,0x44ab:0x4501}),
    'SETBALL': (0xff0,0x1048,0xff4,0x104c,
        {0x35d6:0x36d3,0x5772:0x57c7,0x5765:0x57ba}),
    'SPRINGUP_velocity_rotation': (0x6165,0x61be,0x61df,0x6238,
        {0x33ef:0x34ec}),
    'drain_selection': (0x572,0x59c,0x572,0x59c,{0x69f7:0x6a71}),
    'scored_LOSTBALL_request': (0x5d2,0x5f1,0x5dd,0x5fc,
        {0x5b9a:0x5c14,0x2f1:0x2fc,0x5b06:0x5b80}),
}


def compare_block(d, a, b, name):
    al, ah, bl, bh, immediates = BLOCKS[name]
    xs, ys = linear(d,a,al,ah,768), linear(d,b,bl,bh,768)
    require(len(xs)==len(ys),name+' instruction count')
    for x,y in zip(xs,ys):
        require((x.mnemonic,x.size,len(x.operands))==
                (y.mnemonic,y.size,len(y.operands)),name+' operation')
        for u,v in zip(x.operands,y.operands):
            require((u.type,u.size,u.access)==(v.type,v.size,v.access),name+' operand')
            if u.type==d.x86.X86_OP_REG:
                require(u.reg==v.reg,name+' register')
            elif u.type==d.x86.X86_OP_MEM:
                p,q=u.mem,v.mem
                require((p.segment,p.base,p.index,p.scale)==
                        (q.segment,q.base,q.index,q.scale),name+' addressing')
                require(q.disp==MEMORY.get(p.disp,p.disp),name+' memory')
            elif u.type==d.x86.X86_OP_IMM:
                want=u.imm
                if x.group(d.cs.CS_GRP_JUMP) and name=='SPRINGUP_velocity_rotation':
                    want += bl-al  # includes explicitly reviewed release exits
                elif x.group(d.cs.CS_GRP_JUMP) and al<=u.imm+768<ah:
                    want += bl-al
                else:
                    want=immediates.get(want,want)
                require(v.imm==want,name+' constant/edge')
            else: raise ValueError(name+' unsupported operand')
    return dict(role=name,classification='RELOCATED-IDENTICAL',
                A_file_extent=[al,ah],demo_file_extent=[bl,bh],
                instructions=len(xs),scope='bounded consumer only; opaque callees excluded')


def data_records(a,b):
    rows=[]
    for i in range(8):
        al,bl=0x1c05d+16*i,0x1c1c7+16*i
        av=struct.unpack_from('<5h',a,al);bv=struct.unpack_from('<5h',b,bl)
        require(av==bv,'material record difference')
        rows.append(dict(role='material',index=i,A_file=al,demo_file=bl,
                         values=list(av),classification='RELOCATED-IDENTICAL',
                         consumed_on_witness=False))
    for label,al,bl,callback,want in [
        ('BYGEL1',0x1ab3f,0x1abcb,0x24f6,(5,455,15,465)),
        ('BYGEL2',0x1ab49,0x1abd5,0x257b,(284,455,294,465))]:
        av=struct.unpack_from('<5H',a,al);bv=struct.unpack_from('<5H',b,bl)
        require(av[:4]==bv[:4]==want and av[4]+9==bv[4]==callback,
                label+' geometry/callback binding')
        rows.append(dict(role=label,A_file=al,demo_file=bl,rectangle=list(want),
                         A_callback_file=av[4]+768,demo_callback_file=bv[4]+768,
                         classification='RELOCATED-IDENTICAL',consumed_on_witness=False))
    # Exact first settled SETBALL collision ring samples; these are input data
    # checks, not claims that a fresh prefix has reached this state.
    from re import findall
    collision=(ROOT/'internal/physics/collision.go').read_text()
    ring=collision.split('var ring = [44][2]int16{',1)[1].split('\n}',1)[0]
    points=[tuple(map(int,p)) for p in findall(r'\{(\d+), (\d+)\}',ring)]
    require(len(points)==44,'native ring declaration')
    for i,(dx,dy) in enumerate(points):
        x,y=296+dx,529+dy;local=y*40+(x>>3);mask=128>>(x&7)
        al,bl=0x3b930+local,0x3baa0+local
        require(a[al]==b[bl],'settled launch collision sample difference')
        rows.append(dict(role='SETBALL_lower_collision_sample',index=i,
                         x=x,y=y,A_file=al,demo_file=bl,
                         occupied=bool(a[al]&mask),
                         classification='RELOCATED-IDENTICAL',
                         consumed_on_witness=False))
    require(a[0x1e340:0x1e340+5120]==b[0x1e4b0:0x1e4b0+5120],
            'sine lookup record difference')
    rows.append(dict(role='sine_lookup',A_file=0x1e340,demo_file=0x1e4b0,
                     size=5120,classification='RELOCATED-IDENTICAL',
                     consumed_on_witness=False))
    return rows


def entry_gate(d,a,b,root=ROOT):
    d.expect(a,768,0x4500,'add','word ptr [0x33ef], 0x406')
    d.expect(b,768,0x4556,'add','word ptr [0x34ec], 0x406')
    d.expect(b,768,0x455c,'cmp','byte ptr [0x3812], 0xff')
    require(struct.unpack_from('<H',a,0x19d40+0x33ef)[0]==0 and
            struct.unpack_from('<H',b,0x19db0+0x34ec)[0]==0,'file initializer')
    native=(root/'internal/partyland/game.go').read_text()
    require('clock                                                                          uint16' in native,
            'native clock field drift')
    require('g.clock += 1030' in native,'native clock advancement drift')
    constructor=native.split('func New(',1)[1].split('func (g *Game) emit',1)[0]
    require('clock:' not in constructor,'native clock initializer drift')
    require('g.Release(charge, uint8(g.clock))' in
            (root/'internal/partyland/presentation.go').read_text(),'native release seed')
    session=(root/'internal/partyland/session.go').read_text()
    attract=session.split('func (g *Game) PresentationAudioSync()',1)[1].split('\n',1)[0]
    require('g.audioTick()' in attract and 'clock' not in attract,
            'native attract clock boundary drift')
    return dict(raw_file_initializer=0,native_constructor_clock=0,
                shared_update=1030,low_byte_step=6,
                demo_update_precedes_demomode_test=True,
                reachable_new_game_seed='UNKNOWN',
                classification='TRANSFER-BLOCKING-DIFFERENCE',
                classification_reason='native fresh constructor discards the demo pre-game release-clock history; not a proven inequality at every possible entry',
                native_bridge='TRANSFER BLOCKED BY UNPROVED ENTRY STATE',
                source_reset_boundary='no claim of complete alias-writer exclusion',
                permission_to_search=False,
                rejected_inference='raw file zero implies fresh gameplay clock zero')


def audit(data,canonical,historical):
    demo,full=pinned(data,canonical,historical)
    a,b=full['TABLE1.PRG'],demo['TABLE1.PRG'];d=Decoder()
    records=data_records(a,b)
    transfer=[compare_block(d,a,b,n) for n in BLOCKS]
    gate=entry_gate(d,a,b)
    return dict(verdict='FRESH_BYGEL_DRAIN_PROVENANCE = NOT_PROVED',
      drain_verdict='DRAIN_35877_MEMBERSHIP_UNKNOWN',
      SCORED_DRAIN_35877_PROVENANCE='NOT_PROVED',
      status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',
      transfer_map=transfer,selected_physical_region_data_records=records,
      entry_state_transfer_map=[dict(role='SLUMP_COUNTERN release-clock history',
        **gate)],
      fresh_initial_state=dict(SETBALL_local=dict(x=297,y=530,vx=10,vy=0,high=False,
        hold=False,wait_visits=80),actual_reachable_entry=None,
        launch_seed_gate=gate,initial_aggregate_zero='prior linked reset premise',
        initial_XXBALLE=False,initial_INH_EFF=False,
        boundary='local initializer statements; not a joined fresh state'),
      transfer_complete=False,transfer_blocking_difference_found=True,
      native_witness_authorized=False,
      unattempted_after_entry_gate=['physics integration consumer closure',
        'collision/region/material consumers on a real flight',
        'whole-flight aggregate and INH_EFF writer avoidance',
        'native-to-demo full transition projection'],
      deterministic_input_scripts=[],trajectory_checkpoints=[],
      calculation_indices=dict(BYGEL=None,scored_drain=None),
      witness_states=dict(aggregates=None,XXBALLE=None,INH_EFF=None,LOSTBALL_admitted=None),
      membership={str(n):'UNKNOWN' for n in (35876,35877,35878)},
      search=dict(scripts_executed=0,calculations_executed=0,
        space_covered='none: required transfer gate has not passed',
        failed_search_is_exclusion=False,fast_forward=False),
      theorem=dict(proved=False,reason=MISSING),
      smallest_unresolved_fact=MISSING,
      tests_unavailable=['deterministic witness replay','one-input replay mutation',
        'zero aggregates throughout witness','XXBALLE=false throughout witness',
        'INH_EFF=false at admitted LOSTBALL','exact drain index','reachable neighbors'],
      downstream_suffix='not investigated')


def main():
    p=argparse.ArgumentParser()
    for n,e in [('data','PF_10MIN_DEMO_DATA'),('canonical','PF_RUNTIME_DATA'),
                ('historical','PF_DMO0_HISTORICAL_SOURCE')]:
        p.add_argument('--'+n,type=Path,default=os.getenv(e))
    p.add_argument('--output',type=Path,
                   default=Path('/private/tmp/pf-dmo0-fresh-bygel-drain.json'))
    args=p.parse_args();require(all((args.data,args.canonical,args.historical)),'private inputs required')
    report=audit(args.data,args.canonical,args.historical)
    args.output.write_text(json.dumps(report,indent=2)+'\n')
    print(report['drain_verdict']);print(report['verdict']);return 2


if __name__=='__main__':raise SystemExit(main())
