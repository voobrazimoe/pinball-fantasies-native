//go:build matrixdebug

package frontend

import (
	"fmt"
	"os"
	"pinballfantasies/internal/partyland"
)

func matrixTestCheat(m *Model, key Key) bool {
	if m.Mode != Playing || m.Selected != 1 {
		return false
	}
	g, ok := m.Session.(*partyland.Game)
	if !ok {
		return false
	}
	lane := 0
	switch key {
	case 45: // X: arm the original side-lane lamps.
	case 38:
		lane = -1 // L: place the launched ball at the left lane.
	case 19:
		lane = 1 // R: place the launched ball at the right lane.
	default:
		return false
	}
	accepted := g.TestSideLaneExtraBall(lane)
	fmt.Fprintf(os.Stderr, "MATRIX TEST tick=%d lane=%d accepted=%t\n", g.Tick, lane, accepted)
	return true
}

func matrixTestTrace(m *Model) {
	if g, ok := m.Session.(*partyland.Game); ok {
		for _, event := range g.Events {
			fmt.Fprintf(os.Stderr, "MATRIX TEST %+v op=%s\n", event, g.Display.CurrentOperation())
		}
	}
}
