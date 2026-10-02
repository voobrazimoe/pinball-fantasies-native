package platform

import (
	"fmt"
	"os"
	"pinballfantasies/internal/frontend"
	"runtime"
	"time"
)

type PhysicsControls struct{ Left, Right, Down, Release, Tilt, MusicToggle bool }

type hostEvent struct {
	kind, key   int
	alt, repeat bool
}

// ShowFrontend drives deterministic 60 Hz INTRO and unchanged 71 Hz table syncs.
func ShowFrontend(r *frontend.Runtime, duration time.Duration, device *AudioDevice) error {
	return showFrontend(r, duration, device, nil)
}

// pumpHook is used only by the bounded native transition smoke.
func showFrontend(r *frontend.Runtime, duration time.Duration, device *AudioDevice, pumpHook func(*hostWindow)) error {
	first := r.Frame()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	host, err := openHost(first)
	if err != nil {
		return err
	}
	defer host.Close()
	audit, err := openPacingLog()
	if err != nil {
		return err
	}
	defer audit.Close()
	noPresent := os.Getenv("PF12_NO_PRESENT") == "1"
	if os.Getenv("PF12_AUDIT_FULLSCREEN") == "1" {
		if e := host.presentation.ToggleFullscreen(); e != nil {
			return e
		}
	}
	present := func() error {
		start := time.Now()
		frame := r.Frame()
		audit.mark("frame", start, host.presentation.IsFullscreen(), device)
		if noPresent {
			return nil
		}
		start = time.Now()
		err := host.Present(frame)
		audit.mark("present", start, host.presentation.IsFullscreen(), device)
		return err
	}
	keys := hostKeys{}
	suspended := false

	deadline := time.Now().Add(duration)
	nextSync := time.Now()

	done := false
	pendingFullscreen := 0
	step := func() error {
		start := time.Now()
		defer func() { audit.mark("update", start, host.presentation.IsFullscreen(), device) }()
		input := frontend.Input{}
		for {
			e := host.Event()
			event := e.kind
			if event == 0 {
				break
			}
			switch event {
			case 1:
				input.Close = true
			case 2:
				switch keys.down(e.key, e.alt, e.repeat) {
				case dispatchKey:
					input.Keys = append(input.Keys, frontend.Key(e.key))
				case toggleFullscreen:
					pendingFullscreen++
				}
			case 3:
				input.Release = true
			case 4:
				keys.enterUp()
			case 5:
				keys.enterUp()
				input.FocusLost = true
			}
		}
		bits := host.Held()
		input.Left = bits&1 != 0
		input.Right = bits&2 != 0
		input.Down = bits&4 != 0
		input.Tilt = bits&8 != 0
		hz := r.Model.Hz()
		priorSource := r.AudioSource()
		priorMode := r.Model.Mode
		priorTable := r.Model.Selected
		if err := r.Update(input); err != nil {
			return err
		}
		if priorMode != r.Model.Mode || priorTable != r.Model.Selected {
			fmt.Printf("PF6: %s (table %d)\n", r.Model.Mode, r.Model.Selected)
		}
		if r.Model.Mode == frontend.Quit {
			done = true
			return nil
		}
		if device != nil {
			if priorSource != r.AudioSource() && !r.Model.Suspended() {
				device.Suspend(false)
			}
			if r.Model.Suspended() != suspended {
				suspended = r.Model.Suspended()
				device.Suspend(suspended)
			}
			if err := device.Queue(r.PCM); err != nil {
				return err
			}
		}

		nextSync = nextSync.Add(time.Second / time.Duration(hz))

		return nil
	}

	// Every overdue source sync produces its PCM before another host blit. A
	// slow display may skip intermediate host frames; it never skips game ticks.
	advance := func() error {
		if done {
			return nil
		}
		for !time.Now().Before(nextSync) {
			if duration > 0 && time.Now().After(deadline) {
				done = true
				return nil
			}
			if err := step(); err != nil {
				return err
			}
			if done {
				return nil
			}
		}
		return nil
	}
	// GDI can block in multiple independent calls during a resized frame. Refill
	// deadlines that elapsed between those calls without recursively presenting.
	host.SetSourceTick(advance)
	host.SetModalTick(func() error {
		if done || time.Now().Before(nextSync) {
			return nil
		}
		if err := advance(); err != nil {
			return err
		}
		if done {
			return nil
		}
		return present()
	})
	for {
		host.Pump()
		if pumpHook != nil {
			pumpHook(host)
		}
		if err := advance(); err != nil {
			return err
		}
		if done {
			return nil
		}
		// Shortcut makes are consumed during input collection, but the blocking
		// host transaction waits until every deadline already due has submitted PCM.
		// Recheck before each operation; never generate a future source tick.
		for pendingFullscreen > 0 {
			if err := advance(); err != nil {
				return err
			}
			if done {
				return nil
			}
			pendingFullscreen--
			if err := host.ToggleFullscreen(device); err != nil {
				return err
			}
		}
		// A transition may itself cross source deadlines. Refill before rendering.
		if err := advance(); err != nil {
			return err
		}
		if done {
			return nil
		}
		if err := present(); err != nil {
			return err
		}
		if done {
			return nil
		}
		if remaining := time.Until(nextSync); remaining > 0 {
			time.Sleep(remaining)
		}
	}
}
