#!/usr/bin/env python3
"""Narrow DMO0 expiry slice. Conditional traces are NOT reachability witnesses.

The only open obligation selected by this pass is the normal bonus producer's
wait-age/guard provenance at first equality. No global task closure is claimed.
"""
import argparse
from dataclasses import dataclass, field
import json
import os
from pathlib import Path
import struct
from audit_10min_demo_graph import Decoder, pinned, require
from audit_10min_demo_programs import identities

THRESHOLD = 35998
EXPIRY = ('_CLEAR4', '_SCROLL', '_FLASHON', '_PRINT13_NUMBER', '_WAIT',
          '_FLASHOFF', '_SCROLL', '_FADE', '_WAIT', 'QUIT')
UNKNOWN = ('Can NEW_BALL_TASK produced by the real normal demo bonus continuation '
           'survive to the first ElectronicsCalculation 35998 with shared wait '
           'word DS:0x36cd=30 and PARTYFLASH=VISAKEYS=false at its task visit?')
# Consumer anchors bind this slice to linked demo operands, never to a native
# runtime address allowlist. These are regression guards, not CFG reachability.
ANCHORS = (
 (0x782,'mov','dx, 0xbbb'), (0x785,'call','0x5b80'),
 (0xebb,'mov','dx, 0x1e'), (0xebe,'mov','bx, 0x36cd'),
 (0xec1,'call','0x57c7'), (0xec4,'jb','0xbca'),
 (0xeca,'call','0xbce'), (0xee0,'call','0x37bc'),
 (0xf7d,'mov','dx, 0xcce'), (0xf83,'mov','dx, 0xcf4'),
 (0xf89,'mov','dx, 0xca8'),
 (0xff4,'mov','dx, 0x50'), (0xff7,'mov','bx, 0x36d3'),
 (0xffa,'call','0x57c7'), (0x103d,'mov','byte ptr [0x3026], 0'),
 (0x3af8,'call','0x3842'), (0x3afb,'call','0x37af'),
 (0x3ab3,'mov','di, 0x36c9'), (0x3ab6,'mov','cx, 0x32'),
 (0x3afe,'cmp','byte ptr [0xd0], 0xff'),
 (0x3b10,'cmp','byte ptr [0x34f1], 0xff'),
 (0x3b1a,'mov','byte ptr [0x34f1], 0'),
 (0x3b3a,'mov','bx, 0x1ade'), (0x3b3d,'call','0x4501'),
 (0x5ac7,'cmp','word ptr [bx], dx'), (0x5ace,'inc','word ptr [bx]'),
 (0x5ad2,'mov','word ptr [bx], 0'),
 (0x5e80,'mov','bx, 0x3417'), (0x5e9c,'mov','word ptr [bx], dx'),
 (0x5ea2,'mov','bx, 0x3417'), (0x5eab,'call','word ptr [bx]'),
 (0x5d16,'call','0x5b9f'),
 (0x5ba,'mov','dx, 0x1e'), (0x5bd,'mov','bx, 0x36c9'),
 (0x5c9,'mov','byte ptr [0xd0], 0xff'), (0x5ce,'call','0xbce'),
 (0x5d2,'cmp','byte ptr [0x34cf], 0xff'),
 (0x622,'mov','si, 0x6f1'), (0x625,'call','0x5c14'),
)


def wait_visit(age, limit):
    """Shared callsite word, compare BEFORE increment, reset on match."""
    return (0, True) if age == limit else ((age + 1) & 65535, False)


@dataclass
class Slice:
    """Explicitly conditional slice; omits areas, targets, input and physics.

    Matrix completion is an external premise, not a modeled cadence. Ages may
    only support conditional tests. This class never marks a state reachable.
    """
    counter: int = 35997
    expired: bool = False
    hold: bool = False
    party: bool = False
    visa: bool = False
    program: str = 'ordinary'
    cursor: int = 0
    tasks: list = field(default_factory=list)
    waits: dict = field(default_factory=dict)
    quit: bool = False
    events: list = field(default_factory=list)

    def add(self, task):
        # Model actual first-free insertion, including reuse behind scan cursor.
        i = next((i for i, t in enumerate(self.tasks) if t is None), len(self.tasks))
        require(i < 50, 'conditional task list exhausted')
        if i == len(self.tasks): self.tasks.append(task)
        else: self.tasks[i] = task
        self.events.append(('insert', task, i))
        return i

    def install(self, program):
        self.program, self.cursor = program, 0
        self.events.append(('program', program))

    def electronics(self):
        self.counter = (self.counter + 1) & 65535
        if self.counter == THRESHOLD:
            self.expired = self.hold = True
            self.install('expiry')
        # Other suffix producers are deliberately absent; this is not a full
        # ElectronicsCalculation interpreter or a proof of their exclusion.
        i = 0
        while i < len(self.tasks):
            task = self.tasks[i]
            if task:
                limit = {'new_ball':30, 'party_on':30, 'setball':80,
                         'sound_new':50, 'sound_brick':5}[task]
                self.waits[task], fire = wait_visit(self.waits.get(task,0), limit)
                if fire:
                    self.events.append(('fire', task, i))
                    if task in ('new_ball', 'party_on'):
                        if task == 'party_on': self.party = True
                        self.tasks = [None] * len(self.tasks)
                        self.waits.clear()
                        if not self.party:
                            if self.visa: self.visa = False
                            else: self.install('show_player')
                        self.hold = True
                        for t in ('sound_new', 'setball', 'sound_brick'): self.add(t)
                    else:
                        self.tasks[i] = None
                        if task == 'setball': self.hold = False
            i += 1

    def complete_matrix_operation(self):
        """Assume next operation completed, including budget/wait/scroll/fade.

        No tick duration or fairness inference is made by this test operation.
        """
        if self.program == 'expiry' and self.cursor < len(EXPIRY):
            self.events.append(('completed', EXPIRY[self.cursor]))
            self.cursor += 1
            if self.cursor == len(EXPIRY): self.quit = True
        elif self.program == 'show_player':
            self.cursor = min(self.cursor + 1, 3)

    def scored_drain(self, effect_admitted):
        # Admission includes consumed effect flags and cue readiness/priority.
        # Passing True is a hypothesis, not a proof of an actual next drain.
        if self.expired and effect_admitted: self.install('expiry')


def linked(b, d):
    for at, op, operand in ANCHORS: d.expect(b, 0x300, at, op, operand)
    # Check entire bounded add-task body has no WAITLIST zero store.
    at = 0x5e80
    while at < 0x5e9f:
        x = d.instruction(b, 0x300, at)
        if x.mnemonic == 'mov' and x.operands[0].type == d.x86.X86_OP_MEM:
            require(x.op_str == 'word ptr [bx], dx', 'task insertion gained wait writer')
        at += x.size
    require(at == 0x5e9f, 'task insertion extent')
    return len(ANCHORS)


def program(b, handlers, start):
    at, nodes = start, []
    while True:
        h = struct.unpack_from('<H', b, at)[0]
        if not h: return nodes
        require(h in handlers and len(nodes) < 12, 'narrow program extent drift')
        r = handlers[h]
        args = list(struct.unpack_from('<'+'H'*r['arity'], b, at+2))
        nodes.append(dict(site=at, op=r['op'], args=args))
        at += 2 * (1+r['arity'])


def report(data, canonical, historical):
    demo, unused = pinned(data, canonical, historical)
    b, d = demo['TABLE1.PRG'], Decoder()
    anchors = linked(b,d)
    handlers, unused = identities(b,historical)
    expiry = program(b,handlers,0x1ba17)
    show = program(b,handlers,0x1b88e)
    require(tuple(n['op'] for n in expiry) == EXPIRY, 'expiry graph drift')
    require([n['op'] for n in show] == ['_CLEAR4','_PRINT5','_PRINT5'], 'SHOWPLAYERSTS drift')
    require(expiry[-1]['args'] == [0], 'expiry exit status')
    effect = 0x19db0+0x6f1
    require(struct.unpack_from('<H',b,effect+26)[0]+0x19db0 == 0x1ba17,
            'scored expired drain replay program drift')
    classes = []
    for name, task, age, phase, hold, conditions in (
        ('ordinary active ball',None,None,'playing',False,'no pending replacement; no suffix effect'),
        ('normal bonus/new-ball before firing','new_ball',29,'bonus tail / between balls',None,'surviving normal bonus producer'),
        ('normal new-ball equality firing','new_ball',30,'between balls',None,'PARTYFLASH=VISAKEYS=false permits replacement'),
        ('normal new-ball later firing','new_ball',0,'between balls',None,'31 future task visits; expiry must still be active'),
        ('SETBALL equality firing','setball',80,'new ball / plunge preparation',True,'NEW_BALL_PART_TWO producer; surviving wait'),
        ('SETBALL later firing','setball',79,'new ball / plunge preparation',True,'next admitted task scan; expiry not exited'),
        ('unscored drain continuation','party_on',30,'unscored drain',None,'sets PARTYFLASH before NEW_BALL reset'),
        ('scored drain at equality',None,None,'bonus before threshold',None,'early drain sees expired=false; equality supersedes bonus'),
        ('other held-ball continuation',None,None,'table-specific capture/mode',True,'producer not traversed in this pass'),
    ):
        classes.append(dict(name=name,reachability='UNKNOWN_AT_FIRST_EQUALITY',
          pending_task=task,shared_wait_age=age,player='unchanged by normal demo continuation',
          current_matrix='phase program; exact producer/cursor not proved',
          HOLDSTILL=hold,expired_before_increment=False,
          ball_down='phase-dependent UNKNOWN',LOOSING='phase-dependent UNKNOWN',
          I_UTSKJUT='phase-dependent UNKNOWN',PARTYFLASH='producer-dependent UNKNOWN',
          VISAKEYS='producer-dependent UNKNOWN',bonus_phase=phase,conditions=conditions,
          eventual_exit='UNKNOWN; conditional local transitions do not establish session outcome'))
    return dict(verdict='EXPIRY_INTERLEAVING = NOT_PROVED',
      status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',
      production_base='306d11a0c479c7ac5ee6e245f3f72eacbc665abd',
      research_head='37ff8bc38d7af4a09a70319675d6e505b0bd6a5e',
      premises=['CANONICAL_A_TIMING_INHERITANCE = PROVED','NATIVE_AUDIO_BOUNDARY = PROVED'],
      boundary='canonical A native reference; admitted ElectronicsCalculation counts only',
      checked_consumer_anchors=anchors,threshold_state_classes=classes,
      classes_complete=False,smallest_remaining_fact=UNKNOWN,
      task_slot_provenance=[
        dict(task='NEW_BALL_TASK',producer=0x785,program_node=0x1b533,
             wait_word_DS=0x36cd,limit=30,first_scan='next electronics after matrix producer',
             fresh_enqueue_example=35967,first_visit=35968,thirtieth_visit=35997,
             equality_visit=35998,pre_equality_age=30,
             witness='CONDITIONAL arithmetic; not a real reachable execution prefix'),
        dict(task='SETBALL',producer=0xf86,wait_word_DS=0x36d3,limit=80,
             first_scan='same scan iff insertion slot is ahead; otherwise next scan',
             wait_origin='NEW_BALL reset zeros shared wait words, then remaining scan may age new slots'),
        dict(task='PARTY_ON_TASK1',producer=0x5b6,wait_word_DS=0x36c9,limit=30,
             body='PARTYFLASH=true then NEW_BALL',reachable_condition='unscored drain; threshold alignment unknown')],
      shared_wait_rule='compare age with limit; match resets to zero and fires; mismatch increments uint16; insertion does not zero it',
      programs=dict(expiry=expiry,show_player=show),
      replacement_edges=[
        dict(source='NEW_BALL reset',target='SHOWPLAYERSTS',guards='not PARTYFLASH and not VISAKEYS',
             expires_test=False,old_cursor_saved=False,reachability='UNKNOWN at equality'),
        dict(source='scored drain with expired=true',target='expiry start',effect_file=effect,
             guard='effect must be admitted; INH_EFF/SPECIALMODE and effect/cue result matter',
             unconditional_restore=False,reachability='UNKNOWN after replacement')],
      HOLDSTILL_transitions=['equality sets true','NEW_BALL sets true','SETBALL clears without expired test'],
      consumed_release='ball movement may resume on subsequent physics; no direct program cancellation or timer guard; a later drain needs a real physics/score prefix',
      other_producers=dict(considered=['PARTY_ON_TASK1','SOUNDNEWBALL','SOUNDBRICKUPP','expired scored-drain replay'],
        sound_tasks='cue calls only locally; no direct DO_MATRIX/HOLDSTILL store',
        exclusion_proved=False,reason='no other threshold classes asserted; stop at selected normal-bonus age provenance'),
      eventual_exit=dict(installed_at_equality=True,replacement_reachable='UNKNOWN',
        restoration='conditional scored-drain replay starts at entry; no cursor resumption',
        progression='only while expiry remains current and matrix operations receive completion',
        QUIT_unreachable='UNKNOWN for a real session; SHOWPLAYERSTS has no QUIT',
        indefinite_TABLE1='UNKNOWN',counter_continues_on_admission=True,
        wrap='arithmetic only; requires 29538 further calculations after first equality',
        re_equality='arithmetic only; requires 65536 further calculations after first equality',
        wrap_or_re_equality_reachable='UNKNOWN',other_termination='UNKNOWN; no out-of-scope paths traversed'),
      native_contract='UNRESOLVED: preserve local replace/release edges as obligations; neither atomic expiry nor eventual QUIT may be asserted',
      test_boundary='conditional slice regressions and linked operand mutations; NOT producer reachability or eventual-session proof')


def main():
    p=argparse.ArgumentParser()
    for name,env in [('data','PF_10MIN_DEMO_DATA'),('canonical','PF_RUNTIME_DATA'),('historical','PF_DMO0_HISTORICAL_SOURCE')]:
        p.add_argument('--'+name,type=Path,default=os.getenv(env))
    p.add_argument('--output',type=Path,default=Path('/private/tmp/pf-dmo0-expiry-interleaving.json'))
    args=p.parse_args();require(all((args.data,args.canonical,args.historical)),'private inputs required')
    r=report(args.data,args.canonical,args.historical)
    args.output.write_text(json.dumps(r,indent=2)+'\n')
    print(r['verdict']);return 2

if __name__=='__main__': raise SystemExit(main())
