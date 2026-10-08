#!/usr/bin/env python3
"""Real post-35998 continuation under the accepted native A semantics.

Original payload stays private. Isolated copied source contains demo-only
consumed timer, expired-drain and matrix overlays; production is never edited.
"""
import argparse
import copy
import json
import os
from pathlib import Path
import subprocess
import struct
import audit_10min_demo_first_equality_collision as collision
import audit_10min_demo_deterministic_replay as replay
import audit_10min_demo_expiry_interleaving as expiry
from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_programs import identities

ROOT = replay.ROOT
INPUT = Path('/private/tmp/pf-dmo0-first-equality-collision.json')
OUTPUT = Path('/private/tmp/pf-dmo0-post-collision-termination.json')


def configuration(b, h):
    nodes = expiry.program(b,h,0x1ba17)
    require(tuple(n['op'] for n in nodes)==expiry.EXPIRY,'expiry operations')
    require([n['args'] for n in nodes]==[[],[7313],[1],[18101,344],[100],[1],[7412],[256],[100],[0]],'expiry operands')
    # Pin the reviewed linked cadence itself, not just the program operands.
    d = Decoder()
    for at, op, operands in (
        (0x7206, 'mov', 'cx, 2'),
        (0x721a, 'cmp', 'byte ptr es:[si + 0x14], 0xff'),
        (0x721f, 'jne', '0x6f36'),
        (0x72cc, 'dec', 'byte ptr cs:[0x7000]'),
        (0x72d1, 'jne', '0x6fec'),
        (0x72d6, 'mov', 'byte ptr cs:[0x7000], 8'),
        (0x72dd, 'inc', 'si'),
        (0x72f7, 'loop', '0x6ffb'),
        (0x72fb, 'jmp', '0x6f09'),
    ):
        d.expect(b, 768, at, op, operands)
    commands=[];lengths=[]
    for n in nodes:
        args=n['args'];op=n['op'];nums={str(i):a for i,a in enumerate(args)}
        if op=='_SCROLL':
            start=0x19db0+args[0];length=b.index(255,start)-start
            require(21<=length<200,'bounded scroll text')
            # Every actual glyph and the owned zero-cell representative has
            # only bounded literal dot stores then RET. No phase/clock branch.
            for ch in set(b[start:start+length]) | {0}:
                for table, code in ((0x5f00, 0x7f70), (0x6100, 0x7400)):
                    q = code + struct.unpack_from('<H', b, 0x19db0+table+2*ch)[0]
                    for _ in range(1024):
                        if b[q] == 0xc3:
                            break
                        require(b[q:q+2] in (b'\x88\x87', b'\x88\xa7'), 'glyph is not literal MOV/RET')
                        q += 4
                    else:
                        require(False, 'unbounded glyph stores')
            lengths.append(length);args=['RESEARCH_SCROLL'+str(len(lengths))];nums={}
        elif op=='_PRINT13_NUMBER':args=['SIFFRORNA',str(args[1])];nums={'1':nums['1']}
        else:args=list(map(str,args))
        commands.append(dict(op=op,args=args,nums=nums))
    commands.append(dict(op='0',args=[]))
    require(lengths==[98,88],'consumed scroll lengths')
    require(list(b[0x19db0+0xc9d:0x19db0+0xc9d+3])==[13,0,255],'demo cue metadata')
    require(not any(b[0x1a4a1+2:0x1a4a1+26]),'expiry effect arithmetic')
    return dict(TimingBasis="FF-inclusive-v1",Commands=commands,ScrollLengths=lengths,Jingle=[13,0,255]),nodes


def prepare(harness, canonical, cfg):
    # Remove only this pass's generated files before refreshing copied sources.
    for name in ('research_post_overlay.go','dmo0_post_collision_test.go','dmo0_post_pause_test.go'):
        (harness/'internal/partyland'/name).unlink(missing_ok=True)
    # Reuse the passive handoff constructor, without changing saved artifact.
    fresh,identity=collision.reconstruct(harness,canonical,ROOT/'.tools/go/bin/go')
    collision.handoff(json.loads(collision.INPUT.read_text())['deterministic_replay'],fresh)
    p=harness/'internal/partyland/game.go'
    replay.replace_once(p,' ResearchObserve func(string)',
        ' ResearchExpiryVisits int\n ResearchDemoEnabled, ResearchExpired, ResearchQuit bool\n ResearchLastCalculation uint64\n ResearchTimer uint16\n ResearchAfterWait func(string)\n ResearchObserve func(string)')
    replay.replace_once(p,'\tg.checkAreas()',
        '\tresearchElectronics(g)\n\tg.checkAreas()')
    # Loss tasks emulate the common electronics suffix in the native reference.
    s=p.read_text();anchor='\t\tg.Display.Flash() // MATRIX_BLINKOR precedes DO_TASKS.'
    require(s.count(anchor)==2,'native loss suffix anchors')
    p.write_text(s.replace(anchor,'\t\tresearchElectronics(g)\n'+anchor))
    p=harness/'internal/partyland/timing.go'
    replay.replace_once(p,'if !g.waitReady(site, n) {\n\t\t\treturn false\n\t\t}\n\t\tf()',
        'if !g.waitReady(site, n) {\n\t\t\treturn false\n\t\t}\n\t\tf()\n if g.ResearchAfterWait!=nil {g.ResearchAfterWait(site)}')
    replay.replace_once(p,'\t\tcase "_CLEAR4":',
        '\t\tcase "_FADE":\n m.remaining=uint16(c.Num(0))\n\t\tcase "QUIT":\n g.ResearchQuit=true;g.Physics.Stopped=true;m.active=false;g.Phase=GameOver;return\n\t\tcase "_CLEAR4":')
    # A real accepted effect can enter expiry without researchInstallExpiry.
    # Observe installation at the shared matrix start, after label resolution.
    replay.replace_once(p, '\tg.emit("MatrixStarted", label, 0)',
        '\tif label=="RESEARCH_EXPIRY" {g.ResearchExpiryVisits=0}\n\tg.emit("MatrixStarted", label, 0)')
    # Passive admission counter, after budget and active-state checks. This
    # does not determine completion and resets only on a real expiry install.
    replay.replace_once(p, '\tdone := false\n\tswitch m.op {',
        '\tif !m.sourceProgram && m.next>expiryIndex && m.next<=expiryIndex+10 {g.ResearchExpiryVisits++}\n\tdone := false\n\tswitch m.op {')
    # Only the linked scored+expired branch differs from A. Effect still uses
    # the actual cue/admission consumer; an effect pointer alone never installs.
    p=harness/'internal/partyland/flow.go'
    anchor='\tg.Audio.Priority = 0\n\tif g.ResearchObserve!=nil {g.ResearchObserve("LOSTBALL_request")}'
    replay.replace_once(p,anchor,
        '\tif g.ResearchDemoEnabled&&g.ResearchExpired {\n if g.ResearchObserve!=nil {g.ResearchObserve("RESEARCH_MINUTE5_request")}\n g.effect("RESEARCH_MINUTE5",0,0)\n if g.ResearchObserve!=nil {g.ResearchObserve("RESEARCH_MINUTE5_result")}\n return\n }\n'+anchor)
    # Restore demo PARTY_ON task body, which sets its flag before NEW_BALL.
    replay.replace_once(p,'g.waitAt("PARTY_ON_TASK1", 30, g.newBall)',
        'g.waitAt("PARTY_ON_TASK1", 30, func(){g.partyFlash=true;g.newBall()})')
    # Split ordinary helper declarations from test-only constructors so frontend
    # imports the same overlay Game without importing testing/harness functions.
    src=(ROOT/'tools/dmo0_post_collision_reference.go.txt').read_text()
    cut=src.index('func postSnapshot(')
    common=src[:cut].replace(' "pinballfantasies/internal/physics"\n','').replace(' "testing"\n','')
    (harness/'internal/partyland/research_post_overlay.go').write_text(common)
    test='package partyland\nimport("encoding/json";"os";"testing";"pinballfantasies/internal/physics")\n'+src[cut:]
    (harness/'internal/partyland/dmo0_post_collision_test.go').write_text(test)
    (harness/'internal/partyland/dmo0_post_pause_test.go').write_bytes((ROOT/'tools/dmo0_post_collision_pause_test.go.txt').read_bytes())
    # Demo equality cue was omitted from primitive-only previous handoff probe.
    # Replay it at its established phase before tasks; never change ball/waits.
    p=harness/'internal/partyland/dmo0_collision_handoff_test.go'
    replay.replace_once(p,'g.audioTick();g.Display.Flash();g.runTasks()',
        'g.audioTick();if n==35998 {a:=g.musicClock();a.Play(g.Display.Jingle("S_GAMEOVER2"),62,timing.Cues);g.storeMusicClock(a)};g.Display.Flash();g.runTasks()')
    (harness/'post-config.json').write_text(json.dumps(cfg))
    return identity


def classify(finite_run=None, cycle=None):
    """A finite run can establish QUIT; it can never establish nontermination."""
    if cycle and cycle.get('closed_transition') and cycle.get('gameplay_state_equal') and cycle.get('no_exit'):
        return 'EVENTUAL_QUIT_NOT_GUARANTEED'
    return 'EVENTUAL_TERMINATION_UNKNOWN'


def linked(b, full, d, h):
    collision.linked(b,d,h)
    configuration(b,h)
    for at,op,args in (
        (0x5cd9,'pushaw',''),(0x5cdb,'inc','word ptr [0x34cd]'),
        (0x5cdf,'cmp','word ptr [0x34cd], 0x8c9e'),
        (0x5d2,'cmp','byte ptr [0x34cf], 0xff'),
        (0x622,'mov','si, 0x6f1'),(0x625,'call','0x5c14'),
        (0x5256,'push','word ptr [bx + 2]'),
        (0x5259,'pop','word ptr [0x34e8]'),
        (0x525d,'mov','word ptr [0x34e6], 0x52aa'),
        (0x55aa,'dec','si'),(0x2ebb,'cmp','byte ptr [0x5b2], 0xff'),
        (0x2ec0,'jne','0x2bd4'),(0x2ed4,'ret',''),(0x3cdb,'mov','byte ptr [0x3025], 0')):
        d.expect(b,768,at,op,args)
    require(next(k for k,v in h.items() if v['op']=='_FADE')+768==0x5256,'fade binding')
    require(next(k for k,v in h.items() if v['op']=='QUIT')+768==0x3cdb,'quit binding')
    d.expect(full,768,0x2eb2,'cmp','byte ptr [0x5b2], 0xff')
    d.expect(full,768,0x2eb7,'jne','0x2bcb')
    d.expect(full,768,0x2ecb,'ret','')
    pairs=replay.compare_instructions(d,full,b,json.loads((ROOT/'tools/dmo0_post_collision_correspondence.json').read_text()))
    for role,aa,bb in [('DSCORE',0x19d40+0x7ed,0x19db0+0x809),('BYGELSETB',0x19d40+0xb8c,0x19db0+0xba8)]:
        require(full[aa+2:aa+26]==b[bb+2:bb+26],role+' arithmetic')
    import struct
    for i in (7,8):
        aa=struct.unpack_from('<5H',full,0x1aadb+10*i)
        bb=struct.unpack_from('<5H',b,0x1ab67+10*i)
        require(aa[:4]==bb[:4] and bb[4]-aa[4]==9,'consumed area binding')
    require(struct.unpack_from('<H',b,0x1a4a1)[0]==0xc9d,'expired-drain cue record')
    require(struct.unpack_from('<H',b,0x1a4a1+26)[0]+0x19db0==0x1ba17,'expired-drain matrix record')
    # Selected queued-task direct stores and the two concretely indexed writes.
    writers=[]
    from audit_10min_demo_source_recurrence import linear
    for task,lo,hi in [('SOUNDBRICKUPP',0xfa8,0xfce),('SOUNDNEWBALL',0xfce,0xff4),('SETBALL',0xff4,0x104c)]:
        direct=[]
        for x in linear(d,b,lo,hi,768):
            for operand in x.operands:
                if operand.type==d.x86.X86_OP_MEM and operand.access&d.cs.CS_AC_WRITE:
                    mem=operand.mem
                    require(not mem.base and not mem.index,'new indexed direct task writer')
                    require(not mem.disp<=0x34cf<mem.disp+operand.size,'expired writer in queued task')
                    direct.append(dict(site=x.address+768,offset=mem.disp,width=operand.size))
        writers.append(dict(task=task,extent=[lo,hi],direct_writes=direct))
    return dict(additional_consumed_correspondence=pairs,queued_task_writes=writers,
        indexed_writes={'WAITSYNCS':'BX = linked task DS:36cf/36d1/36d3, compare before increment or zero on match',
                        'SUICIDE':'current live task slot, DS:3417..341b for these three tasks'},
        matrix_writes='accepted A CLEAR4 and PRINT5 consumers write bounded Display state; native state has no expired writer; linked SHOWPLAYERSTS has only these operations',
        sticky_scope='consumed native contract for this trace only; no global DOS writer or IRQ closure',
        timer_admission='one logical electronics admission per examined unpaused calculation, including native BallLost common task/matrix suffix; drain samples old flag before this calculation admission',
        fade='linked SI decrement from 256; palette and volume stores do not gate NEXT_A. QUIT is entry at interrupts_on=false, before out-of-scope teardown.')


def validate(r):
    corrected = r.get('configuration',{}).get('TimingBasis') == 'FF-inclusive-v1'
    start=r['starting_state'];require(start['calculation']==35998 and start['timer']==35998,'start calculation')
    require(start['expired'] and start['ball']['Hold'],'start flags')
    require(start['live_task_sites'][:3]==['SOUNDNEWBALL','SETBALL','SOUNDBRICKUPP'],'start slots')
    require([start['WAITLIST'][k] for k in ('0x36d1','0x36d3','0x36cf')]==[0,1,1],'start age asymmetry')
    for key in ('no_input','single_launch','periodic_launch'):
        v=r['continuations'][key]
        # Every selected native handoff field is retained; matrix remaining was
        # omitted in the old primitive probe and is checked separately here.
        for name in r['native_identity_fields']:
            require(v['start'][name]==start[name],'native start identity: '+name)
        require(v['start']['matrix_remaining']==4,'admitted start matrix visit')
        setball=next(x for x in v['rows'] if x.get('boundary','').startswith('immediately after SETBALL'))
        require(setball['calculation']==36078 and setball['expired'],'SETBALL/expired')
        require((setball['ball']['PixelX'],setball['ball']['PixelY'],setball['ball']['VX'],setball['ball']['VY'],setball['ball']['Hold'])==(297,530,10,0,False),'SETBALL body state')
        require([x for x in setball['live_task_sites'] if x]==['SETBALL'],'SETBALL executing slot identity')
        end=next(x for x in v['rows'] if x['calculation']==36078 and not x.get('boundary'))
        require(not any(end['live_task_sites']),'no live tasks after SETBALL returns')
        require(next(x for x in v['matrix_progression'] if x['op']=='0')['calculation']==36004,'show terminator')
    no=r['continuations']['no_input'];require(not no['drain_boundaries'],'no-input must not invent drain')
    require(no['final']['QUIT'] and no['final']['calculation']==(102583 if corrected else 102575),'no-input concrete quit')
    single=r['continuations']['single_launch'];require(single['drain_boundaries'][0]['calculation']==36288 and not single['drain_boundaries'][0]['SCORECHANGED'],'actual unscored drain')
    scored=r['continuations']['periodic_launch'];req,res=scored['drain_boundaries'][1:3]
    require(req['calculation']==37084 and req['SCORECHANGED'] and req['expired'],'actual scored expired drain')
    require(not req['INH_EFF'] and not req['SPECIALMODE'] and res['effect_accepted'],'actual effect admission')
    require(req['audio_priority']==255 and res['audio_priority']==255,'linked expiry cue priority')
    require(res['matrix_remaining']==5 and res['matrix_pc']==r['expiry_native_entry']+1,'reinstall from entry')
    require(scored['final']['QUIT'] and scored['final']['calculation']==(38133 if corrected else 38125),'second expiry concrete quit')
    require(r['eventual']==classify(cycle=r['pause_cycle']),'cycle theorem')
    return True


def enrich(r):
    n=r['continuations']['no_input'];p=r['continuations']['periodic_launch'];s=r['continuations']['single_launch']
    for name,trace in r['continuations'].items():
        mode=trace['input']['mode'];mutation=trace['input']['mutation']
        trace['input']=dict(mode=mode,mutation=mutation,script=('all controls false' if mode==0 else 'Down36079..36100; Release36101; Right36101..37100' if mode==1 else 'Down36206..36227; Release36228; d=n-36228>=0: Left=(d+46)%52<8, Right=(d+22)%30<21'))
    r['audio_handoff_completion']={
        'input_primitive_A_priority':r['input_artifact_state']['audio_priority'],
        'consumed_demo_priority':r['starting_state']['audio_priority'],
        'provenance':'already proved 35998 expiry cue, linked DS:c9d={position13,repeat0,priority255}; interpreted before existing NEW_BALL reset. Native A source record priority1 is not the demo record. Reset sets readiness but does not clear current priority; lower-priority spring is rejected.',
        'unchanged':'every input ball/flipper/task/wait/guard/score/matrix/timer field is revalidated; input artifact is not rewritten'}
    r['immediate_suffix']={
        'initial_ages':{'slot0_SOUNDNEWBALL':0,'slot1_SETBALL':1,'slot2_SOUNDBRICKUPP':1},
        'fires':{'SOUNDBRICKUPP':36003,'SOUNDNEWBALL':36049,'SETBALL':36078},
        'rule':'compare before increment, reset to zero on equality; subsequent scans begin at slot0',
        'SETBALL_state':next(x for x in n['rows'] if x.get('boundary','').startswith('immediately after SETBALL')),
        'calculation_36078_end_state':next(x for x in n['rows'] if x['calculation']==36078 and not x.get('boundary')),
        'expired_writer_check':r['linked_evidence']['queued_task_writes']}
    r['SHOWPLAYERSTS']={
        'start_program':0x1b88e,'start_cursor':0x1b890,'start_clear_remaining':4,
        'admitted_visits_including_35998':7,'additional_admitted_visits':6,
        'progression':[{'calculations':[35998,36001],'op':'_CLEAR4','next_cursor':0x1b890},
            {'calculation':36002,'dispatch':'_PRINT5','next_cursor':0x1b896},
            {'calculation':36003,'dispatch':'_PRINT5','next_cursor':0x1b89c},
            {'calculation':36004,'dispatch':'terminator','next_cursor':0,'routine':'NODOT/DUMRET'}],
        'automatic_program_at_terminator':None,'automatic_expiry_restore':False,
        'SETBALL_replaces_program':False,
        'idle_semantics':'ONLY_SCORE; in_chute=true blocks automatic player panel. The native next array index is diagnostic, not a saved DOS NEXT_A.'}
    r['no_input_summary']={
        'first_subsequent_drain':None,'score_changed':False,'score':50030,
        'settled_ball':{'PixelX':301,'PixelY':537,'VX':0,'VY':7,'X':308400,'Y':549914,'Hold':False},
        'wrap':{'calculation':65536,'timer':0,'reachable':True},
        'expiry_reinstall':{'reason':'second equality after uint16 wrap','calculation':101534,'timer':35998,'program':0x1ba17,'entry_cursor':0x1ba19,'live_tasks':[]},
        'QUIT_calculation':102583,'expiry_admitted_visits':1050,
        'clock':'from this supplied handoff: timer=calculation modulo65536 for every examined unpaused logical admission; no wall-clock/PIT claim'}
    r['first_actual_post_SETBALL_drain']={
        'witness':'single_launch','calculation':36288,'classification':'unscored',
        'expired_old':True,'expiry_effect_request':False,'program':'PARTY_ONTS',
        'PARTY_ON_TASK1_fires':36318,'second_SETBALL_fires':36398,
        'branch':'unscored guard precedes expired test; PARTYFLASH=true before NEW_BALL; reset preserves PARTY_ON program and expired',
        'witness_followup':'no further plunger input; new ball returns to chute; next timer equality101534; QUIT102583'}
    req,res=p['drain_boundaries'][1:3]
    r['expired_scored_drain_admission']={
        'witness':'periodic_launch','calculation':37084,'old_timer':req['timer'],'old_expired':req['expired'],
        'SCORECHANGED':req['SCORECHANGED'],'INH_EFF':req['INH_EFF'],'SPECIALMODE':req['SPECIALMODE'],
        'effect_file':0x1a4a1,'cue':{'position':13,'repeat':0,'priority':255},
        'priority_before':req['audio_priority'],'priority_after':res['audio_priority'],'admitted':res['effect_accepted'],
        'program':0x1ba17,'cursor_at_install':0x1ba19,'old_cursor_restored':False,'clear_remaining_at_install':5,
        'live_tasks_at_drain':[x for x in req['live_task_sites'] if x],
        'input_script':{'Down':[36206,36227],'Release':36228,'flippers':'for d=n-36228>=0: Left iff (d+46)%52<8; Right iff (d+22)%30<21; remaining controls false'},
        'QUIT_calculation':38133,'matrix_visits':1050,'wrap':'unreachable on this witness'}
    r['complete_expiry_progression']=[]
    for key,start in [('no_input',101534),('single_launch',101534),('periodic_launch',37084)]:
        trace=r['continuations'][key]
        progression=[]
        ns=r['expiry_program']
        operations=[x for x in trace['matrix_progression'] if x['calculation']>=start]
        for node,event in zip(ns,operations):
            progression.append(dict(event,node=node['site'],decoded_next_cursor=node['site']+2*(1+len(node['args'])),admitted_visit_at_dispatch=event['calculation']-start+1))
        r['complete_expiry_progression'].append(dict(witness=key,installed_calculation=start,operations=progression,QUIT_calculation=trace['final']['calculation'],total_matrix_visits=1050))
    r['input_dependence']=[
        {'control':'Down/plunger','classification':'may alter ball trajectory/drain time only','detail':'can defer drain until timer re-equality by withholding release; no expired clear or timer guard; no indefinite active-game witness proved'},
        {'control':'flippers','classification':'UNKNOWN','detail':'meaningful after SETBALL; real periodic script reaches scored drain. No cycle witness or exclusion for arbitrary unpaused control sequences.'},
        {'control':'Space/tilt','classification':'UNKNOWN','detail':'physical push can alter trajectory; outside chute, tilt can install TILTTS. No infinite active replacement cycle proved; actual scored expiry is held/stopped BallLost, and its Sync skips tilt.'},
        {'control':'P/pause','classification':'may prevent the required drain indefinitely','detail':'actual frontend P at supplied35998 reaches suspended self-loop; no electronics, matrix, tasks or audio advance until another key'},
        {'control':'Esc/quit','classification':'can cause an alternate table exit','detail':'native Esc requires SessionReady; P then Esc then Y returns selector. Close exits app. These are available exits, not forced progress.'}]
    r['eventual_theorem']={
        'verdict':r['eventual'],'witness':'legitimate P at calculation35998 then empty input forever; actual native closed suspended transition',
        'finite_run_is_not_nontermination_proof':True,
        'pause_fairness':'no mandatory resume or unpaused-only fairness was supplied by the task; if such a restriction is imposed, this counterexample is excluded and active-only universal termination remains UNKNOWN',
        'active_only_smallest_dependency':'whether every infinite unpaused input continuation from this state admits an expiry installation with no subsequent reachable replacement before QUIT',
        'scope':'TABLE1 session under the accepted native A frontend/calculation contract; QUIT means linked handler entry disables gameplay interrupts, not completion of DOS teardown'}
    r['expiry_interleaving_remaining']={
        'collision_class':'three concrete active witnesses and pause-cycle outcome determined; arbitrary unpaused replacement/admission continuation theorem is not proved',
        'other_first_equality_class':'PARTY_ON_TASK1 firing at age30 after an unscored drain at first equality; its first-equality reachability remains UNKNOWN and is not inferred from this later36288 drain',
        'not_a_reachability_claim':'no additional independently reachable first-equality class is asserted in this narrow pass'}
    r['native_contract_witness_requirements']=[
        'equality can install expiry before a pending NEW_BALL replacement in the same update; retain that ordering',
        'expired survives queued sounds, SHOWPLAYERSTS termination and SETBALL; new ball becomes active',
        'unscored drain can start PARTY_ON and NEW_BALL despite expired=true',
        'scored expired drain consumes priority/inhibit/special admission; admitted effect starts expiry at entry, never resumes old cursor',
        'no-input chute survives uint16 wrap and next equality; no pending replacement then means1050 admitted visits to QUIT',
        'legitimate pause suspends electronics and matrix work and admits an infinite waiting continuation']
    r['native_contract_complete']=False


def main():
    p=argparse.ArgumentParser();p.add_argument('--harness',type=Path,default=Path('/private/tmp/pf-post-collision-reference'));a=p.parse_args()
    data,canonical,historical=[Path(os.environ[n]) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
    demo,full=pinned(data,canonical,historical);b=demo['TABLE1.PRG'];d=Decoder();h,_=identities(b,historical)
    linked_evidence=linked(b,full['TABLE1.PRG'],d,h);cfg,nodes=configuration(b,h)
    identity=prepare(a.harness,canonical,cfg)
    env=os.environ.copy();env.update(PF_POST_CONFIG=str(a.harness/'post-config.json'),PF_POST_OUTPUT=str(a.harness/'post-output.json'),PF_POST_PAUSE_OUTPUT=str(a.harness/'pause-output.json'),GOCACHE='/private/tmp/pf-dmo0-go-cache')
    run=subprocess.run([str(ROOT/'.tools/go/bin/go'),'test','./internal/partyland','-run','^TestResearchPostCollision(PauseCycle)?$','-count=1','-v'],cwd=a.harness,env=env,capture_output=True,text=True)
    (a.harness/'post-tests.log').write_text(run.stdout+run.stderr);require(run.returncode==0,'post continuation failed: isolated post-tests.log')
    result=json.loads((a.harness/'post-output.json').read_text())
    cycle=json.loads((a.harness/'pause-output.json').read_text())
    input_state=json.loads(INPUT.read_text())['final_post_collision_state']
    completed_state=dict(input_state,audio_priority=result['no_input']['start']['audio_priority'])
    report=dict(linked_evidence=linked_evidence,pause_cycle=cycle,input_artifact_state=input_state,starting_state=completed_state,native_reference_identity=identity,
        expiry_program=nodes,configuration=cfg,continuations=result,tests={},status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',
        verdict='POST_COLLISION_TERMINATION = PROVED',eventual=classify(cycle=cycle),expiry_interleaving='EXPIRY_INTERLEAVING = NOT_PROVED')
    report['native_identity_fields']=[k for k in report['starting_state'] if k in result['no_input']['start'] and k not in ('matrix_remaining','events')]
    report['expiry_native_entry']=next(x['pc']-1 for x in result['periodic_launch']['matrix_progression'] if x['calculation']==37084)
    enrich(report)
    validate(report)
    OUTPUT.write_text(json.dumps(report,separators=(',',':'))+'\n')
    print(json.dumps({k:{'final':v['final']['calculation'],'QUIT':v['final']['QUIT'],'drains':len(v['drain_boundaries'])} for k,v in result.items()}))

if __name__=='__main__':main()
