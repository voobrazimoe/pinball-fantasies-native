// Package main builds with -buildmode=c-archive or c-shared. No window backend.
package main

/*
#include "abi.h"
#include <stdlib.h>
#include <string.h>
static void pf_deliver(pf_pcm_sink sink, void *ctx, const uint8_t *p, uint32_t n) {
 if (sink) sink(ctx,p,n);
}
*/
import "C"
import (
	"pinballfantasies/internal/engine"
	"sync"
	"unsafe"
)

type instance struct {
	mu       sync.Mutex
	engine   *engine.Engine
	frame    unsafe.Pointer
	capacity int
}

var registry = struct {
	sync.Mutex
	next      uint64
	instances map[uint64]*instance
}{instances: make(map[uint64]*instance)}

func invoke(handle C.uint64_t, f func(*instance) C.int32_t) C.int32_t {
	registry.Lock()
	i := registry.instances[uint64(handle)]
	registry.Unlock()
	if i == nil {
		return C.PF_INVALID
	}
	if !i.mu.TryLock() {
		return C.PF_BUSY
	}
	defer i.mu.Unlock()
	if i.engine == nil {
		return C.PF_INVALID
	}
	return f(i)
}
func result(err error) C.int32_t {
	if err != nil {
		return C.PF_ERROR
	}
	return C.PF_OK
}

//export pf_engine_abi_version
func pf_engine_abi_version() C.int32_t { return C.PF_ABI_VERSION }

//export pf_engine_create
func pf_engine_create(data, state *C.char, ns C.int64_t, message *C.char, capacity C.uint32_t) C.uint64_t {
	if message != nil && capacity > 0 {
		*message = 0
	}
	if data == nil || state == nil {
		return 0
	}
	e, err := engine.Load(C.GoString(data), C.GoString(state), int64(ns))
	if err != nil {
		if message != nil && capacity > 0 {
			b := unsafe.Slice((*byte)(unsafe.Pointer(message)), int(capacity))
			n := copy(b[:len(b)-1], err.Error())
			b[n] = 0
		}
		return 0
	}
	registry.Lock()
	defer registry.Unlock()
	registry.next++
	h := registry.next
	registry.instances[h] = &instance{engine: e}
	return C.uint64_t(h)
}

//export pf_engine_destroy
func pf_engine_destroy(h C.uint64_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t {
		err := i.engine.Close()
		i.engine = nil
		C.free(i.frame)
		i.frame = nil
		registry.Lock()
		delete(registry.instances, uint64(h))
		registry.Unlock()
		return result(err)
	})
}

//export pf_engine_suspend
func pf_engine_suspend(h C.uint64_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t { return result(i.engine.Suspend()) })
}

//export pf_engine_resume
func pf_engine_resume(h C.uint64_t, ns C.int64_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t { return result(i.engine.Resume(int64(ns))) })
}

//export pf_engine_set_action
func pf_engine_set_action(h C.uint64_t, a C.uint32_t, down C.int32_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t { return result(i.engine.SetAction(engine.Action(a), down != 0)) })
}

//export pf_engine_key
func pf_engine_key(h C.uint64_t, k C.uint8_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t { i.engine.Key(uint8(k)); return C.PF_OK })
}

//export pf_engine_release
func pf_engine_release(h C.uint64_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t { i.engine.Release(); return C.PF_OK })
}

//export pf_engine_plunger_delta
func pf_engine_plunger_delta(h C.uint64_t, d C.int32_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t { i.engine.PlungerDelta(int32(d)); return C.PF_OK })
}

//export pf_engine_plunger_fire
func pf_engine_plunger_fire(h C.uint64_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t { i.engine.PlungerFire(); return C.PF_OK })
}

//export pf_engine_advance
func pf_engine_advance(h C.uint64_t, ns C.int64_t, sink C.pf_pcm_sink, ctx unsafe.Pointer) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t {
		return result(i.engine.Advance(int64(ns), func(p []byte) error {
			if len(p) > 0 {
				C.pf_deliver(sink, ctx, (*C.uint8_t)(unsafe.Pointer(&p[0])), C.uint32_t(len(p)))
			}
			return nil
		}))
	})
}

//export pf_engine_frame
func pf_engine_frame(h C.uint64_t, p **C.uint8_t, w, height, stride *C.int32_t) C.int32_t {
	if p == nil || w == nil || height == nil || stride == nil {
		return C.PF_INVALID
	}
	return invoke(h, func(i *instance) C.int32_t {
		frame := i.engine.Frame()
		n := len(frame.Pix)
		if n > i.capacity {
			buf := C.realloc(i.frame, C.size_t(n))
			if buf == nil {
				return C.PF_ERROR
			}
			i.frame = buf
			i.capacity = n
		}
		if n > 0 {
			C.memcpy(i.frame, unsafe.Pointer(&frame.Pix[0]), C.size_t(n))
		}
		*p = (*C.uint8_t)(i.frame)
		*w = C.int32_t(frame.Rect.Dx())
		*height = C.int32_t(frame.Rect.Dy())
		*stride = C.int32_t(frame.Stride)
		return C.PF_OK
	})
}

//export pf_engine_state
func pf_engine_state(h C.uint64_t, tick *C.uint64_t, mode, table, flags *C.uint32_t) C.int32_t {
	if tick == nil || mode == nil || table == nil || flags == nil {
		return C.PF_INVALID
	}
	return invoke(h, func(i *instance) C.int32_t {
		s := i.engine.State()
		*tick = C.uint64_t(s.Tick)
		*mode = C.uint32_t(s.Mode)
		*table = C.uint32_t(s.Table)
		*flags = 0
		if s.Suspended {
			*flags |= 1
		}
		if s.Done {
			*flags |= 2
		}
		if s.MouseActive {
			*flags |= 4
		}
		return C.PF_OK
	})
}
func main() {}

//export pf_engine_set_presentation
func pf_engine_set_presentation(h C.uint64_t, full C.int32_t) C.int32_t {
	if full != 0 && full != 1 {
		return C.PF_INVALID
	}
	return invoke(h, func(i *instance) C.int32_t { i.engine.SetPresentation(full == 1); return C.PF_OK })
}
