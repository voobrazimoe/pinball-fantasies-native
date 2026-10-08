#!/usr/bin/env python3
"""Narrow native audio boundary evidence, NOT a whole-DOS certificate.

Private modules are identity-checked and decoded in memory. Only semantic
metadata is exported. Local callback CFG counts deliberately summarize calls
as returning opaque effects: they cannot certify transitive/reentrant counts.
"""
import argparse
import json
import os
from pathlib import Path

from audit_10min_demo_graph import Decoder, require, sha
from audit_10min_demo import FILES
from audit_10min_demo_sdr import SDR, unpack, callback_api


def body_counts(dec, binary, root, base=0x300, counted=0x5cd9):
    """All local paths, with matched fallthrough at opaque calls; fail on loops.

    Counts are overapproximations (both conditional successors). Export every
    opaque dependency, rather than implying that callee effects were proved.
    """
    memo, active, calls, returns, stack_reads = {}, set(), {}, set(), set()

    def walk(at):
        if at in memo:
            return memo[at]
        require(at not in active, 'local callback cycle needs a separate proof')
        require(base <= at < 0xaed0, 'callback body outside TABLE1 CS')
        active.add(at)
        x = dec.instruction(binary, base, at)
        for op in x.operands:
            if op.type == dec.x86.X86_OP_MEM and op.access & dec.cs.CS_AC_READ:
                if op.mem.segment == dec.x86.X86_REG_SS:
                    stack_reads.add(at)
        next_at = at + x.size
        if x.group(dec.cs.CS_GRP_RET):
            require(x.mnemonic == 'retf' and not x.op_str, 'callback return changed')
            returns.add(at)
            counts = {0}
        elif x.group(dec.cs.CS_GRP_CALL):
            target = (base + x.operands[0].imm
                      if x.operands[0].type == dec.x86.X86_OP_IMM else None)
            calls[at] = dict(site=at, target=target,
                             effects='OPAQUE returning callee; not an admission certificate')
            counts = {n + int(target == counted) for n in walk(next_at)}
        elif x.group(dec.cs.CS_GRP_JUMP):
            require(x.operands[0].type == dec.x86.X86_OP_IMM,
                    'indirect callback body branch')
            counts = walk(base + x.operands[0].imm)
            if x.mnemonic != 'jmp':
                counts = counts | walk(next_at)
        else:
            counts = walk(next_at)
        active.remove(at)
        memo[at] = counts
        return counts

    counts = walk(root)
    return dict(root=root, local_path_counts=sorted(counts), instructions=len(memo),
                return_sites=sorted(returns), opaque_calls=list(calls.values()),
                explicit_ss_read_sites=sorted(stack_reads),
                scope='own body only; includes syntactically feasible branches; '
                      'no claim of transitive effects or IRQ feasibility')


def dispatch_shape(dec, raw, call, indexed):
    """Verify the budget -> AX -> far callback -> AX sentinel handshake."""
    helper = dec.instruction(raw, 0, call-12)
    require(helper.mnemonic == 'call' and
            helper.operands[0].type == dec.x86.X86_OP_IMM, 'budget helper producer')
    dec.expect(raw, 0, call-9, 'mov', 'ax, 0')
    dec.expect(raw, 0, call-6, 'jae', hex(call-1))
    dec.expect(raw, 0, call-4, 'mov', 'ax, 0xffff')
    dec.expect(raw, 0, call-1, 'push', 'ax')
    dec.expect(raw, 0, call, 'lcall')
    end = call + dec.instruction(raw, 0, call).size
    dec.expect(raw, 0, end, 'pop', 'bx')
    if indexed:
        dec.expect(raw, 0, call-28, 'cmp')
        dec.expect(raw, 0, call-23, 'jb', hex(call+9))
        dec.expect(raw, 0, call-21, 'push')
        dec.expect(raw, 0, call-16, 'mov')
        dec.expect(raw, 0, call+4, 'pop')
        sentinel = call+14
    else:
        sentinel = call+10
    dec.expect(raw, 0, sentinel, 'cmp', 'ax, 0x3039')
    helper_at = helper.operands[0].imm
    first = dec.instruction(raw, 0, helper_at)
    if first.mnemonic == 'clc':
        dec.expect(raw, 0, helper_at+1, 'ret')
        budget = 'always AX=0'
    else:
        dec.expect(raw, 0, helper_at, 'push')
        dec.expect(raw, 0, helper_at+3, 'pop', 'ds')
        dec.expect(raw, 0, helper_at+4, 'cmp')
        dec.expect(raw, 0, helper_at+28, 'cmp', 'ax, bx')
        dec.expect(raw, 0, helper_at+30, 'ja', hex(helper_at+34))
        dec.expect(raw, 0, helper_at+32, 'stc')
        dec.expect(raw, 0, helper_at+33, 'ret')
        dec.expect(raw, 0, helper_at+34, 'clc')
        dec.expect(raw, 0, helper_at+35, 'ret')
        budget = 'AX=0 or 65535 from audio active / remaining-buffer predicate'
    return dict(consumer=call, budget_helper=helper_at, budget=budget,
                priority_gate=call-28 if indexed else None,
                sentinel_consumer=sentinel,
                scope='local handshake only; helper input feasibility / IRQ trace unproved')


def paired_trace(events, last=False, enabled=True):
    """Authored serial TABLE1 admission abstraction, not a DOS emulator.

    Useful counterexamples to an unconditional callback-count mapping, not
    evidence that a particular hardware scheduler admits an authored trace.
    Busy guards and nesting are deliberately outside this miniature model.
    """
    counts = []
    for event in events:
        if event == 'pause':
            enabled = False
        elif event == 'resume':
            enabled = True
        elif event == 'P':
            admitted = enabled and not last
            counts.append(int(admitted))
            if admitted:
                last = True
        elif event == 'L' and enabled:
            last = False
        else:
            require(event == 'L', 'unknown authored event')
    return dict(electronics_counts=counts, total=sum(counts), last=last)


def audit(data, saved):
    data = Path(data)
    b = (data/'TABLE1.PRG').read_bytes()
    size, digest = FILES['TABLE1.PRG']
    require(len(b) == size and sha(b) == digest, 'not pinned TABLE1')
    dec = Decoder()
    # New boundary facts, distinct from previously established primary guards.
    for at, mnemonic, operand in [
        (0x4537, 'or', 'ax, ax'), (0x4543, 'mov', 'byte ptr [0x37f1], 0'),
        (0x472a, 'cmp', 'byte ptr [0x37f1], 0'), (0x472f, 'je', '0x4450'),
        (0x4750, 'call', '0x5968'),
        (0x4709, 'mov', 'byte ptr cs:[0x449c], 0xff'),
        (0x4714, 'mov', 'byte ptr cs:[0x449c], 0xff'),
        (0x5939, 'cmp', 'byte ptr [0x2409], 0'),
        (0x5943, 'cmp', 'byte ptr [0x3025], 0xff'),
        (0x596c, 'cmp', 'byte ptr [0x24a5], 0'),
        (0x5973, 'cmp', 'byte ptr cs:[0x449c], 0xff'),
        (0x597b, 'call', '0x4069'),
        (0x597e, 'cmp', 'byte ptr [0x24a6], 0'),
        (0x5a69, 'mov', 'byte ptr cs:[0x449c], 0'),
        (0x5a6f, 'mov', 'byte ptr [0x24a5], 0'),
        (0x34d8, 'mov', 'byte ptr [0x3025], 0'),
        (0x3530, 'mov', 'byte ptr [0x3025], 0xff'),
        (0x6351, 'mov', 'cx, 0x108'), (0x635e, 'mov', 'cx, 0xae'),
    ]:
        dec.expect(b, 0x300, at, mnemonic, operand)
    primary = body_counts(dec, b, 0x4517)
    later = body_counts(dec, b, 0x592b)
    require(primary['local_path_counts'] == [0, 1] and
            later['local_path_counts'] == [0], 'callback local count changed')
    historical = json.loads(Path(saved).read_text())
    drivers = []
    for name, (size, digest, extent) in SDR.items():
        packed = (data/name).read_bytes()
        require(len(packed) == size and sha(packed) == digest, 'not pinned SDR '+name)
        raw, unused = unpack(packed)
        require(len(raw) == extent, 'SDR extent')
        api = callback_api(name, raw, dec)
        indexed = 'record_base' in api['callback_registration'][1]['producer']
        # Reuse prior producer discovery as candidate metadata only. Validate
        # its current linkage and instructions; never upgrade completeness.
        prior = next(d for d in historical['callback_and_interrupt_admission']['drivers']
                     if d['driver'] == name)
        calls = sorted(set(v['site'] for v in prior['invocation_producers']))
        require(len(calls) == 1, 'unique candidate consumer required')
        c = calls[0]
        if indexed:
            dec.expect(raw, 0, c, 'lcall', '[si - 4]')
        else:
            field = api['callback_registration'][0]['producer']['pointer_ds_offset']
            dec.expect(raw, 0, c, 'lcall', '[%s]' % hex(field))
        row = dispatch_shape(dec, raw, c, indexed)
        row.update(driver=name, family='indexed raster records' if indexed else 'alternating direct pointers',
                   registration=api['callback_registration'], scheduler_complete=False)
        if not indexed:
            late = api['callback_registration'][1]['candidate_absolute_consumers']
            require(len(late) == 1, 'unique later consumer')
            row['later_handshake'] = dispatch_shape(dec, raw, late[0], False)
        drivers.append(row)
    return dict(verdict='NATIVE_AUDIO_BOUNDARY = NOT_PROVED',
        scope='TABLE1 local CFG and supplied-driver semantic handshakes; no DOS closure change',
        primary=primary, later=later, drivers=drivers,
        resolution=dict(low_cx=264, low_cx_hex='0x108', high_cx=174),
        authored_serial_traces={''.join(t): paired_trace(t) for t in
                               [('P','L','P'), ('P','P','L'), ('L','P','P'),
                                ('P','pause','L','resume','P','L','P')]},
        entry_abi=dict(proved_input='AX zero/nonzero -> shared TIME_LEFT',
            memory_before_guards=['increment SYNC_COUNTER', 'set INT_WAS_HERE', 'set INT_WAS_HERE2'],
            output='AX=12345 normal return suppresses driver emergency API8 fallback',
            saved_budget_copy='driver retains pushed crisis word and pops it after callback; '
                              'not proved to be an incoming TABLE1 argument',
            transitive_inputs='NOT CLOSED; local absence of explicit SS reads is not a callee ABI proof'),
        minimal_unresolved_semantic_fact='whether a later callback can complete inside an '
            'unfinished primary rest-of-update, clearing LAST_WAS_VB and overwriting '
            'shared TIME_LEFT, with a further primary delivery before the first returns; '
            'the local priority predicate permits this but source-phase feasibility '
            'and the resulting logical update sequence are not certified',
        next_dependency='one semantic transition certificate for paired callback delivery: '
            'indexed active priority/current record and direct alternating phase, projecting '
            'only LAST_WAS_VB, busy guards and shared TIME_LEFT; no physical stack proof',
        native_candidate=dict(event='ElectronicsCalculation', status='LOCAL necessary condition only',
            order=['ball/drain','UPDATE_COUNTERS','ElectronicsCalculation',
                   'areas/targets/shift','KEYTASK','tasks'],
            no_mapping_to=['Runner ticks','presentation frames','physics substeps','native audio callbacks']),
        obligations={
            'unrestricted SOUND.CFG EXEC':'BELOW-NATIVE-BOUNDARY: native does not EXEC arbitrary programs; scope supplied semantics only',
            'complete INT66 vector admission':'BELOW-NATIVE-BOUNDARY: native uses typed services; their observable effects remain relevant',
            'SDR resident placement':'BELOW-NATIVE-BOUNDARY: no resident DOS address model',
            'SDR handler physical stack and return balance':'BELOW-NATIVE-BOUNDARY: native does not model IRQ frames',
            'callback source stack provenance':'UNRESOLVED-BOUNDARY: local body has no explicit SS argument reads; transitive ABI not proved',
            'CODE2 stack alias from external stack placement':'BELOW-NATIVE-BOUNDARY: stack-address alias route absent in native; CODE2 semantic data still relevant',
            'API11/API12 order priority nesting and budget':'NATIVE-RELEVANT',
        },
        whole_dos_gate=dict(source=str(saved), table1_total=len(historical['domains']),
            table1_bounded=len(historical['domains'])-historical['unknown_domain_count'],
            table1_unknown=historical['unknown_domain_count'],
            code2_total=len(historical['related_code2']['domains']),
            code2_unknown=historical['related_code2']['unknown_domain_count'],
            exit=2, status='DMO0 NOT CLOSED. DMO1 NOT STARTED.'))


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--data', default=os.getenv('PF_10MIN_DEMO_DATA'))
    p.add_argument('--saved', default='/private/tmp/pf-dmo0-admission-domains.json')
    p.add_argument('--output', required=True)
    args = p.parse_args()
    require(args.data, 'private demo input required')
    result = audit(args.data, args.saved)
    Path(args.output).write_text(json.dumps(result, indent=2)+'\n')
    print(result['verdict'])
    return 2


if __name__ == '__main__':
    raise SystemExit(main())
