//go:build windows

package platform

import (
	"os"
	"pinballfantasies/internal/diagnostics"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

// Display pacing for the GDI host. PF12_VSYNC selects it:
//   - "vblank" (default): wait for the monitor's vertical blank through the
//     kernel display driver, then a fixed offset (PF12_VSYNC_OFFSET_MS,
//     default 3), then draw. DWM composes right after the blank, so drawing a
//     little later lands every frame in the same composition instead of
//     racing it.
//   - "dwm": draw, then wait in DwmFlush for the composition to take it.
//   - "off": the old path, drawing at each source tick and sleeping.
//
// Source ticks keep their own 60/71 Hz deadlines in every mode. Any failure
// falls back to the old path.
type displayWait struct {
	mode    string
	adapter uint32
	source  uint32
	monitor uintptr
	open    bool
	failed  bool
	offset  time.Duration
}

var (
	dwmFlush       = cachedProc(syscall.NewLazyDLL("dwmapi.dll"), "DwmFlush")
	kmtOpenFromHdc = cachedProc(gdi32, "D3DKMTOpenAdapterFromHdc")
	kmtWaitVBlank  = cachedProc(gdi32, "D3DKMTWaitForVerticalBlankEvent")
	kmtClose       = cachedProc(gdi32, "D3DKMTCloseAdapter")
)

type kmtOpenAdapterFromHdc struct {
	DC          uintptr
	Adapter     uint32
	LUIDLow     uint32
	LUIDHigh    int32
	VidPnSource uint32
}
type kmtWaitForVerticalBlank struct{ Adapter, Device, VidPnSource uint32 }
type kmtCloseAdapter struct{ Adapter uint32 }
type monitorInfoEx struct {
	monitorInfo
	Device [32]uint16
}

func (h *hostWindow) vsyncMode() string {
	if h.vsync.mode == "" {
		h.vsync.mode = os.Getenv("PF12_VSYNC")
		if h.vsync.mode != "dwm" && h.vsync.mode != "off" {
			h.vsync.mode = "vblank"
		}
		if h.noPresent {
			h.vsync.mode = "off"
		}
		h.vsync.offset = 3 * time.Millisecond
		if ms, e := strconv.Atoi(os.Getenv("PF12_VSYNC_OFFSET_MS")); e == nil && ms >= 0 && ms <= 12 {
			h.vsync.offset = time.Duration(ms) * time.Millisecond
		}
		diagnostics.Printf("PF12 vsync=%s offset=%s\n", h.vsync.mode, h.vsync.offset)
	}
	return h.vsync.mode
}

// WaitBeforePresent blocks until the window's monitor starts a vertical blank.
func (h *hostWindow) WaitBeforePresent() bool {
	if h.vsyncMode() != "vblank" || h.vsync.failed {
		return false
	}
	if !h.openVBlank() {
		return false
	}
	start := time.Now()
	w := kmtWaitForVerticalBlank{Adapter: h.vsync.adapter, VidPnSource: h.vsync.source}
	if st, _, _ := kmtWaitVBlank.Call(uintptr(unsafe.Pointer(&w))); st != 0 {
		h.closeVBlank()
		return false
	}
	if time.Since(start) >= 100*time.Millisecond {
		return false
	}
	time.Sleep(h.vsync.offset)
	return true
}

// WaitAfterPresent blocks until DWM has composed the frame just drawn.
func (h *hostWindow) WaitAfterPresent() bool {
	if h.vsyncMode() != "dwm" || dwmFlush.Find() != nil {
		return false
	}
	start := time.Now()
	hr, _, _ := dwmFlush.Call()
	return int32(hr) >= 0 && time.Since(start) >= time.Millisecond
}

func (h *hostWindow) openVBlank() bool {
	mon, _, _ := up("MonitorFromWindow").Call(h.hwnd, 2)
	if h.vsync.open && mon == h.vsync.monitor {
		return true
	}
	h.closeVBlank()
	if kmtOpenFromHdc.Find() != nil || kmtWaitVBlank.Find() != nil {
		h.vsync.failed = true
		return false
	}
	mi := monitorInfoEx{}
	mi.Size = uint32(unsafe.Sizeof(mi))
	if v, _, _ := up("GetMonitorInfoW").Call(mon, uintptr(unsafe.Pointer(&mi))); v == 0 {
		return false
	}
	dc, _, _ := gdi32.NewProc("CreateDCW").Call(uintptr(unsafe.Pointer(&mi.Device[0])), uintptr(unsafe.Pointer(&mi.Device[0])), 0, 0)
	if dc == 0 {
		return false
	}
	defer gdi32.NewProc("DeleteDC").Call(dc)
	o := kmtOpenAdapterFromHdc{DC: dc}
	if st, _, _ := kmtOpenFromHdc.Call(uintptr(unsafe.Pointer(&o))); st != 0 {
		diagnostics.Printf("PF12 vsync: D3DKMTOpenAdapterFromHdc status %#x, falling back\n", uint32(st))
		h.vsync.failed = true
		return false
	}
	h.vsync.adapter, h.vsync.source, h.vsync.monitor, h.vsync.open = o.Adapter, o.VidPnSource, mon, true
	return true
}

func (h *hostWindow) closeVBlank() {
	if h.vsync.open {
		c := kmtCloseAdapter{Adapter: h.vsync.adapter}
		kmtClose.Call(uintptr(unsafe.Pointer(&c)))
		h.vsync.open = false
	}
}
