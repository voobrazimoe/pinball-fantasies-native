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
func (h *hostWindow) MouseActive(active bool) {
	if h.mouseActive == active {
		return
	}
	h.mouseActive = active
	h.refreshMouseCursor()
}

// WM_SETCURSOR selects a cursor locally; it neither captures the mouse nor
// changes ShowCursor's process display counter. Mode changes also refresh a
// stationary pointer, but only when this window owns its client hit.
func (h *hostWindow) selectMouseCursor(hit uint16) bool {
	if hit != 1 || !h.focused || !h.mouseActive {
		return false
	}
	up("SetCursor").Call(0)
	return true
}

func (h *hostWindow) refreshMouseCursor() {
	var p point
	if ok, _, _ := up("GetCursorPos").Call(uintptr(unsafe.Pointer(&p))); ok == 0 {
		return
	}
	packed := uintptr(uint64(uint32(p.X)) | uint64(uint32(p.Y))<<32)
	owner, _, _ := up("WindowFromPoint").Call(packed)
	if owner != h.hwnd {
		return
	}
	up("ScreenToClient").Call(h.hwnd, uintptr(unsafe.Pointer(&p)))
	var r rect
	up("GetClientRect").Call(h.hwnd, uintptr(unsafe.Pointer(&r)))
	if p.X < r.Left || p.X >= r.Right || p.Y < r.Top || p.Y >= r.Bottom {
		return
	}
	if !h.selectMouseCursor(1) {
		cursor, _, _ := up("LoadCursorW").Call(0, 32512)
		up("SetCursor").Call(cursor)
	}
}
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
