// Package engine is the small serialized native-host contract over source.Runner.
package engine

import (
	"errors"
	"image"
	"pinballfantasies/internal/frontend"
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
	runner    *source.Runner
	now       time.Time
	last      int64
	held      gameplay.Controls
	mouse     gameplay.Mouse
	delta     int
	fire      bool
	fullTable bool
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
	e := &Engine{now: time.Unix(0, ns), last: ns}
	e.runner = source.New(rt, func() time.Time { return e.now })
	return e
}
func (e *Engine) submit(edge gameplay.Controls) {
	edge.Left, edge.Right, edge.Down, edge.Tilt = e.held.Left, e.held.Right, e.held.Down, e.held.Tilt
	e.runner.Submit(frontend.Input{Gameplay: edge})
}
func (e *Engine) SetAction(a Action, down bool) error {
	if a > Tilt {
		return errors.New("unknown action")
	}
	if e.runner.Suspended {
		return nil
	}
	var edge gameplay.Controls
	switch a {
	case Left:
		e.held.Left = down
	case Right:
		e.held.Right = down
	case Spring:
		edge.Release = e.held.Down && !down
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
	e.runner.Submit(frontend.Input{Gameplay: e.held, Keys: []frontend.Key{frontend.Key(code)}})
}
func (e *Engine) Release() { e.submit(gameplay.Controls{Release: true}) }
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
func (e *Engine) PlungerFire() {
	if e.MouseActive() {
		e.fire = true
	}
}
func (e *Engine) clearMouse() { e.mouse.Clear(); e.delta = 0; e.fire = false }
func (e *Engine) Suspend() error {
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
		e.submit(gameplay.Controls{MouseY: e.mouse.Motion(e.delta), MouseFire: e.fire})
		e.delta = 0
		e.fire = false
	}, pcm)
}

// Frame borrows the runtime framebuffer. Valid until the next mutating engine
// operation or Frame call. The caller must not write it or use it concurrently.
func (e *Engine) Frame() *image.RGBA { return e.runner.Runtime.FramePresentation(e.fullTable) }

type State struct {
	Tick                         uint64
	Mode                         uint32
	Table                        uint32
	Suspended, Done, MouseActive bool
}

func (e *Engine) State() State {
	m := e.runner.Runtime.Model
	return State{Tick: e.runner.Ticks, Mode: uint32(m.Mode), Table: uint32(m.Selected), Suspended: e.runner.Suspended, Done: e.runner.Done, MouseActive: e.MouseActive()}
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
