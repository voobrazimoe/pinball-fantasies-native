package partyland

func (g *Game) special() bool { return g.Happy || g.Mega }
func (g *Game) startDrop() {
	g.Physics.SetBall(15, 47, 0, 0, false)
	g.cameraDrop()
	g.waitAt("DROPTASK1", 30, func() {
		g.flash(3, 7, 0, false)
		g.flash(56, 7, 0, false)
		g.waitAt("DROPTASK2", 27, func() {
			g.Physics.SetBall(15, 47, 0, int16(g.clock&127), true)
			g.Physics.Ball.Hold = false
			g.ScreenForce = -1
			g.sound("SNEWBALL")
			g.endFlash(3)
			g.endFlash(56)
		})
	})
}
func (g *Game) hidden() {
	g.effectEnded = false
	g.effect("HIDDEN", 50000, 20000)
	g.jackAdd()
	g.flash(11, 6, 0, false)
	g.FiveX = true
	g.Physics.SetBall(15, 47, 0, 0, false)
	g.Physics.Ball.Hold = true
	g.task(func() bool {
		if !g.effectEnded {
			return false
		}
		g.waitAt("HIDDENTASK1", 480, func() {
			if g.FiveX {
				g.endFlash(11)
				g.flash(11, 2, 0, false)
				g.waitAt("HIDDENTASK2", 120, g.killFiveX)
			} else {
				g.off(11)
			}
		})
		g.startDrop()
		g.emit("TaskReady", "HIDDENTASK0", 0)
		return true
	})
}
func (g *Game) tunnel() {
	defer func() { g.inhibitEffect = false }()
	g.jackAdd()
	g.addMega()
	if g.SkillTime != 0 {
		g.jackAdd()
		g.SkillTime = 0
		g.SkillTunnel.AddNumber(1000000)
		g.score("SKILL_SCORE", g.SkillTunnel.Uint64())
		g.effect("SKILLTUNNEL", 0, 0)
		g.inhibitEffect = true
		g.party(0)
	} else if g.ReverseTime != 0 {
		g.party(0)
	}
	if g.Lights[12] {
		g.effect("TSCORE3", 5000000, 500000)
		g.TunnelTime = 1440
	} else if g.Lights[14] {
		g.effect("TSCORE2", 3000000, 250000)
		g.TunnelTime = 1440
		g.endFlash(12)
		g.light(12, true)
		g.flash(10, 8, 0, false)
	} else {
		g.effect("TSCORE1", 1000000, 25000)
		g.TunnelTime = 720
		g.endFlash(14)
		g.light(14, true)
		g.flash(12, 8, 0, false)
	}
	g.cameraDrop()
	g.waitUntil("WAIT_FOR_TUNNEL_EFFECT", 130, g.special, g.startDrop)
	if !g.special() {
		g.waitAt("LOCK_BALL_IN_TUNNEL", 2, func() { g.Physics.SetBall(15, 47, 0, 0, false); g.Physics.Ball.Hold = true })
	}

}
func (g *Game) snack() {
	if g.SnackDisabled {
		return
	}
	g.SnackDisabled = true
	g.score("BCD50000", 50000)
	g.bonus(5000)
	g.addMega()
	switch {
	case g.snacks[2]:
		g.effect("SCORE3", 1000000, 100000)
		g.jackAdd()
		g.Pop++
	case g.snacks[1]:
		g.effect("SCORE2", 500000, 50000)
		g.jackAdd()
	case g.snacks[0]:
		g.effect("SCORE1", 250000, 25000)
		g.jackAdd()
	default:
		g.effect("HSCORE", 50000, 40)
	}
	for _, n := range snackLamps {
		g.off(n)
	}
	if g.snacks[2] {
		g.party(1)
		if g.Pop <= 1 {
			if !g.HB {
				g.HB = true
				g.flash(33, 12, g.arrowSync, true)
			}
		} else if !g.DB {
			g.DB = true
			g.flash(36, 12, g.arrowSync, false)
			g.waitAt("TURNING_OFF_DB", 480, func() {
				if g.DB {
					g.endFlash(36)
					g.flash(36, 2, 0, false)
					g.waitAt("TURN_OFF_DB", 120, func() {
						if g.DB {
							g.endFlash(36)
							g.DB = false
						}
					})
				}
			})
		}
	}
	g.snacks = [3]bool{}
	g.Physics.Ball.Hold = true
	g.Physics.SetBall(3, 253, 0, 0, true)
	g.task(func() bool {
		g.Physics.Ball.Hold = true
		g.Physics.SetBall(3, 253, 0, 0, true)
		if !g.waitReady("SNACK_HOLE_TASK1", 40) {
			return false
		}
		g.task(func() bool {
			g.Physics.Ball.Hold = true
			g.Physics.SetBall(3, 253, 0, 0, true)
			if !g.special() && !g.waitReady("SNACK_HOLE_TASK1B", 90) {
				return false
			}
			g.sound("SGROP")
			g.Physics.Ball.Hold = false
			g.Physics.SetBall(3, 253, 0, -2500, true)
			g.waitAt("TURN_IT_ON_AGAIN", 60, func() { g.SnackDisabled = false })
			return true
		})
		return true
	})
}
func (g *Game) dragon() {
	if g.Dragon {
		return
	}
	g.Dragon = true
	g.Physics.SetBall(257, 310, 0, 0, false)
	g.Physics.Ball.Hold = true
	g.addMega()
	delay := uint16(85)
	switch {
	case g.JackpotNormal || g.JackpotTimed:
		g.JackpotNormal = false
		g.JackpotTimed = false
		g.off(35)
		g.music("SJINGLE4")
		label := "JACKPOTTS"
		if g.Happy {
			label = "JACKPOT_SPECIAL_HH_TS"
		} else if g.Mega {
			label = "JACKPOT_SPECIAL_ML_TS"
		}
		g.beginMatrix(label)
		g.Score.Add(g.Jackpot)
		g.emit("ScoreAwarded", "JACKPOT", g.Jackpot.Uint64())
		g.Jackpot = Number(10000000)
		delay = 410
	case g.BallFeature:
		g.off(30)
		g.light(51, true)
		g.ExtraBalls++
		g.effect("EXTRABALL1", 10000, 5000)
		g.BallFeature = false
		delay = 320
		if g.special() {
			delay = 15
		}
	case g.FiveMillion:
		g.effect("MILLION5", 5000000, 10000)
		g.off(27)
		g.FiveMillion = false
		delay = 160
		if g.special() {
			delay = 15
		}
	default:
		g.effect("DSCORE", 250000, 10030)
	}
	g.waitAt("BEFOREFLASH", delay, func() {
		g.flash(21, 7, 0, false)
		g.waitAt("DURINGFLASH", 27, func() {
			g.endFlash(21)
			g.sound("SNEWBALL")
			g.Physics.Ball.Hold = false
			g.Physics.SetBall(257, 310, -575, 1575, false)
			g.Dragon = false
		})
	})
}
func (g *Game) arcade() {
	if g.Physics.Tilted {
		g.startDrop()
		return
	}
	g.Random += 21
	g.addMega()
	if !g.Arcade {
		g.startDrop()
		return
	}
	g.Arcade = false
	g.off(7)
	g.off(55)
	g.savedJingle = g.Audio.ReturnPosition
	g.Display.SetJingleCountdown(25) // PLAND GROPB sets DECCOR before EFFECT MYSTERY.
	g.effect("MYSTERY", 0, 0)
	g.Physics.SetBall(15, 47, 0, 0, false)
	g.Physics.Ball.Hold = true
	if !g.effectAccepted {
		g.spin()
		g.cameraDrop()
		return
	}
	g.spinReady = false
	g.Audio.ReturnPosition = 62
	g.Audio.Priority = 1
	g.Audio.ReadyLogic = false
	g.Audio.ReadyAnim = false
	g.task(func() bool {
		if !g.special() && !g.spinReady {
			return false
		}
		g.Audio.ReturnPosition = g.savedJingle
		g.Audio.JumpCount = 1
		g.Audio.Priority = 0
		g.spin()
		g.emit("TaskReady", "WAIT_FOR_SPIN_TASK", 0)
		return true
	})
	g.cameraDrop()
}
func (g *Game) spin() {
	r := int(g.Random>>1) % 64
	kind := 0
	for i, n := range []int{10, 10, 11, 11, 11, 11} {
		if r < n {
			kind = i
			break
		}
		r -= n
	}
	g.task(func() bool { g.spinReward(kind); return true })
}
func (g *Game) spinReward(kind int) {
	delay := uint16(10)
	switch kind {
	case 0:
		g.light(39, true)
		g.effect("SPINXB", 100000, 5000)
		delay = 160
	case 1:
		g.arcadeCrazy = true
		g.crazy()
		g.arcadeCrazy = false
		delay = 180
	case 2:
		g.effect("SPIN5M", 5000000, 25000)
		delay = 140
	case 3:
		g.effect("SPIN1M", 1000000, 25000)
		delay = 150
	case 4:
		g.effect("SPIN500K", 500000, 25000)
		delay = 70
	case 5:
		g.effect("SPINNS", 0, 0)
		delay = 45
	}
	if g.special() {
		delay = 0
	}
	if !g.effectAccepted && kind != 5 {
		delay = 10
	}
	if !g.effectAccepted || kind == 0 || kind == 5 {
		g.dropTime = delay
		g.task(func() bool {
			if !g.special() && !g.waitReady("START_DROP_TIMED", g.dropTime) {
				return false
			}
			g.startDrop()
			return true
		})
		return
	}
	minimum := map[int]uint16{1: 140, 2: 110, 3: 120, 4: 45}[kind]
	g.waitForAudioDrop(minimum, delay)
}
