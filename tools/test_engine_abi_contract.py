#!/usr/bin/env python3
"""Asset-free checks against a real C shared library, never gameplay acceptance."""
import argparse
import ctypes as C
p=argparse.ArgumentParser(description=__doc__); p.add_argument('--library',required=True); a=p.parse_args()
lib=C.CDLL(a.library)
u64=C.c_uint64; u32=C.c_uint32; i32=C.c_int32; i64=C.c_int64
signatures={
 'abi_version':([],i32),
 'create':([C.c_char_p,C.c_char_p,i64,C.c_void_p,u32],u64),
 'destroy':([u64],i32),'suspend':([u64],i32),'resume':([u64,i64],i32),
 'set_presentation':([u64,i32],i32),
 'set_action':([u64,u32,i32],i32),'key':([u64,C.c_uint8],i32),
 'gamepad':([u64,i32,i32,i32],i32),
 'release':([u64],i32),'plunger_delta':([u64,i32],i32),'plunger_fire':([u64],i32),
 'advance':([u64,i64,C.c_void_p,C.c_void_p],i32),
 'frame':([u64,C.POINTER(C.c_void_p),C.POINTER(i32),C.POINTER(i32),C.POINTER(i32)],i32),
 'state':([u64,C.POINTER(u64),C.POINTER(u32),C.POINTER(u32),C.POINTER(u32)],i32)}
for name,(args,result) in signatures.items():
    fn=getattr(lib,'pf_engine_'+name); fn.argtypes=args; fn.restype=result
assert lib.pf_engine_abi_version()==1
error=C.create_string_buffer(256)
assert lib.pf_engine_create(b'/pf-absent-originals',b'/tmp',0,error,256)==0 and error.value
short=C.create_string_buffer(1)
assert lib.pf_engine_create(b'/pf-absent-originals',b'/tmp',0,short,1)==0 and short.raw==b'\0'
assert lib.pf_engine_create(b'/pf-absent-originals',b'/tmp',0,None,0)==0
pixel=C.c_void_p(); w=i32(); h=i32(); stride=i32(); tick=u64(); mode=u32(); table=u32(); flags=u32()
for handle in (0,2**64-1):
    calls={'destroy':(), 'suspend':(), 'resume':(0,), 'set_action':(0,1),
           'key':(57,), 'gamepad':(0,0,1), 'set_presentation':(1,), 'release':(), 'plunger_delta':(8,), 'plunger_fire':(),
           'advance':(0,None,None), 'frame':(C.byref(pixel),C.byref(w),C.byref(h),C.byref(stride)),
           'state':(C.byref(tick),C.byref(mode),C.byref(table),C.byref(flags))}
    for name,args in calls.items():
        assert getattr(lib,'pf_engine_'+name)(handle,*args)==-1,name
assert lib.pf_engine_frame(0,None,None,None,None)==-1
assert lib.pf_engine_state(0,None,None,None,None)==-1
print('PASS real C shared library: ABI 1, bounded load errors, null outs and every invalid-handle export')
print('UNVERIFIED: successful original-backed gameplay conformance requires the separate four-table replay')
