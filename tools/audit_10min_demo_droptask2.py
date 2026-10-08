#!/usr/bin/env python3
"""Narrow linked DROPTASK2 research; isolated fresh canonical-A input replay."""
import json, os, subprocess, argparse, struct
from pathlib import Path
import audit_10min_demo_setball as setball
import audit_10min_demo_party_on as party
import audit_10min_demo_post_collision as post
import audit_10min_demo_deterministic_replay as replay
from audit_10min_demo_graph import Decoder,pinned,require
from audit_10min_demo_programs import identities
from audit_10min_demo_source_recurrence import linear
ROOT=replay.ROOT
OUTPUT=Path('/private/tmp/pf-dmo0-first-equality-droptask2.json')
MAP=ROOT/'tools/dmo0_droptask2_correspondence.json'

def firing(C,first_next=False,child_next=True,l1=30,l2=27):
    # Ages start at zero, compare-before-increment; F1 is visit l1+1.
    return C+int(first_next)+l1+int(child_next)+l2

def linked(b,full,d,h):
    evidence=setball.linked(b,full,d,h)
    evidence['drop_chain']=replay.compare_instructions(d,full,b,json.loads(MAP.read_text()))
    # Full correspondence plus independent binding to actual consumer constants.
    for at,op,args in ((0x1479,'mov','dx, 0x1e'),(0x147c,'mov','bx, 0x36d7'),(0x149a,'mov','dx, 0x11a3'),(0x14a3,'mov','dx, 0x1b'),(0x14a6,'mov','bx, 0x36d9'),(0x14b2,'mov','word ptr [0x347f], 0xffff'),(0x14b8,'mov','bp, word ptr [0x34ec]'),(0x14bc,'and','bp, 0x7f'),(0x14bf,'add','bp, 0'),(0x14c2,'mov','word ptr [0x2fdc], 0xf'),(0x14c8,'mov','word ptr [0x2fde], 0x2f'),(0x14ce,'mov','byte ptr [0x3416], 0xff'),(0x14f0,'mov','word ptr [0x2fea], bp'),(0x14f4,'mov','word ptr [0x2fe8], 0'),(0x1518,'jmp','0x57ba'),(0x1ebc,'mov','word ptr [0xac], 0x82'),(0x1f1e,'mov','dx, word ptr [0xac]'),(0x1f22,'mov','bx, 0x36f1'),(0x1f3a,'mov','dx, 2'),(0x1f3d,'mov','bx, 0x36f3')):d.expect(b,768,at,op,args)
    av=struct.unpack_from('<5H',full,0x1aadb+30);bv=struct.unpack_from('<5H',b,0x1ab67+30)
    require(av[:4]==bv[:4]==(47,122,67,146) and (av[4]+768,bv[4]+768)==(0x1e29,0x1e2d),'GROPD area producer binding')
    evidence['region_rectangles']=[replay.record(full,b,'area_rectangle',aa+10*i,bb+10*i,8,plane=plane,index=i) for plane,aa,bb,count in [('lower',0x1aadb,0x1ab67,14),('upper',0x1ab69,0x1abf5,12)] for i in range(count)]
    require(struct.unpack_from('<H',b,0x1a93c+26)[0]+0x19db0==0x1b904,'SKILLTUNNEL program consumer binding')
    evidence['capture_matrix_program']=party.expiry.program(b,h,0x1b904)
    require([x['op'] for x in evidence['capture_matrix_program']]==['_CLEAR4','_FLASHON','_PRINT13','_WAIT','_CLEAR4','_FLASHOFF','_FLASHON','_PRINT13_NUMBER_CENT','_WAIT','_CLEAR4','_FLASHOFF'],'capture program shape')
    evidence['capture_binding']=dict(rectangle=list(bv[:4]),A_record=0x1aaf9,demo_record=0x1ab85,A_callback=0x1e29,demo_callback=0x1e2d)
    writes=[]
    for x in linear(d,b,0x14b2,0x151b,768):
        for o in x.operands:
            if o.type==d.x86.X86_OP_MEM and o.access&d.cs.CS_AC_WRITE:
                require(not o.mem.base and not o.mem.index,'indexed drop body writer')
                writes.append(dict(site=x.address+768,DS=o.mem.disp,width=o.size))
    require(not {0x3026,0x34cf,0x34ca,0x34cb,0x3481}.intersection(x['DS'] for x in writes),'unexpected drop hold/expired/down/chute writer')
    # Only bind the smallest independently relevant old-inventory candidate.
    for at,op,args in ((0x2750,'mov','dx, 0x1b'),(0x2753,'mov','bx, 0x3701'),(0x2778,'mov','byte ptr [0x3026], 0'),(0x277e,'mov','word ptr [0x2fdc], 0x101'),(0x2784,'mov','word ptr [0x2fde], 0x136'),(0x27ac,'mov','word ptr [0x2fea], 0x627'),(0x27b2,'mov','word ptr [0x2fe8], 0xfdc1')):d.expect(b,768,at,op,args)
    evidence['DROPTASK2_direct_writes']=writes
    evidence['HOLDSTILL_write']=False
    evidence['body_semantics']=dict(X=15,Y=47,X_fixed=15360,Y_fixed=48128,VX=0,VY='SLUMP_COUNTERN & 127',high=True,SCREENFORCE=-1,HOLDSTILL='unchanged',expired='unchanged',BALL_DOWN='unchanged',LOOSING='unchanged',I_UTSKJUT='unchanged',matrix='unchanged',effect='unchanged',flash_ends=[3,56],sound='SNEWBALL',suicide=0x5aba)
    return evidence

def prepare(harness,canonical,cfg):
    (harness/'internal/partyland/dmo0_droptask2_test.go').unlink(missing_ok=True)
    identity=setball.prepare(harness,canonical,cfg)
    (harness/'internal/partyland/dmo0_setball_test.go').unlink()
    # Generic tasks retain their source closure name, passively decoded by observer.
    p=harness/'internal/partyland/dmo0_collision_handoff_test.go'
    replay.replace_once(p,'package partyland','package partyland\nimport("reflect";"runtime";"strings")')
    replay.replace_once(p,'if out[i]=="" {panic("unmapped live task")}', 'if out[i]=="" {out[i]=strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(t).Pointer()).Name(),"pinballfantasies/internal/partyland.(*Game).")}')
    p=harness/'internal/partyland/timing.go'
    replay.replace_once(p,'func (g *Game) waitAt(site string, n uint16, f func()) {', 'func (g *Game) waitAt(site string, n uint16, f func()) {restore:=researchDropSlots(g,site);defer restore()')
    replay.replace_once(p,'if !bypass() && !g.waitReady(site, n) {','if g.ResearchObserve!=nil {g.ResearchObserve("wait_before:"+site)}\n if !bypass() && !g.waitReady(site, n) {\n if g.ResearchObserve!=nil {g.ResearchObserve("wait_after:"+site)}')
    replay.replace_once(p,'\t\tf()\n\t\treturn true\n\t})\n}', '\t\tf()\n if g.ResearchObserve!=nil {g.ResearchObserve("body_after:"+site)}\n\t\treturn true\n\t})\n if g.ResearchTaskNames==nil {g.ResearchTaskNames=map[uint64]string{}};g.ResearchTaskNames[g.nextTaskID]=site\n}')
    text=p.read_text();lo=text.index('func (g *Game) cameraDrop()');hi=text.index('func (g *Game) waitReady',lo);part=text[lo:hi]
    part=part.replace('g.ScrollPosition -= 5','if g.ResearchObserve!=nil {g.ResearchObserve("camera_child_before")}\n g.ScrollPosition -= 5').replace('g.ScreenForce = 0','g.ScreenForce = 0\n if g.ResearchObserve!=nil {g.ResearchObserve("camera_child_suicide")}').replace('\t\t})\n\t\treturn true','\t\t})\n if g.ResearchObserve!=nil {g.ResearchObserve("camera_parent_after")}\n\t\treturn true')
    p.write_text(text[:lo]+part+text[hi:])
    p=harness/'internal/partyland/regions.go'
    replay.replace_once(p,'func (g *Game) trigger(label string) {','func (g *Game) trigger(label string) {if g.ResearchObserve!=nil {g.ResearchObserve("area_before:"+label);defer g.ResearchObserve("area_after:"+label)}')
    p=harness/'internal/partyland/holes.go'
    replay.replace_once(p,'func (g *Game) startDrop() {','func (g *Game) startDrop() {if g.ResearchObserve!=nil {g.ResearchObserve("start_drop_before");defer g.ResearchObserve("start_drop_after")}')
    replay.replace_once(p,'g.waitAt("DROPTASK2", 27,','g.waitAt("DROPTASK2", researchDropLimit(g),')
    replay.replace_once(p,'g.Physics.Ball.Hold = false\n\t\t\tg.ScreenForce = -1','g.Physics.Ball.Hold = g.ResearchExpired // linked DROPTASK2 does not clear expiry HOLDSTILL; native capture hold is separate.\n if g.ResearchPartyMutation=="clear_hold" {g.Physics.Ball.Hold=false}\n if g.ResearchPartyMutation=="drop_state" {g.Physics.SetBall(16,47,0,0,false)}\n\t\t\tg.ScreenForce = -1')
    p=harness/'internal/physics/ball.go'
    replay.replace_once(p,'func (g *Game) step(input Inputs) error {','func (g *Game) step(input Inputs) error {g.research("step_before")')
    text=p.read_text();lo=text.index('func (g *Game) step(');hi=text.index('func divide(',lo)
    part=text[lo:hi];require(part.count('\treturn nil')==1,'physics step return anchor');part=part.replace('\treturn nil','g.research("step_after")\n\treturn nil');p.write_text(text[:lo]+part+text[hi:])
    p=harness/'internal/partyland/game.go'
    replay.replace_once(p,'func (g *Game) beforeTargets() {','func (g *Game) beforeTargets() {if g.ResearchObserve!=nil {g.ResearchObserve("update_counters_before")}')
    (harness/'internal/partyland/dmo0_droptask2_test.go').write_bytes((ROOT/'tools/dmo0_droptask2_test.go.txt').read_bytes())
    return identity

TARGET=dict(Release=35534,Charge=18,LP=95,LD=57,LO=13,RP=84,RD=34,RO=22,ControlsEnd=35810)
TASKS=dict(WAIT_FOR_TUNNEL_EFFECT=dict(file=0x1f13,wait_DS=0x36f1,limit=130,slot=1),LOCK_BALL_IN_TUNNEL=dict(file=0x1f3a,wait_DS=0x36f3,limit=2,slot=2),DROPTASK1=dict(file=0x1479,wait_DS=0x36d7,limit=30,slot=2),DROPTASK2=dict(file=0x14a3,wait_DS=0x36d9,limit=27,slot=0))
boundary=party.boundary

def visit_trace(start,limit):
    age=0;rows=[]
    for n in range(start,start+limit+1):
        after,fire=party.expiry.wait_visit(age,limit);rows.append(dict(calculation=n,before=age,after=after,fire=fire));age=after
    return rows

def index(w):
    out={}
    for x in w['boundaries']:out.setdefault((x['boundary'],x['calculation']),[]).append(x)
    return out

def validate(r):
    corrected = r.get('timing_oracle_basis') == 'FF-inclusive-v1'
    raw=r['fresh_input_replay'];w=raw['witness'];ix=index(w);b=lambda role,n=35998:ix[(role,n)][0]
    rows={x['calculation']:x for x in w['rows']}
    require(raw['deterministic_equal'] and raw['any_capture_deterministic_equal'] and w['script']==TARGET,'deterministic explicit input script')
    require((w['capture'],w['START_DROP'],w['DROPTASK1_fire'],w['DROPTASK2_fire'])==(35810,35940,35970,35998),'producer arithmetic/witness')
    require(firing(35940)==35998,'predecessor equation')
    require(not any(b('area_before:GROPD',35810)['live_task_sites']),'clean task origin at capture')
    require(b('area_after:GROPD',35810)['live_task_sites'][:3]==['cameraDrop.func1','WAIT_FOR_TUNNEL_EFFECT','LOCK_BALL_IN_TUNNEL'],'capture insertion order')
    require(b('camera_parent_after',35810)['live_task_sites'][3].startswith('cameraDrop.func1.') and b('camera_child_before',35810)['live_task_sites'][3].startswith('cameraDrop.func1.') and not b('tasks_after',35810)['live_task_sites'][3],'camera child same scan')
    require(b('camera_child_suicide',35941)['live_task_sites'][1].startswith('cameraDrop.func1.'),'second camera same scan child suicide')
    require(b('start_drop_before',35940)['live_task_sites']==['','WAIT_FOR_TUNNEL_EFFECT']+['']*48,'start drop actual producer slot')
    require(b('start_drop_after',35940)['live_task_sites'][:3]==['cameraDrop.func1','WAIT_FOR_TUNNEL_EFFECT','DROPTASK1'],'drop insertion order')
    require(b('tasks_after',35940)['waits']['DROPTASK1']==1,'drop1 first visit same scan')
    require(b('body_after:DROPTASK1',35970)['live_task_sites'][:3]==['DROPTASK2','','DROPTASK1'],'child earlier slot before suicide')
    require('DROPTASK2' not in b('tasks_after',35970)['waits'],'child not visited insertion calculation')
    for name,start,limit in [('WAIT_FOR_TUNNEL_EFFECT',35810,130),('LOCK_BALL_IN_TUNNEL',35810,2),('DROPTASK1',35940,30),('DROPTASK2',35971,27)]:
        slot=TASKS[name]['slot'];task_id=b('wait_before:'+name,start)['task_ids'][slot]
        for n in range(start,start+limit+1):
            visits=ix.get(('wait_before:'+name,n),[]);require(len(visits)==1,'one callsite visit per calculation')
            x=visits[0];require(x['waits'].get(name,0)==n-start,'shared wait age boundary')
            require(x['live_task_sites'][slot]==name and x['live_task_sites'].count(name)==1 and x['task_ids'][slot]==task_id,'slot survival/duplicates')
            require(x['reset_count']==1 and not x['QUIT'],'no reset/exit')
            if n<start+limit:require(b('wait_after:'+name,n)['waits'][name]==n-start+1,'increment-before-next-visit')
            else:require(b('body_after:'+name,n)['waits'][name]==0,'compare match resets age')
    for n in range(35812,35998):
        x=rows[n];require(x['ball']['Hold'] and not x['expired'] and not x['linked_HOLDSTILL'],'native capture freeze distinct from source HOLDSTILL')
        require((x['ball']['X'],x['ball']['Y'],x['ball']['VX'],x['ball']['VY'],x['ball']['High'])==(15360,48128,0,0,False),'held capture state')
        require(x['reset_count']==1 and not x['BALL_DOWN'] and not x['LOOSING'] and not x['in_chute'],'concrete suffix survives')
    require(rows[35997]['waits']['DROPTASK2']==27 and not rows[35997]['expired'],'pre equality age')
    early=ix[('step_before',35998)];early_after=ix[('step_after',35998)]
    require(len(early)==len(early_after)==3 and early[0]['ball']==early_after[0]['ball']==early[1]['ball']==early_after[1]['ball'],'two early held-ball passes')
    pre=b('electronics_before');installed=b('expiry_installed');body=b('body_after:DROPTASK2')
    require(pre['timer']==35997 and not pre['expired'] and not pre['linked_HOLDSTILL'],'pre equality linked hold')
    require(installed['timer']==35998 and installed['expired'] and installed['linked_HOLDSTILL'] and installed['audio_priority']==255,'expiry installed')
    require(installed['linked_matrix']==dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=5),'expiry operands/program/cursor')
    require(b('wait_before:DROPTASK2')['waits']['DROPTASK2']==27,'first equality real wait')
    require((body['ball']['X'],body['ball']['Y'],body['ball']['PixelX'],body['ball']['PixelY'],body['ball']['VX'],body['ball']['VY'],body['ball']['High'],body['ball']['Hold'])==(15360,48128,15,47,0,52,True,True),'exact linked drop stores and HOLDSTILL preservation')
    require(body['SCREENFORCE']==-1 and body['expired'] and body['linked_HOLDSTILL'] and not body['BALL_DOWN'] and not body['LOOSING'] and not body['in_chute'],'source state unchanged')
    for role in ('wait_before:DROPTASK2','body_after:DROPTASK2','tasks_after','matrix_before'):
        require(b(role)['linked_matrix']==installed['linked_matrix'],'expiry current without replacement/restart')
    require(not any(b('tasks_after')['live_task_sites']),'DROPTASK2 suicide and empty remaining scan')
    require(b('matrix_after')['linked_matrix']==dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=4),'first matrix visit')
    require(b('late_before')['ball']==b('late_after')['ball']==body['ball'],'late held step does not move/collide/apply gravity')
    require(not any(x[0] in ('ring','response','sine') for x in w['equality_physics_reads']),'no equality collision response reads')
    order=[x['boundary'] for x in w['boundaries'] if x['calculation']==35998]
    require(order.index('update_counters_before')<order.index('electronics_before')<order.index('expiry_installed')<order.index('electronics_after')<order.index('wait_before:DROPTASK2')<order.index('body_after:DROPTASK2')<order.index('tasks_after')<order.index('matrix_before')<order.index('matrix_after')<order.index('late_before')<order.index('late_after'),'exact calculation ordering')
    require(w['final']['QUIT'] and w['final']['calculation']==(37047 if corrected else 37039),'concrete no input QUIT')
    for n in range(35998,(37047 if corrected else 37039)):
        x=rows[n];require(x['expired'] and x['linked_HOLDSTILL'] and x['ball']==body['ball'],'high ball remains frozen through expiry')
        require(not any(x['live_task_sites']) and x['reset_count']==1,'no subsequent tasks/reset')
        require(not any(x['input'].values()),'no input after capture')
    require([(x['calculation'],x['event']['Label']) for x in w['milestones'] if x['calculation']>=35998 and x['event']['Kind']=='MatrixStarted']==[(35998,'RESEARCH_EXPIRY')],'no future replacement/restart')
    require(not any(x['boundary'].startswith(('area_before:','drain_entry')) and x['calculation']>35810 for x in w['boundaries']),'no future capture/drain callback')
    m=raw['wait_plus_one'];require(m['DROPTASK2_fire']==35999 and boundary(m,'wait_before:DROPTASK2',35998)['waits']['DROPTASK2']==27,'wait mutation shifts firing')
    m=raw['child_later_slot'];require(m['DROPTASK2_fire']==35997 and boundary(m,'body_after:DROPTASK1',35970)['live_task_sites'][3]=='DROPTASK2' and boundary(m,'wait_before:DROPTASK2',35970)['waits'].get('DROPTASK2',0)==0,'later slot mutation gives same scan visit')
    m=raw['drop_state'];require(boundary(m,'body_after:DROPTASK2',35998)['ball']['PixelX']==16 and not boundary(m,'body_after:DROPTASK2',35998)['ball']['High'],'drop state mutation detected')
    m=raw['clear_hold'];require(not boundary(m,'body_after:DROPTASK2',35998)['ball']['Hold'] and boundary(m,'late_after',35998)['ball']['Y']==48180,'incorrect native hold-clear mutation releases ball')
    require(r['bounded_inventory_review']['classes_complete']==False and r['exactly_one_remaining_dependency']['task']=='DURINGFLASH','inventory remaining independent class')
    return True

def compact(w,end):
    return {k:([x for x in v if x['calculation']<=end] if k in ('rows','boundaries','event_rows','milestones') else v) for k,v in w.items()}

def report(raw,identity,evidence,nodes,any_capture,search):
    w=raw['witness'];entry=boundary(w,'expiry_installed',35998)['matrix_pc']-1
    for v in raw.values():
        if not isinstance(v,dict) or 'boundaries' not in v:continue
        for x in v['rows']+v['boundaries']+[v['final']]:
            # Capture freeze is the accepted native abstraction. It is not a
            # source HOLDSTILL store: linked lock/drop SETBALLPOS never writes it.
            x['linked_HOLDSTILL']=x['expired'] and x['ball']['Hold']
            if entry<x['matrix_pc']<=entry+len(nodes):x['linked_matrix']=party.matrix_state(x,nodes,entry)
            elif 363<x['matrix_pc']<=374:x['linked_matrix']=party.matrix_state(x,evidence['capture_matrix_program'],363)
            elif x['matrix_pc']==375 and not x['matrix_active']:x['linked_matrix']=dict(program=0x1b904,cursor=0,op='NODOT',last_terminated_node=0x1b92e)
            elif x['matrix_pc']==454 and not x['matrix_active']:x['linked_matrix']=dict(program=0x1b88e,cursor=0,op='NODOT',last_terminated_node=0x1b89c)
            else:x['linked_matrix']=dict(native_pc=x['matrix_pc'],op=x['matrix_op'],active=x['matrix_active'])
    b=lambda role,n=35998:boundary(w,role,n)
    inventory=[]
    for name,outcome in [('ordinary active ball','equality supersedes current program; definition has no suffix effect'),('normal bonus/new-ball before firing','neighboring nonfiring age of NEW_BALL_TASK producer'),('normal new-ball equality firing','previous proved NEW_BALL_TASK replacement class'),('normal new-ball later firing','nonfiring age of NEW_BALL_TASK producer'),('SETBALL equality firing','previous proved SETBALL release class'),('SETBALL later firing','neighboring nonfiring age of SETBALL producer'),('unscored drain continuation','previous proved PARTY_ON_TASK1 preservation class'),('scored drain at equality','early drain sees expired=false; electronics equality supersedes its matrix'),('other held-ball continuation','DROPTASK2 firing now proved, but DURINGFLASH is an independent release candidate')]:inventory.append(dict(old_entry=name,classification=outcome))
    remaining=dict(task='DURINGFLASH',file=0x2750,wait_DS=0x3701,limit=27,HOLDSTILL_clear=0x2778,X=257,Y=310,VX=-575,VY=1575,high=False,dependency='Fresh input-only dragon producer whose actual surviving shared wait reaches 27 on first equality; classify its released ball suffix. Reachability/exclusion UNKNOWN. No search or producer-chain expansion in this pass.',why_independent='Explicit HOLDSTILL clear; different coordinates/velocity, high=false and play-field release. Neither this hold-preserving drop nor the chute SETBALL trajectory establishes its collision/effect continuation.')
    r=dict(timing_oracle_basis='FF-inclusive-v1',class_verdict='FIRST_EQUALITY_DROPTASK2_REACHABLE',expiry_interleaving='EXPIRY_INTERLEAVING = '+('NOT_PROVED' if remaining else 'PROVED'),status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',production_base=replay.BASE,research_head='37ff8bc38d7af4a09a70319675d6e505b0bd6a5e',native_reference_identity=identity,linked_evidence=evidence,
      linked_producer_chain=dict(selected='GROPD non-special tunnel; SKILLTUNNEL admitted and inhibits TSCORE1 matrix; no START_DROP_TIMED/START_DROP_WHEN_READY is consumed',capture_file=0x1e2d,capture_region_file=0x1ab85,task_records=TASKS,camera=dict(RULLGARDIN_file=0x151b,RULLGARDINSLAV_file=0x152c,WAITLIST=None,instances=[dict(parent_slot=0,parent_first_visit=35810,child_slot=3,child_first_visit=35810,child_suicide=35810),dict(parent_slot=0,parent_insert=35940,parent_first_visit=35941,child_slot=1,child_first_visit=35941,child_suicide=35941)]),START_DROP_file=0x142a,allocation='first-free; no wait reset; scan slots ascending 0..49; parent remains occupied until SUICIDE'),
      predecessor_equation=dict(E=35810,C=35940,F1=35970,F=35998,equation='C=E+130; F1=C+30; DROPTASK2 first_visit=F1+1; F=F1+1+27=E+188',off_by_one='Each wait first compares0, then increments. WAIT130 fires visit131; DROP1 wait30 fires visit31 in its later slot2; DROP2 inserted slot0 after scan passed0, first compares0 next calculation and fires visit28.',wrong_same_scan_result=35997,wrong_DROP1_next_scan_result=35999),
      input_script=dict(Down=[35516,35533],Release=35534,flippers=dict(start=35534,end=35810,Left='(n-35534+13)%95 < 57',Right='(n-35534+22)%84 < 34'),otherwise='all controls false; all controls false from35811 through QUIT',script=TARGET),
      any_capture_replay=compact(any_capture['found'],any_capture['found']['DROPTASK2_fire']),search=dict(first_family_trials=any_capture['trials'],failed_long_release_range=[35347,35859],failed_scripts=513,residue_candidates=search['trials'],full_delayed_candidates=search['long_trials'],candidate_rule='short F mod128 matches35998; only proposes a delayed script. Full fresh replay accepts it; no linear shift assumption or exclusion theorem.'),fresh_input_replay=raw,
      capture_checkpoints=[x for x in w['boundaries'] if x['boundary'] in ('area_before:GROPD','area_after:GROPD','body_after:LOCK_BALL_IN_TUNNEL','start_drop_before','start_drop_after','body_after:DROPTASK1')],
      age_timelines={name:[dict(calculation=x['calculation'],slot=TASKS[name]['slot'],DS=TASKS[name]['wait_DS'],age=x['waits'].get(name,0),live_task_ids=x['task_ids']) for x in w['boundaries'] if x['boundary']=='wait_before:'+name] for name in TASKS},
      pre_equality_matrix=dict(capture_program=0x1b904,terminated_at=35920,idle_panel_at=35921,idle_panel_program=0x1b88e,idle_panel_terminated_at=35923,state_before_equality=dict(program=0x1b88e,cursor=0,op='NODOT'),effect_accepted=False,INH_EFF=False,SPECIALMODE=False),
      calculation_35997=next(x for x in w['rows'] if x['calculation']==35997),full_calculation_35998=[x for x in w['boundaries'] if x['calculation']==35998],calculation_35998=next(x for x in w['rows'] if x['calculation']==35998),
      DROPTASK2_result=dict(fires=True,slot=0,task_DS=0x3417,wait_DS=0x36d9,wait_before=27,wait_after=0,X=15,Y=47,VX=0,VY=52,high=True,HOLDSTILL=True,HOLDSTILL_written=False,expired=True,program=0x1ba17,cursor=0x1ba19,clear_remaining=5,SCREENFORCE=-1,BALL_DOWN=False,LOOSING=False,I_UTSKJUT=False,remaining_tasks=0,expiry_replaced=False,expiry_restarted=False,releases_after_expiry=False),
      same_calculation=dict(first_matrix_visit=dict(program=0x1ba17,cursor=0x1ba19,clear_before=5,clear_after=4),late_physics=dict(before=b('late_before')['ball'],after=b('late_after')['ball'],steps=1,physics_reads=w['equality_physics_reads'],collision_material_region='No late collision/material/ramp/region reads: expiry HOLDSTILL skips collision and movement; only flipper2 frame0 copy. Earlier old-low ramp indices2201/2202, five negative low-level regions precede tasks. Existing material6/contact1 metadata is stale, not a consumed late response.',callbacks='Areas/targets/drain phases have returned before DO_TASKS; no late area/target/drain/effect callback, replacement or restart.')),
      concrete_no_input_continuation=dict(QUIT=True,calculation=37047,expiry_visits=1050,high_ball_enters_play=False,high_ball_position=[15,47],velocity=[0,52],frozen_through_QUIT=True,new_captures=0,new_drains=0,new_effects=0,expiry_replacements=0,expiry_restarts=0,final=w['final']),
      consumed_writer_scope='Concrete suffix only. Shared wait callsites36f1/36f3/36d7/36d9 each visited once in its actual sole slot; no duplicate, reset or exit. Camera ends before child insertion. LIGHTFLASH/ENDFLASH operate bounded3654..368f; SUICIDE writes its current TASKLIST entry and count347b. Native counters, display/audio and expiry overlay remain accepted reference primitives. No global task/alias/IRQ closure.',
      projection_scope='Accepted fresh canonical-A reference. Native capture Hold=true models capture freeze, not linked HOLDSTILL: linked GROPD/LOCK/START_DROP/DROPTASK2 SETBALLPOS does not write3026. Fresh source HOLDSTILL=false after initial SETBALL; equality sets true. The isolated DROPTASK2 overlay preserves expiry hold while removing native capture freeze before expiry. This corrects the native convenience Hold=false for the consumed demo expiry path, without production changes. No DOS execution claim.',
      updated_first_equality_class_inventory=[dict(name='NEW_BALL_TASK',status='previously proved replacement'),dict(name='PARTY_ON_TASK1',status='previously proved preservation'),dict(name='SETBALL',status='previously proved chute release'),dict(name='DROPTASK2',status='proved firing; expiry HOLDSTILL preserved, high state set, no release; QUIT37047')],bounded_inventory_review=dict(source='old threshold_state_classes',entries=inventory,classes_complete=not bool(remaining),reason='other held-ball continuation still includes independently relevant explicit HOLDSTILL-clearing DURINGFLASH; this conclusion is from its linked operands, not the old classes_complete boolean'),exactly_one_remaining_dependency=remaining,tests={})
    validate(r);return r

def main():
    p=argparse.ArgumentParser();p.add_argument('--harness',type=Path,default=Path('/private/tmp/pf-droptask2-reference'));p.add_argument('--prepare-only',action='store_true');p.add_argument('--replay',type=Path);a=p.parse_args()
    ps=[Path(os.environ[n]) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
    demo,full=pinned(*ps);d=Decoder();h,_=identities(demo['TABLE1.PRG'],ps[2]);e=linked(demo['TABLE1.PRG'],full['TABLE1.PRG'],d,h);cfg,nodes=post.configuration(demo['TABLE1.PRG'],h)
    if a.replay:identity=json.loads((a.harness/'identity.json').read_text());raw=json.loads(a.replay.read_text())
    else:
        identity=prepare(a.harness,ps[1],cfg);(a.harness/'identity.json').write_text(json.dumps(identity));(a.harness/'linked.json').write_text(json.dumps(e))
        if a.prepare_only:print('Producer binding PASS; C=E+130, F=C+58; targetE35810');return
        script=a.harness/'drop-script.json';script.write_text(json.dumps(TARGET));env=os.environ.copy();env.update(PF_POST_CONFIG=str(a.harness/'post-config.json'),PF_DROP_SCRIPT=str(script),PF_DROP_OUTPUT=str(a.harness/'drop-output.json'),GOCACHE='/private/tmp/pf-dmo0-go-cache')
        run=subprocess.run([str(ROOT/'.tools/go/bin/go'),'test','./internal/partyland','-run','^TestResearchDropWitness$','-count=1','-v'],cwd=a.harness,env=env,capture_output=True,text=True);(a.harness/'drop-tests.log').write_text(run.stdout+run.stderr);require(run.returncode==0,'fresh drop replay failed; see drop-tests.log');raw=json.loads((a.harness/'drop-output.json').read_text())
    any_capture=dict(found=raw['any_capture'],trials=656);search=dict(trials=2574,long_trials=1);r=report(raw,identity,e,nodes,any_capture,search);OUTPUT.write_text(json.dumps(r,separators=(',',':'))+'\n');print(r['class_verdict']);print(r['expiry_interleaving'])
if __name__=='__main__':main()
