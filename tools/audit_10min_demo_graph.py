#!/usr/bin/env python3
"""Pinned, static DMO0 continuation/presentation evidence; never executes DOS.

Requires research-only Capstone. Output contains graph metadata and semantic
records, no instruction listings, executable slices, pictures or sound. Direct
CFG walks deliberately stop at indirect transfers: their presence is reported,
never silently treated as an absence of reachability. This tool cannot close
DMO0 by itself and does not produce a production descriptor.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import struct

from audit_10min_demo import FILES, FINGERPRINT, ROOT, require

HISTORICAL = {
    'FANTASIE.ASM': '52500bf0a0f6f7dbf011cb4b456b04e0800cfb6387bb851e3c1dcbc51f4573b8',
    'PLAND.ASM': '4c162a657f167f723c5c0deed3762bb3163a44673dadae1695531cc5d61e5422',
    'INTRO.ASM': '615e7bd32c3cfcd808ca8a01a3887c0d1b2be141b22cc69cbc7237bae8151770',
    'FANTASIE.MAC': '8e161a6a3bea03822d4d184d7390cef4e45f6d7b7c7ba84be4624a3a8c521327',
}


def sha(b):
    return hashlib.sha256(b).hexdigest()


def pinned(data, canonical, historical):
    demo = {}
    for name, (size, digest) in FILES.items():
        b = (data/name).read_bytes()
        require(len(b) == size and sha(b) == digest, 'not the pinned demo: '+name)
        demo[name] = b
    identity = ''.join(n+'\0'+sha(demo[n])+'\n' for n in FILES)
    require(sha(identity.encode()) == FINGERPRINT, 'demo fingerprint')
    inventory = json.loads((ROOT/'analysis/game-inventory.json').read_text())
    full = {}
    for name in ('INTRO.PRG', 'TABLE1.PRG'):
        b = (canonical/name).read_bytes()
        want = next(r for r in inventory if r['name'] == name)
        require(sha(b) == want['sha256'], 'not canonical A: '+name)
        full[name] = b
    for name, digest in HISTORICAL.items():
        require(sha((historical/name).read_bytes()) == digest, 'historical source drift: '+name)
    return demo, full


def generated(path):
    return json.loads(re.search(r'`(.*)`', (ROOT/path).read_text(), re.S)[1])


def bonus_program(b, ds, demo):
    # Source-derived native declarations give the command arities and numeric
    # arguments. Linked handler identities are inferred consistently over the
    # entire bounded program, then pinned at its actual consumer entry below.
    # No nearest-offset or short-record data equivalence scan is performed.
    commands = generated('internal/partyland/timing_data.go')['commands'][211:270]
    require(commands[0]['op'] == '_CLEAR4' and commands[51]['op'] == '_KOLLA_XXBALL',
            'canonical command declaration drift')
    if demo:
        commands = commands[:52] + [{'op': '_DEMOVER_CHANGE_PLAYER', 'args': []}] + commands[56:]
    positions, known, count = [], [], 0
    for q in commands:
        positions.append(count)
        count += 1
        for i, unused in enumerate(q['args']):
            if str(i) in q.get('nums', {}):
                known.append((count, q['nums'][str(i)] & 65535))
            count += 1
    candidates = []
    for at in range(ds, ds+65536-2*count):
        if all(struct.unpack_from('<H', b, at+i*2)[0] == n for i, n in known):
            candidates.append(at)
    require(len(candidates) == 1, 'nonunique bounded bonus program')
    start = candidates[0]
    require(start == (0x1b459 if demo else 0x1b3cd), 'bonus program consumer drift')
    handlers, rows = {}, []
    for pos, q in zip(positions, commands):
        at = start+pos*2
        h = struct.unpack_from('<H', b, at)[0]
        require(handlers.get(q['op'], h) == h, 'inconsistent linked handler: '+q['op'])
        handlers[q['op']] = h
        rows.append(dict(source=at, node=q['op'], operand_roles=q['args'],
                         handler_file=(h+0x300 if h else None)))
    require(handlers['0'] == 0, 'program terminator')
    if demo:
        require(rows[52]['source'] == 0x1b533 and handlers['_DEMOVER_CHANGE_PLAYER'] == 0x43e,
                'demo continuation consumer')
    return dict(start=start, size=count*2, nodes=rows), handlers


class Decoder:
    def __init__(self):
        # Import only after identity verification; no production dependency.
        import capstone as cs
        from capstone import x86
        self.cs, self.x86 = cs, x86
        self.decoder = cs.Cs(cs.CS_ARCH_X86, cs.CS_MODE_16)
        self.decoder.detail = True

    def instruction(self, b, base, file_at):
        x = next(self.decoder.disasm(b[file_at:file_at+15], file_at-base), None)
        require(x is not None, 'undecodable research instruction')
        return x

    def expect(self, b, base, at, mnemonic, operand=None):
        x = self.instruction(b, base, at)
        require(x.mnemonic == mnemonic and (operand is None or x.op_str == operand),
                'reviewed consumer drift at '+hex(at))

    def graph(self, b, base, roots, limit):
        todo, seen, edges, indirect = list(roots), {}, [], []
        while todo:
            at = todo.pop()
            if at in seen:
                continue
            require(base <= at < limit, 'direct edge outside reviewed CS extent')
            x = self.instruction(b, base, at)
            seen[at] = x
            require(len(seen) < 40000, 'unbounded CFG')
            if x.group(self.cs.CS_GRP_RET) or x.mnemonic == 'iret':
                continue
            transfer = x.group(self.cs.CS_GRP_CALL) or x.group(self.cs.CS_GRP_JUMP)
            if transfer:
                if x.operands[0].type == self.x86.X86_OP_IMM:
                    target = base+x.operands[0].imm
                    edges.append(dict(source=at, target=target,
                                      kind='call' if x.group(self.cs.CS_GRP_CALL) else 'branch'))
                    todo.append(target)
                else:
                    # Address and transfer category only, not disassembly.
                    indirect.append(dict(source=at,
                        kind='call' if x.group(self.cs.CS_GRP_CALL) else 'branch',
                        domain='far' if x.mnemonic in ('lcall', 'ljmp') else 'near'))
                if x.mnemonic in ('jmp', 'ljmp'):
                    continue
            todo.append(at+x.size)
        return dict(root_files=roots, instruction_count=len(seen),
                    direct_edges=sorted(edges, key=lambda r: (r['source'], r['target'])),
                    indirect_boundaries=sorted(indirect, key=lambda r: r['source'])), seen


def inspect(demo, full):
    dec = Decoder()
    table, intro = demo['TABLE1.PRG'], demo['INTRO.PRG']
    program, handlers = bonus_program(table, 0x19db0, True)
    canonical_program, canonical_handlers = bonus_program(full['TABLE1.PRG'], 0x19d40, False)
    for at, op, operand in [
        (0x73f, 'cmp', 'byte ptr [0x5b3], 0xff'),
        (0x756, 'inc', 'byte ptr [0x238a]'),
        (0x75a, 'cmp', 'byte ptr [0x238a], 0x41'),
        (0x764, 'mov', 'byte ptr [0x238a], 0x37'),
        (0x769, 'cmp', 'byte ptr [0x2389], 0x38'),
        (0x773, 'inc', 'byte ptr [0x2389]'),
        (0x77a, 'mov', 'byte ptr [0x2389], 0x38'),
        (0x77f, 'call', '0xe78'),
        (0x782, 'mov', 'dx, 0xbbb'),
        (0x785, 'call', '0x5b80'),
        (0xd75, 'cmp', 'byte ptr [0xcd], 0'),
        (0xd84, 'cmp', 'byte ptr [0x36c3], 0xff'),
        (0xd8f, 'mov', 'al, byte ptr [0x3819]'),
        (0xd92, 'cmp', 'al, byte ptr [0x3815]'),
        (0xe3a, 'dec', 'byte ptr [0xce]'),
        (0xebb, 'mov', 'dx, 0x1e'),
        (0xeca, 'call', '0xbce'),
        (0xee0, 'call', '0x37bc'),
        (0xee3, 'call', '0x9d'),
        (0xef5, 'mov', 'byte ptr [0x3026], 0xff'),
        (0xf72, 'call', '0xc7d'),
        (0xf7d, 'mov', 'dx, 0xcce'),
        (0xf83, 'mov', 'dx, 0xcf4'),
        (0xf89, 'mov', 'dx, 0xca8'),
        (0x5d2, 'cmp', 'byte ptr [0x34cf], 0xff'),
        (0x5d8, 'je', '0x322'),
        (0x622, 'mov', 'si, 0x6f1'),
        (0x64cf, 'mov', 'al, byte ptr [0x3813]'),
        (0x64d9, 'mov', 'byte ptr [0x3815], al'),
        (0x6504, 'call', '0x3853'),
        (0x650d, 'call', '0xbce'),
        (0x5ad6, 'stc', None),
        (0x5e9f, 'mov', 'cx, 0x32'),
        (0x5eab, 'call', 'word ptr [bx]'),
        (0x3afe, 'cmp', 'byte ptr [0xd0], 0xff'),
        (0x3b03, 'je', '0x3841'),
        (0x3b3a, 'mov', 'bx, 0x1ade'),
        (0x3b3d, 'call', '0x4501'),
        (0x4801, 'call', '0x4d85'),
        (0x4804, 'mov', 'byte ptr [0x34f8], 0'),
        (0x4809, 'mov', 'word ptr [0x37f5], 0'),
        (0x482c, 'jmp', 'word ptr [bx]'),
        (0x103d, 'mov', 'byte ptr [0x3026], 0'),
        (0x5cdb, 'inc', 'word ptr [0x34cd]'),
        (0x5cdf, 'cmp', 'word ptr [0x34cd], 0x8c9e'),
        (0x5cfd, 'call', '0x4501'),
        (0x5d16, 'call', '0x5b9f'),
        (0x5d4a, 'call', '0x215'),
        (0x5d4d, 'ret', None),
        (0x4517, 'push', '0x19bb'),
        (0x456a, 'and', 'byte ptr [0x2409], 0'),
        (0x4571, 'cmp', 'byte ptr cs:[0x449c], 0xff'),
        (0x46c0, 'cmp', 'byte ptr [0x24a7], 0'),
        (0x46b8, 'call', '0x5a1f'),
        (0x4720, 'call', '0x2ab5'),
        (0x4723, 'call', '0x59d9'),
        (0x472a, 'cmp', 'byte ptr [0x37f1], 0'),
        (0x4753, 'mov', 'byte ptr [0x24a7], 0'),
    ]:
        dec.expect(table, 0x300, at, op, operand)
    base = 0x36e90
    for at, op, operand in [
        (0x382c7, 'mov', 'cx, 7'),
        (0x382d6, 'mov', 'cx, 0x1b8'),
        (0x382ff, 'mov', 'bx, word ptr [bx + 0x5951]'),
        (0x38306, 'call', '0x4810'),
        (0x3b6bd, 'xor', 'ch, ch'),
        (0x3b6bf, 'shl', 'cx, 1'),
        (0x3b6c8, 'mov', 'ax, bx'),
        (0x3b71d, 'mov', 'word ptr cs:[0x4b66], ax'),
        (0x3b727, 'mul', 'word ptr cs:[0x4b64]'),
        (0x38330, 'call', '0x2ced'),
        (0x3893a, 'call', '0x2527'),
        (0x38983, 'mov', 'bp, 0x28d2'),
        (0x39455, 'mov', 'cl, 0x37'),
        (0x39459, 'mov', 'di, 0x334'),
        (0x3945c, 'mov', 'si, 0x5c80'),
        (0x39481, 'mov', 'di, 0x2a44'),
        (0x39484, 'mov', 'si, 0x70e9'),
        (0x39b87, 'mov', 'di, 0x5cb7'),
        (0x39b8a, 'mov', 'si, 0x5cd0'),
        (0x39b8d, 'mov', 'cx, 0x17b'),
        (0x39b91, 'mov', 'cx, 0x37'),
        (0x39b97, 'add', 'si, 0x19'),
        (0x3a09a, 'mov', 'bx, word ptr [bx + 0x4c01]'),
        (0x38897, 'call', '0x31cd'),
        (0x3b4ed, 'mov', 'word ptr cs:[0x2936], 0x2843'),
        (0x3b21e, 'mov', 'word ptr cs:[0x2936], 0x28bb'),
        (0x3b24e, 'mov', 'bx, 0x4f01'),
    ]:
        dec.expect(intro, base, at, op, operand)
    # Verify full row provenance, not just FORM dimensions.
    require(struct.unpack_from('<7H', intro, 0x6551) == (0, 240, 268, 296, 391, 486, 581), 'unpack destination rows')
    packets, row, at = [], [], 0x663b
    while True:
        require(at < 0x6700, 'raster packet extent')
        v = intro[at]
        at += 1
        if v == 255:
            packets.append(row)
            row = []
            if intro[at] == 255:
                break
        else:
            row.append(v)
    require(len(packets) == 18 and sorted(sum(packets, [])) == list(range(95)), 'raster coverage')
    sidebar = [intro[0x396d3+i*12:0x396d3+(i+1)*12].decode('ascii') for i in range(20)]
    require(sidebar[1] == 'F1 -   PARTY' and sidebar[4] == 'F5 - OPTIONS' and sidebar[8] == 'ESC  -  QUIT', 'demo sidebar')
    require(struct.unpack_from('<5H', intro, 0x5801) == (0x4fbd, 0x4c0d, 0x4c69, 0, 0), 'menu text list')
    # A page consumer prints 12 bounded zero-delimited rows. Publish the
    # interpreted presentation fields, not the DOS memory packing.
    pages = []
    for source in (0x580d, 0x5869):
        cursor, rows = source, []
        for unused in range(12):
            end = intro.index(0, cursor)
            require(end-cursor <= 24, 'menu row bound')
            rows.append(intro[cursor:end].decode('ascii'))
            cursor = end+1
        pages.append(dict(source=source, rows=rows))
    cursor, option_rows = 0x5b01, []
    for unused in range(12):
        end = min(intro.find(0, cursor), cursor+24)
        require(end >= cursor, 'option text terminator')
        option_rows.append(intro[cursor:end].decode('ascii'))
        cursor = end + (end < cursor+24)
    require(any('BALLS:' in r for r in option_rows) and any('SAVE AND EXIT' in r for r in option_rows), 'option rows')
    mz = struct.unpack_from('<14H', table)
    table_load = mz[4]*16
    table_entry = table_load+mz[11]*16+mz[10]
    main_cs_end = table_load+struct.unpack_from('<H', table, 0x56aa)[0]*16
    table_roots = [table_entry, 0x73e, 0xd75, 0xebb, 0xece, 0x1178, 0x104c,
                   0x4517, 0x592b, 0x5cd9, 0x5e9f, 0x629, 0x57a6]
    table_graph, seen = dec.graph(table, 0x300, table_roots, main_cs_end)
    intro_graph, intro_seen = dec.graph(intro, base, [base+5, base+0x293c], 0x3bd70)
    # Conservative raw rel16 scan, before any instruction-boundary filtering.
    # This is supporting evidence ONLY, not an indirect-reachability theorem.
    persistence = []
    for name, lo, hi in [('read', 0x66d2, 0x6706), ('write', 0x6706, 0x673d)]:
        direct = []
        for i in range(0x300, main_cs_end-3):
            if table[i] in (0xe8, 0xe9):
                target = 0x300+((i-0x300+3+struct.unpack_from('<h', table, i+1)[0]) & 65535)
                if lo <= target < hi:
                    direct.append(i)
        pointer = struct.pack('<H', lo-0x300)
        persistence.append(dict(role=name, entry=lo, end_exclusive=hi,
                                raw_rel16_incoming=direct,
                                entry_pointer_in_main_cs=pointer in table[0x300:main_cs_end],
                                entry_pointer_in_ds=pointer in table[0x19db0:0x29db0],
                                reached_by_direct_cfg=lo in seen,
                                indirect_verdict='OPEN: pointer-domain proof required'))
    require(all(not r['raw_rel16_incoming'] and not r['entry_pointer_in_main_cs']
                and not r['entry_pointer_in_ds'] for r in persistence), 'new persistence reference')
    return dict(status='DMO0 NOT CLOSED. DMO1 NOT STARTED.', fingerprint=FINGERPRINT,
                bonus_program=program, canonical_bonus_program=canonical_program,
                table_direct_cfg=table_graph, intro_direct_cfg=intro_graph,
                intro_text_consumer_reached=0x3a09a in intro_seen,
                persistence=persistence,
                drain_count_order=dict(physics_call=0x46b8,drain_call=0x5d4a,
                    physics_return=0x5d4d,counter_call=0x4723,
                    contract='drain handler returns before admitted electronics; drain alone does not skip count'),
                expiry_interleaving=dict(counter_then_tasks=True,
                    new_ball_reset_site=0x3abc, replacement_program_site=0x3b3a,
                    replacement_program='SHOWPLAYERSTS', skip_condition='PARTYFLASH true or VISAKEYS true',
                    matrix_dispatch_site=0x4801, matrix_priority_guard=False,
                    release_hold_site=0x103d, expired_guard_in_these_routines=False,
                    reachability_verdict='OPEN: boundary task/effect preemption reachability required'),
                presentation=dict(cards=dict(source_crop=[0, 0, 440, 95],
                    destinations=[[160, 10], [160, 135]], raster_packets=18,
                    copied_rows=95, palette_banks=[16, 0], page_hold_callbacks=540),
                    sidebar_source=0x396d3, sidebar_rows=sidebar[:10],
                    options_source=0x3974b, options_rows=sidebar[10:], pages=pages,
                    option_page_source=0x5b01, option_page_rows=option_rows),
                remaining_proofs=[
                    'Reachability/domain closure for match, high-score, cheat and indirect matrix/task roots',
                    'External sound-driver callback admission/IRQ contract and native event boundary',
                    'All indirect persistence target domains, including launcher/driver callback boundaries',
                    'Expiry interleaving with pending NEW_BALL/SETBALL and effects; replacement priority/reachability',
                    'Final exhaustive consumed-control differences audit after those closures',
                ])


def audit(data, canonical, historical):
    demo, full = pinned(data, canonical, historical)
    result = inspect(demo, full)
    import audit_10min_demo_sdr
    result['sdr_dispatch'] = audit_10min_demo_sdr.audit(data)
    return result


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    for name, env in [('data', 'PF_10MIN_DEMO_DATA'), ('canonical', 'PF_RUNTIME_DATA'),
                      ('historical', 'PF_DMO0_HISTORICAL_SOURCE')]:
        p.add_argument('--'+name, type=Path, default=os.getenv(env))
    p.add_argument('--output', required=True, type=Path)
    a = p.parse_args()
    require(all((a.data, a.canonical, a.historical)), 'supply pinned private data, canonical and historical paths')
    result = audit(a.data, a.canonical, a.historical)
    a.output.write_text(json.dumps(result, indent=2)+'\n')
    print(result['status'])
    print('Static graph metadata verified; indirect boundaries remain explicit.')
