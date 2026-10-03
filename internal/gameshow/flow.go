package gameshow

func (g *Game) resetTable() {
	g.Prizes = [6]uint8{}
	g.BillionEnabled = false
	g.Multiplier = 1
	g.CashPot = number(500000)
	g.MoneyTotal = Decimal{}
	g.Special = false
	g.MoneyMania = false
	g.LoopsAndTraps = false
	g.timers[3] = 0
	g.timers[4] = 0
	g.timers[5] = 0
	g.timers[6] = 0
	g.timers[9] = 0
	g.timers[10] = 0
	g.Lights = [39]bool{}
	g.Lamps = [39]bool{}
	g.flashes = [30]flash{}
	g.palette = g.basePalette()
	g.gate(1, false)
	for n := 7; n <= 10; n++ {
		g.light(n, true)
		g.gate(n-4, false)
	}
	g.flash(1, 15)
	if g.TopThree {
		for i := 0; i < 3; i++ {
			g.Prizes[i] = 2
			g.light(prizeLamps[i], true)
		}
	}
	if g.AllSix {
		for i := 3; i < 6; i++ {
			g.Prizes[i] = 2
			g.light(prizeLamps[i], true)
		}
	}
	g.gate(7, false)
	g.gate(8, false)
}
func (g *Game) drain() {
	g.Phase = BallLost
	if g.Physics.Configured {
		g.Physics.TargetRaster = g.Physics.BottomRaster()
	} else {
		g.ScreenForce = g.Physics.BottomRaster()
	}
	g.Physics.SetBall(135, 28, 0, 0, false)
	g.Physics.Ball.Hold = true
	g.Physics.AllowFlip = false
	g.Special = false
	g.music.Priority = 0
	g.effect("LOSTBALL")
	g.music.ReturnPosition = 55
	g.wait("SOUNDRINNER", 5, func() { g.sound("SRINNER") })
	g.emit("BallLost", "LOOSE_BALL", 0)
}
func (g *Game) changeBall() bool {
	g.savePlayer() // VARS_2_P_STRUC runs before earned-extra-ball selection.
	if g.Lights[31] {
		g.startMatrix("SHOOT_AGAIN_ONTS", false)
		g.emit("ShootAgain", "LET_HIM_SHOOT_AGAIN", 0)
		return false
	}
	if g.matchBall {
		g.startMatrix("AFTER_XXBALLTS", false)
		return false
	}
	if !g.advancePlayer() {
		g.startMatrix("OUT_OF_BALLSTS", false)
		return false
	}
	g.wait("NEW_BALL_TASK", 60, g.newBall)
	return true // HU_ tail-calls the following matrix handler on this sync.
}
func (g *Game) newBall() {
	g.resetTable()
	g.Physics.ResetTilt()
	g.tasks = [20]func() bool{}
	g.waits = map[string]uint16{}
	if !g.partyFlash {
		g.matrix = matrix{}
		g.Display.Begin("_FLASHOFF", nil)
		g.Display.Visit(0, 0, g.matrixNumber)
	}
	g.lastArea = ""
	g.lastCheck = ""
	g.Physics.Stopped = false
	g.Physics.Ball.Lost = false
	g.Physics.Ball.Hold = true
	g.Physics.Ball.HitX = 0
	g.Physics.Ball.HitY = 0
	g.Physics.SetBall(284, 530, 0, 0, false)
	g.ScreenForce = -1
	g.Physics.TargetRaster = -1
	g.inChute = true
	g.Physics.SpringValid = true
	g.Phase = NewBall
	g.Cue("S_SPRING")
	g.music.ReturnPosition = 0
	g.wait("SOUNDNEWBALL", 50, func() { g.sound("SNEWBALL") })
	g.wait("SETBALL", 80, func() { g.Physics.SetBall(299, 530, 10, 0, false); g.Physics.Ball.Hold = false; g.Phase = Playing })
	g.wait("SOUNDBRICKUPP", 5, func() { g.sound("SBRICKUPP") })
	g.playerText()
	// WHEN_NEW_BALL_RESET keeps the source shoot-again/party message.
	if !g.partyFlash {
		g.beginMatrix("SHOWPLAYERSTS") // WHEN_NEW_BALL_RESET calls DO_MATRIX.
	}
	g.emit("NewBall", "NEW_BALL", uint64(g.BallNumber))
}
func (g *Game) endGame() {
	g.Phase = GameOver
	g.emit("GameOver", "_2_DEMO_MODE", g.Score.Uint64())
}
