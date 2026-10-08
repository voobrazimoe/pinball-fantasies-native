#!/usr/bin/env python3
"""Narrow scored-drain predecessor audit; local transfer facts are not reachability."""
import argparse
import json
import os
from pathlib import Path
import struct
import audit_10min_demo_new_ball_threshold as threshold
from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_programs import identities

AGGREGATES = ('BONUSSIFFRORNA', 'CYCLONECOUNTERBCD', 'HAPPY_HOUR_TOTAL', 'MEGA_LAUGH_TOTAL')
PREDICATE = dict(aggregate_relation='==', aggregate_value=0, aggregates=AGGREGATES, XXBALLE=False)
MISSING = ('A geometry-derived fresh-session launch/outlane/drain transfer relation, '
           'including its admitted calculation indices, that reaches an unlit BYGEL1 '
           'or BYGEL2 without an aggregate producer or INH_EFF setter before scored drain.')
ANCHORS = (
 (0x331,'mov','byte ptr [0xcd], 0'),
 (0x354,'xor','ax, ax'), (0x360,'mov','di, 0x3495'),
 (0x366,'rep stosw','word ptr es:[di], ax'),
 (0x36e,'mov','di, 0xb0'), (0x374,'rep stosw','word ptr es:[di], ax'),
 (0x3ac,'xor','ax, ax'), (0x3b3,'mov','di, 0xf4'),
 (0x3b9,'rep stosw','word ptr es:[di], ax'),
 (0x3bb,'mov','di, 0x100'), (0x3c1,'rep stosw','word ptr es:[di], ax'),
 (0x572,'mov','byte ptr [0x34e1], 0'),
 (0x582,'cmp','byte ptr [0xd2], 0xff'),
 (0x587,'jne','0x28c'),
 (0x58c,'mov','byte ptr [0x34ca], 0xff'),
 (0x5dd,'mov','byte ptr [0x3485], 0'),
 (0x5d36,'cmp','byte ptr [0x2498], 0xff'),
 (0x5d40,'cmp','byte ptr [0x34ca], 0xff'),
 (0x5d4a,'call','0x215'),
 (0x27f6,'cmp','byte ptr [0x36b7], 0xff'),
 (0x2815,'mov','si, 0x1b4'), (0x2818,'mov','di, 0x46b5'),
 (0x281b,'call','0x6ad3'), (0x281e,'mov','byte ptr [0x34dd], 0xff'),
 (0x3adc,'mov','byte ptr [0x3494], 0'),
 (0x5f7b,'popf',''), (0x5f7c,'jb','0x5ca5'),
 (0x5fc5,'cmp','al, byte ptr [0x3485]'), (0x5fc9,'jb','0x5d08'),
 (0x6006,'clc',''), (0x6009,'stc',''),
 (0x6166,'cmp','byte ptr [0x364e], 0'),
 (0x6170,'cmp','byte ptr [0x24a1], 0x20'),
 (0x6175,'jae','0x5e81'), (0x617a,'inc','byte ptr [0x24a1]'),
 (0x61df,'mov','cl, byte ptr [0x24a1]'),
 (0x61ec,'cmp','byte ptr [0x3486], 0'),
 (0x6206,'mov','ax, 0xff5a'), (0x6209,'imul','cx'),
 (0x620d,'mov','ax, word ptr [0x34ec]'), (0x6210,'and','ax, 0xff'),
 (0x6213,'sub','bp, ax'), (0x6223,'mov','word ptr [0x2fea], bp'),
 (0x6233,'and','word ptr [0x2fda], 0xf'),
)


def selected(aggregates, xxball):
    require(len(aggregates)==4, 'four aggregates required')
    return all(x==0 for x in aggregates) and not xxball


def drain_branch(score_changed, expired, puke_forbidden=False):
    if puke_forbidden: return 'inhibited'
    if not score_changed: return 'PARTY_ON'
    return 'MINUTE5' if expired else 'LOSTBALL'


def unlit_lane_transfer(score, aggregates, xxball):
    """Conditional local callback transfer, with no claim of physical entry."""
    return dict(score=(score+50030)%10**12, score_changed=True,
                aggregates=tuple(aggregates), XXBALLE=xxball, reachable='UNKNOWN')


def lostball_admission(inh_eff=False, special=False, priority=0, reset_priority=True):
    require(0<=priority<=255, 'byte priority required')
    current=0 if reset_priority else priority
    return 255>=current and not inh_eff and not special


def charge_step(charge, down=True, cap=32):
    return charge+1 if down and charge<cap else charge


def release_signature(charge, jitter):
    return (-166*charge-(jitter&255), jitter&15)


def charge_only_phase_claim(delays, charge=32, cap=32):
    """Fail closed: charge recurrence is insufficient even for release state."""
    stable=all(charge_step(charge,cap=cap)==charge for unused in delays)
    signatures={release_signature(charge,n) for n in delays}
    return dict(charge_stable=stable, same_release=len(signatures)==1,
                gameplay_lemma=False, reachable_indices='UNKNOWN')


def linked(b,d,handlers):
    threshold.linked(b,d,handlers)
    for at,op,args in ANCHORS: d.expect(b,0x300,at,op,args)
    ds=0x19db0; effect=ds+0x6d5
    require(struct.unpack_from('<H',b,effect)[0]==0xca0,'LOSTBALL jingle pointer drift')
    require(tuple(b[ds+0xca0:ds+0xca0+3])==(6,1,255),'LOSTBALL jingle admission drift')
    require(not any(b[effect+2:effect+26]),'LOSTBALL arithmetic drift')
    require(tuple(b[ds+0x1b4:ds+0x1b4+12])==(0,0,0,0,0,0,0,5,0,0,3,0),'lane score drift')
    require(struct.unpack_from('<5H',b,0x1abcb)==(5,455,15,465,0x24f6),'lane region binding drift')
    at=0x2801;calls=[];destinations=[]
    while at<0x2844:
        x=d.instruction(b,0x300,at)
        if x.mnemonic=='call':calls.append(x.op_str)
        if x.mnemonic=='mov' and x.op_str.startswith('di,'):destinations.append(x.op_str)
        at+=x.size
    require(calls==['0x6ad3','0x4501'] and destinations==['di, 0x46b5'],
            'unlit lane score-only transfer drift')
    return len(ANCHORS)


def report(data,canonical,historical,prior=Path('/private/tmp/pf-dmo0-new-ball-threshold-provenance.json')):
    old=json.loads(prior.read_text())
    require(old['producer_calculation_offset']['admitted_matrix_visits']==91,'prior visit drift')
    require(old['producer_calculation_offset']['route_condition']=='four aggregates zero; XXBALLE=false; no intervening program replacement','prior predicate drift')
    require(old['smallest_remaining_dependency']==threshold.UNKNOWN,'prior selected fact drift')
    demo,unused=pinned(data,canonical,historical);b=demo['TABLE1.PRG'];d=Decoder()
    handlers,unused=identities(b,historical);count=linked(b,d,handlers)
    return dict(verdict='SCORED_DRAIN_35877_PROVENANCE = NOT_PROVED',
      status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',
      corrected_predicate=PREDICATE,
      consistency=dict(classification='REPORT TYPO ONLY',artifact_predicate='zero',extractor_predicate='zero',
        previous_tests='91 visits pinned, but explicit predicate regression absent; added this pass',
        visits=91,producer_offset=90,artifact_rewritten=False),
      real_predecessor_graph=[
        dict(edge='fresh TABLE1 -> new-game/new-ball reset',proof='source initialization edges and linked reset consumers; geometry continuation not closed'),
        dict(edge='reset -> launch chute -> release',proof='input charge/release consumers linked; preservation of whole gameplay state under delay UNKNOWN'),
        dict(edge='release -> unlit BYGEL1/BYGEL2',proof='geometry/input trajectory UNKNOWN; region and callback consumers linked'),
        dict(edge='unlit BYGEL1 -> SCORECHANGED=true, four aggregates unchanged',proof='linked local transfer; no claim callback was reached'),
        dict(edge='ball motion -> BALL_DOWN -> DO_PHYSICS -> LOOSE_BALL',proof='BALLCODE sc_move SC_Y>=banh producer; linked BALL_DOWN/LOOSING tests'),
        dict(edge='LOOSE_BALL -> scored, unexpired LOSTBALL',proof='linked; requires PUKEFORBIDDEN false and SCORECHANGED true'),
        dict(edge='LOSTBALL -> normal bonus 0x1b459',proof='linked admission conditional on INH_EFF=false; SPECIALMODE cleared and priority reset')],
      phase_adjustment=dict(proved_property='SPRINGPOS saturates at 32 while Down remains held',
        admitted_electronics='inherited ordinary Sync schedule; pause excluded',
        limitation='charge stability does not freeze ball motion, electronics counters, modes or task state',
        release='VY=-166*charge-(SLUMP_COUNTERN & 255); rotation=SLUMP_COUNTERN & 15',
        jitter_counter_period=256,calculation_period='UNKNOWN',period_suffices_for_gameplay=False,
        authored_consecutive_jitter_probe=charge_only_phase_claim([100,101,102]),
        lemma='NOT PROVED: no preserved active-ball recurrence joined to an actual scored drain'),
      reachable_calculation_indices=dict(property='UNKNOWN; no interval, congruence, bound or exclusion established',
        checks={str(n):'UNKNOWN' for n in (35876,35877,35878)},
        conditional_producers={str(n):threshold.phase_candidate(n)['producer'] for n in (35876,35877,35878)},
        authored_indices_are_witnesses=False),
      aggregate_provenance=dict(initialization='RESET_VARS2 zeros BONUS/CYCLONE; RESET_VARS zeros HAPPY/MEGA',
        score_only_candidate='unlit side lane: score += 50030; SCORECHANGED=true; no bonus arithmetic',
        producers=dict(BONUSSIFFRORNA=['effect bonus field','bonus arithmetic/restore; not traversed by local unlit-lane callback'],
          CYCLONECOUNTERBCD=['BYGEL13 normal/5x additions','player restore'],
          HAPPY_HOUR_TOTAL=['ADDHAPPY when HAPPY_HOUR=true'],
          MEGA_LAUGH_TOTAL=['ADDMEGALAUGH when MEGA_LAUGH=true']),
        lostball_bonus=0,score_implies_nonzero_aggregate=False,
        same_real_prefix='UNKNOWN: no geometrically derived producer-avoiding trajectory'),
      XXBALLE_provenance=dict(initializer='new-game table reset 0x331 writes false',
        relevant_true_writer='LET_HIM_MATCH; first-ball candidate must not enter match continuation',
        local_lane_and_drain_write=False,same_real_prefix='UNKNOWN; initial false is not whole-prefix writer exclusion'),
      effect_admission=dict(jingle=[6,1,255],matrix=0x1b459,
        priority='scored drain writes current priority=0 immediately before request; 255 admits against every byte priority',
        special='LOOSE_BALL writes SPECIALMODE=false before scored branch',
        suppression='INH_EFF=true still suppresses matrix; new-ball reset clears it but no whole-prefix setter exclusion',
        competitors='no interposed effect inside the inherited synchronous drain call; suffix replacements outside scope',
        same_real_prefix='NOT ESTABLISHED'),
      exact_35877='UNKNOWN; neither membership nor nonmembership proved',
      smallest_remaining_predecessor_fact=MISSING,
      tests_boundary='local transfer/operand mutations and rejection of insufficient phase claims; not a gameplay witness',
      unavailable=['actual zero-aggregate gameplay prefix','proved gameplay phase lemma and mutation of that lemma','reachable index neighbor tests'],
      downstream='task slot, DS:0x36cd uniqueness, PARTYFLASH/VISAKEYS not investigated')


def main():
    p=argparse.ArgumentParser()
    for n,e in [('data','PF_10MIN_DEMO_DATA'),('canonical','PF_RUNTIME_DATA'),('historical','PF_DMO0_HISTORICAL_SOURCE')]:
        p.add_argument('--'+n,type=Path,default=os.getenv(e))
    p.add_argument('--output',type=Path,default=Path('/private/tmp/pf-dmo0-scored-drain-35877.json'))
    a=p.parse_args();require(all((a.data,a.canonical,a.historical)),'private inputs required')
    r=report(a.data,a.canonical,a.historical);a.output.write_text(json.dumps(r,indent=2)+'\n')
    print(r['verdict']);return 2


if __name__=='__main__': raise SystemExit(main())
