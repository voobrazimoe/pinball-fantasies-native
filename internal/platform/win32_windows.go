//go:build windows

package platform

import (
	"fmt"
	"image"
	"os"
	"pinballfantasies/internal/diagnostics"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

var user32 = syscall.NewLazyDLL("user32.dll")
var gdi32 = syscall.NewLazyDLL("gdi32.dll")
var kernel32 = syscall.NewLazyDLL("kernel32.dll")

func up(name string) *syscall.LazyProc { return user32.NewProc(name) }

var defWindowProc = up("DefWindowProcW")
var windowCallback = syscall.NewCallback(windowProc)
var activeWindow *hostWindow // UI thread only; one window per application session.

type rect struct{ Left, Top, Right, Bottom int32 }
type point struct{ X, Y int32 }
type message struct {
	HWND           uintptr
	ID             uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             point
	Private        uint32
}
type windowClass struct {
	Size, Style                        uint32
	Proc                               uintptr
	ClassExtra, WindowExtra            int32
	Instance, Icon, Cursor, Background uintptr
	Menu, Name                         *uint16
	SmallIcon                          uintptr
}
type placement struct {
	Size, Flags, Show uint32
	Min, Max          point
	Normal            rect
}
type monitorInfo struct {
	Size          uint32
	Monitor, Work rect
	Flags         uint32
}
type bitmapInfo struct {
	Size                   uint32
	Width, Height          int32
	Planes, Bits           uint16
	Compression, ImageSize uint32
	XPels, YPels           int32
	Used, Important        uint32
}
type paintStruct struct {
	DC                 uintptr
	Erase              int32
	Paint              rect
	Restore, IncUpdate int32
	Reserved           [32]byte
}
type hostWindow struct {
	surfaceDC, surfaceBitmap, originalBitmap uintptr
	surfaceSize                              image.Point
	dirty, noPresent                         bool
	surfaceNeedsClear                        bool
	paints, erases, blits                    uint64
	savedRect                                rect
	hwnd                                     uintptr
	presentation                             WindowPresentation
	input                                    windowsKeys
	focused                                  bool
	mouseActive, mouseDown                   bool
	inputLog                                 *os.File
	events                                   []hostEvent
	pixels                                   []byte
	size                                     image.Point
	saved                                    placement
	style                                    uintptr
	err                                      error
	modalTick                                func() error
	sourceTick                               func() error
	changingStyle                            bool
	transition                               *fullscreenTransition
	transitionSequence                       uint64
	transitionDrawPending                    bool
	transitionDrawUntil                      time.Time
	transitionSettlingUntil                  time.Time
	drawing                                  bool
	lastDrawCall                             string
	transitionDevice                         *AudioDevice
}

func utf(s string) *uint16                { p, _ := syscall.UTF16PtrFromString(s); return p }
func apiError(name string, e error) error { return fmt.Errorf("Win32 %s: %w", name, e) }
func openHost(first *image.RGBA) (_ *hostWindow, err error) {
	dpiPolicy, err := setDPIAwareness()
	if err != nil {
		return nil, err
	}
	instance, _, e := kernel32.NewProc("GetModuleHandleW").Call(0)
	if instance == 0 {
		return nil, apiError("GetModuleHandleW", e)
	}
	name := utf("PinballFantasiesPF12")
	cursor, _, e := up("LoadCursorW").Call(0, 32512)
	if cursor == 0 {
		return nil, apiError("LoadCursorW", e)
	}
	cls := windowClass{Size: uint32(unsafe.Sizeof(windowClass{})), Proc: windowCallback, Instance: instance, Cursor: cursor, Name: name}
	if v, _, e := up("RegisterClassExW").Call(uintptr(unsafe.Pointer(&cls))); v == 0 {
		return nil, apiError("RegisterClassExW", e)
	}

	defer func() {
		if err != nil {
			up("UnregisterClassW").Call(uintptr(unsafe.Pointer(name)), instance)
		}
	}()
	mon, _, _ := up("MonitorFromPoint").Call(0, 1)
	mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	if v, _, e := up("GetMonitorInfoW").Call(mon, uintptr(unsafe.Pointer(&mi))); v == 0 {
		return nil, apiError("GetMonitorInfoW", e)
	}
	initial := initialWindowSize(image.Pt(int(mi.Work.Right-mi.Work.Left), int(mi.Work.Bottom-mi.Work.Top)))
	bounds := rect{Right: int32(initial.X), Bottom: int32(initial.Y)}
	const style = 0x00cf0000
	if v, _, e := up("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&bounds)), style, 0, 0); v == 0 {
		return nil, apiError("AdjustWindowRectEx", e)
	}
	h := &hostWindow{style: style, noPresent: os.Getenv("PF12_NO_PRESENT") == "1", size: first.Rect.Size(), pixels: frameBGRA(nil, first), dirty: true}
	h.openInputLog()
	defer func() {
		if err != nil && h.inputLog != nil {
			h.inputLog.Close()
		}
	}()
	activeWindow = h
	hwnd, _, e := up("CreateWindowExW").Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(utf("Pinball Fantasies"))), style, 0x80000000, 0x80000000, uintptr(bounds.Right-bounds.Left), uintptr(bounds.Bottom-bounds.Top), 0, 0, instance, 0)
	if hwnd == 0 {
		activeWindow = nil
		return nil, apiError("CreateWindowExW", e)
	}
	h.hwnd = hwnd
	h.presentation.backend = h
	if err := h.registerMouse(); err != nil {
		h.Close()
		return nil, err
	}
	up("ShowWindow").Call(hwnd, 5)
	diagnostics.Printf("Win32 window opened HWND=%#x; %s\n", hwnd, dpiPolicy)
	return h, nil
}
func (h *hostWindow) Close() {
	h.MouseActive(false)
	if h.inputLog != nil {
		defer h.inputLog.Close()
	}
	if v, _, e := up("DestroyWindow").Call(h.hwnd); v == 0 {
		fmt.Fprintln(os.Stderr, apiError("DestroyWindow", e))
	}
	activeWindow = nil
	if e := h.destroySurface(); e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
	diagnostics.Printf("Win32 repaint: paints=%d erases=%d complete blits=%d\n", h.paints, h.erases, h.blits)
	instance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	if v, _, e := up("UnregisterClassW").Call(uintptr(unsafe.Pointer(utf("PinballFantasiesPF12"))), instance); v == 0 {
		fmt.Fprintln(os.Stderr, apiError("UnregisterClassW", e))
	}
}
func windowProc(hwnd uintptr, msg uint32, w, l uintptr) uintptr {
	h := activeWindow
	if h != nil {
		if h.transition != nil {
			if name := transitionMessage(msg); name != "" {
				start := time.Now()
				defer func() { h.transition.event(name, time.Since(start)) }()
			}
		}
		switch msg {
		case 0x20: // WM_SETCURSOR: low word is the hit-test result.
			if w == hwnd && h.selectMouseCursor(uint16(l)) {
				return 1
			}
		case 0x10:
			h.events = append(h.events, hostEvent{kind: 1})
			return 0 // WM_CLOSE: shared lifecycle owns shutdown
		case 0x100, 0x104, 0x101, 0x105:
			// Reconcile before decoding Enter as well as after pumping: a lost Alt
			// break must not turn the next ordinary Enter into a shortcut.
			h.refreshInput("key", uint32(w), l, msg == 0x100 || msg == 0x104)
			if (w == 0x11 || w == 0xa2) && l&(1<<24) == 0 {
				// Inspect without removing the next message. This runs in the
				// window procedure so modal Windows message pumps work too.
				stamp, _, _ := inputMessageTime.Call()
				var next message
				found, _, _ := inputPeekMessage.Call(uintptr(unsafe.Pointer(&next)), 0, 0, 0, 0)
				if found != 0 && next.HWND == hwnd && windowsAltGrControl(uint32(w), l, uint32(stamp), next.ID, uint32(next.WParam), next.LParam, next.Time) {
					h.refreshInput("altgr-control-ignored", uint32(w), l, msg == 0x100 || msg == 0x104)
					return 0
				}
			}
			h.events = append(h.events, h.input.key(uint32(w), l, msg == 0x100 || msg == 0x104)...)
			return 0 // consume system keys too; Alt remains a flipper, Alt+Enter shared shortcut
		case 0xff: // WM_INPUT: unscaled hardware-relative mouse motion.
			h.rawMouse(l)
			// DefWindowProc must clean up foreground raw-input resources.
		case 0x7: // WM_SETFOCUS
			h.focused = true
			h.refreshMouseCursor()
			h.events = append(h.events, hostEvent{kind: 8})
			return 0
		case 0x8:
			h.focused = false
			h.refreshMouseCursor()
			h.mouseDown = false
			h.input = windowsKeys{}
			h.refreshInput("focus-loss", 0, 0, false)
			keep := h.events[:0]
			for _, e := range h.events {
				if e.kind == 1 {
					keep = append(keep, e)
				}
			}
			h.events = append(keep, hostEvent{kind: 5})
			return 0

		case 0x14: // WM_ERASEBKGND: the complete offscreen frame owns the background.
			h.erases++
			return 1
		case 0xf:
			h.paints++
			if h.changingStyle {
				// Validate intermediate paints without touching the visible DC or
				// rebuilding the surface. One invalidation follows the transaction.
				h.transitionCall("ValidateRect", hwnd, 0)
				return 0
			}
			if h.err == nil {
				h.err = h.serviceSources()
			}
			var ps paintStruct
			dc, _, e := h.drawHostCall("BeginPaint", hwnd, uintptr(unsafe.Pointer(&ps)))
			if dc == 0 {
				h.err = apiError("BeginPaint", e)
				up("ValidateRect").Call(hwnd, 0)
				return 0
			}
			if h.hwnd != 0 && h.err == nil {
				h.err = h.drawDC(dc)
			}
			if v, _, e := h.drawHostCall("EndPaint", hwnd, uintptr(unsafe.Pointer(&ps))); v == 0 {
				h.err = apiError("EndPaint", e)
			}
			return 0
		case 0x5:
			h.dirty = true
			if !h.changingStyle {
				up("InvalidateRect").Call(hwnd, 0, 0)
			}
			return 0
		case 0x231: // WM_ENTERSIZEMOVE: DefWindowProc owns a nested modal message loop.
			if h.modalTick != nil {
				if v, _, e := up("SetTimer").Call(hwnd, 1, 8, 0); v == 0 {
					h.err = apiError("resize SetTimer", e)
				}
			}
			return 0
		case 0x232:
			if h.modalTick != nil {
				if v, _, e := up("KillTimer").Call(hwnd, 1); v == 0 {
					h.err = apiError("resize KillTimer", e)
				}
			}
			return 0
		case 0x113:
			if w == 1 && h.modalTick != nil && h.err == nil {
				h.err = h.modalTick()
			}
			return 0
		case 0x2e0:
			if !h.presentation.IsFullscreen() && !h.changingStyle {
				r := (*rect)(unsafe.Pointer(l))
				if v, _, e := up("SetWindowPos").Call(hwnd, 0, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top), 0x14); v == 0 {
					h.err = apiError("DPI SetWindowPos", e)
				}
			}
			return 0
		}
	}
	v, _, _ := defWindowProc.Call(hwnd, uintptr(msg), w, l)
	return v
}
func (h *hostWindow) Pump() {
	defer h.refreshInput("pump", 0, 0, false)
	var m message
	for {
		// A borderless mutation can leave slow host/compositor work for the next
		// message pump. Submit already-due PCM before each such settling call.
		if time.Now().Before(h.transitionSettlingUntil) && h.err == nil {
			h.err = h.serviceSources()
			if h.err != nil {
				return
			}
		}
		v, _, _ := h.drawHostCall("PeekMessageW", uintptr(unsafe.Pointer(&m)), 0, 0, 0, 1)
		if v == 0 {
			break
		}
		if m.ID == 0x12 {
			h.events = append(h.events, hostEvent{kind: 1})
		} else {
			h.drawHostCall("DispatchMessageW", uintptr(unsafe.Pointer(&m)))
		}
	}
}
func (h *hostWindow) SetModalTick(tick func() error)  { h.modalTick = tick }
func (h *hostWindow) SetSourceTick(tick func() error) { h.sourceTick = tick }
func (h *hostWindow) serviceSources() error {
	if h.sourceTick != nil && !h.changingStyle {
		return h.sourceTick()
	}
	return nil
}
func (h *hostWindow) Event() hostEvent {
	if len(h.events) == 0 {
		return hostEvent{}
	}
	e := h.events[0]
	h.events = h.events[1:]
	return e
}
func (h *hostWindow) Held() int { return h.refreshInput("held", 0, 0, false) }
func (h *hostWindow) Present(f *image.RGBA) error {
	if h.err != nil {
		return h.err
	}
	if h.size != f.Rect.Size() {
		h.surfaceNeedsClear = true
		diagnostics.Printf("PF12 presentation %dx%d window ID=%d fullscreen=%t\n", f.Rect.Dx(), f.Rect.Dy(), h.hwnd, h.presentation.IsFullscreen())
	}
	h.size = f.Rect.Size()
	h.pixels = frameBGRA(h.pixels, f)
	h.dirty = true
	if err := h.draw(); err != nil {
		return err
	}
	if !h.noPresent {
		// GetDC presentation covered the entire client. Consume its invalidation
		// so the next Pump does not BeginPaint/blit the same resized frame again.
		if v, _, e := h.drawHostCall("ValidateRect", h.hwnd, 0); v == 0 {
			return apiError("present ValidateRect", e)
		}
	}
	return nil
}

func gdiOK(value uintptr) bool { return value != 0 && uint32(value) != 0xffffffff }
func (h *hostWindow) destroySurface() error {
	if h.surfaceDC == 0 {
		return nil
	}
	var failure error
	if v, _, e := gdi32.NewProc("SelectObject").Call(h.surfaceDC, h.originalBitmap); !gdiOK(v) {
		failure = apiError("restore surface bitmap", e)
	}
	if v, _, e := gdi32.NewProc("DeleteObject").Call(h.surfaceBitmap); v == 0 {
		failure = apiError("DeleteObject surface", e)
	}
	if v, _, e := gdi32.NewProc("DeleteDC").Call(h.surfaceDC); v == 0 {
		failure = apiError("DeleteDC surface", e)
	}
	h.surfaceDC, h.surfaceBitmap, h.originalBitmap = 0, 0, 0
	return failure
}
func (h *hostWindow) ensureSurface(dc uintptr, size image.Point) error {
	if h.surfaceDC != 0 && h.surfaceSize == size {
		return nil
	}
	start := time.Now()
	defer func() { h.surfaceTrace(start, size) }()
	if e := h.destroySurface(); e != nil {
		return e
	}
	mem, _, e := gdi32.NewProc("CreateCompatibleDC").Call(dc)
	if mem == 0 {
		return apiError("CreateCompatibleDC", e)
	}
	bitmap, _, e := gdi32.NewProc("CreateCompatibleBitmap").Call(dc, uintptr(size.X), uintptr(size.Y))
	if bitmap == 0 {
		gdi32.NewProc("DeleteDC").Call(mem)
		return apiError("CreateCompatibleBitmap", e)
	}
	original, _, e := gdi32.NewProc("SelectObject").Call(mem, bitmap)
	if !gdiOK(original) {
		gdi32.NewProc("DeleteObject").Call(bitmap)
		gdi32.NewProc("DeleteDC").Call(mem)
		return apiError("SelectObject surface", e)
	}
	h.surfaceDC, h.surfaceBitmap, h.originalBitmap = mem, bitmap, original
	h.surfaceSize = size
	h.dirty, h.surfaceNeedsClear = true, true
	return nil
}
func (h *hostWindow) draw() error {
	if h.noPresent {
		return nil
	}
	if err := h.serviceSources(); err != nil {
		return err
	}
	dc, _, e := h.drawHostCall("GetDC", h.hwnd)
	if dc == 0 {
		return apiError("GetDC", e)
	}
	defer h.drawHostCall("ReleaseDC", h.hwnd, dc)
	return h.drawDC(dc)
}
func (h *hostWindow) drawDC(dc uintptr) error {
	wasDrawing := h.drawing
	h.drawing = true
	defer func() { h.drawing = wasDrawing }()
	var calls []transitionTime
	measure := !h.changingStyle && time.Now().Before(h.transitionDrawUntil)
	var callTimes *[]transitionTime
	if measure {
		callTimes = &calls
	}

	if measure {
		first := h.transitionDrawPending
		h.transitionDrawPending = false
		start, pre := time.Now(), audioSnapshot(h.transitionDevice)
		defer func() {
			if e := writeTransition(struct {
				Kind       string
				First      bool
				Sequence   uint64
				DurationNS int64
				Pre, Post  transitionAudio
				Calls      []transitionTime
			}{"draw", first, h.transitionSequence, time.Since(start).Nanoseconds(), pre, audioSnapshot(h.transitionDevice), calls}); e != nil {
				h.err = e
			}
		}()
	}
	if h.noPresent || len(h.pixels) == 0 {
		return nil
	}
	var client rect
	if v, _, e := up("GetClientRect").Call(h.hwnd, uintptr(unsafe.Pointer(&client))); v == 0 {
		return apiError("GetClientRect", e)
	}
	size := image.Pt(int(client.Right), int(client.Bottom))
	if size.X <= 0 || size.Y <= 0 {
		return nil
	}
	if e := h.ensureSurface(dc, size); e != nil {
		return e
	}
	if err := h.serviceSources(); err != nil {
		return err
	}
	if h.dirty {
		// Neither the black letterbox clear nor the stretch touches the window DC.
		// Letterbox pixels persist until geometry/logical dimensions change.
		// Re-clearing a desktop-sized bitmap every frame is redundant GDI work.
		if h.surfaceNeedsClear {
			if v, _, e := h.drawCall(callTimes, "PatBlt", h.surfaceDC, 0, 0, uintptr(size.X), uintptr(size.Y), 0x42); v == 0 {
				return apiError("offscreen PatBlt", e)
			}
			h.surfaceNeedsClear = false
		}
		if v, _, e := h.drawCall(callTimes, "SetStretchBltMode", h.surfaceDC, 3); v == 0 {
			return apiError("SetStretchBltMode", e)
		}
		if err := h.serviceSources(); err != nil {
			return err
		}
		dst := aspectRect(size, logicalSize(h.size))
		bi := bitmapInfo{Size: 40, Width: int32(h.size.X), Height: -int32(h.size.Y), Planes: 1, Bits: 32}
		v, _, e := h.drawCall(callTimes, "StretchDIBits", h.surfaceDC, uintptr(dst.Min.X), uintptr(dst.Min.Y), uintptr(dst.Dx()), uintptr(dst.Dy()), 0, 0, uintptr(h.size.X), uintptr(h.size.Y), uintptr(unsafe.Pointer(&h.pixels[0])), uintptr(unsafe.Pointer(&bi)), 0, 0xcc0020)
		runtime.KeepAlive(h.pixels)
		if v == 0 || uint32(v) == 0xffffffff {
			return apiError("offscreen StretchDIBits", e)
		}
		h.dirty = false
	}
	if err := h.serviceSources(); err != nil {
		return err
	}
	if v, _, e := h.drawCall(callTimes, "BitBlt", dc, 0, 0, uintptr(size.X), uintptr(size.Y), h.surfaceDC, 0, 0, 0xcc0020); v == 0 {
		return apiError("BitBlt present", e)
	}
	h.blits++
	return h.serviceSources()
}
func (h *hostWindow) Geometry() WindowGeometry {
	var r rect
	if v, _, e := h.transitionCall("GetWindowRect", h.hwnd, uintptr(unsafe.Pointer(&r))); v == 0 {
		h.err = apiError("GetWindowRect", e)
	}
	return WindowGeometry{X: int(r.Left), Y: int(r.Top), Width: int(r.Right - r.Left), Height: int(r.Bottom - r.Top)}
}
func (h *hostWindow) SetFullscreen(on bool) error {
	var target rect
	if h.err != nil {
		return h.err
	}
	h.changingStyle = true
	defer func() {
		h.changingStyle = false
		h.dirty = true
		// Asynchronous paint; preserve pixels and the complete old surface until
		// normal Present/WM_PAINT rebuilds it at the final client dimensions.
		h.transitionCall("InvalidateRect", h.hwnd, 0, 0)
	}()
	style := h.style
	if on {
		if v, _, e := h.transitionCall("GetWindowRect", h.hwnd, uintptr(unsafe.Pointer(&h.savedRect))); v == 0 {
			return apiError("fullscreen GetWindowRect", e)
		}
		h.saved = placement{Size: uint32(unsafe.Sizeof(placement{}))}
		if v, _, e := h.transitionCall("GetWindowPlacement", h.hwnd, uintptr(unsafe.Pointer(&h.saved))); v == 0 {
			return apiError("GetWindowPlacement", e)
		}
		mon, _, _ := h.transitionCall("MonitorFromWindow", h.hwnd, 2)
		mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
		if v, _, e := h.transitionCall("GetMonitorInfoW", mon, uintptr(unsafe.Pointer(&mi))); v == 0 {
			return apiError("GetMonitorInfoW", e)
		}
		target = mi.Monitor
		old, _, e := h.transitionCall("GetWindowLongPtrW", h.hwnd, ^uintptr(15))
		if old == 0 {
			return apiError("GetWindowLongPtrW", e)
		}
		h.style = old &^ (0x01000000 | 0x20000000)
		if h.saved.Show == 3 || h.saved.Show == 2 {
			h.transitionCall("ShowWindow", h.hwnd, 9)
		}
		style = h.style &^ 0x00cf0000
	}
	if v, _, e := h.transitionCall("SetWindowLongPtrW", h.hwnd, ^uintptr(15), style); v == 0 {
		return apiError("SetWindowLongPtrW", e)
	}
	if !on {
		if v, _, e := h.transitionCall("SetWindowPlacement", h.hwnd, uintptr(unsafe.Pointer(&h.saved))); v == 0 {
			return apiError("SetWindowPlacement", e)
		}

		// WINDOWPLACEMENT retains maximize/normal ownership. A normal window also
		// restores the exact screen-pixel rectangle, without mixing workspace offsets
		// with monitor coordinates (taskbars may be on the left or top).
		// FRAMECHANGED | NOZORDER | NOACTIVATE | NOREDRAW; retain copy bits.
		args, flags := [4]uintptr{}, uintptr(0x3f)
		if h.saved.Show == 1 {
			args = windowRectArgs(h.savedRect)
			flags = 0x3c
		}
		if v, _, e := h.transitionCall("SetWindowPos", h.hwnd, 0, args[0], args[1], args[2], args[3], flags); v == 0 {
			return apiError("restore SetWindowPos", e)
		}
	} else {
		args := windowRectArgs(target)
		if v, _, e := h.transitionCall("SetWindowPos", h.hwnd, 0, args[0], args[1], args[2], args[3], 0x3c); v == 0 {
			return apiError("fullscreen SetWindowPos", e)
		}
	}
	return nil
}
func (h *hostWindow) RestoreGeometry(WindowGeometry) {} // WINDOWPLACEMENT restores normal and maximized geometry.

func (h *hostWindow) FullscreenTrace(clears, starts uint64, device *AudioDevice) {
	if device != nil {
		diagnostics.Printf("PF12 fullscreen audio lifecycle clears=%d/%d starts=%d/%d producer unchanged=true\n", clears, device.LifecycleClears, starts, device.Starts)
	}
}

func windowRectArgs(r rect) [4]uintptr {
	return [4]uintptr{uintptr(r.Left), uintptr(r.Top), uintptr(r.Right - r.Left), uintptr(r.Bottom - r.Top)}
}

// Prefer v2 on modern Windows; early Windows 10 supports per-monitor v1.
// Both policies use client pixels and WM_DPICHANGED, with no bitmap scaling.
var processDPIPolicy string

func setDPIAwareness() (policy string, err error) {
	if processDPIPolicy != "" {
		return processDPIPolicy, nil
	}
	defer func() {
		if err == nil {
			processDPIPolicy = policy
		}
	}()
	dpi := up("SetProcessDpiAwarenessContext")
	if e := dpi.Find(); e == nil {
		if ok, _, e := dpi.Call(^uintptr(3)); ok != 0 {
			return "per-monitor DPI v2", nil
		} else if e != syscall.Errno(87) {
			return "", apiError("SetProcessDpiAwarenessContext v2", e)
		}
		if ok, _, e := dpi.Call(^uintptr(2)); ok != 0 {
			return "per-monitor DPI v1", nil
		} else {
			return "", apiError("SetProcessDpiAwarenessContext v1", e)
		}
	}
	proc := syscall.NewLazyDLL("shcore.dll").NewProc("SetProcessDpiAwareness")
	if e := proc.Find(); e != nil {
		return "", apiError("per-monitor DPI awareness", e)
	}
	result, _, _ := proc.Call(2)
	if result != 0 {
		return "", fmt.Errorf("Win32 SetProcessDpiAwareness: HRESULT %#x", uint32(result))
	}
	return "per-monitor DPI v1", nil
}
