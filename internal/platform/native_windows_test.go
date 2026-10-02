//go:build windows

package platform

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/settings"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestWin32AMD64Layouts(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("PF12 primary target")
	}
	for name, tc := range map[string][2]uintptr{
		"PAINTSTRUCT": {unsafe.Sizeof(paintStruct{}), 72},
		"WNDCLASSEX":  {unsafe.Sizeof(windowClass{}), 80}, "MSG": {unsafe.Sizeof(message{}), 48},
		"WINDOWPLACEMENT": {unsafe.Sizeof(placement{}), 44}, "MONITORINFO": {unsafe.Sizeof(monitorInfo{}), 40},
		"BITMAPINFOHEADER": {unsafe.Sizeof(bitmapInfo{}), 40}, "WAVEHDR": {unsafe.Sizeof(waveHeader{}), 48},
	} {
		if tc[0] != tc[1] {
			t.Fatalf("%s size %d expected %d", name, tc[0], tc[1])
		}
	}
	if unsafe.Offsetof(waveHeader{}.Flags) != 24 || unsafe.Offsetof(windowClass{}.Name) != 64 {
		t.Fatal("native pointer/flag offsets")
	}
}

// Explicitly opt into interactive/audio tests on a desktop host. Compile-only
// cross tests do not count as execution of these tests.
func TestNativeWin32WindowSmoke(t *testing.T) {
	if os.Getenv("PF12_WINDOW_SMOKE") != "1" {
		t.Skip("set PF12_WINDOW_SMOKE=1 on a desktop")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	h, e := openHost(image.NewRGBA(image.Rect(0, 0, 320, 240)))
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	settle := func() {
		for i := 0; i < 30; i++ {
			h.Pump()
			time.Sleep(10 * time.Millisecond)
		}
	}
	hwnd := h.hwnd
	up("SetForegroundWindow").Call(hwnd)
	time.Sleep(100 * time.Millisecond)
	h.Pump()
	// Deliberately inject only a make into the history, with no physical key
	// down and no break. Exercise the real Held/GetAsyncKeyState path.
	for _, vk := range windowsHeldVKs {
		state, _, _ := asyncKeyState.Call(uintptr(vk))
		if state&0x8000 != 0 {
			t.Skip("release gameplay keys before the desktop input smoke")
		}
	}
	for _, vk := range windowsHeldVKs {
		l := uintptr(0)
		if vk == 40 {
			l = 1 << 24
		}
		h.input.key(vk, l, true)
		if h.input.held() == 0 || h.Held() != 0 || h.input.down[vk] {
			t.Fatalf("lost break remained held for VK=%#x", vk)
		}
	}
	for _, size := range []image.Point{image.Pt(640, 240), image.Pt(320, 240), image.Pt(320, 350), image.Pt(320, 609)} {
		f := image.NewRGBA(image.Rectangle{Max: size})
		for y := 0; y < size.Y; y++ {
			for x := 0; x < size.X; x++ {
				f.SetRGBA(x, y, color.RGBA{R: byte(x), G: byte(y), B: byte(x + y), A: 255})
			}
		}
		if e := h.Present(f); e != nil {
			t.Fatal(e)
		}
		if v, _, e := up("SetWindowPos").Call(hwnd, 0, 100, 100, 800, 600, 0x14); v == 0 {
			t.Fatal(e)
		}
		settle()
		before := h.Geometry()
		if e := h.presentation.ToggleFullscreen(); e != nil {
			t.Fatal(e)
		}
		settle()
		t.Logf("logical=%v saved placement show=%d flags=%d rect=%v screen=%v before=%v", size, h.saved.Show, h.saved.Flags, h.saved.Normal, h.savedRect, before)
		if e := h.Present(f); e != nil {
			t.Fatal(e)
		}
		if e := h.presentation.ToggleFullscreen(); e != nil {
			t.Fatal(e)
		}
		settle()
		if after := h.Geometry(); after != before {
			t.Fatalf("restore geometry %v != %v", after, before)
		}
		if h.hwnd != hwnd {
			t.Fatal("HWND changed")
		}
		// Force system repaint/erase without generating another logical frame.
		beforePaints := h.paints
		up("InvalidateRect").Call(hwnd, 0, 1)
		up("UpdateWindow").Call(hwnd)
		if h.paints <= beforePaints || h.surfaceBitmap == 0 || h.dirty {
			t.Fatal("retained complete frame not repainted")
		}
		if got := windowProc(hwnd, 0x14, 0, 0); got != 1 {
			t.Fatal("system background erase was not consumed")
		}
		// WM_KILLFOCUS clears both held controls and queued makes.
		h.input.key(0x10, 0, true)
		h.input.key(40, 0, true)
		windowProc(hwnd, 0x8, 0, 0)
		if h.Held() != 0 {
			t.Fatal("stuck controls after focus loss")
		}
	}
	up("ShowWindow").Call(hwnd, 3)
	h.Pump()
	if e := h.presentation.ToggleFullscreen(); e != nil {
		t.Fatal(e)
	}
	if e := h.presentation.ToggleFullscreen(); e != nil {
		t.Fatal(e)
	}
	var p placement
	p.Size = uint32(unsafe.Sizeof(p))
	if v, _, e := up("GetWindowPlacement").Call(hwnd, uintptr(unsafe.Pointer(&p))); v == 0 {
		t.Fatal(e)
	}
	if p.Show != 3 {
		t.Fatalf("maximized placement not restored: %d", p.Show)
	}
	t.Logf("one HWND=%#x; selector/NORMAL/HIGH/OFF resize/fullscreen/focus/maximize PASS", hwnd)
}
func TestNativeWaveOutSmoke(t *testing.T) {
	if os.Getenv("PF12_AUDIO_SMOKE") != "1" {
		t.Skip("set PF12_AUDIO_SMOKE=1 with an output device")
	}
	a, e := OpenAudio()
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	// Five seconds: distinct 440 Hz left / 660 Hz right at 48k S16 stereo.
	pcm := make([]byte, 480*4)
	for tick := 0; tick < 500; tick++ {
		for i := 0; i < 480; i++ {
			sample := tick*480 + i
			for channel, hz := range []float64{440, 660} {
				value := int16(6000 * math.Sin(2*math.Pi*hz*float64(sample)/48000))
				binary.LittleEndian.PutUint16(pcm[i*4+channel*2:], uint16(value))
			}
		}
		if tick == 200 {
			a.Suspend(true)
			if a.err != nil {
				t.Fatal(a.err)
			}
			time.Sleep(100 * time.Millisecond)
			a.Suspend(false)
		}
		if e := a.Queue(pcm); e != nil {
			t.Fatal(e)
		}
		if len(a.blocks) > 32 {
			t.Fatal("unbounded queue")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if a.Starts < 2 {
		t.Fatalf("pause/resume starts=%d", a.Starts)
	}
	t.Logf("48000 Hz S16 stereo; left=440/right=660; starts=%d resets=%d empty=%d", a.Starts, a.Dropped, a.EmptyQueues)
}

// TestPF12HostDriver is an opt-in automation helper for the Wine X11 harness.
// It changes the Windows window through Win32, since XResizeWindow can be ignored
// by Wine's managed-window bridge. It never touches gameplay/config files.
func TestPF12HostDriver(t *testing.T) {
	action := os.Getenv("PF12_DRIVER_ACTION")
	if action == "" {
		t.Skip("external smoke helper")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	oldDPI, _, _ := up("SetThreadDpiAwarenessContext").Call(^uintptr(3))
	defer up("SetThreadDpiAwarenessContext").Call(oldDPI)
	hwnd, _, e := up("FindWindowW").Call(uintptr(unsafe.Pointer(utf("PinballFantasiesPF12"))), 0)
	if hwnd == 0 {
		t.Fatal("FindWindowW", e)
	}
	if strings.HasPrefix(action, "key:") {
		parts := strings.Split(action, ":")
		if len(parts) != 4 {
			t.Fatal(action)
		}
		keys := map[string]uintptr{"space": 32, "Return": 13, "Escape": 27, "Down": 40, "Up": 38, "Shift_L": 0x10, "Shift_R": 0x10, "Alt_L": 0x12, "Alt_R": 0x12, "F1": 112, "F2": 113, "F3": 114, "F4": 115, "F5": 116}
		vk, ok := keys[parts[1]]
		if !ok && len(parts[1]) == 1 {
			vk = uintptr(strings.ToUpper(parts[1])[0])
			ok = true
		}
		if !ok {
			t.Fatal("unknown key", parts[1])
		}
		event, _ := strconv.Atoi(parts[2])
		state, _ := strconv.Atoi(parts[3])
		scan, _, _ := up("MapVirtualKeyW").Call(vk, 0)
		if parts[1] == "Shift_R" {
			scan = 0x36
		}
		l := uintptr(1) | (scan << 16)
		if parts[1] == "Down" || parts[1] == "Up" {
			l |= 1 << 24
		}
		if parts[1] == "Alt_R" {
			l |= 1 << 24
		}
		if state&8 != 0 {
			l |= 1 << 29
		}
		if state&16 != 0 {
			l |= 1 << 30
		}
		msg := uintptr(0x100)
		if state&8 != 0 || vk == 0x12 {
			msg = 0x104
		}
		if event == 3 {
			msg++
			l |= 1<<30 | 1<<31
		}
		if v, _, e := up("PostMessageW").Call(hwnd, msg, vk, l); v == 0 {
			t.Fatal(e)
		}
		return
	}
	switch action {
	case "resize":
		if v, _, e := up("SetWindowPos").Call(hwnd, 0, 100, 100, 820, 650, 4); v == 0 {
			t.Fatal(e)
		}
	case "geometry":
		var r rect
		if v, _, e := up("GetWindowRect").Call(hwnd, uintptr(unsafe.Pointer(&r))); v == 0 {
			t.Fatal(e)
		}
		t.Logf("geometry=%d,%d,%d,%d", r.Left, r.Top, r.Right, r.Bottom)
	case "client":
		var r rect
		var origin point
		if v, _, e := up("GetClientRect").Call(hwnd, uintptr(unsafe.Pointer(&r))); v == 0 {
			t.Fatal(e)
		}
		if v, _, e := up("ClientToScreen").Call(hwnd, uintptr(unsafe.Pointer(&origin))); v == 0 {
			t.Fatal(e)
		}
		t.Logf("client=%d,%d,%d,%d", origin.X, origin.Y, r.Right, r.Bottom)
	case "focus":
		up("SetForegroundWindow").Call(hwnd)
	default:
		t.Fatalf("unknown driver action %q", action)
	}
}

func TestWin32SignedMonitorGeometry(t *testing.T) {
	for _, r := range []rect{{-1920, 0, 0, 1080}, {0, -1440, 2560, 0}, {1920, 200, 4480, 1640}} {
		args := windowRectArgs(r)
		if int32(args[0]) != r.Left || int32(args[1]) != r.Top || int32(args[2]) != r.Right-r.Left || int32(args[3]) != r.Bottom-r.Top {
			t.Fatalf("monitor rectangle lost signed position: %v %v", r, args)
		}
	}
}

func TestWin32FocusRetainsClose(t *testing.T) {
	prior := activeWindow
	defer func() { activeWindow = prior }()
	h := &hostWindow{events: []hostEvent{{kind: 2, key: 80}, {kind: 1}}}
	h.input.key(0x10, 0, true)
	activeWindow = h
	windowProc(0, 0x8, 0, 0)
	if h.Held() != 0 || len(h.events) != 2 || h.events[0].kind != 1 || h.events[1].kind != 5 {
		t.Fatal("focus must clear controls/makes while preserving close")
	}
}

func TestNativeWin32FocusPause(t *testing.T) {
	if os.Getenv("PF12_WINDOW_SMOKE") != "1" {
		t.Skip("set PF12_WINDOW_SMOKE=1 on a desktop")
	}
	data := os.Getenv("PF12_DATA_DIR")
	if data == "" {
		t.Skip("set PF12_DATA_DIR to the original data directory")
	}
	r, err := frontend.LoadConfigured(data, nil, &settings.Store{Directory: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []frontend.Key{frontend.Space, frontend.F1, frontend.F1} {
		if err := r.Update(frontend.Input{Keys: []frontend.Key{key}}); err != nil {
			t.Fatal(err)
		}
	}
	stage := 0
	err = showFrontend(r, time.Second, nil, func(h *hostWindow) {
		switch stage {
		case 0:
			windowProc(h.hwnd, 0x8, 0, 0)
		case 1:
			if r.Model.Mode != frontend.Paused {
				return // Wait until the next scheduled source update consumes it.
			}
			windowProc(h.hwnd, 0x7, 0, 0)
		case 2:
			if r.Model.Mode != frontend.Paused {
				t.Error("WM_SETFOCUS resumed frontend")
			}
			windowProc(h.hwnd, 0x100, 'P', 25<<16|1)
		case 3:
			if r.Model.Mode != frontend.Playing {
				return
			}
			windowProc(h.hwnd, 0x8, 0, 0)
		case 4:
			if r.Model.Mode != frontend.Paused {
				return
			}
			h.events = append(h.events, hostEvent{kind: 1})
		}
		stage++
	})
	if err != nil || stage < 5 {
		t.Fatal("focus pause journey incomplete", stage, err)
	}
}

// Run on real Windows with original data and an audible output device. This
// automated result verifies host/audio invariants; listening remains required.
func TestNativeFullscreenMusicTransitions(t *testing.T) {
	if os.Getenv("PF12_TRANSITION_SMOKE") != "1" {
		t.Skip("set PF12_TRANSITION_SMOKE=1 and PF12_DATA_DIR")
	}
	data := os.Getenv("PF12_DATA_DIR")
	if data == "" {
		t.Fatal("PF12_DATA_DIR must name the original data directory")
	}
	for _, scene := range []string{"selector", "table"} {
		t.Run(scene, func(t *testing.T) {
			config := &settings.Store{Directory: t.TempDir()}
			if err := config.Save(settings.Legacy()); err != nil {
				t.Fatal(err)
			}
			r, err := frontend.LoadConfigured(data, nil, config)
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range []frontend.Key{frontend.Space} {
				if err := r.Update(frontend.Input{Keys: []frontend.Key{key}}); err != nil {
					t.Fatal(err)
				}
			}
			if scene == "table" {
				for _, key := range []frontend.Key{frontend.F1, frontend.F1} {
					if err := r.Update(frontend.Input{Keys: []frontend.Key{key}}); err != nil {
						t.Fatal(err)
					}
				}
			}
			mode := r.Model.Mode
			a, err := OpenAudio()
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			path := filepath.Join(t.TempDir(), "transitions.jsonl")
			if dir := os.Getenv("PF12_TRANSITION_OUTPUT_DIR"); dir != "" {
				path = filepath.Join(dir, scene+"-transitions.jsonl")
			}
			os.Remove(path)
			t.Setenv("PF12_TRANSITION_LOG", path)
			started := time.Now()
			count := 0
			var hwnd uintptr
			var geometry WindowGeometry
			err = showFrontend(r, 10*time.Second, a, func(h *hostWindow) {
				if hwnd == 0 {
					hwnd = h.hwnd
					// Keep the window away from desktop-manager edge snapping in Wine.
					if v, _, e := up("SetWindowPos").Call(hwnd, 0, 100, 100, 0, 0, 0x15); v == 0 {
						t.Fatal(e)
					}
				}
				if h.hwnd != hwnd {
					t.Error("HWND changed")
				}
				if count < 4 && time.Since(started) >= time.Duration(2+count*2)*time.Second {
					if count == 0 {
						geometry = h.Geometry()
					}
					// Alt+Enter make, repeat, and break use the real Win32 key translator
					// and the normal hostKeys/overdue PCM path. Repeat must be consumed.
					windowProc(hwnd, 0x104, 13, 1<<29|0x1c<<16|1)
					windowProc(hwnd, 0x104, 13, 1<<29|1<<30|0x1c<<16|1)
					windowProc(hwnd, 0x105, 13, 1<<29|1<<30|1<<31|0x1c<<16|1)
					count++
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			transitions := 0
			for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
				var record fullscreenTransition
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					t.Fatal(err)
				}
				if record.Kind != "transition" {
					continue
				}
				transitions++
				t.Log(string(line))
				if record.HWND != hwnd || record.SurfaceRecreations != 0 || record.Error != "" {
					t.Fatal("host transaction invariant", record)
				}
				if record.Pre.Starts != 1 || record.Post.Starts != 1 || record.Post.Clears != 0 || record.Post.WaveOutResets != 0 {
					t.Fatal("audio restarted/reset", record)
				}
				if record.Pre.EmptyObserved || record.Post.EmptyObserved || record.Post.Empty != 0 {
					t.Error("queue drained; real Windows listening required", record)
				}
				if !record.Fullscreen && record.GeometryPost != geometry {
					t.Fatal("window geometry not restored", record.GeometryPost, geometry)
				}
			}
			if transitions != 4 || count != 4 || (scene == "table" && r.Model.Mode != mode) {
				t.Fatal("shortcut count/Enter consumption", transitions, count, r.Model.Mode, mode)
			}
			if a.EmptyQueues != 0 || a.Dropped != 0 || a.Starts != 1 || a.LifecycleClears != 0 {
				t.Fatal("continuous music lifecycle", a.EmptyQueues, a.Dropped, a.Starts, a.LifecycleClears)
			}
		})
	}
}

// Wine reconstructs APPDATA for a Unix-launched process. A native child uses
// the supplied Windows environment, matching CreateProcess on real Windows.
func TestPF12PortableArtifactLaunch(t *testing.T) {
	encoded := os.Getenv("PF12_PORTABLE_COMMAND")
	if encoded == "" {
		t.Skip("portable artifact launch helper")
	}
	var argv []string
	if err := json.Unmarshal([]byte(encoded), &argv); err != nil || len(argv) == 0 {
		t.Fatal("invalid portable command", err)
	}
	if dir := os.Getenv("PF12_PORTABLE_APPDATA"); dir != "" {
		t.Setenv("APPDATA", dir)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, argv[0], argv[1:]...)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}
}
