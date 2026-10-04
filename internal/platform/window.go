//go:build linux

// Package platform is the small host-only SDL2 boundary.
package platform

/*
#cgo CFLAGS: -I${SRCDIR}/../../.tools/sdl2/usr/include/SDL2 -I${SRCDIR}/../../.tools/sdl2/usr/include/x86_64-linux-gnu
#cgo LDFLAGS: -l:libSDL2-2.0.so.0
#include <SDL.h>
static void pf_nearest(void) { SDL_SetHint(SDL_HINT_RENDER_SCALE_QUALITY,"0"); }
static int pf_poll_quit(void) {
 SDL_Event e;
 while (SDL_PollEvent(&e)) {
  if (e.type == SDL_QUIT || (e.type == SDL_KEYDOWN && e.key.keysym.sym == SDLK_ESCAPE)) return 1;
 }
 return 0;
}
*/
import "C"
import (
	"fmt"
	"image"
	"pinballfantasies/internal/diagnostics"
	"runtime"
	"time"
	"unsafe"
)

// hostWindowSize runs only before creation. SDL display bounds and returned
// dimensions are window coordinates, never renderer drawable pixels.
func hostWindowSize() image.Point {
	bounds := C.SDL_Rect{}
	usable := image.Pt(800, 600)
	if C.SDL_GetDisplayUsableBounds(0, &bounds) == 0 {
		usable = image.Pt(int(bounds.w), int(bounds.h))
	}
	return initialWindowSize(usable)
}

// Show uploads one static native RGBA framebuffer; duration=0 waits for close.
func Show(frame *image.RGBA, duration time.Duration) error {
	return show(frame, duration, "Pinball Fantasies — Party Land — PF1")
}

func ShowInitial(frame *image.RGBA, duration time.Duration) error {
	return show(frame, duration, "Pinball Fantasies — Party Land — PF2")
}

func show(frame *image.RGBA, duration time.Duration, windowTitle string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fail := func() error { return fmt.Errorf("SDL2: %s", C.GoString(C.SDL_GetError())) }
	if C.SDL_Init(C.SDL_INIT_VIDEO) != 0 {
		return fail()
	}
	defer C.SDL_Quit()
	C.pf_nearest()
	title := C.CString(windowTitle)
	defer C.SDL_free(unsafe.Pointer(title))
	w, h := C.int(frame.Rect.Dx()), C.int(frame.Rect.Dy())
	initial := hostWindowSize()
	window := C.SDL_CreateWindow(title, C.SDL_WINDOWPOS_CENTERED, C.SDL_WINDOWPOS_CENTERED, C.int(initial.X), C.int(initial.Y), C.SDL_WINDOW_SHOWN|C.SDL_WINDOW_RESIZABLE)
	if window == nil {
		return fail()
	}
	defer C.SDL_DestroyWindow(window)
	// Software renderer avoids relying on an OpenGL driver for this static slice.
	renderer := C.SDL_CreateRenderer(window, -1, C.SDL_RENDERER_SOFTWARE)
	if renderer == nil {
		return fail()
	}
	defer C.SDL_DestroyRenderer(renderer)
	if C.SDL_RenderSetLogicalSize(renderer, w, h) != 0 {
		return fail()
	}
	texture := C.SDL_CreateTexture(renderer, C.SDL_PIXELFORMAT_RGBA32, C.SDL_TEXTUREACCESS_STATIC, w, h)
	if texture == nil {
		return fail()
	}
	defer C.SDL_DestroyTexture(texture)
	if C.SDL_UpdateTexture(texture, nil, unsafe.Pointer(&frame.Pix[0]), C.int(frame.Stride)) != 0 {
		return fail()
	}
	runtime.KeepAlive(frame)
	deadline := time.Now().Add(duration)
	diagnostics.Println(windowTitle, "window opened; SDL driver:", C.GoString(C.SDL_GetCurrentVideoDriver()))
	for {
		if C.pf_poll_quit() != 0 {
			return nil
		}
		if C.SDL_RenderClear(renderer) != 0 || C.SDL_RenderCopy(renderer, texture, nil, nil) != 0 {
			return fail()
		}
		C.SDL_RenderPresent(renderer)
		if duration > 0 && time.Now().After(deadline) {
			return nil
		}
		C.SDL_Delay(50)
	}
}
