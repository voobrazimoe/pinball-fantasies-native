//go:build windows

package platform

import (
	"unsafe"
)

type rawInputDevice struct {
	Page, Usage uint16
	Flags       uint32
	Target      uintptr
}
type rawInputHeader struct {
	Type, Size    uint32
	Device, Param uintptr
}
type rawMouse struct {
	Flags uint16
	// RAWMOUSE's button union is DWORD-aligned.
	Padding    uint16
	Buttons    uint32
	RawButtons uint32
	X, Y       int32
	Extra      uint32
}

func (h *hostWindow) registerMouse() error {
	d := rawInputDevice{Page: 1, Usage: 2, Target: h.hwnd}
	ok, _, err := up("RegisterRawInputDevices").Call(uintptr(unsafe.Pointer(&d)), 1, unsafe.Sizeof(d))
	if ok == 0 {
		return apiError("RegisterRawInputDevices", err)
	}
	return nil
}
func (h *hostWindow) MouseActive(active bool) { h.mouseActive = active }
func (h *hostWindow) rawMouse(handle uintptr) {
	var packet struct {
		Header rawInputHeader
		Mouse  rawMouse
	}
	size := uint32(unsafe.Sizeof(packet))
	n, _, _ := up("GetRawInputData").Call(handle, 0x10000003, uintptr(unsafe.Pointer(&packet)), uintptr(unsafe.Pointer(&size)), unsafe.Sizeof(packet.Header))
	if uint32(n) == 0xffffffff || n < uintptr(unsafe.Sizeof(packet)) || packet.Header.Type != 0 {
		return
	}
	if !h.focused || !h.mouseActive {
		h.mouseDown = false
		return
	}
	m := packet.Mouse
	// Absolute devices (e.g. tablets/RDP) are not relative plunger devices.
	if m.Flags&1 == 0 && m.Y != 0 {
		h.events = append(h.events, hostEvent{kind: 6, mouseY: int(m.Y)})
	}
	buttons := uint16(m.Buttons)
	if buttons&1 != 0 && !h.mouseDown {
		h.events = append(h.events, hostEvent{kind: 7})
		h.mouseDown = true
	}
	if buttons&2 != 0 {
		h.mouseDown = false
	}
}
