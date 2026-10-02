//go:build windows

package platform

import (
	"image"
	"runtime"
	"time"
)

func Show(f *image.RGBA, d time.Duration) error        { return ShowInitial(f, d) }
func ShowInitial(f *image.RGBA, d time.Duration) error { return ShowTablePhysics(f, d, "", nil) }
func ShowPhysics(f *image.RGBA, d time.Duration, next func(PhysicsControls) (*image.RGBA, error)) error {
	return ShowTablePhysics(f, d, "Party Land", next)
}
func ShowTablePhysics(f *image.RGBA, d time.Duration, table string, next func(PhysicsControls) (*image.RGBA, error)) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	h, e := openHost(f)
	if e != nil {
		return e
	}
	defer h.Close()
	deadline, sync := time.Now().Add(d), time.Now()
	paused := false
	keys := hostKeys{}
	for {
		h.Pump()
		if e := h.Present(f); e != nil {
			return e
		}
		if d > 0 && time.Now().After(deadline) {
			return nil
		}
		c := PhysicsControls{}
		focusLost := false
		pendingFullscreen := 0
		for {
			e := h.Event()
			if e.kind == 0 {
				break
			}
			switch e.kind {
			case 1:
				return nil
			case 3:
				c.Release = true
			case 4:
				keys.enterUp()
			case 5:
				keys.enterUp()
				focusLost = true
			case 2:
				action := keys.down(e.key, e.alt, e.repeat)
				if action == toggleFullscreen {
					pendingFullscreen++
					continue
				}
				if action != dispatchKey {
					continue
				}
				if e.key == 1 {
					return nil
				}
				if e.key == 25 {
					paused = !paused
				} else if paused {
					paused = false
				}
				if e.key == 50 {
					c.MusicToggle = true
				}
			}
		}
		if focusLost {
			paused = true
		}
		bits := h.Held()
		c.Left = bits&1 != 0
		c.Right = bits&2 != 0
		c.Down = bits&4 != 0
		c.Tilt = bits&8 != 0
		if paused {
			sync = time.Now()
		} else if next != nil {
			f, e = next(c)
			if e != nil {
				return e
			}
		}
		for pendingFullscreen > 0 {
			pendingFullscreen--
			if err := h.ToggleFullscreen(nil); err != nil {
				return err
			}
		}
		sync = sync.Add(time.Second / 71)
		if delay := time.Until(sync); delay > 0 {
			time.Sleep(delay)
		}
	}
}
