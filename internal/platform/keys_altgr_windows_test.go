//go:build windows

package platform

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestWin32AltGrQueuedControl(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Exercise windowProc's actual PeekMessage/GetMessageTime path using a
	// thread queue, without injecting keys into the user's desktop.
	var m, next message
	inputPeekMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0x100, 0x105, 0)
	id, _, _ := kernel32.NewProc("GetCurrentThreadId").Call()
	post := up("PostThreadMessageW")
	ctrl := uintptr(0x1d<<16 | 1)
	alt := uintptr(0x38<<16 | 1<<24 | 1)
	for attempt := 0; attempt < 10; attempt++ {
		for _, key := range []struct{ msg, vk, l uintptr }{{0x100, 0x11, ctrl}, {0x104, 0x12, alt}} {
			if ok, _, err := post.Call(id, key.msg, key.vk, key.l); ok == 0 {
				t.Fatal(err)
			}
		}
		inputPeekMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0x100, 0x105, 1)
		inputPeekMessage.Call(uintptr(unsafe.Pointer(&next)), 0, 0x100, 0x105, 0)
		if m.Time != next.Time {
			inputPeekMessage.Call(uintptr(unsafe.Pointer(&next)), 0, 0x100, 0x105, 1)
			continue // Posts crossed a millisecond boundary; retry the pair.
		}
		h := &hostWindow{}
		previous := activeWindow
		activeWindow = h
		windowProc(0, m.ID, m.WParam, m.LParam)
		ctrlDown := h.input.down[0xa2]
		inputPeekMessage.Call(uintptr(unsafe.Pointer(&next)), 0, 0x100, 0x105, 1)
		windowProc(0, next.ID, next.WParam, next.LParam)
		activeWindow = previous
		if ctrlDown || !h.input.down[0xa5] || h.input.held() != 2 {
			t.Fatalf("queued AltGr: synthetic Ctrl=%t held=%d", ctrlDown, h.input.held())
		}
		return
	}
	t.Fatal("could not post a same-timestamp AltGr pair")
}
