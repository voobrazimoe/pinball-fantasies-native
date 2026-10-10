package platform

import (
	"image"
	"os"
	"pinballfantasies/internal/diagnostics"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gamepad"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/source"
	"runtime"
	"time"
)

type PhysicsControls struct{ Left, Right, Down, Release, Tilt, MusicToggle bool }

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
	runner := source.New(r, time.Now)
	noPresent := os.Getenv("PF12_NO_PRESENT") == "1"
	if os.Getenv("PF12_AUDIT_FULLSCREEN") == "1" {
		if e := host.presentation.ToggleFullscreen(); e != nil {
			return e
		}
	}
	pad := gamepad.New()
	var padFrame *image.RGBA
	present := func() error {
		start := time.Now()
		frame := runner.Frame()
		m := r.Model
		composed := gamepad.Overlay(padFrame, frame, pad, gamepad.Hints(m.Mode, !runner.Suspended && m.MousePlungerActive(), m.PlayerSelectionOpen(), m.Demo))
		if composed != frame {
			padFrame = composed
		}
		frame = composed
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

	pendingFullscreen := 0
	mouse := gameplay.Mouse{}
	priorMouseActive := false
	collect := func() {
		if duration > 0 && time.Now().After(deadline) {
			runner.Done = true
			return
		}
		mouseActive := !runner.Suspended && r.Model.MousePlungerActive()
		if mouseActive != priorMouseActive {
			mouse.Clear()
			priorMouseActive = mouseActive
		}
		host.MouseActive(mouseActive)
		input := frontend.Input{}
		padRelease := false
		pad.Mode(r.Model.Mode)
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
				input.Gameplay.Release = true
			case 4:
				keys.enterUp()
			case 8:
				pad.Focus(true)
				runner.Resume()
			case 9, 10:
				var translated frontend.Input
				if event == 9 {
					translated = pad.Button(e.key, e.mouseY != 0)
				} else {
					translated = pad.Axis(e.key, e.mouseY)
				}
				input.Keys = append(input.Keys, translated.Keys...)
				padRelease = padRelease || translated.Gameplay.Release
			case 11:
				pad.Disconnect()
			case 12:
				pad.Connect(e.key)
			case 13:
				pad.SetFamily(gamepad.Family(e.key))
			case 6:
				input.Gameplay.MouseY += e.mouseY
			case 7:
				input.Gameplay.MouseFire = true
			case 5:
				pad.Focus(false)
				padRelease = false
				keys.enterUp()
				mouse.Clear()
				runner.Submit(frontend.Input{FocusLost: true})
				input = frontend.Input{Close: input.Close}
			}
		}
		held := host.Controls()
		controller := pad.Controls()
		input.Gameplay.Left = held.Left || controller.Left
		input.Gameplay.Right = held.Right || controller.Right
		input.Gameplay.Down = held.Down || controller.Down
		input.Gameplay.Tilt = held.Tilt || controller.Tilt
		// Launch only after every input source has released the plunger.
		input.Gameplay.Release = (input.Gameplay.Release && !controller.Down) || (padRelease && !held.Down)
		input.Gameplay.MouseY = mouse.Motion(input.Gameplay.MouseY)
		runner.Submit(input)
	}
	priorSource := r.AudioSource()
	priorMode, priorTable := r.Model.Mode, r.Model.Selected
	submitPCM := func(pcm []byte) error {
		if priorMode != r.Model.Mode || priorTable != r.Model.Selected {
			diagnostics.Printf("PF6: %s (table %d)\n", r.Model.Mode, r.Model.Selected)
		}
		priorMode, priorTable = r.Model.Mode, r.Model.Selected
		if runner.Done {
			return nil
		}
		if device != nil {
			if priorSource != r.AudioSource() && !r.Model.Suspended() {
				device.Suspend(false)
			}
			if (r.Model.Suspended() || runner.Suspended) != suspended {
				suspended = r.Model.Suspended() || runner.Suspended
				device.Suspend(suspended)
			}
			if err := device.Queue(pcm); err != nil {
				return err
			}
		}
		priorSource = r.AudioSource()
		return nil
	}
	// All due PCM reaches the host before any blocking presentation transaction.
	advance := func() error {
		if duration > 0 && time.Now().After(deadline) {
			runner.Done = true
			return nil
		}
		start := time.Now()
		err := runner.Advance(collect, submitPCM)
		audit.mark("update", start, host.presentation.IsFullscreen(), device)
		return err
	}

	// GDI can block in multiple independent calls during a resized frame. Refill
	// deadlines that elapsed between those calls without recursively presenting.
	host.SetSourceTick(advance)
	host.SetModalTick(func() error {
		if runner.Done || time.Now().Before(runner.Next) {
			return nil
		}
		if err := advance(); err != nil {
			return err
		}
		if runner.Done {
			return nil
		}
		return present()
	})
	for {
		host.Pump()
		if pumpHook != nil {
			pumpHook(host)
		}
		collect()
		if err := advance(); err != nil {
			return err
		}
		if runner.Done {
			return nil
		}
		// Shortcut makes are consumed during input collection, but the blocking
		// host transaction waits until every deadline already due has submitted PCM.
		// Recheck before each operation; never generate a future source tick.
		for pendingFullscreen > 0 {
			if err := advance(); err != nil {
				return err
			}
			if runner.Done {
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
		if runner.Done {
			return nil
		}
		if err := present(); err != nil {
			return err
		}
		if runner.Done {
			return nil
		}
		if runner.Suspended {
			time.Sleep(time.Second / 60)
		} else if remaining := time.Until(runner.Next); remaining > 0 {
			time.Sleep(remaining)
		}
	}
}
