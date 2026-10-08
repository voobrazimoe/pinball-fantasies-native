#!/usr/bin/env python3
"""Callback/vector admission metadata; candidate discovery never grants closure.

Only semantic records are exported. Driver code stays decoded in memory.
The executable-name producer is an external file read, not a driver allowlist.
"""
from pathlib import Path
from audit_10min_demo_graph import Decoder, require, sha
from audit_10min_demo_sdr import SDR, unpack, callback_api


def invocation_context(kind, ss, sp, returned, return_adjust=0, prefix_bytes=0):
    """Architectural context for an already admitted source, not admission proof."""
    frame = {'near': 2, 'far': 4, 'interrupt': 6}.get(kind)
    expected = {'near': 'ret', 'far': 'retf', 'interrupt': 'iret'}.get(kind)
    complete = frame is not None and returned == expected and return_adjust == 0
    return dict(instruction_class=kind, frame_bytes=frame,
        frame_content={'near':['IP'], 'far':['CS','IP'],
                       'interrupt':['FLAGS','CS','IP']}.get(kind),
        entry_ss=ss, entry_sp=None if sp is None or frame is None else (sp-prefix_bytes-frame)&65535,
        return_instruction=returned, restored_source_sp=(sp-prefix_bytes)&65535 if complete and sp is not None else None,
        complete=complete and ss is not None and sp is not None)


def binding_contract(binding):
    """Validate a producer's certificate across all writers and handler effects.

    This does not derive certificates from booleans. Real extraction below
    supplies incomplete certificates. Authored tests exercise certificate
    consumption; target identity may vary only within a complete effect domain.
    """
    errors=[]
    for key in ('writers_complete','ordering_complete','no_later_overwrite','effect_domain_complete'):
        if binding.get(key) is not True: errors.append(key+' unproved')
    rows=binding.get('handler_summaries',[])
    if not rows:errors.append('handler contract unknown')
    for row in rows:
        if row.get('status')!='BOUNDED':errors.append('handler effects unknown')
        if row.get('return_instruction')!='iret' or row.get('stack_delta')!=0:
            errors.append('unmatched interrupt return')
        if any(row.get(n)!='preserving' for n in ('ss','sp')):
            errors.append('stack transform unproved')
        if any(row.get(n) not in ('preserving','bounded finite effect') for n in ('ds','es','allocation')):
            errors.append('register/allocation effect unknown')
        if row.get('object_memory')!='EXCLUDED':errors.append('lookup-table write not excluded')
        if not isinstance(row.get('maximum_downward_bytes'),int) or row['maximum_downward_bytes']<0:
            errors.append('handler depth unknown')
    return dict(status='UNKNOWN' if errors else 'BOUNDED',unresolved_reason=sorted(set(errors)),
        proof_scope='handler effect certificate consumption; caller physical stack exclusion remains separate',
        candidate_targets=binding.get('candidate_targets',[]),
        installed_target_set=binding.get('candidate_targets') if binding.get('writers_complete') and binding.get('ordering_complete') and binding.get('no_later_overwrite') else None,
        contract_identity='variable' if binding.get('identity_unknown') else 'enumerated')


def callback_contract(callback):
    errors=[];contexts=[]
    if not callback.get('invocations_complete'):errors.append('invocation source set incomplete')
    targets=callback.get('registration_targets',[])
    if not targets or any(v!=callback['root'] for v in targets):errors.append('alternate or unknown registration target')
    rows=callback.get('invocations',[])
    if not rows:errors.append('registration without invocation proof')
    for row in rows:
        if row.get('return_instruction')!=callback.get('return_instruction','retf'):
            errors.append('callback return incompatible with invocation frame')
        c=invocation_context(row.get('kind'),row.get('ss'),row.get('sp'),
            row.get('return_instruction'),row.get('return_adjust',0),row.get('prefix_bytes',0))
        contexts.append(c)
        if not c['complete']:errors.append('invocation stack/return context incomplete')
        if not row.get('linkage_complete'):errors.append('registration consumer linkage incomplete')
    return dict(status='UNKNOWN' if errors else 'BOUNDED',contexts=contexts,
        proof_scope='invocation stack transforms only; cannot authorize object-memory or vector closure',
        unresolved_reason=sorted(set(errors)))


def external_driver_name(r):
    """Check the full read-to-EXEC instruction path, without listing bytes."""
    expected=[(0x661c,'mov','dx, 0x23b6'),(0x661f,'mov','ax, 0x3d02'),
        (0x6622,'int','0x21'),(0x6628,'jb','0x6393'),
        (0x6631,'mov','cx, 0xd'),(0x6634,'mov','dx, 0x23c0'),
        (0x6637,'mov','ax, 0x3f00'),(0x663a,'int','0x21'),
        (0x6651,'mov','ax, 0x4b00'),(0x6657,'mov','dx, 0x23c0'),(0x665a,'int','0x21')]
    for at,m,op in expected:r.dec.expect(r.b,r.base,at,m,op)
    # Read/close/EXEC interval has no conditional transfer or string validation.
    at=0x663a;sites=[]
    while at<=0x665a:
        x=r.seen.get(at);require(x is not None,'driver EXEC path missing')
        require(not x.group(r.cs.CS_GRP_JUMP) and not x.group(r.cs.CS_GRP_CALL),
                'driver name validation/control changed')
        sites.append(at);at+=x.size
    start=r.ds+0x23b6
    require(r.b[start:start+10]==b'SOUND.CFG\0','driver configuration source changed')
    return dict(module='TABLE1',configuration='SOUND.CFG',open_site=0x6622,
        read_site=0x663a,read_bytes=13,destination_ds_offset=0x23c0,
        exec_site=0x665a,exec_ax=0x4b00,exec_name_ds_offset=0x23c0,
        read_to_exec_sites=sites,name_allowlist=False,read_length_checked=False,
        exec_success_checked=False,
        reason='external configuration bytes flow to normal DOS EXEC without supplied-SDR name or content allowlist; filesystem executable identity unpinned')


def partial_graph(dec, raw, roots):
    """Retain undecoded boundaries as UNKNOWN, including candidate false roots."""
    todo=list(roots);seen={};indirect=[];unknown=[]
    while todo:
        p=todo.pop()
        if p in seen:continue
        if not 0<=p<len(raw) or len(seen)>=40000:
            unknown.append(dict(site=p,reason='CFG extent/state bound'));continue
        x=next(dec.decoder.disasm(raw[p:p+15],p),None)
        if x is None:
            unknown.append(dict(site=p,reason='undecodable candidate instruction'));continue
        seen[p]=x
        if x.group(dec.cs.CS_GRP_RET) or x.mnemonic in ('iret','hlt'):continue
        if x.group(dec.cs.CS_GRP_CALL) or x.group(dec.cs.CS_GRP_JUMP):
            if x.mnemonic not in ('lcall','ljmp') and x.operands[0].type==dec.x86.X86_OP_IMM:
                todo.append(x.operands[0].imm)
            else:indirect.append(dict(source=p,kind=x.mnemonic))
            if x.mnemonic in ('jmp','ljmp'):continue
        todo.append(p+x.size)
    return dict(indirect_boundaries=indirect,unknown=unknown),seen


def driver_candidates(data, installers):
    dec=Decoder();rows=[]
    for install in installers:
        name=install['module'];size,digest,extent=SDR[name]
        b=(Path(data)/name).read_bytes();require(len(b)==size and sha(b)==digest,'SDR identity drift')
        raw,_=unpack(b);require(len(raw)==extent,'SDR extent drift')
        api=callback_api(name,raw,dec)  # retained registration producer derivation
        prefix=list(dec.decoder.disasm(raw[api['int66_entry']:api['int66_entry']+5],api['int66_entry']))
        require([(x.mnemonic,x.op_str) for x in prefix]==[('cld',''),('sti',''),('pushaw',''),('push','ds'),('push','es')],
                'INT66 handler save prologue changed')
        primary=api['callback_registration'][0];pointer=primary['producer']['pointer_ds_offset']
        # Prologue matching discovers candidates, never a complete IRQ root set.
        roots=[]
        for p in range(api['data_segment']*16):
            if raw[p]!=0xfc:continue
            xs=list(dec.decoder.disasm(raw[p:p+10],p))
            raster=[(x.mnemonic,x.op_str) for x in xs[:4]]==[('cld',''),('pushaw',''),('push','es'),('push','ds')]
            direct=([x.mnemonic for x in xs[:6]]==['cld','sti','pushaw','push','push','pop']
                    and xs[3].op_str=='es' and xs[4].operands[0].type==dec.x86.X86_OP_IMM and xs[5].op_str=='ds')
            if raster or direct:roots.append(p)
        consumers=[];frontiers=[];vector_writers=[];returns=set();source_returns=set();decode_unknown=[]
        for root in roots:
            cfg,seen=partial_graph(dec,raw,[root])
            decode_unknown.extend(cfg['unknown'])
            frontiers.extend(dict(root=root,**v) for v in cfg['indirect_boundaries'])
            for p,x in seen.items():
                if x.mnemonic=='iret':source_returns.add(p)
                if x.mnemonic!='lcall':continue
                m=x.operands[0].mem
                direct=not m.base and not m.index and m.disp==pointer
                indexed=m.base==dec.x86.X86_REG_SI and not m.index and m.disp==-4 and not m.segment
                if not (direct or indexed):continue
                previous=[v for v in seen.values() if v.address+v.size==p]
                prefix=2 if len(previous)==1 and previous[0].mnemonic=='push' and previous[0].op_str=='ax' else None
                consumers.append(dict(driver=name,root_candidate=root,site=p,instruction_class='far',
                    registration_api=11,registration_producer=primary['producer'],pointer_ds_offset=pointer,
                    linkage='absolute operand' if direct else 'first record when SI=primary pointer+4; indexed SI domain incomplete',
                    linkage_complete=False,entry_cs='registered ES',entry_ip='registered DX (TABLE1 IP 0x4217)',
                    entry_ss='source SS; no SS change by far CALL',entry_sp='u16(source SP-4)',
                    pushed_frame=dict(bytes=4,order=['CS','IP']),preceding_argument_bytes=prefix,
                    ds='driver data at prologue; callback assigns TABLE1 DS',es='UNKNOWN at consumer',
                    return_instruction='TABLE1 RETF candidates; matched return incomplete',
                    restored_caller_ss_sp='UNKNOWN; requires balanced callback plus source admission',
                    complete=False))
        # Entry/API direct graph gives direct writer/return candidates only.
        cfg,seen=partial_graph(dec,raw,[install['entry'],api['int66_entry']])
        decode_unknown.extend(cfg['unknown'])
        for p,x in seen.items():
            if x.mnemonic=='iret':returns.add(p)
            for op in x.operands:
                if op.type==dec.x86.X86_OP_MEM and op.access&dec.cs.CS_AC_WRITE and not op.mem.base and not op.mem.index and 0x198<=op.mem.disp<0x19c:
                    vector_writers.append(dict(site=p,segment_register=x.reg_name(op.mem.segment),offset=op.mem.disp,
                        qualification='candidate; segment identity requires reaching definitions'))
        rows.append(dict(driver=name,installer=install,primary_registration=primary,
            invocation_producers=consumers,invocation_candidates_complete=False,
            source_iret_candidates=sorted(source_returns),
            indirect_frontiers=frontiers,decode_unknown=decode_unknown,direct_vector_writer_candidates=vector_writers,
            handler=dict(entry=api['int66_entry'],architectural_frame=dict(bytes=6,order=['FLAGS','CS','IP']),
                entry_ss='caller SS',entry_sp='u16(caller SP-6)',prologue_save_bytes=20,
                prologue_saves=['PUSHA (16)','DS (2)','ES (2)'],iret_candidates=sorted(returns),
                maximum_downward_bytes=None,ss='UNKNOWN',sp='UNKNOWN',ds='UNKNOWN',es='UNKNOWN',
                object_memory='UNKNOWN',allocation='UNKNOWN',status='UNKNOWN',
                unresolved_reason='AL dispatch/nested calls/aliases and physical stack not closed; prologue and IRET candidates do not prove complete return contract')))
    return rows


def analyze(r, data, installers, provenance):
    name=external_driver_name(r);drivers=driver_candidates(data,installers)
    invokers=[v for d in drivers for v in d['invocation_producers']]
    registration=[]
    for at,m,op in [(0x6345,'mov','ax, 0xb'),(0x634a,'mov','dx, 0x4217'),
                    (0x634d,'push','cs'),(0x634e,'pop','es'),(0x634f,'int','0x66')]:
        r.dec.expect(r.b,r.base,at,m,op);registration.append(at)
    callback=dict(root=0x4517,registration_site=0x6345,registration_interrupt=0x634f,
        registration_api=11,registration_targets=[0x4517],registration_definitions=registration,
        invocation_producers=invokers,candidate_source_modules=[d['driver'] for d in drivers],
        invocation_target_set=None,complete=False,entry_ss=None,entry_sp=None,
        return_transform=None,ds='callback prologue assigns relocated TABLE1 DS',es='UNKNOWN',
        object_memory='UNKNOWN: upstream writer certificate retained',
        unresolved_reason=['external executable named by SOUND.CFG is not restricted to supplied SDR identities',
            'candidate IRQ root/record traversal does not prove complete invocation source set',
            'interrupted/source SS:SP domain and complete matched callback return unproved'])
    # Negative syntactic inventory is explicitly weaker than alias/ABI closure.
    ivt33=[];dos_vectors=[]
    for p,x in r.seen.items():
        for op in x.operands:
            if op.type==r.x86.X86_OP_MEM and op.access&r.cs.CS_AC_WRITE and not op.mem.base and not op.mem.index and 0xcc<=op.mem.disp<0xd0:
                ivt33.append(dict(site=p,segment_register=x.reg_name(op.mem.segment),offset=op.mem.disp))
        if x.mnemonic=='int' and x.op_str=='0x21':
            previous=[q for q,y in r.seen.items() if q+y.size==p and y.mnemonic=='mov' and y.op_str in ('ax, 0x2533','ax, 0x2566','ah, 0x25')]
            if previous:dos_vectors.append(dict(site=p,definitions=previous))
    callback['contract_validation']=callback_contract(dict(callback,
        invocations_complete=False,invocations=[dict(kind='far',ss=None,sp=None,
            return_instruction='retf',linkage_complete=False) for v in invokers]))
    vector_validation=binding_contract(dict(writers_complete=False,ordering_complete=False,
        no_later_overwrite=False,effect_domain_complete=False,
        candidate_targets=[dict(module=v['module'],offset=v['target_offset']) for v in installers],
        handler_summaries=[d['handler'] for d in drivers]))
    matched=[]
    for caller in provenance['entry_callers']:
        for dep in caller['matched_return_effect_dependencies']:
            if dep['site'] not in (0x4726,0x573e):continue
            matched.append(dict(dep,code2_caller=caller['site'],
                required_interrupt_references=['callback_and_interrupt_admission.int33','callback_and_interrupt_admission.int66'] if dep['site']==0x4726 else ['callback_and_interrupt_admission.int66'],
                transitive_incomplete_summaries=[row for row in caller['transitive_stack_effect_dependencies'] if row['status']=='UNKNOWN'],
                finite_rejoin_status='UNKNOWN',ss_sp_return_status='UNKNOWN'))
    return dict(callback_root=callback,driver_name_admission=name,drivers=drivers,
        int66=dict(vector=0x66,vector_address=0x198,initial_state='environmental IVT; unpinned',
            vector_writers=installers,vector_writers_complete=False,
            writer_ordering=dict(local='external DOS EXEC precedes TABLE1 INT66 at 0x6663 on fallthrough',
                complete=False,reason='EXEC failure unchecked; executable identity unrestricted; no postinstall writer exclusion'),
            candidate_targets=[dict(module=v['module'],offset=v['target_offset'],segment=v['target_segment']) for v in installers],
            installed_target_set=None,handler_summaries=[d['handler'] for d in drivers],contract_validation=vector_validation,
            later_replacement='UNKNOWN: external executable/handlers and environmental vector writers unpinned',
            uninstall_restore='UNKNOWN: direct candidate writer inventory does not exclude aliases/external restoration',
            required_sites=[0x6ec,0x6256,0x55ba,0x5fdc],complete=False,
            object_memory='UNKNOWN',unresolved_reason=name['reason']),
        int33=dict(vector=0x33,vector_address=0xcc,required_sites=[0x618f],ax=3,
            candidate_targets=[],installed_target_set=None,identity='unknown environmental implementation',
            invariant_contract=None,contract_status='UNKNOWN',complete=False,
            table1_direct_slot_store_candidates=ivt33,table1_dos_set_vector_candidates=dos_vectors,
            table1_installation='no fixed installer found by direct slot/immediate DOS-vector inventory; indexed aliases/external replacements remain unproved',
            entry_transform=dict(ss='caller SS',sp='u16(caller SP-6)',frame_bytes=6),
            return_transform=None,object_memory='UNKNOWN',
            unresolved_reason='no pinned implementation or mechanically invariant stack/register/object-memory contract supplied'),
        matched_returns=matched,
        code2_callers=[dict(site=v['site'],ss=v['ss'],sp=v['sp'],status=v['status'],
            interrupts=[h['site'] for h in v['admitted_interrupts']],unresolved_reason=v['unresolved_reason']) for v in provenance['entry_callers']],
        startup_int21_required_for_callback_stack=False,
        completeness_flag=False,status='UNKNOWN',
        table1_external_api_effect={hex(p):'partially discharged: candidate bindings retained; admission and transfer/memory contracts open' for p in (0x3a01,0x3d13,0x5640)},
        scope='binding/admission and entry/return effects only; no callback cadence or IRQ frequency')
