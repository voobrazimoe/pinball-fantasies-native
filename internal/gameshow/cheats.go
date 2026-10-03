package gameshow

// SetCheatRules carries FANTASIE's loaded-table globals across NEW_GAME.
// Tilt inhibition and SHIFTKEYS are not reset by WHEN_NEW_GAME_RESET.
func (g *Game) SetCheatRules(tiltDisabled, fastBall bool, balls int) {
	g.Physics.TiltDisabled, g.Physics.FastBall = tiltDisabled, fastBall
	if balls != 0 {
		g.totalBalls = uint8(balls)
	}
}
