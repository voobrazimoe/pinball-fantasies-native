#!/usr/bin/env python3
"""Fresh input-only PARTY_ON first-equality witness; private metadata only."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import struct
import audit_10min_demo_post_collision as post
import audit_10min_demo_deterministic_replay as replay
import audit_10min_demo_expiry_interleaving as expiry
from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_programs import identities
from audit_10min_demo_source_recurrence import linear
ROOT=replay.ROOT
OUTPUT=Path('/private/tmp/pf-dmo0-first-equality-party-on.json')
ANCHORS=(
 (0x46b8,'call','0x5a1f'),(0x46c5,'je','0x4414'),
 (0x5d36,'cmp','byte ptr [0x2498], 0xff'),(0x5d40,'cmp','byte ptr [0x34ca], 0xff'),(0x5d4a,'call','0x215'),
 (0x4720,'call','0x2ab5'),(0x4723,'call','0x59d9'),
 (0x592,'cmp','byte ptr [0x34dd], 0'),(0x597,'jne','0x2d2'),
 (0x59c,'mov','byte ptr [0x3485], 0'),(0x5a2,'mov','bx, 0x1493'),(0x5a5,'call','0x4501'),
 (0x5a8,'mov','si, 0xc8b'),(0x5ab,'call','0x5cb4'),(0x5ae,'mov','byte ptr [0xd1], 0xff'),
 (0x5b3,'mov','dx, 0x2ba'),(0x5b6,'call','0x5b80'),(0x5b9,'ret',''),
 (0x5c3,'jb','0x2c9'),(0x5c8,'ret',''),(0x5d1,'ret',''),
 (0x5e83,'cmp','word ptr [bx], 0x6a71'),(0x5e87,'je','0x5b98'),
 (0x5e8c,'add','bx, 2'),(0x5e8f,'cmp','bx, 0x347b'),(0x5e93,'jne','0x5b83'),
 (0x5e9f,'mov','cx, 0x32'),(0x5eaf,'add','bx, 2'),
 (0x3b03,'je','0x3841'),(0x3b41,'ret',''),
 (0x2efc,'mov','byte ptr [0xd0], 0xff'),(0x2f33,'mov','byte ptr [0xd0], 0xff'),(0x561d,'ret',''),
)

def target_drain(first_visit_same=True,age=30,fire=35998):
    return fire-age-(0 if first_visit_same else 1)

def linked(b,full,d,h):
    expiry.linked(b,d)
    for at,op,args in ANCHORS:d.expect(b,768,at,op,args)
    from audit_10min_demo_fresh_bygel_drain import compare_block
    drain=compare_block(d,full,b,'drain_selection')
    correspondence=replay.compare_instructions(d,full,b)
    for i in (2,3,6,7):
        require(full[0x1c05d+16*i:0x1c05d+16*i+10]==b[0x1c1c7+16*i:0x1c1c7+16*i+10],'consumed material')
    require(full[0x1e340:0x1e340+5120]==b[0x1e4b0:0x1e4b0+5120],'consumed sine table')
    party=expiry.program(b,h,0x1b243)
    require([x['op'] for x in party]==['_CLEAR4','_FLASHON','_PARTYONN','_PRINT13','_PARTYON'],'party program')
    require([x['args'] for x in party]==[[],[3],[1],[7647,336],[1]],'party operands')
    # Only concrete pre-fire consumers: the wait and surviving PARTYRUT.
    writes=[]
    for name,lo,hi in [('PARTY_ON_TASK1',0x5ba,0x5d2),('WAITSYNCS',0x5ac7,0x5ad8),('PARTYRUT',0x561d,0x561e),('_PARTYONN',0x2ef8,0x2f18),('_PARTYON',0x2f33,0x2f47)]:
        for x in linear(d,b,lo,hi,768):
            for o in x.operands:
                if o.type==d.x86.X86_OP_MEM and o.access&d.cs.CS_AC_WRITE:
                    m=o.mem
                    if m.base or m.index:
                        require(name=='WAITSYNCS' and (x.mnemonic,x.op_str) in (('inc','word ptr [bx]'),('mov','word ptr [bx], 0')),'unknown concrete indexed writer')
                        target=0x36c9
                    else:target=m.disp
                    writes.append(dict(consumer=name,site=x.address+768,target=target,width=o.size))
    return dict(anchors=len(ANCHORS),drain_correspondence=drain,consumed_path_correspondence=correspondence,consumed_material_indices=[2,3,6,7],sine_table_equal=True,party_program=party,concrete_consumer_writes=writes,
        ordering={'physics_call':0x46b8,'drain_call':0x5d4a,'rest_branch':0x46c5,'UPDATE_COUNTERS':0x4720,'DO_ELECTRONICS':0x4723,'DO_TASKS':0x5d16,'first_free':0x5e80,'scan_entry':0x5e9f,'scan_call':0x5eab,'scan_next':0x5eaf,'scan_loop':0x5eb2},
        writer_scope='Concrete suffix only: lost held ball, KEYTASK DUMRET, sole PARTY_ON_TASK1; native common counters/display plus linked PARTY_ON ops and wait. No global writer/IRQ closure.')

def prepare(harness,canonical,cfg):
    (harness/'internal/partyland/dmo0_party_on_test.go').unlink(missing_ok=True)
    identity=post.prepare(harness,canonical,cfg)
    p=harness/'internal/partyland/game.go'
    replay.replace_once(p,' ResearchDemoEnabled,',' ResearchPartyMutation string\n ResearchDemoEnabled,')
    p=harness/'internal/partyland/timing.go'
    replay.replace_once(p,'if !g.waitReady(site, n) {','if g.ResearchObserve!=nil {g.ResearchObserve("wait_before:"+site)}\n if !g.waitReady(site, n) {\n if g.ResearchObserve!=nil {g.ResearchObserve("wait_after:"+site)}')
    replay.replace_once(p,'\t\tf()\n if g.ResearchAfterWait','\t\tf()\n if g.ResearchObserve!=nil {g.ResearchObserve("body_after:"+site)}\n if g.ResearchAfterWait')
    p=harness/'internal/partyland/flow.go'
    replay.replace_once(p,'g.waitAt("PARTY_ON_TASK1", 30, func(){g.partyFlash=true;g.newBall()})',
        'g.waitAt("PARTY_ON_TASK1", 30, func(){if g.ResearchPartyMutation!="delay_party" {g.partyFlash=true}; if g.ResearchObserve!=nil {g.ResearchObserve("party_guard")};g.newBall()})\n if g.ResearchObserve!=nil {g.ResearchObserve("party_admitted")}')
    replay.replace_once(p,'\tg.resetBall()','if g.ResearchObserve!=nil {g.ResearchObserve("before_reset")}\n g.resetBall()\n if g.ResearchObserve!=nil {g.ResearchObserve("after_reset")}')
    p=harness/'internal/partyland/research_post_overlay.go'
    replay.replace_once(p,'g.emit("ResearchExpiryInstalled",reason,uint64(g.ResearchTimer))','g.emit("ResearchExpiryInstalled",reason,uint64(g.ResearchTimer))\n if g.ResearchObserve!=nil {g.ResearchObserve("expiry_installed")}')
    (harness/'internal/partyland/dmo0_party_on_test.go').write_bytes((ROOT/'tools/dmo0_party_on_test.go.txt').read_bytes())
    return identity

def boundary(w,role,n):
    return next(x for x in w['boundaries'] if x['boundary']==role and x['calculation']==n)

def matrix_state(x,nodes,entry):
    # Native command-array indices count commands; linked cursors count bytes.
    if entry<=x['matrix_pc']<=entry+len(nodes):
        k=x['matrix_pc']-entry
        if k>0:
            q=nodes[k-1]
            return dict(program=nodes[0]['site'],cursor=q['site']+2*(1+len(q['args'])),op=x['matrix_op'],remaining=x['matrix_remaining'])
    return dict(native_pc=x['matrix_pc'],op=x['matrix_op'],remaining=x['matrix_remaining'])

def validate(r):
    corrected = r.get('timing_oracle_basis') == 'FF-inclusive-v1'
    w=r['replay']['witness'];bs=w['boundaries'];rows={x['calculation']:x for x in w['rows']}
    require(w['drain']==target_drain() and not w['scored'],'target drain')
    require(w['prefix_score_false'] and w['prefix_reset_one'],'fresh prefix')
    require(r['replay']['deterministic_equal'],'determinism')
    drain=boundary(w,'drain_entry',35968);admit=boundary(w,'party_admitted',35968)
    require(not drain['SCORECHANGED'] and not drain['expired'] and drain['timer']==35967,'drain old timer')
    require(admit['live_task_sites']==['PARTY_ON_TASK1']+['']*49,'actual slot')
    require(admit['waits'].get('PARTY_ON_TASK1',0)==0,'initial word')
    for n in range(35968,35998):
        before=[x for x in bs if x['calculation']==n and x['boundary']=='wait_before:PARTY_ON_TASK1']
        after=[x for x in bs if x['calculation']==n and x['boundary']=='wait_after:PARTY_ON_TASK1']
        require(len(before)==len(after)==1,'exact visit count')
        require(before[0]['waits'].get('PARTY_ON_TASK1',0)==n-35968 and after[0]['waits']['PARTY_ON_TASK1']==n-35967,'age progression')
        require(rows[n]['live_task_sites']==['PARTY_ON_TASK1']+['']*49 and rows[n]['reset_count']==1,'survival/duplicate')
        require(not rows[n]['VISAKEYS'],'visa')
    installed=boundary(w,'expiry_installed',35998);fire=boundary(w,'party_guard',35998)
    require(installed['expired'] and installed['ball']['Hold'] and installed['timer']==35998,'expiry admission')
    require(installed['waits']['PARTY_ON_TASK1']==30,'fire age')
    scan=boundary(w,'wait_before:PARTY_ON_TASK1',35998)
    require(scan['waits'].get('PARTY_ON_TASK1',0)==30 and scan['live_task_sites']==['PARTY_ON_TASK1']+['']*49,'equality scan age/slot')
    order=[x['boundary'] for x in bs if x['calculation']==35998]
    require(order.index('expiry_installed')<order.index('wait_before:PARTY_ON_TASK1')<order.index('party_guard')<order.index('before_reset')<order.index('after_reset'),'equality/body ordering')
    require(fire['partyflash'] and fire['waits']['PARTY_ON_TASK1']==0,'flag before reset')
    for role in ('before_reset','after_reset','body_after:PARTY_ON_TASK1'):
        x=boundary(w,role,35998)
        require((x['matrix_pc'],x['matrix_remaining'],x['matrix_op'])==(installed['matrix_pc'],5,'_CLEAR4'),'expiry cursor preservation')
    body=boundary(w,'body_after:PARTY_ON_TASK1',35998)
    require(body['live_task_sites'][:3]==['SOUNDNEWBALL','SETBALL','SOUNDBRICKUPP'],'queued tasks')
    require(rows[35998]['waits']=={'SETBALL':1,'SOUNDBRICKUPP':1},'remaining scan')
    require(w['final']['QUIT'] and w['final']['calculation']==(37047 if corrected else 37039),'concrete quit')
    for v,dr,fi in zip(r['replay']['neighbors'],(35967,35969),(35997,35999)):
        require(v['drain']==dr and not v['scored'] and boundary(v,'party_guard',fi),'neighbors')
    clear=r['replay']['clear_party'];require(boundary(clear,'after_reset',35998)['matrix_pc']!=installed['matrix_pc'],'flag clear mutation')
    for k in ('delay_party','clear_before_body'):
        require(boundary(r['replay'][k],'body_after:PARTY_ON_TASK1',35998)['matrix_pc']==installed['matrix_pc'],'redundant/restored party flag')
    return True

def report(raw,identity,evidence,nodes):
    w=raw['witness'];party=evidence['party_program'];entry=boundary(w,'expiry_installed',35998)['matrix_pc']-1
    # Attach exact linked program/cursor to every suffix boundary and row.
    for v in [w,*raw['neighbors'],raw['release_plus_one'],raw['clear_party'],raw['delay_party'],raw['clear_before_body']]:
        for x in v['rows']+v['boundaries']:
            x['linked_matrix']=matrix_state(x,nodes,entry) if x['matrix_pc']>=entry and x['matrix_pc']<=entry+len(nodes) else matrix_state(x,party,55)
    timeline=[]
    for n in range(35968,35999):
        x=boundary(w,'wait_before:PARTY_ON_TASK1',n)
        timeline.append(dict(calculation=n,slot=0,DS_0x36c9_before=x['waits'].get('PARTY_ON_TASK1',0),after=0 if n==35998 else n-35967,fire=n==35998,PARTYFLASH=x['partyflash'],VISAKEYS=x['VISAKEYS'],matrix=x['linked_matrix'],live_tasks=x['live_task_sites']))
    ops=[];prior=None
    for x in w['rows']:
        if x['calculation']>=35998 and x['matrix_op']!=prior:
            ops.append(dict(calculation=x['calculation'],matrix=x['linked_matrix'],text_left=x['matrix_text_left']));prior=x['matrix_op']
    r=dict(timing_oracle_basis='FF-inclusive-v1',reachability='FIRST_EQUALITY_PARTY_ON_COLLISION_REACHABLE',expiry_interleaving='EXPIRY_INTERLEAVING = NOT_PROVED',status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',
        production_base=replay.BASE,research_head='37ff8bc38d7af4a09a70319675d6e505b0bd6a5e',native_reference_identity=identity,linked_evidence=evidence,
        target_drain_arithmetic=dict(first_visit='same calculation; insertion before electronics, first-free slot0, scan starts slot0',D=target_drain(),first_visit_calculation=35968,after_35997_age=30,fires=35998,off_by_one={'35967':35997,'35969':35999},next_update_hypothesis_D=target_drain(False)),
        input_script=dict(Down=[1,35841],Release=35842,all_other_controls=False,after_release='all controls false through QUIT'),
        drain_checkpoint=boundary(w,'drain_entry',35968),task_admission=boundary(w,'party_admitted',35968),task_slot=dict(index=0,DS=0x3417,task_file=0x5ba,task_pointer=0x2ba,wait_DS=0x36c9,initial_age=0,duplicates=1),DS_0x36c9_timeline=timeline,replay=raw,
        guard_provenance='VISAKEYS=false is the accepted linked fresh WHEN_NEW_GAME_RESET -> initial NEW_BALL projection, not a native Go field. Fresh prefix reset_count=1; no F1/new-game input or visa writer on concrete suffix. MUSICOK DS:d1 is distinct. PARTYFLASH is native-observed: false at drain; _PARTYONN sets true35973; _PARTYON true35975; task body again sets true35998.',
        admission='Unscored branch directly calls DO_MATRIX(PARTY_ONTS), PLAYJINGLE(S_SPRING), then ADDTASK; no generic effect admission gate. MatrixStarted event, real op/cursor and slot prove admission. effect_accepted is stale and is not used as proof.',
        concrete_writer_result='DS:36c9 starts at zero after initial WAITLIST reset; first-free insertion does not write it. Exactly one sole PARTY task wait visit per suffix scan. No other tasks, geometry callbacks, reset, task clear or duplicate through35997. Linked PARTYRUT is RET and preceding PARTY matrix ops have no36c9 writer. At35998 compare-match writes zero, then NEW_BALL clears WAITLIST/tasklist.',
        expiry_install=boundary(w,'expiry_installed',35998),party_firing=boundary(w,'party_guard',35998),reset_before=boundary(w,'before_reset',35998),reset_after=boundary(w,'after_reset',35998),queued_tasks=boundary(w,'body_after:PARTY_ON_TASK1',35998),
        matrix_outcome=dict(program=0x1ba17,cursor=0x1ba19,remaining_before_first_visit=5,remaining_after_first_visit=4,replaced=False,restarted=False,preserved=True,reason='PARTYFLASH=true takes reset JE directly to RET before VISAKEYS/SHOWPLAYERSTS; task/wait reset precedes guard'),
        expiry_progression=ops,eventual_result=dict(QUIT=True,calculation=37047,visits=1050,SETBALL=36078,subsequent_input='none',additional_drains=0,scope='this concrete continuation; no universal termination inference'),
        remaining_first_equality_classes=[dict(name='NEW_BALL_TASK',reachability='previously PROVED',PARTYFLASH=False,VISAKEYS=False,outcome='SHOWPLAYERSTS replaces expiry before first matrix visit'),dict(name='PARTY_ON_TASK1',reachability='PROVED this pass',PARTYFLASH=True,VISAKEYS=False,outcome='preserves expiry cursor; this input continuation QUIT37047'),dict(name='SETBALL at equality',reachability='UNKNOWN; no independently reachable evidence asserted')],
        exactly_one_remaining_dependency='Reachability or exclusion and effect classification of SETBALL clearing HOLDSTILL on the first-equality calculation with expiry current; current first-equality class coverage is not claimed complete.',
        search_scope='Finite deterministic input search; no impossibility inference. First found at release35842, charge0, no flippers.',tests={})
    validate(r);return r

def main():
    p=argparse.ArgumentParser();p.add_argument('--harness',type=Path,default=Path('/private/tmp/pf-party-on-reference'));p.add_argument('--replay',type=Path);a=p.parse_args()
    paths=[Path(os.environ[n]) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
    demo,full=pinned(*paths);b=demo['TABLE1.PRG'];d=Decoder();h,_=identities(b,paths[2]);evidence=linked(b,full['TABLE1.PRG'],d,h);cfg,nodes=post.configuration(b,h)
    if a.replay:
        raw=json.loads(a.replay.read_text());identity=json.loads((a.harness/'identity.json').read_text())
    else:
        identity=prepare(a.harness,paths[1],cfg)
        env=os.environ.copy();env.update(PF_POST_CONFIG=str(a.harness/'post-config.json'),PF_PARTY_OUTPUT=str(a.harness/'party-output.json'),GOCACHE='/private/tmp/pf-dmo0-go-cache')
        run=subprocess.run([str(ROOT/'.tools/go/bin/go'),'test','./internal/partyland','-run','^TestResearchPartySearch$','-count=1','-v'],cwd=a.harness,env=env,capture_output=True,text=True)
        (a.harness/'party-tests.log').write_text(run.stdout+run.stderr);require(run.returncode==0,'fresh party replay failed; see isolated party-tests.log')
        raw=json.loads((a.harness/'party-output.json').read_text())
    r=report(raw,identity,evidence,nodes);OUTPUT.write_text(json.dumps(r,separators=(',',':'))+'\n');print(r['reachability'])
if __name__=='__main__':main()
