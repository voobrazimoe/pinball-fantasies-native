#!/usr/bin/env python3
"""Symbolic DOS EXEC placement certificates, never segment-name heuristics.

Contract: successful normal MZ EXEC in a valid, nonwrapping DOS arena.
No conventional-memory ceiling, fixed base, allocation strategy, or A20 state
is assumed. Unknown transfers remain lifecycle/stack obligations.
"""
import struct

SPACE = 1 << 20
EXEC_SOURCE = 'https://github.com/microsoft/MS-DOS/blob/main/v4.0/src/DOS/EXEC.ASM'
ALLOC_SOURCE = 'https://github.com/microsoft/MS-DOS/blob/main/v4.0/src/DOS/ALLOC.ASM'


def mz(binary):
    from audit_10min_demo_control import relocations
    h = struct.unpack_from('<14H', binary)
    if h[0] != 0x5a4d or not h[2] or h[1] > 511:
        raise ValueError('invalid MZ length/header')
    size = (h[2]-1)*512 + (h[1] or 512)
    load, relocs = relocations(binary)
    if size != len(binary) or size <= load or h[5] > h[6]:
        raise ValueError('MZ file extent/allocation inconsistent')
    # DOS EXEC reserves whole header-declared pages, including the partial page.
    reserved = h[2]*32-h[4]
    required = reserved+h[5]
    if required+16 > 65535 or h[6] == 0:
        raise ValueError('load-high or oversized EXEC not modeled')
    return dict(header_paragraphs=h[4],header_bytes=load,file_size=size,
        image_bytes=size-load,image_paragraphs=(size-load+15)//16,
        exec_reserved_image_paragraphs=reserved,minimum_extra_paragraphs=h[5],
        maximum_extra_paragraphs=h[6],minimum_psp_block_paragraphs=required+16,
        load_base_domain=dict(symbol='B',minimum=16,maximum=65536-required,
                              stride=1,unit='paragraph'),
        psp_segment='B - 0x10',initial_cs_relative=h[11],initial_ip=h[10],
        initial_ss_relative=h[7],initial_sp=h[8],entry_file=load+16*h[11]+h[10],
        relocation_count=len(relocs),relocation_words=sorted(relocs),
        relocation_effect='u16(word + B), once before entry; not runtime mutation',
        allocation_contract='successful normal EXEC; valid nonwrapping DOS arena; '
                            'PSP plus complete reserved image plus minimum extra',
        symbolic_allocation_constraints=['P = B - 0x10',
            'P + allocated_paragraphs <= E <= 0x10000',
            'allocated_paragraphs >= 0x10 + reserved_image + minimum_extra',
            'therefore B + reserved_image + minimum_extra <= 0x10000'],
        admission_boundary='normal EXEC semantics, not a custom/overlay loader or corrupt arena; '
                           'actual free-block endpoints/OS version are not pinned',
        allocation_ceiling_exclusive=SPACE,sources=[EXEC_SOURCE,ALLOC_SOURCE])


def modular_interval(lo, hi, modulus=SPACE):
    """Half-open conservative range, split across physical/offset wrap."""
    if hi <= lo:
        return []
    if hi-lo >= modulus:
        return [[0, modulus]]
    first=lo % modulus
    length=hi-lo
    return ([[first, first+length]] if first+length <= modulus else
            [[first, modulus], [0, first+length-modulus]])


def intersects(left, right):
    return any(a < d and c < b for a,b in left for c,d in right)


def offset_ranges(offsets, width):
    if not 1 <= width <= 65536:
        raise ValueError('unmodeled store width')
    if offsets is None:
        return [[0, 65536]]
    ranges=[]
    for start in offsets:
        ranges.extend(modular_interval(start, start+width, 65536))
    return ranges


def domain(contract, kind, value, ranges, wrap=True):
    """Envelope over ALL B; widening can only prevent an exclusion.

    Symbolic image tokens are file segment bases, not numeric DOS segments.
    Real-mode segment addition wraps at 16 bits; physical wrap is separately
    modeled. Image-image disjointness uses the shared B only if neither wraps.
    """
    if kind == 'literal':
        base=(value & 65535)*16
        lo=hi=base
    elif kind == 'image':
        b=contract['load_base_domain']
        relative=value-contract['header_bytes']
        if relative < 0 or relative % 16:
            return [[0, SPACE if wrap else SPACE+65536]]
        lo=b['minimum']*16+relative
        hi=b['maximum']*16+relative
        # Segment value itself wraps, even with A20 enabled.
        if hi >= SPACE:
            return [[0, SPACE if wrap else SPACE+65536]]
    else:
        return [[0, SPACE if wrap else SPACE+65536]]
    result=[]
    for a,z in ranges:
        result.extend(modular_interval(lo+a,hi+z) if wrap else [[lo+a,hi+z]])
    return result


def exclusion(contract, segment_evidence, offsets, width, obj):
    target_relative=[obj['start']-contract['header_bytes'],
                     obj['end_exclusive']-contract['header_bytes']]
    b=contract['load_base_domain']
    target_unwrapped=[b['minimum']*16+target_relative[0],
                      b['maximum']*16+target_relative[1]]
    valid_target=(0 <= target_relative[0] < target_relative[1] <= contract['image_bytes']
                  and 0 <= target_unwrapped[0] < target_unwrapped[1] <= SPACE)
    target=modular_interval(*target_unwrapped)
    evidence=dict(target=obj['name'],target_physical_domain=target,
        target_symbolic='16*B + [file_start-header_bytes, file_end-header_bytes)',
        load_base_domain=b,offset_ranges=offset_ranges(offsets,width),
        destination_domains=[],status='UNKNOWN',reason=None)
    if not valid_target or segment_evidence['unknown'] or not segment_evidence['values']:
        evidence['reason']='target outside certified image or segment provenance unresolved'
        return evidence
    possible=False
    for kind,value in segment_evidence['values']:
        ranges=evidence['offset_ranges']
        physical=domain(contract,kind,value,ranges)
        # Also evaluate A20 enabled. Do not select a hardware state on faith.
        unwrapped=domain(contract,kind,value,ranges,False)
        hit=intersects(physical,target) or intersects(unwrapped,[target_unwrapped])
        relative=value-contract['header_bytes']
        if kind == 'image' and b['maximum']*16+relative+65536 <= SPACE:
            # Same image identity and same B: retain correlation, not independent envelopes.
            hit=intersects([[relative+a,relative+z] for a,z in ranges],[target_relative])
        possible |= hit
        evidence['destination_domains'].append(dict(kind=kind,value=value,
            classification='absolute-real-mode-segment' if kind=='literal' else 'relocated-image-segment',
            physical_a20_disabled=physical,linear_a20_enabled=unwrapped))
    evidence.update(status='UNKNOWN' if possible else 'EXCLUDED',
        reason='physical destination/target domains may intersect' if possible else
               'disjoint for every admitted B, offset, byte and both A20 states')
    return evidence


def stack_exclusion(contract, ss, interval, obj, reassignments=(), unknown_transfers=()):
    if reassignments or unknown_transfers or ss is None:
        return dict(status='UNKNOWN',reason='SS reassignment or incoming/transfer SS effects unresolved',
                    reassignments=list(reassignments),unknown_transfers=list(unknown_transfers))
    # A half-open interval of possible written offsets, including CALL/INT bytes.
    offsets=None if interval is None else range(interval[0],interval[1])
    return exclusion(contract,dict(values=[ss],unknown=[]),offsets,1,obj)


def transfers(r, boundaries):
    rows=[]
    for at in boundaries:
        f=r.facts.get(at,{})
        effects={n:'UNKNOWN' for n in ('ds','es','ss','sp','allocation','placement')}
        summary=f.get('callee_segment_summaries',{})
        candidates=summary.get('candidates',[])
        for reg in ('ds','es'):
            if candidates and all(c[reg+'_preserved'] for c in candidates):
                effects[reg]='candidate callees preserve; admitted target contents remain UNKNOWN'
        rows.append(dict(source=at,instruction=r.seen[at].mnemonic,
            candidate_targets=f.get('targets',[]),
            effects=effects,
            reason='target candidates do not establish admitted callee/lifecycle effects'))
    interrupts=[dict(source=p,vector=x.operands[0].imm,
        effects={n:'UNKNOWN' for n in ('ds','es','ss','sp','allocation','placement')},
        stack_effect='at least FLAGS, CS, IP; handler depth/SS switches not bounded')
        for p,x in sorted(r.seen.items()) if x.mnemonic=='int']
    assignments=[]
    for p,x in sorted(r.seen.items()):
        unused,writes=x.regs_access()
        names={x.reg_name(v) for v in writes}
        if names & {'ss','sp','esp'} and x.mnemonic not in (
            'push','pop','pushaw','popaw','pushf','popf','call','lcall','ret','retf','iret','int'):
            assignments.append(dict(source=p,instruction=x.mnemonic,operand=x.op_str,
                                    writes=sorted(names & {'ss','sp','esp'})))
        elif x.mnemonic=='pop' and x.op_str in ('ss','sp','esp'):
            assignments.append(dict(source=p,instruction=x.mnemonic,operand=x.op_str,
                                    writes=[x.op_str]))
    return dict(boundaries=rows,interrupts=interrupts,stack_assignments=assignments,
        stack_direction='decrement SP modulo 65536 before stores',
        safe_initial_ss_offset_interval=[0,65536],
        maximum_depth='UNKNOWN; declared stack size is not a dynamic depth bound',
        lifecycle='UNKNOWN across external transfers, interrupts and additional entry roots')


def derive(r, launcher):
    contract=mz(r.b)
    launch=mz(launcher)
    prefix=list(r.dec.decoder.disasm(launcher[launch['entry_file']:launch['entry_file']+26],
                                     launch['entry_file']))
    if ([x.mnemonic for x in prefix] != ['pushaw','push','pop','popaw','push','pop',
            'push','pop','mov','mov','neg','add','add','int']
            or [x.op_str for x in prefix[1:3]] != ['ds','ds']
            or [x.op_str for x in prefix[4:6]] != ['ds','es']
            or [x.op_str for x in prefix[8:11]] != ['ax, 0x4a00','bx, es','bx']
            or prefix[12].op_str != 'bx, 0x40' or prefix[13].op_str != '0x21'):
        raise ValueError('launcher resize entry grammar changed')
    last=prefix[11]
    if last.address+last.encoding.imm_offset not in launch['relocation_words']:
        raise ValueError('launcher resize endpoint not relocated')
    launch['startup']=dict(resize_site=prefix[13].address,
        endpoint_definition=last.address,retained_image_end_relative=(last.operands[1].imm+64)*16,
        requested_psp_paragraphs=last.operands[1].imm+64+16,
        failure='CF not checked; normal EXEC child still uses its own MZ allocation')
    # Recognize the actual TABLE1 startup, without a site-address allowlist.
    entry=contract['entry_file']
    code=list(r.dec.decoder.disasm(r.b[entry:entry+19],entry))
    expected=['mov','mov','mov','sub','add','int','mov','mov']
    if [x.mnemonic for x in code] != expected or [x.op_str for x in code[:2]] != ['bp, ds','ah, 0x4a']:
        raise ValueError('TABLE1 resize entry grammar changed')
    if (code[3].op_str != 'bx, bp' or code[4].op_str != 'bx, 0x40'
            or code[5].op_str != '0x21' or code[7].op_str != 'ds, ax'):
        raise ValueError('TABLE1 resize/data grammar changed')
    relocs=set(contract['relocation_words'])
    for x in (code[2],code[6]):
        if x.address+x.encoding.imm_offset not in relocs:
            raise ValueError('startup segment immediate not relocated')
    retained=(code[2].operands[1].imm+64)*16
    data=(code[6].operands[1].imm)*16
    contract['startup']=dict(resize_site=code[5].address,
        segment_definition=code[2].address,psp_source=code[0].address,
        requested_psp_paragraphs=code[2].operands[1].imm+64+16,
        retained_image_end_relative=retained,
        released_tail_relative=[retained,contract['exec_reserved_image_paragraphs']*16],
        failure='CF not checked; on failed shrink original larger block remains',
        data_definition=code[6].address,data_segment_relative=data,
        later_lifecycle='UNKNOWN; external allocation/free/resize effects not closed')
    # Launcher direct CFG verifies actual normal EXEC instruction operands.
    base=launch['header_bytes']+16*launch['initial_cs_relative']
    cfg,seen=r.dec.graph(launcher,base,[launch['entry_file']],len(launcher))
    exec_sites=[]
    for p,x in seen.items():
        if x.mnemonic != 'int' or x.op_str != '0x21':continue
        preceding=[q for at,q in seen.items() if at+q.size==p]
        if len(preceding)==1 and preceding[0].mnemonic=='mov' and preceding[0].op_str=='ax, 0x4b00':
            exec_sites.append(p)
    if not exec_sites:
        raise ValueError('launcher normal EXEC grammar not derived')
    contract['launcher']=dict(header=launch,normal_exec_sites=sorted(exec_sites),
        stack_assignments=[dict(source=p,instruction=x.mnemonic,operand=x.op_str)
            for p,x in sorted(seen.items()) if (x.mnemonic=='mov'
            and x.op_str.startswith(('ss,','sp,'))) or (x.mnemonic=='pop' and x.op_str=='ss')],
        lifecycle='separate PSP/stack; child initial SS:SP comes from child MZ; '
                  'parent stack restores do not assign child SS')
    sub=r.related_resolver
    contract['typed_extents']=[dict(module='TABLE1 initial complete load image',relative=[0,contract['image_bytes']]),
        dict(module='TABLE1 code slice',relative=[r.base-contract['header_bytes'],sub.base-contract['header_bytes']]),
        dict(module='CODE2 linked image slice',relative=[sub.base-contract['header_bytes'],data]),
        dict(module='demo data segment addressable window',relative=[data,data+65536]),
        dict(module='post-CODE2 linked image through startup LASTSEG',relative=[data,retained-1024])]
    b=contract['load_base_domain']
    for extent in contract['typed_extents']:
        lo,hi=extent['relative']
        extent.update(symbolic='16*B + relative',physical_envelope=[16*b['minimum']+lo,16*b['maximum']+hi],
                      identity='same TABLE1 MZ allocation; not a separately loaded module')
    contract['separate_allocations']='resident/module allocations not certified; no allocation-based exclusion granted'
    return contract
