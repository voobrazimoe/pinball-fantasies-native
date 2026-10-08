#!/usr/bin/env python3
"""Narrow PIT producer certificate; authored clocks are never real DOS witnesses.

Uses prior paired-transition premises. Exports semantic metadata, no payload.
No CPU/bus timing is invented to correlate PIT time with TABLE1 instructions.
"""
import argparse
from dataclasses import dataclass
import json
import os
from pathlib import Path
from audit_10min_demo_paired_scheduler import Decoder, require, sha, unpack, SDR, FILES
from audit_10min_demo_sdr import callback_api
from audit_10min_demo_admission import partial_graph

MISSING = ('Reachable joint PIT-time/TABLE1 execution relation: residual source time '
           'after eligible nested L, at outer 0x4758, and through 0x479a.')
HARDWARE = {
    'PIT': 'https://www.cs.usfca.edu/~cruse/cs210s07/8254.pdf',
    'PIC': 'https://www.pcjs.org/documents/datasheets/intel/INTEL_8259A_PIC.pdf',
}


def u16(n):
    return n & 65535


def indexed_reload(next_deadline, previous_deadline, latched_count):
    return u16(next_deadline - previous_deadline + latched_count - 10)


def temporal_relation(next_entry, clear, returned):
    """Only useful with a certified common clock; None must fail closed."""
    if None in (next_entry, clear, returned):
        return 'UNKNOWN'
    require(clear < returned, 'empty critical interval')
    if next_entry < clear:
        return 'BEFORE_CLEAR'
    if next_entry < returned:
        return 'IN_TAIL'
    return 'NOT_BEFORE_RETURN'


@dataclass
class PIT:
    """Authored tick abstraction, after load latency; not a CPU timing model.

    Count zero means 65536. Pending PIC bit coalesces edges, even when masked.
    Mode 0 makes one terminal edge; mode 3 supplies periodic rising edges.
    Real load/edge phase and CPU acceptance latency remain in the obligation.
    """
    count: int
    periodic: bool = False
    pending: bool = False
    masked: bool = False
    interrupts: bool = True
    in_service: bool = False
    running: bool = True

    def __post_init__(self):
        self.program(self.count)

    def program(self, count):
        self.count = u16(count) or 65536
        self.remaining = self.count
        self.running = True
        # Reprogramming PIT does not acknowledge a pending PIC request.

    def tick(self, ticks=1):
        require(ticks >= 0, 'negative authored source ticks')
        for _ in range(ticks):
            if not self.running:
                continue
            self.remaining -= 1
            if self.remaining == 0:
                self.pending = True
                if self.periodic:
                    self.remaining = self.count
                else:
                    self.running = False

    def accept(self):
        if not (self.pending and self.interrupts and not self.masked
                and not self.in_service):
            return False
        self.pending = False
        self.in_service = True
        self.interrupts = False
        return True

    def early_eoi_sti(self):
        self.in_service = False
        self.interrupts = True


@dataclass
class Direct:
    countdown: int = 1
    phase: str = 'P'
    active: bool = True
    enabled: bool = True

    def raw_entry(self, pit):
        if not pit.accept():
            return 'NO_SOURCE_ENTRY'
        # Source EOI precedes decrement, but STI only follows the guards.
        pit.in_service = False
        self.countdown = u16(self.countdown - 1)
        if self.countdown:
            pit.interrupts = True  # IRET to authored interrupted IF=1
            return 'NONZERO'
        if not self.active:
            pit.interrupts = True
            return 'INACTIVE_ZERO_RETAINED'
        if not self.enabled:
            self.countdown = 50
            pit.interrupts = True
            return 'DISABLED_RELOAD_50'
        pit.interrupts = True
        return self.phase  # publication is a separate, interruptible step

    def publish_phase(self):
        self.phase = 'L' if self.phase == 'P' else 'P'

    def reload(self, n):
        self.countdown = u16(n)


def linear(d, raw, start, end, base=0):
    xs = []
    while start < end:
        x = d.instruction(raw, base, start)
        xs.append(x)
        start += x.size
    require(start == end, 'instruction extent drift')
    return xs


def expect_sequence(d, raw, start, seq):
    sites = []
    for mnemonic, operand in seq:
        d.expect(raw, 0, start, mnemonic, operand)
        sites.append(start)
        start += d.instruction(raw, 0, start).size
    return sites, start


def pit_write(d, raw, start, load, mode=0x30):
    seq = [('push','ax'),('push','bx'),('mov',f'al, {hex(mode)}'),
           ('pushf',''),('cli',''),('out','0x43, al'),('mov',load),
           ('mov','al, bl'),('out','0x40, al'),('mov','al, bh'),
           ('out','0x40, al'),('popf',''),('pop','bx'),('pop','ax')]
    sites, end = expect_sequence(d, raw, start, seq)
    return dict(mode_command=mode, command=sites[5], low=sites[8], high=sites[10],
                irq_atomic_write=True, end=end)


def source_cfg(d, raw):
    api = callback_api('supplied', raw, d)
    _, seen = partial_graph(d, raw, [0, api['int66_entry']])
    return api, seen


def indexed_source(d, raw, q):
    c, advance = q['callback'], q['advance']
    roots = [p for p in range(c-250,c) if raw[p:p+4] == b'\xfc\x60\x06\x1e']
    require(len(roots)==1, 'ambiguous narrow IRQ root')
    root = roots[0]
    d.expect(raw,0,root+4,'mov','al, 0x20')
    d.expect(raw,0,root+6,'out','0x20, al')
    d.expect(raw,0,root+8,'sti')
    prefix = linear(d,raw,root+9,advance)
    guard = next(x for x in prefix if x.mnemonic=='cmp' and x.op_str.startswith('byte ptr cs:'))
    snapshot = next(x for x in prefix if x.mnemonic=='mov' and x.op_str=='si, word ptr ['+hex(q['cursor'])+']')
    first_test = d.instruction(raw,0,snapshot.address+snapshot.size)
    require(first_test.mnemonic=='cmp' and first_test.operands[0].reg==d.x86.X86_REG_SI,
            'first record comparison')
    first_record = first_test.operands[1].imm
    # Last fixed prefix is shared by every indexed producer, not inferred from STI.
    delta = advance-44
    expect_sequence(d,raw,delta,[('mov','bx, word ptr [si + 2]'),
        ('sub','bx, word ptr [si - 7]'),('mov','al, 0'),('out','0x43, al'),
        ('in','al, 0x40'),('mov','cl, al'),('in','al, 0x40'),('mov','ch, al'),
        ('add','bx, cx'),('sub','bx, 0xa')])
    nonfirst=pit_write(d,raw,advance-21,'bx, bx')
    require(nonfirst['end']==advance, 'deadline publication end')
    first_starts=[x.address for x in prefix if x.mnemonic=='push' and x.op_str=='ax'
                  and x.address < delta and d.instruction(raw,0,x.address+x.size).op_str=='bx']
    require(len(first_starts)==1, 'first deadline block')
    first=pit_write(d,raw,first_starts[0],'bx, word ptr [si + 2]')
    mask_window=None
    if q['driver']=='GUS.SDR':
        expect_sequence(d,raw,first_starts[0]-27,[('in','al, 0x21'),('push','ax'),('or','al, 0xfd'),('out','0x21, al')])
        expect_sequence(d,raw,first['end'],[('pop','ax'),('out','0x21, al')])
        mask_window=dict(set=first_starts[0]-22,restore=first['end']+1,meaning='IRQ0 masked during first-record raster wait/rearm; saved IMR restored before cursor/callback')
    api, seen=source_cfg(d,raw)
    registration=api['callback_registration'][0]['entry']
    cal1=d.instruction(raw,0,registration+22)
    raster=d.instruction(raw,0,registration+29)
    cal2=d.instruction(raw,0,registration+49)
    require(all(x.mnemonic=='call' and x.operands[0].type==d.x86.X86_OP_IMM for x in (cal1,raster,cal2)), 'indexed calibration producer drift')
    expect_sequence(d,raw,raster.operands[0].imm,[('pushf',''),('cli',''),('push','es'),('mov','dx, 0x40')])
    raster_end=linear(d,raw,raster.operands[0].imm,cal1.operands[0].imm)
    require(any(x.mnemonic=='inc' and x.op_str=='cx' for x in raster_end), 'raster line counter absent')
    calibration=dict(registration=registration,first_PIT_measure=cal1.address,raster_measure=raster.address,second_PIT_measure=cal2.address)
    vectors=[]
    for p,x in seen.items():
        if x.mnemonic=='mov' and x.op_str==f'ax, {hex(root)}':
            sites,end=expect_sequence(d,raw,p,[('mov',f'ax, {hex(root)}'),
                ('xchg','word ptr es:[0x20], ax'),('mov',None),
                ('mov','ax, cs'),('xchg','word ptr es:[0x22], ax')])
            expect_sequence(d,raw,p-3,[('push','0'),('pop','es')])
            vectors.append(dict(offset_write=sites[1], segment_write=sites[4]))
    require(len(vectors)==1, 'IRQ0 installation not established')
    return dict(irq_root=root, vector_8_install=vectors[0], eoi=root+6,
        first_sti=root+8, source_enabled_cs=guard.operands[0].mem.disp,
        record_snapshot=snapshot.address, first_record=first_record,
        next_deadline_first=first, next_deadline_nonfirst=nonfirst,temporary_mask_window=mask_window,
        deadline_formula='first: u16([SI+2]); other: u16([SI+2]-[SI-7]+latched_PIT0-10)',
        record_units='calibrated cumulative PIT input ticks relative to raster frame; not CPU cycles',
        publication_window='early STI precedes snapshot; after PIT rearm, expiry before cursor/priority publication needs a temporal proof too. No atomic global record advancement is claimed.',
        source_transition='IRQ0 -> EOI -> STI -> enabled guard -> SI snapshot -> PIT0 mode0 rearm -> cursor advance -> priority callback -> priority restore',
        survives_nested=['PIT state','advanced cursor','shared active priority with per-invocation restoration'],
        recurrence_units='PIT input ticks; hardware zero load means 65536, terminal OUT rise supplies IRQ0',
        minimum_new_load='at least one positive count tick, plus hardware load phase; NOT a residual-at-tail bound',
        calibration=dict(producers=calibration,meaning='API11 measures raster lines and PIT ticks; API12 converts raster position to cumulative PIT deadline'),
        record_reachability='No claim all nine slots execute during P. Only current SI and its successor matter; temporal reachability UNKNOWN.')


DIRECT_SETUP={'INTERNAL.SDR':(0x1554,0x1562,0x14ac,0x15e1),
              'ADLIB.SDR':(0x153e,0x154c,0x14aa,0x1667),
              'THING.SDR':(0x1534,0x1542,0x14a0,0x15c1)}


def direct_source(d,raw,q):
    c=q['callback']; install,rearm,setup,bootstrap=DIRECT_SETUP[q['driver']]
    root=d.instruction(raw,0,install).operands[1].imm
    d.expect(raw,0,install,'mov',f'word ptr es:[0x20], {hex(root)}')
    expect_sequence(d,raw,install-3,[('push','0'),('pop','es')])
    d.expect(raw,0,root,'push','ax');d.expect(raw,0,root+1,'push','ds')
    d.expect(raw,0,c-262,'inc')
    # INC is before EOI; count field aliases the sample-fetch operand, not due state.
    d.expect(raw,0,c-241,'dec',f'word ptr cs:[{hex(q["countdown"])}]')
    d.expect(raw,0,c-236,'je',hex(c-218))
    d.expect(raw,0,c-218,'cmp',f'byte ptr cs:[{hex(q["active"])}], 0xff')
    d.expect(raw,0,c-210,'cmp',f'byte ptr cs:[{hex(q["enabled"])}], 0')
    d.expect(raw,0,c-257,'mov','al, 0x20')
    d.expect(raw,0,c-255,'out','0x20, al')
    d.expect(raw,0,c-253,'cmp');d.expect(raw,0,c-246,'jb',hex(c-241))
    d.expect(raw,0,c-244,'jmp',hex(c+75))
    d.expect(raw,0,c+79,'mov');d.expect(raw,0,c+86,'jmp',hex(c-241))
    d.expect(raw,0,c-227,'mov',f'word ptr cs:[{hex(q["countdown"])}], 0x32')
    sites,end=expect_sequence(d,raw,setup,[('mov','al, 0x36'),('out','0x43, al'),
        ('mov','dx, 0x12'),('mov','ax, 0x34dc'),('mov','bx, word ptr [0x78b]'),
        ('div','bx'),('mov','bx, ax'),('mov','word ptr [0x66c5], ax'),
        ('mov','al, bl'),('out','0x40, al'),('mov','al, bh'),('out','0x40, al'),('sti','')])
    reloaded=pit_write(d,raw,rearm,'bx, word ptr [0x66c5]',0x36)
    expect_sequence(d,raw,bootstrap,[('mov','ax, word ptr [0x78b]'),('mov','bx, 0x64'),
        ('xor','dx, dx'),('div','bx')])
    period=d.instruction(raw,0,bootstrap+10)
    require(period.mnemonic=='mov' and period.operands[1].reg==d.x86.X86_REG_AX,'period initializer')
    full=d.instruction(raw,0,c+29).operands[1].mem.disp
    split=d.instruction(raw,0,c+32).operands[1].mem.disp
    require(full==period.operands[0].mem.disp,'whole-period field mismatch')
    api,seen=source_cfg(d,raw)
    # Common DS segment installation for vector 8 is operand-derived in setup CFG.
    cs_installs=[p for p,x in seen.items() if x.mnemonic=='xchg' and x.op_str=='word ptr es:[bx + 2], ax'
                 and d.instruction(raw,0,p-2).op_str=='ax, cs']
    require(cs_installs,'driver CS vector setup absent')
    return dict(irq_root=root, vector_8_offset_write=install, vector_segment_setup=cs_installs,
        eoi=c-255, raw_countdown_decrement=q['countdown_decrement'],countdown_cs=q['countdown'],phase_cs=q['phase'],active_cs=q['active'],enabled_cs=q['enabled'],
        source_pit_program=dict(command=sites[1],low=sites[9],high=sites[11],mode_command=0x36),
        source_pit_rearm=reloaded, sample_rate_ds=0x78b, divisor_ds=0x66c5,
        divisor_formula='floor(0x1234dc/sample_rate) PIT input ticks; hardware zero divisor means 65536',
        countdown_event='one decrement per SERVICED IRQ0, not per elapsed tick or lost/coalesced edge',
        countdown_transition='u16(n-1); zero enters guards, nonzero IRET; inactive retains zero; disabled reloads 50',
        zero_wrap='loaded n>0 expires after n serviced IRQ0 entries; loaded 0 after 65536; next decrement from 0 -> 65535',
        whole_period_ds=full, split_ds=split, whole_period_initializer=bootstrap,
        whole_period_formula='initial floor(sample_rate/100); P raster correction mutates whole period',
        reloads=dict(P='split if L registered, otherwise whole period', L='u16(whole_period-split)'),
        source_transition='periodic PIT0 IRQ0 -> sample cursor step/wrap -> EOI -> countdown decrement -> guards -> STI -> phase test -> raster correction if P -> phase publication -> countdown reload -> callback',
        publication_window=dict(sti=q['interrupt_enable'],phase_test=q['phase_test'],
            P_phase=q['primary_publish'],P_reload=q['primary_reload'],
            L_phase=q['later_publish'],L_reload=q['later_reload'],
            between='before reload, another serviced IRQ decrements current countdown; from just-expired zero it wraps to 65535. No due callback follows from phase store alone.'),
        survives_nested=['countdown','phase','period/split correction','sample cursor; wrap rejoins same decrement'],
        recurrence_units='serviced IRQ0 entries separated by PIT mode3 divisor, including edge phase and acceptance latency',
        calibration='API11 raster-line measurement supplies split ratio; source rate supplies period. Neither measures TABLE1 tail or VGA OUT latency.')


def tail_shape(d,b):
    xs=linear(d,b,0x4753,0x479b,0x300)
    d.expect(b,0x300,0x4753,'mov','byte ptr [0x24a7], 0')
    d.expect(b,0x300,0x479a,'retf','')
    require(all(not x.group(d.cs.CS_GRP_CALL) and not x.group(d.cs.CS_GRP_JUMP)
                and x.mnemonic not in ('cli','sti','popf','iret','hlt') for x in xs),'tail control drift')
    ports=[x.address+0x300 for x in xs if x.mnemonic=='out']
    require(len(ports)==6 and all(x.op_str=='dx, ax' for x in xs if x.mnemonic=='out'),'tail IO drift')
    return dict(clear=0x4753,first_post_clear=0x4758,returned=0x479a,
        instruction_count_including_clear_return=len(xs),post_clear_count_including_return=len(xs)-1,post_clear_count_before_retf=len(xs)-2,
        graphics_out_sites=ports,paths=1,branches=0,calls=0,loops=0,
        mask_change=False,source_reprogram=False,
        narrower_window='any accepted P between completion of 0x4753 and entry to RETF at 0x479a',
        duration_in_PIT_ticks=None,
        reason='fixed instruction count is not a CPU/VGA-bus elapsed-time bound; no source clock read in tail')


def audit(data,saved='/private/tmp/pf-dmo0-paired-scheduler.json'):
    prior=json.loads(Path(saved).read_text());d=Decoder();data=Path(data)
    b=(data/'TABLE1.PRG').read_bytes()
    require((len(b),sha(b))==FILES['TABLE1.PRG'],'not pinned TABLE1')
    tail=tail_shape(d,b);rows=[]
    for q in prior['drivers']:
        name=q['driver'];packed=(data/name).read_bytes();size,digest,extent=SDR[name]
        require((len(packed),sha(packed))==(size,digest),'not pinned SDR')
        raw,_=unpack(packed);require(len(raw)==extent,'SDR extent')
        source=indexed_source(d,raw,q) if 'indexed' in q['family'] else direct_source(d,raw,q)
        rows.append(dict(driver=name,family=q['family'],actual_source='PIT0 OUT -> master PIC IR0 -> INT08 supplied SDR handler',
            source=source,paired_premises=q,
            enable_mask_variables=['PIT running/gate','master PIC IMR bit0','PIC IRR0/ISR0','CPU IF',
                                   'driver source enabled'] + (['driver active'] if 'direct' in q['family'] else []),
            reentry_permission='normal DOS PIC: early EOI releases ISR0; IF=1 permits next unmasked pending IRQ0 before IRET. Permission alone is not expiry.',
            progresses_during_callback=True,
            temporal_relation_evidence=dict(source_counts=True,callback_clock_bound=False,
                residual_at_post_clear=None,critical_tail=tail,verdict='UNKNOWN'),
            critical_tail_verdict='UNKNOWN',
            traces={k:'UNKNOWN real source feasibility' for k in
                    ['P -> nested L -> outer return','P -> nested L -> P while rest busy -> outer return',
                     'P -> nested L -> P after clear -> outer return']},
            impossible_traces=['no pending source token','IRQ0 masked or IF clear',
                'source stopped with no pending request'] +
                (['P delivery while indexed active priority=200'] if 'indexed' in q['family'] else ['callback on nonzero countdown'])) )
    timer=(data/'TIMER.BIN').read_bytes()
    require(len(timer)==253 and sha(timer)=='783f88891a760b3fab648a7ae64ad8998e1349737df07a1e765329f7a6787d0c','TIMER identity')
    for at,m,op in [(0x15,'mov','al, 0xb2'),(0x17,'out','0x43, al'),(0x44,'mov','cx, 0xaf'),
        (0x85,'mov','cx, 0x7d0'),(0xcd,'mov','cx, 0'),(0xd0,'sub','cx, ax'),
        (0xdd,'mov','ax, 5'),(0xe0,'cmp','cx, 0x1a00'),(0xe7,'cmp','cx, 0x2500'),
        (0xee,'cmp','cx, 0x3400'),(0xf5,'cmp','cx, 0x4400'),(0xfc,'ret','')]:d.expect(timer,0,at,m,op)
    return dict(verdict='SOURCE_DUE_RECURRENCE = NOT_PROVED',
        source_due_reentry_before_outer_return='UNKNOWN',paired_scheduler_contract='PAIRED_SCHEDULER_CONTRACT = NOT_PROVED',
        native_audio_boundary='NATIVE_AUDIO_BOUNDARY = NOT_PROVED',smallest_missing_fact=MISSING,
        scope='supplied SDR producer and critical TABLE1 tail; prior conditional P/L suffixes are premises',
        hardware_references=HARDWARE,drivers=rows,critical_tail=tail,
        families={family:dict(drivers=[r['driver'] for r in rows if r['family']==family],
            producer='PIT0 IRQ0',temporal_verdict='UNKNOWN',critical_tail_verdict='UNKNOWN')
            for family in ('indexed constant-budget','indexed audio-budget','direct audio-budget')},
        calibration=dict(TIMER='PIT2 measures 175 VGA-store loop iterations and 2000 multiply/memory loop iterations; returns five-class threshold result, not exact callback timing',
            driver='raster-to-source calibration gives deadline/countdown units, not instruction/IO-to-PIT timing',
            usable_tail_bound=False),
        indexed_critical_trace_obligations=[
            'outer: SI=current P; rearm successor deadline; advance cursor; active 100',
            'nested: real PIT terminal edge before outer return required; successor L 200 permitted; advance/rearm again',
            'L return restores 100; cursor/deadline survive; following record must actually be P and due',
            'P due before clear is rest-busy case; P due after clear is second-E case; neither temporal placement certified',
            'intervening or rejected records also rearm/advance; nine slot capacity is not temporal reachability'],
        direct_critical_trace_obligations=[
            'outer expiry n=1 ->0; active/enabled; STI; phase=P; publish L then split reload',
            'split effective n (zero=65536) serviced IRQ0 entries required while P remains outstanding',
            'eligible L expiry ->0; publish P then u16(period-split); L returns',
            'remaining effective n serviced entries must place next P before outer return; relative to clear UNKNOWN',
            'before phase/reload IRQ0 can wrap zero; outer reload may overwrite intervening decrements; strict alternation not certified'],
        authored_model_scope='tests provide explicit source ticks and CPU stages; accepted model traces are NOT historical reachable witnesses',
        real_feasible_traces=[],real_impossible_traces=['entry with no pending IRQ0 under unchanged source state'],
        whole_dos_gate=prior['whole_dos_gate'])


def main():
    p=argparse.ArgumentParser();p.add_argument('--data',default=os.getenv('PF_10MIN_DEMO_DATA'))
    p.add_argument('--saved',default='/private/tmp/pf-dmo0-paired-scheduler.json');p.add_argument('--output',required=True)
    args=p.parse_args();require(args.data,'private demo required')
    r=audit(args.data,args.saved);Path(args.output).write_text(json.dumps(r,indent=2)+'\n')
    print(r['verdict']);return 2

if __name__=='__main__':raise SystemExit(main())
