package platform

// WindowGeometry is expressed in host window coordinates, independent of the
// logical framebuffer and drawable pixels. Backends restore maximized state too.
type WindowGeometry struct {
	X, Y, Width, Height int
	Maximized           bool
}

type windowBackend interface {
	Geometry() WindowGeometry
	SetFullscreen(bool) error
	RestoreGeometry(WindowGeometry)
}

// WindowPresentation owns session-only host state. The backend may use SDL or
// native window APIs; no gameplay, configuration, texture or audio state enters
// this operation. Startup is always windowed.
type WindowPresentation struct {
	backend    windowBackend
	fullscreen bool
	windowed   WindowGeometry
}

func (w *WindowPresentation) IsFullscreen() bool { return w.fullscreen }
func (w *WindowPresentation) ToggleFullscreen() error {
	if !w.fullscreen {
		geometry := w.backend.Geometry()
		if err := w.backend.SetFullscreen(true); err != nil {
			return err
		}
		w.windowed = geometry
		w.fullscreen = true
		return nil
	}
	if err := w.backend.SetFullscreen(false); err != nil {
		return err
	}
	w.fullscreen = false
	w.backend.RestoreGeometry(w.windowed)
	return nil
}

type keyAction uint8

const (
	ignoreKey keyAction = iota
	dispatchKey
	toggleFullscreen
)

// Return must have a new make edge. Consume every repeated Return in the
// shortcut, including backends that send duplicate downs without repeat flags.
type hostKeys struct{ enterDown bool }

func (k *hostKeys) down(code int, alt, repeat bool) keyAction {
	if code == 28 {
		held := k.enterDown
		k.enterDown = true
		if held || repeat {
			return ignoreKey
		}
		if alt {
			return toggleFullscreen
		}
	}
	if repeat {
		return ignoreKey
	}
	return dispatchKey
}
func (k *hostKeys) enterUp() { k.enterDown = false }
