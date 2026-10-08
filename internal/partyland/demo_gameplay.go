//go:build dmoimpl1 || demodev

package partyland

import (
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"

	"pinballfantasies/internal/tablelogic"
)

// Admission is complete before LASTCHECK, switch events or callback effects.
func (d *demoConnected) gameplayArea() error {
	if d.failure != nil {
		return d.failure
	}
	if !d.gameplayOperands {
		return d.rejectEdge("checkAreas", "BYGEL operands", "no pinned gameplay admission")
	}
	g := d.game
	if g.Playback != nil {
		return d.rejectEdge("checkAreas", "host playback", "silent candidate only")
	}
	label := g.areaConsumer()
	if label == "" {
		g.lastCheck = ""
		return nil
	}
	if label == g.lastCheck {
		return nil
	}
	switch label {
	case "BYGEL12", "BYGEL28", "CLOSE1":
	case "BYGEL9":
		if g.lastArea == "BYGEL11" {
			return d.rejectEdge("BYGEL9", "loop award chain", "unadmitted nested JACKADD/ADDMEGALAUGH/DO_PARTY_T")
		}
	case "BYGEL11":
		if g.lastArea == "BYGEL9" && g.InhibitReverseTime == 0 && !g.InhibitReverse {
			return d.rejectEdge("BYGEL11", "reverse award chain", "unadmitted nested mode/award consumers")
		}
	case "BYGEL3", "BYGEL4":
	case "BYGEL1", "BYGEL2":
		if g.Lights[39] {
			return d.rejectEdge(label, "EXTRABALL2", "unadmitted nested effect/matrix; no lamps or score changed")
		}
	default:
		return d.rejectEdge("checkAreas", label, "area consumer outside bounded BYGEL family")
	}
	if (label == "BYGEL1" || label == "BYGEL2" || label == "BYGEL3" || label == "BYGEL4") && d.fjantText {
		for _, c := range demoPlayerProgram() {
			if !d.command(c) {
				return d.rejectEdge(label, "SHOWPLAYERSTS", "unadmitted nested score panel")
			}
		}
	}
	g.lastCheck = label
	g.emit("Switch", label, 0)
	switch label {
	case "BYGEL12":
		g.SkillTime = 300
		g.Physics.SpringValid = false
		g.InhibitReverseTime = 120
	case "BYGEL28":
		g.Physics.SpringValid = true
	case "BYGEL9":
	case "BYGEL11":
		if g.lastArea == "BYGEL9" {
			g.InhibitReverseTime = 0
		}
	case "BYGEL1", "BYGEL2":
		g.sound("SBYGEL1")
		if err := d.bygelScore("BCD50030", 50030, 0); err != nil {
			return err
		}
	case "BYGEL3", "BYGEL4":

		if err := d.bygelScore("BYGELSETB", 10040, 1000); err != nil {
			return err
		}
		g.effectAccepted = false
		g.effectEnded = true
		g.sound("SBYGEL2")
	case "CLOSE1":
		if g.lastArea == "BYGEL12" {
			d.addPlayers = false
			g.Session.SelectionOpen = false
			a := g.musicClock()
			accepted := a.Play(tablelogic.JingleSpec{Position: 1, Priority: 1}, 62, timing.Cues)
			g.storeMusicClock(a)
			if accepted {
				g.emit("Music", "S_MAIN", 1)
			} else {
				g.emit("AudioRejected", "S_MAIN", 1)
			}
			g.Audio.ReturnPosition = 1
			if err := d.install(demoPartyOffProgram()); err != nil {
				return err
			}
			g.inChute = false
			g.partyFlash = false
			d.visaKeys = false
		}
	}
	g.lastArea = g.lastCheck
	d.trace = append(d.trace, "area:"+label)
	return nil
}

func demoPartyOffProgram() []presentation.Command {
	return []presentation.Command{{Op: "_CLEAR4"}, {Op: "_PARTYOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}}, {Op: "_DEMO_PARTY_FLASHOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}}, {Op: "0"}}
}

func (d *demoConnected) keyboardEntry(in physics.Inputs) error {
	if d.failure != nil {
		return d.failure
	}
	if d.game.Physics.FastBall {
		return d.rejectEdge("controls", "fastball", "fresh Legacy convention required")
	}
	if in.Tilt || in.MouseY != 0 || in.MouseFire || in.TouchSet {
		return d.rejectEdge("controls", "alternate controls", "only fixed keyboard controls admitted")
	}
	g := d.game
	g.Events = g.Events[:0]
	g.Physics.Events = g.Physics.Events[:0]
	g.Tick++
	g.clock += 1030
	if g.Physics.AllowFlip && g.Phase == Playing {
		if in.Left && !g.previousInput.Left {
			g.sound("SFLIPPUPP")
		}
		if in.Right && !g.previousInput.Right {
			g.sound("SFLIPPUPP")
		}
	}
	if g.plungerSound {
		g.emit("Sound", "SFJADER", uint64(g.plungerVolume))
		g.plungerSound = false
	}
	d.controls = func() error {
		gameplay.Spring(&g.Physics.SpringPosition, g.Physics.SpringValid, in, func(charge uint8) {
			g.Physics.Release(charge, uint8(g.clock))
			g.plungerSound = charge > 0
			g.plungerVolume = charge * 2
		})
		return nil
	}
	return nil
}
func (d *demoConnected) keyboardLamps(in physics.Inputs) {
	if d.failure != nil {
		return
	}
	g := d.game
	if g.Physics.AllowFlip && !g.PukeForbidden && ((in.Left && !g.previousInput.Left) || (in.Right && !g.previousInput.Right)) {
		old := [4]bool{g.Lights[5], g.Lights[2], g.Lights[1], g.Lights[4]}
		for i, n := range []int{4, 5, 2, 1} {
			g.light(n, old[i])
		}
	}
	g.previousInput = in
}

func demoPlayerProgram() []presentation.Command {
	return []presentation.Command{{Op: "_CLEAR4"}, {Op: "_PRINT5", Args: []string{"PLAYERSTEXT", "340"}, Nums: map[int]int{1: 340}}, {Op: "_PRINT5", Args: []string{"BALLSTEXT", "1684"}, Nums: map[int]int{1: 1684}}, {Op: "0"}}
}
func (d *demoCore) gameplayIdle() error {
	if d.failure != nil {
		return d.failure
	}
	g := d.game
	if d.highScoreOperands {
		if err := d.checkHighScorePreflight(); err != nil {
			return err
		}
	}
	if d.infoPresentation && d.infoCount == 720 {
		return d.showInfo()
	}
	defer g.Display.FlushPrint(g.matrixNumber)

	probe := *g.Display
	if probe.TakeIdlePanel(g.inChute) {
		for _, c := range demoPlayerProgram() {
			if !d.command(c) {
				return d.reject("NODOT", "SHOWPLAYERSTS", "unadmitted player panel")
			}
		}
		g.Display.TakeIdlePanel(g.inChute)
		if err := d.installOwned(demoPlayerProgram(), false); err != nil {
			return err
		}
		if d.highScoreOperands && g.Score != Number(0) {
			d.infoCount++
		}
		g.Display.Score(g.Score.String())

		g.matrixDispatch()
		return nil
	}
	if !d.highScoreOperands && g.Score.Uint64() != 0 {
		return d.reject("NODOT", "CHECKHIGHSCORE", "nonzero score idle chain not admitted")
	}
	if d.highScoreOperands && g.Score != Number(0) {
		d.infoCount++
	}
	g.Display.Score(g.Score.String())
	return nil
}

func (d *demoConnected) bygelScore(label string, score, bonus uint64) error {
	if d.failure != nil {
		return d.failure
	}
	g := d.game
	g.score(label, score)
	d.infoCount = 0
	if d.fjantText {
		if err := d.install(demoPlayerProgram()); err != nil {
			return err
		}
		d.fjantText = false
	}
	g.Bonus.AddNumber(bonus)
	return nil
}
