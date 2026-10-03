package stones

import (
	"encoding/binary"
	"fmt"
)

type areaRecord struct {
	Rect    [4]int16
	Handler string
}

// decodeAreas reads linked rectangle records as data. Handler words select
// native behavior; they are never called as machine-code addresses. Each game
// owns its decoded lists. The terminating zero word is part of the contract.
func decodeAreas(data []byte, refs map[string][2]int, handlers map[uint16]string) map[string][]areaRecord {
	out := make(map[string][]areaRecord, len(refs))
	for name, ref := range refs {
		at, count := ref[0], ref[1]
		if at < 0 || at > len(data) || count < 0 || len(data)-at < 2 || count > (len(data)-at-2)/10 {
			panic("invalid PRG area extent")
		}
		if binary.LittleEndian.Uint16(data[at+count*10:]) != 0 {
			panic("invalid PRG area terminator")
		}
		rows := make([]areaRecord, count)
		for i := range rows {
			b := data[at+i*10 : at+i*10+10]
			for j := range rows[i].Rect {
				rows[i].Rect[j] = int16(binary.LittleEndian.Uint16(b[j*2:]))
			}
			handler, ok := handlers[binary.LittleEndian.Uint16(b[8:])]
			if !ok {
				panic("unregistered PRG area handler")
			}
			rows[i].Handler = handler
		}
		out[name] = rows
	}
	return out
}

func (g *Game) checkAreas() {
	b := &g.Physics.Ball
	name := "AREALISTA_L"
	if b.High {
		name = "AREALISTA_U"
	}
	if g.Physics.Tilted {
		name += "_T"
	}
	x, y := uint16(b.PixelX+8), uint16(b.PixelY+8+g.Physics.ScreenOffset)
	for _, r := range g.areas[name] {
		if x >= uint16(r.Rect[0]) && x <= uint16(r.Rect[2]) && y >= uint16(r.Rect[1]) && y <= uint16(r.Rect[3]) {
			if r.Handler != g.lastCheck {
				g.lastCheck = r.Handler
				g.area(r.Handler)
			}
			return
		}
	}
	g.lastCheck = ""
}
func (g *Game) area(l string) {
	switch l {
	case "CLOSE1":
		g.gate("GATE2", false)
		g.holdMulti = false
		if g.enabled(19) {
			g.SulpScore.Add(number(1000000))
			g.score("SULP", g.SulpScore.Uint64())
			g.effect("MILLIONPLUS")
			g.off(19)
		}
		g.award(l, 10000, 1000)
	case "PARTYOFFAREA":
		g.inChute = false
		g.Session.SelectionOpen = false // CLOSE1 writes ADDPLAYERS=FALSE.
		g.partyFlash = false
		g.Physics.Ball.High = false
		g.beginMatrix("PARTY_OFFTS")
		g.Cue("S_MAIN")
		g.music.ReturnPosition = 1
	case "BYGEL16":
		g.gate("GATE2", true)
		g.addGrim()
	case "OPENBUMPERS":
		g.gate("GATE2", true)
	case "BYGEL5":
		g.key(1)
	case "BYGEL6":
		g.key(2)
	case "BYGEL7":
		g.key(3)
	case "BYGEL12":
		g.rip(4)
	case "BYGEL13":
		g.rip(5)
	case "BYGEL14":
		g.rip(6)
	case "GROPA", "GROPA_T":
		g.captureTower()
	case "GROPB", "GROPB_T":
		g.captureWell()
	case "GROPC":
		g.captureVault()
	case "BYGEL3", "BYGEL4":
		g.sound("SBYGEL2")
		g.award(l, 10070, 1080)
	case "BYGEL1":
		g.sound("SBYGEL1")
		g.score(l, 50010)
	case "BYGEL2":
		g.sound("SBYGEL1")
		g.score(l, 50030)
	case "BYGEL28":
		g.Physics.SpringValid = true
	case "BYGELSI":
		g.Physics.SpringValid = false
	case "BYGEL8":
		g.loop()
	case "BYGEL8B":
		g.timers[0] = 300
	case "BYGEL11":
		g.scream()
	case "BYGEL9":
		g.gridLeft()
	case "BYGEL10":
		if !g.Physics.Tilted {
			g.addGrim()
			g.sound("SBYGEL1")
			g.award(l, 10020, 1010)
		}
		g.gate("GATE4A", false)
	case "OPEN6":
		g.gate("GATE4B", true)
	case "CLOSE6":
		g.gate("GATE4B", false)
	case "OPEN7":
		g.gate("GATE4C", true)
	case "CLOSE7":
		g.gate("GATE4C", false)
	case "BYGEL18":
		g.highVault = true
		g.vaultAdd()
	case "BYGEL15", "BYGEL19", "CLOSE4":
	default:
		panic("unknown STONES area " + l)
	}
}
func (g *Game) hold(name string, x, y int16, high bool) bool {
	if g.captured != "" {
		return false
	}
	g.captured = name
	g.Physics.SetBall(x, y, 0, 0, high)
	g.Physics.Ball.Hold = true
	g.emit("Capture", name, 0)
	return true
}
func (g *Game) ejectTask(name string) {
	if g.captured != name {
		return
	}
	g.wait(name+"ENDT", 10, func() { g.eject(name) })
}
func (g *Game) eject(name string) {
	if g.captured != name {
		return
	}
	g.sound("SGROP")
	g.Physics.Ball.Hold = false
	switch name {
	case "TOWER":
		g.Physics.SetBall(141, 143, 0, -4000, true)
		g.inhibit = false
		if g.wasSpecial && !g.Physics.Tilted {
			label := "BACK_2_OFFROADTS"
			g.GhostHunt = true
			if g.grimBackup {
				label = "BACK_2_TURBOTS"
				g.Grim = true
			}
			g.beginMatrix(label)
		}
	case "WELL":
		g.Physics.SetBall(275, 245, -800, 2000, false)
	case "VAULT":
		g.gate("GATE5", true)
		g.Physics.SetBall(2, 532, 0, -2880, false)
		g.wait("VAULTENDTT", 30, func() {
			if !g.kickback {
				g.gate("GATE5", false)
			}
			g.highVault = false
		})
	}
	g.captured = ""
	g.emit("Eject", name, 0)
}
func (g *Game) captureTower() {
	if !g.hold("TOWER", 141, 143, true) {
		return
	}
	if g.Physics.Tilted {
		g.ejectTask("TOWER")
		return
	}
	g.addGrim()
	g.jackAdd()
	g.inhibit = true
	g.wasSpecial = g.Special
	g.grimBackup = g.Grim
	g.towerHuntOrig = g.TowerHunt // GROPA captures TOWERHUNTMODEORIG.
	g.gate("GATE3", false)
	g.gate("GATE2", true)
	g.towerOpen = false
	g.off(7)
	accepted := false
	if g.TowerHunt && g.TowerStage > 0 {
		accepted = g.effect(fmt.Sprintf("TOWERHUNT%d", g.TowerStage))
		if g.music.ReturnPosition != 50 {
			g.music.ReturnPosition++
		}
		g.TowerStage++
		if g.TowerStage == 4 {
			g.TowerStage = 0
		}
		g.openTower()
		g.inhEff = true
	}
	if g.enabled(10) {
		g.off(10)
		accepted = g.effectCore("SUPERJACK", true) || accepted
		g.inhEff = true
	}
	if g.enabled(9) {
		g.off(9)
		accepted = g.effectCore("JACKPOT", true) || accepted
		g.score("JACKVALUE", g.Jackpot.Uint64())
		g.Jackpot = number(10000000)
		g.synced(10, 18, g.syncTower)
		g.openTower()
		g.wait("SJOFF", 780, func() {
			if g.enabled(10) {
				g.off(10)
				g.updateTower()
			}
		})
		g.inhEff = true
	}
	for _, q := range []struct {
		n      int
		effect string
	}{{8, "EXTRABALL"}, {14, "DOUBLEBONUS"}, {13, "HOLDBONUS"}, {12, "TMILLION5"}, {11, "TMILLION"}} {
		if g.enabled(q.n) {
			g.off(q.n)
			accepted = g.effect(q.effect) || accepted
			switch q.n {
			case 8:
				g.ExtraBalls++
			case 14:
				g.Bonus.Add(g.Bonus)
			case 13:
				g.holdBonus = true
			}
			g.inhEff = true
			if q.n == 11 {
				g.inhEff = false
				if !accepted {
					g.ejectTask("TOWER")
				}
				return
			}
		}
	}
	accepted = g.effect("SCORETOWER") || accepted
	g.score("TOWERVALUE", g.TowerValue.Uint64())
	g.TowerValue = number(1000000)
	if g.TowerHunt {
		g.openTower()
	}
	g.inhEff = false
	if !accepted {
		g.ejectTask("TOWER")
	}
}
func (g *Game) teleport() {
	g.holdMulti = true
	g.music.Priority = 0
	g.effect("SHOOTTHEBALL")
	g.inhEff = true
	g.Physics.SetBall(302, 535, 10, 0, false)
	g.Physics.Ball.Hold = false
	g.captured = ""
	g.partyFlash = true
	g.inChute = true
	g.music.ReturnPosition = 0
}
func (g *Game) captureWell() {
	if !g.hold("WELL", 275, 245, false) {
		return
	}
	if g.Physics.Tilted || g.Lights[25] {
		g.ejectTask("WELL")
		return
	}
	g.addGrim()
	g.jackAdd()
	g.score("WELLVALUE", g.WellValue.Uint64())
	accepted := false
	if g.enabled(25) {
		g.teleport()
		g.endFlash(25)
		g.light(25, true)
		accepted = true
	}
	accepted = g.effect("SCOREWELL") || accepted
	if g.enabled(24) {
		g.off(24)
		if g.bonusPointer <= 43 {
			if g.Multiplier == 1 {
				g.Multiplier = 2
			} else {
				g.Multiplier += 2
			}
			g.light(int(g.bonusPointer), true)
			accepted = g.effect(fmt.Sprintf("M%d", g.Multiplier)) || accepted
			g.bonusPointer++
			text := append([]byte(nil), g.Display.Content.Texts["BONUS_TEXT"]...)
			text[11] = g.Multiplier + 7
			g.Display.Content.Texts["BONUS_TEXT"] = text
		}
	}
	g.inhEff = false
	if !accepted {
		g.ejectTask("WELL")
	}
}
func (g *Game) captureVault() {
	if !g.hold("VAULT", 2, 532, false) {
		return
	}
	if !g.highVault {
		g.off(44)
		g.kickback = false
	}
	if g.Physics.Tilted || g.Lights[15] {
		g.ejectTask("VAULT")
		return
	}
	g.jackAdd()
	accepted := false
	if g.enabled(15) {
		g.teleport()
		g.endFlash(15)
		g.light(15, true)
		accepted = true
	}
	if g.ghostFlashing {
		g.ghostFlashing = false
		g.off(16)
		g.endFlash(ghostLamps[g.GhostCounter])
		g.light(ghostLamps[g.GhostCounter], true)
		idx := g.GhostCounter
		if g.Special && (idx == 4 || idx == 7) {
			g.task(func() bool {
				if g.Special {
					return false
				}
				if g.Phase == BallLost {
					return true
				}
				g.fixTeleport = true
				g.effect(ghostEffects[idx])
				return true
			})
		} else {
			accepted = g.effect(ghostEffects[idx]) || accepted
		}
		switch idx {
		case 1:
			g.TowerHunt = true
			g.TowerStage = 1
			g.timers[5] = 2400
			g.openTower()
			g.music.ReturnPosition = 46
		case 2:
			g.synced(8, 18, g.syncTower)
			g.openTower()
		case 4, 7:
			g.synced(9, 18, g.syncTower)
			g.openTower()
		case 5:
			g.MultiDemon = true
			g.timers[1] = 2100
			g.holdMulti = false
			g.flash(15, 18)
			g.flash(25, 18)
			g.flash(18, 18)
		}
		g.GhostCounter++
		if g.GhostCounter == 8 {
			g.GhostCounter = 0
			g.ghostInhibit = true
			for n := 31; n <= 38; n++ {
				g.flash(n, 2)
			}
			g.wait("GHOSTOFF", 240, func() {
				for n := 31; n <= 38; n++ {
					g.off(n)
				}
				g.ghostInhibit = false
			})
		}
	}
	g.addGrim()
	g.score("VAULTVALUE", g.VaultValue.Uint64())
	g.inhEff = accepted
	accepted = g.effect("SCOREVAULT") || accepted
	g.inhEff = false
	if !accepted {
		g.ejectTask("VAULT")
	}
}
