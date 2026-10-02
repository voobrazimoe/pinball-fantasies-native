package stones

import "pinballfantasies/internal/tablelogic"

func (g *Game) resetBall() {
	// PLAYER_STRUC keeps target banks, ghosts, kickback, skill score and screams.
	persistent := g.Lights
	g.Lights = [45]bool{}
	g.Lamps = [45]bool{}
	g.flashes = [64]flash{}
	g.palette = g.basePalette()
	g.palette = g.Display.SourceMatrixPalette(g.palette)
	g.matrixOn = true
	for _, n := range append(append([]int{}, stoneBone...), 4, 5, 6) {
		g.light(n, persistent[n])
	}
	if g.kickback {
		g.flash(44, 18)
		g.gate("GATE5", true)
	} else {
		g.gate("GATE5", false)
	}
	for i := 0; i < int(g.GhostCounter); i++ {
		g.light(ghostLamps[i], true)
	}
	if g.ghostFlashing {
		g.flash(ghostLamps[g.GhostCounter], 32)
		g.flash(16, 18)
	}
	g.syncSulp = 0
	g.syncTower = 0
	g.Multiplier = 1
	bonusText := append([]byte(nil), g.Display.Content.Texts["BONUS_TEXT"]...)
	bonusText[11] = '8'
	g.Display.Content.Texts["BONUS_TEXT"] = bonusText
	g.bonusPointer = 39
	g.TowerValue = number(1000000)
	g.VaultValue = number(500000)
	g.WellValue = number(100000)
	g.SulpScore = Decimal{}
	g.GhostTotal = Decimal{}
	g.GrimTotal = Decimal{}
	g.Special = false
	g.GhostHunt = false
	g.Grim = false
	g.TowerHunt = false
	g.MultiDemon = false
	g.timers = [6]uint16{}
	g.TowerStage = 0
	g.ComboStage = 0
	g.towerNext = 0
	g.holdBonus = false
	g.holdMulti = false
	g.inhibit = false
	g.wasSpecial = false
	g.fixTeleport = false
	g.captured = ""
	g.keyDisabled = false
	g.keyRollDisabled = true
	g.ripDisabled = false
	g.ripRollDisabled = false
	g.sbDisabled = false
	g.ghostInhibit = false
	g.sbGuards = [11]bool{}
	g.towerOpen = false
	g.highVault = false
	if g.NextJump == 0 {
		g.NextJump = 10
	}
	g.SkillKey = uint8(g.clock)%3 + 1
	g.flash(int(g.SkillKey), 1)
	g.gate("GATE3", false)
}
func (g *Game) drain() {
	g.Phase = BallLost
	if g.Physics.Configured {
		g.Physics.TargetRaster = g.Physics.BottomRaster()
	} else {
		g.ScreenForce = g.Physics.BottomRaster()
	}
	g.Physics.SetBall(140, 210, 0, 0, false)
	g.Physics.Ball.Hold = true
	g.Physics.AllowFlip = false
	g.Special = false
	g.off(38)
	g.music.Priority = 0
	g.effect("LOSTBALL")
	g.wait("SOUNDRINNER", 5, func() { g.sound("SRINNER") })
	g.emit("BallLost", "LOOSE_BALL", 0)
}
func (g *Game) changeBall() {
	if g.holdBonus {
		g.Bonus = g.heldBonus
	}
	g.savePlayer() // VARS_2_P_STRUC runs before earned-extra-ball selection.
	if g.ExtraBalls > 0 {
		g.ExtraBalls--
		g.beginMatrix("SHOOT_AGAIN_ONTS")
		return
	}
	if g.matchBall {
		g.endGame()
		return
	}
	if !g.advancePlayer() {
		g.beginMatrix("OUT_OF_BALLSTS")
		return
	}
	g.wait("NEW_BALL_TASK", 30, g.newBall)
}
func (g *Game) newBall() {
	g.resetBall()
	g.Physics.ResetTilt()
	g.tasks = [20]func() bool{}
	g.waits = map[string]uint16{}
	g.lastCheck = ""
	g.Physics.Stopped = false
	g.Physics.Ball.Lost = false
	g.Physics.Ball.Hold = true
	g.Physics.SetBall(282, 530, 0, 0, false)
	g.ScreenForce = -1
	g.Physics.TargetRaster = -1
	g.inChute = true
	g.Physics.SpringValid = true
	g.Phase = NewBall
	g.Cue("S_SPRING")
	g.music.ReturnPosition = 0
	if !g.partyFlash {
		g.matrix = matrix{}
		g.Display.Begin("_FLASHOFF", nil)
		g.Display.Visit(0, 0, g.matrixNumber)
	}
	g.wait("SOUNDNEWBALL", 45, func() { g.sound("SNEWBALL") })
	g.wait("SETBALL", 80, func() { g.Physics.SetBall(297, 530, 10, 0, false); g.Physics.Ball.Hold = false; g.Phase = Playing })
	g.playerText()
	// WHEN_NEW_BALL_RESET keeps the source shoot-again/party message.
	if !g.partyFlash {
		g.Display.ShowPlayerBall(g.Score.String())
	}
	g.emit("NewBall", "NEW_BALL", uint64(g.BallNumber))
}
func (g *Game) endGame() {
	g.Phase = GameOver
	g.matrix.active = false
	g.emit("GameOver", "_2_DEMO_MODE", g.Score.Uint64())
}
func (g *Game) runTasks() { tablelogic.Run(g.tasks[:], g.ids[:]) }
