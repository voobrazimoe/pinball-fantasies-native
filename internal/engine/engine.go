// Package engine is the small serialized native-host contract over source.Runner.
package engine

import (
	"errors"
	"image"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gamepad"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/source"
	"time"
)

type Action uint32

const (
	Left Action = iota
	Right
	Spring
	Tilt
)

// Engine operations, including Frame and PCM sinks, must be serialized by the
// host. A PCM sink may copy/enqueue samples but must not reenter the engine.
type Engine struct {
	runner      *source.Runner
	now         time.Time
	last        int64
	held        gameplay.Controls
	pad         *gamepad.Input
	padFrame    *image.RGBA
	mouse       gameplay.Mouse
	delta       int
	fire        bool
	fullTable   bool
	touchSet    bool
	touchTarget int
}

func Load(data, state string, ns int64) (*Engine, error) {
	if ns < 0 || state == "" {
		return nil, errors.New("nonnegative clock and state directory required")
	}
	rt, err := frontend.LoadConfigured(data, frontend.FileStore{Directory: state, SeedDirectory: data}, &settings.Store{Directory: state, SeedDirectory: data})
	if err != nil {
		return nil, err
	}
	return New(rt, ns), nil
}

// New wraps an existing authoritative runtime; desktop hosts can keep using
// source.Runner directly. No second implementation of source scheduling exists.
func New(rt *frontend.Runtime, ns int64) *Engine {
	e := &Engine{now: time.Unix(0, ns), last: ns, pad: gamepad.New()}
	e.runner = source.New(rt, func() time.Time { return e.now })
	return e
}
func (e *Engine) submit(edge gameplay.Controls) {
	held := e.controls()
	edge.Left, edge.Right, edge.Down, edge.Tilt = held.Left, held.Right, held.Down, held.Tilt
	e.runner.Submit(frontend.Input{Gameplay: edge})
}
func (e *Engine) SetAction(a Action, down bool) error {
	if a > Tilt {
		return errors.New("unknown action")
	}
	if e.runner.Suspended {
		return nil
	}
	e.controls() // Reconcile a screen change before examining controller holds.
	var edge gameplay.Controls
	switch a {
	case Left:
		e.held.Left = down
	case Right:
		e.held.Right = down
	case Spring:
		edge.Release = e.held.Down && !down && !e.pad.Controls().Down
		e.held.Down = down
	case Tilt:
		e.held.Tilt = down
	}
	e.submit(edge)
	return nil
}

// Key is an original logical DOS make code, never a native OS keycode.
// Hosts suppress OS repeats and submit each physical make once.
func (e *Engine) Key(code uint8) {
	e.runner.Submit(frontend.Input{Gameplay: e.controls(), Keys: []frontend.Key{frontend.Key(code)}})
}
func (e *Engine) Release() {
	e.controls()
	e.submit(gameplay.Controls{Release: !e.pad.Controls().Down})
}
func (e *Engine) MouseActive() bool {
	return !e.runner.Suspended && e.runner.Runtime.Model.MousePlungerActive()
}

// PlungerDelta takes raw relative vertical counts (eight counts per source
// adjustment), discarded outside SPRING_VALID. Magnitude cannot speed charging.
func (e *Engine) PlungerDelta(delta int32) {
	if e.MouseActive() {
		e.delta += int(delta)
	} else {
		e.clearMouse()
	}
}

// PlungerTarget is an additive touch-only absolute charge input. It is latched
// until the next source task, never applied directly to a table from host code.
func (e *Engine) PlungerTarget(target int32) {
	if !e.MouseActive() {
		e.clearMouse()
		return
	}
	if target < 0 {
		target = 0
	}
	if target > 32 {
		target = 32
	}
	e.touchTarget = int(target)
	e.touchSet = true
}
func (e *Engine) PlungerFire() {
	if e.MouseActive() {
		e.fire = true
	}
}
func (e *Engine) clearMouse() { e.mouse.Clear(); e.delta = 0; e.fire = false; e.touchSet = false }
func (e *Engine) Suspend() error {
	e.pad.Focus(false)
	e.held = gameplay.Controls{}
	e.clearMouse()
	return e.runner.LoseFocus()
}
func (e *Engine) clock(ns int64) error {
	if ns < e.last {
		return errors.New("monotonic clock moved backwards")
	}
	e.last = ns
	e.now = time.Unix(0, ns)
	return nil
}
func (e *Engine) Resume(ns int64) error {
	if err := e.clock(ns); err != nil {
		return err
	}
	if e.runner.Suspended {
		e.pad.Focus(true)
		e.held = gameplay.Controls{}
		e.clearMouse()
		e.runner.Resume()
	}
	return nil
}
func (e *Engine) Advance(ns int64, pcm func([]byte) error) error {
	if err := e.clock(ns); err != nil {
		return err
	}
	return e.runner.Advance(func() {
		if !e.MouseActive() {
			e.clearMouse()
		}
		e.submit(gameplay.Controls{MouseY: e.mouse.Motion(e.delta), MouseFire: e.fire, TouchSet: e.touchSet, TouchTarget: e.touchTarget})
		e.touchSet = false
		e.delta = 0
		e.fire = false
	}, pcm)
}

// Frame borrows the runtime framebuffer. Valid until the next mutating engine
// operation or Frame call. The caller must not write it or use it concurrently.
func (e *Engine) Frame() *image.RGBA {
	m := e.runner.Runtime.Model
	frame := e.runner.Runtime.FramePresentation(e.fullTable)
	result := gamepad.Overlay(e.padFrame, frame, e.pad, gamepad.Hints(m.Mode, e.MouseActive(), m.PlayerSelectionOpen(), m.Demo))
	if result != frame {
		e.padFrame = result
	}
	return result
}

type State struct {
	Tick                         uint64
	Mode                         uint32
	Table                        uint32
	Suspended, Done, MouseActive bool
	Gamepad                      bool
}

func (e *Engine) State() State {
	m := e.runner.Runtime.Model
	return State{Tick: e.runner.Ticks, Mode: uint32(m.Mode), Table: uint32(m.Selected), Suspended: e.runner.Suspended, Done: e.runner.Done, MouseActive: e.MouseActive(), Gamepad: e.pad.Connected()}
}

// Close completes the shared frontend shutdown, including settings persistence.
func (e *Engine) Close() error {
	e.held = gameplay.Controls{}
	e.clearMouse()
	e.runner.Suspend()
	e.runner.Submit(frontend.Input{Close: true})
	return e.runner.Advance(nil, nil)
}

// SetPresentation changes only subsequent frame composition.
func (e *Engine) SetPresentation(full bool) { e.fullTable = full }
