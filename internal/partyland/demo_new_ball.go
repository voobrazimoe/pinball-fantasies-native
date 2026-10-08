//go:build dmoimpl1 || demodev

package partyland

import (
	"bytes"
	"pinballfantasies/internal/audio"
)

// These names denote shared WAITLIST call sites, not per-instance ages.
// Linked DS:36d1 / DS:36d3 / DS:36cf, respectively.
func demoChild(site string, delay uint16) bool {
	return (site == "SOUNDNEWBALL" && delay == 50) ||
		(site == "SETBALL" && delay == 80) || (site == "SOUNDBRICKUPP" && delay == 5)
}

// Uses pinned private inputs through the same adapter as presentation.
// The three reviewed RESTOREDn source pointers relocate by +0x100.
func (d *demoCore) loadResetOperands(demo []byte) error {
	if d.failure != nil {
		return d.failure
	}
	const ds = 0x19db0
	for i, at := range []int{0x69b0, 0x69f0, 0x6a40} {
		n := []int{30, 30, 15}[i]
		if ds+at+n > len(demo) || !bytes.Equal(demo[ds+at:ds+at+n], d.game.duckUp[i]) {
			return d.reject("NEW_BALL operands", "RESTORED1/2/3", "shared mask correspondence unavailable")
		}
	}
	text := d.game.Display.Content.Texts["BONUS_TEXT"]
	const textAt = ds + 0x2325
	if len(text) < 12 || textAt+len(text) > len(demo) || !bytes.Equal(text, demo[textAt:textAt+len(text)]) {
		return d.reject("NEW_BALL operands", "BONUS_TEXT", "mutable reset text correspondence unavailable")
	}
	d.game.Display.Content.Texts["BONUS_TEXT"] = append([]byte(nil), text...)

	for _, cue := range []struct {
		at           int
		label        string
		sample, note byte
	}{{0xc39, "SBRICKUPP", 23, 23}, {0xc51, "SNEWBALL", 28, 18}} {
		at := ds + cue.at
		if at+4 > len(demo) || demo[at] != cue.sample || demo[at+1] != cue.note || demo[at+3] != 3 || audio.Effects[cue.label] != (audio.Effect{Sample: int(cue.sample), Note: int(cue.note)}) {
			return d.reject("child operands", cue.label, "shared effect correspondence unavailable")
		}
	}
	d.childOperands = true
	d.resetOperands = true
	return nil
}

func (d *demoCore) preflightNewBall() error {
	if d.failure != nil {
		return d.failure
	}
	g := d.game

	if !d.resetOperands {
		return d.reject("PARTY_ON_TASK1", "NEW_BALL reset operands", "verified reset consumers unavailable")
	}
	if d.addPlayers || !g.musicOK {
		return d.reject("PARTY_ON_TASK1", "NEW_BALL ADDPLAYERS/MUSICOK", "unadmitted SNART_NEW_BALL or spring cue consumer")
	}
	if g.Session.CurrentPlayer < 1 || g.Session.CurrentPlayer > len(g.Session.Players) || g.Physics.Table == nil {
		return d.reject("PARTY_ON_TASK1", "P_STRUC_2_VARS", "saved player/table unavailable")
	}
	if len(g.Display.Content.Texts["BONUS_TEXT"]) < 12 {
		return d.reject("PARTY_ON_TASK1", "BONUS_TEXT", "reset text operand missing")
	}
	for i, size := range []int{30, 30, 15} {
		if len(g.duckUp[i]) != size {
			return d.reject("PARTY_ON_TASK1", "RESTORED1/2/3", "verified shared duck mask input unavailable")
		}
	}
	return nil
}

// Infallible after preflight, within the existing DO_TASKS scan. Reviewed
// generic/table resets are explicit; canonical NEW_BALL is never invoked.
func (d *demoCore) newBallHandoff() {
	if d.failure != nil {
		return
	}
	g := d.game
	d.loosing, d.specialMode = false, false
	g.Physics.Ball.Lost = false
	g.Physics.SpringValid = true
	d.shiftPressed, d.inhibitCountdown = false, false
	d.keyTaskEmpty, d.dotReady = true, true

	g.inChute = true
	g.Display.Content.Texts["BONUS_TEXT"][11] = '8'

	g.Lights = [57]bool{}
	g.Lamps = [57]bool{}
	g.lampPalette = g.paletteFor(g.Lamps)
	g.flashes = [15]flash{}
	g.tasks = [50]func() bool{}
	g.waitCounters = make(map[string]uint16)
	if d.scoredNewBall && !g.partyFlash {
		g.Display.KillFlash()
		if d.visaKeys {
			d.visaKeys = false
		} else {
			_ = d.install(demoScoredPlayerProgram())
		}
	}
	g.Audio.ReadyAnim = true
	g.Audio.ReadyLogic = true
	g.inhibitEffect = false
	g.effectEnded = true
	g.modeIntro = false
	g.snackSync = 0
	g.arrowSync = 0
	g.pukeSync = 0
	g.trainSync = 0
	g.touchDisabled = false
	g.duckDisabled = [3]bool{}
	g.snacks = [3]bool{}
	g.SnackNext = 0
	g.Pop = 0
	g.CrazyNext = 0
	g.Skyride = 0
	g.Puke = 0
	g.Balloon = 0
	g.Arcade = false
	g.SnackDisabled = false
	g.Dragon = false
	g.PukeForbidden = false
	g.MB = false
	g.HB = false
	g.DB = false
	g.FiveX = false
	g.FiveMillion = false
	g.BallFeature = false
	g.JackpotNormal = false
	g.JackpotTimed = false
	g.HoldBonus = false
	g.Happy = false
	g.Mega = false
	g.HappyPending, g.MegaPending = false, false
	g.InhibitLoop = false
	g.InhibitReverse = false
	g.LoopTime = 0
	g.ReverseTime = 0
	g.InhibitReverseTime = 0
	g.TunnelTime = 0
	g.Multiplier = 1
	d.bonusX = 1
	g.lastCheck = ""
	g.ScoreChanged = false
	for i := 0; i < 3; i++ {
		g.light(52+i, true)
		g.duckMask(i, false)
	}
	g.flash(14, 8, 0, false)
	g.flash(26, 9, 0, false)

	g.LoadPlayerState(g.Session.Load())
	for _, n := range []int{1, 2, 4, 5, 6, 8, 9, 41, 38, 34, 31, 28, 42, 43, 44, 45, 46} {
		g.light(n, g.Lights[n])
	}
	if g.ExtraBalls != 0 {
		g.light(51, true)
	}
	g.HappyTotal, g.MegaTotal = Decimal{}, Decimal{}
	d.holdStill = true
	g.Physics.SetBall(282, 530, 0, 0, false)
	g.Physics.TargetRaster = -1
	d.screenForce2 = -1
	g.Audio.ReturnPosition = 0
	g.ScoreChanged = false
	g.Phase = NewBall
	d.handoffs++

	for _, t := range []demoTask{
		{Site: "SOUNDNEWBALL", Delay: 50, Action: demoChildWait},
		{Site: "SETBALL", Delay: 80, Action: demoChildWait},
		{Site: "SOUNDBRICKUPP", Delay: 5, Action: demoChildWait},
	} {
		_ = d.queue(t)
	}
	g.Physics.ResetTilt()
}
