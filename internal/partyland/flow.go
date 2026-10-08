package partyland

func (g *Game) drain() {
	g.Phase = BallLost
	g.emit("BallLost", "LOOSE_BALL", 0)
	g.Physics.SetBall(15, 47, 0, 0, false)
	g.Physics.Ball.Hold = true
	g.Happy = false
	g.Mega = false
	g.ModeTime = 0
	if !g.ScoreChanged {
		g.Audio.Priority = 0
		g.beginMatrix("PARTY_ONTS")
		g.music("S_SPRING")
		g.musicOK = true
		g.waitAt("PARTY_ON_TASK1", 30, g.newBall)
		return
	}
	g.Audio.Priority = 0
	if !g.demoScoredDrain() {
		g.effect("LOSTBALL", 0, 0)
	}
	g.Audio.Priority = 0
	g.Audio.ReturnPosition = 62
	g.waitAt("SOUNDRINNER", 5, func() { g.sound("SRINNER") })
}
func (g *Game) changeBall() bool {
	g.savePlayer() // VARS_2_P_STRUC runs before earned-extra-ball selection.
	if g.Lights[51] {
		g.ExtraBalls--
		g.emit("ShootAgain", "LET_HIM_SHOOT_AGAIN", 0)
		g.startMatrix("SHOOT_AGAIN_ONTS", false)
		return false
	}
	if g.matchBall {
		g.startMatrix("AFTER_XXBALLTS", false)
		return false
	}
	if !g.advancePlayer() {
		g.endGameMatch()
		return false
	}
	g.waitAt("NEW_BALL_TASK", 30, g.newBall)
	return true // HU_ tail-calls the following matrix handler on this sync.
}
func (g *Game) endGameMatch() { g.startMatrix("OUT_OF_BALLSTS", false) }
func (g *Game) newBall() {
	// NEW_BALL does not reset flipper angles, spin or ramp gravity.
	g.Physics.Stopped = false
	g.Physics.Ball.Lost = false
	g.Physics.Ball.HitX = 0
	g.Physics.Ball.HitY = 0
	g.resetBall()
	g.demoNewBall()
	g.HappyTotal = Decimal{}
	g.MegaTotal = Decimal{}
	g.Phase = NewBall
	g.Physics.Raster = (g.Physics.BottomRaster() + 33) * 16
	g.Physics.Ball.Hold = true
	g.Physics.SetBall(282, 530, 0, 0, false)
	if !g.musicOK {
		g.music("S_SPRING")
	}
	g.Audio.ReturnPosition = 0
	g.ScreenForce = -1
	g.Physics.TargetRaster = -1
	g.playerText()
	if !g.partyFlash {
		g.beginMatrix(g.playerPanel()) // WHEN_NEW_BALL_RESET calls DO_MATRIX.
	}
	g.emit("NewBall", "NEW_BALL", uint64(g.BallNumber))
	g.waitAt("SOUNDNEWBALL", 50, func() { g.sound("SNEWBALL") })
	g.waitAt("SETBALL", 80, func() { g.Physics.SetBall(297, 530, 10, 0, false); g.Physics.Ball.Hold = false; g.Phase = Playing })
	g.waitAt("SOUNDBRICKUPP", 5, func() { g.sound("SBRICKUPP") })
}
