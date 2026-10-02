//go:build linux

package platform

/*
#include <SDL.h>
*/
import "C"
import "fmt"

// SDL uses the current desktop mode; never request a physical mode switch.
type sdlWindow struct{ window *C.SDL_Window }

func (w sdlWindow) Geometry() WindowGeometry {
	var x, y, width, height C.int
	C.SDL_GetWindowPosition(w.window, &x, &y)
	C.SDL_GetWindowSize(w.window, &width, &height)
	// SDL's X11 position getter reports the client origin, while the window
	// manager interprets move requests as the decorated frame origin. Preserve
	// that frame origin so restoring a decorated window does not drift downward.
	var top, left, bottom, right C.int
	if C.SDL_GetWindowBordersSize(w.window, &top, &left, &bottom, &right) == 0 {
		x -= left
		y -= top
	}
	return WindowGeometry{int(x), int(y), int(width), int(height), C.SDL_GetWindowFlags(w.window)&C.SDL_WINDOW_MAXIMIZED != 0}
}
func (w sdlWindow) SetFullscreen(on bool) error {
	var flags C.Uint32
	if on {
		flags = C.SDL_WINDOW_FULLSCREEN_DESKTOP
	}
	if C.SDL_SetWindowFullscreen(w.window, flags) != 0 {
		return fmt.Errorf("SDL2 fullscreen: %s", C.GoString(C.SDL_GetError()))
	}
	return nil
}
func (w sdlWindow) RestoreGeometry(g WindowGeometry) {
	if g.Maximized {
		// Let SDL/window-manager ownership retain the pre-maximize resize state.
		C.SDL_MaximizeWindow(w.window)
		return
	}
	C.SDL_RestoreWindow(w.window)
	C.SDL_SetWindowSize(w.window, C.int(g.Width), C.int(g.Height))
	C.SDL_SetWindowPosition(w.window, C.int(g.X), C.int(g.Y))
}
