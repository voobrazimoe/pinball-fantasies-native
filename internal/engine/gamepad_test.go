package engine

import (
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gamepad"
	"testing"
)

func TestGamepadSourceChargeAndOwners(t *testing.T) {
	e, s := inputEngine(t)
	e.SetAction(Left, true)
	e.Gamepad(0, gamepad.LeftShoulder, 1)
	e.SetAction(Left, false)
	e.Gamepad(0, gamepad.X, 1)
	for i := 0; i < 10; i++ {
		e.Gamepad(0, gamepad.X, 1)
	}
	tickEngine(t, e)
	if !s.inputs[0].Left || s.position != 1 {
		t.Fatal("owner merge or source charging", s)
	}
	e.SetAction(Spring, true)
	e.Gamepad(0, gamepad.X, 0)
	tickEngine(t, e)
	if len(s.launches) != 0 || s.position != 2 {
		t.Fatal("controller release fired another owner's spring", s)
	}
	e.SetAction(Spring, false)
	tickEngine(t, e)
	if len(s.launches) != 1 || s.launches[0] != 2 {
		t.Fatal("final owner release lost", s.launches)
	}
	e.Gamepad(0, gamepad.LeftShoulder, 0)
	tickEngine(t, e)
	if s.inputs[len(s.inputs)-1].Left {
		t.Fatal("flipper stuck")
	}
}

func TestGamepadDisconnectAndSuspend(t *testing.T) {
	e, s := inputEngine(t)
	e.Gamepad(0, gamepad.X, 1)
	tickEngine(t, e)
	e.Gamepad(3, 0, 0)
	tickEngine(t, e)
	if len(s.launches) != 0 || s.inputs[len(s.inputs)-1].Down {
		t.Fatal("disconnect fired or stuck")
	}
	e.Gamepad(0, gamepad.LeftShoulder, 1)
	e.Suspend()
	e.Gamepad(0, gamepad.Y, 1)
	e.Resume(e.last)
	e.runner.Runtime.Model.Mode = frontend.Playing
	tickEngine(t, e)
	last := s.inputs[len(s.inputs)-1]
	if last.Left || last.Tilt {
		t.Fatal("background/retained press leaked", last)
	}
	e.Gamepad(0, gamepad.LeftShoulder, 0)
	e.Gamepad(0, gamepad.LeftShoulder, 1)
	tickEngine(t, e)
	if !s.inputs[len(s.inputs)-1].Left {
		t.Fatal("fresh press suppressed")
	}
}

func TestGamepadInvalidContract(t *testing.T) {
	e, _ := inputEngine(t)
	for _, v := range [][3]int{{0, 17, 1}, {0, 0, 2}, {1, 0, 10}, {1, 4, -1}, {1, 5, 32768}, {2, 1 << 17, 0}, {6, 0, 0}, {4, 4, 0}, {5, 4, 0}, {5, 0, 9}} {
		if e.Gamepad(v[0], v[1], v[2]) == nil {
			t.Fatal("invalid accepted", v)
		}
	}
}
