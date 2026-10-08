//go:build dmoimpl1 || demodev

package partyland

import (
	"bytes"

	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"
	"reflect"
)

// Only the linked zero-aggregate route is admitted. These are real source
// branches, not a zero-bonus shortcut: every node is visited in source order.
func demoLostProgram() []presentation.Command {
	p := []presentation.Command{{Op: "_CLEAR4"}, {Op: "_PRINT13", Args: []string{"DEMO_DRAIN_TEXT", "344"}, Nums: map[int]int{1: 344}}, {Op: "_WAIT", Args: []string{"80"}, Nums: map[int]int{0: 80}}, {Op: "_CLEAR4"}}
	for _, name := range []string{"BONUSSIFFRORNA", "CYCLONECOUNTERBCD", "HAPPY_HOUR_TOTAL", "MEGA_LAUGH_TOTAL", "BONUSSIFFRORNA"} {
		p = append(p, presentation.Command{Op: "_DEMO_ZERO_BONUS", Args: []string{name}})
	}
	return append(p, presentation.Command{Op: "_DEMO_KOLLA_XXBALL"}, presentation.Command{Op: "_DEMOVER_CHANGE_PLAYER"}, presentation.Command{Op: "_CLEAR4"}, presentation.Command{Op: "_WAIT", Args: []string{"32000"}, Nums: map[int]int{0: 32000}}, presentation.Command{Op: "0"})
}
func demoScoredPlayerProgram() []presentation.Command {
	return []presentation.Command{{Op: "_CLEAR4"}, {Op: "_PRINT5", Args: []string{"PLAYERSTEXT", "336"}, Nums: map[int]int{1: 336}}, {Op: "_PRINT5", Args: []string{"DEMO_BALLSTEXT", "1684"}, Nums: map[int]int{1: 1684}}, {Op: "0"}}
}

func (d *demoCore) scoredCommand(c presentation.Command) bool {
	if !d.scoredDrainOperands {
		return false
	}
	for _, p := range [][]presentation.Command{demoLostProgram(), demoScoredPlayerProgram()} {
		for _, w := range p {
			if c.Op != w.Op || !reflect.DeepEqual(c.Args, w.Args) || !sameDemoNums(c.Nums, w.Nums) {
				continue
			}
			switch c.Op {
			case "_PRINT13":
				return d.game.Display.Content.Fonts["13"] == d.drainFont && bytes.Equal(d.game.Display.Content.Texts["DEMO_DRAIN_TEXT"], d.drainText)
			case "_PRINT5":
				return d.game.Display.Content.Fonts["5"] == d.gameplayFont && bytes.Equal(d.game.Display.Content.Texts["DEMO_BALLSTEXT"], d.drainBallText) && bytes.Equal(d.game.Display.Content.Texts["PLAYERSTEXT"], d.panelTexts[0])
			case "_DEMO_ZERO_BONUS", "_DEMO_KOLLA_XXBALL", "_DEMOVER_CHANGE_PLAYER":
				return true
			}
		}
	}
	return false
}
func (d *demoConnected) scoredDrain() error {
	if d.failure != nil {
		return d.failure
	}
	if d.expired {
		return d.rejectEdge("LOOSE_BALL", "minute5 effect0x1a4a1", "expired scored restart not admitted before effect side effects")
	}
	g := d.game

	if !d.scoredDrainOperands || !g.ScoreChanged || !d.loosing || !d.holdStill || g.Playback != nil || !d.scoredCommand(demoLostProgram()[1]) {
		return d.rejectEdge("LOOSE_BALL", "LOSTBALL operands", "mandatory effect/sound/presentation unavailable")
	}

	if d.fjantText {
		for _, c := range demoScoredPlayerProgram() {
			if !d.command(c) {
				return d.rejectEdge("LOSTBALL", "UPDAT_INFOBAR panel", "source panel unavailable before effect")
			}
		}
	}
	g.Physics.TargetRaster = d.screenForce2
	g.Audio.Priority = 0
	a := g.musicClock()
	accepted := true
	if !g.inhibitEffect {
		accepted = a.Play(tablelogic.JingleSpec{Position: 6, Repeat: 1, Priority: 255}, 62, timing.Cues)
		g.storeMusicClock(a)
		g.emit("Music", "S_LOSTBALL", 6)
	}
	d.infoCount = 0
	if d.fjantText {
		if err := d.install(demoScoredPlayerProgram()); err != nil {
			return err
		}
		d.fjantText = false
	}
	g.Score.AddNumber(0)
	g.ScoreChanged = true
	g.Bonus.AddNumber(0)
	g.effectAccepted = accepted && !g.inhibitEffect && !d.specialMode
	g.effectEnded = !g.effectAccepted
	if accepted && !g.inhibitEffect && !d.specialMode {
		if err := d.install(demoLostProgram()); err != nil {
			return err
		}
	}
	g.Audio.Priority = 0
	g.Audio.ReturnPosition = 62
	if err := d.queue(demoTask{Site: "SOUNDRINNER", Delay: 5, Action: demoScoredWait}); err != nil {
		return err
	}
	d.snapshot("LOSTBALL installed")
	return nil
}
func (d *demoCore) hasFreeTask() bool {
	for _, t := range d.game.tasks {
		if t == nil {
			return true
		}
	}
	return false
}
func (d *demoCore) preflightScoredTask(site string) error {
	if d.failure != nil {
		return d.failure
	}
	if !d.scoredDrainOperands || d.game.Playback != nil {
		return d.reject(site, "scored operands", "mandatory consumer unavailable before WAIT reset")
	}
	if site == "SOUNDRINNER" {
		if audio.Effects["SRINNER"] != (audio.Effect{Sample: 28, Note: 18}) {
			return d.reject(site, "SRINNER", "shared sound operands changed before WAIT reset")
		}
		return nil
	}
	if site != "NEW_BALL_TASK" {
		return d.reject(site, "identity", "unknown scored task")
	}
	g := d.game
	if d.addPlayers || !d.resetOperands || !d.childOperands || g.Physics.Table == nil || g.Session.CurrentPlayer < 1 || g.Session.CurrentPlayer > len(g.Session.Players) || len(g.Display.Content.Texts["BONUS_TEXT"]) < 12 {
		return d.reject(site, "NEW_BALL", "reset/child consumers unavailable before WAIT reset")
	}
	for i, n := range []int{30, 30, 15} {
		if len(g.duckUp[i]) != n {
			return d.reject(site, "duck masks", "missing reset operand")
		}
	}
	if !g.partyFlash && !d.visaKeys {
		for _, c := range demoScoredPlayerProgram() {
			if !d.command(c) {
				return d.reject(site, "SHOWPLAYERSTS", "missing source reset panel before WAIT reset")
			}
		}
	}
	return nil
}
func (d *demoCore) executeScoredTask(site string) {
	if d.preflightScoredTask(site) != nil {
		return
	}
	g := d.game
	if site == "SOUNDRINNER" {
		g.sound("SRINNER")
	} else {
		d.scoredDueMatrix = g.matrix
		d.scoredDueAudio = g.Audio
		d.scoredDueExpiry = len(g.Display.Content.Commands) > 1 && g.Display.Content.Commands[1].Op == "_SCROLL" && g.Display.Content.Commands[1].Args[0] == "DEMO_EXPIRY_TEXT1"

		d.scoredNewBall = true
		d.newBallHandoff()
		d.scoredNewBall = false
		if !g.musicOK {
			a := g.musicClock()
			accepted := a.Play(tablelogic.JingleSpec{Position: 0, Repeat: 0, Priority: 1}, 62, timing.Cues)
			g.storeMusicClock(a)
			if accepted {
				g.emit("Music", "S_SPRING", 0)
			} else {
				g.emit("AudioRejected", "S_SPRING", 1)
			}
			g.musicOK = false
		}
		g.Audio.ReturnPosition = 0
	}
	d.drainFires = append(d.drainFires, demoChildFire{site, d.counter, g.Physics.Ball, d.holdStill, g.inChute, g.Audio})
}
func (d *demoCore) dispatchScored(c matrixCommand) bool {
	if c.Op != "_DEMO_ZERO_BONUS" && c.Op != "_DEMO_KOLLA_XXBALL" && c.Op != "_DEMOVER_CHANGE_PLAYER" {
		return false
	}
	if d.failure != nil {
		return true
	}
	if !d.scoredDrainOperands {
		d.reject(c.Op, "admission", "missing scored operands")
		return true
	}
	g := d.game
	p := demoLostProgram()
	at := g.matrix.next - 1
	if at < 0 || at >= len(p) || c.Op != p[at].Op || !reflect.DeepEqual(c.Args, p[at].Args) || !sameDemoNums(c.Nums, p[at].Nums) {
		d.reject(c.Op, "source cursor/operands", "unadmitted bonus dispatch")
		return true
	}
	if c.Op == "_DEMO_ZERO_BONUS" {
		var value uint64
		switch c.Args[0] {
		case "BONUSSIFFRORNA":
			value = g.Bonus.Uint64()
		case "CYCLONECOUNTERBCD":
			value = uint64(g.Cyclones)
		case "HAPPY_HOUR_TOTAL":
			value = g.HappyTotal.Uint64()
		case "MEGA_LAUGH_TOTAL":
			value = g.MegaTotal.Uint64()
		}
		if value != 0 {
			d.reject("JBCDZ", c.Args[0], "nonzero fallthrough consumer not admitted")
			return true
		}
		d.bonusNodes = append(d.bonusNodes, d.counter)
	} else if c.Op == "_DEMO_KOLLA_XXBALL" {
		if g.matchBall {
			d.reject("KOLLA_XXBALL", "XXBALLE=true", "alternate match/extra ball consumer unavailable")
			return true
		}
	} else {
		if g.HoldBonus || !d.hasFreeTask() || !bytes.Equal(g.Display.Content.Texts["DEMO_BALLSTEXT"], d.drainBallText) {
			d.reject("_DEMOVER_CHANGE_PLAYER", "hold/allocation/text", "unadmitted producer branch")
			return true
		}
		text := g.Display.Content.Texts["DEMO_BALLSTEXT"]
		text[5]++
		if text[5] >= 65 {
			text[5] = 55
			if text[4] == 56 {
				text[4]++
			} else {
				text[4] = 56
			}
		}
		d.drainBallText = append([]byte(nil), text...)
		g.Session.Save(g.SavePlayerState())
		if d.queue(demoTask{Site: "NEW_BALL_TASK", Delay: 30, Action: demoScoredWait}) != nil {
			return true
		}
	}

	next := g.matrix.next
	if next >= len(p) || next >= len(g.Display.Content.Commands) {
		d.reject("HU_", "cursor", "missing next source consumer")
		return true
	}
	actual := g.Display.Content.Commands[next]
	want := p[next]
	if actual.Op != want.Op || !reflect.DeepEqual(actual.Args, want.Args) || !sameDemoNums(actual.Nums, want.Nums) || !d.command(actual) {
		d.reject("HU_", "next consumer", "unadmitted tail consumer before side effects")
		return true
	}
	g.matrixDispatch()
	return true
}
