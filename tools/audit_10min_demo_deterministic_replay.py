#!/usr/bin/env python3
"""Owner-local, research-only native replay and consumed-path projection.

Production files are read, never edited. Original PRGs stay external. The
output contains numeric trajectory/consumer metadata, never payload slices.
"""
import argparse
import json
import os
from pathlib import Path
import shutil
import struct
import subprocess

from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_source_recurrence import linear
from audit_10min_demo_fresh_bygel_drain import compare_block
from audit_10min_demo_canonical_a_jitter import audit as entry_audit
from audit_10min_demo_scored_drain import linked as scored_linked
from audit_10min_demo_programs import identities

ROOT = Path(__file__).resolve().parents[1]
TEMPLATE = ROOT/'tools/dmo0_deterministic_replay_test.go.txt'
MAP = ROOT/'tools/dmo0_replay_correspondence.json'
BASE = '306d11a0c479c7ac5ee6e245f3f72eacbc665abd'
TARGET = dict(Release=35460, Charge=22, LP=52, LD=8, LO=46, RP=30, RD=21, RO=22)


def replace_once(path, old, new):
    s=path.read_text();require(s.count(old)==1,'observer anchor drift: '+path.name)
    path.write_text(s.replace(old,new,1))


def prepare(harness, canonical):
    """Copy tracked sources, verify identity, then add passive observers only."""
    harness.mkdir(parents=True,exist_ok=True)
    names=subprocess.check_output(['git','ls-files'],cwd=ROOT,text=True).splitlines()
    base_names=set(subprocess.check_output(['git','ls-tree','-r','--name-only',BASE],cwd=ROOT,text=True).splitlines())
    copied=[];research_tests=[]
    for name in names:
        if name.startswith('internal/') or name.endswith('.go') or name in ('go.mod','go.sum'):
            src=ROOT/name;dst=harness/name;dst.parent.mkdir(parents=True,exist_ok=True)
            shutil.copyfile(src,dst)
            if name.endswith('.go') or name in ('go.mod','go.sum'):
                if name in base_names:
                    require(dst.read_bytes()==subprocess.check_output(['git','show',BASE+':'+name],cwd=ROOT), 'production base drift: '+name)
                else:
                    require(name.endswith('_test.go'),'new production source after base: '+name)
                    research_tests.append(name)
                copied.append(name)
    for name in ('TABLE1.PRG','PINBALL.CFG'):
        p=harness/name
        if p.is_symlink():p.unlink()
        require(not p.exists(),'refuse to replace private fixture copy')
        p.symlink_to(canonical/name)
    patches={
      'internal/physics/ball.go':[
       ('type Game struct {','type Game struct {\n ResearchTrace func(string,...int)'),
       ('if g.BeforeTargets != nil {','g.research("electronics")\n if g.BeforeTargets != nil {'),
       ('start := int(f.Frame) * f.stride','g.research("flipper_copy",i,int(f.Frame))\n start := int(f.Frame) * f.stride'),
       ('if occupied == 0 {','g.research("ramp",int(boolInt(b.High)),i,int(occupied),int(ramp))\n if occupied == 0 {'),
       ('for _, r := range g.Table.levels[index] {','for j, r := range g.Table.levels[index] {\n g.research("level",index,j,int(boolInt(r.contains(g.Ball.PixelX+8,g.Ball.PixelY+8+g.ScreenOffset))))')],
      'internal/physics/collision.go':[
       ('if !bit(m, x, y) {','g.research("ring",int(boolInt(b.High)),int(x),int(y),i,int(boolInt(bit(m,x,y))))\n if !bit(m, x, y) {'),
       ('m := g.Table.Materials[c.material]','g.research("response",int(c.material),int(c.angle),int(c.count),int(c.xadd),int(c.yadd))\n m := g.Table.Materials[c.material]'),
       ('s, co := g.Table.Sin[inv],','g.research("sine",int(inv),int(inv)+512)\n s, co := g.Table.Sin[inv],'),
       ('s, co = g.Table.Sin[c.angle],','g.research("sine",int(c.angle),int(c.angle)+512)\n s, co = g.Table.Sin[c.angle],')],
      'internal/partyland/game.go':[
       ('type Game struct {','type Game struct {\n ResearchObserve func(string)'),
       ('func (g *Game) Release(charge, jitter uint8) {','func (g *Game) Release(charge, jitter uint8) {\n if g.ResearchObserve!=nil {g.ResearchObserve("spring_release")}'),
       ('case physics.EventTargetHit:','case physics.EventTargetHit:\n if g.ResearchObserve!=nil {g.ResearchObserve("target_callback")}')],
      'internal/partyland/regions.go':[
       ('g.emit("Switch", label, 0)','if g.ResearchObserve!=nil {g.ResearchObserve(label)}\n g.emit("Switch", label, 0)')],
      'internal/partyland/flow.go':[
       ('func (g *Game) drain() {','func (g *Game) drain() {\n if g.ResearchObserve!=nil {g.ResearchObserve("drain_entry")}'),
       ('g.effect("LOSTBALL", 0, 0)','if g.ResearchObserve!=nil {g.ResearchObserve("LOSTBALL_request")}\n g.effect("LOSTBALL", 0, 0)\n if g.ResearchObserve!=nil {g.ResearchObserve("LOSTBALL_result")}')],
    }
    for name,pairs in patches.items():
        for old,new in pairs:replace_once(harness/name,old,new)
    (harness/'internal/physics/research_observer.go').write_text(
      'package physics\nfunc boolInt(b bool) int {if b{return 1};return 0}\n'
      'func(g *Game) research(role string,v ...int){if g.ResearchTrace!=nil {g.ResearchTrace(role,v...)}}\n')
    (harness/'internal/partyland/dmo0_deterministic_replay_test.go').write_bytes(TEMPLATE.read_bytes())
    return dict(base=BASE,tracked_go_module_files=len(copied),all_base_files_equal=True,research_tests_added_after_base=research_tests,
                construction='partyland.New(DecodePartyLand(A), A); Configure(settings.Legacy())',
                input_only=True,observers='passive reads in isolated copied sources; no production edits')


def compare_instructions(d,a,b,rows=None):
    result=[]
    for r in rows or json.loads(MAP.read_text()):
        al,ah=r['A_extent'];bl,bh=r['demo_extent'];role=r['role']
        xs,ys=linear(d,a,al,ah,768),linear(d,b,bl,bh,768)
        require(len(xs)==len(ys),role+' instruction count')
        reloc={(at,i,kind):(old,new) for at,i,kind,old,new in r['relocated_operands']}
        consumed=set()
        for x,y in zip(xs,ys):
            require((x.mnemonic,x.size,len(x.operands))==(y.mnemonic,y.size,len(y.operands)),role+' operation')
            for i,(u,v) in enumerate(zip(x.operands,y.operands)):
                require((u.type,u.size,u.access)==(v.type,v.size,v.access),role+' operand')
                if u.type==d.x86.X86_OP_REG:require(u.reg==v.reg,role+' register')
                elif u.type in (d.x86.X86_OP_MEM,d.x86.X86_OP_IMM):
                    kind='memory' if u.type==d.x86.X86_OP_MEM else 'immediate'
                    if kind=='memory':
                        p,q=u.mem,v.mem
                        require((p.segment,p.base,p.index,p.scale)==(q.segment,q.base,q.index,q.scale),role+' addressing')
                        av,bv=p.disp,q.disp
                    else:av,bv=u.imm,v.imm
                    key=(x.address+768,i,kind)
                    if key in reloc:
                        old,new=reloc[key];require(av==old and bv==new,role+' reviewed relocation');consumed.add(key)
                    else:require(av==bv,role+' unmapped difference')
                else:raise ValueError(role+' unsupported operand')
        require(consumed==set(reloc),role+' unused relocation')
        result.append(dict(role=role,A_extent=[al,ah],demo_extent=[bl,bh],instructions=len(xs),
                           classification='RELOCATED-IDENTICAL',scope='bounded consumer correspondence, not arbitrary caller/alias closure'))
    return result


def record(a,b,role,al,bl,size,**metadata):
    require(a[al:al+size]==b[bl:bl+size],role+' data difference')
    return dict(role=role,A_file=al,demo_file=bl,size=size,classification='RELOCATED-IDENTICAL',**metadata)


def consumed_projection(d,a,b,replay):
    w=replay['witness'];checkpoints=w['Checkpoints']
    rows=compare_instructions(d,a,b)
    rows += [compare_block(d,a,b,n) for n in ('BYGEL1_unlit','BYGEL2_unlit','drain_selection','scored_LOSTBALL_request')]
    reads=[v for block in w['PhysicsReads'] for v in (block or [])]
    materials=sorted({r[1] for r in reads if r[0]=='response'})
    for i in materials:
        rows.append(record(a,b,'material',0x1c05d+16*i,0x1c1c7+16*i,10,index=i,
          values=list(struct.unpack_from('<5h',a,0x1c05d+16*i))))
    rows.append(record(a,b,'sine_lookup',0x1e340,0x1e4b0,5120,
      consumed_indices=sorted({i for r in reads if r[0]=='sine' for i in r[1:]})))
    rows.append(record(a,b,'XY_LIST',0x20520,0x20690,176,
      note='44 coordinate pairs; checkpoint angles/order also paired in sc_krock'))
    rows.append(record(a,b,'physics_velocity_spin_push_constants',0x19d40+0x68a2,0x19db0+0x69a2,14))
    rows.append(record(a,b,'high_resolution_ramp_records',0x19d40+0x72,0x19db0+0x72,16,
      setting='accepted Legacy low angle: paired TABLE_ANGLE subtracts 3 from each Y word; Configure uses the same four results'))
    for i,(off,size) in enumerate([(0x68b0,30),(0x68f0,30),(0x6940,15)]):
        rows.append(record(a,b,'fresh_duck_up_mask_patch',0x19d40+off,0x19db0+off+0x100,size,index=i))
    # All six sources include negative bitmap tests, material and ramp reads.
    # Dynamic duck/gate initialization and flipper OR/copy consumers are paired.
    for role,al,size in [('mask12',0x3b930,23040),('mask11',0x41330,23040),
       ('mask22',0x46d30,23040),('mask13',0x71b30,23040),
       ('mask21',0x77530,20400),('mask23',0x7cf30,20400)]:
        rows.append(record(a,b,role,al,al+0x170,size,
          note='complete bounded bitmap equality covers every actual byte read; dynamic copy frames compared separately'))
    for i,(al,size) in enumerate([(0x4c730,8904),(0x4fe10,8904),(0x4ed50,4284)]):
        rows.append(record(a,b,'flipper_collision_frames',al,al+0x170,size,index=i,
          consumed_frames=sorted({r[2] for r in reads if r[0]=='flipper_copy' and r[1]==i})))
        rows.append(record(a,b,'flipper_numeric_descriptor',0x20690+60*i,0x20800+60*i,42,index=i,
          note='type, bounds, centers, angle/speed/frame and acceleration; relocated pointer fields excluded'))
    rows.append(record(a,b,'high_level_regions',0x1ac05,0x1ac91,82))
    rows.append(record(a,b,'low_level_regions',0x1ac57,0x1ace3,42))
    for i in range(4):rows.append(record(a,b,'target_test_rectangle',0x1aab1+10*i,0x1ab3d+10*i,8,index=i))
    for plane,al,bl,count in [('lower',0x1aadb,0x1ab67,14),('upper',0x1ab69,0x1abf5,12)]:
        for i in range(count):rows.append(record(a,b,'area_test_rectangle',al+10*i,bl+10*i,8,plane=plane,index=i))
    for role,al,bl in [('BYGEL11',0x1aadb,0x1ab67),('BYGEL9',0x1aae5,0x1ab71),
       ('CLOSE1',0x1ab17,0x1aba3),('BYGEL1',0x1ab3f,0x1abcb),
       ('BYGEL2',0x1ab49,0x1abd5),('BYGEL12',0x1ab53,0x1abdf),('BYGEL28',0x1ab5d,0x1abe9),
       ('TOUCHER',0x1aab1,0x1ab3d)]:
        av,bv=struct.unpack_from('<5H',a,al),struct.unpack_from('<5H',b,bl)
        require(av[:4]==bv[:4],role+' binding geometry')
        bindings={'BYGEL11':(0x19b6,0x19ba),'BYGEL9':(0x1baf,0x1bb3),
          'CLOSE1':(0x2564,0x2568),'BYGEL1':(0x27ed,0x27f6),'BYGEL2':(0x2872,0x287b),
          'BYGEL12':(0x28bd,0x28c6),'BYGEL28':(0x28d0,0x28d9),'TOUCHER':(0x164c,0x1650)}
        require((av[4]+768,bv[4]+768)==bindings[role],role+' callback binding')
        rows.append(dict(role=role+'_binding',A_file=al,demo_file=bl,
          A_callback=av[4]+768,demo_callback=bv[4]+768,classification='RELOCATED-IDENTICAL'))
    # Additional demo store at the same CLOSE1 boundary, before its paired RET.
    d.expect(b,768,0x2596,'mov','byte ptr [0xd1], 0')
    rows.append(dict(role='CLOSE1 extra VISAKEYS=false store',demo_extent=[0x2596,0x259b],
      classification='DEMO-DIFFERENT-BUT-IRRELEVANT',reason='VISAKEYS does not feed ball, spring, flipper, geometry, score, aggregates or this drain admission; deferred NEW_BALL_TASK consumer remains outside this pass'))
    for label in ('BYGEL9','BYGEL11'):
        cp=next(c for c in checkpoints if c['boundary']==label)
        if label=='BYGEL9':require(cp['last_area']=='CLOSE1','BYGEL9 actual no-loop guard')
        else:require(cp['last_area']=='BYGEL9' and cp['inhibit_reverse_time']>0,'BYGEL11 actual inhibited reverse guard')
    rows.append(dict(role='demo expiry guard before scored drain',classification='DEMO-DIFFERENT-BUT-IRRELEVANT',
      reason=f"concrete drain {w['Drain']} precedes inherited first equality 35998; no expiry effect is admitted on this prefix"))
    rows.append(dict(role='native bitmap bounds policy',classification='VALUE-EQUIVALENT',
      observed_out_of_bounds_ring_samples=sum(1 for r in reads if r[0]=='ring' and not 0<=r[3]*40+(r[2]>>3)<23040),
      reason='accepted canonical-A native bit() treats out-of-map samples as empty; demo projection inherits the same bounded-map consumer and equal map dimensions. This is a native-reference witness, not a physical DOS adjacent-memory trajectory oracle. No historical phase or whole physics closure follows.'))
    # Explicit linked LOSTBALL record/program and admission assertions.
    handlers,_=identities(b,Path(os.environ['PF_DMO0_HISTORICAL_SOURCE']))
    scored_linked(b,d,handlers)
    require(struct.unpack_from('<H',b,0x19db0+0x6d5+26)[0]+0x19db0==0x1b459,'normal bonus program binding')
    rows.append(dict(role='LOSTBALL admission and normal bonus binding',classification='VALUE-EQUIVALENT',
      demo_effect_file=0x19db0+0x6d5,program=0x1b459,
      proof='paired effect consumers; linked priority/admission and score-only record checks; actual uninhibited non-special admitted native request'))
    return rows


def validate_replay(r):
    w=r['witness'];require(w['Valid'],'invalid native witness')
    require((w['Release'],w['Bygel'],w['Drain'])==(35460,35790,35877),'exact indices')
    next_index=1
    for row in w['Rows']:
        require(row['start']==next_index and row['end']>=row['start'],'trace coverage')
        next_index=row['end']+1;s=row['state']
        require(s['totals']==[0,0,0,0] and not s['XXBALLE'],'zero aggregate/XXBALLE prefix')
    require(next_index==35878,'complete trace to drain')
    cp={c['boundary']:c for c in w['Checkpoints']}
    require(not cp['BYGEL1']['light39'] and cp['BYGEL1']['score']==0,'unlit score-only entry')
    require(cp['spring_release']['spring']==22,'actual spring release charge')
    for label in ('drain_entry','LOSTBALL_request','LOSTBALL_result'):
        s=cp[label];require(s['score']==50030 and s['SCORECHANGED'],'scored drain origin')
        require(not s['INH_EFF'] and not s['SPECIALMODE'],'LOSTBALL state')
    require(cp['LOSTBALL_result']['effect_accepted'],'LOSTBALL admission')
    require(not r['mutation']['Valid'] or (r['mutation']['Bygel'],r['mutation']['Drain'])!=(35790,35877),'input mutation')
    return True


def main():
    p=argparse.ArgumentParser();p.add_argument('--harness',type=Path,default=Path('/private/tmp/pf-bygel-deterministic-harness'))
    p.add_argument('--output',type=Path,default=Path('/private/tmp/pf-dmo0-deterministic-bygel-drain.json'))
    p.add_argument('--go',type=Path,default=ROOT/'.tools/go/bin/go');args=p.parse_args()
    data,canonical,historical=[Path(os.environ[n]) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
    demo,full=pinned(data,canonical,historical)
    entry=entry_audit(data,canonical,historical)
    require(entry['verdict']=='CANONICAL_A_JITTER_INHERITANCE = PROVED','fresh reference premise')
    identity=prepare(args.harness,canonical)
    script=args.harness/'script.json';script.write_text(json.dumps(TARGET))
    replay_path=args.harness/'replay.json';variation_path=args.harness/'variations.json'
    env=os.environ.copy();env.update(PF_REPLAY_SCRIPT=str(script),PF_REPLAY_OUTPUT=str(replay_path),
      PF_REPLAY_VARIATIONS=str(variation_path),GOCACHE='/private/tmp/pf-dmo0-go-cache')
    run=subprocess.run([str(args.go),'test','./internal/partyland','-run','^TestResearch(DeterministicReplay|DelayAndHold)$','-count=1','-v'],cwd=args.harness,env=env,capture_output=True,text=True)
    (args.harness/'tests.log').write_text(run.stdout+run.stderr);require(run.returncode==0,'deterministic harness tests failed; inspect owner-local tests.log')
    replay=json.loads(replay_path.read_text());validate_replay(replay)
    projection=consumed_projection(Decoder(),full['TABLE1.PRG'],demo['TABLE1.PRG'],replay)
    variations=json.loads(variation_path.read_text())
    # Preserve the first simple witness as an independently traced fresh replay.
    simple_path=args.harness/'simple-replay.json';simple_script=args.harness/'simple-script.json'
    simple_script.write_text(json.dumps(dict(Release=61,RightEnd=1000)))
    env.update(PF_REPLAY_SCRIPT=str(simple_script),PF_REPLAY_OUTPUT=str(simple_path))
    simple_run=subprocess.run([str(args.go),'test','./internal/partyland','-run','^TestResearchDeterministicReplay$','-count=1','-v'],cwd=args.harness,env=env,capture_output=True,text=True)
    (args.harness/'simple-tests.log').write_text(simple_run.stdout+simple_run.stderr)
    require(simple_run.returncode==0,'simple replay/mutation failed')
    simple=json.loads(simple_path.read_text());sw=simple['witness']
    require(sw['Valid'] and (sw['Release'],sw['Bygel'],sw['Drain'])==(61,255,349),'first witness drift')
    simple_projection=consumed_projection(Decoder(),full['TABLE1.PRG'],demo['TABLE1.PRG'],simple)
    suffix=json.loads(Path('/private/tmp/pf-dmo0-new-ball-threshold-provenance.json').read_text())
    require(suffix['producer_calculation_offset']['admitted_matrix_visits']==91,'existing suffix premise')
    report=dict(verdict='FRESH_BYGEL_DRAIN_PROVENANCE = PROVED',drain_verdict='DRAIN_35877_REACHABLE',
      SCORED_DRAIN_35877_PROVENANCE='PROVED',status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',
      fresh_start_identity=identity,input_script=TARGET,
      input_semantics='wait with no inputs; Down on [35438,35459]; release edge at 35460; d=n-35460>=0: Left iff (d+46)%52<8, Right iff (d+22)%30<21; all other controls false',
      deterministic_replay=replay,path_consumed_correspondence=projection,release_delay_tests=variations,
      first_simple_witness=simple,first_simple_correspondence=simple_projection,
      expiration=dict(at_drain=False,before_pre_electronics_check=True,first_equality=35998),
      LOSTBALL=dict(admitted=True,normal_demo_bonus_program=0x1b459),
      search_coverage=dict(no_flippers=dict(release=[2,512],scripts=511,limit_after_release=2200),
        fixed_or_periodic_flippers=dict(release=[18,160],modes=12,scripts=1716,limit_after_release=2200),
        late_constant_right=dict(release=[35200,35800],scripts=601,limit_after_release=2200),
        right_intervals=dict(release=[33,160],starts=list(range(0,181,5)),durations=[10,20,40,80,160,1000],scripts=28416,limit_after_release=1400),
        charge_and_right_intervals=dict(release=[33,160],charge=[18,32],starts=list(range(0,101,10)),durations=[20,80,1000],scripts=63360,limit_after_release=1400),
        generated_periodic_controls=dict(generator='Go math/rand.NewSource(1); release 33+Intn(128), charge 18+Intn(15), LP/RP 20+Intn(200), LD/RD/LO/RO bounded by periods',trials=8610,valid_scored_BYGEL_drains=157,limit_after_release=1499,invalid_aggregate_prefixes_pruned=True),
        phase_use='128-period arithmetic proposed one candidate; full fresh long replay verified directly; no state periodicity reduction or exclusion theorem'),
      neighbors={'35876':'UNKNOWN: direct release variations did not land on this index','35877':'REACHABLE: concrete projected replay','35878':'UNKNOWN: direct release variations did not land on this index'},
      reused_suffix=dict(visits=91,producer_after=35967,conditional_NEW_BALL_TASK=35998,
        premise='existing zero aggregate, XXBALLE=false branch; arithmetic not rediscovered',
        NEW_BALL_TASK_collision_reachability='NOT_PROVED',deferred=['actual task slot','DS:0x36cd initial/unique age','task survival','PARTYFLASH','VISAKEYS']),
      smallest_unresolved_fact=None,tests=dict(new_pass='PASS',OriginalTrajectories='NOT AVAILABLE',
        PF6SessionKeepsGameplayOracle='known reproduced baseline: 2311040 versus 2300000; not trajectory evidence'))
    args.output.write_text(json.dumps(report,separators=(',',':'))+'\n')
    print(report['drain_verdict']);print(report['verdict'])
    return 0


if __name__=='__main__':raise SystemExit(main())
