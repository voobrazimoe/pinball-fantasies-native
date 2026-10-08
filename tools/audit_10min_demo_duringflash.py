#!/usr/bin/env python3
"""Linked DURINGFLASH first-equality certificate in isolated fresh A reference."""
import argparse,json,os,struct,subprocess
from pathlib import Path
import audit_10min_demo_droptask2 as drop
import audit_10min_demo_setball as setball
import audit_10min_demo_post_collision as post
import audit_10min_demo_deterministic_replay as replay
import audit_10min_demo_party_on as party
from audit_10min_demo_graph import Decoder,pinned,require
from audit_10min_demo_programs import identities
from audit_10min_demo_source_recurrence import linear
ROOT=replay.ROOT
OUTPUT=Path('/private/tmp/pf-dmo0-first-equality-duringflash.json')
TARGET=dict(Release=34711,Charge=19,LP=93,LD=61,LO=76,RP=186,RD=66,RO=82,ControlsEnd=35886)
SHORT=dict(Release=131,Charge=21,LP=79,LD=43,LO=48,RP=166,RD=44,RO=29,ControlsEnd=798)
ANCHORS=(
 (0x259d,'cmp','byte ptr [0xd4], 0xff'),(0x25a7,'mov','byte ptr [0xd4], 0xff'),
 (0x25e6,'mov','byte ptr [0x3026], 0xff'),(0x25f2,'call','0x2bbb'),
 (0x25f5,'call','0x232e'),(0x25fe,'call','0x2368'),(0x2607,'call','0x23ab'),
 (0x261b,'mov','si, 0x809'),(0x2621,'mov','word ptr [0xa6], 0x55'),
 (0x2627,'mov','dx, 0x242e'),(0x262a,'call','0x5b80'),
 (0x272e,'mov','dx, word ptr [0xa6]'),(0x2732,'mov','bx, 0x36ff'),
 (0x2735,'call','0x57c7'),(0x2747,'mov','dx, 0x2450'),(0x274a,'call','0x5b80'),
 (0x274d,'jmp','0x57ba'),(0x2750,'mov','dx, 0x1b'),(0x2753,'mov','bx, 0x3701'),
 (0x2756,'call','0x57c7'),(0x275f,'mov','dl, 0x15'),(0x2761,'call','0x5817'),
 (0x2766,'mov','cl, byte ptr [0xc51]'),(0x2776,'int','0x66'),
 (0x2778,'mov','byte ptr [0x3026], 0'),(0x277e,'mov','word ptr [0x2fdc], 0x101'),
 (0x2784,'mov','word ptr [0x2fde], 0x136'),(0x278a,'mov','byte ptr [0x3416], 0'),
 (0x27ac,'mov','word ptr [0x2fea], 0x627'),(0x27b2,'mov','word ptr [0x2fe8], 0xfdc1'),
 (0x27b8,'mov','byte ptr [0xd4], 0'),(0x27bd,'jmp','0x57ba'),
 (0x3ab3,'mov','di, 0x36c9'),(0x3ab6,'mov','cx, 0x32'),(0x3ab9,'rep stosw','word ptr es:[di], ax'),
 (0x515,'mov','byte ptr [0x3026], 0xff'),
 (0x20a6,'cmp','byte ptr [0x34e1], 0xff'),(0x20b1,'cmp','byte ptr [0xd3], 0xff'),
 (0x20d3,'mov','byte ptr [0x3485], 0'),(0x20d9,'call','0x1ddf'),
)

def firing(C,delay=85,producer_next=False,child_next=False,wait=27):
    return C+int(producer_next)+delay+int(child_next)+wait


def linked(b,a,d,h):
    # Reuse accepted common ordering, matrix, physical consumers/data and dragon
    # correspondence. No previously closed class is researched again here.
    e=setball.linked(b,a,d,h)
    for at,op,args in ANCHORS:d.expect(b,768,at,op,args)
    av=struct.unpack_from('<5H',a,0x1aadb+70);bv=struct.unpack_from('<5H',b,0x1ab67+70)
    require(av[:4]==bv[:4]==(260,312,268,321) and av[4]+768==0x2594 and bv[4]+768==0x259d,'GROPE linked gameplay rectangle')
    e['fresh_prefix_correspondence']=replay.compare_instructions(d,a,b,json.loads((ROOT/'tools/dmo0_duringflash_correspondence.json').read_text()))
    e['target_records']=[replay.record(a,b,'target_test_rectangle',0x1aab1+10*i,0x1ab3d+10*i,8,index=i) for i in range(4)]
    for i in range(4):require(struct.unpack_from('<H',b,0x1ab3d+10*i+8)[0]-struct.unpack_from('<H',a,0x1aab1+10*i+8)[0]==4,'target callback relocation')
    e['region_rectangles']=[replay.record(a,b,'area_rectangle',aa+10*i,bb+10*i,8,plane=plane,index=i) for plane,aa,bb,count in [('lower',0x1aadb,0x1ab67,14),('upper',0x1ab69,0x1abf5,12)] for i in range(count)]
    e['GROPE']=dict(A_record=0x1ab21,demo_record=0x1abad,rectangle=list(bv[:4]),A_callback=0x2594,demo_callback=0x259d,coordinate_test='ball center PixelX+8, PixelY+8+ScreenOffset; low level; first matching region; last-check suppresses repeated entry')
    e['producer_consumer_instructions']=[dict(site=x.address+768,op=x.mnemonic,args=x.op_str) for x in linear(d,b,0x259d,0x27c0,768)]
    writes=[]
    for x in linear(d,b,0x275f,0x27c0,768):
        for o in x.operands:
            if o.type==d.x86.X86_OP_MEM and o.access&d.cs.CS_AC_WRITE:
                require(not o.mem.base and not o.mem.index,'unbound indexed DURINGFLASH store')
                writes.append(dict(site=x.address+768,DS=o.mem.disp,width=o.size))
    require(0x3026 in {x['DS'] for x in writes} and not {0x34cf,0x34e6,0x34e8}.intersection(x['DS'] for x in writes),'hold clear with no expired/matrix store')
    e['DURINGFLASH_direct_writes']=writes
    # All helper scopes are inherited bounded native primitives. Raw cue bytes
    # and original effect/scroll text are never emitted to evidence.
    e['mask_patch_data']=[replay.record(a,b,'duck_dynamic_mask',0x19d40+off,0x19db0+off+0x100,size) for off,size in [(0x68b0,30),(0x68d0,30),(0x68f0,30),(0x6910,30),(0x6930,15),(0x6940,15)]]
    e['effect_arithmetic']=[replay.record(a,b,name+'_arithmetic',0x19d40+aa+2,0x19db0+bb+2,24) for name,aa,bb in [('DSCORE',0x7ed,0x809),('RIDE1',0xae3,0xaff),('BYGELSETB',0xb8c,0xba8)]]
    e['duck_selfmod_initial_counters']=[replay.record(a,b,'duck_task_initial_counter',aa,aa+4,1) for aa in (0x16f5,0x1775,0x17f5)]
    e['DSCORE_program']=party.expiry.program(b,h,struct.unpack_from('<H',b,0x19db0+0x809+26)[0]+0x19db0)
    e['default_effect']=dict(file=0x1a5b9,program=e['DSCORE_program'][0]['site'])
    e['shared_wait_writers']=dict(BEFOREFLASH=dict(call=0x2735,BX=0x36ff),DURINGFLASH=dict(call=0x2756,BX=0x3701),compare_increment_reset='paired WAITSYNCS5ac7..5ad8 writes only current BX word',reset=dict(file=0x3ab3,start_DS=0x36c9,words=50,end_exclusive=0x372d),concrete_scope='Fresh wait map zero; sole actual instances and one initial reset; no wait write by insertion, duplicate, subsequent reset or replacement. Not a global alias proof.')
    e['capture_source_HOLDSTILL']=0x25e6;e['release_source_HOLDSTILL_clear']=0x2778
    e['independent_pending_capture_candidate']=dict(task='WAIT_FOR_SPIN_TASK',file=0x20a6,guards=['SPECIALMODE DS:34e1','spin-ready DS:d3','jingle-ready DS:240c selects START_DROP alternative'],priority_clear_file=0x20d3,priority_DS=0x3485,then='SPIN at0x20df inserts selected reward task; direct priority clear occurs after timer expiry if this task is admitted',first_equality_reachability='UNKNOWN',equivalence_to_five_classes='NOT_PROVED')
    return e


def prepare(harness,canonical,cfg):
    (harness/'internal/partyland/dmo0_duringflash_test.go').unlink(missing_ok=True)
    identity=drop.prepare(harness,canonical,cfg)
    (harness/'internal/partyland/dmo0_droptask2_test.go').unlink()
    p=harness/'internal/partyland/game.go';replay.replace_once(p,' ResearchPartyMutation string',' ResearchSourceHold, ResearchCaptureHold bool\n ResearchPartyMutation string')
    p=harness/'internal/partyland/holes.go'
    replay.replace_once(p,'g.Dragon = true','g.Dragon = true;g.ResearchSourceHold=true;g.ResearchCaptureHold=true')
    replay.replace_once(p,'g.waitAt("DURINGFLASH", 27,','g.waitAt("DURINGFLASH", researchFlashLimit(g),')
    replay.replace_once(p,'g.Physics.SetBall(257, 310, -575, 1575, false)','g.ResearchCaptureHold=false;if g.ResearchPartyMutation!="remove_hold_clear" {g.ResearchSourceHold=false};g.Physics.Ball.Hold=g.ResearchSourceHold\n g.Physics.SetBall(257,310,-575,1575,false);if g.ResearchPartyMutation=="release_physics" {g.Physics.SetBall(258,310,-574,1575,true)}')
    p=harness/'internal/partyland/research_post_overlay.go';replay.replace_once(p,'g.ResearchExpired=true','g.ResearchExpired=true;g.ResearchSourceHold=true')
    p=harness/'internal/partyland/flow.go';replay.replace_once(p,'func (g *Game) newBall() {','func (g *Game) newBall() {g.ResearchSourceHold=true;g.ResearchCaptureHold=false')
    replay.replace_once(p,'g.Physics.Ball.Hold = false; g.Phase = Playing','g.Physics.Ball.Hold = false;g.ResearchSourceHold=false;g.ResearchCaptureHold=false; g.Phase = Playing')
    # Linked demo LOOSE_BALL file515 has a source HOLDSTILL store.
    p=harness/'internal/partyland/flow.go';replay.replace_once(p,'g.Phase = BallLost','g.Phase = BallLost;g.ResearchSourceHold=true;g.ResearchCaptureHold=false')
    (harness/'internal/partyland/dmo0_duringflash_test.go').write_bytes((ROOT/'tools/dmo0_duringflash_test.go.txt').read_bytes())
    return identity

boundary=party.boundary
index=drop.index

def validate(r):
    corrected = r.get('timing_oracle_basis') == 'FF-inclusive-v1'
    raw=r['fresh_input_replay'];w=raw['witness'];ix=index(w);rows={x['calculation']:x for x in w['rows']}
    b=lambda role,n=35998:ix[(role,n)][0]
    require(raw['deterministic_equal'] and raw['any_capture_deterministic_equal'] and w['script']==TARGET,'fresh deterministic fixed controls')
    require((w['capture'],w['BEFOREFLASH_fire'],w['DURINGFLASH_fire'])==(35886,35971,35998),'exact producer and firing')
    require(firing(35886)==35998 and firing(35886,child_next=True)==35999,'arithmetic/scan distinction')
    require(raw['any_capture']['script']==SHORT and (raw['any_capture']['capture'],raw['any_capture']['DURINGFLASH_fire'])==(798,910),'first real dragon family')
    require(not any(b('area_before:GROPE',35886)['live_task_sites']),'actual first free empty capture')
    capture=b('area_after:GROPE',35886)
    require(capture['live_task_sites']==['BEFOREFLASH']+['']*49,'producer actual slot0')
    require(capture['source_HOLDSTILL'] and capture['native_capture_Hold'] and capture['Dragon'],'source and capture holds set')
    for key in ('FiveMillion','BallFeature','JackpotNormal','JackpotTimed','SPECIALMODE'):require(not b('area_before:GROPE',35886)[key],'ordinary DSCORE branch guard')
    require(b('body_after:BEFOREFLASH',35971)['live_task_sites'][:2]==['BEFOREFLASH','DURINGFLASH'],'parent occupied and child first free slot1')
    for name,start,limit,slot in [('BEFOREFLASH',35886,85,0),('DURINGFLASH',35971,27,1)]:
        tid=b('wait_before:'+name,start)['task_ids'][slot]
        for n in range(start,start+limit+1):
            visits=ix.get(('wait_before:'+name,n),[]);require(len(visits)==1,'exact one shared-word visit')
            x=visits[0];require(x['waits'].get(name,0)==n-start,'age arithmetic')
            require(x['live_task_sites'][slot]==name and x['live_task_sites'].count(name)==1 and x['task_ids'][slot]==tid,'no duplicate/replacement slot identity')
            require(x['reset_count']==1 and not x['QUIT'],'no resetting reset or premature exit')
            require(x['source_HOLDSTILL'] and x['native_capture_Hold'],'capture survives each visit')
            if n<start+limit:require(b('wait_after:'+name,n)['waits'][name]==n-start+1,'increment chronology')
            else:require(b('body_after:'+name,n)['waits'][name]==0,'firing resets shared word')
    require(rows[35997]['waits']['DURINGFLASH']==27 and not rows[35997]['expired'],'exact end35997 checkpoint')
    for n in range(35886,35998):
        x=rows[n];require(x['ball']['Hold'] and x['source_HOLDSTILL'] and x['native_capture_Hold'],'captured suffix')
        require((x['ball']['X'],x['ball']['Y'],x['ball']['VX'],x['ball']['VY'],x['ball']['High'])==(263168,317440,0,0,False),'held physical state')
        require(not x['SPECIALMODE'] and not x['INH_EFF'] and x['reset_count']==1,'no interfering mode/reset')
    order=[x['boundary'] for x in w['boundaries'] if x['calculation']==35998]
    require(order.index('update_counters_before')<order.index('electronics_before')<order.index('expiry_installed')<order.index('electronics_after')<order.index('wait_before:DURINGFLASH')<order.index('body_after:DURINGFLASH')<order.index('tasks_after')<order.index('matrix_before')<order.index('matrix_after')<order.index('late_before')<order.index('late_after'),'complete equality phase order')
    for before,after in zip(ix[('step_before',35998)][:2],ix[('step_after',35998)][:2]):require(before['ball']==after['ball'] and before['ball']['Hold'],'early captured passes unchanged')
    require(len(ix[('step_before',35998)])==3 and len(ix[('step_after',35998)])==3,'two early and one late physical pass')
    require(b('electronics_before')['timer']==35997 and not b('electronics_before')['expired'],'old timer sampled')
    installed=b('expiry_installed');body=b('body_after:DURINGFLASH')
    require(installed['timer']==35998 and installed['expired'] and installed['source_HOLDSTILL'] and installed['audio_priority']==255,'expiry flags and priority')
    state=dict(program=0x1ba17,cursor=0x1ba19,op='_CLEAR4',remaining=5)
    for role in ('expiry_installed','body_after:DURINGFLASH','tasks_after','matrix_before'):require(b(role)['linked_matrix']==state and b(role)['expired'] and b(role)['audio_priority']==255,'expiry cursor, flag and priority preserved')
    ball=body['ball'];require((ball['PixelX'],ball['PixelY'],ball['X'],ball['Y'],ball['VX'],ball['VY'],ball['High'],ball['Lost'])==(257,310,263168,317440,-575,1575,False,False),'exact release physics')
    require(not ball['Hold'] and not body['source_HOLDSTILL'] and not body['native_capture_Hold'] and not body['Dragon'] and body['expired'],'source clear and native capture removal independently bound')
    require(not any(b('tasks_after')['live_task_sites']),'suicide and empty surviving scan')
    require(b('matrix_after')['linked_matrix']==dict(state,remaining=4),'clear5 to4')
    require(b('late_after')['expired'] and not b('late_after')['source_HOLDSTILL'] and not b('late_after')['ball']['Lost'] and b('late_after')['linked_matrix']==dict(state,remaining=4),'end-of-equality expiry/release state')
    late=b('late_after')['ball'];require((late['X'],late['Y'],late['PixelX'],late['PixelY'],late['VX'],late['VY'],late['Rotation'])==(262593,319015,256,311,-575,1582,-4200),'same-calculation late physics result')
    reads=[x for x in w['physics_reads'] if x[1]==35998];require(sum(x[0]=='ring' for x in reads)==44 and not any(x[0] in ('response','sine') for x in reads),'all44 late ring samples empty; stale material6 is not consumed')
    require(all(x[-1]==0 for x in reads if x[0]=='ring'),'empty mask bits')
    drain=b('drain_entry',36173);req=b('RESEARCH_MINUTE5_request',36173);result=b('RESEARCH_MINUTE5_result',36173)
    require(drain['expired'] and drain['SCORECHANGED'] and not drain['INH_EFF'] and not drain['SPECIALMODE'] and drain['audio_priority']==255,'real scored expired drain admission')
    require(drain['ball']['Lost'] and (drain['ball']['X'],drain['ball']['Y'],drain['ball']['VX'],drain['ball']['VY'])==(167360,590318,538,796),'real terminal released trajectory')
    require(req['source_HOLDSTILL'] and req['ball']['Hold'] and result['effect_accepted'] and result['linked_matrix']==state,'source hold and actual expiry entry restart')
    require(w['final']['QUIT'] and w['final']['calculation']==(37220 if corrected else 37212),'real no-input quit')
    for n in range(35887,37213):
        x=rows[n];require(not any(x['input'].values()),'no further input');require(x['reset_count']==1 and not x['INH_EFF'] and not x['SPECIALMODE'],'no subsequent reset/mode/inhibit')
    for n in range(35998,37213):require(rows[n]['expired'],'sticky expired')
    matrices=[(x['calculation'],x['event']['Label']) for x in w['milestones'] if x['calculation']>=35998 and x['event']['Kind']=='MatrixStarted']
    require(matrices==[(35998,'RESEARCH_EXPIRY'),(36173,'RESEARCH_EXPIRY')],'one actual restart only')
    require(not any(x['boundary'].startswith('area_before:') for x in w['boundaries'] if x['calculation']>35886),'no post-release area/capture')
    for m in ('wait_plus_one','child_earlier_slot'):require(raw[m]['DURINGFLASH_fire']==35999,'wait/ordering mutation delays actual fire')
    m=raw['child_earlier_slot'];require(boundary(m,'wait_before:DURINGFLASH',35972)['live_task_sites'][0]=='DURINGFLASH','earlier slot mutation next scan')
    m=raw['remove_hold_clear'];x=boundary(m,'body_after:DURINGFLASH',35998);require(x['source_HOLDSTILL'] and not x['native_capture_Hold'] and x['ball']['Hold'],'native capture false alone does not clear source')
    require(boundary(m,'late_after',35998)['ball']['Y']==317440,'hold clear mutation suppresses movement')
    require(boundary(raw['release_physics'],'body_after:DURINGFLASH',35998)['ball']['High'],'release mutation observed')
    inventory=r['independent_inventory_producer'];iv=inventory['found'];require(inventory['trials']==333 and iv['arcade_pending_producer']==1001 and iv['script']==dict(Release=33,Charge=21,LP=20,LD=5,LO=3,RP=171,RD=100,RO=15,ControlsEnd=1001),'fresh independent gameplay arcade producer')
    require('arcade.func1' in boundary(iv,'area_after:GROPB',1001)['live_task_sites'],'real pending arcade task; threshold admission still unproved')
    require(r['class_verdict']=='FIRST_EQUALITY_DURINGFLASH_REACHABLE' and r['expiry_interleaving']=='EXPIRY_INTERLEAVING = NOT_PROVED','honest class versus completeness verdict')
    require(r['exactly_one_remaining_dependency']['task']=='WAIT_FOR_SPIN_TASK','independent admission dependency')
    return True


def report(raw,identity,e,nodes,search,inventory_producer=None):
    w=raw['witness'];entry=boundary(w,'expiry_installed',35998)['matrix_pc']-1
    for v in raw.values():
        if not isinstance(v,dict) or 'boundaries' not in v:continue
        for x in v['rows']+v['boundaries']+[v['final']]:
            if entry<x['matrix_pc']<=entry+len(nodes):x['linked_matrix']=party.matrix_state(x,nodes,entry)
            elif 282<x['matrix_pc']<=282+len(e['DSCORE_program']):x['linked_matrix']=party.matrix_state(x,e['DSCORE_program'],282)
            elif not x['matrix_active']:x['linked_matrix']=dict(program=0x1b88e,cursor=0,op='NODOT')
            else:x['linked_matrix']=dict(native_pc=x['matrix_pc'],op=x['matrix_op'],remaining=x['matrix_remaining'])
    b=lambda role,n=35998:boundary(w,role,n)
    chronology={name:[dict(calculation=x['calculation'],slot=slot,wait_DS=ds,before=x['waits'].get(name,0),after=0 if x['calculation']==end else x['waits'].get(name,0)+1,fire=x['calculation']==end,task_id=x['task_ids'][slot]) for x in w['boundaries'] if x['boundary']=='wait_before:'+name] for name,slot,ds,end in [('BEFOREFLASH',0,0x36ff,35971),('DURINGFLASH',1,0x3701,35998)]}
    reads=w['physics_reads'];responses=[x for x in reads if x[0]=='response'];transitions=[];prior=None
    for x in w['rows']:
        if x['calculation']>=35998 and x['matrix_op']!=prior:
            transitions.append(dict(calculation=x['calculation'],matrix=x['linked_matrix'],text_left=x['matrix_text_left']));prior=x['matrix_op']
    remaining=dict(task='WAIT_FOR_SPIN_TASK',file=0x20a6,priority_clear=0x20d3,dependency='Fresh input-only first-equality admission or exclusion of the arcade pending-spin producer, including its actual priority-zero/reward suffix. This independent linked task is not covered by the five proved witnesses.',gameplay_producer='fresh Release33/Charge21 prefix reaches GROPB at1001 with real arcade.func1 pending; reproduced twice',reachability='UNKNOWN_AT_FIRST_EQUALITY',why_independent='Direct CURRENT_PRIORITY=0 after expiry can admit queued reward effects; no expiry guard. No threshold witness or exclusion invariant was established here.')
    inventory=[dict(name=n,status=s) for n,s in [('NEW_BALL_TASK','previously proved replacement'),('PARTY_ON_TASK1','previously proved flagged preservation'),('SETBALL','previously proved chute release'),('DROPTASK2','previously proved hold-preserving state change'),('DURINGFLASH','proved source-hold clearing play-field release; scored drain36173 restarts expiry; QUIT37220'),('nonfiring ages','same task family, no equality body; not extra firing classes; later suffix requires concrete admission'),('early effects','early physics/area effects precede expiry installation and are superseded; this alone does not exclude pending task effects'),('ordinary active/held without releasing task','expiry hold remains in equality suffix'),('other pending capture/mode producers','independent arcade WAIT_FOR_SPIN_TASK priority-clear edge remains unclassified at equality')]]
    r=dict(timing_oracle_basis='FF-inclusive-v1',class_verdict='FIRST_EQUALITY_DURINGFLASH_REACHABLE',expiry_interleaving='EXPIRY_INTERLEAVING = NOT_PROVED',status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',production_base=replay.BASE,research_head='37ff8bc38d7af4a09a70319675d6e505b0bd6a5e',native_reference_identity=identity,
      producer_graph=dict(gameplay=e['GROPE'],nodes=[dict(name='GROPE',file=0x259d,capture=35886,HOLDSTILL_store=0x25e6,guards='Dragon false; low area center match; no active jackpot/ball/5m or SPECIALMODE',mode_delays=dict(default=85,jackpot=410,ball=320,five_million=160,special_ball_or_5m=15)),dict(name='BEFOREFLASH',file=0x272e,wait_DS=0x36ff,limit=85,insert=35886,slot=0,first_visit=35886,fire=35971),dict(name='DURINGFLASH',file=0x2750,wait_DS=0x3701,limit=27,insert=35971,slot=1,first_visit=35971,fire=35998,release=35998)],edges=['GROPE -> first-free BEFOREFLASH','BEFOREFLASH equality -> flash21 and first-free DURINGFLASH -> parent suicide','DURINGFLASH equality -> endflash21, SNEWBALL, source hold clear, release, Dragon false, suicide'],allocation='50 slots; first free; parent occupied until suicide; ascending scan; insertion does not zero wait'),
      producer_equation=dict(C=35886,B=35971,F=35998,equation='B=C+85; first_DURINGFLASH=B; F=B+27=C+112',target='C=35998-85-27=35886',predecessor=35971,neighbors=dict(C_minus_one=35997,C_plus_one=35999,incorrect_next_child=35999),visit_rule='compare before increment: wait85 fires on visit86, wait27 on visit28; both first visits are same scan here'),
      input_script=dict(target=TARGET,first_dragon=SHORT,controls='Down34692..34710; Release34711; Left iff(n-34711+76)%93<61, Right iff(n-34711+82)%186<66 on34711..35886; every other control false; all controls false after35886',construction='fresh New(DecodePartyLand(A), A), Configure(Legacy); no state/task/wait/timer seed'),
      search=dict(trials=search['trials'],full_delayed_candidates=search['long_trials'],seed=1,candidate_rule='short firing residue mod128 proposes delay only; every delayed candidate is reconstructed fresh; no shift theorem assumed',finite_search='4579 deterministic short scripts, 2 delayed fresh candidates; accepted fixed scripts replayed twice'),
      linked_evidence=e,independent_inventory_producer=inventory_producer,task_wait_chronology=chronology,fresh_input_replay=raw,
      checkpoints=dict(capture_before=b('area_before:GROPE',35886),capture_after=b('area_after:GROPE',35886),predecessor_body=b('body_after:BEFOREFLASH',35971),pre_firing_35997=next(x for x in w['rows'] if x['calculation']==35997)),
      full_scan35998=[x for x in w['boundaries'] if x['calculation']==35998],task_scan35998=[dict(slot=i,DS=0x3417+2*i,initial='DURINGFLASH' if i==1 else 'empty',operation='wait27 fires and suicides' if i==1 else 'empty slot',after='empty') for i in range(50)],
      DURINGFLASH_body=b('body_after:DURINGFLASH'),late_physics=dict(before=b('late_before'),after=b('late_after'),consumed_reads=[x for x in reads if x[1]==35998],explanation='44 zero low-mask ring bits; no material/sine response; early ramp index12752 and5 negative level records; no late ramp/level/area callback; material6 snapshot is stale. Movement uses retained GY7 and Rotation decay -4202 to -4200.'),
      concrete_continuation=dict(control='no input',trajectory=w['rows'],collision_responses=responses,physical_events=w.get('physical_events',[]),area_target_events=[x for x in w['boundaries'] if x['calculation']>35998 and (x['boundary'].startswith('area_before:') or x['boundary']=='target_callback')],drain_calculation=36173,drain=b('drain_entry',36173),effect_request=b('RESEARCH_MINUTE5_request',36173),effect_result=b('RESEARCH_MINUTE5_result',36173),expiry_restarts=1,expiry_replacements_by_other_program=0,matrix_progression=transitions,QUIT=37220,final=w['final']),
      bounded_inventory_review=dict(entries=inventory,classes_complete=False,argument='The five named firing bodies are now witnessed. This does not form an exhaustive pending-task census: linked arcade WAIT_FOR_SPIN_TASK has a direct priority-clear/reward path after expiry, absent in all five traces. Nonfiring ages and superseded early effects do not discharge that independent admission. No claim of full coverage or universal termination.'),exactly_one_remaining_dependency=remaining,
      scope='Accepted canonical-A native reference, linked consumed-path correspondence and explicit demo overlay; no DOS execution/admission/alias/IRQ claim. Source HOLDSTILL tracked independently from native Ball.Hold and dragon capture. Demo LOOSE_BALL source hold store515 is modeled. Native BallLost skips further physics. No payload bytes.',tests={})
    validate(r);return r


def main():
    p=argparse.ArgumentParser();p.add_argument('--harness',type=Path,default=Path('/private/tmp/pf-duringflash-reference'));p.add_argument('--replay',type=Path);p.add_argument('--inventory-replay',type=Path,default=Path('/private/tmp/pf-flash-inventory.json'));p.add_argument('--search',action='store_true');p.add_argument('--prepare-only',action='store_true');args=p.parse_args()
    paths=[Path(os.environ[n]) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
    demo,full=pinned(*paths);d=Decoder();h,_=identities(demo['TABLE1.PRG'],paths[2]);e=linked(demo['TABLE1.PRG'],full['TABLE1.PRG'],d,h);cfg,nodes=post.configuration(demo['TABLE1.PRG'],h)
    if args.replay:identity=json.loads((args.harness/'identity.json').read_text());raw=json.loads(args.replay.read_text())
    else:
        identity=prepare(args.harness,paths[1],cfg);(args.harness/'identity.json').write_text(json.dumps(identity))
        if args.prepare_only:print('linked producer PASS');return
        script=args.harness/'flash-script.json';script.write_text(json.dumps(TARGET));out=args.harness/'flash-output.json'
        env=os.environ.copy();env.update(PF_POST_CONFIG=str(args.harness/'post-config.json'),PF_FLASH_SCRIPT=str(script),PF_FLASH_OUTPUT=str(out),GOCACHE='/private/tmp/pf-dmo0-go-cache')
        run=subprocess.run([str(ROOT/'.tools/go/bin/go'),'test','./internal/partyland','-run','^TestResearchFlash'+('Search' if args.search else 'Witness')+'$','-count=1','-v'],cwd=args.harness,env=env,capture_output=True,text=True);(args.harness/'flash-tests.log').write_text(run.stdout+run.stderr);require(run.returncode==0,'fresh flash replay failed; see isolated flash-tests.log');raw=json.loads(out.read_text())
        if args.search:print('finite search',raw['trials'],raw['found']['script']);return
    if not args.replay:
        invout=args.harness/'flash-inventory.json';env['PF_FLASH_OUTPUT']=str(invout)
        invrun=subprocess.run([str(ROOT/'.tools/go/bin/go'),'test','./internal/partyland','-run','^TestResearchFlashInventoryProducer$','-count=1','-v'],cwd=args.harness,env=env,capture_output=True,text=True)
        (args.harness/'flash-inventory-tests.log').write_text(invrun.stdout+invrun.stderr);require(invrun.returncode==0,'fresh inventory producer replay failed')
        inventory_producer=json.loads(invout.read_text())
    else:inventory_producer=json.loads(args.inventory_replay.read_text()) if args.inventory_replay.exists() else None
    require(inventory_producer is not None,'fresh independent arcade inventory producer metadata required')
    r=report(raw,identity,e,nodes,dict(trials=4579,long_trials=2),inventory_producer);OUTPUT.write_text(json.dumps(r,separators=(',',':'))+'\n');print(r['class_verdict']);print(r['expiry_interleaving'])
if __name__=='__main__':main()
