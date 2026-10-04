//go:build linux

package platform

/*
#cgo CFLAGS: -I${SRCDIR}/../../.tools/sdl2/usr/include/SDL2 -I${SRCDIR}/../../.tools/sdl2/usr/include/x86_64-linux-gnu
#cgo LDFLAGS: -l:libSDL2-2.0.so.0
#include <SDL.h>
static int pf6_key(SDL_Scancode k) {
 switch(k) {
 case SDL_SCANCODE_ESCAPE:return 1;case SDL_SCANCODE_RETURN:return 28;case SDL_SCANCODE_SPACE:return 57;
 case SDL_SCANCODE_F1:return 59;case SDL_SCANCODE_F2:return 60;case SDL_SCANCODE_F3:return 61;case SDL_SCANCODE_F4:return 62;
 case SDL_SCANCODE_F5:return 63;case SDL_SCANCODE_F6:return 64;case SDL_SCANCODE_F7:return 65;case SDL_SCANCODE_F8:return 66;case SDL_SCANCODE_UP:return 72;case SDL_SCANCODE_DOWN:return 80;
 case SDL_SCANCODE_KP_MULTIPLY:return 55;
 }
 const char *letters="qwertyuiopasdfghjklzxcvbnm";
 const int scans[]={16,17,18,19,20,21,22,23,24,25,30,31,32,33,34,35,36,37,38,44,45,46,47,48,49,50};
 for(int i=0;i<26;i++) if(k==SDL_SCANCODE_A+letters[i]-'a') return scans[i];
 // Other make codes must resume pause and answer quit prompts too.
 return 127;
}
static int pf6_event(int *key, int *alt, int *repeat, int *mouseY) {
 SDL_Event e;
 while(SDL_PollEvent(&e)) {
  if(e.type==SDL_MOUSEMOTION) { *mouseY=e.motion.yrel; return 6; }
  if(e.type==SDL_MOUSEBUTTONDOWN && e.button.button==SDL_BUTTON_LEFT) return 7;
  if(e.type==SDL_WINDOWEVENT && e.window.event==SDL_WINDOWEVENT_FOCUS_GAINED) return 8;
  if(e.type==SDL_QUIT) return 1;
  if(e.type==SDL_WINDOWEVENT && e.window.event==SDL_WINDOWEVENT_FOCUS_LOST) return 5;
  if(e.type==SDL_KEYDOWN) {*key=pf6_key(e.key.keysym.scancode);*alt=(e.key.keysym.mod & KMOD_ALT)!=0;*repeat=e.key.repeat;return 2;}
  if(e.type==SDL_KEYUP && e.key.keysym.scancode==SDL_SCANCODE_RETURN) return 4;
  if(e.type==SDL_KEYUP && e.key.keysym.scancode==SDL_SCANCODE_DOWN) return 3;
 }
 return 0;
}
static int pf6_held(void) {const Uint8 *k=SDL_GetKeyboardState(NULL);return ((k[SDL_SCANCODE_LSHIFT]||k[SDL_SCANCODE_LCTRL]||k[SDL_SCANCODE_LALT])?1:0)|((k[SDL_SCANCODE_RSHIFT]||k[SDL_SCANCODE_RCTRL]||k[SDL_SCANCODE_RALT])?2:0)|(k[SDL_SCANCODE_DOWN]?4:0)|(k[SDL_SCANCODE_SPACE]?8:0);}

*/
import "C"
import (
	"fmt"
	"image"
	"pinballfantasies/internal/diagnostics"
	"pinballfantasies/internal/gameplay"
	"runtime"
	"unsafe"
)

type hostWindow struct {
	window       *C.SDL_Window
	renderer     *C.SDL_Renderer
	texture      *C.SDL_Texture
	size         image.Point
	presentation WindowPresentation
	mouseActive  bool
}

func openHost(first *image.RGBA) (_ *hostWindow, err error) {
	fail := func() error { return fmt.Errorf("SDL2: %s", C.GoString(C.SDL_GetError())) }
	if C.SDL_Init(C.SDL_INIT_VIDEO) != 0 {
		return nil, fail()
	}
	defer func() {
		if err != nil {
			C.SDL_QuitSubSystem(C.SDL_INIT_VIDEO)
		}
	}()
	hint := C.CString("SDL_MOUSE_RELATIVE_SCALING")
	value := C.CString("0")
	C.SDL_SetHint(hint, value)
	C.SDL_free(unsafe.Pointer(hint))
	C.SDL_free(unsafe.Pointer(value))
	title := C.CString("Pinball Fantasies")
	defer C.SDL_free(unsafe.Pointer(title))
	w, h := C.int(first.Rect.Dx()), C.int(first.Rect.Dy())
	initial := hostWindowSize()
	window := C.SDL_CreateWindow(title, C.SDL_WINDOWPOS_CENTERED, C.SDL_WINDOWPOS_CENTERED, C.int(initial.X), C.int(initial.Y), C.SDL_WINDOW_SHOWN|C.SDL_WINDOW_RESIZABLE)
	if window == nil {
		return nil, fail()
	}
	defer func() {
		if err != nil {
			C.SDL_DestroyWindow(window)
		}
	}()
	renderer := C.SDL_CreateRenderer(window, -1, C.SDL_RENDERER_SOFTWARE)
	if renderer == nil {
		return nil, fail()
	}
	defer func() {
		if err != nil {
			C.SDL_DestroyRenderer(renderer)
		}
	}()
	if C.SDL_RenderSetLogicalSize(renderer, w, h) != 0 {
		return nil, fail()
	}

	hst := &hostWindow{window: window, renderer: renderer}
	hst.presentation.backend = sdlWindow{window}
	diagnostics.Println("PF6 window opened; SDL driver:", C.GoString(C.SDL_GetCurrentVideoDriver()))
	return hst, nil
}
func (h *hostWindow) Close() {
	if h.texture != nil {
		C.SDL_DestroyTexture(h.texture)
	}
	C.SDL_DestroyRenderer(h.renderer)
	C.SDL_DestroyWindow(h.window)
	C.SDL_QuitSubSystem(C.SDL_INIT_VIDEO)
}
func (h *hostWindow) Present(frame *image.RGBA) error {
	fail := func() error { return fmt.Errorf("SDL2: %s", C.GoString(C.SDL_GetError())) }
	if h.size != frame.Rect.Size() {
		if h.texture != nil {
			C.SDL_DestroyTexture(h.texture)
			h.texture = nil
		}
		h.size = frame.Rect.Size()
		diagnostics.Printf("PF11.2 presentation %dx%d window ID=%d fullscreen=%t\n", h.size.X, h.size.Y, uint32(C.SDL_GetWindowID(h.window)), h.presentation.IsFullscreen())
		logical := logicalSize(h.size)
		if C.SDL_RenderSetLogicalSize(h.renderer, C.int(logical.X), C.int(logical.Y)) != 0 {
			return fail()
		}
		h.texture = C.SDL_CreateTexture(h.renderer, C.SDL_PIXELFORMAT_RGBA32, C.SDL_TEXTUREACCESS_STREAMING, C.int(h.size.X), C.int(h.size.Y))
		if h.texture == nil {
			return fail()
		}
	}
	if C.SDL_UpdateTexture(h.texture, nil, unsafe.Pointer(&frame.Pix[0]), C.int(frame.Stride)) != 0 {
		return fail()
	}
	runtime.KeepAlive(frame)
	if C.SDL_RenderClear(h.renderer) != 0 || C.SDL_RenderCopy(h.renderer, h.texture, nil, nil) != 0 {
		return fail()
	}
	C.SDL_RenderPresent(h.renderer)
	return nil
}
func (h *hostWindow) Event() hostEvent {
	for {
		var key, alt, repeat, mouseY C.int
		kind := int(C.pf6_event(&key, &alt, &repeat, &mouseY))
		if (kind == 6 || kind == 7) && !h.mouseActive {
			continue
		}
		return hostEvent{kind: kind, key: int(key), mouseY: int(mouseY), alt: alt != 0, repeat: repeat != 0}
	}
}
func (h *hostWindow) Held() int {
	if C.SDL_GetWindowFlags(h.window)&C.SDL_WINDOW_INPUT_FOCUS == 0 {
		return 0
	}
	return int(C.pf6_held())
}

func (h *hostWindow) FullscreenTrace(clears, starts uint64, device *AudioDevice) {
	diagnostics.Printf("PF11.2 fullscreen=%t logical=%dx%d window ID=%d\n", h.presentation.IsFullscreen(), h.size.X, h.size.Y, uint32(C.SDL_GetWindowID(h.window)))
	if device != nil {
		diagnostics.Printf("PF11.2 fullscreen audio lifecycle clears=%d/%d starts=%d/%d producer unchanged=true\n", clears, device.LifecycleClears, starts, device.Starts)
	}
}

func (h *hostWindow) Pump()                     {}
func (h *hostWindow) SetModalTick(func() error) {}

func (h *hostWindow) ToggleFullscreen(device *AudioDevice) error {
	var clears, starts uint64
	if device != nil {
		clears, starts = device.LifecycleClears, device.Starts
	}
	if err := h.presentation.ToggleFullscreen(); err != nil {
		return err
	}
	h.FullscreenTrace(clears, starts, device)
	return nil
}

func (h *hostWindow) SetSourceTick(func() error) {}

// SDL relative mode removes pointer bounds and logical-renderer scaling.
func (h *hostWindow) MouseActive(active bool) {
	enabled := C.SDL_bool(C.SDL_FALSE)
	if active && C.SDL_GetWindowFlags(h.window)&C.SDL_WINDOW_INPUT_FOCUS != 0 {
		enabled = C.SDL_TRUE
	}
	h.mouseActive = enabled == C.SDL_TRUE
	if C.SDL_GetRelativeMouseMode() != enabled {
		C.SDL_SetRelativeMouseMode(enabled)
	}
}

// Controls translates host held state into logical source controls.
func (h *hostWindow) Controls() gameplay.Controls {
	bits := h.Held()
	return gameplay.Controls{Left: bits&1 != 0, Right: bits&2 != 0, Down: bits&4 != 0, Tilt: bits&8 != 0}
}
