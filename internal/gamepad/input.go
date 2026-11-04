// Package gamepad translates standard controller input at the host boundary.
package gamepad

import (
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameplay"
)

// Button values match SDL2's standard GameController buttons, not raw joystick
// indices. Labels refer to Xbox/Steam Deck positions.
const (
	A = iota
	B
	X
	Y
	Back
	Guide
	Start
	LeftStick
	RightStick
	LeftShoulder
	RightShoulder
	Up
	Down
	Left
	Right
	leftTrigger
	rightTrigger
)

// Input retains physical presses independently of game actions. Mode/focus
// changes clear actions; a button held across either boundary needs a fresh
// press before it can act again.
type Input struct {
	buttons   [17]bool
	held      [17]bool
	mode      frontend.Mode
	focused   bool
	connected bool
	family    Family
	glyphs    [4]Glyph
}

func New() *Input { return &Input{focused: true} }

func (p *Input) clear() { p.held = [17]bool{} }

func (p *Input) Mode(mode frontend.Mode) {
	if p.mode != mode {
		p.clear()
		p.mode = mode
	}
}

func (p *Input) Focus(focused bool) { p.clear(); p.focused = focused }

// Disconnect clears physical history too, without generating a launch edge.
func (p *Input) Disconnect() {
	p.clear()
	p.buttons = [17]bool{}
	p.connected = false
	p.family = Generic
	p.glyphs = [4]Glyph{}
}

// Connect seeds only physical history, preventing already held controls from
// becoming makes when a device is opened or replaced.
func (p *Input) Connect(buttons int) {
	p.Disconnect()
	p.connected = true
	for i := range p.buttons {
		p.buttons[i] = buttons&(1<<i) != 0
	}
}

func (p *Input) Controls() gameplay.Controls {
	return gameplay.Controls{Left: p.held[LeftShoulder] || p.held[leftTrigger], Right: p.held[RightShoulder] || p.held[rightTrigger], Down: p.held[X], Tilt: p.held[Y]}
}

// Axis accepts SDL's normalized trigger axes (4/5). Hysteresis prevents noise
// around the threshold from producing flipper chatter.
func (p *Input) Axis(axis, value int) frontend.Input {
	button := leftTrigger
	if axis == 5 {
		button = rightTrigger
	} else if axis != 4 {
		return frontend.Input{}
	}
	down := p.buttons[button]
	if value >= 16000 {
		down = true
	} else if value <= 12000 {
		down = false
	}
	return p.Button(button, down)
}

func (p *Input) Button(button int, down bool) frontend.Input {
	var out frontend.Input
	if button < 0 || button >= len(p.buttons) || p.buttons[button] == down {
		return out
	}
	p.buttons[button] = down
	if !down {
		out.Gameplay.Release = button == X && p.held[X]
		p.held[button] = false
		return out
	}
	if !p.focused {
		return out
	}
	key := frontend.Key(0)
	switch p.mode {
	case frontend.Startup:
		if button == A || button == Start {
			key = frontend.Space
		}
	case frontend.Selector, frontend.SelectorText:
		switch button {
		case A:
			key = frontend.F1
		case B:
			key = frontend.F2
		case X:
			key = frontend.F3
		case Y:
			key = frontend.F4
		case Start:
			key = frontend.F5
		case Back:
			key = frontend.Escape
		}
	case frontend.Options:
		switch button {
		case Up:
			key = frontend.Up
		case Down:
			key = frontend.Down
		case A:
			key = frontend.Enter
		case B, Back:
			key = frontend.Escape
		}
	case frontend.TableAttract:
		switch button {
		case A:
			key = frontend.F1
		case X:
			key = frontend.Enter
		case B, Back:
			key = frontend.Escape
		}
	case frontend.Playing:
		switch button {
		case LeftShoulder, RightShoulder, leftTrigger, rightTrigger, X, Y:
			p.held[button] = true
		case A:
			key = frontend.Enter
		case Start:
			key = frontend.P
		case Back:
			key = frontend.Escape
		}
	case frontend.Paused:
		if button == A || button == Start {
			key = frontend.Space
		}
		if button == B || button == Back {
			key = frontend.Escape
		}
	case frontend.QuitQuestion:
		if button == A {
			key = 21
		} // DOS Y
		if button == B || button == Back {
			key = 49
		} // DOS N
	case frontend.Closing:
		if button == B || button == Back {
			key = frontend.Escape
		}
	}
	if key != 0 {
		out.Keys = []frontend.Key{key}
	}
	return out
}
