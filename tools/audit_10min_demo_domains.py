#!/usr/bin/env python3
"""Fail-closed bounded target-domain research, pinned pinbfan inputs only.

No x86 execution, payload export or production descriptor. The worklist follows
all direct conditional edges (an overapproximation, not a state reachability
proof), derives bounded producers/tables, adds their targets, and iterates to a
fixed point. Unsupported reaching definitions are UNKNOWN, never an empty set.
"""
import argparse
from collections import defaultdict
import json
import os
from pathlib import Path
import struct

import audit_10min_demo_graph as graph


class Resolver:
    def __init__(self, binary, base, data_base, roots, limit, decoder=None):
        self.b = binary
        self.base, self.ds, self.roots, self.limit = base, data_base, set(roots), limit
        self.dec = decoder or graph.Decoder()
        self.cs, self.x86 = self.dec.cs, self.dec.x86
        self.programs = {}
        self.handlers = {}
        self.command_targets = set()
        self.program_root_unknowns = []
        self.effects = []
        self.facts = {}
        self.rounds = []
        self.preserved = {}

    def preserves(self, entry, requested, active=()):
        """Bounded save/restore summary, not execution of register values.

        Track only ownership of the incoming register token and stack saves.
        Unsupported stack manipulation, recursive calls or an indirect tail
        prevent a proof. Conditional branches are both followed.
        """
        key=(entry,requested)
        if key in self.preserved:
            return self.preserved[key]
        if key in active or not self.base <= entry < self.limit:
            return False
        regs=('ax','cx','dx','bx','sp','bp','si','di')
        initial=tuple('input' if n==requested else None for n in regs)
        todo=[(entry,initial,())];seen=set();returned=False
        while todo:
            at,values,stack=todo.pop()
            state=(at,values,stack)
            if state in seen:continue
            seen.add(state)
            if len(seen)>4096 or len(stack)>32 or not self.base<=at<self.limit:
                self.preserved[key]=False;return False
            x=self.dec.instruction(self.b,self.base,at)
            v=dict(zip(regs,values));ops=x.operands
            if x.group(self.cs.CS_GRP_RET):
                if v[requested]!='input' or stack or ops:
                    self.preserved[key]=False;return False
                returned=True;continue
            if x.mnemonic in ('enter','leave','iret'):
                self.preserved[key]=False;return False
            if x.mnemonic in ('pushaw','pushal'):
                stack=stack+tuple(v[n] for n in regs)
            elif x.mnemonic in ('popaw','popal'):
                if len(stack)<8:self.preserved[key]=False;return False
                restored=dict(zip(regs,stack[-8:]));stack=stack[:-8]
                for n in regs:
                    if n!='sp':v[n]=restored[n]
            elif x.mnemonic=='push':
                value=v.get(x.reg_name(ops[0].reg)) if ops[0].type==self.x86.X86_OP_REG else None
                stack=stack+(value,)
            elif x.mnemonic=='pop':
                if not stack:self.preserved[key]=False;return False
                if ops[0].type==self.x86.X86_OP_REG:
                    name=x.reg_name(ops[0].reg)
                    if name in v:v[name]=stack[-1]
                stack=stack[:-1]
            elif x.group(self.cs.CS_GRP_CALL):
                for n in regs:
                    if v[n]=='input' and (ops[0].type!=self.x86.X86_OP_IMM
                            or not self.preserves(self.base+ops[0].imm,n,active+(key,))):
                        v[n]=None
            elif x.mnemonic=='mov' and ops[0].type==self.x86.X86_OP_REG:
                name=x.reg_name(ops[0].reg)
                name=name[1:] if name.startswith('e') else name
                if name in v:
                    if name=='sp':self.preserved[key]=False;return False
                    source=x.reg_name(ops[1].reg) if ops[1].type==self.x86.X86_OP_REG else ''
                    source=source[1:] if source.startswith('e') else source
                    v[name]=v.get(source)
                elif name in ('al','ah','bl','bh','cl','ch','dl','dh'):
                    v[name[0]+'x']=None
            elif x.mnemonic=='int':
                v={n:None for n in regs}
            else:
                unused,writes=x.regs_access()
                for r in writes:
                    name=x.reg_name(r)
                    name=name[1:] if name.startswith('e') else name
                    if name=='sp':self.preserved[key]=False;return False
                    if name in v:v[name]=None
                    elif name in ('al','ah','bl','bh','cl','ch','dl','dh'):v[name[0]+'x']=None
            successors=[]
            if x.group(self.cs.CS_GRP_JUMP):
                if ops[0].type!=self.x86.X86_OP_IMM:
                    self.preserved[key]=False;return False
                successors.append(self.base+ops[0].imm)
            if x.mnemonic not in ('jmp','ljmp'):
                successors.append(at+x.size)
            todo.extend((p,tuple(v[n] for n in regs),stack) for p in successors)
        self.preserved[key]=returned
        return returned

    def word(self, at):
        graph.require(0 <= at <= len(self.b)-2, 'bounded word extent')
        return struct.unpack_from('<H', self.b, at)[0]

    def immediate(self, x, reg):
        ops = x.operands
        if (x.mnemonic == 'mov' and len(ops) == 2
                and ops[0].type == self.x86.X86_OP_REG
                and x.reg_name(ops[0].reg) == reg and ops[1].type == self.x86.X86_OP_IMM):
            return ops[1].imm & 65535
        return None

    def prepare(self):
        cfg, self.seen = self.dec.graph(self.b, self.base, sorted(self.roots), self.limit)
        self.cfg = cfg
        self.pred = defaultdict(set)
        for at, x in self.seen.items():
            if x.group(self.cs.CS_GRP_RET) or x.mnemonic == 'iret':
                continue
            if x.group(self.cs.CS_GRP_JUMP) and x.operands[0].type == self.x86.X86_OP_IMM:
                target = self.base+x.operands[0].imm
                self.pred[target].add(at)
            if x.mnemonic not in ('jmp', 'ljmp'):
                self.pred[at+x.size].add(at)
        self.producers = defaultdict(list)
        for at, x in self.seen.items():
            ops = x.operands
            if not ops or ops[0].type != self.x86.X86_OP_MEM:
                continue
            m = ops[0].mem
            if m.base or m.index or m.segment:
                continue  # indexed/far/ES writers require a separate domain proof
            try:
                unused, writes = x.regs_access()
            except Exception:
                writes = []
            if x.mnemonic in ('cmp', 'test', 'push', 'call', 'jmp'):
                continue
            if x.mnemonic == 'mov' and len(ops) == 2 and ops[1].type == self.x86.X86_OP_IMM:
                self.producers[m.disp].append((at, ops[1].imm & 65535, None))
            else:
                self.producers[m.disp].append((at, None, 'nonliteral field writer'))

    def reaching(self, at, reg, budget=256):
        """Backward CFG slice, with explicit call/stack/load boundaries.

        Values are constants or ('field', DS word). No call is assumed to
        preserve a register. This intentionally under-resolves unsupported
        definitions instead of promoting a plausible literal to completeness.
        """
        todo, visited, values, defs, unknown = list(self.pred[at]), set(), set(), set(), set()
        while todo:
            p = todo.pop()
            if (p, reg) in visited:
                continue
            visited.add((p, reg))
            if len(visited) > budget:
                unknown.add('backward slice bound'); break
            x = self.seen[p]
            ops = x.operands
            value = self.immediate(x, reg)
            if value is not None:
                values.add(value); defs.add(p); continue
            if (x.mnemonic == 'mov' and len(ops) == 2
                    and ops[0].type == self.x86.X86_OP_REG and x.reg_name(ops[0].reg) == reg
                    and ops[1].type == self.x86.X86_OP_MEM):
                m = ops[1].mem
                if not m.base and not m.index and not m.segment:
                    values.add(('field', m.disp)); defs.add(p); continue
            if x.group(self.cs.CS_GRP_CALL):
                if (x.operands[0].type == self.x86.X86_OP_IMM
                        and self.preserves(self.base+x.operands[0].imm,reg)):
                    todo.extend(self.pred[p]);continue
                unknown.add('call preservation not proved at '+hex(p)); continue
            unused, writes = x.regs_access()
            names = {x.reg_name(r) for r in writes}
            aliases = {reg, reg[0]+'l', reg[0]+'h'} if reg in ('ax','bx','cx','dx') else {reg}
            if names & aliases:
                unknown.add('unsupported register definition at '+hex(p)); continue
            parents = self.pred[p]
            if not parents:
                unknown.add('entry register definition at '+hex(p))
            todo.extend(parents)
        return values, sorted(defs), sorted(unknown)

    def field(self, disp):
        values, defs, unknown = set(), [], []
        for at, v, error in self.producers[disp]:
            defs.append(at)
            if error:
                unknown.append(error+' at '+hex(at))
            else:
                values.add(v)
        # Initial literal values are included; zero is not silently discarded.
        values.add(self.word(self.ds+disp))
        return values, defs, unknown

    def expand_values(self, values, defs, unknown):
        targets = set()
        for value in values:
            if isinstance(value, tuple):
                v, d, u = self.field(value[1]); targets.update(v); defs.extend(d); unknown.extend(u)
            else:
                targets.add(value)
        return targets, sorted(set(defs)), sorted(set(unknown))

    def task_selector(self, call_at):
        """Byte state → half-index → bounded byte LUT → task word table."""
        if call_at != 0x20f4:
            return None
        start=0x20df
        sites=[];at=start
        while at<call_at:
            x=self.seen.get(at)
            if x is None:return None
            sites.append((at,x));at+=x.size
        checks=['mov','xor','shr','mov','mov','add','lodsw','mov']
        graph.require([x.mnemonic for unused,x in sites]==checks,'task selector shape')
        ops=[x.operands for unused,x in sites]
        graph.require(sites[0][1].op_str.startswith('bl, byte ptr')
            and sites[1][1].op_str=='bh, bh'
            and sites[2][1].op_str.startswith('bx, ')
            and sites[3][1].op_str.startswith('bl, byte ptr [bx + ')
            and sites[4][1].op_str=='si, bx'
            and sites[5][1].op_str.startswith('si, ')
            and sites[7][1].op_str=='dx, ax','task selector definitions')
        shift=ops[2][1].imm
        bound=(1 << (8*ops[0][0].size)) >> shift
        lookup=ops[3][1].mem.disp; words=ops[5][1].imm
        selectors=set(self.b[self.ds+lookup:self.ds+lookup+bound])
        targets={self.word(self.ds+words+i) for i in selectors}
        return targets,[p for p,unused in sites],dict(selector_base=self.ds+lookup,
            selector_count=bound,selector_values=sorted(selectors),
            pointer_sources=sorted(self.ds+words+i for i in selectors))

    def table(self, start, stride, target_offset, terminator='zero'):
        rows, at = [], self.ds+start
        for unused in range(512):
            first = self.word(at)
            if (terminator == 'zero' and first == 0) or (terminator == 'ff' and first == 255):
                return rows, at
            target = self.word(at+target_offset)
            rows.append(dict(source=at, target=self.base+target))
            if terminator == 'ff':
                end = self.b.find(b'$', at+2, min(at+16, self.ds+65536))
                graph.require(end >= 0, 'cheat record bound/terminator')
                at = end+1
            else:
                at += stride
            graph.require(at < self.ds+65536, 'bounded table DS extent')
        raise ValueError('bounded table has no terminator')

    def programs_from_consumers(self):
        roots, unknown = defaultdict(set), []
        effects=[]
        for at,x in self.seen.items():
            if (x.mnemonic!='call' or x.operands[0].type!=self.x86.X86_OP_IMM
                    or x.operands[0].imm not in (0x5c14,0x5bb5)):
                continue
            v,defs,errors=self.reaching(at,'si')
            v,defs,errors=self.expand_values(v,defs,errors)
            for value in v:
                source=self.ds+value
                no_jingle=self.word(source)==0
                program=self.word(source+26+int(no_jingle))
                if program:
                    if self.word(self.ds+program) in self.handlers:
                        roots[self.ds+program].add(at)
                    else:
                        errors.append('effect matrix pointer has unknown command head')
                effects.append(dict(source=source,consumer_call=at,definitions=defs,
                    optional_priority_byte=no_jingle,program=self.ds+program if program else None))
            if errors:
                unknown.append(dict(source=at,definitions=defs,unknown=sorted(set(errors))))
        self.effects=effects
        for at, x in self.seen.items():
            op = x.operands[0] if x.operands else None
            matrix_call = (x.mnemonic == 'call' and op.type == self.x86.X86_OP_IMM
                           and op.imm == 0x4501)
            matrix_jump = (x.mnemonic == 'jmp' and op.type == self.x86.X86_OP_MEM
                           and op.mem.base == self.x86.X86_REG_BX and not op.mem.segment)
            if not (matrix_call or matrix_jump):
                continue
            # These consume a typed current-command cursor, not a new external
            # root. Their branch words are parsed below with the opcode arity.
            if matrix_jump and at in (0x2fce,0x2ffe,0xa2a):
                if at==0x2fce:
                    self.dec.expect(self.b,self.base,0x2fb1,'mov','bx, word ptr [bx + 4]')
                elif at==0x2ffe:
                    self.dec.expect(self.b,self.base,0x2ff5,'add','bx, 2')
                    self.dec.expect(self.b,self.base,0x2ff8,'mov','bx, word ptr [bx]')
                else:
                    self.dec.expect(self.b,self.base,0xa21,'add','bx, 4')
                    self.dec.expect(self.b,self.base,0xa24,'mov','bx, word ptr [bx]')
                continue
            if matrix_jump and at in (0x2fe4,0x57d2,0xe38):
                if at==0x2fe4:
                    self.dec.expect(self.b,self.base,0x2fde,'mov','bx, word ptr [bx]')
                elif at==0x57d2:
                    self.dec.expect(self.b,self.base,0x57cf,'add','bx, 2')
                else:
                    self.dec.expect(self.b,self.base,0xe21,'add','bx, 2')
                continue
            if matrix_call and at in (0x5f08,0x5f9f):
                continue  # effect pointer source is derived through API calls
            v, defs, errors = self.reaching(at, 'bx')
            v, defs, errors = self.expand_values(v, defs, errors)
            for value in v:
                if 0 <= value < 65536 and self.word(self.ds+value) in self.handlers:
                    roots[self.ds+value].add(at)
                else:
                    errors.append('producer does not name a known program head')
            if errors:
                unknown.append(dict(source=at, definitions=defs, unknown=sorted(set(errors))))
        todo = list(roots)
        parsed, targets = {}, set()
        branches = {'_JMP': (0,), '_JBCDZ': (1,), '_JBONUSX1': (0,),
                    '_SHOW_SCORE': (0,), '_LOOP_': (1,)}
        while todo:
            start = todo.pop()
            if start in parsed:
                continue
            nodes, at, errors = [], start, []
            for unused in range(1024):
                if not self.ds <= at < self.ds+65536-2:
                    errors.append('program leaves bounded DS'); break
                h = self.word(at)
                if h == 0:
                    break
                spec = self.handlers.get(h)
                if spec is None:
                    errors.append('unresolved command identity at '+hex(at)); break
                op, arity = spec['op'], spec['arity']
                args = [self.word(at+2+2*j) for j in range(arity)]
                successors = []
                for j in branches.get(op, ()):
                    target = self.ds+args[j]
                    successors.append(target); roots[target].add(at); todo.append(target)
                nodes.append(dict(source=at, node=op, handler=self.base+h,
                                  operands=args, branch_programs=successors))
                targets.add(h)
                at += 2*(1+arity)
                if op in ('_JMP', 'QUIT'):
                    break
            else:
                errors.append('program decode bound')
            parsed[start] = dict(start=start, predecessors=sorted(roots[start]), nodes=nodes,
                                 status='UNKNOWN' if errors else 'DECODED', unknown=errors)
        for start in parsed:
            parsed[start]['predecessors'] = sorted(roots[start])
        self.programs = parsed
        self.command_targets = targets
        self.program_root_unknowns = unknown

    def resolve(self, at):
        x = self.seen[at]
        op = x.operands[0]
        targets, defs, errors, recipe, tables = set(), [], [], 'unsupported indirect operand', []
        if x.mnemonic in ('lcall', 'ljmp'):
            return dict(source=at, domain='far', status='UNKNOWN', targets=[], definitions=[],
                        recipe='segment/runtime API binding', unknown=['far callback segment/target domain'])
        if op.type == self.x86.X86_OP_MEM:
            m = op.mem
            if not m.base and not m.index and not m.segment:
                targets, defs, errors = self.field(m.disp)
                recipe = 'DS callback field writers plus initial value'
            elif not m.index and m.base == self.x86.X86_REG_BX and not m.segment:
                if at == 0x5eab:
                    # DOADDTASK is a distinct producer API. Derive arguments
                    # from every currently reachable call, repeat after closure.
                    recipe = 'task API reaching DX definitions'
                    reset = self.seen.get(0x3b42)
                    reset_value = self.immediate(reset, 'ax') if reset else None
                    if reset_value is None:
                        errors.append('task reset producer not reached/proved')
                    else:
                        targets.add(reset_value); defs.append(0x3b42)
                    for p, q in self.seen.items():
                        if q.mnemonic == 'call' and q.operands[0].type == self.x86.X86_OP_IMM and q.operands[0].imm == 0x5b80:
                            v,d,u = self.reaching(p,'dx'); v,d,u = self.expand_values(v,d,u)
                            selector=self.task_selector(p)
                            if selector:
                                v,d,record=selector;u=[];tables.append(record)
                            targets.update(v);defs.extend(d);errors.extend(u)
                    errors.append('indexed task-list writer domain not fully closed')
                else:
                    recipe = 'matrix program command domain'
                    targets.update(self.command_targets)
                    if self.program_root_unknowns or any(p['status']=='UNKNOWN' for p in self.programs.values()):
                        errors.append('matrix root/branch producer closure incomplete')
        elif op.type == self.x86.X86_OP_REG:
            reg = x.reg_name(op.reg)
            if reg == 'bx' and at == 0x38fb:
                # Slice recovers SI at the parser's record-list entry. The
                # decoder checks each LOAD/advance/terminator instruction.
                self.dec.expect(self.b,self.base,0x3894,'mov','si, 0x383c')
                self.dec.expect(self.b,self.base,0x389f,'lodsw')
                self.dec.expect(self.b,self.base,0x38a0,'cmp','ax, 0xff')
                self.dec.expect(self.b,self.base,0x38a8,'mov','bx, ax')
                self.dec.expect(self.b,self.base,0x38ae,'cmp','al, 0x24')
                rows,end = self.table(self.immediate(self.seen[0x3894],'si'),0,0,'ff')
                targets.update(r['target']-self.base for r in rows)
                tables.append(dict(start=self.ds+0x383c,end=end,rows=rows))
                defs=[0x3894,0x389f,0x38a8];recipe='bounded word + dollar-terminated cheat records'
            elif reg == 'ax' and at in (0x6095,0x6151):
                sites = (0x6064,) if at == 0x6095 else (0x6104,0x6111,0x611e)
                recipe = 'rectangle records: four bounds + near handler, zero sentinel'
                for p in sites:
                    q = self.seen.get(p)
                    v = self.immediate(q,'si') if q else None
                    if v is None:
                        errors.append('area table base producer not verified at '+hex(p));continue
                    rows,end=self.table(v,10,8)
                    targets.update(r['target']-self.base for r in rows)
                    defs.append(p);tables.append(dict(start=self.ds+v,end=end,rows=rows))
                errors.append('area index/DS-alias completeness not fully closed')
            elif reg == 'dx' and at in (0x7263,0x729a):
                start=0x724f if at==0x7263 else 0x7286
                chain=[];p=start
                while p<at:
                    q=self.seen[p];chain.append((p,q));p+=q.size
                graph.require([q.mnemonic for unused,q in chain]==['mov','and','shl','add','mov','add'],
                              'glyph callback index shape')
                graph.require(chain[1][1].op_str=='di, 0xff'
                    and chain[2][1].op_str=='di, 1'
                    and chain[4][1].op_str=='dx, word ptr es:[di]', 'glyph callback definitions')
                base=chain[3][1].operands[1].imm
                add=chain[5][1].operands[1].imm
                count=chain[1][1].operands[1].imm+1
                targets.update((self.word(self.ds+base+2*i)+add)&65535 for i in range(count))
                defs=[p for p,unused in chain]
                tables=[dict(start=self.ds+base,entries=count,stride=2,code_addend=add)]
                recipe='masked byte glyph index into word offsets + CS code base'
            else:
                recipe = 'backward register reaching definitions'
                v,defs,errors=self.reaching(at,reg)
                targets,defs,errors=self.expand_values(v,defs,errors)
        invalid = [v for v in targets if not v or not self.base <= self.base+v < self.limit]
        if invalid:
            errors.append('initial/derived non-code target requires lifecycle proof')
        targets={self.base+v for v in targets if v not in invalid}
        if not targets and not errors:
            errors.append('no bounded target producer')
        return dict(source=at,domain='near',status='UNKNOWN' if errors else 'BOUNDED',
                    targets=sorted(targets),definitions=sorted(set(defs)),recipe=recipe,
                    tables=tables,unknown=sorted(set(errors)))

    def run(self):
        for iteration in range(32):
            self.prepare()
            if self.handlers:
                self.programs_from_consumers()
            facts={r['source']:self.resolve(r['source']) for r in self.cfg['indirect_boundaries']}
            added=({t for f in facts.values() for t in f['targets']}
                   | {self.base+h for h in self.command_targets})-self.roots
            self.rounds.append(dict(iteration=iteration,roots=len(self.roots),
                                    instructions=len(self.seen),indirect_sites=len(facts),added_roots=len(added)))
            self.facts=facts
            if not added:
                unknown=[f for f in facts.values() if f['status']=='UNKNOWN']
                return dict(fixed_point=True,rounds=self.rounds,domains=list(facts.values()),
                            unknown_domain_count=len(unknown),unknown_sites=[f['source'] for f in unknown],
                            root_files=sorted(self.roots),
                            preservation_summaries=[dict(entry=p,register=n,preserved=v)
                                for (p,n),v in sorted(self.preserved.items())],
                            persistence_candidates=[dict(start=lo,end_exclusive=hi,
                                body_in_direct_or_candidate_cfg=any(lo<=p<hi for p in self.seen),
                                verdict='UNKNOWN: remaining domain/lifecycle proofs required')
                                for lo,hi in ((0x66d2,0x6706),(0x6706,0x673d))],
                            programs=list(self.programs.values()),
                            effects=self.effects,
                            unknown_program_roots=self.program_root_unknowns,
                            unknown_program_body_count=sum(p['status']=='UNKNOWN' for p in self.programs.values()),
                            conservative_reachability=True)
            self.roots.update(added)
        raise ValueError('target resolver did not reach bounded fixed point')


def audit(data, canonical, historical):
    demo,full=graph.pinned(Path(data),Path(canonical),Path(historical))
    table=demo['TABLE1.PRG']
    h=struct.unpack_from('<14H',table)
    load=h[4]*16; cs=load+h[11]*16
    entry=cs+h[10]
    # Prior proved INT66 primary/later registrations are retained roots. The
    # resolver starts from the actual MZ entry rather than a mid-startup label.
    limit=load+struct.unpack_from('<H',table,0x56aa)[0]*16
    r=Resolver(table,cs,0x19db0,[entry,0x4517,0x592b],limit)
    import audit_10min_demo_programs as programs
    r.handlers, identities = programs.identities(table, Path(historical))
    result=r.run()
    result.update(status='DMO0 NOT CLOSED. DMO1 NOT STARTED.',table_entry=entry,main_cs_end=limit,
                  scope='TABLE1 conservative target closure; INTRO/SDR semantic admission gates remain separate')
    result['command_identity_derivation'] = identities
    result['open_scope_gates'] = ['INTRO indirect control domain closure',
        'all eleven SDR IRQ/admission contracts',
        'state-feasible match/high-score/exit and expiry interleaving closure',
        'control-object writer aliases and callback initialization/lifecycle']
    return result


def require_closed(result):
    graph.require(not result['unknown_domain_count'] and not result['open_scope_gates'],
                  'DMO0 unresolved target/semantic scopes')


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for n,e in [('data','PF_10MIN_DEMO_DATA'),('canonical','PF_RUNTIME_DATA'),('historical','PF_DMO0_HISTORICAL_SOURCE')]:
        p.add_argument('--'+n,type=Path,default=os.getenv(e))
    p.add_argument('--output',type=Path,required=True)
    p.add_argument('--allow-open',action='store_true',help='emit partial metadata without claiming closure')
    a=p.parse_args();graph.require(all((a.data,a.canonical,a.historical)),'private paths required')
    result=audit(a.data,a.canonical,a.historical)
    a.output.write_text(json.dumps(result,indent=2)+'\n')
    print('fixed point:',result['fixed_point'],'UNKNOWN TABLE1 domains:',result['unknown_domain_count'])
    if not a.allow_open:
        try: require_closed(result)
        except ValueError: raise SystemExit(2)
