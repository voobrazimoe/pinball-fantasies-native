//go:build linux

package platform

/*
#cgo CFLAGS: -I${SRCDIR}/../../.tools/sdl2/usr/include/SDL2 -I${SRCDIR}/../../.tools/sdl2/usr/include/x86_64-linux-gnu
#cgo LDFLAGS: -l:libSDL2-2.0.so.0
#include <SDL.h>
static int pf3_controls(void) {
 SDL_Event e; int result=0;
 while(SDL_PollEvent(&e)) {
  if(e.type==SDL_QUIT || (e.type==SDL_KEYDOWN && e.key.keysym.sym==SDLK_ESCAPE)) result|=1;
  if(e.type==SDL_WINDOWEVENT && e.window.event==SDL_WINDOWEVENT_FOCUS_LOST) result|=512;
  if(e.type==SDL_KEYDOWN && !e.key.repeat) {result|=256; if(e.key.keysym.sym==SDLK_p)result|=64; if(e.key.keysym.sym==SDLK_m)result|=128;}
  if(e.type==SDL_KEYUP && e.key.keysym.sym==SDLK_DOWN) result|=8;
 }
 const Uint8 *keys=SDL_GetKeyboardState(NULL);
 if(keys[SDL_SCANCODE_LSHIFT]||keys[SDL_SCANCODE_LCTRL]||keys[SDL_SCANCODE_LALT]) result|=2;
 if(keys[SDL_SCANCODE_RSHIFT]||keys[SDL_SCANCODE_RCTRL]||keys[SDL_SCANCODE_RALT]) result|=4;
 if(keys[SDL_SCANCODE_DOWN]) result|=16;
 if(keys[SDL_SCANCODE_SPACE]) result|=32;
 return result;
}
*/
import "C"
import (
	"fmt"
	"image"
	"runtime"
	"time"
	"unsafe"
)

// ShowPhysics keeps one deterministic original sync per presentation step.
// Fixed 71 Hz deadlines include frame work. Late frames catch up using whole
// original Sync calls; host elapsed time is never passed into physics.
func ShowPhysics(first *image.RGBA, duration time.Duration, next func(PhysicsControls) (*image.RGBA, error)) error {
	return ShowTablePhysics(first, duration, "Party Land", next)
}
func ShowTablePhysics(first *image.RGBA, duration time.Duration, table string, next func(PhysicsControls) (*image.RGBA, error)) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fail := func() error { return fmt.Errorf("SDL2: %s", C.GoString(C.SDL_GetError())) }
	if C.SDL_Init(C.SDL_INIT_VIDEO) != 0 {
		return fail()
	}
	defer C.SDL_QuitSubSystem(C.SDL_INIT_VIDEO)
	title := C.CString("Pinball Fantasies — " + table + " (Down spring; Space push; Shift/Ctrl/Alt flippers)")
	defer C.SDL_free(unsafe.Pointer(title))
	w, h := C.int(first.Rect.Dx()), C.int(first.Rect.Dy())
	initial := hostWindowSize()
	window := C.SDL_CreateWindow(title, C.SDL_WINDOWPOS_CENTERED, C.SDL_WINDOWPOS_CENTERED, C.int(initial.X), C.int(initial.Y), C.SDL_WINDOW_SHOWN|C.SDL_WINDOW_RESIZABLE)
	if window == nil {
		return fail()
	}
	defer C.SDL_DestroyWindow(window)
	renderer := C.SDL_CreateRenderer(window, -1, C.SDL_RENDERER_SOFTWARE)
	if renderer == nil {
		return fail()
	}
	defer C.SDL_DestroyRenderer(renderer)
	if C.SDL_RenderSetLogicalSize(renderer, w, h) != 0 {
		return fail()
	}
	texture := C.SDL_CreateTexture(renderer, C.SDL_PIXELFORMAT_RGBA32, C.SDL_TEXTUREACCESS_STREAMING, w, h)
	if texture == nil {
		return fail()
	}
	defer C.SDL_DestroyTexture(texture)
	fmt.Println(table+" window opened; SDL driver:", C.GoString(C.SDL_GetCurrentVideoDriver()))
	frame := first
	paused := false
	deadline := time.Now().Add(duration)
	nextSync := time.Now()
	for {
		if C.SDL_UpdateTexture(texture, nil, unsafe.Pointer(&frame.Pix[0]), C.int(frame.Stride)) != 0 {
			return fail()
		}
		runtime.KeepAlive(frame)
		if C.SDL_RenderClear(renderer) != 0 || C.SDL_RenderCopy(renderer, texture, nil, nil) != 0 {
			return fail()
		}
		C.SDL_RenderPresent(renderer)
		if duration > 0 && time.Now().After(deadline) {
			return nil
		}
		bits := int(C.pf3_controls())
		if bits&1 != 0 {
			return nil
		}
		if bits&512 != 0 {
			paused = true
		} else if bits&64 != 0 {
			paused = !paused
		} else if paused && bits&256 != 0 {
			paused = false
		}
		if paused {
			nextSync = time.Now()
			time.Sleep(time.Second / 71)
			continue
		}
		var err error
		frame, err = next(PhysicsControls{Left: bits&2 != 0, Right: bits&4 != 0, Release: bits&8 != 0, Down: bits&16 != 0, Tilt: bits&32 != 0, MusicToggle: bits&128 != 0})
		if err != nil {
			return err
		}
		nextSync = nextSync.Add(time.Second / 71)
		if remaining := time.Until(nextSync); remaining > 0 {
			time.Sleep(remaining)
		}
	}
}
