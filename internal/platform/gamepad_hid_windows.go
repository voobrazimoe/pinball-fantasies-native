//go:build windows

package platform

import (
	"pinballfantasies/internal/gamepad"
	"unsafe"
)

// Sony controllers (DualSense, DualShock 4) are HID gamepads, not XInput
// devices. Raw input delivers their reports to the window in the foreground,
// so they work without Steam Input, DS4Windows or SDL.

const (
	rawInputHID    = 2
	ridevDevNotify = 0x2000
	ridiDeviceInfo = 0x2000000b
	gidcRemoval    = 2
	ridInput       = 0x10000003
	ridHeader      = 0x10000005
)

type hidGamepad struct {
	device  uintptr // raw input handle of the Sony controller in use
	product uint16
	state   gamepad.Snapshot
	present bool
	known   map[uintptr]uint16 // device handle -> Sony product id, 0 for other devices
	buffer  []byte
}

type ridDeviceInfoHID struct {
	Size, Type       uint32
	Vendor, Product  uint32
	Version          uint32
	UsagePage, Usage uint16
	_                [8]byte // rest of the RID_DEVICE_INFO union
}

func (h *hostWindow) registerGamepads() {
	// Generic Desktop / Game Pad. Failure only loses native Sony support;
	// XInput controllers are polled independently.
	d := rawInputDevice{Page: 1, Usage: 5, Flags: ridevDevNotify, Target: h.hwnd}
	up("RegisterRawInputDevices").Call(uintptr(unsafe.Pointer(&d)), 1, unsafe.Sizeof(d))
}

func (p *hidGamepad) sonyProduct(device uintptr) uint16 {
	if product, ok := p.known[device]; ok {
		return product
	}
	if p.known == nil {
		p.known = map[uintptr]uint16{}
	}
	var info ridDeviceInfoHID
	info.Size = uint32(unsafe.Sizeof(info))
	size := info.Size
	n, _, _ := up("GetRawInputDeviceInfoW").Call(device, ridiDeviceInfo, uintptr(unsafe.Pointer(&info)), uintptr(unsafe.Pointer(&size)))
	var product uint16
	if int32(n) > 0 && info.Type == rawInputHID && gamepad.SonyController(uint16(info.Vendor), uint16(info.Product)) {
		product = uint16(info.Product)
	}
	p.known[device] = product
	return product
}

// rawInput dispatches WM_INPUT by device type.
func (h *hostWindow) rawInput(handle uintptr) {
	var hdr rawInputHeader
	size := uint32(unsafe.Sizeof(hdr))
	n, _, _ := up("GetRawInputData").Call(handle, ridHeader, uintptr(unsafe.Pointer(&hdr)), uintptr(unsafe.Pointer(&size)), unsafe.Sizeof(hdr))
	if uint32(n) == 0xffffffff {
		return
	}
	switch hdr.Type {
	case 0:
		h.rawMouse(handle)
	case rawInputHID:
		h.rawHID(handle)
	}
}

// rawHID reads one HID packet from a Sony controller.
func (h *hostWindow) rawHID(handle uintptr) {
	p := &h.gamepad.hid
	var size uint32
	header := unsafe.Sizeof(rawInputHeader{})
	up("GetRawInputData").Call(handle, ridInput, 0, uintptr(unsafe.Pointer(&size)), header)
	if size < uint32(header)+8 || size > 1<<16 {
		return
	}
	if cap(p.buffer) < int(size) {
		p.buffer = make([]byte, size)
	}
	buf := p.buffer[:size]
	n, _, _ := up("GetRawInputData").Call(handle, ridInput, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), header)
	if uint32(n) == 0xffffffff || uint32(n) < uint32(header)+8 {
		return
	}
	hdr := (*rawInputHeader)(unsafe.Pointer(&buf[0]))
	if hdr.Type != rawInputHID {
		return
	}
	// One device at a time: the first Sony controller that reports keeps
	// ownership until it is removed.
	if p.present && hdr.Device != p.device {
		return
	}
	product := p.sonyProduct(hdr.Device)
	if product == 0 {
		return
	}
	// RAWHID: dwSizeHid, dwCount, then dwCount reports of dwSizeHid bytes.
	body := buf[header:n]
	reportSize := *(*uint32)(unsafe.Pointer(&body[0]))
	count := *(*uint32)(unsafe.Pointer(&body[4]))
	data := body[8:]
	if reportSize == 0 || uint64(reportSize)*uint64(count) > uint64(len(data)) {
		return
	}
	for i := uint32(0); i < count; i++ {
		report := data[i*reportSize : (i+1)*reportSize]
		if s, ok := gamepad.SonySnapshot(product, report); ok {
			p.device, p.product, p.state, p.present = hdr.Device, product, s, true
		}
	}
}

// hidDeviceChange handles WM_INPUT_DEVICE_CHANGE.
func (h *hostWindow) hidDeviceChange(change, device uintptr) {
	p := &h.gamepad.hid
	if change != gidcRemoval {
		return
	}
	delete(p.known, device)
	if p.present && device == p.device {
		p.present = false
		p.device = 0
		p.state = gamepad.Snapshot{}
	}
}
