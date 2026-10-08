//go:build demodev

package main

/*
#include "abi.h"
*/
import "C"
import (
	"pinballfantasies/internal/engine"
	"pinballfantasies/internal/frontend"
	"unsafe"
)

func demoMessage(message *C.char, capacity C.uint32_t, value string) {
	if message != nil && capacity > 0 {
		b := unsafe.Slice((*byte)(unsafe.Pointer(message)), int(capacity))
		n := copy(b[:len(b)-1], value)
		b[n] = 0
	}
}

//export pf_engine_create_demo
func pf_engine_create_demo(demo, canonical, log *C.char, ns C.int64_t, message *C.char, capacity C.uint32_t) C.uint64_t {
	demoMessage(message, capacity, "")
	if demo == nil || canonical == nil || log == nil || ns < 0 {
		return 0
	}
	rt, err := frontend.LoadExperimentalDemo(C.GoString(demo), C.GoString(canonical), C.GoString(log))
	if err != nil {
		demoMessage(message, capacity, err.Error())
		return 0
	}
	registry.Lock()
	defer registry.Unlock()
	registry.next++
	h := registry.next
	registry.instances[h] = &instance{engine: engine.New(rt, int64(ns))}
	return C.uint64_t(h)
}

//export pf_engine_demo_diagnostic
func pf_engine_demo_diagnostic(h C.uint64_t, message *C.char, capacity C.uint32_t) C.int32_t {
	return invoke(h, func(i *instance) C.int32_t { demoMessage(message, capacity, i.engine.Diagnostic()); return C.PF_OK })
}
