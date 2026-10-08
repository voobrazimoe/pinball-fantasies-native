#!/usr/bin/env python3
"""Bounded semantic diamonds, never a physical-delivery reachability proof.

Exports metadata only. Returning opaque gameplay calls are equal common effects,
not a new proof of their domains or VGA noninterference.
"""
import argparse
from copy import deepcopy
from dataclasses import asdict
import json
import os
from pathlib import Path
from audit_10min_demo_paired_scheduler import Table
from audit_10min_demo_audio_boundary import Decoder, FILES, sha, require
from audit_10min_demo_source_recurrence import linear


def matrix_step(s):
    # Checked active-animation nonterminal path: 7330 -> 7336 -> 73e9 -> 73ed.
    s['animation_remaining'] = (s['animation_remaining'] - 1) & 65535
    if s['animation_remaining'] == 0:
        s['frame_transition_due'] = True


def pair_a(ax=65535, remaining=2, persistent=True):
    initial = dict(table=asdict(Table(LAST_WAS_VB=True, INSIDE_RESTOFVBLANK=True)),
        electronics=1, ball_updates=1, sync=1, animation_remaining=remaining,
        frame_transition_due=False, program_cursor='active animation; unchanged',
        printtask='DUMRET', task_advancements=1,
        gameplay='common completed outer electronics/physics effects',
        flags={'INT_WAS_HERE':255,'INT_WAS_HERE2':255},
        witness_guards={'matrix_busy':False,'intflag':False,'active_animation':True,
            'PRINTTASK':'DUMRET','L_keyboard':0,'L_keyboard_busy':True,
            'L_3003':0,'L_3007':0,'first_task_slot':'DUMRET'},
        common_L_effects='same scroll/raster/keytask queue effects, no matrix replacement in witness')
    results=[]
    for before in (False, True):
        s=deepcopy(initial);t=Table(**s['table'])
        def later():
            r=t.enter('L',ax);require(r['stop']=='L body pending','ineligible authored L')
            t.later_return()
            s['queued_KEYTASK']=1
        if before: later()
        admitted=t.matrix_test()
        if admitted and persistent: matrix_step(s)
        t.rest_return()
        if not before: later()
        s['table']=asdict(t)
        results.append(dict(id='A1' if before else 'A0',event='same successor L1',
            delivery='before 472a' if before else 'after outer P RETF',
            matrix_admitted=admitted,final=s))
    return dict(initial=initial,traces=results,
        source_event=dict(identity='L1 successor of outer P0',AX=ax,
            final='same cursor/phase; priority restored; equal residual source state is a premise'),
        TABLE1_state_deltas={k:[results[0]['final']['table'][k],results[1]['final']['table'][k]] for k in initial['table']},
        matrix_program_task_deltas={k:[results[0]['final'][k],results[1]['final'][k]] for k in ('animation_remaining','program_cursor','printtask','task_advancements','queued_KEYTASK')},
        first_divergence=('DATA2:042a animation frame countdown at 7330'
                          if results[0]['final']!=results[1]['final'] else None))


def pair_b(ax=0, callbacks=1, tail_native_write=False):
    # Bounded projection: identical complete P transformer and same event list.
    # This algebra does not assume the opaque DOS P transformer is VGA-independent.
    initial=dict(table=asdict(Table()),electronics=1,ball_updates=1,sync=1,
        animation_remaining=8,frame_transition_due=False,task_advancements=1,
        flags={'INT_WAS_HERE':255,'INT_WAS_HERE2':255},
        native_probe=0,source_events=[])
    out=[]
    for nested in (False,True):
        s=deepcopy(initial)
        def tail():
            if tail_native_write:s['native_probe']=s['electronics']
        def p(i):
            t=Table(**s['table']);s['sync']+=1
            r=t.enter('P',ax);require(r['ball'],'P not admitted')
            s['ball_updates']+=1;require(t.ball_return()['electronics'],'rest blocked')
            s['electronics']+=1;s['task_advancements']+=1
            if t.matrix_test():matrix_step(s)
            t.rest_return();s['table']=asdict(t);s['source_events'].append('P'+str(i))
        if not nested:tail()
        for i in range(callbacks):
            # Additional eligible L event before next P, equal on both sides.
            if i:
                t=Table(**s['table']);t.enter('L',0);t.later_return();s['table']=asdict(t)
                s['source_events'].append('L'+str(i))
            p(i)
        if nested:tail()
        out.append(dict(id='B1' if nested else 'B0',event='same next P0',
            delivery='4758 before 479a' if nested else 'after 479a',final=s))
    return dict(initial=initial,traces=out,
        source_event=dict(identity='P0 next event, counted once on each side',AX=ax,
            final='same published source fields and empty callback priority stack by premise'),
        TABLE1_state_deltas={k:[out[0]['final']['table'][k],out[1]['final']['table'][k]] for k in initial['table']},
        matrix_program_task_deltas={k:[out[0]['final'][k],out[1]['final'][k]] for k in ('animation_remaining','task_advancements')},
        projected_equal=out[0]['final']==out[1]['final'],
        scope='bounded common P transformer; opaque call/VGA noninterference not certified')


def tail_certificate(d,b):
    xs=linear(d,b,0x4758,0x479b,0x300);rows=[]
    for x in xs:
        at=x.address+0x300
        require(x.mnemonic in ('pop','push','mov','out','retf'),'tail operation drift')
        require(not any(o.type==d.x86.X86_OP_MEM for o in x.operands),
                'tail touches explicit memory/native state')
        if x.mnemonic=='out':
            require(x.op_str=='dx, ax','tail output ABI drift');category='DOS rendering/VGA-only'
        elif x.mnemonic=='retf':category='return/calling machinery'
        elif at==0x4797:
            d.expect(b,0x300,at,'mov','ax, 0x3039');category='callback ABI state'
        elif x.mnemonic in ('pop','push'):category='return/calling machinery'
        else:category='DOS rendering/VGA-only'
        rows.append(dict(site=at,category=category,
            reads=['invocation-local saved register values'] if x.mnemonic in ('pop','push','out') else [],
            registers_read=[d.decoder.reg_name(r) for r in x.regs_access()[0]],
            registers_written=[d.decoder.reg_name(r) for r in x.regs_access()[1]],
            writes=['VGA register'] if x.mnemonic=='out' else
                   ['return AX=12345'] if at==0x4797 else ['local registers/stack']))
    require(len(xs)==38 and sum(x.mnemonic=='out' for x in xs)==6,'tail shape drift')
    return dict(operations=rows,native_reads=[],native_writes=[],
        VGA_restores=['GC5','GC8','GC4','GC1','GC0','SEQ2'],
        implicit_stack='three saved words; balanced push/pop scratch; then RETF frame',
        ABI='AX=12345; SDR separately retains incoming AX crisis value',
        logical_diamond='PROVED for any VGA-independent complete P transformer',
        full_DOS_diamond='NOT_PROVED: opaque P rendering callees not given VGA noninterference certificate',
        repeated_normalization='finite fixed callback word can move across logical-identity tail; event additions/reordering cannot',
        hardware_caveat='P reads VGA registers at 4591/459b/45a4/45ae/45b7/45c1 for saves. Tail outputs affect those reads. Full P cannot be declared hardware-independent from empty tail RAM footprint alone.')


def tail_reorder(d,b,cut):
    """Execute the checked tail operands, with an ABI-preserving callback macro.

    Tests the tail diamond itself, not noninterference of opaque P callees.
    The macro saves/restores six hardware registers, as the checked P suffix
    does, and applies an arbitrary common native transformation.
    """
    xs=linear(d,b,0x4758,0x479b,0x300)
    require(0 <= cut <= len(xs)-1,'tail split')
    initial=dict(reg={'ax':0,'dx':0},stack=[0x1234,0x5678,0x9abc],
        gfx={(0x3ce,i):17+i for i in (0,1,4,5,8)} | {(0x3c4,2):19},
        logical=1,events=[])
    def run(nested):
        s=deepcopy(initial)
        def read(n):
            if n in ('ah','al'):return (s['reg']['ax'] >> (8 if n=='ah' else 0)) & 255
            return s['reg'][n]
        def write(n,v):
            if n in ('ah','al'):
                shift=8 if n=='ah' else 0;s['reg']['ax']=(s['reg']['ax'] & ~(255<<shift)) | ((v&255)<<shift)
            else:s['reg'][n]=v&65535
        def p():
            saved=deepcopy(s['gfx']);s['logical']+=1;s['events'].append('same P1')
            # Arbitrary callback rendering changes, restored before return.
            for key in s['gfx']:s['gfx'][key]^=255
            s['gfx']=saved
            # SDR wrapper restores interrupted registers, not callback return AX.
        for i,x in enumerate(xs):
            if nested and i==cut:p()
            if x.mnemonic=='retf':break
            dst=x.operands[0]
            name=d.decoder.reg_name(dst.reg)
            if x.mnemonic=='pop':write(name,s['stack'].pop())
            elif x.mnemonic=='push':s['stack'].append(read(name))
            elif x.mnemonic=='mov':
                src=x.operands[1];v=src.imm if src.type==d.x86.X86_OP_IMM else read(d.decoder.reg_name(src.reg));write(name,v)
            elif x.mnemonic=='out':s['gfx'][(read('dx'),read('al'))]=read('ah')
            else:raise ValueError('tail operation drift')
        if not nested:p()
        return s
    return run(False),run(True)

def dependency_certificate(d,b):
    # Only the direct active-animation countdown dependency; no task domain audit.
    anchors=[(0x472a,'cmp','byte ptr [0x37f1], 0'),(0x472f,'je','0x4450'),
        (0x4746,'call','0x449d'),(0x479d,'call','0x44b2'),
        (0x47bb,'nop',''),(0x47bc,'mov','dx, word ptr [0x34e6]'),
        (0x47c0,'mov','si, word ptr [0x34e8]'),(0x47c4,'call','dx'),
        (0x47ca,'mov','word ptr [0x34e8], si'),
        (0x485d,'mov','word ptr [0x34e6], 0x701e'),
        (0x4863,'mov','word ptr [0x34e8], si'),
        (0x7328,'push','0x206c'),(0x732b,'pop','ds'),
        (0x732c,'mov','word ptr [0x42e], si'),
        (0x7330,'dec','word ptr [0x42a]'),(0x7334,'je','0x7039'),
        (0x7336,'jmp','0x70e9'),(0x73e9,'mov','si, word ptr [0x42e]'),
        (0x73ed,'ret',''),(0x7339,'mov','bx, word ptr [0x42c]'),
        (0x7355,'add','word ptr [0x42c], 4'),
        (0x735a,'push','word ptr [bx + si + 2]'),(0x735d,'pop','word ptr [0x42a]'),
        (0x47a7,'call','word ptr [0x37f9]'),
        (0x47ab,'mov','word ptr [0x37f9], 0x6a71'),(0x6d71,'ret','')]
    anchors += [(0x597b,'call','0x4069'),(0x59ca,'call','0x8820'),
        (0x8b20,'cmp','byte ptr cs:[0x883a], 0xff'),(0x8b26,'je','0x8839'),
        (0x8b39,'ret',''),(0x59cd,'cmp','byte ptr [0x3003], 0xff'),
        (0x59d2,'jne','0x56df'),(0x59df,'cmp','byte ptr [0x3007], 0xff'),
        (0x59e4,'jne','0x56ec'),(0x59ec,'mov','al, byte ptr [0x3813]'),
        (0x59f1,'je','0x572a'),(0x5a24,'mov','dx, 0x625e'),
        (0x5a27,'call','0x5b80'),(0x5e80,'mov','bx, 0x3417'),
        (0x5e83,'cmp','word ptr [bx], 0x6a71'),(0x5e87,'je','0x5b98'),
        (0x5e98,'inc','word ptr [0x347b]'),(0x5e9c,'mov','word ptr [bx], dx'),
        (0x5e9e,'ret','')]
    for at,m,op in anchors:d.expect(b,0x300,at,m,op)
    graph,seen=d.graph(b,0x300,[0x4369],0x4491)
    require(not graph['indirect_boundaries'],'L scroll indirect dependency')
    writes=set()
    for at,x in seen.items():
        require(x.mnemonic not in ('call','lcall','in'),'L scroll opaque/input dependency')
        for op in x.operands:
            if op.type==d.x86.X86_OP_REG and op.access & d.cs.CS_AC_WRITE:
                require(op.reg!=d.x86.X86_REG_DS,'L scroll segment change')
            if op.type==d.x86.X86_OP_MEM and op.access & d.cs.CS_AC_WRITE:
                require(not op.mem.base and not op.mem.index,'L scroll indexed write')
                writes.add(op.mem.disp)
    require(writes=={0x3000,0x2ffe},'L scroll writes drift')
    return dict(checked_sites=[a for a,_,_ in anchors],
        first_persistent_difference='DATA2:042a: 2 -> 1 admitted, 2 retained suppressed',
        next_equal_matrix_visit='1 -> 0 enters frame selection at 7339; 2 -> 1 retains frame. Cannot erase phase difference.',
        L_dependency=dict(scroll_writes=sorted(writes),
            keyboard_busy_guard='8b20 -> 8b39, returns without keyboard body',
            presentation_guards='3003=3007=0 skip redraw helpers; keyboard=0 skips DO_MATRIX',
            queue='first free slot at DS:3417 receives KEYTASK 625e; DS:347b increments. No task execution until another P.',
            disjoint='DS19bb scroll/task writes do not touch DATA2:042a; active nonterminal animation does not touch scroll/task queue'),
        program_task='active handler 701e / nonzero SI stay equal in witness; PRINTTASK=DUMRET stays equal; no task-list exploration',
        native_consumer='internal/partyland/timing.go:472 -> presentation/matrix.go:571 -> tablelogic.Animation: frameTime controls frame/loop/program advancement',
        no_VGA_input_on_witness='731e..7336 outputs only; nonterminal branch returns via 73e9. Countdown operand never comes from VGA.',
        admission='active animation producer 485d/4863 checked; concrete global reachability and physical L placement remain UNKNOWN')


def audit(data):
    b=(Path(data)/'TABLE1.PRG').read_bytes()
    require((len(b),sha(b))==FILES['TABLE1.PRG'],'not pinned TABLE1')
    d=Decoder();dep=dependency_certificate(d,b);tail=tail_certificate(d,b)
    require(all(x==y for x,y in (tail_reorder(d,b,k) for k in range(38))), 'decoded tail reorder drift')
    tail['decoded_reorder_checks']=dict(boundaries=38,status='PASS',
        scope='exact tail interpreter with ABI-preserving save/restore callback macro; not full opaque P semantics')
    return dict(verdict='SCHEDULER_TIMING_CONFLUENCE = NOT_PROVED',
        classification='REFERENCE_TIMING_REQUIRED',
        classification_scope='for preserving exact matrix logical progress, conditional on these unresolved placements; NOT a claim both physical traces reachable',
        paired_scheduler_contract='PAIRED_SCHEDULER_CONTRACT = NOT_PROVED',
        native_audio_boundary='NATIVE_AUDIO_BOUNDARY = NOT_PROVED',
        first_minimal_native_divergence=dep['first_persistent_difference'],
        synchronization='after outer P and compared L/P returned, before next independent event',
        common_source_state=['same event identities/order and AX','same registration/enable state',
            'same cursor/phase/priority after returns','same residual PIT/pending/countdown at synchronization as comparison premise; not physically proved'],
        physical_reachability='UNKNOWN; this pass tests semantic dependence only',
        case_A=[pair_a(0),pair_a(65535)],case_B=[pair_b(0),pair_b(65535)],
        dependency=dep,tail=tail,
        normalization='No general L move across 472a. Fixed P word commutes with logical tail under explicit VGA-noninterference premise; no full DOS theorem promoted.',
        recursion='Further event count changes are distinct logical traces; P,L,P cannot be reordered into P,P,L. A countdown divergence already blocks general bisimulation.',
        unchanged_domains='electronics/gameplay/ball/score/tasks common up to decision; future identical matrix input distinguishes countdown. No claim all future gameplay stays equal.',
        whole_dos_status='DMO0 NOT CLOSED. DMO1 NOT STARTED.')


def main():
    p=argparse.ArgumentParser();p.add_argument('--data',default=os.getenv('PF_10MIN_DEMO_DATA'));p.add_argument('--output',required=True)
    args=p.parse_args();require(args.data,'private demo required');r=audit(args.data)
    Path(args.output).write_text(json.dumps(r,indent=2)+'\n');print(r['verdict']);return 2
if __name__=='__main__':raise SystemExit(main())
