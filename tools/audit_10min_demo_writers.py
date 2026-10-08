#!/usr/bin/env python3
"""Conservative object overlap evidence for an admitted typed research CFG.

An enumerated writer is not an excluded writer. Unknown segment provenance,
index, string bounds, stack placement or transfer effects remain obligations.
No object or consumer address selects the analysis rules. No payload export.
"""
from collections import deque


def call_entries(r):
    callers={}
    for q,x in r.seen.items():
        if not x.group(r.cs.CS_GRP_CALL):continue
        targets=([r.base+x.operands[0].imm] if x.mnemonic=='call' and x.operands[0].type==r.x86.X86_OP_IMM
                 else r.facts.get(q,{}).get('targets',[]))
        for target in targets:callers.setdefault(target,set()).add(q)
    return callers


def definitions(r, at, requested):
    """Finite backward slice, retaining unknown at every unsupported definition.

    Image segment tokens originate only in CS or relocated immediate words.
    Literal segment numbers are absolute, never silently loader-relative.
    Calls require a register preservation summary; DS and ES are independent.
    """
    todo=deque([(at,requested)]);seen=set();values=set();sources=set();unknown=set()
    from audit_10min_demo_control import relocations
    load,relocs=relocations(r.b)
    callers=getattr(r,'writer_callers',None)
    if callers is None:
        callers=call_entries(r)
    while todo:
        p,reg=todo.popleft()
        if (p,reg) in seen:continue
        seen.add((p,reg))
        if len(seen)>2048:
            unknown.add('definition worklist bound');break
        parents=r.pred[p]
        incoming=callers.get(p,set())
        todo.extend((caller,reg) for caller in incoming)
        if not parents and not incoming:unknown.add('entry context not derived for '+reg+' at '+hex(p))
        for q in parents:
            x=r.seen[q];ops=x.operands
            assigned=(ops and ops[0].type==r.x86.X86_OP_REG and x.reg_name(ops[0].reg)==reg)
            producer=q
            if assigned and x.mnemonic=='pop':
                previous=r.pred[q]
                if len(previous)!=1:
                    unknown.add('stack producer ambiguous at '+hex(q));continue
                producer=next(iter(previous));push=r.seen[producer]
                if push.mnemonic!='push':
                    unknown.add('stack producer unsupported at '+hex(q));continue
                source=push.operands[0];sources.update((producer,q))
            elif assigned and x.mnemonic=='mov':
                source=ops[1];sources.add(q)
            else:
                unused,writes=x.regs_access()
                names={x.reg_name(v) for v in writes}
                aliases={reg,reg[0]+'l',reg[0]+'h'} if reg in ('ax','bx','cx','dx') else {reg}
                if x.group(r.cs.CS_GRP_CALL):
                    targets=([r.base+ops[0].imm] if x.mnemonic=='call' and ops[0].type==r.x86.X86_OP_IMM
                             else r.facts.get(q,{}).get('targets',[]))
                    if not targets or not all(r.preserves(t,reg) for t in targets):
                        unknown.add('call effect not summarized at '+hex(q));continue
                elif names&aliases or x.mnemonic in ('int','iret'):
                    unknown.add('definition unsupported at '+hex(q));continue
                todo.append((q,reg));continue
            if source.type==r.x86.X86_OP_REG:
                name=x.reg_name(source.reg)
                if name=='cs':values.add(('image',r.base))
                else:todo.append((producer,name))
            elif source.type==r.x86.X86_OP_IMM:
                instruction=r.seen[producer]
                relocated=producer+instruction.encoding.imm_offset in relocs
                values.add(('image',load+16*(source.imm&65535)) if relocated
                           else ('literal',source.imm&65535))
            else:unknown.add('memory/register producer unsupported at '+hex(producer))
    if not values and not unknown:unknown.add('definition cycle without producer')
    return dict(values=sorted(values),definitions=sorted(sources),unknown=sorted(unknown))


def overlap(segment, offsets, width, start, end):
    """16-bit EA wrap, byte/word extent, inclusive possible-index coverage.

    Unknown offsets cover all 65536 starts, even with a known segment. Unknown
    segment includes loader placement and therefore cannot exclude any object.
    """
    if segment is None:return True
    if offsets is None:return segment < end and start < segment+65536+width-1
    return any(start<=segment+((offset+i)&65535)<end
               for offset in offsets for i in range(width))


def scan(r, objects, placement=None, module=None, stack_parent=None):
    """Scan every instruction/operand; classify exclusions separately from proof.

    String destinations cover the entire segment unless lifecycle/range analysis
    is supplied by a future proof. Implicit stack writers are not discarded.
    Missing operand access metadata is UNKNOWN rather than read-only by default.
    """
    cache={}
    r.writer_callers=call_entries(r)
    def query(at,reg):
        key=(at,reg)
        if key not in cache:cache[key]=definitions(r,at,reg)
        return cache[key]
    records=[];implicit=[];access_unknown=[]
    stack_ops={'push','pushf','pushfd','pushaw','pushal','enter','call','lcall','int'}
    read_only={'cmp','test','lea','push','call','lcall','jmp','ljmp','lds','les','lfs','lgs','lss','outsb','outsw','outsd','lodsb','lodsw','lodsd','cmpsb','cmpsw','cmpsd','scasb','scasw','scasd'}
    for p,x in sorted(r.seen.items()):
        if x.mnemonic in stack_ops:implicit.append(p)
        for j,op in enumerate(x.operands):
            if op.type!=r.x86.X86_OP_MEM:continue
            if not op.access:
                if x.mnemonic.split()[-1] in read_only:continue
                access_unknown.append(dict(source=p,operand=j));continue
            if not op.access&r.cs.CS_AC_WRITE:continue
            m=op.mem
            seg=x.reg_name(m.segment) if m.segment else ('ss' if x.reg_name(m.base) in ('bp','sp','ebp','esp') else 'ds')
            s=(dict(values=[('image',r.base)],definitions=[],unknown=[]) if seg=='cs' else query(p,seg))
            offsets={m.disp&65535};offset_unknown=[];index_defs=[]
            for reg in (m.base,m.index):
                if not reg:continue
                a=query(p,x.reg_name(reg));index_defs.extend(a['definitions'])
                nums={v for kind,v in a['values'] if kind=='literal'}
                if a['unknown'] or len(nums)!=len(a['values']) or not nums:
                    offset_unknown.extend(a['unknown'] or ['non-numeric index']);offsets=None;break
                scale=m.scale if reg==m.index else 1
                offsets={(v+scale*n)&65535 for v in offsets for n in nums}
                if len(offsets)>4096:offset_unknown.append('EA set bound');offsets=None;break
            string=any(w in x.mnemonic for w in ('stos','movs','insb','insw','insd'))
            if string:
                count=query(p,'cx') if x.mnemonic.startswith(('rep ','repe ','repne ')) else dict(values=[('literal',1)],definitions=[],unknown=[])
                counts={v for kind,v in count['values'] if kind=='literal'}
                index_defs.extend(count['definitions'])
                if offsets is not None and counts and not count['unknown'] and len(counts)==len(count['values']) and max(counts)*op.size<=4096:
                    # Unknown DF is BOTH directions, not a forward-copy assumption.
                    offsets={(start+direction*i*op.size)&65535 for start in offsets
                             for n in counts for i in range(n) for direction in (-1,1)}
                else:
                    offsets=None;offset_unknown.append('string count/direction/recurrence range not derived')
            bases=[v for kind,v in s['values'] if kind=='image']
            seg_unknown=bool(s['unknown'] or len(bases)!=len(s['values']) or not bases)
            before=[o['name'] for o in objects if seg_unknown or any(
                overlap(base,offsets,op.size,o['start'],o['end_exclusive']) for base in bases)]
            proofs=[]
            hits=before
            if placement is not None:
                from audit_10min_demo_placement import exclusion
                proofs=[exclusion(placement,s,offsets,op.size,o) for o in objects]
                # Physical wrap can invalidate an old file-relative exclusion.
                # Only the placement certificate authorizes this verdict.
                hits=[v['target'] for v in proofs
                      if v['status']!='EXCLUDED']
            records.append(dict(source=p,operand=j,width=op.size,segment=seg,
                segment_evidence=s,index_definitions=sorted(set(index_defs)),
                offset_evidence=dict(address_bits=16,displacement=m.disp,
                    base=x.reg_name(m.base),index=x.reg_name(m.index),scale=m.scale,
                    definitions=sorted(set(index_defs)),unknown=offset_unknown,
                    string_count_evidence=count if string else None,
                    string_direction='both DF values' if string else None),
                offsets=None if offsets is None else sorted(offsets),
                range_unknown=offset_unknown,possible_objects=hits,
                possible_objects_before_placement=before,placement_proofs=proofs,
                verdict='MAY OVERLAP' if hits else 'EXCLUDED'))
    boundaries=[p for p,x in sorted(r.seen.items())
                if (x.group(r.cs.CS_GRP_JUMP) or x.group(r.cs.CS_GRP_CALL))
                and (x.mnemonic in ('lcall','ljmp') or not x.operands
                     or x.operands[0].type!=r.x86.X86_OP_IMM)
                and r.facts.get(p,{}).get('status')!='BOUNDED']
    effects=None
    if placement is not None:
        import audit_10min_demo_stack as stack
        effects,context=stack.analyze(r,objects,placement,records,implicit,boundaries,
                                     module or 'typed module',stack_parent)
        r.stack_context=context
    open_stack=implicit if effects is None else [v['site'] for v in effects['implicit_sites'] if v['status']=='UNKNOWN']
    open_transfers=boundaries if effects is None else [v['site'] for v in effects['boundaries'] if v['status']=='UNKNOWN']
    rows=[]
    for o in objects:
        possible=[w['source'] for w in records if o['name'] in w['possible_objects']]
        object_stack=open_stack if effects is None else [v['site'] for v in effects['implicit_sites'] if o['name'] in v['possible_objects']]
        unknown=[]
        if possible:unknown.append('possible writer aliases/ranges not excluded')
        if object_stack:unknown.append('implicit SS:SP stack placement/range not proved')
        if access_unknown:unknown.append('operand access semantics unmodelled')
        if open_transfers:unknown.append('indirect/external memory effects not closed')
        if effects and effects['interrupts']:unknown.append('admitted INT handler memory/placement effects unmodelled')
        rows.append(dict(o,status='UNKNOWN' if unknown else 'BOUNDED',
                         possible_writers=possible,implicit_stack_writers=object_stack,
                         unresolved_transfer_effects=open_transfers,unknown=unknown))
    proved=all(o['status']=='BOUNDED' for o in rows)
    return dict(status='BOUNDED' if proved else 'UNKNOWN',
        scope='all instructions in admitted typed CFG; not whole-image closure',
        instruction_count=len(r.seen),explicit_writer_count=len(records),writers=records,
        implicit_stack_writers=implicit,unknown_operand_access=access_unknown,
        unsummarized_transfer_boundaries=open_transfers,objects=rows,
        original_transfer_boundaries=boundaries,unresolved_implicit_stack_writers=open_stack,
        placement_effects=effects,
        possible_operand_counts_before={o['name']:sum(o['name'] in w['possible_objects_before_placement'] for w in records) for o in objects},
        possible_operand_counts_after={o['name']:sum(o['name'] in w['possible_objects'] for w in records) for o in objects},
        complete_inventory=not access_unknown,
        runtime_immutability_proved=proved,
        unknown=sorted({error for o in rows for error in o['unknown']}))


def target_validity(r, facts):
    """Typed table roots only; reject inconsistent instruction byte ownership."""
    owned={};collisions=set()
    for p,x in sorted(r.seen.items()):
        for byte in range(p,p+x.size):
            if byte in owned:collisions.update((p,owned[byte]))
            owned[byte]=p
    results=[]
    for f in facts:
        routines=[]
        for target in f['targets']:
            todo=[target];seen=set();errors=set();returned=False
            while todo:
                at=todo.pop()
                if at in seen:continue
                seen.add(at)
                x=r.seen.get(at)
                if x is None or at in collisions:
                    errors.add('invalid instruction boundary');continue
                if x.group(r.cs.CS_GRP_RET):
                    returned=True
                    if x.mnemonic!='ret' or x.operands:errors.add('not an unadjusted near return')
                    continue
                if x.group(r.cs.CS_GRP_CALL) or x.group(r.cs.CS_GRP_JUMP):
                    if x.operands[0].type!=r.x86.X86_OP_IMM:
                        errors.add('unresolved recursive transfer');continue
                    todo.append(r.base+x.operands[0].imm)
                if x.mnemonic not in ('jmp','ljmp'):todo.append(at+x.size)
            if not returned:errors.add('no near return reached')
            routines.append(dict(entry=target,instruction_count=len(seen),unknown=sorted(errors)))
        valid=not collisions and all(not v['unknown'] for v in routines)
        results.append(dict(source=f['source'],raw_entry_count=sum(t['entries'] for t in f['tables']),
            distinct_target_count=len(f['targets']),
            all_at_instruction_boundaries=all(t in r.seen for t in f['targets']),
            instruction_overlap_sources=sorted(collisions),candidate_routines=routines,
            status='VALID TYPED CANDIDATES' if valid else 'UNKNOWN',
            premise='operand-derived table roots; does not prove admitted table contents'))
    return results
