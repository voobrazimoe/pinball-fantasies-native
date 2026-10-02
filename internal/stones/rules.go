package stones

import "fmt"

var stoneBone = []int{20, 21, 22, 23, 26, 27, 28, 29, 30}
var ghostLamps = []int{32, 34, 35, 37, 36, 31, 33, 38}
var ghostEffects = []string{"BATMAN", "TOWERHUNT", "SMILER", "REDDEVIL", "GHOSTHUNT", "MULTIDEMONS", "MUMMYHEAD", "GRIMR"}

func (g *Game) addGhost() {
	if g.GhostHunt {
		g.GhostTotal.Add(number(1000000))
	}
}
func (g *Game) addGrim() {
	if g.Grim {
		g.GrimTotal.Add(number(5000000))
	}
}
func (g *Game) jackAdd()  { g.Jackpot.Add(number(100000)) }
func (g *Game) vaultAdd() { g.VaultValue.Add(number(82150)) }
func (g *Game) towerAdd() { g.TowerValue.Add(number(223470)) }
func (g *Game) wellAdd()  { g.WellValue.Add(number(64190)) }
func (g *Game) all(ns []int) bool {
	for _, n := range ns {
		if !g.Lights[n] {
			return false
		}
	}
	return true
}
func (g *Game) rotateGroups() {
	for _, group := range []struct {
		start    int
		disabled bool
	}{{1, g.keyRollDisabled}, {4, g.ripRollDisabled}} {
		if group.disabled {
			continue
		}
		a := g.Lights[group.start+2]
		g.light(group.start+2, g.Lights[group.start+1])
		g.light(group.start+1, g.Lights[group.start])
		g.light(group.start, a)
	}
}
func (g *Game) touch(i int) {
	n := []int{21, 22, 20, 23, 28, 27, 26, 29, 30}[i]
	if g.ghostInhibit || g.sbDisabled || g.sbGuards[n-20] {
		return
	}
	g.sbGuards[n-20] = true
	g.addGhost()
	if n < 26 {
		g.sound("S_TOUCH2")
		g.award("TOUCHSETA", 27530, 510)
	} else {
		g.sound("S_TOUCH1")
		g.award("TOUCHSETB", 17520, 750)
	}
	g.Lights[n] = true
	if g.all(stoneBone) {
		g.jackAdd()
		g.sbDisabled = true
		g.sbGuards = [11]bool{}
		for _, l := range stoneBone {
			g.endFlash(l)
		}
		if g.ghostFlashing {
			g.effect("EVENT_LIT")
		} else {
			g.effect(fmt.Sprintf("EVENT_LIT%d", g.GhostCounter+1))
			g.ghostFlashing = true
			g.flash(ghostLamps[g.GhostCounter], 32)
			g.flash(16, 18)
		}
		g.wait("ALL_STONE_BONE_2", 2, func() {
			for _, l := range stoneBone {
				g.flash(l, 2)
			}
			g.wait("ALL_STONE_BONE_3", 70, func() {
				for _, l := range stoneBone {
					g.off(l)
				}
				g.sbDisabled = false
			})
		})
		return
	}
	g.flash(n, 2)
	g.wait(fmt.Sprintf("SB_%d", n), 10, func() {
		g.endFlash(n)
		if !g.sbDisabled {
			g.light(n, true)
		}
		g.sbGuards[n-20] = false
	})
}
func (g *Game) key(n int) {
	if g.keyDisabled {
		return
	}
	g.sound("SBYGEL1")
	g.award("KEY", 10060, 1010)
	g.Lights[n] = true
	if g.SkillKey == uint8(n) {
		g.endFlash(n)
		g.SkillScore.Add(number(1000000))
		g.score("SKILLSHOT", g.SkillScore.Uint64())
		g.effect("SKILLSHOT")
		g.vaultAdd()
		g.towerAdd()
		g.jackAdd()
		g.wellAdd()
	}
	if g.SkillKey != 0 {
		for l := 1; l <= 3; l++ {
			if l != n {
				g.off(l)
			}
		}
	}
	g.SkillKey = 0
	if g.all([]int{1, 2, 3}) {
		g.vaultAdd()
		g.jackAdd()
		g.keyDisabled = true
		g.keyRollDisabled = true
		for l := 1; l <= 3; l++ {
			g.flash(l, 2)
		}
		award := 11
		switch g.towerNext {
		case 1:
			award = 12
		case 2:
			award = 14
		case 3:
			award = 13
		default:
			if g.towerNext >= 4 {
				award = 12
			}
		}
		g.towerNext++
		if !g.enabled(award) {
			g.synced(award, 18, g.syncTower)
		}
		if !g.towerOpen {
			g.effect("OPENTOWER")
			g.openTower()
		}
		g.wait("ALL_KEY_2", 70, func() {
			for l := 1; l <= 3; l++ {
				g.off(l)
			}
			g.keyDisabled = false
			g.keyRollDisabled = false
		})
		return
	}
	g.keyRollDisabled = true
	g.flash(n, 2)
	g.wait(fmt.Sprintf("KEY_%d", n), 10, func() {
		if !g.keyDisabled {
			g.endFlash(n)
			g.light(n, true)
			g.keyRollDisabled = false
		}
	})
}
func (g *Game) rip(n int) {
	if g.ripDisabled {
		return
	}
	g.sound("SBYGEL1")
	g.award("RIP", 10070, 1080)
	g.Lights[n] = true
	if g.all([]int{4, 5, 6}) {
		g.ripDisabled = true
		g.ripRollDisabled = true
		for l := 4; l <= 6; l++ {
			g.flash(l, 2)
		}
		g.kickback = true
		g.flash(44, 18)
		g.gate("GATE5", true)
		g.effect("RIPEFF")
		g.wait("ALL_RIP_2", 70, func() {
			for l := 4; l <= 6; l++ {
				g.off(l)
			}
			g.ripDisabled = false
			g.ripRollDisabled = false
		})
		return
	}
	g.ripRollDisabled = true
	g.flash(n, 2)
	g.wait(fmt.Sprintf("RIP_%d", n), 20, func() {
		g.endFlash(n)
		if !g.ripDisabled {
			g.light(n, true)
		}
		g.ripRollDisabled = false
	})
}
func (g *Game) openTower() { g.towerOpen = true; g.flash(7, 20); g.gate("GATE3", true) }
func (g *Game) enabled(n int) bool {
	for _, f := range g.flashes {
		if int(f.lamp) == n {
			return true
		}
	}
	return false
}
func (g *Game) updateTower() {
	for n := 8; n <= 14; n++ {
		if g.enabled(n) {
			return
		}
	}
	g.towerOpen = false
	g.off(7)
	g.gate("GATE3", false)
}
func (g *Game) scream() {
	g.sound("SBYGEL1")
	g.award("SCREAM", 10060, 1050)
	if g.enabled(18) {
		g.off(18)
		g.timers[1] = 1
		n := 0
		if g.Lights[15] {
			n++
		}
		if g.Lights[25] {
			n++
		}
		g.effect([]string{"MILLION5", "MILLION10", "MILLION20"}[n])
		g.light(15, false)
		g.light(25, false)
	}
	if g.timers[4] > 0 && g.ComboStage == 1 {
		g.ComboStage = 2
	} else {
		g.ComboStage = 0
	}
	g.addGrim()
	g.wellAdd()
	g.Screams++
	if g.Screams == 1 {
		g.Screams++
	}
	if g.enabled(17) {
		g.off(17)
		g.wait("THESECONDTIME", 2, g.scream)
	}
	if g.Screams == g.NextJump {
		if g.Screams == 10 {
			g.synced(8, 18, g.syncTower)
			g.effect("GETXBALL")
		} else {
			g.synced(12, 18, g.syncTower)
			g.writeMilesText()
			g.effect("OPENTOWER")
		}
		g.openTower()
		g.NextJump += 10
	} else {
		l := "JUMP_AT"
		if g.Screams > 10 {
			l = "JUMP_AT2"
		}
		g.writeJumpText()
		g.writeMilesText()
		g.effect(l)
	}
}

// STONES.ASM Put_In_Text: write three encoded decimal digits, then
// LOOPEN_BERTIL blanks the entire leading zero run (all zero => "***").
// Only the text selected by the source branch is mutated before EFFECT.
func putInText(buf []byte, at int, value uint16) {
	for i := 2; i >= 0; i-- {
		buf[at+i] = byte(value%10) + '7'
		value /= 10
	}
	for i := at; i < at+3 && buf[i] == '7'; i++ {
		buf[i] = '*'
	}
}
func (g *Game) writeMilesText() {
	putInText(g.Display.MutableText("MILES_TEXT", 7), 4, g.Screams)
}
func (g *Game) writeJumpText() {
	if g.Screams > 10 {
		putInText(g.Display.MutableText("JUMP_AT_TEXT2", 3), 0, g.NextJump)
	} else {
		putInText(g.Display.MutableText("JUMP_AT_TEXT", 3), 0, g.NextJump)
	}
}

func (g *Game) gridLeft() {
	if !g.Physics.Tilted {
		g.addGrim()
		g.vaultAdd()
		g.sound("SBYGEL1")
		g.award("GRID_LEFT", 10030, 1040)
		if !g.enabled(19) {
			g.synced(19, 16, g.syncSulp)
		}
		if !g.enabled(17) {
			g.synced(17, 16, g.syncSulp)
		}
		g.timers[2] = 450
		if g.bonusPointer <= 43 && !g.enabled(24) {
			g.synced(24, 16, g.syncSulp)
			g.timers[3] = 570
		}
		g.ComboStage = 1
		g.timers[4] = 780
	}
	g.gate("GATE4A", true)
}
func (g *Game) loop() {
	g.addGrim()
	g.vaultAdd()
	if g.timers[4] > 0 && g.ComboStage == 2 {
		g.effect("LOOPCOMBO")
	}
	g.ComboStage = 0
	old := g.timers[0]
	g.timers[0] = 300
	if old != 0 {
		g.effect("MILLION")
	} else {
		g.sound("SBYGEL1")
		g.award("LOOP", 10030, 1020)
	}
}
func (g *Game) updateCounters() {
	if g.inhibit || g.Phase == BallLost {
		return
	}
	g.syncSulp = (g.syncSulp + 1) % 32
	g.syncTower = (g.syncTower + 1) % 36
	if g.timers[0] > 0 {
		g.timers[0]--
	}
	if !g.holdMulti && g.timers[1] > 0 {
		if g.timers[1] == 1 {
			g.MultiDemon = false
			if g.enabled(15) {
				g.off(15)
			}
			if g.enabled(25) {
				g.off(25)
			}
		}
		g.timers[1]--
	}
	for _, q := range []struct {
		i  int
		ls []int
	}{{2, []int{19, 17}}, {3, []int{24}}} {
		if g.timers[q.i] > 0 {
			g.timers[q.i]--
			if g.timers[q.i] == 0 {
				for _, n := range q.ls {
					if g.enabled(n) {
						g.off(n)
					}
				}
			} else if g.timers[q.i] == 90 {
				for _, n := range q.ls {
					if g.enabled(n) {
						g.flash(n, 1)
					}
				}
			}
		}
	}
	if g.timers[4] > 0 {
		g.timers[4]--
		if g.timers[4] == 0 {
			g.ComboStage = 0
		}
	}
	if g.timers[5] > 0 {
		g.timers[5]--
		if g.timers[5] == 0 {
			g.TowerHunt = false
			g.TowerStage = 0
			g.playJingle("SJINGLE16")
			g.music.ReturnPosition = 1
		}
	}
}
