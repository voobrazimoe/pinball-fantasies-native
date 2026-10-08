#!/usr/bin/env python3
"""Static, in-memory EXEPACK/SDR dispatch evidence; no DOS execution/output bytes.

This decodes container blocks, not instructions. No uncompressed executable is
written to disk. Reviewed Capstone sites below describe two dispatch families;
full callback admission closure for every driver remains a separate obligation.
"""
from pathlib import Path
import struct

from audit_10min_demo_graph import Decoder, require, sha

SDR = {
    'ADLIB.SDR': (10966, '19d2e5e7f6b4b0b11d9b6ae8070dc558a0303d56647142966ec2713b4a805724', 36048),
    'GUS.SDR': (10502, 'b32e0306634bb9285ff402e2985eeebe66b2eff09a0c313b7fe5c9960e1d42ac', 17488),
    'INTERNAL.SDR': (10711, '281e62186567dbf3bc670d397ccf7fbfba37eee5844440796856c0e1d39a6fe8', 35664),
    'NOSOUND.SDR': (2883, '223fdd845fa8541e2801c3371f2f97a61244e4ea54bcd81f2a51d312c005366d', 3312),
    'PAS16.SDR': (10712, 'e863cb16e9e30cc2506638b7f91a2a31e26d19c242429af065ad455554352fd2', 36160),
    'SB16.SDR': (11349, '81a7c4d41a14ae042a63899b38e88bc4d33a547053f44b17b3bd324ebe7bab17', 36592),
    'SB20.SDR': (11814, '4cba5dff50926c9c2485d5242e935a4bc4d134b1d8b8fce4db0d5faa0e044e9c', 35856),
    'SBLASTER.SDR': (11421, 'c6c016f94f985c0f70d6bba7acb7e4572db8bbf88d99b2714aa150313be4b21f', 35536),
    'SBPRO.SDR': (11886, '3dfdea95cd3cd7c2096a4bf59fcf0af4802e20ca1879e149c9f1f39f57e27f20', 35920),
    'SM2.SDR': (11348, 'd305e3401899af6d4938c09f85d0a1dc0aec0361b6ce8cb434933bdd298208b8', 35408),
    'THING.SDR': (10369, 'c9cd962aae321ce1e2637287a559a7d2aa6032c468ae6d794c3a7f4e3e491836', 34688),
}


def unpack(b):
    require(len(b) >= 28 and b[:2] == b'MZ', 'SDR MZ header')
    h = struct.unpack_from('<14H', b)
    load, stub = h[4]*16, (h[4]+h[11])*16
    require(load < stub and stub+18 <= len(b), 'EXEPACK extent')
    header = struct.unpack_from('<9H', b, stub)
    # GUS has entry IP 22; the other pinned containers have entry IP 6.
    # Identity verification in audit precedes this container interpretation.
    require(header[0] in (6, 22) and header[1:3] == (0, 0)
            and header[8] == 0x4252 and header[7] == 1,
            'reviewed EXEPACK variant')
    p, out, count = stub-1, bytearray(), 0
    while b[p] == 255:
        p -= 1
        require(p >= load, 'EXEPACK padding')
    while True:
        require(p-2 >= load, 'EXEPACK command bound')
        command = b[p]
        p -= 1
        n = struct.unpack_from('<H', b, p-1)[0]
        p -= 2
        require(command & 254 in (176, 178), 'EXEPACK block type')
        if command & 254 == 176:
            require(p >= load, 'EXEPACK repeat bound')
            out.extend(bytes([b[p]])*n)
            p -= 1
        else:
            require(p-n+1 >= load, 'EXEPACK literal bound')
            out.extend(reversed(b[p-n+1:p+1]))
            p -= n
        count += 1
        require(len(out) <= header[6]*16 and count < 10000, 'EXEPACK output bound')
        if command & 1:
            break
    out.reverse()
    # Terminal block leaves a literal leading prefix in the load image.
    raw = b[load:p+1]+bytes(out)
    require(len(raw) == header[6]*16, 'EXEPACK destination extent')
    return raw, count


def audit(data):
    sources, rows = {}, []
    for name, (size, digest, unpacked_size) in SDR.items():
        b = (Path(data)/name).read_bytes()
        require(len(b) == size and sha(b) == digest, 'not the pinned SDR: '+name)
        raw, blocks = unpack(b)
        require(len(raw) == unpacked_size, 'SDR decoded size')
        sources[name] = raw
        rows.append(dict(name=name, sha256=digest, decoded_size=len(raw), blocks=blocks))
    dec = Decoder()
    for name, at, mnemonic, operand in [
        ('NOSOUND.SDR', 0x6c4, 'mov', 'word ptr [0x46d], dx'),
        ('NOSOUND.SDR', 0x6c8, 'mov', 'word ptr [0x46f], es'),
        ('NOSOUND.SDR', 0x5bf, 'cld', None),
        ('NOSOUND.SDR', 0x5c7, 'sti', None),
        ('NOSOUND.SDR', 0x66b, 'cmp', 'al, byte ptr cs:[0x6a1]'),
        ('NOSOUND.SDR', 0x670, 'jb', '0x690'),
        ('NOSOUND.SDR', 0x677, 'mov', 'byte ptr cs:[0x6a1], al'),
        ('NOSOUND.SDR', 0x687, 'lcall', '[si - 4]'),
        ('ADLIB.SDR', 0x1a89, 'mov', 'word ptr [0x6bcd], dx'),
        ('ADLIB.SDR', 0x1a8d, 'mov', 'word ptr [0x6bcf], es'),
        ('ADLIB.SDR', 0xbd1, 'call', '0x1a32'),
        ('ADLIB.SDR', 0xbdd, 'lcall', '[0x6bcd]'),
        ('ADLIB.SDR', 0xc11, 'lcall', '[0x6c33]'),
        ('ADLIB.SDR', 0x1a36, 'cmp', 'byte ptr cs:[0x156b], 0'),
    ]:
        dec.expect(sources[name], 0, at, mnemonic, operand)
    return dict(status='OPEN: all-driver admission/IRQ proof not complete', inputs=rows,
                api_domains=[callback_api(name, raw, dec) for name, raw in sources.items()],
                reviewed_dispatch=[
                    dict(driver='NOSOUND.SDR', domain='decoded module offsets',
                         callback_install=0x6bc, irq_entry=0x5bf, priority_gate=0x66b,
                         callback=0x687, kind='raster-calibrated IRQ0 record scheduler'),
                    dict(driver='ADLIB.SDR', domain='decoded module offsets',
                         callback_install=0x1a81, vblank_callback=0xbdd,
                         late_callback=0xc11, budget_check=0x1a32,
                         kind='buffer/interrupt callback scheduler'),
                ])


def callback_api(name, raw, dec):
    """Derive INT66 entry, AL dispatch and DX:ES callback storage in all SDRs.

    A callback call found by its operand is a candidate consumer, not an IRQ
    admission proof. Keep that distinction explicit in the exported metadata.
    """
    from capstone import x86
    entry = 22 if name == 'GUS.SDR' else 6
    initial = list(dec.decoder.disasm(raw[entry:entry+40], entry))
    data_segment = next(x.operands[0].imm for x in initial
                        if x.mnemonic == 'push' and x.operands[0].type == x86.X86_OP_IMM)
    api = None
    for i, x in enumerate(initial):
        if (x.mnemonic == 'mov' and len(x.operands) == 2
                and x.operands[0].type == x86.X86_OP_MEM
                and x.operands[0].mem.disp == 0x198):
            previous = initial[i-1]
            require(previous.mnemonic == 'mov' and previous.op_str.startswith('ax, ')
                    and previous.operands[1].type == x86.X86_OP_IMM,
                    'INT66 vector producer')
            api = previous.operands[1].imm
    require(api is not None, 'INT66 entry not derived')
    dispatch, at = {}, api+5
    for unused in range(64):
        cmp = dec.instruction(raw, 0, at)
        if cmp.mnemonic != 'cmp' or cmp.op_str.split(',')[0] != 'al':
            break
        branch = dec.instruction(raw, 0, at+cmp.size)
        jump = dec.instruction(raw, 0, at+cmp.size+branch.size)
        if branch.mnemonic != 'jne' or jump.mnemonic != 'jmp':
            require(11 in dispatch and 12 in dispatch, 'SDR AL dispatcher before callback APIs')
            break
        dispatch[cmp.operands[1].imm] = jump.operands[0].imm
        at = branch.operands[0].imm
    require(11 in dispatch and 12 in dispatch, 'primary/later registration APIs')
    registrations = []
    for code in (11, 12):
        unused, seen = dec.graph(raw, 0, [dispatch[code]], len(raw))
        pairs = []
        for p, x in seen.items():
            ops = x.operands
            if (x.mnemonic == 'mov' and len(ops) == 2
                    and ops[0].type == x86.X86_OP_MEM
                    and not ops[0].mem.base and not ops[0].mem.index
                    and ops[1].type == x86.X86_OP_REG and x.reg_name(ops[1].reg) == 'dx'):
                offset = ops[0].mem.disp
                adjacent = [q for q,y in seen.items() if y.mnemonic == 'mov'
                            and len(y.operands)==2 and y.operands[0].type==x86.X86_OP_MEM
                            and not y.operands[0].mem.base and not y.operands[0].mem.index
                            and y.operands[0].mem.disp==offset+2
                            and y.operands[1].type==x86.X86_OP_REG
                            and y.reg_name(y.operands[1].reg)=='es']
                if adjacent:
                    pairs.append(dict(offset_store=p,segment_stores=sorted(adjacent),
                                      pointer_ds_offset=offset))
        if not pairs:
            prefix=list(dec.decoder.disasm(raw[dispatch[code]:dispatch[code]+48],dispatch[code]))
            bases=[x.operands[1].imm for x in prefix if x.mnemonic=='mov'
                   and len(x.operands)==2 and x.op_str.startswith('si, ')
                   and x.operands[1].type==x86.X86_OP_IMM]
            counts=[x.operands[1].imm for x in prefix if x.mnemonic=='mov'
                    and len(x.operands)==2 and x.op_str.startswith('ax, ')
                    and x.operands[1].type==x86.X86_OP_IMM]
            pop_fields={}
            for q,y in seen.items():
                if y.mnemonic=='pop' and y.operands[0].type==x86.X86_OP_MEM:
                    m=y.operands[0].mem
                    if m.base==x86.X86_REG_SI and not m.index:
                        pop_fields[m.disp]=q
            require(bases and counts and 5 in pop_fields and 7 in pop_fields,
                    'indexed callback record producer: '+name)
            registrations.append(dict(api=code,entry=dispatch[code],producer=dict(
                kind='sorted indexed record insertion',record_base=bases[0],
                loop_bound=counts[0],record_stride=9,offset_store=pop_fields[5],
                segment_store=pop_fields[7]),candidate_absolute_consumers=[],
                unknown=['record insertion/removal lifecycle and IRQ consumer bound']))
            continue
        require(len(pairs)==1,'unique registration callback storage: '+name)
        pair=pairs[0]
        # Locate operand-compatible far consumers throughout decoded CS.
        # CFG/IRQ reachability must separately qualify those candidates.
        candidates=[]
        for p in range(data_segment*16):
            if raw[p] not in (0xff,0x2e):
                continue
            x=next(dec.decoder.disasm(raw[p:p+8],p),None)
            if x and x.mnemonic=='lcall' and x.operands[0].type==x86.X86_OP_MEM:
                m=x.operands[0].mem
                if not m.base and not m.index and m.disp==pair['pointer_ds_offset']:
                    candidates.append(p)
        registrations.append(dict(api=code,entry=dispatch[code],producer=pair,
                                  candidate_absolute_consumers=candidates))
    return dict(driver=name,int66_entry=api,data_segment=data_segment,
                callback_registration=registrations,
                status='registration domain derived; IRQ/admission UNKNOWN')
