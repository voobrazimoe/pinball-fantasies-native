package physics

// TILTLOGIC is a make-edge latch distinct from TILT0's held physical push.
// DO_PHYSICS decays TILTCOUNTER once per sync before electronics/key tasks.
func (g *Game) TiltInput(held, inChute bool) (warning, tilted bool) {
	edge := held && !g.tiltLatched
	g.tiltLatched = held
	if !edge || inChute || g.Tilted || g.Ball.Lost {
		return
	}
	g.TiltCounter += 60
	if g.TiltCounter > 120 {
		g.Tilted = true
		g.AllowFlip = false
		return false, true
	}
	return g.TiltCounter > 60, false
}
func (g *Game) ResetTilt() { g.Tilted = false; g.TiltCounter = 0; g.AllowFlip = true }
func (g *Game) push(held bool) {
	// BALLCODE/TILT0: unsigned JBE on upward motion, signed JGE on return.
	if held {
		g.ScreenSpeed = 600
		g.ScreenPosition += 600
		if uint16(g.ScreenPosition) > 2048 {
			g.ScreenSpeed = 0
			g.ScreenPosition = 2048
		}
	} else {
		g.ScreenSpeed = -200
		g.ScreenPosition -= 200
		if g.ScreenPosition < 0 {
			g.ScreenSpeed = 0
			g.ScreenPosition = 0
		}
	}
	g.ScreenOffset = int16(uint16(g.ScreenPosition) >> 9)
}
