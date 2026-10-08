#!/usr/bin/env python3
"""Partial source transition certificate; due-entry edges remain explicitly open.

No executable payload is exported. Authored transitions are conditional tests,
never witnesses that a real scheduler supplies a due event before callback return.
"""
import argparse
from dataclasses import dataclass, asdict
import json
import os
from pathlib import Path
from audit_10min_demo_audio_boundary import Decoder, SDR, FILES, sha, require, unpack, dispatch_shape

UNRESOLVED = ('Due-entry recurrence while a callback is outstanding: the reachable '
    'next-record deadline / direct countdown (including zero and wrap), source enable '
    'state and publication window must supply L then P before outer P return. '
    'STI and a permissive priority/phase branch do not establish that temporal edge.')


@dataclass
class Table:
    LAST_WAS_VB: bool = False
    INSIDE_BALLHANDLER: bool = False
    INSIDE_RESTOFVBLANK: bool = False
    INSIDE_RASTINT: bool = False
    INTERRUPTS_ON: bool = True
    TIME_LEFT: bool = True

    def enter(self, kind, ax=0):
        require(kind in ('P', 'L') and ax in (0, 65535), 'semantic callback input')
        before = asdict(self)
        result = dict(kind=kind, ax=ax, before=before, ball=False,
                      electronics=False, matrix=False, stop=None)
        if not self.INTERRUPTS_ON:
            result['stop'] = 'INTERRUPTS_ON'
        else:
            self.TIME_LEFT = ax == 0
            if kind == 'L':
                if self.INSIDE_RASTINT:
                    result['stop'] = 'INSIDE_RASTINT'
                elif not self.LAST_WAS_VB:
                    result['stop'] = 'LAST_WAS_VB clear'
                elif self.INSIDE_BALLHANDLER:
                    result['stop'] = 'INSIDE_BALLHANDLER'
                else:
                    self.INSIDE_RASTINT = True
                    result['stop'] = 'L body pending'
            elif self.LAST_WAS_VB:
                result['stop'] = 'LAST_WAS_VB set'
            elif self.INSIDE_BALLHANDLER:
                result['stop'] = 'INSIDE_BALLHANDLER'
            else:
                self.INSIDE_BALLHANDLER = True
                result['ball'] = True
                result['stop'] = 'P ball pending'
        result['after'] = asdict(self)
        return result

    def ball_return(self):
        self.INSIDE_BALLHANDLER = False
        self.LAST_WAS_VB = True
        if self.INSIDE_RESTOFVBLANK:
            return dict(electronics=False, stop='INSIDE_RESTOFVBLANK')
        self.INSIDE_RESTOFVBLANK = True
        return dict(electronics=True, stop='P rest pending')

    def later_return(self):
        require(self.INSIDE_RASTINT, 'no admitted L')
        self.LAST_WAS_VB = False
        self.INSIDE_RASTINT = False

    def matrix_test(self):
        return self.TIME_LEFT  # necessary budget condition; matrix busy is separate

    def rest_return(self):
        self.INSIDE_RESTOFVBLANK = False


class Scheduler:
    """Published-state suffix model, conditional on an explicit due-source token.

    Record contents/lifecycle and time until due are not silently nondeterministic
    real-source evidence. A false due token rejects entry; unknown stays UNKNOWN.
    """
    def __init__(self, family, records=('P', 'L')):
        require(family in ('indexed', 'direct'), 'family')
        self.family, self.records = family, records
        self.current = 0
        self.active_priority = 0
        self.phase = 'P'
        self.countdown = None
        self.stack = []

    def enter(self, kind, due=None, reload=None):
        require(kind in ('P', 'L'), 'callback kind')
        if due is None:
            return 'UNKNOWN due-entry recurrence'
        if not due:
            return 'REJECT source not due'
        if self.family == 'indexed':
            if self.records[self.current] != kind:
                return 'REJECT current record'
            self.current = (self.current + 1) % len(self.records)
            priority = {'P': 100, 'L': 200}[kind]
            if priority < self.active_priority:
                return 'DROP priority'  # record advanced even on drop
            self.stack.append(self.active_priority)
            self.active_priority = priority
        else:
            if self.phase != kind:
                return 'REJECT phase'
            require(reload is not None, 'direct reload is semantic state')
            self.phase = 'L' if kind == 'P' else 'P'
            self.countdown = reload & 65535
            self.stack.append(kind)
        return 'DELIVER'

    def leave(self):
        require(self.stack, 'no callback outstanding')
        old = self.stack.pop()
        if self.family == 'indexed':
            self.active_priority = old
        # Direct return restores neither phase nor countdown.


def checked(d, raw, at, mnemonic, operand=None):
    d.expect(raw, 0, at, mnemonic, operand)
    return d.instruction(raw, 0, at)


def indexed_shape(d, raw, c):
    dispatch_shape(d, raw, c, True)
    advance = checked(d, raw, c-54, 'add')
    require(advance.op_str.endswith(', 9'), 'record stride')
    cursor = advance.operands[0].mem.disp
    d.expect(raw, 0, c-49, 'mov')
    d.expect(raw, 0, c-46, 'cmp', 'word ptr [si + 2], ax')
    d.expect(raw, 0, c-43, 'jne', hex(c-35))
    reset = checked(d, raw, c-41, 'mov')
    require(reset.operands[0].mem.disp == cursor, 'wrap cursor')
    d.expect(raw, 0, c-35, 'sti')
    d.expect(raw, 0, c-31, 'mov', 'al, byte ptr [si - 5]')
    gate = checked(d, raw, c-28, 'cmp')
    priority = gate.operands[1].mem.disp
    for at, mnemonic in [(c-21, 'push'), (c-16, 'mov'), (c+4, 'pop')]:
        x = checked(d, raw, at, mnemonic)
        require(x.operands[0].mem.disp == priority, 'active priority save/set/restore')
    return dict(cursor=cursor, priority=priority, advance=c-54,
                wrap=c-41, interrupt_enable=c-35, priority_test=c-28,
                priority_restore=c+4, suffix_complete=True,
                source_due_complete=False)


def direct_shape(d, raw, c):
    dispatch_shape(d, raw, c, False)
    decrement = checked(d, raw, c-241, 'dec')
    countdown = decrement.operands[0].mem.disp
    d.expect(raw, 0, c-236, 'je', hex(c-218))
    active = checked(d, raw, c-218, 'cmp').operands[0].mem.disp
    d.expect(raw, 0, c-212, 'jne', hex(c-234))
    enabled = checked(d, raw, c-210, 'cmp').operands[0].mem.disp
    d.expect(raw, 0, c-204, 'je', hex(c-227))
    d.expect(raw, 0, c-201, 'sti')
    phase = checked(d, raw, c-194, 'cmp').operands[0].mem.disp
    d.expect(raw, 0, c-188, 'jne', hex(c-183))
    d.expect(raw, 0, c-186, 'jmp', hex(c+23))
    d.expect(raw, 0, c-32, 'cmp')  # registered later selector
    d.expect(raw, 0, c-27, 'je', hex(c-16))
    d.expect(raw, 0, c-25, 'mov')  # calibrated split load
    x = checked(d, raw, c-22, 'mov')
    require(x.operands[0].mem.disp == phase and x.operands[1].imm == 255,
            'primary phase publication')
    x = checked(d, raw, c-16, 'mov')
    require(x.operands[0].mem.disp == countdown, 'primary countdown reload')
    x = checked(d, raw, c+23, 'mov')
    require(x.operands[0].mem.disp == phase and x.operands[1].imm == 0,
            'later phase publication')
    d.expect(raw, 0, c+29, 'mov')
    d.expect(raw, 0, c+32, 'sub')
    x = checked(d, raw, c+36, 'mov')
    require(x.operands[0].mem.disp == countdown, 'later countdown reload')
    dispatch_shape(d, raw, c+52, False)
    return dict(countdown=countdown, phase=phase, active=active, enabled=enabled, countdown_decrement=c-241,
                interrupt_enable=c-201, phase_test=c-194,
                primary_publish=c-22, primary_reload=c-16,
                later_publish=c+23, later_reload=c+36,
                source_due_complete=False,
                publication_window='interrupts enabled before phase test/reload; '
                'published-state alternation is not a complete entry model',
                reloads=['P: split when registered; otherwise whole period, phase unchanged',
                         'L: whole period minus split'],
                return_effect='neither countdown nor phase restored')


def conditional_counterexample(family, ax=65535, tail=False):
    s, t = Scheduler(family), Table()
    edges = [s.enter('P', due=True, reload=1)]
    p = t.enter('P'); outer = t.ball_return()
    edges.append(s.enter('L', due=True, reload=1))
    l = t.enter('L', ax); t.later_return(); s.leave()
    if tail:
        t.rest_return()
    edges.append(s.enter('P', due=True, reload=1))
    second = t.enter('P', ax)
    second_rest = t.ball_return() if second['ball'] else None
    return dict(status='CONDITIONAL_ONLY; real-source feasibility UNKNOWN',
                assumptions=['registered P/L pair', 'ordinary enabled TABLE1 mode',
                             'returning local calls', 'two due entries before outer return'],
                scheduler_edges=edges, outer_primary=p, outer_rest=outer,
                nested_later=l, nested_primary=second, nested_primary_rest=second_rest,
                outer_matrix_budget=t.matrix_test(), final_state=asdict(t))


def audit(data, saved='/private/tmp/pf-dmo0-audio-boundary.json'):
    prior = json.loads(Path(saved).read_text())
    d = Decoder(); data = Path(data); b = (data/'TABLE1.PRG').read_bytes()
    require((len(b), sha(b)) == FILES['TABLE1.PRG'], 'not pinned TABLE1')
    # Only newly needed busy lifetime anchors, not a rerun of known body counts.
    for at, m, op in [(0x46bb,'mov','byte ptr [0x24a6], 0'),
                      (0x46c0,'cmp','byte ptr [0x24a7], 0'),
                      (0x46c5,'je','0x4414'),
                      (0x471b,'mov','byte ptr [0x24a7], 0xff'),
                      (0x4753,'mov','byte ptr [0x24a7], 0')]:
        d.expect(b, 0x300, at, m, op)
    rows = []
    for q in prior['drivers']:
        name = q['driver']; packed = (data/name).read_bytes()
        size, digest, extent = SDR[name]
        require(len(packed) == size and sha(packed) == digest, 'not pinned SDR')
        raw, _ = unpack(packed); require(len(raw) == extent, 'extent')
        indexed = q['priority_gate'] is not None
        row = indexed_shape(d, raw, q['consumer']) if indexed else direct_shape(d, raw, q['consumer'])
        row.update(driver=name, family=('indexed constant-budget' if name in
            ('NOSOUND.SDR','GUS.SDR') else 'indexed audio-budget') if indexed else 'direct audio-budget',
            budget=q['budget'], callback=q['consumer'])
        rows.append(row)
    return dict(verdict='PAIRED_SCHEDULER_CONTRACT = NOT_PROVED',
        native_audio_boundary='NATIVE_AUDIO_BOUNDARY = NOT_PROVED',
        unresolved_reason=UNRESOLVED, drivers=rows,
        scheduler_family_states={
            'indexed':['current record / next deadline','active priority','saved priority stack',
                       'source enable/due token','record snapshot held by current invocation'],
            'direct':['16-bit countdown','phase','whole period / calibrated split',
                      'registered later selector','source active/enabled','publication stage','due token']},
        transition_edges={
            'indexed':['due -> snapshot record -> program next deadline -> advance/wrap cursor',
                       'priority below active -> drop, cursor stays advanced',
                       'priority >= active -> save/set priority -> budget -> callback',
                       'return -> restore priority, cursor NOT restored'],
            'direct':['source tick -> decrement modulo 65536; only zero enters paired dispatch',
                      'zero + active/enabled -> STI -> phase selection',
                      'P + registered L -> phase=L; countdown=split; budget -> P',
                      'L -> phase=P; countdown=period-split; budget -> L',
                      'return -> no phase/countdown restore',
                      'STI-to-publication nested edge UNKNOWN; cannot assume atomic selection']},
        nesting_rules={
            'indexed':'published P permits L and equal P; published L drops P, permits equal L; due still required',
            'direct':'published phase chooses next kind; no priority exclusion; due still required',
            'impossible':'no due => no delivery; indexed P at active 200 => drop; wrong published direct phase => reject'},
        budget_production={'constant':'AX=0', 'audio':'AX=0/65535 from audio predicate; exact temporal correlation UNKNOWN'},
        table1_projection=dict(fields=list(asdict(Table())),
            scope='ordinary gameplay; previously proved slowdown zero; attract/task semantics excluded',
            P=['disabled exits before TIME_LEFT store','enabled stores TIME_LEFT before latch/busy guards',
               'latch or ball busy => no ball/no electronics',
               'ball return clears ball busy; rest busy => latch set, no electronics',
               'otherwise latch/rest set; one direct electronics event; shared matrix budget read later',
               'rest cleared at 0x4753 BEFORE callback return 0x479a'],
            L=['disabled exits before TIME_LEFT store','enabled stores TIME_LEFT before raster/latch/ball guards',
               'eligible L sets raster busy; completing L clears latch and raster busy',
               'early L exits retain latch; rest busy is NOT a later guard']),
        authored_counterexamples={f'{f}_{stage}':conditional_counterexample(f, tail=stage=='tail')
            for f in ('indexed','direct') for stage in ('rest','tail')},
        reachability={'real_source_counterexample':'UNKNOWN, not certified reachable or unreachable',
                      'conditional_suffix_counterexample':'accepted in both published-state models',
                      'unreachable_suffixes':['P delivered at indexed active L priority',
                                              'direct wrong phase without another publication', 'entry without due']},
        pause_resume=dict(latch='retained while disabled; paused P/L do not update TIME_LEFT or clear latch',
            first_order='UNKNOWN due/record/countdown state across stop/resume',
            first_P='enabled budget write then latch exit until eligible L',
            first_L='if raster/ball guards clear, releases latch; next admitted P may calculate'),
        family_comparison='No invariant certificate: constant/audio TIME_LEFT differs; '
            'indexed priority exclusion differs from direct phase selection. Native minimal state not certified.',
        whole_dos_gate=prior['whole_dos_gate'])


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--data', default=os.getenv('PF_10MIN_DEMO_DATA'))
    p.add_argument('--output', required=True)
    args = p.parse_args(); require(args.data, 'private data required')
    r = audit(args.data); Path(args.output).write_text(json.dumps(r, indent=2)+'\n')
    print(r['verdict']); return 2

if __name__ == '__main__':
    raise SystemExit(main())
