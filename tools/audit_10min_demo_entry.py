#!/usr/bin/env python3
"""CODE2 caller-relative reaching definitions, independent of global stack gates.

Candidate control edges are retained; they do not prove callback scheduling or
pointer immutability. No external vector, root context or ABI is synthesized.
All output is semantic metadata, never executable slices.
"""
from collections import defaultdict, deque
from audit_10min_demo_stack import operation, targets, Summaries
from audit_10min_demo_control import relocations
from audit_10min_demo_placement import exclusion


class Admission:
    def __init__(self, r, contract, summaries, roots=None, handlers=None):
        self.r=r;self.contract=contract;self.summaries=summaries
        self.roots=roots or {};self.handlers=handlers or {};self.pred=defaultdict(set)
        self.load,self.relocs=relocations(r.b)
        for p,x in r.seen.items():
            if x.mnemonic in ('ret','retf','iret','hlt'):continue
            call=x.group(r.cs.CS_GRP_CALL);jump=x.group(r.cs.CS_GRP_JUMP)
            if call or jump:
                for t in targets(r,p)[0]:
                    if t in r.seen:self.pred[t].add((p,'call-entry' if call else 'branch'))
            if x.mnemonic not in ('jmp','ljmp') and p+x.size in r.seen:
                self.pred[p+x.size].add((p,'call-return' if call else 'next'))

    def slice(self, caller):
        todo=[caller];seen=set()
        while todo:
            p=todo.pop()
            if p in seen:continue
            seen.add(p);todo.extend(q for q,_ in self.pred[p])
        return seen

    def handler(self, site):
        # Bindings must be derived by the caller from installed vector admission.
        # An arbitrary list of candidate offsets is explicitly insufficient.
        binding=self.handlers.get(site)
        result=dict(site=site,vector=self.r.seen[site].operands[0].imm,
            vector_address=4*self.r.seen[site].operands[0].imm,
            immediate_frame=dict(bytes=6,order=['FLAGS','CS','IP'],relative_interval=[-6,0]),
            identities=[],admitted_target_set=None,target_set_complete=False,summaries=[],status='UNKNOWN',effects={n:'UNKNOWN' for n in ('ss','sp','ds','es','memory','placement')},
            unresolved_reason=[])
        if not binding or not binding.get('complete') or not binding.get('targets'):
            result['unresolved_reason']=['installed vector target admission not derived; external OS/BIOS implementation not pinned']
            if result['vector']==0x66:result['unresolved_reason'].append('INT66 binding/admission incomplete; callback cadence is not a premise')
            return result
        for target in binding['targets']:
            result['identities'].append(dict(module=target['summaries'].module,entry=target['entry']))
            original=target['summaries']
            # Handler placement restrictions must follow nested direct calls.
            handler_summaries=Summaries(original.r,[w for rows in original.writers.values() for w in rows],
                original.module,stack_only=original.stack_only,handler_only=True)
            result['summaries'].append(handler_summaries.routine(target['entry'],far='interrupt'))
        result.update(admitted_target_set=result['identities'],target_set_complete=True)
        rows=result['summaries']
        if any(v.get('proof_scope')=='stack only; cannot authorize memory/transfer closure' or v['status']!='BOUNDED' or v['conditional_boundaries'] or v['effects']['ss']!='preserving'
               or v['effects']['sp']!='preserving' or v['effects']['ds']=='UNKNOWN' or v['effects']['es']=='UNKNOWN' for v in rows):
            result['unresolved_reason']=['handler stack/segment/return or target-memory effects incomplete']
            return result
        result.update(status='BOUNDED',effects={n:'preserving' if all(v['effects'][n]=='preserving' for v in rows) else 'bounded finite effect' for n in ('ss','sp','ds','es','placement')},
            maximum_downward_bytes=max(v['maximum_downward_bytes'] for v in rows),
            return_contract='matched IRET consumes original six-byte frame after balanced local stack',
            memory_contract='all explicit destination placement certificates exclude scoped tables')
        result['effects']['memory']='bounded finite effect'
        return result

    def reaching(self, caller, reg):
        todo=deque([(caller,reg,0)]);seen=set();values=set();defs=set();errors=set();interrupts=set();calls=set()
        def value(kind,v,delta):
            if reg=='sp' and kind!='literal':errors.add('SP is not a finite numeric offset');return
            values.add((kind,(v+delta)&65535 if kind=='literal' else v+delta))
        while todo:
            p,n,delta=todo.popleft()
            if (p,n,delta) in seen:continue
            seen.add((p,n,delta))
            if len(seen)>4096:
                errors.add('reaching-definition state bound; unbalanced cycle or incomplete slice');break
            if p==self.contract['entry_file']:
                if n=='ss':value('image',self.load+16*self.contract['initial_ss_relative'],delta);defs.add(p)
                elif n=='sp':value('literal',self.contract['initial_sp'],delta);defs.add(p)
                else:errors.add('MZ entry register '+n+' not derived')
            if p in self.roots:
                vals=self.roots[p].get(n)
                if vals is None:errors.add('admitted root '+n+' unknown at '+hex(p))
                else:
                    for kind,v in vals:value(kind,v,delta)
                    defs.add(p)
            parents=self.pred[p]
            if not parents and p!=self.contract['entry_file'] and p not in self.roots:
                errors.add('external/root '+n+' admission unknown at '+hex(p))
            for q,edge in parents:
                x=self.r.seen[q];ops=x.operands;m=x.mnemonic
                if edge=='call-entry':
                    todo.append((q,n,(delta+operation(self.r,x)['delta'])&65535 if n=='sp' else delta));continue
                if edge=='call-return':
                    calls.add(q);ts,complete=targets(self.r,q)
                    rows=[self.summaries.routine(t,m=='lcall') for t in ts]
                    if not complete or not rows or any(v['status']!='BOUNDED' or v['conditional_boundaries'] for v in rows):
                        errors.add('return effect on '+n+' incomplete at '+hex(q));continue
                    if n=='sp':todo.append((q,n,delta));continue
                    for row in rows:
                        output=row['register_outputs'].get(n,['unknown'])
                        if output==['input',n]:todo.append((q,n,delta))
                        elif output[0] in ('image','literal'):value(*output,delta);defs.add(q)
                        else:errors.add('callee output '+n+' unknown at '+hex(q))
                    continue
                if m=='int':
                    interrupts.add(q);h=self.handler(q)
                    if h['status']!='BOUNDED' or h['effects'].get(n)!='preserving':
                        errors.add('interrupt '+n+' effect unknown at '+hex(q));continue
                    todo.append((q,n,delta));continue
                names={x.reg_name(v) for v in x.regs_access()[1]}
                assigned=ops and ops[0].type==self.r.x86.X86_OP_REG and x.reg_name(ops[0].reg)==n
                if m=='mov' and assigned:
                    defs.add(q);src=ops[1]
                    if src.type==self.r.x86.X86_OP_IMM:
                        relocated=q+x.encoding.imm_offset in self.relocs
                        value('image' if relocated else 'literal',self.load+16*(src.imm&65535) if relocated else src.imm&65535,delta)
                    elif src.type==self.r.x86.X86_OP_REG:todo.append((q,x.reg_name(src.reg),delta))
                    else:errors.add('memory definition of '+n+' unknown at '+hex(q))
                    continue
                if n=='sp' and m in ('add','sub') and assigned and ops[1].type==self.r.x86.X86_OP_IMM:
                    defs.add(q);todo.append((q,n,(delta+(1 if m=='add' else -1)*ops[1].imm)&65535));continue
                effect=operation(self.r,x)
                if n=='sp' and effect and m not in ('enter','leave') and not (m=='pop' and assigned):
                    todo.append((q,n,(delta+effect['delta'])&65535));continue
                if n in names or (n in ('ax','bx','cx','dx') and names&{n[0]+'l',n[0]+'h'}):
                    defs.add(q);errors.add('unsupported '+n+' definition at '+hex(q));continue
                todo.append((q,n,delta))
        if not values and not errors:errors.add('definition cycle without admitted producer')
        return dict(values=sorted(values),definitions=sorted(defs),unknown=sorted(errors),
            admitted_interrupts=sorted(interrupts),return_effect_dependencies=sorted(calls),
            status='UNKNOWN' if errors else 'BOUNDED')

    def caller(self, p, objects, local_bound):
        r=self.r;ss=self.reaching(p,'ss');sp=self.reaching(p,'sp');nodes=self.slice(p)
        dependencies=[]
        for q in sorted(set(ss['return_effect_dependencies'])|set(sp['return_effect_dependencies'])):
            ts,complete=targets(r,q)
            rows=[self.summaries.routine(t,r.seen[q].mnemonic=='lcall') for t in ts]
            dependencies.append(dict(site=q,target_admission_complete=complete,target_set=ts,summaries=rows))
        # Include only handler frontiers actually encountered in these matched
        # return-effect dependencies, not all interrupts in TABLE1.
        dependency_ints=set();todo=[row for dep in dependencies for row in dep['summaries']];visited=set();nested=[]
        while todo:
            row=todo.pop();key=(row['entry'],row.get('return_kind'))
            if key in visited:continue
            visited.add(key);nested.append(row)
            for q in row.get('sites',[]):
                if r.seen[q].mnemonic=='int':dependency_ints.add(q)
                if r.seen[q].group(r.cs.CS_GRP_CALL):
                    for t in targets(r,q)[0]:
                        child=self.summaries.cache.get((t,r.seen[q].mnemonic=='lcall'))
                        if child is not None:todo.append(child)
        interrupts=sorted({q for q in nodes if r.seen[q].mnemonic=='int'} | set(ss['admitted_interrupts']) | set(sp['admitted_interrupts']) | dependency_ints)
        frame=operation(r,r.seen[p]);width=frame['width'];entry_sp=None if sp['unknown'] else sorted({(v-width)&65535 for kind,v in sp['values']})
        offsets=None if entry_sp is None or local_bound is None else {(v-i)&65535 for v in entry_sp for i in range(1,local_bound+1)} | {(v+i)&65535 for v in entry_sp for i in range(width)}
        segment=dict(values=ss['values'],unknown=ss['unknown'])
        proofs=[exclusion(self.contract,segment,offsets,1,o) for o in objects]
        handlers=[self.handler(q) for q in interrupts]
        for h in handlers:
            incoming_ss=self.reaching(h['site'],'ss');incoming_sp=self.reaching(h['site'],'sp')
            h['incoming_ss']=incoming_ss;h['incoming_sp']=incoming_sp
            starts=None if incoming_sp['unknown'] else {(v-6)&65535 for _,v in incoming_sp['values']}
            h['immediate_frame']['physical_proofs']=[exclusion(self.contract,incoming_ss,starts,6,o) for o in objects]
            extent=h.get('maximum_downward_bytes')
            writes=None if starts is None or extent is None else {(v-i)&65535 for v in starts for i in range(1,extent+1)}
            h['local_stack_physical_proofs']=[exclusion(self.contract,incoming_ss,writes,1,o) for o in objects]
            h['physical_stack_complete']=h['status']=='BOUNDED' and all(v['status']=='EXCLUDED' for v in h['immediate_frame']['physical_proofs']+h['local_stack_physical_proofs'])
        errors=ss['unknown']+sp['unknown']
        if local_bound is None:errors.append('local CODE2 bound incomplete')
        if any(v['status']!='EXCLUDED' for v in proofs):errors.append('complete CODE2 physical stack domain may overlap tables')
        if any(not h['physical_stack_complete'] for h in handlers):errors.append('required interrupt handler/physical stack effects unresolved')
        assignments=[]
        for q in sorted(nodes|{q for row in nested for q in row.get('sites',[])}):
            x=r.seen[q];effect=operation(r,x)
            names={x.reg_name(v) for v in x.regs_access()[1]}&{'ss','sp','esp'}
            if names and (effect is None or x.mnemonic in ('enter','leave') or
                    (x.mnemonic=='pop' and x.op_str in ('ss','sp','esp'))):
                assignments.append(dict(site=q,instruction=x.mnemonic,writes=sorted(names)))
        return dict(site=p,target_module='CODE2',target=next((t['target_file'] for t in r.facts.get(p,{}).get('module_targets',[])),None),
            backward_provenance_slice=sorted(nodes),slice_roots=sorted(q for q in nodes if not self.pred[q]),
            slice_root_admissions=[dict(site=q,ss=self.roots.get(q,{}).get('ss'),sp=self.roots.get(q,{}).get('sp')) for q in sorted(nodes) if not self.pred[q] and q!=self.contract['entry_file']],
            reachable_stack_assignments=assignments,
            matched_return_effect_dependencies=dependencies,
            transitive_stack_effect_dependencies=nested,
            ss_definitions=ss,sp_definitions=sp,ss=None if ss['unknown'] else ss['values'],sp=None if sp['unknown'] else [v for _,v in sp['values']],
            far_call_frame=dict(bytes=width,relative_interval=[-width,0]),far_call_bytes=width,
            code2_entry_ss=None if ss['unknown'] else ss['values'],code2_entry_sp=entry_sp,
            local_code2_stack_bound=local_bound,physical_stack_ranges=proofs,
            physical_stack_destination_domain={'a20_disabled':[[0,1<<20]],'a20_enabled':[[0,(1<<20)+65520]]} if ss['unknown'] else [v['destination_domains'] for v in proofs[:1]],
            physical_stack_calculation='16*SS + u16([caller_SP - far_frame_bytes - local_bound, caller_SP)); both A20 states, all legal B; UNKNOWN inputs retain full domains',
            target_table_physical_ranges=[dict(name=v['target'],ranges=v['target_physical_domain']) for v in proofs],
            admitted_interrupts=handlers,overlap_result='DISJOINT' if all(v['status']=='EXCLUDED' for v in proofs) else 'UNKNOWN',
            status='UNKNOWN' if errors else 'BOUNDED',completeness_reason=None if errors else 'all reaching entry contexts and required handlers bounded; shared placement proof excludes complete local stack plus far frame',
            unresolved_reason=sorted(set(errors)),global_table1_stack_completeness_required=False,
            scope='typed candidate entry projection; does not authorize TABLE1 far-pointer promotion')


def analyze(r, contract, summaries, objects, local_bound, roots=None, handlers=None):
    # A callee's unrelated explicit store cannot make its SS/SP effect top.
    # Its memory effects remain a separate immutability obligation.
    stack_summaries=Summaries(r,[],'TABLE1',stack_only=True)
    admission=Admission(r,contract,stack_summaries,roots,handlers)
    tables=[o for o in objects if o['name'].startswith('glyph-table-')]
    callers=sorted(p for p,f in r.facts.items() if any(t['module']=='CODE2' for t in f.get('module_targets',[])))
    rows=[admission.caller(p,tables,local_bound) for p in callers]
    startup=[];at=contract['entry_file']
    while at in r.seen and len(startup)<128:
        x=r.seen[at];startup.append(at)
        if x.mnemonic=='int' or x.group(r.cs.CS_GRP_CALL) or x.group(r.cs.CS_GRP_JUMP) or x.group(r.cs.CS_GRP_RET):break
        at+=x.size
    frontier=admission.handler(at) if at in r.seen and r.seen[at].mnemonic=='int' else None
    if frontier:
        si=admission.reaching(at,'ss');pi=admission.reaching(at,'sp')
        frontier['incoming_ss']=si;frontier['incoming_sp']=pi
        starts=None if pi['unknown'] else {(v-6)&65535 for _,v in pi['values']}
        frontier['immediate_frame']['physical_proofs']=[exclusion(contract,si,starts,6,o) for o in tables]
    return dict(entry_callers=rows,startup_provenance_prefix=startup,startup_frontier=frontier,
        initial_mz=dict(entry=contract['entry_file'],ss=['image',contract['header_bytes']+16*contract['initial_ss_relative']],sp=contract['initial_sp'],
            dependence='SS=B+header SS modulo 65536; SP is header literal; relocation identities retained'),
        status='BOUNDED' if rows and all(row['status']=='BOUNDED' for row in rows) else 'UNKNOWN',
        completeness_scope='CODE2 entry only; unrelated TABLE1 stack records are not premises')
