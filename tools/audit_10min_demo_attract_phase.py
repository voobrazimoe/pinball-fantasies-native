#!/usr/bin/env python3
"""Fresh attract-phase gate: counter arithmetic is never a reachability witness.

Only linked startup/start/reset and the two actual counter producers are read.
No flight, BYGEL input search, PIT model or global alias proof is performed.
"""
import argparse
import json
import os
from pathlib import Path
import struct

from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_source_recurrence import linear
from audit_10min_demo_expiry_interleaving import wait_visit
from audit_10min_demo_fresh_bygel_drain import ROOT, entry_gate

COUNTER=0x34ec
MISSING=('Establish one reachable joint foreground-MAIN/callback residue at the '
         'automatic fresh start\'s first post-SETBALL SPRINGTASK: actual MAIN INC '
         'visits M and enabled primary ADD visits P, with actual initial C0, must satisfy C0 + M + 6*P = 6 '
         '(mod 256), through the real startup/start prefix.')
# File offsets are bound to decoded operands/edges, not counter-state seeds.
ANCHORS=(
 (0x329f,'mov','bp, ds'),(0x32ad,'mov','ax, 0x19bb'),(0x32b0,'mov','ds, ax'),
 (0x32b7,'mov','bx, 0x372d'),(0x32ba,'int','0x65'),
 (0x336e,'mov','byte ptr [0x3025], 0'),
 (0x3373,'call','0x3730'),(0x3376,'call','4'),(0x337a,'call','0x6074'),
 (0x33a2,'call','0x6045'),(0x33a5,'call','0x5945'),(0x33e2,'mov','byte ptr [0x3025], 0xff'),
 (0x33e7,'jmp','0x35fe'),
 (0x3a74,'mov','byte ptr [0x3813], 0x3b'),
 (0x6345,'mov','ax, 0xb'),(0x6348,'mov','bl, 0x64'),
 (0x634a,'mov','dx, 0x4217'),(0x634f,'int','0x66'),
 (0x6361,'mov','dx, 0x562b'),(0x6364,'mov','bl, 0xc8'),
 (0x6368,'mov','ax, 0xc'),(0x636b,'int','0x66'),
 (0x4529,'cmp','byte ptr [0x3025], 0xff'),
 (0x4556,'add','word ptr [0x34ec], 0x406'),
 (0x455c,'cmp','byte ptr [0x3812], 0xff'),
 (0x4561,'jne','0x4266'),(0x4563,'jmp','0x6086'),
 (0x3947,'cmp','byte ptr [0x3812], 0xff'),(0x3951,'call','0x33fa'),
 (0x36fa,'mov','al, byte ptr [0x37e8]'),(0x36fd,'cmp','al, 0x3b'),(0x36ff,'jb','0x342a'),(0x372c,'jne','0x3475'),(0x3775,'ret',''),
 (0x3704,'cmp','al, 0x42'),(0x371f,'mov','byte ptr [0x3813], al'),
 (0x372a,'cmp','al, 0x1c'),(0x376c,'mov','byte ptr [0x3813], al'),
 (0x39c5,'inc','word ptr [0x34ec]'),(0x39d3,'mov','ax, word ptr [0x34ec]'),(0x39d6,'mov','word ptr [0x2fda], ax'),
 (0x3a06,'jmp','0x35fe'),
 (0x6380,'mov','byte ptr [0x3812], 0xff'),
 (0x6443,'je','0x614c'),(0x644b,'retf',''),
 (0x64c5,'cmp','byte ptr [0x3813], 0'),(0x64ca,'je','0x621b'),
 (0x64cf,'mov','al, byte ptr [0x3813]'),(0x64d2,'mov','byte ptr [0x3813], 0'),
 (0x6501,'call','0x606e'),(0x636e,'mov','byte ptr [0x3812], 0'),
 (0x6504,'call','0x3853'),(0x6507,'call','0xb'),
 (0x650d,'call','0xbce'),(0x6510,'mov','byte ptr [0x34e2], 0'),
 (0x6515,'mov','dx, 0x625e'),(0x6518,'call','0x5b80'),
 (0xee0,'call','0x37bc'),(0xee3,'call','0x9d'),
 (0xf67,'cmp','byte ptr [0x34e2], 0xff'),(0xf6d,'je','0xc76'),
 (0xf76,'mov','dx, 0x6275'),(0xf79,'call','0x5b80'),
 (0x6575,'mov','dx, 0x1e'),(0x6578,'mov','bx, 0x3717'),
 (0x657b,'call','0x57c7'),(0x6584,'call','0xc7d'),
 (0x655e,'mov','dx, 0xf'),(0x6561,'mov','bx, 0x3715'),
 (0xfce,'mov','dx, 0x32'),(0xfd1,'mov','bx, 0x36d1'),
 (0xfa8,'mov','dx, 5'),(0xfab,'mov','bx, 0x36cf'),
 (0xf7d,'mov','dx, 0xcce'),(0xf83,'mov','dx, 0xcf4'),
 (0xf89,'mov','dx, 0xca8'),
 (0xff4,'mov','dx, 0x50'),(0xff7,'mov','bx, 0x36d3'),
 (0xffa,'call','0x57c7'),(0x103d,'mov','byte ptr [0x3026], 0'),
 (0x5ac7,'cmp','word ptr [bx], dx'),(0x5ace,'inc','word ptr [bx]'),
 (0x5ad2,'mov','word ptr [bx], 0'),
 (0x5e80,'mov','bx, 0x3417'),(0x5e83,'cmp','word ptr [bx], 0x6a71'),
 (0x5e8c,'add','bx, 2'),(0x5e9c,'mov','word ptr [bx], dx'),
 (0x5e9f,'mov','cx, 0x32'),(0x5eab,'call','word ptr [bx]'),
 (0x5eaf,'add','bx, 2'),(0x5eb2,'loop','0x5ba5'),
 (0x4723,'call','0x59d9'),(0x5d16,'call','0x5b9f'),
 (0x4726,'call','word ptr [0x24a3]'),
)
# Complete local bodies; returning external/rendering callees are NOT in this
# inventory. This deliberately cannot certify all runtime alias effects.
RESET_RANGES={
 'WHEN_NEW_GAME_RESET_TABLE':(0x30b,0x354),
 'RESET_VARS2':(0x354,0x39d),
 'WHEN_NEW_BALL_RESET_TABLE':(0x39d,0x3aa),
 'RESET_VARS':(0x3aa,0x4cd),
 'JUST_ONE_TIME_RESET':(0x3a30,0x3aa8),
 'WHEN_NEW_BALL_RESET':(0x3abc,0x3b42),
 'WHEN_NEW_GAME_RESET':(0x3b53,0x3bb4),
 'GO_GAME_MODE':(0x636e,0x6374),
 'NEW_BALL':(0xece,0xf7d),
 'NEW_BALL_PART_TWO':(0xf7d,0xfa8),
 'SETBALL':(0xff4,0x104c),
 'RESET_TASK_LIST':(0x3b42,0x3b53),
 'RESET_WAITLIST':(0x3aaf,0x3abc),
}


def counter_inventory(b,d):
    rows=[];stores=[];ranges=[]
    for role,(lo,hi) in RESET_RANGES.items():
        xs=linear(d,b,lo,hi,0x300)
        di=cx=None;es_ds=False
        for x in xs:
            site=x.address+0x300;ops=x.operands
            if x.mnemonic=='push' and x.op_str=='ds':es_ds='pushed DS'
            elif x.mnemonic=='pop' and x.op_str=='es' and es_ds=='pushed DS':es_ds=True
            if x.mnemonic=='mov' and len(ops)==2 and ops[0].type==d.x86.X86_OP_REG and ops[1].type==d.x86.X86_OP_IMM:
                name=x.reg_name(ops[0].reg)
                if name=='di':di=ops[1].imm
                if name=='cx':cx=ops[1].imm
            for op in ops:
                if op.type!=d.x86.X86_OP_MEM or not op.access&d.cs.CS_AC_WRITE:continue
                m=op.mem
                if not m.base and not m.index and not m.segment:
                    interval=[m.disp,m.disp+op.size]
                    require(not(interval[0]<COUNTER+2 and COUNTER<interval[1]),
                            'new-game/reset path gained SLUMP_COUNTERN writer')
                    stores.append(dict(role=role,site=site,DS_interval=interval))
                if x.mnemonic.startswith('rep stos') and es_ds is True and di is not None and cx is not None:
                    interval=[di,di+cx*op.size]
                    require(not(interval[0]<COUNTER+2 and COUNTER<interval[1]),
                            'reset string range gained counter writer')
                    ranges.append(dict(role=role,site=site,DS_interval=interval,
                                       derivation='ES=DS, literal DI and CX, architectural store width',
                                       forward_direction_premise='source startup CLD contract; does not prove external DF effects'))
        rows.append(dict(role=role,extent=[lo,hi],counter_direct_store=False,
            instruction_count=len(xs),scope='own-body direct and recognized reset string destinations only'))
    for site,delta,role in ((0x39c5,1,'foreground MAIN'),(0x4556,1030,'enabled primary callback')):
        x=d.instruction(b,0x300,site)
        rows.append(dict(role=role,site=site,DS_interval=[COUNTER,COUNTER+2],
                         delta=delta,low_byte_delta=delta&255,instruction=x.mnemonic))
    return dict(local_bodies=rows,direct_stores=stores,bounded_string_destinations=ranges,
        concrete_counter_writers=[rows[-2],rows[-1]],
        global_alias_closure=False,complete_prefix_writer_proof=False,
        scope='only this word and reviewed local reset bodies; opaque init/render/API callees not closed')


def task_prefix():
    """Finite fresh-start task scan, conditional on one admitted scan per P.

    This is a word/wait/slot derivation, not native physics, a counter seed or a
    real CPU callback/foreground scheduling witness.
    """
    tasks=['SNART','ALLOW']+[None]*48;ages={};events=[]
    limits={'SNART':30,'ALLOW':15,'SOUNDNEW':50,'SETBALL':80,'BRICK':5}
    for visit in range(1,113):
        for i in range(50):
            role=tasks[i]
            if role is None:continue
            ages[role],fire=wait_visit(ages.get(role,0),limits[role])
            if not fire:continue
            events.append(dict(visit=visit,slot=i,role=role))
            if role=='SNART':
                for new in ('SOUNDNEW','SETBALL','BRICK'):
                    free=tasks.index(None);tasks[free]=new
                    events.append(dict(visit=visit,slot=free,installed=new,
                                       visited_this_scan=free>i))
            tasks[i]=None
            if role=='SETBALL':return dict(events=events,
                installation_visit=next(e['visit'] for e in events if e.get('installed')=='SETBALL'),
                body_visit=visit,own_wait_first_visit='same ascending scan as SNART firing',
                native_trajectory_calculations=0,reachability='CONDITIONAL task-prefix derivation only')
    raise ValueError('bounded wait prefix did not reach SETBALL')


def phase(primary_visits,main_visits=0,initial=0):
    return (initial+1030*primary_visits+main_visits)&255


def solve(native=6,post=111,main=0):
    return [n for n in range(128) if phase(n+post,main)==native]


def native_boundary(root=ROOT):
    game=(root/'internal/partyland/game.go').read_text()
    physics=(root/'internal/physics/ball.go').read_text()
    presentation=(root/'internal/partyland/presentation.go').read_text()
    constructor=game.split('func New(',1)[1].split('func (g *Game) emit',1)[0]
    require('clock:' not in constructor and game.count('g.clock += 1030')==1,'native fresh phase changed')
    require('g.clock += 1030' in game.split('func (g *Game) SyncWithMatrixBudget',1)[1],'native update boundary drift')
    require('g.Physics.BeforeTargets = g.beforeTargets' in constructor and
            'g.Physics.AfterTargets = g.afterTargets' in constructor and
            game.index('g.runTasks()',game.index('func (g *Game) afterTargets')) <
            game.index('g.springControl(input)',game.index('func (g *Game) afterTargets')),
            'native task/spring ordering drift')
    require('g.Release(charge, uint8(g.clock))' in presentation,'native spring phase consumer drift')
    require('g.SpringValid = true' in physics and 'VX: 10, GY: 8' in physics and
            'SpringPosition' in physics,'native fresh physics boundary drift')
    return dict(constructor_clock=0,first_sync_clock=1030,first_spring_low_byte=6,
        selected='immediately before first springControl of a fresh native Game; DOS first post-SETBALL SPRINGTASK',
        distinction='constructor ready-state phase is 0; first actual spring consumer follows one +1030 and reads 6',
        spring_consumption='low 8 for velocity, low 4 for rotation',
        full_physics_transfer_proved=False)


def linked(b,d):
    for site,mnemonic,operand in ANCHORS:d.expect(b,0x300,site,mnemonic,operand)
    require(struct.unpack_from('<H',b,0x19db0+COUNTER)[0]==0,'linked counter initializer changed')
    # Ensure the MAIN back-edge includes the writer, and the primary branch
    # includes the increment before either DEMOMODE successor.
    main=linear(d,b,0x38fe,0x3a09,0x300)
    require(sum(x.mnemonic=='inc' and x.op_str=='word ptr [0x34ec]' for x in main)==1,
            'MAIN counter recurrence drift')
    primary=linear(d,b,0x4556,0x4566,0x300)
    require([x.mnemonic for x in primary]==['add','cmp','jne','jmp'],
            'increment before DEMOMODE branch changed')
    return dict(anchor_count=len(ANCHORS),counter_file=0x19db0+COUNTER,
        actual_process_entry=0x329f,callbacks=dict(primary=0x4517,later=0x592b,
        attract_primary=0x6386,attract_later=0x643a),
        startup_path=[0x329f,0x32ad,0x32b0,0x32b7,0x32ba,0x336e,0x3373,
                      0x3a74,0x3376,0x337a,0x6380,0x33a2,0x6345,0x636b,0x33a5,0x33e2,0x33e7,0x38fe],
        start_path=[0x64c5,0x64cf,0x64d2,0x6501,0x636e,0x6504,0x6507,0x650d,
                    0xee0,0xee3,0xf67,0xf76,0x6510,0x6515],
        opaque_initialization_callees=True)


def reset_checks(b,d,inventory):
    checks=[]
    for role,sites,field,value in (
        ('score',(0x358,0x35b,0x35e),'DS:46b5..46c0',0),
        ('aggregate totals',(0x3b3,0x3b6,0x3b9,0x3bb,0x3be,0x3c1),'DS:f4..10b',0),
        ('XXBALLE',(0x331,),'DS:cd',False),
        ('INH_EFF',(0x3ae6,),'DS:34e1',False),
        ('task list',(0x3b42,0x3b45,0x3b48,0x3b4b),'50 task words', 'DUMRET'),
        ('wait list',(0x3ab3,0x3ab6,0x3ab9),'DS:36c9..372c',0),
        ('new ball held position',(0xefb,0xf01,0xf29,0xf2f),'ball',dict(x=282,y=530,vx=0,vy=0)),
        ('SETBALL position',(0x1003,0x1009,0x1031,0x1037,0x103d),'ball',dict(x=297,y=530,vx=10,vy=0,hold=False)),
    ):
        checks.append(dict(role=role,sites=list(sites),field=field,value=value,
            result='LOCAL statements checked; fresh joined runtime state not promoted'))
    for at,op in ((0x331,'byte ptr [0xcd], 0'),(0x3ae6,'byte ptr [0x34e1], 0'),
                  (0x3b4b,'word ptr [bx], ax'),(0x3ab9,'word ptr es:[di], ax')):
        d.expect(b,0x300,at,'rep stosw' if at==0x3ab9 else 'mov',op)
    for at,mn,op in (
        (0x358,'mov','di, 0x46b5'),(0x35b,'mov','cx, 6'),
        (0x35e,'rep stosw','word ptr es:[di], ax'),
        (0x3b3,'mov','di, 0xf4'),(0x3b6,'mov','cx, 6'),
        (0x3bb,'mov','di, 0x100'),(0x3be,'mov','cx, 6'),
        (0x3b42,'mov','ax, 0x6a71'),(0x3b45,'mov','cx, 0x32'),
        (0x3b48,'mov','bx, 0x3417'),
        (0x3ab3,'mov','di, 0x36c9'),(0x3ab6,'mov','cx, 0x32'),
        (0xefb,'mov','word ptr [0x2fdc], 0x11a'),
        (0xf01,'mov','word ptr [0x2fde], 0x212'),
        (0xf29,'mov','word ptr [0x2fea], 0'),(0xf2f,'mov','word ptr [0x2fe8], 0'),
        (0x1003,'mov','word ptr [0x2fdc], 0x129'),
        (0x1009,'mov','word ptr [0x2fde], 0x212'),
        (0x1031,'mov','word ptr [0x2fea], 0'),(0x1037,'mov','word ptr [0x2fe8], 0xa'),
    ):d.expect(b,0x300,at,mn,op)
    checks.extend([
        dict(role='BYGEL light 39',result='SLACK_LIGHTS called by new-game/new-ball; no complete lamp-reset theorem claimed'),
        dict(role='spring charge',DS=0x24a1,linked_value=b[0x19db0+0x24a1],
             result='file zero and no-input spring contract only; no chosen phase-aligned prefix'),
        dict(role='held rotation',writer=0x39d6,source='MAIN reads same counter while HOLDSTILL',
             result='SETBALL local macro does not reconstruct every physics field; release overwrites rotation with low nibble'),
    ])
    return checks


def audit(data,canonical,historical):
    demo,full=pinned(data,canonical,historical);b=demo['TABLE1.PRG'];d=Decoder()
    path=linked(b,d);inventory=counter_inventory(b,d)
    # Retain the existing clock/update/consumer premises; do not rewrite the
    # previous fresh-BYGEL report or its verdict.
    entry_gate(d,full['TABLE1.PRG'],b)
    native=native_boundary();tasks=task_prefix();G=tasks['body_visit']
    require(tasks['installation_visit']==31 and G==111,'post-start task contribution changed')
    return dict(verdict='ATTRACT_PHASE_ALIGNMENT = NOT_PROVED',
        FRESH_BYGEL_ENTRY_TRANSFER='NOT_PROVED',status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',
        premises=dict(CANONICAL_A_TIMING_INHERITANCE='PROVED',NATIVE_AUDIO_BOUNDARY='PROVED'),
        startup_attract_control_path=path,
        counter_initial_state=dict(DS=COUNTER,file=path['counter_file'],linked_value=0,
            loaded_zero_is_runtime_phase=False,initial_value_used_only_under_concrete_writer_premise=True),
        concrete_writers=inventory,
        primary_increments=dict(site=0x4556,step=1030,step_mod256=6,before_demomode=True,
            guard='INTERRUPTS_ON=true; disabled callback returns before ADD',
            attract_own_body_increments=1,attract_return=0x6439,
            recurrence_scope='primary own body returns to caller; opaque rendering effects not closed'),
        attract_dwell=dict(initial_pending_start=0x3b,producer=0x3a74,consumer=0x64c5,
            no_key_is_no_start_request=False,
            first_enabled_later_accepts_pending_start=True,
            at_least_128_enabled_primary_callbacks_proved=False,
            arbitrary_no_start_dwell_proved=False,
            reason='demo preloads automatic F1; L consumes it without a keyboard event',
            serialized_P_L_example=dict(attract_primary_visits=1,external_start_key=False,
                runtime_counter_residue_certified=False)),
        start_sampling_point=dict(key_scan=0x36fa,request=0x3813,acceptance=0x64c5,
            where='foreground CHECKSTARTKEYS latches F1/F8 or Enter; later attract callback consumes request',
            relative_to_primary='selected P then L ordering: primary increment precedes acceptance; key latching itself is asynchronous foreground',
            automatic_request='already nonzero before callback installation'),
        phase_formula=dict(full='u16(C0 + 1030*P + M)',low_byte='(C0 + 6*P + M) mod 256',
            definitions=dict(N='enabled primary ADD visits before automatic start acceptance',
                M='all foreground MAIN INC visits up to the specified boundary; M varies between boundaries',
                j='enabled gameplay primary visits after start, on selected serialized prefix'),
            points=dict(start_acceptance='(C0+6*N+M_accept) mod 256',
                GO_GAME_MODE='same as acceptance within serialized L',NEW_BALL='same as acceptance within serialized L',
                SETBALL_installation='(C0+6*(N+31)+M_install) mod 256',
                SETBALL_body='(C0+6*(N+111)+M_body) mod 256',
                first_gameplay_spring='(C0+6*(N+1)+M_first) mod 256',
                first_post_SETBALL_spring='same as SETBALL body within serialized P'),
            fixed_post_start_contribution=dict(installation_primary_visits=31,body_primary_visits=G,
                raw=1030*G,low_byte=(1030*G)&255,
                scope='conditional fresh task-prefix under inherited serial scan contract; MAIN contribution is not fixed',
                K_is_not_complete_runtime_constant=True),
            simple_N_only_formula_valid=False),
        post_start_wait_derivation=tasks,native_comparison_boundary=native,
        solved_N_values=dict(if_C0_zero_and_M_zero=solve(native=6,post=G),modulus=128,
            conditional_pairs=[dict(N=n,required_M_mod256=(6-6*(n+G))&255) for n in range(128)],
            arithmetic_only=True),
        arithmetic_candidate=dict(N=1,M_mod256=(6-6*(1+G))&255,
            demo_low_byte=phase(1+G,(6-6*(1+G))&255),native_low_byte=6,reachable=False),
        chosen_reachable_alignment_witness=None,
        reset_state_checks=reset_checks(b,d,inventory),
        reset_writer_proof=dict(local_counter_reset=False,complete_runtime_prefix_proof=False,
            counter_preserved_in_reviewed_reset_bodies=True,scope=inventory['scope']),
        smallest_unresolved_fact=MISSING,
        trajectory_bridge_clear=False,
        search=dict(BYGEL_scripts_executed=0,native_trajectory_calculations=0,
                    physical_schedule_search=False),
        unchanged_prior_verdicts=dict(FRESH_BYGEL_DRAIN_PROVENANCE='NOT_PROVED',
          SCORED_DRAIN_35877_PROVENANCE='NOT_PROVED',DRAIN_35877_MEMBERSHIP='UNKNOWN'))


def main():
    p=argparse.ArgumentParser(description=__doc__)
    for n,e in [('data','PF_10MIN_DEMO_DATA'),('canonical','PF_RUNTIME_DATA'),('historical','PF_DMO0_HISTORICAL_SOURCE')]:
        p.add_argument('--'+n,type=Path,default=os.getenv(e))
    p.add_argument('--output',type=Path,default=Path('/private/tmp/pf-dmo0-attract-phase-alignment.json'))
    args=p.parse_args();require(all((args.data,args.canonical,args.historical)),'private inputs required')
    r=audit(args.data,args.canonical,args.historical)
    args.output.write_text(json.dumps(r,indent=2)+'\n');print(r['verdict']);return 2


if __name__=='__main__':raise SystemExit(main())
