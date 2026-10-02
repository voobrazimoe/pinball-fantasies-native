package platform

import (
	"errors"
	"testing"
)

type fakeWindow struct {
	geometry   WindowGeometry
	fullscreen bool
	changes    int
	fail       bool
}

func (w *fakeWindow) Geometry() WindowGeometry { return w.geometry }
func (w *fakeWindow) SetFullscreen(on bool) error {
	if w.fail {
		return errors.New("host failure")
	}
	w.fullscreen = on
	w.changes++
	w.geometry = WindowGeometry{Width: 1920, Height: 1080}
	return nil
}
func (w *fakeWindow) RestoreGeometry(g WindowGeometry) { w.geometry = g }
func TestFullscreenSessionGeometry(t *testing.T) {
	for _, maximized := range []bool{false, true} {
		geometry := WindowGeometry{37, 59, 1000, 800, maximized}
		backend := &fakeWindow{geometry: geometry}
		host := WindowPresentation{backend: backend}
		if host.IsFullscreen() {
			t.Fatal("startup fullscreen")
		}
		for n := 0; n < 3; n++ {
			if e := host.ToggleFullscreen(); e != nil || !host.IsFullscreen() || !backend.fullscreen {
				t.Fatal(e)
			}
			if e := host.ToggleFullscreen(); e != nil || host.IsFullscreen() || backend.fullscreen || backend.geometry != geometry {
				t.Fatal("geometry", backend.geometry, e)
			}
			geometry.Width += 10
			backend.geometry = geometry // later user resize becomes the new restoration target
		}
		backend.fail = true
		if e := host.ToggleFullscreen(); e == nil || host.IsFullscreen() {
			t.Fatal("failed entry changed state")
		}
		backend.fail = false
		host.ToggleFullscreen()
		backend.fail = true
		if e := host.ToggleFullscreen(); e == nil || !host.IsFullscreen() {
			t.Fatal("failed exit changed state")
		}
	}
}
func TestFullscreenMakeEdgeConsumed(t *testing.T) {
	for _, altName := range []string{"left Alt", "right Alt"} {
		t.Run(altName, func(t *testing.T) {
			keys := hostKeys{}
			if keys.down(127, false, false) != dispatchKey {
				t.Fatal("standalone Alt lost")
			}
			if keys.down(28, true, false) != toggleFullscreen {
				t.Fatal("shortcut not recognized")
			}
			for _, repeat := range []bool{true, false, true} {
				if keys.down(28, true, repeat) != ignoreKey {
					t.Fatal("duplicate shortcut")
				}
			}
			keys.enterUp()
			if keys.down(28, true, false) != toggleFullscreen {
				t.Fatal("new edge")
			}
			keys.enterUp()
			if keys.down(28, false, false) != dispatchKey {
				t.Fatal("ordinary Enter lost")
			}
			if keys.down(28, true, false) != ignoreKey {
				t.Fatal("modifier change is not make edge")
			}
			keys.enterUp()
			if keys.down(28, true, true) != ignoreKey {
				t.Fatal("repeat cannot toggle")
			}
		})
	}
}
