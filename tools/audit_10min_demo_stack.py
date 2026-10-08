#!/usr/bin/env python3
"""Scoped stack/transfer certificates for the existing typed candidate CFG.

Local callee effects and runtime target admission are separate certificates.
No interrupt handler is invented. An admitted INT invalidates its continuation
until its handler/ABI effects are supplied. No allowance is a stack-depth proof.
"""
from collections import defaultdict, deque
import json

from audit_10min_demo_placement import exclusion, SPACE


def operation(r, x):
    """Architectural bytes, relative to incoming SP; offset wrap is explicit."""
    m=x.mnemonic
    wide=0x66 in x.prefix
    unit=4 if wide else 2
    if m in ('push','pushf','pushfd'): delta=-unit; width=unit
    elif m in ('pushaw','pushal'): delta=-8*unit; width=8*unit
    elif m in ('call','lcall'): delta=-unit*(2 if m=='lcall' else 1); width=-delta
    elif m=='int': delta=-6; width=6
    elif m in ('pop','popf','popfd'): delta=unit; width=0
    elif m in ('popaw','popal'): delta=8*unit; width=0
    elif m in ('ret','retf','iret'):
        delta=unit*(3 if m=='iret' else 2 if m=='retf' else 1)
        delta+=x.operands[0].imm if x.operands else 0
        width=0
    elif m in ('enter','leave'): return dict(effect_class=m,delta=None,width=None,
        unknown=['frame/stack-pointer effect not summarized'])
    else:return None
    return dict(effect_class=m,delta=delta,width=width,
        write_relative=[delta,0] if width else [],
        growth='downward modulo 65536' if width else 'upward modulo 65536',unknown=[])


def targets(r, at):
    x=r.seen[at]
    f=r.facts.get(at,{})
    if x.mnemonic not in ('lcall','ljmp') and x.operands and x.operands[0].type==r.x86.X86_OP_IMM:
        return [r.base+x.operands[0].imm],True
    return list(f.get('targets',[])),f.get('status')=='BOUNDED'


class Summaries:
    """Finite path stack-save analysis; calls use matched callee summaries.

    Repeated (site, stack) states close balanced loops. A different stack at a
    cycle is retained until a resource guard returns UNKNOWN, never a depth
    certificate. Recursive call SCCs are UNKNOWN without a bounded admission.
    """
    def __init__(self,r,writers,module,stack_only=False,handler_only=False):
        self.r=r;self.module=module;self.writers=defaultdict(list);self.cache={}
        self.stack_only=stack_only
        self.handler_only=handler_only
        for w in writers:self.writers[w['source']].append(w)

    def routine(self, entry, far=False, active=()):
        key=(entry,far)
        if key in self.cache:return self.cache[key]
        if key in active:return self.open(entry,'recursive call SCC has no bounded admission')
        r=self.r
        registers=('ss','ds','es','ax','bx','cx','dx','bp','si','di')
        initial=tuple(('input',n) for n in registers)
        todo=[(entry,(),initial)];visited=set();sites=set();errors=set();returns=set()
        maximum=0;writes=set();conditional=set();stack_sites={};ss_assignments=set()
        while todo:
            at,stack,values=todo.pop();state=(at,stack,values)
            regs=dict(zip(registers,values));ss=regs["ss"]
            if state in visited:continue
            visited.add(state)
            if len(visited)>8192 or len(stack)>128:
                errors.add('stack state exploration incomplete (cycle/depth/resource guard)');break
            x=r.seen.get(at)
            if x is None:errors.add('callee leaves admitted typed instruction domain');continue
            sites.add(at);ops=x.operands;m=x.mnemonic
            effect=operation(r,x)
            if (self.handler_only or self.stack_only) and effect and ss!=('input','ss'):
                errors.add('stack access after SS substitution has no matched physical-frame proof');continue
            if r.cs.CS_GRP_RET in x.groups or m=='iret':
                expected='iret' if far=='interrupt' else 'retf' if far else 'ret'
                if stack or m!=expected or ops:errors.add('unmatched/adjusted return or return kind')
                else:returns.add(values)
                continue
            if m in ('int','iret','enter','leave','hlt'):
                errors.add('unmodelled '+m+' control/stack/handler effects');continue
            if (far=='interrupt' or self.handler_only) and (m in ('out','outsb','outsw','outsd','lgdt','lidt','lmsw','wrmsr')
                    or any(op.type==r.x86.X86_OP_REG and x.reg_name(op.reg).startswith(('cr','dr')) for op in ops)):
                errors.add('handler machine/placement effect not summarized');continue
            depth=sum(v[0] for v in stack)
            for op in ops:
                if not self.stack_only and op.type==r.x86.X86_OP_MEM and not op.access:
                    errors.add('memory operand effect metadata incomplete at '+hex(at))
            for w in self.writers.get(at,[]):
                writes.update(w['possible_objects'])
                if not self.stack_only and w['verdict']!='EXCLUDED':errors.add('possible target-object explicit memory write')
            if effect and effect['width']:
                stack_sites.setdefault(at,set()).add(-depth)
                maximum=max(maximum,depth+effect['width'])
            if m in ('push','pushf','pushfd','pushaw','pushal'):
                token=None
                if m=='push' and ops[0].type==r.x86.X86_OP_REG:token=regs.get(x.reg_name(ops[0].reg))
                elif m=='push' and ops[0].type==r.x86.X86_OP_IMM:
                    from audit_10min_demo_control import relocations
                    load,relocs=relocations(r.b)
                    token=('image',load+16*(ops[0].imm&65535)) if at+x.encoding.imm_offset in relocs else ('literal',ops[0].imm&65535)
                stack=stack+((effect['width'],token),)
            elif m in ('pop','popf','popfd','popaw','popal'):
                count=effect['delta'];popped=[]
                while stack and count>0:
                    n,token=stack[-1];stack=stack[:-1];count-=n;popped.append(token)
                if count!=0:errors.add('unmatched stack pop');continue
                if m=='pop' and ops[0].type==r.x86.X86_OP_REG:
                    name=x.reg_name(ops[0].reg)
                    if name in ('sp','esp'):errors.add('stack-pointer pop effect unknown');continue
                    if name in regs:regs[name]=popped[0] or ('unknown',)
                    if name=='ss':ss_assignments.add(at)
            elif x.group(r.cs.CS_GRP_CALL):
                ts,admitted=targets(r,at)
                if not ts:errors.add('call target/effects unknown');continue
                if not admitted:conditional.add(at)
                callees=[self.routine(t,m=='lcall',active+(key,)) for t in ts]
                if any(c['status']=='UNKNOWN' for c in callees):
                    errors.add('callee stack/memory/control effects unknown at '+hex(at));continue
                if any(c['effects']['ss']!='preserving' for c in callees):
                    errors.add('callee SS substitution not derived in local frame');continue
                maximum=max(maximum,depth+effect['width']+max(c['maximum_downward_bytes'] for c in callees))
                for c in callees:conditional.update(c['conditional_boundaries'])
                for n in registers:
                    outputs={tuple(c['register_outputs'].get(n,['unknown'])) for c in callees}
                    value=next(iter(outputs)) if len(outputs)==1 else ('unknown',)
                    regs[n]=regs[n] if value==('input',n) else value
            else:
                _,access=x.regs_access();names={x.reg_name(v) for v in access}
                if names&{'sp','esp'}:
                    errors.add('SP definition/adjustment unsupported at '+hex(at));continue
                if m=='mov' and ops[0].type==r.x86.X86_OP_REG:
                    name=x.reg_name(ops[0].reg)
                    if name in regs:
                        source=ops[1]
                        if source.type==r.x86.X86_OP_REG:regs[name]=regs.get(x.reg_name(source.reg),('unknown',))
                        elif source.type==r.x86.X86_OP_IMM:
                            from audit_10min_demo_control import relocations
                            load,relocs=relocations(r.b)
                            regs[name]=('image',load+16*(source.imm&65535)) if at+x.encoding.imm_offset in relocs else ('literal',source.imm&65535)
                        else:regs[name]=('unknown',)
                        if name=='ss':ss_assignments.add(at)
                else:
                    for n in names:
                        if n in regs:regs[n]=('unknown',)
                        if n in ('al','ah','bl','bh','cl','ch','dl','dh'):regs[n[0]+'x']=('unknown',)
                if regs['ss']!=('input','ss') and regs['ss'][0] not in ('literal','image'):
                    errors.add('SS definition unsupported at '+hex(at));continue
            nexts=[]
            if x.group(r.cs.CS_GRP_JUMP):
                ts,admitted=targets(r,at)
                if not ts:errors.add('branch domain unknown');continue
                if not admitted:conditional.add(at)
                nexts.extend(ts)
            if m not in ('jmp','ljmp'):nexts.append(at+x.size)
            todo.extend((p,stack,tuple(regs[n] for n in registers)) for p in nexts)
        if not returns:errors.add('normal matched return not proved')
        outputs={n:{v[i] for v in returns} for i,n in enumerate(registers)}
        if not outputs['ss'] or any(v!=('input','ss') and v[0] not in ('literal','image') for v in outputs['ss']):errors.add('return SS domain unknown')
        ss_effect='preserving' if outputs['ss']=={('input','ss')} else 'bounded finite effect' if outputs['ss'] and ('unknown',) not in outputs['ss'] else 'UNKNOWN'
        result=dict(module=self.module,entry=entry,return_kind='interrupt' if far=='interrupt' else 'far' if far else 'near',
            proof_scope='stack only; cannot authorize memory/transfer closure' if self.stack_only else 'stack and scoped explicit memory',
            status='UNKNOWN' if errors else 'BOUNDED',instruction_count=len(sites),sites=sorted(sites),
            effects=dict(ds='UNKNOWN',es='UNKNOWN',ss=ss_effect,sp='preserving' if not errors else 'UNKNOWN',
                memory_writes='possible target-object write' if writes else 'explicit target-object writes excluded; stack separate',
                allocation='preserving' if not errors else 'UNKNOWN',placement='preserving' if not errors else 'UNKNOWN',
                control_return='matched normal return' if not errors else 'UNKNOWN'),
            ss_output=[list(v) for v in sorted(outputs['ss'],key=str)],ss_assignments=sorted(ss_assignments),
            register_outputs={n:list(next(iter(v))) if len(v)==1 else ['unknown'] for n,v in outputs.items()},
            maximum_downward_bytes=maximum if not errors else None,
            stack_sites=[dict(site=p,incoming_sp_relative=sorted(v)) for p,v in sorted(stack_sites.items())],
            conditional_boundaries=sorted(conditional),possible_objects=sorted(writes),
            completeness_reason='closed direct paths and matched stack/callee effects within enumerated candidate domain' if not errors else None,
            unresolved_reason=sorted(errors))
        if not errors and not self.stack_only:
            for reg in ('ds','es'):
                preserved=(outputs[reg]=={('input',reg)}) if far=='interrupt' else r.preserves(entry,reg)
                result['effects'][reg]='preserving' if preserved else 'bounded finite effect' if outputs[reg] and all(v[0] in ('literal','image') for v in outputs[reg]) else 'UNKNOWN'
        if self.stack_only:result['effects']['memory_writes']='UNKNOWN: excluded from stack-only summary'
        self.cache[key]=result
        return result

    def open(self,entry,reason):
        return dict(module=self.module,entry=entry,status='UNKNOWN',maximum_downward_bytes=None,
            conditional_boundaries=[],effects={n:'UNKNOWN' for n in ('ds','es','ss','sp','memory_writes','allocation','placement','control_return')},
            unresolved_reason=[reason])


def transfer(r,at,summaries,target_domain=None):
    ts,complete=targets(r,at);x=r.seen[at]
    if target_domain is not None:ts=list(target_domain)
    callees=[summaries.routine(t,x.mnemonic=='lcall') for t in ts]
    errors=[]
    if summaries.stack_only:errors.append('stack-only summary cannot authorize transfer/memory closure')
    if not complete:errors.append('runtime target admission/immutability not proved')
    if not callees or any(c['status']=='UNKNOWN' for c in callees):errors.append('admitted callee effects incomplete')
    if any(c.get('conditional_boundaries') for c in callees):errors.append('nested target admission/immutability not proved')
    effects={n:'UNKNOWN' for n in ('ds','es','ss','sp','memory_writes','allocation','placement','control_return')}
    if callees and all(c['status']=='BOUNDED' for c in callees):
        for n in effects:
            vals={c['effects'][n] for c in callees}
            effects[n]=next(iter(vals)) if len(vals)==1 else 'bounded finite effect'
    if any(effects[n]=='UNKNOWN' for n in ('ds','es','ss','sp','allocation','placement','control_return')):
        errors.append('callee segment/stack/lifecycle effect domain incomplete')
    local=bool(callees) and all(c['status']=='BOUNDED' for c in callees)
    candidate_classification=('preserving' if all(effects[n]=='preserving' for n in ('ds','es','ss','sp')) and all(not summaries.writers.get(p) for c in callees for p in c.get('sites',[])) else 'bounded finite effect') if local else 'UNKNOWN'
    classification='UNKNOWN' if errors else candidate_classification
    memory=[dict(site=p,operand=w['operand'],width=w['width'],segment=w['segment'],
        placement_proofs=w['placement_proofs'],possible_objects=w['possible_objects'])
        for p in sorted({p for c in callees for p in c.get('sites',[])})
        for w in summaries.writers.get(p,[])]
    return dict(module=summaries.module,site=at,source=at,instruction=x.mnemonic+' '+x.op_str,
        effect_class=classification,classification=classification,status='UNKNOWN' if errors else 'BOUNDED',
        candidate_targets=ts,target_module=summaries.module,admitted_target_domain_complete=complete,
        incoming_domains={n:'symbolic caller input' for n in ('ds','es','ss','sp')},effects=effects,
        physical_write_domain='callee writer placement certificates; implicit stack needs caller SS:SP',
        overlap_decision='UNKNOWN' if errors else 'explicit writes excluded; stack obligation separate',
        candidate_effects_complete=local,candidate_effect_class=candidate_classification,callee_summaries=callees,
        explicit_memory_effects=memory,
        maximum_callee_downward_bytes=max(c['maximum_downward_bytes'] for c in callees) if local else None,
        completeness_reason='every admitted callee summarized' if not errors else None,unresolved_reason=errors)


def merge_state(old,new):
    if old is None:return new
    result=[]
    for a,b in zip(old,new):
        result.append(None if a is None or b is None else a|b)
    # State resource widening never authorizes a bounded range.
    if result[1] is not None and len(result[1])>256:result[1]=None
    return tuple(result)


def contexts(r,contract,summaries,entry_contexts=None,candidate_effects=False):
    """Incoming SS/SP fixed point across instruction edges and matched calls.

    Unknown callback/root contexts and INT continuations remain top. Offset sets
    are finite modulo 65536; widening to top only loses exclusions.
    """
    states={};queue=deque();reasons=defaultdict(set)
    initial=(frozenset({('image',contract['header_bytes']+16*contract['initial_ss_relative'])}),frozenset({contract['initial_sp']}))
    def put(p,state,reason=None):
        if p not in r.seen:return
        if reason:reasons[p].add(reason)
        v=merge_state(states.get(p),state)
        if states.get(p)!=v:states[p]=v;queue.append(p)
    if entry_contexts is not None:
        for p,state in entry_contexts.items():put(p,state,'typed module caller context')
    elif contract['entry_file'] in r.seen:
        put(contract['entry_file'],initial,'MZ initial SS:SP')
    # Roots without a typed incoming edge do not acquire a startup context.
    incoming=set()
    for p,x in r.seen.items():
        if x.group(r.cs.CS_GRP_CALL) or x.group(r.cs.CS_GRP_JUMP):incoming.update(targets(r,p)[0])
        if x.mnemonic not in ('ret','retf','iret','jmp','ljmp'):incoming.add(p+x.size)
    for p in r.roots:
        if p not in incoming and p not in states:put(p,(None,None),'entry SS:SP admission not derived')
    def advance(state,delta):
        return (state[0],None if state[1] is None else frozenset((p+delta)&65535 for p in state[1]))
    def drain():
        while queue:
            p=queue.popleft();state=states[p];x=r.seen[p];m=x.mnemonic;ops=x.operands
            effect=operation(r,x);out=state
            if m in ('ret','retf','iret'):continue
            if effect and effect['delta'] is not None:out=advance(state,effect['delta'])
            if x.group(r.cs.CS_GRP_CALL):
                ts,admitted=targets(r,p)
                for t in ts:put(t,out)
                cs=[summaries.routine(t,m=='lcall') for t in ts]
                if (admitted or candidate_effects) and cs and all(c['status']=='BOUNDED' and c['effects']['ss']=='preserving' and (candidate_effects or not c['conditional_boundaries']) for c in cs):out=state
                else:out=(None,None);reasons[p+x.size].add('call target/SS:SP return effects incomplete at '+hex(p))
            elif m=='int':
                out=(None,None);reasons[p+x.size].add('admitted INT handler/SS:SP effects unknown at '+hex(p))
            elif m in ('enter','leave') or (m=='pop' and x.op_str in ('ss','sp','esp')):out=(None,None)
            elif effect is None:
                _,regs=x.regs_access();names={x.reg_name(v) for v in regs}
                if 'ss' in names:out=(None,out[1])
                if names&{'sp','esp'}:out=(out[0],None)
            if x.group(r.cs.CS_GRP_JUMP):
                for t in targets(r,p)[0]:put(t,out)
            if m not in ('jmp','ljmp'):put(p+x.size,out)
    drain()
    for p in r.roots:
        if p not in states:put(p,(None,None),'disconnected candidate/root context unknown')
    drain()
    return states,reasons


def stack_record(r,p,contract,objects,state,module,reasons=()):
    x=r.seen[p];effect=operation(r,x);ss,sp=state;errors=[]
    s=dict(values=[] if ss is None else sorted(ss),unknown=['incoming SS provenance unresolved'] if ss is None else [])
    offsets=None if sp is None or effect['delta'] is None else {(v+effect['delta'])&65535 for v in sp}
    proofs=[exclusion(contract,s,offsets,effect['width'] or 1,o) for o in objects] if effect['width'] else []
    if ss is None:errors.append('incoming SS provenance unresolved')
    if sp is None:errors.append('incoming SP/depth unresolved')
    errors.extend(effect['unknown'])
    hits=[v['target'] for v in proofs if v['status']!='EXCLUDED'] if effect['width'] else [o['name'] for o in objects]
    # Full-offset disjointness is a valid proof even without a depth bound.
    instruction_closed=not effect['unknown'] and bool(proofs) and not hits
    closed=instruction_closed
    if x.mnemonic=='int':
        closed=False;hits=[o['name'] for o in objects]
        errors.append('admitted interrupt handler SS/SP/depth/memory effects unresolved')
    ts,complete=targets(r,p) if x.group(r.cs.CS_GRP_CALL) else ([],True)
    return dict(module=module,site=p,source=p,instruction=x.mnemonic+' '+x.op_str,
        effect_class=effect['effect_class'],bytes_written=effect['width'],stack_delta=effect['delta'],growth=effect.get('growth'),
        incoming_domains=dict(ss=s,sp=None if sp is None else sorted(sp)),
        sp_derivation='instruction fixed point with matched call/return summaries; unknown states widened to full offsets',
        write_offsets=None if offsets is None else sorted(offsets),write_relative=effect.get('write_relative'),
        physical_write_domain=proofs,
        physical_destination_domain={'a20_disabled':[[0,SPACE]],'a20_enabled':[[0,SPACE+65536]]} if ss is None else [v['destination_domains'] for v in proofs[:1]],
        architectural_write_status='BOUNDED' if instruction_closed else 'UNKNOWN',
        handler_physical_write_domain='UNKNOWN' if x.mnemonic=='int' else None,
        callee_target_domain=ts,callee_module_domain=r.facts.get(p,{}).get('module_targets',[]),
        callee_admission_complete=complete,possible_objects=hits,overlap_decision='DISJOINT' if closed else 'UNKNOWN',
        status='BOUNDED' if closed else 'UNKNOWN',context_sources=list(reasons),
        completeness_reason='placement certificate excludes every written byte for all incoming contexts' if closed else None,
        unresolved_reason=[] if closed else sorted(set(errors+['instruction or unmodelled handler physical destination may overlap target objects'])))


def groups(records):
    grouped=defaultdict(list)
    for row in records:
        signature={n:row[n] for n in ('effect_class','bytes_written','stack_delta','incoming_domains','write_offsets','status','possible_objects','unresolved_reason')}
        grouped[json.dumps(signature,sort_keys=True)].append(row)
    return [dict(group=i,number_of_sites=len(rows),representative_sites=[r['site'] for r in rows[:6]],
        sites=[r['site'] for r in rows],effect_class=rows[0]['effect_class'],incoming_domains=rows[0]['incoming_domains'],
        sp_derivation=rows[0]['sp_derivation'],maximum_write_interval=rows[0]['physical_write_domain'][0]['offset_ranges'] if rows[0]['physical_write_domain'] else None,
        maximum_downward_extent='UNKNOWN' if rows[0]['incoming_domains']['sp'] is None else 'finite incoming SP set and architectural width',
        physical_write_domain=rows[0]['physical_write_domain'],physical_destination_domain=rows[0]['physical_destination_domain'],
        architectural_write_status=rows[0]['architectural_write_status'],handler_physical_write_domain=rows[0]['handler_physical_write_domain'],
        overlap_decision=rows[0]['overlap_decision'],
        status=rows[0]['status'],unresolved_reason=rows[0]['unresolved_reason']) for i,rows in enumerate(grouped.values())]


def analyze(r,objects,contract,writers,implicit,boundaries,module,parent=None):
    summaries=Summaries(r,writers,module)
    entry_contexts=None;caller_context=[]
    if parent is not None:
        states=parent['states'];entries=defaultdict(lambda:None)
        for p,f in parent['resolver'].facts.items():
            for target in f.get('module_targets',[]):
                if target['module']!=module:continue
                ss,sp=states.get(p,(None,None));op=operation(parent['resolver'],parent['resolver'].seen[p])
                incoming=(ss,None if sp is None else frozenset((v+op['delta'])&65535 for v in sp))
                entries[target['target_file']]=merge_state(entries[target['target_file']],incoming)
                caller_context.append(dict(module='TABLE1',site=p,target_module=module,target=target['target_file'],
                    ss=None if ss is None else sorted(ss),sp=None if sp is None else sorted(sp),far_call_bytes=op['width'],
                    unresolved_reason='upstream caller SS:SP admission/effects incomplete' if None in incoming else None))
        entry_contexts=dict(entries)
    states,reasons=contexts(r,contract,summaries,entry_contexts)
    records=[stack_record(r,p,contract,objects,states.get(p,(None,None)),module,sorted(reasons[p])) for p in implicit]
    transfers=[transfer(r,p,summaries) for p in boundaries]
    relative=None
    if entry_contexts is not None and len(entry_contexts)==1:
        entry=next(iter(entry_contexts))
        rel_states,_=contexts(r,contract,summaries,{entry:(frozenset({('symbolic',0)}),frozenset({0}))},candidate_effects=True)
        relative=summaries.routine(entry,far=True)
        for row in records:
            vals=rel_states.get(row['site'],(None,None))[1]
            row['conditional_incoming_sp_relative']=None if vals is None else sorted(v if v<32768 else v-65536 for v in vals)
            row['conditional_written_sp_relative']=[[v+row['stack_delta'],v] for v in row['conditional_incoming_sp_relative']] if row['conditional_incoming_sp_relative'] is not None else None
            row['conditional_stack_premise']='exact enumerated candidate calls; caller SS:SP and runtime admission remain unproved'
    effects=dict(implicit_sites=records,stack_groups=groups(records),boundaries=transfers,
        closed_stack_site_count=sum(v['status']=='BOUNDED' for v in records),
        excluded_architectural_write_count=sum(v['architectural_write_status']=='BOUNDED' for v in records),
        unresolved_stack_site_count=sum(v['status']=='UNKNOWN' for v in records),
        unresolved_transfer_count=sum(v['status']=='UNKNOWN' for v in transfers),
        caller_contexts=caller_context,conditional_module_stack_summary=relative,
        reusable_callee_summaries=list(summaries.cache.values()) if module=='CODE2' else [],
        stack_transitions=[dict(module=module,site=p,instruction=x.mnemonic+' '+x.op_str,**operation(r,x)) for p,x in sorted(r.seen.items()) if operation(r,x)],
        initial_stack_domain=dict(ss=['image',contract['header_bytes']+16*contract['initial_ss_relative']],sp=contract['initial_sp']),
        interrupts=[dict(module=module,site=p,instruction=x.mnemonic+' '+x.op_str,effect_class='interrupt',
            bytes_written=6,status='UNKNOWN',unresolved_reason='admitted INT handler memory/segment/stack/allocation effects not supplied')
            for p,x in sorted(r.seen.items()) if x.mnemonic=='int'],
        stack_direction='decrement SP modulo 65536 before stores',
        lifecycle='runtime admission distinct from conditional candidate callee summaries')
    return effects,dict(states=states,resolver=r,summaries=summaries)
