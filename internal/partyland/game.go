// Package partyland ports PLAND's native rule subsystem. Physics remains
// the PF3 integer implementation; this package owns its original callback state.
package partyland

import (
	"fmt"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/hotseat"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"
)

// Preserve the PF4 public types while sharing original ADDSCOREBCD semantics.
type Decimal = tablelogic.Decimal

func Number(n uint64) Decimal { return tablelogic.Number(n) }

type Phase uint8

const (
	Playing Phase = iota
	BallLost
	NewBall
	GameOver
)

type Event struct {
	Tick        uint64
	Kind, Label string
	Value       uint64
	Object      int
}
type flash struct{ number, counter, speed uint8 }
type lampContent struct {
	start int
	rgb   []byte
}

type Game struct {
	Session     hotseat.Session[PlayerState]
	MusicOff    bool
	Display     *presentation.Display
	lampPalette [768]byte

	highScore                                                                      *Decimal
	alreadyBeaten                                                                  bool
	savedJingle                                                                    uint8
	musicOK                                                                        bool
	matrixTimeLeft                                                                 bool
	dropTime                                                                       uint16
	ScreenForce, ScrollPosition                                                    int16
	modeIntro                                                                      bool
	matchRemaining, matchTimer                                                     uint16
	matchLast                                                                      uint16
	Tick                                                                           uint64
	Audio                                                                          silentJingle
	Playback                                                                       *audio.Player
	AudioPCM                                                                       []byte
	matrix                                                                         matrixState
	waitCounters                                                                   map[string]uint16
	scrollPhase                                                                    uint8
	dropMinimum, dropMaximum                                                       uint16
	dropReady, arcadeCrazy                                                         bool
	spinReady, effectEnded, effectAccepted, inhibitEffect                          bool
	plungerSound                                                                   bool
	plungerVolume                                                                  byte
	taskIDs                                                                        [50]uint64
	nextTaskID                                                                     uint64
	Physics                                                                        *physics.Game
	Score, Bonus, Jackpot, SkillTunnel, SkillCyclone, HappyTotal, MegaTotal        Decimal
	BallNumber, ExtraBalls, Multiplier                                             uint8
	Phase                                                                          Phase
	ScoreChanged                                                                   bool
	Lights                                                                         [57]bool // LIGHTSTATUS: flashing changes the palette, never these bits.
	Lamps                                                                          [57]bool // physical LON/LOFF palette state
	Events                                                                         []Event  // bounded to the current sync, including deferred audio intents
	Cyclones                                                                       uint16
	Skyride, Puke                                                                  uint16
	SnackNext, Pop, CrazyNext, Balloon, Random                                     uint8
	Arcade, SnackDisabled, Dragon, PukeForbidden                                   bool
	MB, HB, DB, FiveX, FiveMillion, BallFeature, JackpotNormal, JackpotTimed       bool
	HoldBonus, Happy, Mega, HappyPending, MegaPending, InhibitLoop, InhibitReverse bool
	SkillTime, LoopTime, ReverseTime, InhibitReverseTime, TunnelTime               uint16
	ModeTime                                                                       uint16
	lastCheck, lastArea                                                            string
	previousInput                                                                  physics.Inputs
	tasks                                                                          [50]func() bool
	flashes                                                                        [15]flash
	snackSync, arrowSync, pukeSync, trainSync                                      uint8
	content                                                                        [57]lampContent
	font5, font13                                                                  []byte
	duckUp, duckDown                                                               [3][]byte
	gateClosed                                                                     []byte
	totalBalls                                                                     uint8
	clock                                                                          uint16
	matchBall                                                                      bool
	partyFlash                                                                     bool
	inChute                                                                        bool
	touchDisabled                                                                  bool
	duckDisabled                                                                   [3]bool
	snacks                                                                         [3]bool
}

func New(table *physics.Table, data []byte) *Game {
	g := &Game{matrixTimeLeft: true, ScreenForce: -1, Physics: physics.New(table), totalBalls: 3, scrollPhase: 8}
	g.Display = presentation.New(1, data)
	g.Physics.SpringGraphics = g.Display.SpringGraphics()
	g.Audio.LastPriority = 1
	g.Audio.Priority = 1
	g.Audio.ReadyAnim = true
	g.Audio.ReadyLogic = true
	// Original external content, statically located in SHA-validated TABLE1.PRG.
	g.font13 = append([]byte(nil), data[0x1ff40:0x1ff40+36*13]...)
	g.font5 = append([]byte(nil), data[0x20450:0x20450+36*5]...)
	p := 0x1adb9
	for i := 1; i <= 56; i++ {
		if i == 40 {
			g.content[i] = g.content[39]
		}
		start, count := int(data[p]), int(data[p+1])
		p += 2
		if i != 40 {
			g.content[i] = lampContent{start, append([]byte(nil), data[p:p+3*count]...)}
		}
		p += 3 * count
	}
	for i, o := range []int{0x68b0, 0x68f0, 0x6940} {
		size := 30
		if i == 2 {
			size = 15
		}
		g.duckUp[i] = append([]byte(nil), data[0x19d40+o:0x19d40+o+size]...)
	}
	for i, o := range []int{0x68d0, 0x6910, 0x6930} {
		size := 30
		if i == 2 {
			size = 15
		}
		g.duckDown[i] = append([]byte(nil), data[0x19d40+o:0x19d40+o+size]...)
	}
	g.gateClosed = g.Physics.MaskRegion(true, 14, 15, 2, 18)
	g.Physics.OnEvent = g.consume
	g.Physics.BeforeTargets = g.beforeTargets
	g.Physics.AfterTargets = g.afterTargets
	g.Physics.BeforeLate = g.presentationTick
	g.Physics.ScrollForce = func() int16 { return g.ScreenForce }
	g.BallNumber = 1
	g.Jackpot = Number(10_000_000)
	g.resetBall()
	g.Session.Initialize(1, g.SavePlayerState())
	return g
}
func (g *Game) emit(kind, label string, value uint64) {
	g.Events = append(g.Events, Event{Tick: g.Tick, Kind: kind, Label: label, Value: value})
}
func (g *Game) sound(label string) {
	g.emit("Sound", label, 0)
	if g.Playback != nil {
		g.Playback.Sound(label)
	}
}
func (g *Game) music(label string) { g.playJingle(label) }
func (g *Game) score(label string, n uint64) {
	g.Score.AddNumber(n)
	g.ScoreChanged = true
	g.emit("ScoreAwarded", label, n)
}
func (g *Game) bonus(n uint64) {
	for i := uint8(0); i < g.Multiplier; i++ {
		g.Bonus.AddNumber(n)
	}
}

// DOEFFECT requests the jingle before arithmetic. Priority gates the matrix,
// never the score or bonus. Suppression explicitly sets the shared EOTS flag.
func (g *Game) effect(label string, score, bonus uint64) {
	accepted := g.effectTiming(label)
	g.score(label, score)
	g.Bonus.AddNumber(bonus)
	g.effectAccepted = accepted
	if accepted {
		key := label
		if key == "MULTIBONUS" {
			key = fmt.Sprintf("M%d", g.Multiplier+2)
			if g.Multiplier == 1 {
				key = "M2"
			}
		}
		g.beginMatrix(timing.Effects[key].Matrix)
	} else {
		g.effectEnded = true
	}
}
func (g *Game) jackAdd()             { g.Jackpot.AddNumber(50_000) }
func (g *Game) light(n int, on bool) { g.Lights[n] = on; g.lamp(n, on) }
func (g *Game) lamp(n int, on bool) {
	g.Lamps[n] = on
	g.applyLamp(&g.lampPalette, n, on)
	if n == 39 || n == 40 {
		g.Lamps[39] = on
		g.Lamps[40] = on
	}
	g.emit("LampChanged", fmt.Sprint(n), uint64(boolByte(on)))
}
func boolByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
func (g *Game) flash(n, speed int, sync uint8, invert bool) {
	if invert {
		sync = (sync + uint8(speed)) % uint8(2*speed)
	}
	for i := range g.flashes {
		if g.flashes[i].number == 0 {
			g.flashes[i] = flash{uint8(n), sync, uint8(speed)}
			return
		}
	}
}
func (g *Game) endFlash(n int) {
	for i, f := range g.flashes {
		if int(f.number) == n {
			g.flashes[i] = flash{}
			return
		}
	}
}
func (g *Game) off(n int)          { g.endFlash(n); g.light(n, false) }
func (g *Game) task(f func() bool) { tablelogic.Add(g.tasks[:], g.taskIDs[:], &g.nextTaskID, f) }
func (g *Game) runTasks()          { tablelogic.Run(g.tasks[:], g.taskIDs[:]) }
func dec(v *uint16) {
	if *v != 0 {
		*v--
	}
}
func (g *Game) Sync(input physics.Inputs) error { return g.SyncWithMatrixBudget(input, true) }

// SyncWithMatrixBudget represents the original sound-driver TIME_LEFT input.
// Silent play has the budget every sync. Explicit inputs allow deterministic
// coverage of a missed matrix scan without consulting host performance.
func (g *Game) SyncWithMatrixBudget(input physics.Inputs, timeLeft bool) error {
	g.matrixTimeLeft = timeLeft
	g.Events = g.Events[:0]
	g.Tick++
	if !g.inChute {
		g.Session.SelectionOpen = false
	}
	// Original keyboard DSSK/DSSKR triggers the fourth voice on each press edge.
	if g.Phase == Playing {
		if g.Physics.AllowFlip && input.Left && !g.previousInput.Left {
			g.sound("SFLIPPUPP")
		}
		if g.Physics.AllowFlip && input.Right && !g.previousInput.Right {
			g.sound("SFLIPPUPP")
		}
	}
	if g.plungerSound {
		g.emit("Sound", "SFJADER", uint64(g.plungerVolume))
		if g.Playback != nil {
			g.Playback.SoundVolume("SFJADER", g.plungerVolume)
		}
		g.plungerSound = false
	}
	g.clock += 1030
	g.audioTick()
	if g.Phase == GameOver {
		return nil
	}
	if g.Phase == BallLost {

		g.Display.Flash() // MATRIX_BLINKOR precedes DO_TASKS.
		g.runTasks()
		g.flashTick()
		if g.Phase == GameOver {
			return nil // Frontend continues DEMOMODE NODOT in this sync.
		}
		g.matrixTick()
		if g.ScreenForce >= 0 {
			g.Physics.Raster = (g.ScreenForce + 33) * 16
		}
		return nil
	}
	err := g.Physics.Sync(input)
	// Drain bypasses electronics in PF3. Continue the ball-loss display tasks.
	if g.Phase == BallLost {
		g.Display.Flash() // MATRIX_BLINKOR precedes DO_TASKS.
		g.runTasks()
		g.flashTick()
		if g.Phase == GameOver {
			return nil // Frontend continues DEMOMODE NODOT in this sync.
		}
		g.matrixTick()
	}
	if g.ScreenForce >= 0 {
		g.Physics.Raster = (g.ScreenForce + 33) * 16
	}
	return err
}
func (g *Game) Release(charge, jitter uint8) {
	if g.Phase == Playing || g.Phase == NewBall {
		g.Physics.Release(charge, jitter)
		g.plungerSound = charge > 0
		g.plungerVolume = charge * 2
	}
}
func (g *Game) consume(e physics.Event) {
	switch e.Kind {
	case physics.EventBumperHit:
		g.score("BUMPER", 1000)
		g.sound(fmt.Sprintf("SBUMPER%d", e.Object+1))
		g.addHappy()
	case physics.EventSlingshotHit:
		g.score("KICKER", 500)
		g.sound("SKICKER")
		g.addHappy()
	case physics.EventTargetHit:
		if e.Object == 0 {
			g.toucher()
		} else {
			g.duck(e.Object - 1)
		}
	case physics.EventDrain:
		g.drain()
	}
}
func (g *Game) flashTick() {
	for i := range g.flashes {
		f := &g.flashes[i]
		if f.number == 0 {
			continue
		}
		if f.counter == 2*f.speed {
			f.counter = 0
		}
		if f.counter == 0 {
			g.lamp(int(f.number), true)
		} else if f.counter == f.speed {
			g.lamp(int(f.number), false)
		}
		f.counter++
	}
}
func (g *Game) beforeTargets() {
	g.Random++
	dec(&g.SkillTime)
	dec(&g.LoopTime)
	dec(&g.ReverseTime)
	dec(&g.InhibitReverseTime)
	g.snackSync = (g.snackSync + 1) % 16
	g.arrowSync = (g.arrowSync + 1) % 24
	g.pukeSync = (g.pukeSync + 1) % 4
	g.trainSync = (g.trainSync + 1) % 28
	if g.TunnelTime != 0 {
		g.TunnelTime--
		if g.TunnelTime == 720 {
			g.off(10)
			g.Lights[12] = false
			g.flash(12, 8, 0, false)
		}
		if g.TunnelTime == 0 {
			g.off(12)
			g.Lights[14] = false
			g.flash(14, 8, 0, false)
		}
	}
	g.checkAreas()
}
func (g *Game) afterTargets(input physics.Inputs) {
	g.tiltControl(input)
	if g.Physics.AllowFlip && !g.PukeForbidden && ((input.Left && !g.previousInput.Left) || (input.Right && !g.previousInput.Right)) {
		old := [4]bool{g.Lights[5], g.Lights[2], g.Lights[1], g.Lights[4]}
		for i, n := range []int{4, 5, 2, 1} {
			g.light(n, old[i])
		}
	}
	g.previousInput = input
	g.Display.Flash() // MATRIX_BLINKOR precedes DO_TASKS.
	g.runTasks()
	g.springControl(input)
	g.flashTick()
}
func (g *Game) resetBall() {
	g.Physics.ResetTilt()
	g.inChute = true
	// PLAYER_STRUC's progression lamps survive; temporary lamps and timers reset.
	keep := g.Lights
	g.Lights = [57]bool{}
	g.Lamps = [57]bool{}
	g.lampPalette = g.paletteFor(g.Lamps)
	g.flashes = [15]flash{}
	g.tasks = [50]func() bool{}
	g.waitCounters = make(map[string]uint16)
	if !g.partyFlash {
		g.matrix = matrixState{}
	}
	g.Audio.ReadyAnim = true
	g.Audio.ReadyLogic = true
	g.inhibitEffect = false
	g.effectEnded = true
	g.modeIntro = false
	for _, n := range []int{1, 2, 4, 5, 6, 8, 9, 41, 38, 34, 31, 28, 42, 43, 44, 45, 46} {
		g.light(n, keep[n])
	}
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
	g.SkillTime = 0
	g.LoopTime = 0
	g.ReverseTime = 0
	g.InhibitReverseTime = 0
	g.TunnelTime = 0
	g.ModeTime = 0
	g.Multiplier = 1
	g.lastCheck = ""
	g.lastArea = ""
	g.ScoreChanged = false
	for i := 0; i < 3; i++ {
		g.light(52+i, true)
		g.duckMask(i, false)
	}
	g.flash(14, 8, 0, false)
	g.flash(26, 9, 0, false)
	if g.ExtraBalls != 0 {
		g.light(51, true)
	}
}
func (g *Game) duckMask(i int, down bool) {
	x := [3]int{18, 19, 20}
	y := [3]int{277, 295, 313}
	w := 2
	if i == 2 {
		w = 1
	}
	data := g.duckUp[i]
	if down {
		data = g.duckDown[i]
	}
	g.Physics.PatchMask(false, x[i], y[i], w, 15, data)
}
func (g *Game) gate(open bool) {
	data := g.gateClosed
	if open {
		data = make([]byte, 36)
	}
	g.Physics.PatchMask(true, 14, 15, 2, 18, data)
}

// SetOriginalBallSetting selects FANTASIES' S_BALLS path (0: three, nonzero:
// five). The frontend reads only this original setting; no options UI is added.
func (g *Game) SetOriginalBallSetting(setting byte) {
	g.totalBalls = 3
	if setting != 0 {
		g.totalBalls = 5
	}
}

// AttachAudio adds deterministic native playback without changing semantic state.
// The renderer's Bxx target passes through the already established handler state.
func (g *Game) AttachAudio(m *audio.Module) {
	g.Playback = audio.New(m)
	if g.Audio.Active {
		g.Playback.Force(int(g.Audio.Position))
	}
	g.Playback.Jump = func(next int) int {
		if g.Audio.JumpCount == 1 {
			return int(g.Audio.ReturnPosition)
		}
		return next
	}
}

// PlungerValid exposes the one canonical source spring gate to host adapters.
func (g *Game) PlungerValid() bool { return g.Physics.SpringValid }
