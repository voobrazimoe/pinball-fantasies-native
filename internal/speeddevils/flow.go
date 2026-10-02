package speeddevils

func (g *Game) drain() {
	g.Phase = BallLost
	g.emit("BallLost", "LOOSE_BALL", 0)
	if g.Physics.Configured {
		g.Physics.TargetRaster = g.Physics.BottomRaster()
	} else {
		g.ScreenForce = g.Physics.BottomRaster()
	}
	g.Physics.SetBall(280, 560, 0, 0, false)
	g.Physics.Ball.Hold = true
	g.Special = false
	g.music.Priority = 0
	g.effect("LOSTBALL")
	g.wait("SOUNDRINNER", 5, func() { g.sound("SRINNER") })
}
func (g *Game) changeBall() {
	if g.HoldBonus {
		g.Bonus = g.matrix.held
	}
	g.savePlayer() // VARS_2_P_STRUC runs before earned-extra-ball selection.
	if g.Lights[55] {
		g.emit("ShootAgain", "LET_HIM_SHOOT_AGAIN", 0)
		g.beginMatrix("SHOOT_AGAIN_ONTS")
		return
	}
	if g.matchBall {
		g.Phase = GameOver
		g.emit("GameOver", "AFTER_XXBALLTS", 0)
		return
	}
	if !g.advancePlayer() {
		g.beginMatrix("OUT_OF_BALLSTS")
		return
	}
	g.wait("NEW_BALL_TASK", 60, g.newBall)
}
func (g *Game) newBall() {
	if g.Session.PlayerCount > 1 {
		g.OffRoadTotal, g.TurboTotal = Decimal{}, Decimal{}
	}
	g.Physics.ResetTilt()
	// P_STRUC_2_VARS preserves score, bonus, miles, speed, gear, position,
	// car parts and mode totals; RESET_VARS clears only temporary ball electronics.
	keep := g.Lights
	g.Lights = [68]bool{}
	g.Lamps = [68]bool{}
	g.lampPalette = g.paletteFor(g.Lamps)
	g.flashes = [64]flash{}
	g.tasks = [20]func() bool{}
	g.waits = map[string]uint16{}
	g.matrix = matrix{}
	for _, n := range []int{9, 10, 11, 12, 13, 14, 22, 23, 24, 25, 50, 51, 52, 53, 54} {
		g.light(n, keep[n])
	}
	for n := uint16(0); n < g.Gear; n++ {
		g.light(int(26+n), true)
	}
	for n := uint16(0); n < g.Position; n++ {
		g.light(int(32+n), true)
	}
	for n := g.Position; n < g.PositionAvailable; n++ {
		g.syncFlash(int(32+n), n%2 != 0)
	}
	for n := uint16(0); n < g.Speed && n < 12; n++ {
		g.light(int(56+n), true)
	}
	g.Multiplier = 1
	g.MBStock, g.MBCollected = 0, 0
	g.pitDown = [3]uint16{}
	g.pitAll, g.gearDown, g.loopHigh, g.loopLow = 0, 0, 0, 0
	g.HoldBonus, g.Special, g.OffRoad, g.Turbo = false, false, false, false
	g.touchBusy = [6]bool{}
	g.snackDisabled = false
	g.loopH, g.loopL, g.jump = false, false, false
	g.lastArea, g.lastCheck = "", ""
	g.posSync = 0
	g.inChute = true
	g.ModeTime = 0
	g.inhibitCountdown = false
	g.Physics.Stopped = false
	g.Physics.Ball.Lost = false
	g.Physics.Ball.Hold = true
	g.Physics.Ball.HitX, g.Physics.Ball.HitY = 0, 0
	g.Physics.SetBall(285, 530, 0, 0, false)
	g.ScreenForce = -1
	g.Physics.TargetRaster = -1
	g.Physics.SpringValid = true
	g.Phase = NewBall
	g.scoreChanged = false
	g.Cue("S_SPRING")
	g.music.ReturnPosition = 0
	g.wait("SOUNDNEWBALL", 50, func() { g.sound("SNEWBALL") })
	g.wait("SOUNDBRICKUPP", 5, func() { g.sound("SBRICKUPP") })
	g.wait("SETBALL", 80, func() { g.Physics.SetBall(300, 530, 10, 0, false); g.Physics.Ball.Hold = false; g.Phase = Playing })
	g.playerText()
	g.Display.ShowPlayerBall(g.Score.String())
	g.emit("NewBall", "NEW_BALL", uint64(g.BallNumber))
}
func (g *Game) checkHighScore() bool {
	if g.inChute || g.Special {
		return false
	}
	return g.beatHighScore()
}

// _BEATEN_MATRIX at ball loss omits CHECKHIGHSCORE's chute/mode guards.
func (g *Game) beatHighScore() bool {
	if g.top == nil || g.beaten || g.Score.Uint64() <= g.top.Uint64() {
		return false
	}
	g.beaten = true
	g.emit("HighScoreBeaten", "CHECKHIGHSCORE", g.Score.Uint64())
	return true
}
