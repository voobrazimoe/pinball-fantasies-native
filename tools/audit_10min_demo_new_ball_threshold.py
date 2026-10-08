#!/usr/bin/env python3
"""One predecessor pass. A linked zero-aggregate route is NOT a gameplay witness."""
import argparse
import json
import os
from pathlib import Path
import struct
from audit_10min_demo_graph import Decoder, pinned, require
import audit_10min_demo_expiry_interleaving as expiry
from audit_10min_demo_programs import identities

UNKNOWN = ('Does a real normal-game input prefix reach an admitted scored-drain '
           'LOSTBALL installation at calculation 35877, with BONUSSIFFRORNA, '
           'CYCLONECOUNTERBCD, HAPPY_HOUR_TOTAL and MEGA_LAUGH_TOTAL zero and '
           'XXBALLE=false, on the all-budget-true native schedule?')
# Real linked branch destinations, not a target-domain closure.
ROUTE = (
 (0x1b459, '_CLEAR4', (), 5),
 (0x1b45b, '_PRINT13', (7630,344), 1),
 (0x1b461, '_WAIT', (80,), 80),
 (0x1b465, '_CLEAR4', (), 5),
 (0x1b467, '_JBCDZ', (13461,5871), 0),
 (0x1b49f, '_JBCDZ', (176,5915), 0),
 (0x1b4cb, '_JBCDZ', (244,5947), 0),
 (0x1b4eb, '_JBCDZ', (256,5979), 0),
 (0x1b50b, '_JBCDZ', (13461,6017), 0),
 (0x1b531, '_KOLLA_XXBALL', (), 0),
 (0x1b533, '_DEMOVER_CHANGE_PLAYER', (), 0),
)
ANCHORS = (
 (0x592,'cmp','byte ptr [0x34dd], 0'),
 (0x597,'jne','0x2d2'), (0x5d2,'cmp','byte ptr [0x34cf], 0xff'),
 (0x5e3,'mov','si, 0x6d5'), (0x5e6,'call','0x5c14'),
 (0x5f89,'cmp','byte ptr [0x3494], 0xff'),
 (0x5f93,'cmp','byte ptr [0x34e1], 0xff'),
 (0x5f9d,'mov','bx, ax'), (0x5f9f,'call','0x4501'),
 (0x51ca,'mov','word ptr [0x34ea], 0'),
 (0x51d3,'add','word ptr [0x34ea], 0xa8'),
 (0x51d9,'cmp','word ptr [0x34ea], 0x348'),
 (0x4ea4,'push','1'), (0x50a3,'push','word ptr [bx + 2]'),
 (0x55d6,'dec','si'), (0xd75,'cmp','byte ptr [0xcd], 0'),
 (0xd7c,'jmp','0xb21'), (0x788,'pop','bx'),
 (0x2591,'mov','byte ptr [0xd0], 0'),
 (0x3bae,'mov','byte ptr [0x34f1], 0xff'),
 (0x3b1a,'mov','byte ptr [0x34f1], 0'),
 (0x5e83,'cmp','word ptr [bx], 0x6a71'),
 (0x5e8c,'add','bx, 2'), (0x5e8f,'cmp','bx, 0x347b'),
 (0x5e9f,'mov','cx, 0x32'), (0x5eaf,'add','bx, 2'),
)


def linked(b, d, handlers):
    expiry.linked(b,d)
    for at,op,args in ANCHORS: d.expect(b,0x300,at,op,args)
    for at,op,args,cost in ROUTE:
        h=struct.unpack_from('<H',b,at)[0]
        require(h in handlers and handlers[h]['op']==op, 'predecessor route handler drift')
        arity=handlers[h]['arity']
        values=struct.unpack_from('<'+'H'*arity,b,at+2)
        require(values==args,'predecessor route operand drift')
        if op=='_JBCDZ':
            i=next(i for i,r in enumerate(ROUTE) if r[0]==at)
            require(0x19db0+args[1]==ROUTE[i+1][0], 'zero branch predecessor drift')
    e=0x19db0+0x6d5
    require(struct.unpack_from('<H',b,e+26)[0]+0x19db0==ROUTE[0][0],
            'LOSTBALL effect predecessor drift')
    return len(ANCHORS)+len(expiry.ANCHORS)


def producer_offset(route=ROUTE, budget_holes=0):
    require(budget_holes>=0, 'negative budget holes')
    # DO_MATRIX installs first routine during drain before that update's E.
    # The 91st admitted post-scan matrix visit dispatches the producer.
    return sum(r[3] for r in route)-1+budget_holes


def phase_candidate(drain, holes=0):
    """Arithmetic candidate only; never labels a supplied drain reachable."""
    return dict(producer=drain+producer_offset(budget_holes=holes), reachable='UNKNOWN')


def conditional_suffix(producer=35967, age=0, slots=(None,),
                       party=False, visa=False, reset_at=None, interfere_at=None):
    """Authored local regression; list/guard/writer exclusion is an assumption."""
    s=expiry.Slice(counter=producer, tasks=list(slots), party=party, visa=visa)
    slot=s.add('new_ball');s.waits['new_ball']=age
    firings=[]
    for calc in range(producer+1,35999):
        if calc==reset_at: s.tasks=[None]*len(s.tasks);s.waits.clear()
        if calc==interfere_at:
            s.waits['new_ball'],unused=expiry.wait_visit(s.waits.get('new_ball',0),30)
        before=len(s.events);s.electronics()
        firings += [calc for e in s.events[before:] if e[:2]==('fire','new_ball')]
    return dict(slot=slot,wait=s.waits.get('new_ball'),fires=firings,program=s.program,
                reachable=False)


def report(data, canonical, historical):
    demo,unused=pinned(data,canonical,historical);b=demo['TABLE1.PRG'];d=Decoder()
    handlers,unused=identities(b,historical);anchors=linked(b,d,handlers)
    return dict(verdict='NEW_BALL_THRESHOLD_PROVENANCE = NOT_PROVED',
      expiry_interleaving='EXPIRY_INTERLEAVING = NOT_PROVED',
      status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',
      premises=['CANONICAL_A_TIMING_INHERITANCE = PROVED','NATIVE_AUDIO_BOUNDARY = PROVED'],
      checked_anchors=anchors,real_predecessor_chain=dict(
        edges=['normal gameplay -> actual drain decision (input/physics prefix UNKNOWN)',
         'LOOSE_BALL sets LOOSING; clears SPECIALMODE/HAPPY_HOUR/MEGA_LAUGH',
         'SCORECHANGED!=false and expired!=true -> LOSTBALL effect',
         'jingle acceptance plus INH_EFF=false and SPECIALMODE=false -> DO_MATRIX(BALL_LOSTTS)',
         'linked zero-aggregate branch route -> XXBALLE=false -> node 0x1b533',
         '_DEMOVER_CHANGE_PLAYER saves/restores BX and inserts NEW_BALL_TASK first-free'],
        reachability='NOT ESTABLISHED: linked edges do not establish input prefix',
        normal_bonus_entry=0x1b459,producer_node=0x1b533,
        alternative_routes='nonzero bonus/countdown/high-score/XXBALLE routes are not excluded; not a global impossibility claim'),
      producer_calculation_offset=dict(
        route=[dict(site=at,operation=op,args=args,matrix_visits=cost) for at,op,args,cost in ROUTE],
        route_condition='four aggregates zero; XXBALLE=false; no intervening program replacement',
        animation_visits=0,wait_visits=80,clear_visits=10,print_visits=1,
        admitted_matrix_visits=91,formula='producer_calculation = drain_calculation + 90 + H',
        H='budget-rejected matrix updates from drain update through producer; zero for Sync budget=true',
        first_visit='drain update: producer install precedes electronics; task scan precedes budget matrix work',
        generic_route_formula='D + sum(actual routine visits on traversed branch route) - 1 + H; branch route requires actual gameplay values',
        proof_form='linked conditional relative offset, not real predecessor reachability'),
      reachable_phase_evidence=dict(verdict='UNKNOWN',congruence='NONE PROVED',
        candidate_adjustment='unreleased launch-chute continuation',
        missing='no binary/native-derived real scored-drain input prefix joined to a preserved launch delay was established',
        pause='excluded: no admitted electronics progress',
        authored_phase_candidates_are_reachable=False),
      threshold_equation=dict(producer=35967,equation='D + 90 + H = 35967',
        candidate_D='35877 - H',all_budget_true_D=35877,
        phase_candidate=phase_candidate(35877),first_visit=35968,
        visit_indices='35968..35997: visits 1..30; 35998: visit 31',
        off_by_one=dict(producer_35966='fires 35997 before equality',
                       producer_35968='age 29 before equality visit; age 30 after it; fires 35999'),
        arithmetic='PROVED CONDITIONALLY; not retested as reachability'),
      slot_identity_provenance=dict(rule='s=min{i in 0..49: TASKLIST[i]=DUMRET}; DS=0x3417+2*s',
        actual_slot='UNKNOWN without real predecessor task-list contents',
        suffix_survival='NOT ESTABLISHED; authored model assumes no earlier reset, overwrite or NEW_BALL',
        producer_preserves_BX=True,tail='HU_ uses saved matrix BX, not following task slot',
        reset='WHEN_NEW_BALL_RESET calls RESET_TASK_LIST and RESET_WAITLIST before guard tests'),
      wait_word=dict(DS=0x36cd,initial_value='UNKNOWN for real predecessor; last NEW_BALL reset would zero it',
        direct_use='NEW_BALL_TASK at 0xebb..0xec4 -> WAITSYNCS helper',
        writes=['helper mismatch increments','helper equality zeroes','RESET_WAITLIST zeroes DS:0x36c9..0x372c'],
        insertion_zeroes=False,slot_local=False,
        slot_movement='does not transfer age; instances share the same callsite word',
        other_instance='duplicate NEW_BALL_TASK visits could advance/reset it; absence not proved on real suffix',
        reachable_writers='cannot restrict to a concrete reachable suffix before predecessor is established; no global writer closure attempted'),
      PARTYFLASH=dict(DS=0xd0,clear_writer=0x2591,source='CLOSE1 after LASTAREA=BYGEL12',
        set_writer=0x5c9,source_set='unscored-drain PARTY_ON_TASK1 before NEW_BALL',
        value_at_threshold='UNKNOWN: no real prefix proving CLOSE1 and excluding later set on this suffix'),
      VISAKEYS=dict(DS=0x34f1,set_writer=0x3bae,source='WHEN_NEW_GAME_RESET',
        clear_writer=0x3b1a,source_clear='WHEN_NEW_BALL_RESET if PARTYFLASH is not true',
        value_at_threshold='UNKNOWN: requires actual prior reset/guard path, not default zero'),
      collision=dict(witness=None,exclusion=None,reachability='UNKNOWN',
        conditional_consequence='expiry installed then NEW_BALL_TASK can replace with SHOWPLAYERSTS if age=30 and guards false',
        atomic_expiry_disproved_for_real_session=False,eventual_termination='not investigated'),
      smallest_remaining_dependency=UNKNOWN,
      dependency_boundary='one selected earlier gameplay predecessor fact; slot/guard obligations remain conditional and are not separate claimed closures',
      tests_boundary='authored suffix and pinned operand/path mutations only; passing tests cannot discharge UNKNOWN')


def main():
    p=argparse.ArgumentParser()
    for n,e in [('data','PF_10MIN_DEMO_DATA'),('canonical','PF_RUNTIME_DATA'),('historical','PF_DMO0_HISTORICAL_SOURCE')]:
        p.add_argument('--'+n,type=Path,default=os.getenv(e))
    p.add_argument('--output',type=Path,default=Path('/private/tmp/pf-dmo0-new-ball-threshold-provenance.json'))
    args=p.parse_args();require(all((args.data,args.canonical,args.historical)),'private inputs required')
    r=report(args.data,args.canonical,args.historical)
    args.output.write_text(json.dumps(r,indent=2)+'\n');print(r['verdict']);return 2

if __name__=='__main__':raise SystemExit(main())
