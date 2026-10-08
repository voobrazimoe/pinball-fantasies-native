#!/usr/bin/env python3
"""Operand-derived control evidence. Candidate sets are not closure proofs.

All offsets refer to the private input or an in-memory decoded module. No
payload is exported. In particular, a literal relocated pointer does not prove
the absence of later aliased writes, and an API installer does not prove which
installer ran before an admitted consumer. Those obligations remain UNKNOWN.
"""
import struct

import audit_10min_demo_graph as graph


def predecessors(r, at, count):
    chain = []
    for unused in range(count):
        parents = r.pred[at]
        graph.require(len(parents) == 1, 'nonunique idiom predecessor')
        at = next(iter(parents))
        chain.append((at, r.seen[at]))
    return list(reversed(chain))


def local_value(r, at, reg, active=()):
    """All predecessor paths; only MOV and an adjacent balanced PUSH/POP.

    Segment tokens retain the module identity. Do not assign entry-register
    values, cross a call on faith, or treat a linear listing as a CFG proof.
    """
    key = (at, reg)
    if key in active:
        return set(), [], []  # loop backedge contributes no new definition
    if len(active) >= 128:
        return set(), [], ['segment reaching-definition bound']
    values, definitions, errors = set(), [], []
    if not r.pred[at]:
        return set(), [], ['segment/register entry value is unknown']
    for p in r.pred[at]:
        x = r.seen[p]
        ops = x.operands
        assigned = (ops and ops[0].type == r.x86.X86_OP_REG
                    and x.reg_name(ops[0].reg) == reg)
        if assigned and x.mnemonic == 'pop':
            parents = r.pred[p]
            if len(parents) != 1:
                errors.append('nonunique stack producer at '+hex(p)); continue
            q = next(iter(parents)); push = r.seen[q]
            if push.mnemonic != 'push':
                errors.append('unproved stack producer at '+hex(p)); continue
            source = push.operands[0]; definitions.extend((p,q))
        elif assigned and x.mnemonic == 'mov':
            source = ops[1]; q = p; definitions.append(p)
        else:
            unused, writes = x.regs_access()
            names = {x.reg_name(v) for v in writes}
            if x.group(r.cs.CS_GRP_CALL):
                targets=[]
                if ops[0].type==r.x86.X86_OP_IMM:
                    targets=[r.base+ops[0].imm]
                elif p in r.facts:
                    targets=r.facts[p]['targets']
                if not targets or not all(r.preserves(t,reg) for t in targets):
                    errors.append('segment/register call clobber not summarized at '+hex(p));continue
                definitions.append(p)
            elif reg in names:
                errors.append('segment/register clobber not summarized at '+hex(p)); continue
            v,d,u = local_value(r,p,reg,active+(key,))
            values.update(v);definitions.extend(d);errors.extend(u);continue
        if source.type == r.x86.X86_OP_IMM:
            values.add(source.imm & 65535)
        elif source.type == r.x86.X86_OP_REG:
            name = x.reg_name(source.reg)
            if name == 'cs':
                values.add(('module',r.base))
            else:
                v,d,u = local_value(r,q,name,active+(key,))
                values.update(v);definitions.extend(d);errors.extend(u)
        else:
            errors.append('unsupported segment/register source at '+hex(q))
    return values, sorted(set(definitions)), sorted(set(errors))


def relocations(b):
    h = struct.unpack_from('<14H',b)
    load = h[4]*16
    graph.require(h[12]+4*h[3] <= load, 'MZ relocation extent')
    entries=[load+segment*16+offset for offset,segment in
             struct.iter_unpack('<HH',b[h[12]:h[12]+4*h[3]])]
    graph.require(len(entries)==len(set(entries)), 'duplicate MZ relocation lifecycle unmodelled')
    graph.require(all(load<=p<len(b)-1 for p in entries), 'MZ relocation word outside image')
    return load, set(entries)


def far_operand(r, at):
    x = r.seen[at]; op = x.operands[0]
    f = dict(source=at,domain='far',status='UNKNOWN',targets=[],definitions=[],
             tables=[],module_targets=[],recipe='far operand reaching segment definitions',
             unknown=[],operand_class='memory far pointer',domain_category='unresolved far')
    if op.type != r.x86.X86_OP_MEM or op.mem.base or op.mem.index:
        f['unknown']=['indexed/non-memory far-pointer construction not proved'];return f
    m = op.mem
    segment = x.reg_name(m.segment) if m.segment else 'ds'
    values,defs,errors = local_value(r,at,segment)
    f['definitions']=defs;f['segment_sources']=[list(v) if isinstance(v,tuple) else v
                                               for v in sorted(values,key=str)]
    f['unknown'].extend(errors)
    for value in values:
        if value == 0:
            f['domain_category']='external/API binding'
            f['binding_slot']=dict(segment=0,offset=m.disp)
            f['unknown'].append('vector installer/admitted consumer ordering and writer closure not proved')
        elif value == ('module',r.base):
            source=r.base+m.disp
            if not r.base <= source < r.limit-3:
                f['unknown'].append('far pointer outside owning CS extent');continue
            load,relocs=relocations(r.b)
            offset,seg=struct.unpack_from('<HH',r.b,source)
            f['tables'].append(dict(pointer_source=source,offset=offset,
                                   image_segment=seg,segment_relocation=source+2 in relocs))
            if any(p in relocs for p in (source-1,source,source+1,source+3)):
                f['unknown'].append('far pointer has additional overlapping loader relocation');continue
            if source+2 not in relocs:
                f['unknown'].append('far segment has no MZ relocation producer');continue
            base=load+seg*16;target=base+offset
            f['domain_category']='relocated module pointer'
            # This is a typed CALL operand, not an arbitrary address scan.
            # Only the segment's offset-zero entry is admitted for inspection.
            if offset or not r.limit <= base < r.ds:
                f['unknown'].append('destination code-entry role/extent not established');continue
            f['module_targets'].append(dict(module='CODE2',segment_file_base=base,
                                            offset=offset,target_file=target))
            f['unknown'].append('CS pointer immutability requires complete ES/DS/indexed writer aliases')
        else:
            f['unknown'].append('unclassified far operand segment source')
    if not values and not f['unknown']:
        f['unknown'].append('no far-pointer segment producer')
    f['unknown']=sorted(set(f['unknown']))
    return f


def area_operand(r, at):
    """Derive all bases at the ten-byte rectangle loop, excluding its recurrence.

    The excluded edge is proved to restore the saved row cursor then advance
    one whole row. It is not an arbitrary backward-slice cut. DS provenance and
    table immutability are deliberately independent obligations.
    """
    result=dict(offsets=[],definitions=[],tables=[],
                recipe='CFG rectangle loop: four bounds, handler, zero sentinel',unknown=[])
    try:
        # Find the final handler LODSW through the optional previous-area guard.
        chain=predecessors(r,at,1)
        p,x=chain[0]
        if x.mnemonic != 'lodsw':
            chain=predecessors(r,at,7)
            p,x=chain[0]
        graph.require(x.mnemonic=='lodsw','area handler word load')
        # The loop body from saved cursor to final load has a fixed grammar.
        headers=[h for h,q in r.seen.items() if h<p and p-h<64
                 and q.mnemonic=='push' and q.op_str=='si']
        graph.require(len(headers)==1,'unique area saved cursor')
        header=headers[0];cursor=header+r.seen[header].size;body=[]
        while cursor<=p:
            q=r.seen.get(cursor)
            graph.require(q is not None,'area loop instruction boundary')
            body.append((cursor,q));cursor+=q.size
        graph.require([q.mnemonic for unused,q in body if q.mnemonic=='lodsw']==['lodsw']*5,
                      'four area bounds plus handler word')
        graph.require(all(q.op_str=='ax, word ptr [si]' for unused,q in body
                          if q.mnemonic=='lodsw'),'area load base/width')
        increments=[s for s in r.pred[header] if r.seen[s].mnemonic=='add'
                    and r.seen[s].op_str=='si, 0xa']
        graph.require(len(increments)==1,'unique area cursor increment')
        increment=increments[0]
        back=predecessors(r,increment,1)+[(increment,r.seen[increment])]
        graph.require(back[0][1].mnemonic=='pop' and back[0][1].op_str=='si'
                      and back[1][1].mnemonic=='add' and back[1][1].op_str=='si, 0xa',
                      'area cursor recurrence')
        graph.require([q.mnemonic for unused,q in body if q.group(r.cs.CS_GRP_JUMP)]
                      ==['je','jb','jb','ja','ja'],'area sentinel/bounds branches')
        branches=[(s,q) for s,q in body if q.group(r.cs.CS_GRP_JUMP)]
        graph.require(all(r.base+q.operands[0].imm==back[0][0] for unused,q in branches[1:]),
                      'area rejection branches restore row cursor')
        graph.require([q.op_str for unused,q in body if q.mnemonic=='cmp']
                      ==['cx, ax','dx, ax','cx, ax','dx, ax'],'area bound comparisons')
        graph.require(body[0][1].mnemonic=='lodsw' and body[1][1].op_str=='ax, ax',
                      'area zero sentinel')
        graph.require(all(q.mnemonic in ('lodsw','or','je','jb','ja','cmp','nop')
                          for unused,q in body),'area loop has no other definitions')
        sentinel=r.base+branches[0][1].operands[0].imm
        # The zero row must bypass the handler call.
        graph.require(sentinel>at,'area sentinel bypasses call')
        values,defs,errors=r.reaching(header,'si',recurrence_edges={(back[-1][0],header)})
        graph.require(all(isinstance(v,int) for v in values),'area base must be immediate')
        targets=set()
        for base in sorted(values):
            rows,end=r.table(base,10,8)
            targets.update(q['target']-r.base for q in rows)
            result['tables'].append(dict(start=r.ds+base,end=end,rows=rows))
        result.update(offsets=sorted(targets),definitions=sorted(set(defs+[s for s,q in body]+[header]
                                                                  +[s for s,q in back])))
        result['unknown'].extend(errors)
        result['unknown'].append('DS entry/callee alias and table writer completeness not proved')
    except (ValueError,KeyError,IndexError) as e:
        result['unknown'].append('area grammar not proved: '+str(e))
    return result


def glyph_operand(r, at):
    """Operand-derived masked word-table candidates, independent of site address."""
    result=dict(source=at,domain='near',domain_category='masked glyph table',status='UNKNOWN',
                instruction='call',operand_class='register',target_module='CODE2',
                completeness_reason=None,targets=[],definitions=[],tables=[],unknown=[])
    try:
        call=r.seen[at]
        graph.require(call.mnemonic=='call' and call.op_str=='dx','glyph CALL DX consumer')
        chain=[];cursor=at
        for unused in range(16):
            p,x=predecessors(r,cursor,1)[0];chain.append((p,x));cursor=p
            if x.mnemonic=='and':break
        chain.reverse()
        # Match the suffix from AND to CALL; preceding character input is arbitrary.
        cuts=[i for i,(p,x) in enumerate(chain) if x.mnemonic=='and']
        graph.require(len(cuts)==1,'unique glyph mask')
        chain=chain[cuts[0]:]
        meaningful=[(p,x) for p,x in chain if not (x.mnemonic=='mov' and
                    x.operands[0].type==r.x86.X86_OP_REG and x.op_str.startswith('ax, ')
                    and x.operands[1].type==r.x86.X86_OP_IMM)]
        graph.require([x.mnemonic for p,x in meaningful]==['and','shl','add','mov','mov','add'],
                      'glyph table grammar')
        mask,shift,base,transfer,load,add=[x for p,x in meaningful]
        index=mask.reg_name(mask.operands[0].reg)
        graph.require(mask.operands[1].type==r.x86.X86_OP_IMM
                      and shift.op_str==index+', 1' and base.reg_name(base.operands[0].reg)==index
                      and base.operands[1].type==r.x86.X86_OP_IMM,'glyph bounded index')
        graph.require(transfer.op_str=='di, '+index and load.op_str=='dx, word ptr es:[di]'
                      and add.operands[1].type==r.x86.X86_OP_IMM
                      and add.reg_name(add.operands[0].reg)=='dx','glyph load/code addend')
        bound=mask.operands[1].imm
        graph.require(0<=bound<=255 and bound&(bound+1)==0,'contiguous glyph mask')
        table=base.operands[1].imm;addend=add.operands[1].imm
        targets={r.base+((r.word(r.ds+table+2*i)+addend)&65535) for i in range(bound+1)}
        graph.require(all(r.base<=v<r.limit for v in targets),'glyph targets leave module code extent')
        result.update(targets=sorted(targets),definitions=[p for p,x in chain],
                      tables=[dict(start=r.ds+table,entries=bound+1,stride=2,code_addend=addend)])
        result['unknown']=['ES alias/table writer completeness not proved',
                           'caller far-pointer domain remains UNKNOWN']
    except (ValueError,KeyError,IndexError) as e:
        result['unknown']=['glyph grammar not proved: '+str(e)]
    return result


def related_code2(r, facts):
    from audit_10min_demo_domains import Resolver
    entries=sorted({t['target_file'] for f in facts for t in f.get('module_targets',[])})
    if not entries:
        return dict(rounds=[],domains=[],root_files=[],instruction_count=0,
                    bounded_domain_count=0,unknown_domain_count=0)
    graph.require(len(entries)==1,'ambiguous CODE2 entry modules')
    sub=Resolver(r.b,entries[0],r.ds,entries,r.ds,decoder=r.dec)
    rounds=[]
    for iteration in range(32):
        sub.prepare()
        facts=[glyph_operand(sub,f['source']) for f in sub.cfg['indirect_boundaries']]
        sub.facts={f['source']:f for f in facts}
        added={t for f in facts for t in f['targets']}-sub.roots
        rounds.append(dict(iteration=iteration,roots=len(sub.roots),instructions=len(sub.seen),
                           indirect_sites=len(facts),added_roots=len(added)))
        if not added:
            for f in facts:
                if not f['tables']:continue
                load_at=next(p for p in f['definitions'] if sub.seen[p].op_str=='dx, word ptr es:[di]')
                values,definitions,errors=local_value(sub,load_at,'es')
                f['es_alias_evidence']=dict(values=sorted(values,key=str),definitions=definitions,
                                            unknown=errors,
                    premise='candidate glyph call domains; their summaries do not prove table immutability')
                image_load,relocs=relocations(sub.b)
                immediate_sources=[p for p in definitions if sub.seen[p].mnemonic=='push'
                    and sub.seen[p].operands[0].type==sub.x86.X86_OP_IMM]
                f['es_alias_evidence']['image_segment_bindings']=[dict(source=p,
                    image_segment=sub.seen[p].operands[0].imm,
                    file_base=image_load+16*sub.seen[p].operands[0].imm,
                    relocated=p+sub.seen[p].encoding.imm_offset in relocs) for p in immediate_sources]
                bindings=f['es_alias_evidence']['image_segment_bindings']
                if (not errors and values=={(sub.ds-image_load)//16} and bindings
                        and all(b['relocated'] and b['file_base']==sub.ds for b in bindings)):
                    f['unknown']=['word-table writer completeness not proved',
                                  'caller far-pointer domain remains UNKNOWN']
                f['candidate_targets']=list(f['targets'])
            import audit_10min_demo_writers as writers
            r.related_resolver=sub
            es_summaries=[dict(source=f['source'],candidate_count=len(f['targets']),
                candidates=[dict(entry=t,es_preserved=sub.preserves(t,'es'),
                                 ds_preserved=sub.preserves(t,'ds')) for t in f['targets']])
                for f in facts]
            for f,summary in zip(facts,es_summaries):
                f['callee_segment_summaries']=summary
                if not all(c['es_preserved'] for c in summary['candidates']):
                    f['unknown'].append('candidate ES behavior unresolved')
            return dict(rounds=rounds,domains=facts,root_files=sorted(sub.roots),
                        target_validity=writers.target_validity(sub,facts),
                        callee_segment_summaries=es_summaries,
                        instruction_count=len(sub.seen),
                        bounded_domain_count=sum(f['status']=='BOUNDED' for f in facts),
                        unknown_domain_count=sum(f['status']=='UNKNOWN' for f in facts),
                        scope='candidate CODE2 CFG; aliases and upstream domains remain UNKNOWN')
        sub.roots.update(added)
    raise ValueError('CODE2 candidate fixed point did not converge')


def api_installers(data):
    """Actual entry-path vector stores in all supplied SDRs, without IRQ expansion."""
    import audit_10min_demo_sdr as sdr
    dec=graph.Decoder();rows=[]
    for name,(size,digest,extent) in sdr.SDR.items():
        b=(data/name).read_bytes()
        graph.require(len(b)==size and graph.sha(b)==digest,'not the pinned SDR: '+name)
        raw,unused=sdr.unpack(b)
        graph.require(len(raw)==extent,'SDR decoded extent')
        # Container entry is encoded in the retained EXEPACK header, not raw word 0.
        h=struct.unpack_from('<14H',b);stub=(h[4]+h[11])*16
        entry=struct.unpack_from('<H',b,stub)[0]
        prefix=list(dec.decoder.disasm(raw[entry:entry+48],entry))
        stores=[(i,x) for i,x in enumerate(prefix) if x.mnemonic=='mov'
                and x.op_str.startswith('word ptr es:[0x198], ')]
        graph.require(len(stores)==1,'unique SDR vector offset store')
        i,x=stores[0]
        graph.require(x.op_str=='word ptr es:[0x198], ax','SDR vector offset source')
        graph.require([q.op_str for q in prefix[i-3:i]]==['ax, 0','es, ax',prefix[i-1].op_str]
                      and prefix[i-1].mnemonic=='mov'
                      and prefix[i-1].op_str.startswith('ax, ')
                      and prefix[i-1].operands[1].type==dec.x86.X86_OP_IMM,
                      'SDR vector ES/offset producer')
        graph.require(prefix[i+1].op_str=='ax, cs' and prefix[i+2].op_str=='word ptr es:[0x19a], ax',
                      'SDR vector segment producer')
        rows.append(dict(module=name,entry=entry,vector_segment=0,vector_offset=0x198,
                         target_offset=prefix[i-1].operands[1].imm,
                         target_segment='resident SDR CS',
                         definitions=[q.address for q in prefix[i-3:i+3]],
                         status='INSTALLER DERIVED; launch/binding completeness UNKNOWN'))
    return rows


def task_writer_evidence(r, at):
    """Inventory candidate stores, retaining unsupported aliases as obligations."""
    direct_api=[];resets=[];other=[]
    for p,x in sorted(r.seen.items()):
        ops=x.operands
        if x.mnemonic!='mov' or len(ops)!=2 or ops[0].type!=r.x86.X86_OP_MEM:
            continue
        m=ops[0].mem
        if m.base!=r.x86.X86_REG_BX or m.index or m.disp or m.segment:
            continue
        v,defs,errors=r.reaching(p,'bx')
        record=dict(source=p,operand_class='indexed memory writer',
                    base_values=[list(a) if isinstance(a,tuple) else a for a in sorted(v,key=str)],
                    base_definitions=defs,base_unknown=errors)
        if ops[1].type==r.x86.X86_OP_REG:
            reg=x.reg_name(ops[1].reg);record['value_register']=reg
            if reg=='dx':direct_api.append(record)
            elif reg=='ax':resets.append(record)
            else:other.append(record)
        else:
            other.append(record)
    return dict(consumer=at,task_api_writer_candidates=direct_api,
                reset_writer_candidates=resets,other_bx_writer_candidates=other,
                unresolved='slot extents, current-slot aliases, all writer admission and initialization '
                           'must be proved before this inventory is complete')


def extend_evidence(r, result, data):
    installers=api_installers(data)
    result['api_vector_installers']=installers
    result['related_code2']=related_code2(r,result['domains'])
    import audit_10min_demo_writers as writers
    pointers=sorted({t['pointer_source'] for f in result['domains']
                     if f.get('domain_category')=='relocated module pointer' for t in f['tables']})
    tables={t['start']:t['start']+t['entries']*t['stride']
            for f in result['related_code2']['domains'] for t in f['tables']}
    objects=[dict(name='far-pointer-'+hex(p),start=p,end_exclusive=p+4) for p in pointers]
    objects += [dict(name='glyph-table-'+hex(p),start=p,end_exclusive=tables[p]) for p in sorted(tables)]
    import audit_10min_demo_placement as placement
    contract=placement.derive(r,(data/'PINBALL.EXE').read_bytes())
    result['placement_contract']=contract
    result['object_writer_evidence']=[dict(module='TABLE1',**writers.scan(r,objects,contract,module="TABLE1"))]
    if hasattr(r,'related_resolver'):
        sub=r.related_resolver
        result['object_writer_evidence'].append(dict(module='CODE2',**writers.scan(sub,objects,contract,module="CODE2",stack_parent=r.stack_context)))
    # Cross-module effects retain CODE2 identity. Conditional summaries never
    # discharge the relocated pointer object's runtime admission obligation.
    if hasattr(r,'related_resolver'):
        import audit_10min_demo_stack as stack
        import audit_10min_demo_entry as entry
        code2_effects=result['object_writer_evidence'][1]['placement_effects']
        local=code2_effects['conditional_module_stack_summary']
        provenance=entry.analyze(r,contract,r.stack_context['summaries'],objects,
            local['maximum_downward_bytes'] if local['status']=='BOUNDED' else None)
        code2_effects['entry_stack_provenance']=provenance
        import audit_10min_demo_admission as admission
        result['callback_and_interrupt_admission']=admission.analyze(r,data,installers,provenance)
        for caller in provenance['entry_callers']:
            for handler in caller['admitted_interrupts']:
                key='int66' if handler['vector']==0x66 else 'int33' if handler['vector']==0x33 else None
                if key:
                    handler['binding_evidence_reference']='callback_and_interrupt_admission.'+key
                    handler['unresolved_reason'].append(result['callback_and_interrupt_admission'][key]['unresolved_reason'])
        for boundary in code2_effects['boundaries']:
            boundary['entry_stack_handler_status']=provenance['status']
            boundary['lookup_table_runtime_immutability']='UNKNOWN: upstream memory/admission certificate and required entry/handler contracts remain open'
        effects=result['object_writer_evidence'][0]['placement_effects']
        for i,boundary in enumerate(effects['boundaries']):
            f=r.facts.get(boundary['site'],{})
            modules=f.get('module_targets',[])
            if modules and all(t['module']=='CODE2' for t in modules):
                effects['boundaries'][i]=stack.transfer(r,boundary['site'],
                    sub.stack_context['summaries'],target_domain=[t['target_file'] for t in modules])
                effects['boundaries'][i]['module']='TABLE1'
                effects['boundaries'][i]['typed_target_domain']=modules
    load,relocs=relocations(r.b)
    result['pointer_lifecycle']=[dict(object_file=p,offset_word=p,segment_word=p+2,
        loader_relocation_words=sorted(v for v in relocs if p<=v<p+4),
        loader_effect='MZ loader adds load segment to relocated segment word only',
        executable_initialization='UNKNOWN: writer admission/lifecycle not proved',
        runtime_offset_immutable=False,runtime_segment_immutable=False,
        status='UNKNOWN: loader relocation is not runtime mutation evidence') for p in pointers]
    for f in result['domains']:
        x=r.seen[f['source']]
        f['instruction']=x.mnemonic
        f.setdefault('operand_class', 'register' if x.operands[0].type==r.x86.X86_OP_REG
                     else 'memory near pointer')
        f.setdefault('domain_category',f['recipe'])
        f['target_module']='TABLE1' if f['domain']=='near' else 'see module_targets/binding_slot'
        f['completeness_reason']=('bounded under existing resolver grammar' if f['status']=='BOUNDED'
                                  else None)
        if f.get('binding_slot')==dict(segment=0,offset=0x198):
            f['binding_producers']=installers
        op=x.operands[0]
        if f['domain']=='near' and op.type==r.x86.X86_OP_MEM and not (
                op.mem.base or op.mem.index or op.mem.segment):
            disp=op.mem.disp
            f['field_evidence']=dict(ds_offset=disp,initial_value=r.word(r.ds+disp),
                writers=[dict(source=p,value=v,unknown=u) for p,v,u in r.producers[disp]],
                unresolved='initial-value exclusion requires initializer-before-admission proof; '
                           'indexed/segment-alias writers are not enumerated by direct-field scan')
        if f['source']==0x5eab:
            f['domain_category']='task-list consumer; not a writer'
            f['writer_evidence']=task_writer_evidence(r,f['source'])
        f['candidate_targets']=list(f['targets']) if f['status']=='UNKNOWN' else []
