package partyland

import "fmt"

var hitLamps = [3]int{16, 18, 24}
var snackLamps = [3]int{23, 17, 20}

func (g *Game) toucher() {
	// TOUCHER has no score award, only a twenty-sync repeat inhibitor.
	if g.touchDisabled {
		return
	}
	g.touchDisabled = true
	g.sound("S_TOUCH2")
	g.waitAt("ENABLETOUCHER", 20, func() { g.touchDisabled = false })
	if !g.Arcade {
		g.Arcade = true
		g.flash(7, 12, 0, false)
		g.flash(55, 12, 0, false)
	}
}
func (g *Game) duck(i int) {
	if !g.Lights[52+i] || g.duckDisabled[i] {
		return
	}
	g.duckDisabled[i] = true
	g.light(52+i, false)
	g.waitAt(fmt.Sprintf("KILL_THE_D%d", i+1), 20, func() { g.duckMask(i, true) })
	g.sound("SBRICKNER")
	g.score("DROPSETA", 7510)
	g.addHappy()
	g.bonus(750)
	if g.Lights[52] || g.Lights[53] || g.Lights[54] {
		g.flash(hitLamps[i], 3, 0, false)
		g.waitAt(fmt.Sprintf("TURNINGONDUCK%d", i+1), 13, func() { g.endFlash(hitLamps[i]); g.duckDisabled[i] = false })
		return
	}
	g.effect("ALLDUCKS", 0, 0)
	// Retain the source's repeated D2DISABLED write; D3 is already set by its hit.
	g.duckDisabled[0] = true
	g.duckDisabled[1] = true
	for _, n := range hitLamps {
		g.endFlash(n)
		g.flash(n, 2, 0, false)
	}
	next := int(g.SnackNext)
	if !g.snacks[next] {
		g.snacks[next] = true
		g.flash(snackLamps[next], 8, g.snackSync, next == 1)
	}
	g.SnackNext = (g.SnackNext + 1) % 3
	g.waitAt("TURNOFFDUCKS", 71, func() {
		for j, n := range hitLamps {
			g.endFlash(n)
			g.light(52+j, true)
			g.duckMask(j, false)
			g.lamp(n, false)
			g.duckDisabled[j] = false
		}
		g.sound("SBRICKUPP")
	})
}
func (g *Game) party(i int) {
	if g.Lights[42+i] {
		return
	}
	// PARTY_P awards before its lamp; all other letters lamp before award.
	if i == 0 {
		g.effect("PARTYSCORE1", 250000, 25000)
		g.light(42, true)
	} else {
		g.light(42+i, true)
		g.effect("PARTYSCORE"+string(rune('1'+i)), 250000, 25000)
	}
	complete := true
	for n := 42; n <= 46; n++ {
		complete = complete && g.Lights[n]
	}
	if complete {
		if g.Happy || g.Mega {
			g.HappyPending = true
		} else {
			g.startHappy()
		}
	}
}
func (g *Game) startHappy() {
	g.effect("HAPPYHOUR", 0, 0)
	for n := 42; n <= 46; n++ {
		g.light(n, false)
	}
	if !g.JackpotNormal && !g.JackpotTimed {
		g.flash(35, 14, g.trainSync, false)
	}
	g.JackpotTimed = true
	g.Audio.ReturnPosition = 43
	g.modeIntro = true
	g.Happy = true
	g.ModeTime = 0
	g.flash(32, 8, 0, false)
}
func (g *Game) startMega() {
	g.effect("CRAZYSCORE", 0, 0)
	if !g.JackpotNormal && !g.JackpotTimed {
		g.flash(35, 14, g.trainSync, false)
	}
	g.JackpotTimed = true
	g.Audio.ReturnPosition = 25
	g.modeIntro = true
	g.Mega = true
	g.ModeTime = 0
	g.flash(37, 8, 0, false)
}
func (g *Game) addHappy() {
	if g.Happy {
		g.HappyTotal.AddNumber(1000000)
	}
}
func (g *Game) addMega() {
	if g.Mega {
		g.MegaTotal.AddNumber(5000000)
	}
}
func (g *Game) modeTick() {
	if (!g.Happy && !g.Mega) || g.ModeTime == 0 {
		return
	}
	dec(&g.ModeTime)
	if g.ModeTime == 2*71 && !g.JackpotNormal && g.JackpotTimed {
		g.endFlash(35)
		g.flash(35, 2, 0, false)
	}
	if g.ModeTime != 0 {
		return
	}
	wasHappy := g.Happy
	if wasHappy {
		g.Happy = false
		g.off(32)
		g.effect("EOHAPPYHOUR", 0, 1000000)
	} else {
		g.Mega = false
		g.off(37)
		g.effect("EOMEGALAUGH", 0, 5000000)
	}
	g.Audio.ReturnPosition = 1
	g.off(35)
	g.JackpotTimed = false
	if g.JackpotNormal {
		g.flash(35, 14, g.trainSync, false)
	}
	if wasHappy && g.MegaPending {
		g.MegaPending = false
		g.waitAt("MEGA_START", 400, g.startMega)
	}
	if !wasHappy && g.HappyPending {
		g.HappyPending = false
		g.waitAt("HAPPY_START", 400, g.startHappy)
	}
}
func (g *Game) crazy() {
	g.jackAdd()
	label := "ADVANCE3"
	if g.arcadeCrazy {
		label = "ADVANCE3B"
	}
	g.effect(label, 0, 0)
	n := [5]int{41, 38, 34, 31, 28}[g.CrazyNext]
	g.light(n, true)
	g.CrazyNext++
	if g.CrazyNext == 5 {
		g.CrazyNext = 0
		for _, n := range []int{28, 31, 34, 38, 41} {
			g.light(n, false)
		}
		if g.Happy || g.Mega {
			g.MegaPending = true
		} else {
			g.startMega()
		}
	}
}
func (g *Game) loop() {
	if g.InhibitLoop {
		return
	}
	g.jackAdd()
	g.addMega()
	if g.LoopTime != 0 {
		g.party(3)
	}
	g.LoopTime = 600
	if !g.Lights[6] {
		g.effect("THELOOP1", 100000, 10000)
		g.flash(6, 2, 0, false)
		g.waitAt("MAD_M_TASK", 14, func() { g.endFlash(6); g.Lights[6] = true })
		return
	}
	if !g.Lights[8] {
		g.effect("THELOOP2", 250000, 25000)
		g.flash(8, 2, 0, false)
		g.waitAt("MAD_A_TASK", 14, func() { g.endFlash(8); g.Lights[8] = true })
		return
	}
	g.Lights[6] = false
	g.Lights[8] = false
	for _, n := range []int{6, 8, 9} {
		g.flash(n, 2, 0, false)
	}
	g.InhibitLoop = true
	g.waitAt("TURN_OFF_MAD", 120, func() {
		for _, n := range []int{6, 8, 9} {
			g.endFlash(n)
		}
		g.InhibitLoop = false
	})
	g.effect("THELOOP3", 500000, 50000)
	g.crazy()
}
func (g *Game) reverse() {
	if g.InhibitReverseTime != 0 {
		g.InhibitReverseTime = 0
		return
	}
	if g.InhibitReverse {
		return
	}
	g.jackAdd()
	g.addMega()
	if g.LoopTime != 0 {
		g.party(3)
	}
	g.LoopTime = 600
	g.ReverseTime = 300
	g.Balloon++
	switch g.Balloon {
	case 1:
		g.endFlash(26)
		g.lamp(26, true)
		g.flash(25, 9, 0, false)
		g.effect("RSCORE1", 250000, 10000)
	case 2:
		g.endFlash(25)
		g.lamp(25, true)
		g.flash(19, 9, 0, false)
		g.effect("RSCORE2", 500000, 10000)
	default:
		g.Balloon = 0
		g.endFlash(19)
		for _, n := range []int{19, 25, 26} {
			g.flash(n, 2, 0, false)
		}
		g.effect("RSCORE3", 750000, 10000)
		g.InhibitReverse = true
		g.waitAt("TURN_OFF_BALOONS", 120, func() {
			for _, n := range []int{19, 25, 26} {
				g.endFlash(n)
			}
			g.flash(26, 9, 0, false)
			g.InhibitReverse = false
		})
	}
	if g.MB {
		g.MB = false
		g.off(29)
		for i, n := range []int{47, 49, 50, 48} {
			if !g.Lights[n] {
				g.light(n, true)
				g.effect("MULTIBONUS", 10000, 5000)
				g.Multiplier = uint8(2 * (i + 1))
				break
			}
		}
	}
	if g.HB {
		g.HB = false
		g.off(33)
		g.effect("HOLDBONUS", 0, 0)
		g.HoldBonus = true
	}
	if g.DB {
		g.DB = false
		g.off(36)
		g.effect("DOUBLEBONUS", 0, 0)
		g.Bonus.Add(g.Bonus)
	}
}
func (g *Game) skyride() {
	g.addMega()
	g.gate(false)
	if g.LoopTime != 0 {
		g.party(3)
	}
	g.LoopTime = 600
	g.Skyride++
	switch g.Skyride {
	case 1:
		g.lamp(22, true)
		g.effect("RIDE1", 100000, 10000)
	case 2:
		g.lamp(13, true)
		g.effect("RIDE2", 250000, 25000)
	default:
		g.effect("RIDE3", 500000, 50000)
		for _, n := range []int{22, 13, 15} {
			g.flash(n, 2, 0, false)
		}
		g.waitAt("KILL_FLASHING_ROCKETS", 120, func() {
			for _, n := range []int{22, 13, 15} {
				g.endFlash(n)
			}
		})
		g.Skyride = 0
		if !g.MB && !g.Lights[48] {
			g.MB = true
			g.flash(29, 12, g.arrowSync, false)
			g.effect("GETBONUS", 0, 0)
		}
	}
}
func (g *Game) pukeLetter(n int) {
	g.sound("SBYGEL1")
	if g.Lights[n] {
		return
	}
	g.Lights[n] = true
	g.score("BCD20070", 20070)
	g.bonus(1000)
	g.jackAdd()
	if !(g.Lights[1] && g.Lights[2] && g.Lights[4] && g.Lights[5]) {
		g.PukeForbidden = true
		g.flash(n, 2, 0, false)
		g.waitAt(fmt.Sprintf("PUKEP%d", n), 13, func() { g.endFlash(n); g.PukeForbidden = false })
		return
	}
	for _, n := range []int{1, 5, 2, 4} {
		g.flash(n, 2, g.pukeSync, n == 2 || n == 4)
	}
	g.waitAt("TURNOFFPUKE", 100, func() {
		for _, n := range []int{1, 2, 4, 5} {
			g.off(n)
		}
	})
	g.Puke++
	switch g.Puke {
	case 1:
		if !g.FiveMillion {
			g.flash(27, 14, g.trainSync, false)
			g.FiveMillion = true
		}
	case 2:
		if !g.BallFeature {
			g.flash(30, 14, g.trainSync, true)
			g.BallFeature = true
		}
	case 3:
		if !g.JackpotNormal {
			if !g.JackpotTimed {
				g.flash(35, 14, g.trainSync, false)
			}
			g.JackpotNormal = true
		}
	}
	g.party(4)
}
func (g *Game) killFiveX() { g.off(11); g.FiveX = false }
func (g *Game) cyclone() {
	defer func() { g.inhibitEffect = false }()
	g.addMega()
	if g.SkillTime != 0 {
		g.jackAdd()
		g.jackAdd()
		g.effect("SKILLCYCLONE", 0, 0)
		g.SkillCyclone.AddNumber(1000000)
		g.score("SKILL_SCORE2", g.SkillCyclone.Uint64())
		g.effect("SKILLCYCLONE", 0, 0)
		g.inhibitEffect = true
		g.party(2)
		g.killFiveX()
	} else if g.ReverseTime != 0 {
		g.party(2)
		g.killFiveX()
	}
	if g.FiveX {
		g.Cyclones += 5
		g.effect("CYCLCOUNT2", 2500000, 500000)
	} else {
		g.Cyclones++
		if g.Cyclones == 1 {
			g.Cyclones++
		}
		g.effect("CYCLCOUNT", 100000, 25000)
	}
	g.killFiveX()
}
