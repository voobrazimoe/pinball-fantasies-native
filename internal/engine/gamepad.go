package engine

import (
	"errors"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gamepad"
	"pinballfantasies/internal/gameplay"
)

func (e *Engine) controls() gameplay.Controls {
	e.pad.Mode(e.runner.Runtime.Model.Mode)
	p := e.pad.Controls()
	return gameplay.Controls{Left: e.held.Left || p.Left, Right: e.held.Right || p.Right,
		Down: e.held.Down || p.Down, Tilt: e.held.Tilt || p.Tilt}
}

// Gamepad accepts standardized host input, serialized like every engine call.
// kind: 0 button, 1 trigger axis, 2 connection snapshot, 3 disconnect,
// 4 controller family, 5 face-button glyph.
func (e *Engine) Gamepad(kind, a, b int) error {
	e.pad.Mode(e.runner.Runtime.Model.Mode)
	var input frontend.Input
	switch kind {
	case 0:
		if a < 0 || a > 16 || (b != 0 && b != 1) {
			return errors.New("invalid gamepad button")
		}
		input = e.pad.Button(a, b != 0)
	case 1:
		if (a != 4 && a != 5) || b < 0 || b > 32767 {
			return errors.New("invalid gamepad trigger")
		}
		input = e.pad.Axis(a, b)
	case 2:
		if a < 0 || a >= 1<<17 {
			return errors.New("invalid gamepad snapshot")
		}
		e.pad.Connect(a)
	case 3:
		e.pad.Disconnect()
	case 4:
		if a < 0 || a > 3 {
			return errors.New("invalid gamepad family")
		}
		e.pad.SetFamily(gamepad.Family(a))
	case 5:
		if a < 0 || a > 3 || b < 0 || b > 8 {
			return errors.New("invalid gamepad glyph")
		}
		e.pad.SetGlyph(a, gamepad.Glyph(b))
	default:
		return errors.New("invalid gamepad event")
	}
	p := e.controls()
	input.Gameplay.Left, input.Gameplay.Right = p.Left, p.Right
	input.Gameplay.Down, input.Gameplay.Tilt = p.Down, p.Tilt
	input.Gameplay.Release = input.Gameplay.Release && !e.held.Down
	e.runner.Submit(input)
	return nil
}
