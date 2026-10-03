//go:build matrixdebug

package partyland

// TestSideLaneExtraBall modifies only the reproducer's initial conditions.
// AREALISTA/BYGEL1/2, DO_PHYSICS/drain and LOOSE_BALL still run normally.
func (g *Game) TestSideLaneExtraBall(lane int) bool {
	if g.Phase != Playing || lane != 0 && (g.inChute || g.Physics.Ball.Hold || g.Physics.Ball.Lost) {
		return false
	}
	g.light(39, true)
	g.light(40, true)
	if lane != 0 {
		x := int16(2)
		if lane > 0 {
			x = 281
		}
		g.Physics.SetBall(x, 447, 0, 700, false)
	}
	return true
}
