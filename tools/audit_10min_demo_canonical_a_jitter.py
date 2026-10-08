#!/usr/bin/env python3
"""Canonical A launch-jitter reference inheritance, not DOS phase alignment.

Only bounded launch/reset consumers are compared. No trajectory is executed,
no physical interleaving or global writer/task-domain closure is inferred.
"""
import argparse
import json
import os
from pathlib import Path
import subprocess

from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_source_recurrence import linear
from audit_10min_demo_attract_phase import native_boundary, reset_checks, counter_inventory
from audit_10min_demo_fresh_bygel_drain import compare_block, ROOT

BASE = '306d11a0c479c7ac5ee6e245f3f72eacbc665abd'
HISTORY = '8879eb0^'
# Reviewed operand correspondences, per bounded body. These are not a
# runtime indirect-target admission list, nor a range-wide relocation rule.
SLICES = {
 'table_new_game': ((779,852,779,852), {}, {22626:22711,22492:22577,25104:25226,7190:7208,7208:7226}),
 'RESET_VARS2': ((852,925,852,925), {}, {17848:18101,13211:13461,7190:7208}),
 'table_new_ball': ((925,938,925,938), {}, {22626:22711,22492:22577,3400:3404}),
 'RESET_VARS': ((938,1229,938,1229), {8760:9008,13084:13334,13235:13485,13306:13559}, {22492:22577,22626:22711,22403:22488,22429:22514,22803:22888,4277:4281,4313:4317,4349:4353}),
 'generic_new_ball_prefix': ((15077,15143,15036,15102), {13080:13330,13196:13446,13197:13447,13198:13448,13200:13450,13201:13451,13210:13460,13264:13514,13284:13537,13305:13558,13307:13560}, {27127:27249,22492:22577,14443:14402,14296:14255}),
 'reset_tasks': ((15211,15228,15170,15187), {}, {27127:27249,13085:13335,14452:14411}),
 'reset_waits': ((15064,15077,15023,15036), {}, {13772:14025}),
 'NEW_BALL': ((3786,3961,3790,3965), {13264:13514,9118:9368,13282:13535,12076:12326,12002:12252,12004:12254,13084:13334,12006:12256,12008:12258,12010:12260,12012:12262,12016:12266,12014:12264,13191:13441,13285:13538,8971:9221,13280:13533}, {14309:14268,22403:22488,3183:3211,23610:23732,3193:3197,25083:25205,23302:23424}),
 'NEW_BALL_PART_TWO': ((3961,4004,3965,4008), {13265:13515,8978:9228,8980:9230,13280:13533}, {3274:3278,23302:23424,3312:3316,3236:3240}),
 'SPRINGTASK': ((24840,25089,24962,25211), {13650:13903,9127:9377,9129:9379,13196:13446,13654:13907,13295:13548,12016:12266,12014:12264,12000:12250,3133:3161,3134:3162,3136:3164}, {26107:26229,24165:24287,24044:24166,13295:13548,24072:24194}),
 'GO_GAME_MODE': ((25332,25338,25454,25460), {14101:14354}, {}),
 'SLACK_LIGHTS': ((0x5b62,0x5b8f,0x5bb7,0x5be4), {0x2f2d:0x3027,0x3593:0x3690,0x12bd:0x12d9}, {0x69f7:0x6a71,0x588f:0x58e4}),
 'SNART_NEW_BALL': ((25851,25866,25973,25988), {}, {13850:14103,22386:22471}),
}


def correspondence(d, a, b):
    rows=[]
    for role,(extent,mem,imm) in SLICES.items():
        al,ah,bl,bh=extent
        xs,ys=linear(d,a,al,ah,768),linear(d,b,bl,bh,768)
        require(len(xs)==len(ys),role+' instruction count')
        for x,y in zip(xs,ys):
            require((x.mnemonic,x.size,len(x.operands))==(y.mnemonic,y.size,len(y.operands)),role+' operation')
            for u,v in zip(x.operands,y.operands):
                require((u.type,u.size,u.access)==(v.type,v.size,v.access),role+' operand')
                if u.type==d.x86.X86_OP_REG:require(u.reg==v.reg,role+' register')
                elif u.type==d.x86.X86_OP_MEM:
                    p,q=u.mem,v.mem
                    require((p.segment,p.base,p.index,p.scale)==(q.segment,q.base,q.index,q.scale),role+' addressing')
                    require(q.disp==mem.get(p.disp,p.disp),role+' memory')
                elif u.type==d.x86.X86_OP_IMM:
                    # Every branch in these bodies has the reviewed local
                    # control relocation, including the task/release exits.
                    want=imm.get(u.imm,u.imm+(bl-al)) if x.group(d.cs.CS_GRP_JUMP) else imm.get(u.imm,u.imm)
                    require(v.imm==want,role+' constant/edge')
                else:raise ValueError(role+' unsupported operand')
        rows.append(dict(role=role,A_extent=[al,ah],demo_extent=[bl,bh],instructions=len(xs),
                         classification='RELOCATED-IDENTICAL',opaque_callees_closed=False))
    for name in ('SETBALL','SPRINGUP_velocity_rotation'):
        rows.append(compare_block(d,a,b,name))
    for role,aa,bb,mn,ao,bo in (
        ('primary increment',0x4500,0x4556,'add','word ptr [0x33ef], 0x406','word ptr [0x34ec], 0x406'),
        ('MAIN increment',0x39e9,0x39c5,'inc','word ptr [0x33ef]','word ptr [0x34ec]'),
        ('DEMOMODE test after increment',0x4506,0x455c,'cmp','byte ptr [0x3715], 0xff','byte ptr [0x3812], 0xff')):
        d.expect(a,768,aa,mn,ao);d.expect(b,768,bb,mn,bo)
        rows.append(dict(role=role,A_site=aa,demo_site=bb,classification='RELOCATED-IDENTICAL'))
    for at,mn,op in ((0x6206,'mov','ax, 0xff5a'),(0x6210,'and','ax, 0xff'),
                     (0x6213,'sub','bp, ax'),(0x6233,'and','word ptr [0x2fda], 0xf')):
        d.expect(b,768,at,mn,op)
    return rows


def native_policy(root=ROOT):
    r=native_boundary(root)
    game=(root/'internal/partyland/game.go').read_text()
    physics=(root/'internal/physics/ball.go').read_text()
    session=(root/'internal/partyland/session.go').read_text()
    runtime=(root/'internal/frontend/runtime.go').read_text()
    model=(root/'internal/frontend/model.go').read_text()
    require('g := partyland.New(table, tableData)' in runtime and
            's, e := m.tableFactory()(m.Scores[m.Selected-1][0].Digits)' in
            model.split('func (m *Model) startPlayers',1)[1].split('func ',1)[0], 'native fresh factory drift')
    require('func (g *Game) CarryLoadedTableState' not in '\n'.join(p.read_text() for p in (root/'internal/partyland').glob('*.go')),
            'native Party Land gained carried state')
    constructor=game.split('func New(',1)[1].split('func (g *Game) emit',1)[0]
    for field in ('Score','SkillTunnel','SkillCyclone','HappyTotal','MegaTotal','ExtraBalls','inhibitEffect','clock'):
        require(field+':' not in constructor, 'native fresh selected state drift: '+field)
    require('clock' not in constructor, 'native constructor gained counter history')
    require('g := &Game{' in constructor and 'g.resetBall()' in constructor and
            'g.BallNumber = 1' in constructor and 'g.Session.Initialize(1,' in constructor,
            'native fresh state drift')
    sync=game.split('func (g *Game) SyncWithMatrixBudget',1)[1].split('func (g *Game) Release',1)[0]
    require(sync.index('g.clock += 1030') < sync.index('g.audioTick()') < sync.index('g.Physics.Sync(input)'),
            'native clock ordering drift')
    require('func (g *Game) PresentationAudioSync() { g.audioTick() }' in session,
            'native attract counter history drift')
    require('g.Ball.VY = -166*int16(charge) - int16(jitter)' in physics and
            'g.Ball.Rotation = int16(jitter & 15)' in physics,'native source arithmetic drift')
    require('clock' not in (root/'internal/partyland/hotseat.go').read_text(),'native player carries clock')
    r.update(fresh_factory=True,pre_game_counter_history_discarded=True,
        recurrence='u16(1030*n) at gameplay Sync n; spring reads low8, rotation low4',
        advancement='before audio/phase branches/physics; tasks precede springControl',
        physical_DOS_foreground_interleaving_represented=False,
        arithmetic='high-resolution VY=-166*charge-low8(clock); VX=0; rotation=low4(clock)',
        scope='fresh launch initialization only; existing A logical clock policy retained')
    reset=game.split('func (g *Game) resetBall()',1)[1].split('func (g *Game)',1)[0]
    require('clock' not in reset and 'g.tasks = [50]func() bool{}' in reset and
            'g.waitCounters = make(map[string]uint16)' in reset and
            'g.inhibitEffect = false' in reset, 'native reset-state preservation drift')
    r['reset_clock_each_ball']=False
    r['presentation_carry']='matrix display memory may carry; selected gameplay fields and clock do not'
    return r


def historical_policy():
    def read(path):
        return subprocess.check_output(['git','show',HISTORY+':'+path],cwd=ROOT,text=True)
    pf8=read('analysis/pf8-dos-parity.md');pf6=read('analysis/pf6-frontend.md')
    require('remains the existing clock convention; exact DOS SLUMP_COUNTERN callback phase' in pf8 and
            'is unresolved and is not described as authentic.' in pf8,'historical A reference boundary drift')
    require('Table-attract state; fresh native Party Land session' in pf6 and
            'only presentation audio, never ball/rules/tasks/gameplay Tick' in pf6,
            'historical fresh-session contract drift')
    return dict(revision=subprocess.check_output(['git','rev-parse',HISTORY],cwd=ROOT,text=True).strip(),
        PF8=dict(file='analysis/pf8-dos-parity.md',lines=[197,204],
                 deterministic_native_clock_accepted=True,exact_DOS_callback_phase='UNRESOLVED_NOT_AUTHENTIC'),
        PF6=dict(file='analysis/pf6-frontend.md',lines=[18,59],fresh_session=True,attract_gameplay_frozen=True),
        accepted_tests=[dict(file='internal/physics/ball_test.go',test='TestOriginalRelease',constraint='explicit jitter=255; VY=-5567, rotation=15; arithmetic, not startup phase'),
            dict(file='internal/partyland/presentation_test.go',test='TestDownSpringChargeReleaseAndSpaceTilt',constraint='fresh direct controls do not Sync; clock=0, VY=-5312'),
            dict(file='internal/frontend/model_test.go',test='TestPF6SessionKeepsGameplayOracle',constraint='historical fresh frontend 1200-Sync rule oracle; current known baseline FAIL score=2311040 versus 2300000; not physical phase evidence'),
            dict(file='internal/partyland/session_test.go',test='TestAttractRenderDoesNotMutateSession',constraint='rendering does not run gameplay')],
        historical_pre_game_counter_phase_required_by_these_tests=False,
        TestOriginalTrajectories='NOT AVAILABLE; not executed or manufactured')


def selected_state(d,a,b):
    # Reuse the previous narrow state checks, then strengthen the four totals:
    # RESET_VARS2 clears SkillTunnel/SkillCyclone; RESET_VARS clears Happy/Mega.
    inv=counter_inventory(b,d);rows=reset_checks(b,d,inv)
    d.expect(b,768,0x33dd,'mov','byte ptr [0x3027], 0xff')
    require(b[0x19db0+0x24a1]==0 and a[0x19d40+0x23a7]==0, 'pre-game spring charge initializer drift')
    # MAKE_BAD has no launch effect on the valid-source A reference. The
    # demo eliminates A checksum failure branching, not the jitter formula.
    for blob,sites in ((a,((0x3bd0,'byte ptr [0x33f1], 0'),)),
                       (b,((0x3b75,'byte ptr [0x34ee], 0'),(0x3b9f,'byte ptr [0x34ee], 0')))):
        for at,op in sites:d.expect(blob,768,at,'mov',op)
    for at,mn,op in ((0x37e,'mov','di, 0xdc'),(0x381,'mov','cx, 6'),
                     (0x384,'rep stosw','word ptr es:[di], ax'),(0x386,'mov','di, 0xe8'),
                     (0x389,'mov','cx, 6'),(0x38c,'rep stosw','word ptr es:[di], ax')):
        d.expect(a,768,at,mn,op);d.expect(b,768,at,mn,op)
    for blob,sites in ((a,((0x3b87,'byte ptr [0x33de], 1'),(0x3b94,'byte ptr [0x371c], 1'))),
                       (b,((0x3b5e,'byte ptr [0x34db], 1'),(0x3b6b,'byte ptr [0x3819], 1')))):
        for at,op in sites:d.expect(blob,768,at,'mov',op)
    # Both table resets call the respective SLACK_LIGHTS. Source early guard
    # is retained: startup UPPSTARTAD must be true on this selected start.
    for blob,lo,hi,ds in ((a,0x5b62,0x5b8f,0x3593),(b,0x5bb7,0x5be4,0x3690)):
        xs=linear(d,blob,lo,hi,768)
        require(any(x.mnemonic=='mov' and x.op_str=='byte ptr [bx + '+hex(ds)+'], 0' for x in xs),
                'selected lights reset writer drift')
    rows.extend([dict(role='four totals',value=0,sites=[0x37e,0x386,0x3b3,0x3bb],
                     history='overwritten by two paired REP pairs'),
        dict(role='player/current ball',value=1,sites=[0x3b5e,0x3b6b],history='literal overwrite'),
        dict(role='held rotation',history='derived only from SLUMP_COUNTERN; SPRINGUP overwrites low4 before launched flight'),
        dict(role='spring charge',history='loaded zero; attract branch bypasses SPRINGTASK; selected no-launch/no-input prefix'),
        dict(role='tasks/waits',history='reset before selected SNART/ALLOW and SETBALL waits; prefix is paired, not whole task-list closure'),
        dict(role='relevant lights',history='SLACK_LIGHTS clears LIGHTSTATUS, player progression seeded by table new-game reset; selected unlit BYGEL state')])
    return dict(fields=rows,another_pre_game_trajectory_residue=None,
        player_current_ball_correspondence=dict(classification='VALUE-EQUIVALENT',A_sites=[0x3b87,0x3b94],demo_sites=[0x3b5e,0x3b6b],values=[1,1]),
        only_surviving_historical_timing_residue='SLUMP_COUNTERN and its derived held rotation',
        scope='selected fresh one-player, no-input launch construction; not whole DOS runtime state/alias closure',
        physical_prefix_join_proved=False)


def audit(data,canonical,historical):
    demo,full=pinned(Path(data),Path(canonical),Path(historical));d=Decoder()
    a,b=full['TABLE1.PRG'],demo['TABLE1.PRG']
    matched=correspondence(d,a,b);native=native_policy();history=historical_policy()
    state=selected_state(d,a,b)
    return dict(verdict='CANONICAL_A_JITTER_INHERITANCE = PROVED',production_base=BASE,
        A_native_jitter_policy=native,historical_A_evidence_boundary=history,
        A_demo_linked_jitter_comparison=matched,
        demo_specific=[dict(role='automatic F1 versus selected F1',classification='DEMO-SPECIFIC-BUT-REFERENCE-IRRELEVANT',reason='pre-session acceptance history below launch-jitter reference boundary'),
            dict(role='generic new-game UI/checksum',classification='DEMO-SPECIFIC-BUT-REFERENCE-IRRELEVANT',reason='player/current ball stores retained; MAKE_BAD ends false for valid A and demo; UI text/checksum differences do not choose jitter')],
        surviving_pre_game_state_audit=state,
        reference_boundary=dict(scope='launch-jitter initialization ONLY',
            classification=dict.fromkeys(['historical C0','foreground MAIN count M before gameplay','callback count P before gameplay','automatic F1 acceptance phase','CPU/IRQ relative timing before session construction'],'BELOW-NATIVE-REFERENCE-BOUNDARY'),
            native_semantic_start='canonical A fresh clock=0; first springControl after Sync reads 6',
            preserves='source arithmetic for every selected logical gameplay clock input; no physical DOS sequence identity claimed'),
        HISTORICAL_ATTRACT_PHASE='NOT_PROVED_BUT_BELOW_NATIVE_REFERENCE_BOUNDARY',
        historical_phase_theorem_proved=False,reachable_C0_M_P_witness=None,
        ATTRACT_PHASE_ALIGNMENT='NOT_PROVED',attract_alignment_blocks_native_equivalence=False,
        FRESH_BYGEL_ENTRY_TRANSFER='PROVED',fresh_entry_scope='bounded selected fresh-session launch reference; does not prove flight/drain provenance',
        remaining_transfer_reason=None,next_pass='deterministic BYGEL trajectory search may now run under inherited A reference; not run here',
        unchanged=dict(FRESH_BYGEL_DRAIN_PROVENANCE='NOT_PROVED',DRAIN_35877_MEMBERSHIP='UNKNOWN',DMO0='NOT CLOSED',DMO1='NOT STARTED'),
        search=dict(BYGEL_scripts_executed=0,native_trajectory_calculations=0,physical_schedule_search=False),
        prior_artifacts='unchanged',checks={})


def main():
    p=argparse.ArgumentParser(description=__doc__)
    for n,e in [('data','PF_10MIN_DEMO_DATA'),('canonical','PF_RUNTIME_DATA'),('historical','PF_DMO0_HISTORICAL_SOURCE')]:p.add_argument('--'+n,default=os.getenv(e))
    p.add_argument('--output',type=Path,default=Path('/private/tmp/pf-dmo0-canonical-a-jitter.json'))
    args=p.parse_args();require(all((args.data,args.canonical,args.historical)),'private inputs required')
    r=audit(args.data,args.canonical,args.historical)
    args.output.write_text(json.dumps(r,indent=2)+'\n');print(r['verdict'])
if __name__=='__main__':main()
