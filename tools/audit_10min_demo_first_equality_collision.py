#!/usr/bin/env python3
"""Concrete replay handoff -> linked demo-only suffix. No eventual exit claim."""
import argparse
import copy
import json
import os
from pathlib import Path
import subprocess
import struct
import audit_10min_demo_deterministic_replay as replay
import audit_10min_demo_new_ball_threshold as threshold
from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_programs import identities

ROOT = replay.ROOT
INPUT = Path('/private/tmp/pf-dmo0-deterministic-bygel-drain.json')
OUTPUT = Path('/private/tmp/pf-dmo0-first-equality-collision.json')
# File offsets paired to callsite words by the linked operands below.
TASKS = {'SOUNDRINNER': (0x5fc,0x36cb,5),
         'ENABLETOUCHER': (0x169b,0x36db,20),
         'NEW_BALL_TASK': (0xebb,0x36cd,30),
         'SOUNDNEWBALL': (0xfce,0x36d1,50),
         'SOUNDBRICKUPP': (0xfa8,0x36cf,5),
         'SETBALL': (0xff4,0x36d3,80)}
ANCHORS = (
 (0x2596,'mov','byte ptr [0xd1], 0'), # MUSICOK, NOT VISAKEYS.
 (0x5fc,'mov','dx, 5'),(0x5ff,'mov','bx, 0x36cb'),
 (0x169b,'mov','dx, 0x14'),(0x169e,'mov','bx, 0x36db'),
 (0x16aa,'mov','byte ptr [0x94], 0'),
 (0x3bae,'mov','byte ptr [0x34f1], 0xff'),
 (0x3ab9,'rep stosw','word ptr es:[di], ax'),
 (0x5ea7,'mov','word ptr [0x347d], bx'),
 (0x5eb2,'loop','0x5ba5'),
 (0x5ead,'pop','bx'),(0x5cfd,'call','0x4501'),
 (0x5cdb,'inc','word ptr [0x34cd]'),
 (0x5cdf,'cmp','word ptr [0x34cd], 0x8c9e'),
 (0x5cea,'mov','byte ptr [0x34cf], 0xff'),
 (0x5cef,'mov','byte ptr [0x3026], 0xff'),
 (0x5cf4,'mov','si, 0xc9d'),(0x5cfa,'mov','bx, 0x1c67'),
)


def linked(b,d,h):
    threshold.linked(b,d,h)
    for at,op,args in ANCHORS:d.expect(b,768,at,op,args)
    for name,(at,word,limit) in TASKS.items():
        d.expect(b,768,at,'mov',f'dx, {hex(limit) if limit>9 else limit}')
        d.expect(b,768,at+3,'mov',f'bx, {hex(word)}')
        d.expect(b,768,at+6,'call','0x57c7')
    # Source-backed fresh reset: VISAKEYS=true, then initial NEW_BALL clears it
    # on PARTYFLASH=false. The extra CLOSE1 store is the adjacent MUSICOK byte.
    require(b[0x19db0+0xd0]==0,'fresh PARTYFLASH data initializer drift')
    require(struct.unpack_from('<8H',b,0x1b88e)==(20158,19559,9081,336,19559,9093,1684,0),'SHOWPLAYERSTS mutation')
    require(h[struct.unpack_from('<H',b,0x1b535)[0]]['op']=='_CLEAR4','bonus tail clear mutation')
    require(h[struct.unpack_from('<H',b,0x1b537)[0]]['op']=='_WAIT' and struct.unpack_from('<H',b,0x1b539)[0]==32000,'bonus tail wait mutation')
    return len(ANCHORS)


def reconstruct(harness,canonical,go):
    identity=replay.prepare(harness,canonical)
    replay.replace_once(harness/'internal/partyland/game.go',' ResearchObserve func(string)',
        ' ResearchObserve func(string)\n ResearchTaskNames map[uint64]string\n ResearchResetCount int')
    replay.replace_once(harness/'internal/partyland/game.go','func (g *Game) resetBall() {',
        'func (g *Game) resetBall() {\n g.ResearchResetCount++')
    replay.replace_once(harness/'internal/partyland/timing.go','func (g *Game) waitAt(site string, n uint16, f func()) {\n\tg.task(func() bool {\n\t\tif !g.waitReady(site, n) {\n\t\t\treturn false\n\t\t}\n\t\tf()\n\t\treturn true\n\t})',
        'func (g *Game) waitAt(site string, n uint16, f func()) {\n\tg.task(func() bool {\n\t\tif !g.waitReady(site, n) {\n\t\t\treturn false\n\t\t}\n\t\tf()\n\t\treturn true\n\t})\n if g.ResearchTaskNames==nil {g.ResearchTaskNames=make(map[uint64]string)}\n g.ResearchTaskNames[g.nextTaskID]=site')
    p=harness/'internal/partyland/dmo0_deterministic_replay_test.go'
    replay.replace_once(p,'return map[string]any{"ball":',
        'return map[string]any{"live_task_sites":liveTaskSites(g),"reset_count":g.ResearchResetCount,"ball":')
    replay.replace_once(p,'};return out}', '};collisionReplayGame=g;return out}')
    replay.replace_once(p,'if !reflect.DeepEqual(a,b) {t.Fatal("nondeterministic replay")}', 'nativeTasks:=collisionNativeTasks(collisionReplayGame)\n if !reflect.DeepEqual(a,b) {t.Fatal("nondeterministic replay")}')
    replay.replace_once(p,'out:=map[string]any{"witness":a,', 'out:=map[string]any{"native_task_reference":nativeTasks,"witness":a,')
    (harness/'internal/partyland/dmo0_collision_handoff_test.go').write_bytes(
        (ROOT/'tools/dmo0_collision_handoff_test.go.txt').read_bytes())
    script=harness/'collision-script.json';script.write_text(json.dumps(replay.TARGET))
    out=harness/'collision-handoff.json'
    env=os.environ.copy();env.update(PF_REPLAY_SCRIPT=str(script),PF_REPLAY_OUTPUT=str(out),GOCACHE='/private/tmp/pf-dmo0-go-cache')
    run=subprocess.run([str(go),'test','./internal/partyland','-run','^TestResearchDeterministicReplay$','-count=1','-v'],cwd=harness,env=env,capture_output=True,text=True)
    (harness/'collision-tests.log').write_text(run.stdout+run.stderr)
    require(run.returncode==0,'handoff replay failed: see isolated collision-tests.log')
    return json.loads(out.read_text()),identity


def handoff(saved,fresh):
    replay.validate_replay(saved);replay.validate_replay(fresh)
    old=saved['witness'];new=fresh['witness']
    # Reconstruct missing live function slots only. Verify all original selected
    # fields at every recorded boundary and every compressed trace run.
    def strip(v):
        v=copy.deepcopy(v);v.pop('live_task_sites',None);v.pop('reset_count',None);return v
    require([strip(c) for c in new['Checkpoints']]==old['Checkpoints'],'replay checkpoint handoff drift')
    require([dict(r,state=strip(r['state'])) for r in new['Rows']]==old['Rows'],'replay rows handoff drift')
    cp={c['boundary']:c for c in new['Checkpoints']}
    require(all(r['state']['reset_count']==1 for r in new['Rows']),'unexpected fresh-prefix NEW_BALL/reset')
    require({t for r in new['Rows'] for t in r['state']['live_task_sites'] if t} <= {'ENABLETOUCHER','SOUNDRINNER'},'unexpected prefix task effect')
    require(all('NEW_BALL_TASK' not in r['state']['live_task_sites'] for r in new['Rows']),'earlier NEW_BALL_TASK')
    c=cp['LOSTBALL_result'];after=cp['after drain calculation']
    require(c['live_task_sites']==['']*50,'drain installation live slots drift')
    require(after['live_task_sites']==['SOUNDRINNER']+['']*49,'drain post-scan live slots drift')
    require(c['waits']=={'ENABLETOUCHER':0},'drain shared waits drift')
    require(after['waits']=={'ENABLETOUCHER':0,'SOUNDRINNER':1},'drain task handoff drift')
    close=cp['CLOSE1'];require(close['last_area']=='BYGEL12','CLOSE1 guard')
    # Fresh reference begins after paired initial reset, with waits zero. There
    # are no later resets and no callsite use of 0x36cd on the concrete prefix.
    # Initial demo VISAKEYS=true is consumed/cleared by initial NEW_BALL, before
    # input calculation 1. CLOSE1 is not its writer.
    return dict(installation=c,after_drain=after,
        TASKLIST=[TASKS[t][0]-768 if t else 0x6a71 for t in c['live_task_sites']],
        live_sites=c['live_task_sites'],WAITLIST={hex(w):c['waits'].get(n,0) for n,(_,w,_) in TASKS.items()},
        full_WAITLIST_DS={hex(0x36c9+2*i):0 for i in range(50)},
        DS_0x36cd=0,PARTYFLASH=c['partyflash'],VISAKEYS=False,
        VISAKEYS_provenance='fresh initial WHEN_NEW_GAME_RESET sets DS:34f1=ff; PARTYFLASH initial 0; initial NEW_BALL/reset clears DS:34f1; concrete prefix reset_count=1 and no later fresh-game entry; CLOSE1 DS:d1 is MUSICOK',
        CLOSE1_calculation=close['calculation'],expired=False,HOLDSTILL=c['ball']['Hold'],
        BALL_DOWN=c['BALL_DOWN'],LOOSING=c['LOOSING'],I_UTSKJUT=c['in_chute'],
        demo_program=0x1b459,demo_cursor=0x1b45b,
        native_cursor=c['matrix_pc'],cursor_mapping='linked first _CLEAR4 installed; next linked node 0x1b45b, native index 212',
        handoff_validation='all original replay boundaries and compressed rows identical; missing live slots reconstructed by passive reads')


def concrete_suffix(start,producer=35967,mutation=None):
    """Execute only the admitted zero-aggregate concrete continuation.

    Route/cadence is the previously proved 91-visit premise. Existing live task
    effects are linked sound-only SOUNDRINNER and ENABLETOUCHER suicide. No
    global indirect-domain closure or DOS interrupt model is inferred here.
    """
    require(start['installation']['totals']==[0]*4 and not start['installation']['XXBALLE'],'route handoff')
    slots=start['live_sites'].copy();waits={n:start['installation']['waits'].get(n,0) for n in TASKS}
    party=start['PARTYFLASH'];visa=start['VISAKEYS'];program=0x1b459;cursor=0x1b45b
    expired=False;hold=True;lost=True;loosing=True;chute=False
    rows=[];events=[];checkpoint=None;route=threshold.ROUTE;node=0;remaining=route[0][3]
    producer_shift=producer-35967
    route=list(route);wait=route[2];route[2]=wait[:3]+(wait[3]+producer_shift,)
    tail_remaining=None;matrix_op='_CLEAR4'
    for calc in range(35877,max(35998,producer+31)+1):
        row=dict(calculation=calc,timer_before=calc-1,PARTYFLASH=party,VISAKEYS=visa,
                 tasklist_before=slots.copy(),wait_before=waits['NEW_BALL_TASK'],events=[])
        def event(kind,**kw):
            e=dict(calculation=calc,kind=kind,**kw);events.append(e);row['events'].append(e)
        if calc==35877:
            s=slots.index('');slots[s]='SOUNDRINNER'
            event('drain_tail_insert',slot=s,task='SOUNDRINNER',wait=waits['SOUNDRINNER'])
        # Demo inherited schedule: drain/counters then equality inside electronics,
        # then ordinary KEYTASK/DO_TASKS and admitted matrix work.
        if calc==35998:
            expired=hold=True;program=0x1ba17;cursor=0x1ba19
            event('expiry_install',timer=calc,expired=True,HOLDSTILL=True,jingle='S_GAMEOVER2',program=program,cursor=cursor)
        if mutation and calc==35980:
            if mutation=='slot':slots[0]=''
            if mutation=='wait':waits['NEW_BALL_TASK']+=1
            if mutation=='party':party=True
            if mutation=='visa':visa=True
        event('KEYTASK',operation='DUMRET')
        visits=[]
        for i in range(50):
            t=slots[i]
            if not t:continue
            require(t in TASKS,'unexpected reachable task')
            age=waits[t];waits[t],fire=threshold.expiry.wait_visit(age,TASKS[t][2])
            visits.append(dict(slot=i,task=t,before=age,after=waits[t],fire=fire))
            if not fire:continue
            event('task_fire',slot=i,task=t,wait_after=0)
            if t=='NEW_BALL_TASK':
                event('NEW_BALL',program_before=program,cursor_before=cursor)
                slots=['']*50;waits={n:0 for n in TASKS};lost=loosing=False;chute=True
                event('reset_tasks_waits',slots=slots.copy(),waits=waits.copy())
                event('guards',PARTYFLASH=party,VISAKEYS=visa)
                if not party:
                    if visa:visa=False
                    else:
                        old=program;program=0x1b88e;cursor=0x1b890
                        event('SHOWPLAYERSTS',program_before=old,program_after=program,cursor=cursor,saved_expiry_cursor=None)
                hold=True
                for n in ('SOUNDNEWBALL','SETBALL','SOUNDBRICKUPP'):
                    j=slots.index('');slots[j]=n;event('insert',slot=j,task=n)
                # Current-slot reset followed by add: do not suicide the new slot.
            else:
                slots[i]='';event('suicide',slot=i,task=t)
        row['visits']=visits
        # The concrete drain installs CLEAR before its first admitted visit.
        # Retain accepted route cost; instant branches tail-dispatch in same visit.
        if program==0x1b459:
            if tail_remaining is not None:
                tail_remaining-=1
                event('matrix_visit',op=matrix_op,budget=True,remaining=tail_remaining)
                if tail_remaining==0:
                    require(matrix_op=='_CLEAR4','unexpected tail completion')
                    matrix_op='_WAIT';tail_remaining=32000;cursor=0x1b53b
            elif node<len(route):
                event('matrix_visit',node=route[node][0],op=route[node][1],budget=True)
                remaining-=1
                if remaining==0:
                    node+=1
                    while node<len(route) and route[node][3]==0:
                        if route[node][1]=='_DEMOVER_CHANGE_PLAYER':
                            require('NEW_BALL_TASK' not in slots,'duplicate before producer')
                            before=slots.copy();s=slots.index('');slots[s]='NEW_BALL_TASK'
                            event('producer',node=0x1b533,slot=s,wait=waits['NEW_BALL_TASK'],after_task_scan=True)
                            checkpoint=dict(calculation=calc,boundary=f'after producer on calc {calc}',
                                TASKLIST_before=before,TASKLIST_after=slots.copy(),slot=s,DS_slot=0x3417+2*s,
                                DS_0x36cd=waits['NEW_BALL_TASK'],duplicates=slots.count('NEW_BALL_TASK'),
                                PARTYFLASH=party,VISAKEYS=visa,program=program,cursor=0x1b537,
                                TASKLIST_words_before=[TASKS[t][0]-768 if t else 0x6a71 for t in before],
                                TASKLIST_words_after=[TASKS[t][0]-768 if t else 0x6a71 for t in slots])
                            node=len(route);cursor=0x1b537;tail_remaining=5;matrix_op='_CLEAR4'
                            event('matrix_tail_install',node=0x1b535,op=matrix_op,next_cursor=cursor,remaining=5)
                            break
                        event('matrix_tail',node=route[node][0],op=route[node][1]);node+=1
                    if node<len(route):
                        remaining=route[node][3];cursor=route[node+1][0];matrix_op=route[node][1]
        if program==0x1b88e:
            matrix_op='_CLEAR4';tail_remaining=4
            event('matrix_visit',op=matrix_op,budget=True,remaining=4)
        row.update(matrix_op=matrix_op,matrix_remaining=tail_remaining if tail_remaining is not None else remaining,wait_after=waits['NEW_BALL_TASK'],tasklist_after=slots.copy(),
            PARTYFLASH_after=party,VISAKEYS_after=visa,program=program,cursor=cursor,
            expired=expired,HOLDSTILL=hold,BALL_DOWN=lost,LOOSING=loosing,I_UTSKJUT=chute,
            WAITLIST=waits.copy())
        row['TASKLIST_words']=[TASKS[t][0]-768 if t else 0x6a71 for t in slots]
        row['WAITLIST_DS']={hex(0x36c9+2*i):0 for i in range(50)}
        row['WAITLIST_DS'].update({hex(w):waits[n] for n,(_,w,_) in TASKS.items()})
        rows.append(row)
    collision=any(e['kind']=='SHOWPLAYERSTS' and e['calculation']==35998 and e['program_before']==0x1ba17 for e in events)
    return dict(producer=checkpoint,rows=rows,events=events,collision=collision,
        post_collision=rows[35998-35877],old_expiry_cursor_retained=False,
        matrix_visits_through_producer=sum(e['kind']=='matrix_visit' and e['calculation']<=checkpoint['calculation'] for e in events),route_premise='accepted zero-aggregate 91 admitted visits, H=0; producer after scan',
        reachable_effects=['SOUNDRINNER sound then suicide at 35882','ENABLETOUCHER already completed before LOSTBALL installation; zero shared wait remains',
          'KEYTASK DUMRET; no F1/new-game input','held lost ball skips new geometry/target/drain callbacks',
          'matrix CLEAR/PRINT/WAIT/JBCDZ/KOLLA_XXBALL/DEMOVER producer; no reset/replacement before producer',
          'sole NEW_BALL_TASK wait visits until equality; timer equality is only intervening program installer'])


def validate_suffix(s):
    p=s['producer'];require(s['matrix_visits_through_producer']==91,'admitted bonus matrix visits')
    require(p is not None and p['calculation']==35967,'producer calculation')
    require(p['slot']==0 and p['TASKLIST_before']==['']*50,'first-free actual slot')
    require(p['DS_0x36cd']==0 and p['duplicates']==1,'initial shared wait / duplicate')
    for n in range(35968,35998):
        r=s['rows'][n-35877];v=[v for v in r['visits'] if v['task']=='NEW_BALL_TASK']
        require(r['tasklist_before'][0]=='NEW_BALL_TASK' and len(v)==1,'slot survival')
        require((v[0]['before'],v[0]['after'],v[0]['fire'])==(n-35968,n-35967,False),'shared wait interference')
        require(not r['PARTYFLASH_after'] and not r['VISAKEYS_after'],'guard writer interference')
    r=s['rows'][35998-35877];k=[e['kind'] for e in r['events']]
    require(k.index('expiry_install')<k.index('task_fire')<k.index('NEW_BALL')<k.index('SHOWPLAYERSTS'),'expiry/task ordering or guard')
    v=next(v for v in r['visits'] if v['task']=='NEW_BALL_TASK')
    require((v['before'],v['after'],v['fire'])==(30,0,True),'first equality firing')
    require(r['program']==0x1b88e and not s['old_expiry_cursor_retained'],'replacement/cursor')
    return True


def main():
    p=argparse.ArgumentParser();p.add_argument('--harness',type=Path,default=Path('/private/tmp/pf-first-equality-harness'))
    p.add_argument('--output',type=Path,default=OUTPUT);p.add_argument('--reconstructed-handoff',type=Path);a=p.parse_args()
    paths=[Path(os.environ[n]) for n in ('PF_10MIN_DEMO_DATA','PF_RUNTIME_DATA','PF_DMO0_HISTORICAL_SOURCE')]
    demo,full=pinned(*paths);d=Decoder();
    from audit_10min_demo_canonical_a_jitter import correspondence
    reset_correspondence=correspondence(d,full['TABLE1.PRG'],demo['TABLE1.PRG'])
    h,_=identities(demo['TABLE1.PRG'],paths[2]);linked(demo['TABLE1.PRG'],d,h)
    if a.reconstructed_handoff:
        fresh=json.loads(a.reconstructed_handoff.read_text());identity={'construction':'previous passive fresh rerun; every saved row and checkpoint revalidated'}
    else:
        fresh,identity=reconstruct(a.harness,paths[1],ROOT/'.tools/go/bin/go')
    saved=json.loads(INPUT.read_text());state=handoff(saved['deterministic_replay'],fresh)
    suffix=concrete_suffix(state);validate_suffix(suffix)
    native=fresh.get('native_task_reference')
    require(native is not None,'actual native task/reset probe required')
    for r in native[1:]:
        projected=suffix['rows'][r['calculation']-35877]
        require(r['live_task_sites']==projected['tasklist_after'],'native task-slot continuation mismatch')
        require(r['waits'].get('NEW_BALL_TASK',0)==projected['wait_after'],'native wait continuation mismatch')
    require(native[-1]['reset_count']==2 and native[-1]['phase']==2,'native NEW_BALL reset boundary')
    report=dict(verdict='NEW_BALL_THRESHOLD_PROVENANCE = PROVED',
        collision_verdict='FIRST_EQUALITY_COLLISION_REACHABLE',
        theorem='FIRST_EQUALITY_EXPIRY_REPLACEMENT = REACHABLE',
        atomic_expiry='ATOMIC_EXPIRY_MODEL = DISPROVED',expiry_interleaving='EXPIRY_INTERLEAVING = NOT_PROVED',
        status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',concrete_drain_handoff=state,
        reconstructed_replay_identity=identity,reset_correspondence=reset_correspondence,suffix=suffix,
        native_task_reference=native,final_post_collision_state=native[-1],
        DS_0x36cd_timeline=[dict(calculation=r['calculation'],before_scan=r['wait_before'],after_calculation=r['wait_after']) for r in suffix['rows'] if r['calculation']>=35967],
        PARTYFLASH_timeline=[dict(calculation=r['calculation'],value=r['PARTYFLASH_after']) for r in suffix['rows']],
        VISAKEYS_timeline=[dict(calculation=r['calculation'],value=r['VISAKEYS_after']) for r in suffix['rows']],
        duplicate_instance_check={'before_producer':0,'after_producer':1,'scans_35968_35998':1},
        native_reference_boundary='actual replay Game continued with existing native task/reset primitives; linked demo matrix supplies producer; expiry/guard projection and current demo cursor are in suffix.post_collision',
        guard_correction='CLOSE1 DS:d1 clears MUSICOK, not DS:34f1 VISAKEYS; VISAKEYS=false comes from initial new-ball reset',
        proof_scope='accepted fresh native semantic schedule plus linked demo suffix; no physical DOS timing or whole writer closure',
        eventual_termination='NOT INVESTIGATED',tests={})
    # Merge exact native reset/ball fields with the linked demo matrix state.
    report['final_post_collision_state']=dict(native[-1],demo_program=suffix['post_collision']['program'],
        demo_cursor=suffix['post_collision']['cursor'],matrix_remaining=suffix['post_collision']['matrix_remaining'],
        expired=True,HOLDSTILL=True,VISAKEYS=False,PARTYFLASH=False,
        TASKLIST=suffix['post_collision']['TASKLIST_words'],WAITLIST=suffix['post_collision']['WAITLIST_DS'],
        old_expiry_cursor=None,timer=35998)
    a.output.write_text(json.dumps(report,indent=2)+'\n');print(report['verdict']);print(report['collision_verdict'])

if __name__=='__main__':main()
