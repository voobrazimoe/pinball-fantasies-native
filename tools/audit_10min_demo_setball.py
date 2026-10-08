#!/usr/bin/env python3
"""Input-only first-equality SETBALL witness in an isolated A reference."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import audit_10min_demo_party_on as party
import audit_10min_demo_post_collision as post
import audit_10min_demo_deterministic_replay as replay
from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_programs import identities
from audit_10min_demo_source_recurrence import linear
ROOT=replay.ROOT
OUTPUT=Path('/private/tmp/pf-dmo0-first-equality-setball.json')

def target_drain(party_limit=30,setball_limit=80,same_scan=True):
    return 35998-party_limit-setball_limit-(0 if same_scan else 1)

def linked(b,full,d,h):
    evidence=party.linked(b,full,d,h)
    evidence['post_consumers']=post.linked(b,full,d,h)
    from audit_10min_demo_fresh_bygel_drain import compare_block
    evidence['SETBALL_correspondence']=compare_block(d,full,b,'SETBALL')
    records=[]
    for role,at,size in [('mask12',0x3b930,23040),('mask11',0x41330,23040),('mask22',0x46d30,23040),('mask13',0x71b30,23040),('mask21',0x77530,20400),('mask23',0x7cf30,20400),('flipper0',0x4c730,8904),('flipper1',0x4fe10,8904),('flipper2',0x4ed50,4284),('XY_LIST',0x20520,176)]:
        records.append(replay.record(full,b,role,at,at+0x170,size))
    for role,aa,bb,size in [('physics_constants',0x19d40+0x68a2,0x19db0+0x69a2,14),('high_ramps',0x19d40+0x72,0x19db0+0x72,16),('high_levels',0x1ac05,0x1ac91,82),('low_levels',0x1ac57,0x1ace3,42)]:
        records.append(replay.record(full,b,role,aa,bb,size))
    for i in range(8):records.append(replay.record(full,b,'material',0x1c05d+16*i,0x1c1c7+16*i,10,index=i))
    for i,(off,size) in enumerate([(0x68b0,30),(0x68f0,30),(0x6940,15)]):records.append(replay.record(full,b,'fresh_duck_patch',0x19d40+off,0x19db0+off+0x100,size,index=i))
    for i in range(3):records.append(replay.record(full,b,'flipper_descriptor',0x20690+60*i,0x20800+60*i,42,index=i))
    evidence['consumed_physics_data']=records
    for at,op,args in ((0x1003,'mov','word ptr [0x2fdc], 0x129'),(0x1009,'mov','word ptr [0x2fde], 0x212'),(0x100f,'mov','byte ptr [0x3416], 0'),(0x1031,'mov','word ptr [0x2fea], 0'),(0x1037,'mov','word ptr [0x2fe8], 0xa'),(0x1043,'mov','word ptr [0x3481], 0xffff'),(0x1049,'jmp','0x57ba')):
        d.expect(b,768,at,op,args)
    evidence['SETBALL_stores']=[dict(site=x.address+768,op=x.mnemonic,operand=x.op_str) for x in linear(d,b,0xff4,0x104c,768)]
    return evidence

def prepare(harness,canonical,cfg):
    identity=party.prepare(harness,canonical,cfg)
    (harness/'internal/partyland/dmo0_party_on_test.go').unlink()
    p=harness/'internal/partyland/flow.go'
    replay.replace_once(p,'g.waitAt("SETBALL", 80,', 'g.waitAt("SETBALL", researchSetballLimit(g),')
    replay.replace_once(p,'g.Phase = Playing })','g.Phase = Playing; if g.ResearchPartyMutation=="matrix_replace" {g.beginMatrix("SHOWPLAYERSTS")} })')
    p=harness/'internal/partyland/game.go'
    replay.replace_once(p,'func (g *Game) runTasks()          { tablelogic.Run(g.tasks[:], g.taskIDs[:]) }','func (g *Game) runTasks() {if g.ResearchObserve!=nil {g.ResearchObserve("tasks_before")};tablelogic.Run(g.tasks[:],g.taskIDs[:]);if g.ResearchObserve!=nil {g.ResearchObserve("tasks_after")}}')
    p=harness/'internal/partyland/research_post_overlay.go'
    replay.replace_once(p,'g.ResearchLastCalculation=g.Tick;g.ResearchTimer++','if g.ResearchObserve!=nil {g.ResearchObserve("electronics_before")}\n g.ResearchLastCalculation=g.Tick;g.ResearchTimer++')
    replay.replace_once(p,'if g.ResearchTimer==35998 {researchInstallExpiry(g,"timer equality")}','if g.ResearchTimer==35998 {researchInstallExpiry(g,"timer equality")}\n if g.ResearchObserve!=nil {g.ResearchObserve("electronics_after")}')
    p=harness/'internal/partyland/presentation.go'
    replay.replace_once(p,'func (g *Game) presentationTick() { g.matrixTick() }','func (g *Game) presentationTick() {if g.ResearchObserve!=nil {g.ResearchObserve("matrix_before")};g.matrixTick();if g.ResearchObserve!=nil {g.ResearchObserve("matrix_after")}}')
    p=harness/'internal/physics/ball.go'
    replay.replace_once(p,'g.scroll()','g.research("late_before")\n g.scroll()')
    replay.replace_once(p,'g.Syncs++','g.research("late_after")\n g.Syncs++')
    (harness/'internal/partyland/dmo0_setball_test.go').write_bytes((ROOT/'tools/dmo0_setball_test.go.txt').read_bytes())
    return identity

boundary=party.boundary

def main():
    p=argparse.ArgumentParser();p.add_argument('--harness',type=Path,default=Path('/private/tmp/pf-setball-reference'));p.add_argument('--replay',type=Path);a=p.parse_args()
    paths=[Path(os.environ[n]) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
    demo,full=pinned(*paths);b=demo['TABLE1.PRG'];d=Decoder();h,_=identities(b,paths[2]);evidence=linked(b,full['TABLE1.PRG'],d,h);cfg,nodes=post.configuration(b,h)
    if a.replay:
        raw=json.loads(a.replay.read_text());identity=json.loads((a.harness/'identity.json').read_text())
    else:
        identity=prepare(a.harness,paths[1],cfg);(a.harness/'identity.json').write_text(json.dumps(identity))
        env=os.environ.copy();env.update(PF_POST_CONFIG=str(a.harness/'post-config.json'),PF_SETBALL_OUTPUT=str(a.harness/'setball-output.json'),GOCACHE='/private/tmp/pf-dmo0-go-cache')
        run=subprocess.run([str(ROOT/'.tools/go/bin/go'),'test','./internal/partyland','-run','^TestResearchSetballSearch$','-count=1','-v'],cwd=a.harness,env=env,capture_output=True,text=True)
        (a.harness/'setball-tests.log').write_text(run.stdout+run.stderr);require(run.returncode==0,'SETBALL replay failed; see isolated setball-tests.log')
        raw=json.loads((a.harness/'setball-output.json').read_text())
    r=report(raw,identity,evidence,nodes);OUTPUT.write_text(json.dumps(r,separators=(',',':'))+'\n');print(r['class_verdict'])

def firing_from_drain(D,party_limit=30,setball_limit=80):
    # Independent visit-by-visit arithmetic; no trajectory or reachability claim.
    age=0;n=D
    while True:
        age,fire=party.expiry.wait_visit(age,party_limit)
        if fire:break
        n+=1
    party_fire=n;age=0
    while True:
        age,fire=party.expiry.wait_visit(age,setball_limit)
        if fire:return party_fire,n
        n+=1

# Report/validation are intentionally shared with the tests below.
def validate(r):
    corrected = r.get('timing_oracle_basis') == 'FF-inclusive-v1'
    w=r['fresh_input_replay']['witness'];bs=w['boundaries'];rows={x['calculation']:x for x in w['rows']}
    b=lambda role,n:boundary(w,role,n)
    require(target_drain()==35888 and w['drain']==35888,'target off-by-one')
    require(r['fresh_input_replay']['deterministic_equal'] and w['prefix_score_false'] and w['prefix_reset_one'],'fresh deterministic prefix')
    x=b('drain_entry',35888)
    require(x['timer']==35887 and not x['expired'] and not x['SCORECHANGED'],'unscored old timer/expired')
    require(not any(x['live_task_sites']) and x['reset_count']==1,'drain task/reset origin')
    x=b('party_admitted',35888)
    require(x['live_task_sites']==['PARTY_ON_TASK1']+['']*49 and x['waits'].get('PARTY_ON_TASK1',0)==0,'party slot/initial word')
    for n in range(35888,35918):
        pre=b('wait_before:PARTY_ON_TASK1',n);after=b('wait_after:PARTY_ON_TASK1',n)
        require(pre['waits'].get('PARTY_ON_TASK1',0)==n-35888 and after['waits']['PARTY_ON_TASK1']==n-35887,'party wait progression')
        require(sum(x['calculation']==n and x['boundary']=='wait_before:PARTY_ON_TASK1' for x in bs)==1,'party once per scan')
    require(b('wait_before:PARTY_ON_TASK1',35918)['waits']['PARTY_ON_TASK1']==30 and b('party_guard',35918)['partyflash'],'party fire date/flag')
    before=b('before_reset',35918);after=b('after_reset',35918)
    require(before['matrix_pc']==after['matrix_pc'] and before['matrix_remaining']==after['matrix_remaining'],'party reset matrix preservation')
    require(not any(after['live_task_sites']) and not after['waits'] and after['reset_count']==2,'reset clears')
    hand=b('body_after:PARTY_ON_TASK1',35918)
    require(hand['live_task_sites'][:3]==['SOUNDNEWBALL','SETBALL','SOUNDBRICKUPP'] and not hand['waits'],'new task slots/zero waits')
    require(rows[35918]['waits']=={'SETBALL':1,'SOUNDBRICKUPP':1},'same scan initial visit; SOUND age zero')
    for n in range(35918,35998):
        pre=b('wait_before:SETBALL',n);after=b('wait_after:SETBALL',n)
        require(pre['waits'].get('SETBALL',0)==n-35918 and after['waits']['SETBALL']==n-35917,'SETBALL age')
        require(sum(x['calculation']==n and x['boundary']=='wait_before:SETBALL' for x in bs)==1,'SETBALL once per scan')
        require(pre['live_task_sites'][1]=='SETBALL' and pre['live_task_sites'].count('SETBALL')==1 and pre['reset_count']==2,'SETBALL slot/survival/duplicate/reset')
        require(not rows[n]['QUIT'],'no TABLE1 exit')
    pre=b('electronics_before',35998);installed=b('expiry_installed',35998);scan=b('wait_before:SETBALL',35998);body=b('body_after:SETBALL',35998)
    require(pre['timer']==35997 and not pre['expired'] and pre['ball']['Hold'] and not pre['BALL_DOWN'] and not pre['LOOSING'],'pre-electronics')
    require(installed['timer']==35998 and installed['expired'] and installed['ball']['Hold'] and installed['audio_priority']==255,'expiry equality/cue')
    require(installed['linked_matrix']==dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=5),'expiry initial cursor')
    require(scan['waits']['SETBALL']==80 and scan['live_task_sites']==['','SETBALL']+['']*48,'real equality slot/age')
    order=[x['boundary'] for x in bs if x['calculation']==35998]
    require(order.index('electronics_before')<order.index('expiry_installed')<order.index('wait_before:SETBALL')<order.index('body_after:SETBALL')<order.index('tasks_after')<order.index('matrix_before')<order.index('matrix_after')<order.index('late_before')<order.index('late_after'),'whole calculation order')
    require((body['ball']['PixelX'],body['ball']['PixelY'],body['ball']['VX'],body['ball']['VY'],body['ball']['High'],body['ball']['Hold'])==(297,530,10,0,False,False),'SETBALL concrete stores')
    require(body['expired'] and body['in_chute'] and body['waits']['SETBALL']==0,'SETBALL expired/I_UTSKJUT/wait')
    for role in ('wait_before:SETBALL','body_after:SETBALL','tasks_after','matrix_before'):
        require(b(role,35998)['linked_matrix']==installed['linked_matrix'],'expiry untouched before matrix')
    matrix=b('matrix_after',35998);late0=b('late_before',35998);late1=b('late_after',35998)
    require(matrix['matrix_remaining']==4 and matrix['matrix_pc']==installed['matrix_pc'],'first expiry visit')
    require(late0['ball']==body['ball'] and late1['ball']['X']==body['ball']['X']+10 and late1['ball']['Y']==body['ball']['Y'],'one same-call late physics step')
    require(late1['ball']['VY']==body['ball']['GY'] and late1['ball']['VX']==10+body['ball']['GX'],'late gravity')
    require([x for x in w['equality_physics_reads'] if x[0]=='response']==[['response',7,1024,13,0,0]],'actual late contact')
    require(rows[35998]['ball']==late1['ball'] and rows[35998]['expired'] and rows[35998]['matrix_remaining']==4,'calculation end')
    require(w['final']['QUIT'] and w['final']['calculation']==(37047 if corrected else 37039),'no-input QUIT')
    require([x['calculation'] for x in bs if x['boundary']=='drain_entry']==[35888],'no subsequent drain')
    for v,dr,fire in zip(r['fresh_input_replay']['neighbors'],(35887,),(35997,)):
        require(v['drain']==dr and not v['scored'] and boundary(v,'body_after:SETBALL',fire),'real neighbor off-by-one')
    m=r['fresh_input_replay']['limit_plus_one']
    require(not any(x['calculation']==35998 and x['boundary']=='body_after:SETBALL' for x in m['boundaries']) and boundary(m,'body_after:SETBALL',35999),'limit mutation fire shifts')
    m=r['fresh_input_replay']['matrix_replace']
    require(boundary(m,'body_after:SETBALL',35998)['matrix_pc']!=installed['matrix_pc'],'matrix replacement mutation detected')
    require(r['expiry_interleaving']=='EXPIRY_INTERLEAVING = NOT_PROVED' and 'DROPTASK2' in r['exactly_one_remaining_dependency'],'inventory incompleteness retained')
    return True

def report(raw,identity,evidence,nodes):
    w=raw['witness'];installed=boundary(w,'expiry_installed',35998);entry=installed['matrix_pc']-1
    for v in [w,*raw['neighbors'],raw['limit_plus_one'],raw['matrix_replace']]:
        for x in v['rows']+v['boundaries']:
            x['linked_matrix']=party.matrix_state(x,nodes,entry) if entry<=x['matrix_pc']<=entry+len(nodes) else party.matrix_state(x,evidence['party_program'],55)
    b=lambda role,n:boundary(w,role,n)
    rows={x['calculation']:x for x in w['rows']}
    settle=37047
    while settle>35998 and rows[settle-1]['ball']==w['final']['ball']:settle-=1
    r=dict(timing_oracle_basis='FF-inclusive-v1',class_verdict='FIRST_EQUALITY_SETBALL_REACHABLE',expiry_interleaving='EXPIRY_INTERLEAVING = NOT_PROVED',status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',production_base=replay.BASE,research_head='37ff8bc38d7af4a09a70319675d6e505b0bd6a5e',native_reference_identity=identity,linked_evidence=evidence,
      target_predecessor_arithmetic=dict(D=target_drain(),party_first_visit='D; zero -> one',party_fire='D+30 = 35918; compare30 then zero',SETBALL_first_visit='D+30; zero -> one in resumed slot1 scan',SETBALL_age80='D+109 = 35997 after scan',SETBALL_fire='D+110 = 35998; visit81 compares80 before increment',neighbors={'35887':35997,'35889':35999},wrong_next_scan_D=target_drain(same_scan=False)),
      neighbor_search=dict(D_minus_one='real replay Release35758, Down1..35757; drain35887; SETBALL35997',D_plus_one='UNKNOWN: no input-only35889 drain found in finite searches',finite_release_search={'long_hold_release':[35300,35780],'additional_charge_search':{'release':[35720,35780],'charges':list(range(1,17))}},arithmetic_mutations=[dict(D=D,party_fire=firing_from_drain(D)[0],SETBALL_fire=firing_from_drain(D)[1],scope='counterfactual arithmetic only') for D in (35887,35889)]),
      input_script=dict(Down=[1,w['release']-1],Release=w['release'],all_other_controls=False,after_release='all false through QUIT'),fresh_input_replay=raw,
      drain_checkpoint=b('drain_entry',35888),party_admission=b('party_admitted',35888),PARTY_ON_TASK1_age_trace=[x for x in w['boundaries'] if x['boundary'].startswith(('wait_before:PARTY_ON_TASK1','wait_after:PARTY_ON_TASK1'))],
      new_ball_handoff={role:b(role,35918) for role in ('party_guard','before_reset','after_reset','body_after:PARTY_ON_TASK1','tasks_after')},
      task_wait_projection=dict(SOUNDNEWBALL={'slot':0,'task_DS':0x3417,'wait_DS':0x36d1,'initial':0,'after_scan':0},SETBALL={'slot':1,'task_DS':0x3419,'wait_DS':0x36d3,'initial':0,'after_scan':1},SOUNDBRICKUPP={'slot':2,'task_DS':0x341b,'wait_DS':0x36cf,'initial':0,'after_scan':1}),
      SETBALL_age_trace=[x for x in w['boundaries'] if x['boundary'].startswith(('wait_before:SETBALL','wait_after:SETBALL','body_after:SETBALL'))],
      full_calculation_35998=[x for x in w['boundaries'] if x['calculation']==35998],calculation_35997=next(x for x in w['rows'] if x['calculation']==35997),calculation_35998=next(x for x in w['rows'] if x['calculation']==35998),
      SETBALL_effects=dict(expired=True,HOLDSTILL=False,high=False,I_UTSKJUT=65535,program=0x1ba17,cursor=0x1ba19,clear_remaining=5,preserved=True,restarted=False,replaced=False),
      same_calculation=dict(first_expiry_matrix_visit=True,clear_remaining=4,late_physics_steps=1,late_before=b('late_before',35998),late_after=b('late_after',35998),callback_result='area/target/drain dispatch phases have already returned before DO_TASKS; late sc_program has no same-calculation area/drain dispatch; actual late collision is material7/angle1024/contact count13; response exits with n<=0, so it changes contact metadata but no velocity or game callback/effect; next early DO_PHYSICS would consume any late loss'),
      concrete_no_input_continuation=dict(QUIT=True,calculation=w['final']['calculation'],additional_drains=0,expiry_visits=1050,ball_settled_from=settle,settled_state=w['final']['ball'],in_chute_through_QUIT=all(x['in_chute'] for x in w['rows'] if x['calculation']>=35998),final=w['final'],matrix_starts=[dict(calculation=x['calculation'],label=e['Label']) for x in w['event_rows'] for e in x['events'] if e['Kind']=='MatrixStarted' and x['calculation']>=35998]),
      consumed_writer_scope='Actual suffix: wait macros for three queued tasks, SOUNDBRICKUPP/SOUNDNEWBALL cue calls and SUICIDE; PARTYRUT RET; ordinary native counters/display. Direct linked task store operands and relocated consumed correspondence checked. SETBALL has no36d3 writer except WAITSYNCS reset. One actual slot1 visit per scan, no duplicates/reset/exit through firing. Not a global writer/IRQ theorem.',
      projection_scope='VISAKEYS=false from accepted fresh reset and no visa-writing suffix. DS wait words correspond to shared wait map, not per-slot age. I_UTSKJUT=ffff from linked SETBALL store and native in_chute=true. A native trajectory with consumed linked demo timer/expiry overlays; no DOS execution claim.',
      updated_first_equality_class_inventory=[dict(name='NEW_BALL_TASK',reachability='PROVED previous pass',outcome='SHOWPLAYERSTS replacement before expiry visit'),dict(name='PARTY_ON_TASK1',reachability='PROVED previous pass',outcome='expiry program/cursor preserved; no-input QUIT37047'),dict(name='SETBALL',reachability='PROVED this pass',outcome='expiry preserved; releases ball for same-call late physics; no-input QUIT37047')],
      bounded_inventory_review=dict(source='audit_10min_demo_expiry_interleaving.audit threshold_state_classes',classes_complete=False,covered=['normal new-ball equality firing','unscored drain continuation','SETBALL equality firing'],nonfiring_classes=['ordinary active ball with no suffix effect','normal bonus/new-ball before or later firing','SETBALL later firing'],early_drain='scored drain sees expired=false before electronics; equality supersedes its early matrix installation',remaining_bucket='other held-ball continuation',why_not_exhaustive='Previous bounded inventory explicitly left table-specific capture/mode producers untraversed; these three witnesses do not exclude a different HOLDSTILL-clearing task at equality.'),
      exactly_one_remaining_dependency='DROPTASK2 capture-release at first equality: reachability or exclusion of a fresh input-only surviving wait at its firing age, and its actual post-expiry release/same-calculation suffix. This is one instance of the previously open other held-ball continuation bucket; no universal termination or all-task reopening.',tests={})
    validate(r);return r

if __name__=='__main__':main()
