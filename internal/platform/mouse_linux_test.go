//go:build linux

package platform

import (
	"image"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// Exercise SDL's real relative event path, including a resized window. XTest
// generates host input; no fixture hooks are compiled into the game binary.
func TestNativeSDLRelativeMouseSmoke(t *testing.T) {
	if os.Getenv("PF12_WINDOW_SMOKE") != "1" {
		t.Skip("set PF12_WINDOW_SMOKE=1 on an isolated X11 display")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	h, err := openHost(image.NewRGBA(image.Rect(0, 0, 320, 240)))
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	inject := func(action string) {
		t.Helper()
		script := `import ctypes as C,sys
x=C.CDLL('libX11.so.6');xt=C.CDLL('libXtst.so.6')
x.XOpenDisplay.restype=C.c_void_p
x.XDefaultRootWindow.argtypes=[C.c_void_p];x.XDefaultRootWindow.restype=C.c_ulong
x.XQueryTree.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_ulong),C.POINTER(C.c_ulong),C.POINTER(C.POINTER(C.c_ulong)),C.POINTER(C.c_uint)]
x.XFetchName.argtypes=[C.c_void_p,C.c_ulong,C.POINTER(C.c_char_p)]
x.XFree.argtypes=[C.c_void_p]
x.XSetInputFocus.argtypes=[C.c_void_p,C.c_ulong,C.c_int,C.c_ulong]
x.XSync.argtypes=[C.c_void_p,C.c_int]
xt.XTestFakeRelativeMotionEvent.argtypes=[C.c_void_p,C.c_int,C.c_int,C.c_ulong]
xt.XTestFakeButtonEvent.argtypes=[C.c_void_p,C.c_uint,C.c_int,C.c_ulong]
d=x.XOpenDisplay(None);assert d
if sys.argv[1]=='focus':
 root=x.XDefaultRootWindow(d);rt,pa=C.c_ulong(),C.c_ulong();children=C.POINTER(C.c_ulong)();n=C.c_uint()
 assert x.XQueryTree(d,root,C.byref(rt),C.byref(pa),C.byref(children),C.byref(n))
 found=False
 for i in range(n.value):
  name=C.c_char_p();x.XFetchName(d,children[i],C.byref(name))
  if name.value==b'Pinball Fantasies':x.XSetInputFocus(d,children[i],1,0);found=True
  if name:x.XFree(name)
 assert found
else:
 assert xt.XTestFakeRelativeMotionEvent(d,0,16,0)
 assert xt.XTestFakeButtonEvent(d,1,1,0)
 assert xt.XTestFakeButtonEvent(d,1,0,0)
x.XSync(d,0)
`
		if output, err := exec.Command("python3", "-c", script, action).CombinedOutput(); err != nil {
			t.Fatalf("XTest: %v %s", err, output)
		}
	}
	drain := func() (int, int) {
		y, fires := 0, 0
		until := time.Now().Add(100 * time.Millisecond)
		for time.Now().Before(until) {
			for e := h.Event(); e.kind != 0; e = h.Event() {
				if e.kind == 6 {
					y += e.mouseY
				}
				if e.kind == 7 {
					fires++
				}
			}
			time.Sleep(time.Millisecond)
		}
		return y, fires
	}
	inject("focus")
	drain()
	h.MouseActive(true)
	drain()
	baseline := 0
	for _, size := range []image.Point{image.Pt(640, 480), image.Pt(800, 600)} {
		h.presentation.backend.RestoreGeometry(WindowGeometry{X: 100, Y: 100, Width: size.X, Height: size.Y})
		drain()
		inject("motion")
		y, fires := drain()
		if baseline == 0 {
			baseline = y
		}
		if y <= 0 || y != baseline || fires != 1 {
			t.Fatalf("size %v: relative Y=%d fire edges=%d", size, y, fires)
		}
	}
}
