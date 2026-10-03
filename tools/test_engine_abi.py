#!/usr/bin/env python3
"""Compare a real c-shared engine to the direct Go Runner, with external assets.
Usage: test_engine_abi.py --library LIB --oracle PFTRACE --data ORIGINALS
No fixture or commercial payload is written to the repository.
"""
import argparse, ctypes as C, hashlib, json, pathlib, subprocess, tempfile

def trace(table):
    steps=[]; ns=0
    def step(**kw):
        nonlocal ns
        steps.append(dict(NS=ns, **kw)); ns+=16_666_667 if len(steps)<3 else 14_084_508
    step(Frame=True); step(Keys=[57],Frame=True);step(Keys=[58+table],Frame=True)
    letters='QWERTYUIOPASDFGHJKLZXCVBNM'
    codes=[16,17,18,19,20,21,22,23,24,25,30,31,32,33,34,35,36,37,38,44,45,46,47,48,49,50]
    for word in ['EARTHQUAKE','EXTRA BALLS','FAIR PLAY','SNAIL','CHEAT','JOHAN','DANIEL','GABRIEL','TSP','TECH','ROBBAN','STEIN','GREET','EXTRA BALLS']:
        for char in word:step(Keys=[57 if char==' ' else codes[letters.index(char)]])
        step(Frame=True)
    step(Keys=[60],Frame=True) # F2: two players
    for tick in range(2400):
        if tick==300:
            step(Suspend=True,Frame=True);ns+=3_600_000_000_000
            step(Actions=[[0,1],[2,1]],Delta=4096,Fire=True,Frame=True)
            step(Resume=True,Frame=True);step(Keys=[25],Frame=True)
        if tick%137==0:ns+=70_422_535 # slow wake: all five overdue tasks
        step(Actions=[[0,int(tick%93<18)],[1,int(tick%71<14)],[2,int(tick<32)],[3,int(tick==120)]],
             Delta=8 if tick>=600 and tick%600<32 else 0,Fire=tick>=600 and tick%600==32,Frame=tick%211==0)
    return dict(Steps=steps)

def main():
    p=argparse.ArgumentParser();p.add_argument('--library',required=True);p.add_argument('--oracle',required=True);p.add_argument('--data',required=True);a=p.parse_args()
    lib=C.CDLL(a.library);u64=C.c_uint64;u32=C.c_uint32;i32=C.c_int32;i64=C.c_int64;ptr=C.c_void_p
    sink_type=C.CFUNCTYPE(None,ptr,C.POINTER(C.c_uint8),u32)
    signatures={
      'create':([C.c_char_p,C.c_char_p,i64,C.c_char_p,u32],u64),'destroy':([u64],i32),
      'suspend':([u64],i32),'resume':([u64,i64],i32),'set_action':([u64,u32,i32],i32),
      'key':([u64,C.c_uint8],i32),'release':([u64],i32),'plunger_delta':([u64,i32],i32),
      'plunger_fire':([u64],i32),'advance':([u64,i64,sink_type,ptr],i32),
      'frame':([u64,C.POINTER(ptr),C.POINTER(i32),C.POINTER(i32),C.POINTER(i32)],i32),
      'state':([u64,C.POINTER(u64),C.POINTER(u32),C.POINTER(u32),C.POINTER(u32)],i32)}
    for name,(args,ret) in signatures.items():fn=getattr(lib,'pf_engine_'+name);fn.argtypes=args;fn.restype=ret
    def call(name,*args):
        result=getattr(lib,'pf_engine_'+name)(*args)
        assert result==0,(name,result)
    assert lib.pf_engine_abi_version()==1
    assert lib.pf_engine_destroy(0)==-1
    error=C.create_string_buffer(256)
    assert lib.pf_engine_create(b'/nonexistent',b'/tmp',0,error,256)==0 and error.value
    for table,name in enumerate(['Party Land','Speed Devils','Billion Dollar Gameshow',"Stones 'N Bones"],1):
      for scroll in [2,3]:
        with tempfile.TemporaryDirectory(prefix='pf-abi-') as tmp:
          state=pathlib.Path(tmp)/'state';state.mkdir()
          # Existing shared native config encoding, SOFT and full-table OFF.
          (state/'PINBALL.CFG').write_bytes(b'PFNC\x01\x05'+bytes([0,1,scroll,0,1]))
          replay=trace(table)
          expected=json.loads(subprocess.check_output([a.oracle,'-data',a.data,'-state',str(state)],input=json.dumps(replay).encode()))
          handle=lib.pf_engine_create(a.data.encode(),str(state).encode(),0,error,256)
          assert handle,error.value
          count=0;digest=None;callback_errors=[]
          def receive(ctx,samples,size):
            nonlocal count
            if size%4:callback_errors.append('partial PCM frame')
            count+=size//4;digest.update(C.string_at(samples,size))
            if lib.pf_engine_set_action(handle,0,1)!=-2:callback_errors.append('reentry not rejected')
          sink=sink_type(receive)
          for index,(step,want) in enumerate(zip(replay['Steps'],expected)):
            ns=step['NS']
            if step.get('Suspend'):call('suspend',handle)
            if step.get('Resume'):call('resume',handle,ns)
            for action,down in step.get('Actions',[]):call('set_action',handle,action,down)
            for key in step.get('Keys',[]):call('key',handle,key)
            call('plunger_delta',handle,step.get('Delta',0))
            if step.get('Fire'):call('plunger_fire',handle)
            digest=hashlib.sha256();count=0;call('advance',handle,ns,sink,None)
            tick=u64();mode=u32();selected=u32();flags=u32();call('state',handle,C.byref(tick),C.byref(mode),C.byref(selected),C.byref(flags))
            got=dict(Tick=tick.value,Mode=mode.value,Table=selected.value,Flags=flags.value,PCM=digest.hexdigest(),Frames=count,Frame='',Width=0,Height=0,Stride=0)
            if step.get('Frame'):
              pixels=ptr();w=i32();h=i32();stride=i32();call('frame',handle,C.byref(pixels),C.byref(w),C.byref(h),C.byref(stride))
              got.update(Frame=hashlib.sha256(C.string_at(pixels,stride.value*h.value)).hexdigest(),Width=w.value,Height=h.value,Stride=stride.value)
              # Borrowed storage survives advance and stays at the same address
              # until another retrieval/destroy (growth may change its address).
            assert got==want,(name,scroll,index,got,want)
          assert not callback_errors,callback_errors
          assert lib.pf_engine_advance(handle,-1,sink,None)==-3
          assert lib.pf_engine_set_action(handle,99,1)==-3
          call('destroy',handle);assert lib.pf_engine_destroy(handle)==-1
          print(f'PASS {name} scroll={scroll}: {len(expected)} checkpoints; tasks={expected[-1]["Tick"]}; PCM/frame/state parity')
if __name__=='__main__':main()
