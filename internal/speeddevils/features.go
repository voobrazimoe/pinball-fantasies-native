package speeddevils

import (
	"fmt"
	"pinballfantasies/internal/tablelogic"
)

func (g *Game) touch(i int) {
	if g.touchBusy[i] {
		return
	}
	g.touchBusy[i] = true
	g.Lights[16+i] = true
	g.addOffRoad()
	award, bonus := uint64(7510), uint64(550)
	if i >= 3 {
		award, bonus = 7520, 570
	}
	g.score("BURNIN", award)
	g.Bonus.Add(number(bonus))
	g.sound("S_TOUCH1")
	base, gear := 16, 24
	if i >= 3 {
		base, gear = 19, 25
	}
	if g.Lights[base] && g.Lights[base+1] && g.Lights[base+2] {
		g.jackAdd()
		for n := base; n < base+3; n++ {
			g.touchBusy[n-16] = true
			g.flash(n, 1)
		}
		g.Lights[gear] = true
		g.checkGear()
		g.flash(gear, 1)
		g.wait(fmt.Sprintf("END_%d", gear), 40, func() {
			g.endFlash(gear)
			g.lamp(gear, true)
			for n := base; n < base+3; n++ {
				g.off(n)
				g.touchBusy[n-16] = false
			}
		})
	} else {
		g.flash(16+i, 1)
		g.wait(fmt.Sprintf("END_TOUCH_%d", i), 10, func() {
			g.endFlash(16 + i)
			if g.touchBusy[i] {
				g.lamp(16+i, true)
				g.touchBusy[i] = false
			}
		})
	}
}
func (g *Game) checkGear() bool {
	for n := 22; n <= 25; n++ {
		if !g.Lights[n] {
			return false
		}
	}
	if g.PositionAvailable < 10 {
		n := int(32 + g.PositionAvailable)
		g.syncFlash(n, false)
		g.syncFlash(n+1, true)
		g.PositionAvailable += 2
	} else {
		g.effect("NUMBERONE")
	}
	if g.Gear != 5 {
		g.light(int(26+g.Gear), true)
		g.Gear++
	} else {
		g.gearDown = 30
		for n := 26; n <= 31; n++ {
			g.flash(n, 1)
		}
	}
	for n := 22; n <= 25; n++ {
		g.Lights[n] = false
		g.flash(n, 2)
	}
	g.wait("STOP_GEAR", 45, func() {
		for n := 22; n <= 25; n++ {
			g.off(n)
		}
	})
	g.effect("GEARCHANGE")
	return true
}
func (g *Game) resetGear() {
	g.Gear = 0
	for n := 26; n <= 31; n++ {
		g.off(n)
	}
	if !g.Lights[2] {
		g.Lights[2] = true
		g.syncFlash(2, false)
	}
}
func (g *Game) pit(i int) {
	if g.pitDown[i] != 0 {
		return
	}
	g.Lights[6+i] = true
	g.addOffRoad()
	g.sound("S_TOUCH1")
	g.effect("BYGELSETC")
	if g.Lights[6] && g.Lights[7] && g.Lights[8] && g.pitAll == 0 {
		g.effect("MBLIT")
		if g.MBCollected+g.MBStock >= 8 {
			g.effect("MILLION")
		} else {
			g.MBStock++
			if g.MBStock == 1 {
				g.Lights[4] = true
				g.flash(4, 10)
			}
		}
		for n := 6; n <= 8; n++ {
			g.flash(n, 2)
		}
		g.pitAll = 40
	} else {
		g.pitDown[i] = 20
		g.flash(6+i, 1)
	}
}
func (g *Game) offroadLane() {
	if g.Turbo {
		g.effect("TURBOEFFECT")
		g.addTurbo()
	}
	g.effect("BYGELSETE")
	g.part(11, 50, "PART1")
	g.part(12, 51, "PART4")
	if g.Lights[4] {
		g.Multiplier++
		g.MBCollected++
		g.MBStock--
		if g.MBStock == 0 {
			g.off(4)
		}
		n := 42
		for n < 49 && g.Lights[n] {
			n++
		}
		g.effect(fmt.Sprintf("M%d", n-40))
		g.light(n, true)
	}
	g.Lights[23] = true
	g.jackAdd()
	if !g.checkGear() {
		g.flash(23, 1)
		g.wait("STOP_E", 10, func() { g.endFlash(23); g.lamp(23, true) })
	}
}
func (g *Game) part(source, dest int, label string) {
	if !g.Lights[source] {
		return
	}
	g.off(source)
	g.effect(label)
	g.Lights[dest] = true
	g.syncFlash(dest, dest%2 == 0)
	g.partTest()
}
func (g *Game) partTest() {
	for n := 50; n <= 54; n++ {
		if !g.Lights[n] {
			return
		}
	}
	for n := 50; n <= 54; n++ {
		g.off(n)
		g.flash(n, 1)
	}
	g.CarPart = 0
	g.wait("ENDFLASHA_CARPARTS", 120, func() {
		for n := 50; n <= 54; n++ {
			g.endFlash(n)
		}
	})
}
func (g *Game) jumpLane() {
	if g.lastArea != "BYGEL12" {
		if g.lastArea == "BYGEL17" {
			g.part(14, 52, "PART3")
			g.part(10, 54, "PART5")
			if g.Lights[1] {
				g.effect("XBALL")
				g.off(1)
				g.light(55, true)
			}
		}
		return
	}
	if g.Turbo {
		g.effect("TURBOEFFECT")
		g.addTurbo()
	}
	g.jackAdd()
	if g.Lights[15] {
		g.score("JACKPOT", g.Jackpot.Uint64())
		g.Jackpot = number(5_000_000)
		if g.Turbo {
			g.music.Priority = 0
			g.inhibitCountdown = true
			g.Special = false
			g.effect("JACKPOT")
			g.Special = true
		} else {
			g.effect("JACKPOT")
		}
		g.off(15)
		g.Lights[3] = true
		g.syncFlash(3, false)
		g.wait("TURNOFF_SUPER", 1200, func() { g.off(3) })
	}
	if g.Lights[5] {
		g.effect("JUMP")
		g.off(5)
	}
	g.part(13, 53, "PART2")
	g.jump = true
	g.speedometer()
	g.Lights[22] = true
	g.jackAdd()
	if !g.checkGear() {
		g.flash(22, 1)
		g.wait("STOP_G", 10, func() { g.endFlash(22); g.lamp(22, true) })
	}
}
func (g *Game) speedometer() {
	if !g.loopH || !g.loopL || !g.jump {
		return
	}
	g.effect("SPEEDO")
	g.loopH, g.loopL, g.jump = false, false, false
	n := int(g.Speed) + 56
	if n%2 != 0 && g.CarPart < 5 {
		g.CarPart++
		lamp := []int{11, 13, 14, 12, 10}[g.CarPart-1]
		g.Lights[lamp] = true
		g.syncFlash(lamp, false)
	}
	g.Speed++
	if n >= 67 {
		g.lamp(67, true)
	} else {
		g.light(n, true)
	}
}
func (g *Game) loop(high bool) {
	if high {
		g.loopHigh = 390
		if g.loopLow != 0 {
			g.loopLow = 0
			g.twoLoops()
		}
	} else {
		g.loopLow = 390
		if g.loopHigh != 0 {
			g.loopHigh = 0
			g.twoLoops()
		}
	}
	g.jackAdd()
	s := g.Speed
	if s > 11 {
		s = 11
	}
	g.effect(fmt.Sprintf("SSCORE%d", s+1))
	if g.Turbo {
		g.effect("TURBOEFFECT")
		g.addTurbo()
	}
	g.Miles++
	if g.Miles == 1 {
		g.Miles++
	}
	if g.Miles <= 20 {
		if g.Miles > 10 {
			g.effect("XBALL_AT_20")
			if g.Miles == 20 {
				g.Lights[1] = true
				g.flash(1, 15)
				g.effect("LITXBALL")
			}
		}
		if g.Miles == 10 {
			g.startOffRoad()
		} else if g.Miles < 10 {
			g.effect("OFFROAD_AT_10")
		}
	}
	if g.Miles == g.NextJump {
		g.NextJump += 20
		if !g.Lights[5] {
			g.Lights[5] = true
			g.flash(5, 15)
			g.effect("LGT_JUMP")
		}
	} else if uint16(g.NextJump-g.Miles) < 10 {
		g.writeJumpText()
		g.effect("JUMP_AT")
	}
	if uint16(g.NextOffRoad-g.Miles) <= 10 {
		if g.Miles == g.NextOffRoad {
			g.NextOffRoad += 20
			g.startOffRoad()
		}
		g.writeOffRoadText()
		g.effect("OFFROAD_AT")
	}
	// NotStandardSeries: the miles count goes into the scrolltext on every
	// award, before any queued MilesTS program prints it.
	g.writeMilesText()
}
func (g *Game) twoLoops() {
	if g.PositionAvailable > g.Position {
		g.light(int(32+g.Position), true)
		g.endFlash(int(32 + g.Position))
		g.Position++
		g.effect("OVERTAKE")
		if g.Position == 10 {
			g.Lights[9] = true
			g.syncFlash(9, true)
			g.Lights[15] = true
			g.syncFlash(15, false)
			g.jackDown = 1200
			g.effect("GOALLIT")
		}
	}
	g.effect("MILLION")
	g.loopH, g.loopL = true, true
}
func (g *Game) startOffRoad() {
	if g.Special {
		g.task(func() bool {
			if g.Phase == BallLost {
				return true
			}
			if g.Special {
				return false
			}
			g.startOffRoad()
			return true
		})
		return
	}
	g.music.Priority = 0
	g.playJingle("SJINGLE22")
	g.Special, g.OffRoad = true, true
	g.beginMatrix("OFFROADTS")
}
func (g *Game) startTurbo() {
	for n := 32; n <= 41; n++ {
		g.off(n)
	}
	g.Position, g.PositionAvailable = 0, 0
	if g.Special {
		g.task(func() bool {
			if g.Phase == BallLost {
				return true
			}
			if g.Special {
				return false
			}
			g.startTurbo()
			return true
		})
		return
	}
	g.effect("GETGOAL")
	g.Special, g.Turbo = true, true
}
func (g *Game) pitstop() {
	if g.Lights[3] {
		g.off(3)
		if g.Lights[9] {
			g.effect("SUPERJACK2")
			g.capture(150)
			return
		}
		g.inhibitCountdown = true
		g.music.Priority = 0
		old := g.Special
		g.Special = false
		g.effect("SUPERJACK")
		g.Special = old
	}
	if g.Lights[9] {
		g.off(9)
		g.startTurbo()
	}
	if g.Lights[2] {
		g.effect("HOLDBONUS")
		g.off(2)
		g.HoldBonus = true
	}
	g.capture(20)
}
func (g *Game) capture(wait uint16) {
	if g.snackDisabled {
		return
	}
	// SDEV tests SNACK_DISABLED but never sets it on capture. LASTCHECK
	// suppresses repeated area entry; preserve that original behavior.
	g.task(func() bool {
		g.Physics.Ball.Hold = true
		g.Physics.SetBall(256, 41, 0, 0, false)
		if !tableWait(g, "SNACK_HOLE_TASK", wait) {
			return false
		}
		g.task(func() bool {
			g.Physics.Ball.Hold = true
			g.Physics.SetBall(256, 41, 0, 0, false)
			if !g.Special && !tableWait(g, "SNACK_HOLE_TASK1B", 60) {
				return false
			}
			g.sound("SGROP")
			g.Physics.Ball.Hold = false
			g.Physics.SetBall(256, 41, -2100, 800, false)
			g.wait("TURN_IT_ON_AGAIN", 60, func() { g.snackDisabled = false })
			return true
		})
		return true
	})
}

func tableWait(g *Game, site string, n uint16) bool { return tablelogic.Wait(g.waits, site, n) }
