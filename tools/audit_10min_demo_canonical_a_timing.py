#!/usr/bin/env python3
"""Narrow canonical-A reference inheritance; metadata only, no DOS closure."""
import argparse
import json
import os
import struct
from pathlib import Path
from audit_10min_demo_graph import Decoder, pinned, bonus_program, require
from audit_10min_demo_source_recurrence import linear
from audit_10min_demo_scheduler_confluence import pair_a, matrix_step
from audit_10min_demo_sdr import SDR, sha, unpack, callback_api
from audit_10min_demo_paired_scheduler import indexed_shape
from audit_10min_demo_audio_boundary import dispatch_shape
ROOT = Path(__file__).resolve().parents[1]
BLOCKS = {'animation': (29366, 29584, 29476, 29694),
 'electronics_suffix': (23686, 23717, 23808, 23839),
 'later': (22742, 23075, 22827, 23160),
 'matrix': (18247, 18390, 18333, 18476),
 'physics_boundary': (23717, 23764, 23839, 23886),
 'primary': (17601, 18247, 17687, 18333),
 'registration': (25291, 25332, 25413, 25454)}
MEMORY = {('electronics_suffix', 0, 13198): 13448,
 ('later', 0, 3173): 3201,
 ('later', 0, 3174): 3202,
 ('later', 0, 3176): 3204,
 ('later', 0, 7579): 7827,
 ('later', 0, 8975): 9225,
 ('later', 0, 9131): 9381,
 ('later', 0, 9132): 9382,
 ('later', 0, 12041): 12291,
 ('later', 0, 12043): 12293,
 ('later', 0, 12045): 12295,
 ('later', 0, 12075): 12325,
 ('later', 0, 13285): 13538,
 ('later', 0, 13647): 13900,
 ('later', 0, 13648): 13901,
 ('later', 0, 14068): 14321,
 ('later', 0, 14101): 14354,
 ('later', 0, 14102): 14355,
 ('later', 0, 14104): 14357,
 ('later', 11, 17478): 17564,
 ('matrix', 0, 13209): 13459,
 ('matrix', 0, 13283): 13536,
 ('matrix', 0, 13287): 13540,
 ('matrix', 0, 13289): 13542,
 ('matrix', 0, 13291): 13544,
 ('matrix', 0, 13307): 13560,
 ('matrix', 0, 14072): 14325,
 ('matrix', 0, 14076): 14329,
 ('matrix', 0, 14078): 14331,
 ('matrix', 0, 14080): 14333,
 ('matrix', 28, 17896): 18149,
 ('physics_boundary', 0, 8980): 9230,
 ('physics_boundary', 0, 9118): 9368,
 ('physics_boundary', 0, 13264): 13514,
 ('primary', 0, 8975): 9225,
 ('primary', 0, 9129): 9379,
 ('primary', 0, 9132): 9382,
 ('primary', 0, 9133): 9383,
 ('primary', 0, 12004): 12254,
 ('primary', 0, 12036): 12286,
 ('primary', 0, 12040): 12290,
 ('primary', 0, 12041): 12291,
 ('primary', 0, 12043): 12293,
 ('primary', 0, 12044): 12294,
 ('primary', 0, 12045): 12295,
 ('primary', 0, 12075): 12325,
 ('primary', 0, 13295): 13548,
 ('primary', 0, 13298): 13551,
 ('primary', 0, 13645): 13898,
 ('primary', 0, 13647): 13900,
 ('primary', 0, 13648): 13901,
 ('primary', 0, 13654): 13907,
 ('primary', 0, 14068): 14321,
 ('primary', 0, 14101): 14354,
 ('primary', 11, 17477): 17563,
 ('primary', 11, 17478): 17564,
 ('registration', 0, 13654): 13907}
IMMEDIATE = {('animation', 29370, 0, 8277): 8300,
 ('animation', 29467, 0, 28686): 28796,
 ('animation', 29484, 0, 28686): 28796,
 ('animation', 29491, 0, 28686): 28796,
 ('animation', 29537, 0, 28756): 28866,
 ('animation', 29554, 0, 28756): 28866,
 ('animation', 29561, 0, 28756): 28866,
 ('electronics_suffix', 23686, 0, 23915): 24037,
 ('electronics_suffix', 23689, 0, 23732): 23854,
 ('electronics_suffix', 23692, 0, 4044): 4048,
 ('electronics_suffix', 23695, 0, 19576): 19662,
 ('electronics_suffix', 23702, 1, 27127): 27249,
 ('electronics_suffix', 23708, 0, 23333): 23455,
 ('electronics_suffix', 23711, 0, 22507): 22592,
 ('later', 22742, 0, 6580): 6587,
 ('later', 22804, 0, 24768): 24890,
 ('later', 22822, 0, 16403): 16489,
 ('later', 22901, 0, 34736): 34848,
 ('later', 22919, 0, 16699): 16785,
 ('later', 22932, 0, 21959): 22044,
 ('later', 22960, 1, 6890): 6908,
 ('later', 22963, 0, 17579): 17665,
 ('later', 22991, 1, 25060): 25182,
 ('later', 22994, 0, 23302): 23424,
 ('matrix', 18247, 0, 17500): 17586,
 ('matrix', 18261, 1, 27127): 27249,
 ('matrix', 18288, 0, 6580): 6587,
 ('matrix', 18304, 1, 21334): 21420,
 ('matrix', 18322, 0, 17598): 17684,
 ('matrix', 18328, 0, 19759): 19845,
 ('matrix', 18347, 0, 19759): 19845,
 ('matrix', 18367, 0, 6580): 6587,
 ('matrix', 18374, 1, 17884): 18137,
 ('physics_boundary', 23731, 0, 23241): 23363,
 ('physics_boundary', 23734, 0, 22996): 23118,
 ('physics_boundary', 23737, 0, 23839): 23961,
 ('primary', 17601, 0, 6580): 6587,
 ('primary', 17677, 0, 24588): 24710,
 ('primary', 17853, 0, 21959): 22044,
 ('primary', 17903, 0, 16699): 16785,
 ('primary', 17906, 0, 34736): 34848,
 ('primary', 17922, 0, 21946): 22031,
 ('primary', 17925, 0, 34736): 34848,
 ('primary', 17961, 0, 16699): 16785,
 ('primary', 17964, 0, 34736): 34848,
 ('primary', 17970, 0, 21946): 22031,
 ('primary', 17973, 0, 34736): 34848,
 ('primary', 18015, 0, 21959): 22044,
 ('primary', 18018, 0, 22949): 23071,
 ('primary', 18122, 0, 10924): 10933,
 ('primary', 18125, 0, 22916): 23001,
 ('primary', 18160, 0, 17479): 17565,
 ('primary', 18170, 0, 22803): 22888,
 ('registration', 25296, 1, 16833): 16919,
 ('registration', 25319, 1, 21974): 22059}
CALLS = {17853: (17939, 22727, 22812),
 17903: (17989, 17467, 17553),
 17906: (17992, 35504, 35616),
 17922: (18008, 22714, 22799),
 17925: (18011, 35504, 35616),
 17961: (18047, 17467, 17553),
 17964: (18050, 35504, 35616),
 17970: (18056, 22714, 22799),
 17973: (18059, 35504, 35616),
 18015: (18101, 22727, 22812),
 18018: (18104, 23717, 23839),
 18122: (18208, 11692, 11701),
 18125: (18211, 23684, 23769),
 18160: (18246, 18247, 18333),
 18170: (18256, 23571, 23656),
 18247: (18333, 18268, 18354),
 18322: (18408, 18366, 18452),
 18328: (18414, 20527, 20613),
 18347: (18433, 20527, 20613),
 22822: (22907, 17171, 17257),
 22901: (22986, 35504, 35616),
 22919: (23004, 17467, 17553),
 22932: (23017, 22727, 22812),
 22963: (23048, 18347, 18433),
 22994: (23079, 24070, 24192),
 23686: (23808, 24683, 24805),
 23689: (23811, 24500, 24622),
 23692: (23814, 4812, 4816),
 23695: (23817, 20344, 20430),
 23708: (23830, 24101, 24223),
 23711: (23833, 23275, 23360)}
# Active animation producer ties the dispatch pointer to the matched routine.
BLOCKS['animation_producer'] = (0x47f0, 0x4814, 0x4846, 0x486a)
MEMORY.update({('animation_producer', 0, 0x33e9): 0x34e6,
               ('animation_producer', 0, 0x33eb): 0x34e8})
IMMEDIATE.update({('animation_producer', 0x4803, 0, 0x19b4): 0x19bb,
                  ('animation_producer', 0x4807, 1, 0x6fb0): 0x701e,
                  ('animation_producer', 0x4811, 0, 0x533f): 0x5395})



def match(d, a, b):
    """Full bounded blocks; only reviewed operand relocations may differ.

    These are correspondence maps, not target-domain/immutability allowlists.
    Opaque gameplay callees are paired call edges, never claimed closed here.
    """
    rows = []
    for name, (al, ah, bl, bh) in BLOCKS.items():
        xs, ys = linear(d, a, al, ah, 0x300), linear(d, b, bl, bh, 0x300)
        require(len(xs) == len(ys), name+' instruction count drift')
        for x, y in zip(xs, ys):
            site = x.address+0x300
            require(x.mnemonic == y.mnemonic and x.size == y.size and
                    len(x.operands) == len(y.operands), name+' operation/order drift')
            for i, (u, v) in enumerate(zip(x.operands, y.operands)):
                require((u.type, u.size, u.access) == (v.type, v.size, v.access), name+' operand drift')
                if u.type == d.x86.X86_OP_REG:
                    require(u.reg == v.reg, name+' register drift')
                elif u.type == d.x86.X86_OP_MEM:
                    p, q = u.mem, v.mem
                    require((p.segment,p.base,p.index,p.scale) == (q.segment,q.base,q.index,q.scale), name+' memory addressing drift')
                    want = MEMORY.get((name,p.segment,p.disp),p.disp)
                    require(q.disp == want, name+' memory relocation drift')
                elif u.type == d.x86.X86_OP_IMM:
                    want = u.imm
                    if x.group(d.cs.CS_GRP_JUMP) and al <= u.imm+0x300 < ah:
                        want += bl-al
                    else:
                        want = IMMEDIATE.get((name,site,i,u.imm),want)
                    require(v.imm == want, name+' constant/control target drift')
                else:
                    raise ValueError('unsupported operand')
        rows.append(dict(role=name, canonical_A=[al,ah], demo=[bl,bh],
                         instructions=len(xs), classification='RELOCATED-IDENTICAL',
                         scope='entire bounded block; returning gameplay calls remain opaque'))
    return rows


def interference(d,b,canonical=None):
    # Electronics prefix's sole pre-threshold effect is the word increment.
    anchors=[(0x5cdb,'inc','word ptr [0x34cd]'),
             (0x5cdf,'cmp','word ptr [0x34cd], 0x8c9e'),
             (0x5ce5,'jne','0x5a00'),
             (0x5cea,'mov','byte ptr [0x34cf], 0xff'),
             (0x5cef,'mov','byte ptr [0x3026], 0xff'),
             (0x5cf7,'call','0x5cb4'),(0x5cfd,'call','0x4501'),
             (0x4720,'call','0x2ab5'),(0x4723,'call','0x59d9'),
             (0x472a,'cmp','byte ptr [0x37f1], 0'),
             (0x73f,'cmp','byte ptr [0x5b3], 0xff'),
             (0x785,'call','0x5b80'),(0xebb,'mov','dx, 0x1e'),
             (0x5a05,'mov','bx, 0x1afc')]
    for at,m,op in anchors:d.expect(b,0x300,at,m,op)
    prefix=linear(d,b,0x5cd9,0x5d00,0x300)
    require([x.mnemonic for x in prefix[:5]] == ['pushaw','push','inc','cmp','jne'], 'timer prefix drift')
    require(all(x.mnemonic=='nop' for x in prefix[5:8]),'expiry branch padding drift')
    require(tuple(b[0x1aa38:0x1aa3b])==(62,0,0),'S_EMPTY drift')
    require(tuple(b[0x1aa4d:0x1aa50])==(13,0,255),'S_GAMEOVER2 drift')
    require(struct.unpack_from('<8H',b,0x1b88e)==(20158,19559,9081,336,19559,9093,1684,0),'PLAYERSTEXT consumed program drift')
    if canonical is not None:
        require(tuple(canonical[0x1a9ac:0x1a9af])==(62,0,1),'A S_EMPTY drift')
        require(tuple(canonical[0x1a9c1:0x1a9c4])==(13,0,1),'A S_GAMEOVER2 drift')
    program,_=bonus_program(b,0x19db0,True)
    require(program['nodes'][52]['node']=='_DEMOVER_CHANGE_PLAYER','demo continuation drift')
    return [
        dict(feature='timer increment/equality 35998',classification='DEMO-SPECIFIC-BUT-NOT-TIMING',
             sites=[0x5cdb,0x5cdf,0x5ce5,0x5d00],
             effect='one word increment at admitted electronics; unequal branch reaches common suffix before tasks/matrix'),
        dict(feature='normal-ball continuation / NEW_BALL_TASK',classification='DEMO-SPECIFIC-BUT-NOT-TIMING',
             sites=[0x73e,0x785,0xebb],effect='linked matrix continuation enqueues task; common callback scan order preserved; expiry/task replacement reachability still open'),
        dict(feature='expiry program',classification='DEMO-SPECIFIC-BUT-NOT-TIMING',
             sites=[0x5cea,0x5cef,0x5cf7,0x5cfd],effect='at equality sets expiry/hold and installs matrix in electronics before existing tasks and matrix visit; not executed below threshold'),
        dict(feature='PLAYERSTEXT',classification='DEMO-SPECIFIC-BUT-NOT-TIMING',
             effect='SHOWPLAYERSTS PRINT position 336 versus 340; content/placement, no callback/budget/countdown change'),
        dict(feature='S_EMPTY / S_GAMEOVER2 priorities',classification='DEMO-SPECIFIC-BUT-NOT-TIMING',
             effect='typed jingle cue/priority differences preserved; cue may change readiness/program duration but cannot rewrite serialized source budget or schedule'),
        dict(feature='INTRO',classification='DEMO-SPECIFIC-BUT-NOT-TIMING',
             effect='separate frontend mode/cards; entering TABLE1 registers the checked P/L pair; no INTRO rendering work inserted inside TABLE1 update')]


def timer_preexpiry(counter):
    counter=(counter+1)&65535
    return counter, counter==35998, ['electronics','tasks','matrix']


def native_order(root=ROOT):
    def body(path,start,end):
        text=(root/path).read_text();return text.split(start,1)[1].split(end,1)[0]
    def ordered(text,items):
        last=-1
        for item in items:
            at=text.find(item,last+1);require(at>last,'native order drift: '+item);last=at
    game=body('internal/partyland/game.go','func (g *Game) SyncWithMatrixBudget','func (g *Game) Release')
    ordered(game,['g.matrixTimeLeft = timeLeft','g.Tick++','g.audioTick()','if g.Phase == GameOver','if g.Phase == BallLost','g.Physics.Sync(input)'])
    phy=body('internal/physics/ball.go','func (g *Game) Sync','func (g *Game) step')
    ordered(phy,['i < 2','g.checkRamps()','g.checkLevels()','if g.Ball.Lost','g.BeforeTargets()','g.checkSpringAndTargets()','g.AfterTargets(input)','g.BeforeLate()','g.scroll()','lateSteps := 1','g.step(input)'])
    hooks=(root/'internal/partyland/game.go').read_text()
    for h in ('g.Physics.BeforeTargets = g.beforeTargets','g.Physics.AfterTargets = g.afterTargets','g.Physics.BeforeLate = g.presentationTick'):
        require(h in hooks,'native hook drift')
    tasks=body('internal/partyland/game.go','func (g *Game) afterTargets','func (g *Game) resetBall')
    ordered(tasks,['g.tiltControl(input)','g.Display.Flash()','g.runTasks()','g.springControl(input)','g.flashTick()'])
    matrix=body('internal/partyland/timing.go','func (g *Game) matrixTick()', 'func (g *Game)')
    ordered(matrix,['if !g.matrixTimeLeft','return','case "_ANIMATION"','g.Display.StepAnimation'])
    require('return g.SyncWithMatrixBudget(input, true)' in hooks,'native default budget drift')
    return dict(ordinary=['budget captured once (default true)','tick/clock','audio tracker/PCM and logical cue callbacks',
        'two early physics passes','bumper/ramps/levels and drain decision',
        'counters/areas (BeforeTargets)','targets','shift/keyboard; matrix blink; tasks; spring; lamps (AfterTargets)',
        'budget-gated matrix/animation (BeforeLate)','scroll','one late physics pass (two fast-ball)'],
        ball_loss='already lost: audio, blink, tasks, lamps, matrix; new early-physics drain returns before electronics, then runs loss tasks/matrix; GameOver ends gameplay branch',
        audio_budget='audio runs before electronics; PCM/device performance never rewrites budget; default Sync chooses true, explicit SyncWithMatrixBudget captures deterministic input',
        case_A_choice='A0: P budget=0 permits matrix 2->1; L delivered after that decision/primary logical work, no asynchronous TIME_LEFT overwrite',
        latch='P/L admission machinery replaced by one serialized logical update; no native physical latch or IRQ nesting',
        evidence='source extraction plus existing strict-input independent oracle tests; not cycle-exact hardware evidence')


ORACLES = [
    dict(test='TestEveryOriginalAnimationDrawSync',file='internal/presentation/temporal_test.go',
         constrains='original-backed independent frame draw/completion visits for every decoded animation, including DATA2:042a timer semantics'),
    dict(test='TestPF45IndependentMatrixTraces',file='internal/partyland/timing_test.go',
         constrains='original PLAND data + TABLE1.MOD independently generate exact matrix command ticks for HIDDEN/MYSTERY/HAPPY/MEGA; timers count successful logical visits'),
    dict(test='TestPF45MissedMatrixBudget',file='internal/partyland/timing_test.go',
         constrains='10 skipped scans preserve animation/program, task runs at tick 2; 171 successful scans at tick 181 start countdown'),
    dict(test='TestPF45MatrixBeforeLatePhysics',file='internal/partyland/timing_test.go',
         constrains='mode expires on matrix visit before late physics boundary'),
    dict(test='TestPF45IndependentDrainToNextBallTrace',file='internal/partyland/timing_test.go',
         constrains='independent original-data event checkpoints across drain, bonus/tasks and next ball'),
    dict(test='TestPF45TaskSlotAndSharedWaitOrdering',file='internal/partyland/timing_test.go',
         constrains='task slots/WAITSYNCS scan ordering'),
    dict(test='TestAllSourceMatrixRoutineBoundaries',file='internal/partyland/matrix_temporal_test.go',
         constrains='every reachable source routine duration including ANIM old-BX/loop and timer, no estimated duration'),
    dict(test='TestPF45PendingModeStartsAfterNextTaskScan',file='internal/partyland/timing_test.go',
         constrains='matrix expiry task begins on next task scan, 401-call wait boundary'),
    dict(test='TestEdgesConsumedOnceAndSlowHostCannotDropTicks',file='internal/source/runner_test.go',
         constrains='every due update and ordered PCM retained; host delay cannot drop a matrix visit')]


def scheduler_abi(d,data,canonical):
    rows=[]
    for name,(size,digest,_) in SDR.items():
        b=(data/name).read_bytes();a=(canonical/name).read_bytes()
        require(len(b)==size and sha(b)==digest,'demo driver identity drift')
        row=dict(driver=name,packed_bytes_equal=a==b)
        if a==b:
            row.update(classification='IDENTICAL',scheduler_semantics='shared complete supplied binary')
        else:
            require(name in ('PAS16.SDR','SB16.SDR'),'unreviewed differing SDR')
            ar,_=unpack(a);br,_=unpack(b)
            aa=callback_api(name,ar,d);bb=callback_api(name,br,d)
            # The differing modules move this code region by two bytes. The
            # low common service return region stays fixed. Data DS is derived
            # independently from each module entry, not guessed from filenames.
            def relocated(at):return at-2 if at>=0x1000 else at
            def same(x,y):
                require(x.mnemonic==y.mnemonic and x.size==y.size and len(x.operands)==len(y.operands),'SDR scheduler operation drift')
                for u,v in zip(x.operands,y.operands):
                    require((u.type,u.size,u.access)==(v.type,v.size,v.access),'SDR operand drift')
                    if u.type==d.x86.X86_OP_REG:require(u.reg==v.reg,'SDR register drift')
                    elif u.type==d.x86.X86_OP_MEM:
                        p,q=u.mem,v.mem
                        require((p.segment,p.base,p.index,p.scale)==(q.segment,q.base,q.index,q.scale),'SDR addressing drift')
                        priority={'PAS16.SDR':0x1c29,'SB16.SDR':0x1de6}[name]
                        want=p.disp-2 if p.segment==d.x86.X86_REG_CS and p.disp==priority else p.disp
                        require(q.disp==want,'SDR scheduler memory drift at '+hex(x.address)+' '+hex(p.disp)+' -> '+hex(q.disp))
                    elif u.type==d.x86.X86_OP_IMM:
                        want=u.imm
                        if x.group(d.cs.CS_GRP_JUMP) or x.group(d.cs.CS_GRP_CALL) or x.mnemonic=='loop':want=relocated(want)
                        elif x.mnemonic=='push' and want==aa['data_segment']:want=bb['data_segment']
                        require(v.imm==want,'SDR budget/constant drift at '+hex(x.address)+' expected '+hex(want)+' actual '+hex(v.imm))
                    else:raise ValueError('SDR operand unsupported')
            todo=[p['entry'] for p in aa['callback_registration']];seen=set()
            require([p['entry'] for p in bb['callback_registration']]==[relocated(p['entry']) for p in aa['callback_registration']],'SDR registration linkage drift')
            while todo:
                at=todo.pop()
                if at in seen:continue
                seen.add(at)
                try:
                    x=d.instruction(ar,0,at)
                    same(x,d.instruction(br,0,relocated(at)))
                except ValueError as e:raise ValueError(name+' registration at '+hex(at)+': '+str(e))
                if x.group(d.cs.CS_GRP_RET) or x.mnemonic=='iret':continue
                if x.group(d.cs.CS_GRP_JUMP):
                    require(x.operands[0].type==d.x86.X86_OP_IMM,'indirect registration branch')
                    todo.append(x.operands[0].imm)
                    if x.mnemonic=='jmp':continue
                todo.append(at+x.size)
            # Checked, unique indexed consumer and the complete budget helper.
            c={'PAS16.SDR':0x1c0f,'SB16.SDR':0x1dcc}[name]
            sa,sb=indexed_shape(d,ar,c),indexed_shape(d,br,c-2)
            ha,hb=dispatch_shape(d,ar,c,True),dispatch_shape(d,br,c-2,True)
            require(ha['budget']==hb['budget'],'SDR budget convention differs')
            xs=linear(d,ar,c-54,ha['budget_helper']+36,0)
            ys=linear(d,br,c-56,hb['budget_helper']+36,0)
            require(len(xs)==len(ys),'SDR dispatch extent drift')
            for x,y in zip(xs,ys):same(x,y)
            row.update(classification='RELOCATED-IDENTICAL',registration_instructions=len(seen),
                dispatch_budget_instructions=len(xs),canonical_consumer=c,demo_consumer=c-2,
                canonical_registration=aa['callback_registration'],demo_registration=bb['callback_registration'],
                budget=ha['budget'],scheduler_semantics='API11/API12 local registration branches, pointer/record stores, BL priority handling; advance/wrap/save/set/restore and full budget predicate operand-matched',
                scope='scheduler slice only; registration service calls paired opaque effects; no IRQ closure')
        rows.append(row)
    return rows

def audit(data,canonical,historical):
    demo,full=pinned(Path(data),Path(canonical),Path(historical));d=Decoder()
    matches=match(d,full['TABLE1.PRG'],demo['TABLE1.PRG'])
    drivers=scheduler_abi(d,Path(data),Path(canonical))
    inter=interference(d,demo['TABLE1.PRG'],full['TABLE1.PRG']);native=native_order()
    # Consumer chain and the Case A guards/budget/countdown were matched above.
    witness=pair_a();a0,a1=[t['final']['animation_remaining'] for t in witness['traces']]
    require((a0,a1)==(1,2),'Case A witness drift')
    return dict(verdict='CANONICAL_A_TIMING_INHERITANCE = PROVED',
        production_base='306d11a0c479c7ac5ee6e245f3f72eacbc665abd',research_head='37ff8bc38d7af4a09a70319675d6e505b0bd6a5e',
        proof_boundary='canonical A chosen native semantic reference only; neither all DOS traces nor physical placement uniqueness',
        callback_matches=matches, call_edges=[dict(canonical_site=k,demo_site=v[0],canonical_target=v[1],demo_target=v[2]) for k,v in CALLS.items()],
        semantic_correspondence=dict(TIME_LEFT=[0x36f4,0x37f1],LAST_WAS_VB_CS=[0x4446,0x449c],
            primary_budget=[0x44e1,0x4537],later_budget=[0x5901,0x5956],
            UPDATE_COUNTERS=[0x46ca,0x4720],DO_ELECTRONICS=[0x46cd,0x4723],
            matrix_budget_decision=[0x46d4,0x472a],matrix_visit=[0x46f0,0x4746],
            later_latch_clear=[0x5a14,0x5a69],rest_busy_clear=[0x46fd,0x4753],
            animation_countdown=[0x72c2,0x7330],countdown_operand='DATA2:042a identical word; DEC/zero branch/reload old BX/loop/end identical'),
        scheduler_ABI=drivers,
        scheduler_ABI_correspondence=dict(API11='primary DX:ES, BL100',API12='later DX:ES, BL200, CX264/174',
            input_budget='AX zero -> TIME_LEFT true; nonzero -> false, shared field written before guards',
            handshake='P sets LAST_WAS_VB after ball; admitted L clears it after later body; neither rest busy nor budget stored per invocation',
            native_projection='one serialized primary update with captured budget, matrix before late work; no device-driven nesting'),
        native_A_event_order=native,oracle_test_evidence=ORACLES,
        oracle_scope='strict A original identity via testinputs.Require/oracle.Verify; independent source data expectations, deterministic schedule observations, not DOS IRQ waveform captures',
        demo_specific_interference=inter,
        case_A_projection=dict(A_admits_same_abstract_divergence=True,AX_P=0,AX_L=65535,
            countdown_after_matrix_then_L=a0,countdown_after_L_then_matrix=a1,
            chosen='A0, serialized primary matrix before later logical event; default native matrix budget true',
            validation='independent exact matrix traces + missed-budget/temporal/order tests constrain chosen visit contract; no claim oracle uniquely rules out every DOS schedule',
            demo_interference='no callback/control/animation differences; timer increment inside existing electronics, continuation/cues preserved as data/control differences',
            reuse='same semantic schedule admissible without proving A1 physically impossible'),
        native_audio_boundary='NATIVE_AUDIO_BOUNDARY = PROVED',
        native_audio_scope='under canonical-A serialized reference and supplied SDR semantics only; cue priority/duration differences retained, arbitrary SOUND.CFG/physical IRQ not modeled',
        historical_unknowns=dict.fromkeys(['PIT due-entry recurrence','CPU-dependent L placement','physical nested callback timing','VGA tail readback/opaque rendering timing','physical audio-budget phase'],'BELOW-NATIVE-REFERENCE-BOUNDARY'),
        remaining_DMO0_blockers=['TABLE1/CODE2 consumed indirect-domain and writer/alias closure',
            'demo expiry versus pending NEW_BALL/SETBALL/effect replacement reachability',
            'negative persistence and alternate termination/control path closure',
            'exhaustive consumed-control difference audit and demo presentation coverage'],
        whole_DOS_gate=dict(exit=2,status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',changed=False),
        prior_artifacts='unchanged; previous DOS-level NOT_PROVED verdicts remain historical facts',
        checks='populate separate current-pass run record after execution')


def main():
    p=argparse.ArgumentParser(description=__doc__)
    for name,env in [('data','PF_10MIN_DEMO_DATA'),('canonical','PF_RUNTIME_DATA'),('historical','PF_DMO0_HISTORICAL_SOURCE')]:
        p.add_argument('--'+name,default=os.getenv(env))
    p.add_argument('--output',required=True,type=Path);args=p.parse_args()
    result=audit(args.data,args.canonical,args.historical)
    args.output.write_text(json.dumps(result,indent=2)+'\n');print(result['verdict'])
if __name__=='__main__':main()
