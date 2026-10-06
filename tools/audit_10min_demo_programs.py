#!/usr/bin/env python3
"""Linked command identities from bounded source-shaped program constraints.

A unique structural match identifies command handlers, not a reachable program.
The consuming-code resolver alone supplies roots. Candidate labels remain
untrusted until a predecessor supplies that address; shared source is not the
exact demo build. Numeric/symbolic operand differences are not normalized away.
"""
import struct
import re
import matrix_operand_schema as schema
import audit_10min_demo_graph as g


def identities(b, historical):
    p=g.generated('internal/presentation/content.go')['1']; cmds=p['commands']; labs=p['labels']
    i=labs['NO_BONUS2TS']+1
    cmds=cmds[:i]+[{'op':'_DEMOVER_CHANGE_PLAYER','args':[]}]+cmds[i+4:]
    labs={n:v-(3 if v>=i+4 else 0) for n,v in labs.items() if not i<=v<i+4}
    unused,known=g.bonus_program(b,0x19db0,True)
    rev={v:k for k,v in known.items()};proglist={}
    for n,i in labs.items():
     c=[]
     while i<len(cmds):
      q=cmds[i];c.append(q);i+=1
      if q['op']=='0':break
     proglist[n]=c
    # Parse bounded shared DATA declarations omitted by the gameplay/frontend
    # extractor (notably quit confirmation). Source supplies arities and labels;
    # the demo binary still supplies every handler identity and operand.
    shared=(historical/'FANTASIE.ASM').read_bytes().decode('latin1').upper()
    shared=shared.split('DATA\tENDS',1)[0]
    current=None
    for line in shared.splitlines():
        line=line.split(';')[0].strip()
        label=re.match(r'^(\w+)\s+LABEL\s+(?:WORD|BYTE)$',line)
        if label:
            current=label[1];proglist.setdefault(current,[]);continue
        inline=re.match(r'^(\w+)\s+DW\s+(.+)$',line)
        if inline:
            current=inline[1];proglist.setdefault(current,[]);line='DW '+inline[2]
        if line=='CLEARIT' and current:
            proglist[current].append(dict(op='_CLEAR4',args=[]));continue
        dw=re.match(r'^DW\s+(.+)$',line)
        if dw and current:
            parts=[v.strip().replace('OFFSET ','') for v in dw[1].split(',')]
            op,args=parts[0],parts[1:]
            if not (op.startswith('_') or op in ('QUIT','0')):
                current=None;continue
            nums={}
            for j,arg in enumerate(args):
                try: nums[str(j)]=schema.resolve_expression(arg,{'SW':336,'TOTCENT':0})
                except (schema.OperandError,ValueError,KeyError,SyntaxError,ZeroDivisionError):pass
            proglist[current].append(dict(op=op,args=args,nums=nums))
            if op in ('0','_JMP'):
                current=None
    proglist={k:v for k,v in proglist.items() if v}
    linked={}; remain=dict(proglist)
    while True:
     changes=0
     for name,cmd in sorted(remain.items(),key=lambda x:-len(x[1])):
      pos=[];numeric=[];count=0
      for q in cmd:
       pos.append(count);count+=1
       for j,a in enumerate(q['args']):
        v=q.get('nums',{}).get(str(j))
        if v is None and a.isdigit():v=int(a)
        if v is not None:numeric.append((count,v&65535))
        count+=1
      candidates=[]
      first=known.get(cmd[0]["op"]); starts=[]
      if first is None: starts=range(0x19db0,0x29db0-2*count)
      else:
       st=0x19db0; needle=struct.pack("<H",first)
       while True:
        st=b.find(needle,st,0x29db0-2*count+2)
        if st<0:break
        starts.append(st);st+=1
      for start in starts:
       if not all(struct.unpack_from('<H',b,start+idx*2)[0]==v for idx,v in numeric):continue
       local=dict(known);r=dict(rev);ok=True
       for q,idx in zip(cmd,pos):
        h=struct.unpack_from('<H',b,start+idx*2)[0];op=q['op']
        if op in local and local[op]!=h:ok=False;break
        if h in r and r[h]!=op:ok=False;break
        if op!='0' and not 0<h<0xacd0:ok=False;break
        local[op]=h;r[h]=op
       if ok:candidates.append((start,local,r))
      if candidates:
       # A label can have several structurally identical candidates while a
       # handler identity is invariant over all of them. Bind that invariant,
       # retaining the label ambiguity; do not arbitrarily pick a root.
       for op in {q['op'] for q in cmd}-set(known):
        values={local[op] for unused,local,unused_rev in candidates}
        if len(values)==1:
         h=values.pop();g.require(h not in rev,'consensus handler alias')
         known[op]=h;rev[h]=op;changes+=1
      if len(candidates)==1:
       st,unused_local,unused_rev=candidates[0];linked[name]=dict(start=st,size=count*2,nodes=[dict(op=q['op'],at=st+idx*2) for q,idx in zip(cmd,pos)]);del remain[name];changes+=1
     if not changes:break
    arities={}
    for q in [q for program in proglist.values() for q in program]:
        if q['op'] in arities:
            g.require(arities[q['op']]==len(q['args']), 'command arity inconsistency')
        arities[q['op']]=len(q['args'])
    # Already proved demo expiry sequence; linked words supply only the two
    # identities absent from the full frontend declaration stream.
    expiry=['_CLEAR4','_SCROLL','_FLASHON','_PRINT13_NUMBER','_WAIT',
            '_FLASHOFF','_SCROLL','_FADE','_WAIT','QUIT','0']
    arities.update(_FADE=1,QUIT=1)
    at=0x1ba17
    for op in expiry:
        h=struct.unpack_from('<H',b,at)[0]
        g.require(known.get(op,h)==h, 'expiry identity disagrees with linked command')
        g.require(rev.get(h,op)==op, 'expiry identity alias')
        known[op]=h;rev[h]=op;at+=2*(1+arities[op])
    g.require(at==0x1ba17+42,'expiry command extent')
    return {h:dict(op=op,arity=arities[op]) for op,h in known.items() if h},dict(
        structural_candidates={n:r['start'] for n,r in linked.items()},
        unlinked_source_labels=sorted(remain),identity_count=len(known)-1)
