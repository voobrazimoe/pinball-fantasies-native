// Package gameshow implements SHOW's native table-local electronics and rules.
package gameshow

import (
	"encoding/json"
	"fmt"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/hotseat"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"
	"strings"
)

type Decimal = tablelogic.Decimal

func number(n uint64) Decimal { return tablelogic.Number(n) }

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
}
type effectSpec struct {
	Jingle, Matrix string
	Priority       uint8
	Score, Bonus   uint64
}
type lamp struct {
	Ref   [2]int
	Start int
	RGB   []byte
}
type gate struct {
	Number, X, Y, Width, Height int
	High                        bool
	OpenedRef                   [4]int `json:"opened_ref"`
	ClosedRef                   [4]int `json:"closed_ref"`
}
type content struct {
	Effects   map[string]effectSpec
	Jingles   map[string]tablelogic.JingleSpec
	Cues      map[string]tablelogic.CueSpec
	Lamps     map[string]lamp
	LampOrder []int
	Gates     []gate
}

var programs = func() content {
	var c content
	if e := json.Unmarshal([]byte(contentJSON), &c); e != nil {
		panic(e)
	}
	return c
}()

type Game struct {
	Session                                                               hotseat.Session[PlayerState]
	Physics                                                               *physics.Game
	Display                                                               *presentation.Display
	Playback                                                              *audio.Player
	AudioPCM                                                              []byte
	Score, Bonus, Jackpot, CashPot, CashPot5, MoneyTotal, RaisingMillions Decimal
	BallNumber, Multiplier                                                uint8
	Phase                                                                 Phase
	Tick                                                                  uint64
	Lights, Lamps                                                         [39]bool
	Events                                                                []Event
	Prizes                                                                [6]uint8 // TV, TRIP, CAR, BOAT, HOUSE, PLANE: 0/unlit,1/lit,2/won
	TopThree, AllSix, BillionEnabled, Special, MoneyMania, LoopsAndTraps  bool
	Skills                                                                uint16
	ModeTime                                                              uint16
	ScreenForce                                                           int16
	MusicOff                                                              bool
	tasks                                                                 [20]func() bool
	ids                                                                   [20]uint64
	nextID                                                                uint64
	waits                                                                 map[string]uint16
	flashes                                                               [30]flash
	palette                                                               [768]byte
	music                                                                 tablelogic.MusicClock
	matrix                                                                matrix
	timers                                                                [11]uint16 // SJP, JP, MB, TV, BOAT, TRIP, CAR, CASH5, LOOP, HOUSE, PLANE
	syncFlasher                                                           uint8
	spinCounter                                                           uint16
	spinIndex, spinLight                                                  int
	spinning                                                              bool
	spinScore                                                             uint64
	previous                                                              physics.Inputs
	partyFlash                                                            bool
	inChute                                                               bool
	lastArea, lastCheck                                                   string
	top                                                                   *Decimal
	beaten, matchBall                                                     bool
	clock, matchLast                                                      uint16
	totalBalls                                                            uint8
}
type flash struct{ lamp, counter, speed uint8 }

func New(t *physics.Table, data []byte) *Game {
	g := &Game{Physics: physics.New(t), Display: presentation.New(3, data), BallNumber: 1, Multiplier: 1, ScreenForce: -1, inChute: true, totalBalls: 3, waits: map[string]uint16{}}
	g.Physics.SpringGraphics = g.Display.SpringGraphics()
	g.music.ReadyAnim = true
	g.music.ReadyLogic = true
	g.Jackpot = number(10_000_000)
	g.resetTable()
	g.Physics.OnEvent = g.consume
	g.Physics.BeforeTargets = g.beforeTargets
	g.Physics.AfterTargets = g.afterTargets
	g.Physics.BeforeLate = func() { g.Display.Flash(); g.matrixTick() }
	g.Physics.ScrollForce = func() int16 { return g.ScreenForce }
	g.Cue("S_SPRING")
	g.Session.Initialize(1, g.SavePlayerState())
	return g
}
func (g *Game) emit(k, l string, n uint64) { g.Events = append(g.Events, Event{g.Tick, k, l, n}) }
func (g *Game) score(l string, n uint64)   { g.Score.Add(number(n)); g.emit("ScoreAwarded", l, n) }
func (g *Game) effect(label string) {
	label = strings.ToUpper(label)
	s, ok := programs.Effects[label]
	if !ok {
		panic("unknown SHOW effect " + label)
	}
	accepted := s.Priority >= g.music.Priority
	if s.Jingle != "0" {
		if !g.Special || label == "LOSTBALL" {
			accepted = g.playJingle(s.Jingle)
		}
	}
	score := s.Score
	switch label {
	case "JACKPOT":
		score = g.Jackpot.Uint64()
	case "CASHPOT":
		score = g.CashPot.Uint64()
	case "CASHPOT5":
		score = g.CashPot5.Uint64()
	case "RAISING_M":
		score = g.RaisingMillions.Uint64()
	}
	g.score(label, score)
	g.Bonus.Add(number(s.Bonus))
	if accepted && !g.Special && s.Matrix != "0" {
		g.beginMatrix(s.Matrix)
	}
}
func (g *Game) task(f func() bool) { tablelogic.Add(g.tasks[:], g.ids[:], &g.nextID, f) }
func (g *Game) wait(site string, n uint16, f func()) {
	g.task(func() bool {
		if !tablelogic.Wait(g.waits, site, n) {
			return false
		}
		g.emit("TaskReady", site, 0)
		f()
		return true
	})
}
func (g *Game) runTasks()            { tablelogic.Run(g.tasks[:], g.ids[:]) }
func (g *Game) light(n int, on bool) { g.Lights[n] = on; g.lamp(n, on) }
func (g *Game) lamp(n int, on bool) {
	g.Lamps[n] = on
	g.applyLamp(&g.palette, n, on)
	g.emit("LampChanged", fmt.Sprint(n), 0)
}
func (g *Game) flash(n, speed int) {
	g.endFlash(n)
	for i := range g.flashes {
		if g.flashes[i].lamp == 0 {
			g.flashes[i] = flash{uint8(n), 0, uint8(speed)}
			return
		}
	}
}
func (g *Game) endFlash(n int) {
	for i, f := range g.flashes {
		if int(f.lamp) == n {
			g.flashes[i] = flash{}
			return
		}
	}
}
func (g *Game) off(n int) { g.endFlash(n); g.light(n, false) }
func (g *Game) synced(n int, inverted bool) {
	g.flash(n, 10)
	for i := range g.flashes {
		f := &g.flashes[i]
		if int(f.lamp) == n {
			f.counter = g.syncFlasher
			if inverted {
				f.counter = (f.counter + 10) % 20
			}
			return
		}
	}
}
func (g *Game) flashTick() {
	for i := range g.flashes {
		f := &g.flashes[i]
		if f.lamp == 0 {
			continue
		}
		if f.counter == 2*f.speed {
			f.counter = 0
		}
		if f.counter == 0 {
			g.lamp(int(f.lamp), true)
		} else if f.counter == f.speed {
			g.lamp(int(f.lamp), false)
		}
		f.counter++
	}
}
func (g *Game) gate(n int, open bool) {
	c := programs.Gates[n-1]
	b := g.Display.StridedRecord(c.ClosedRef)
	if open {
		b = g.Display.StridedRecord(c.OpenedRef)
	}
	g.Physics.PatchMask(c.High, c.X, c.Y, c.Width, c.Height, b)
	g.emit("Gate", fmt.Sprint(n), 0)
}
func (g *Game) Sync(in physics.Inputs) error {
	g.Events = g.Events[:0]
	g.Tick++
	if !g.inChute {
		g.Session.SelectionOpen = false
	}
	g.clock += 1030
	g.audioTick()
	if g.Phase == GameOver {
		return nil
	}
	if g.Phase == BallLost {
		g.runTasks()
		g.flashTick()
		g.Display.Flash()
		g.matrixTick()
		return nil
	}
	if g.Physics.AllowFlip && (in.Left && !g.previous.Left || in.Right && !g.previous.Right) {
		g.sound("SFLIPPUPP")
	}
	return g.Physics.Sync(in)
}
func (g *Game) Release(charge, jitter uint8) {
	if g.Phase == Playing || g.Phase == NewBall {
		g.Physics.Release(charge, jitter)
		if g.Playback != nil && charge > 0 {
			g.Playback.SoundEffectVolume(g.soundSpec("SFJADER"), charge*2)
		}
	}
}
func (g *Game) consume(e physics.Event) {
	switch e.Kind {
	case physics.EventBumperHit:
		g.score("BUMPER", 1000)
		g.sound("SBUMPER")
		g.addMoney(false)
	case physics.EventSlingshotHit:
		g.score("KICKER", 500)
		g.sound("SKICKER")
	case physics.EventTargetHit:
		g.touch(e.Object)
	case physics.EventDrain:
		g.drain()
	}
}
func (g *Game) beforeTargets() { g.updateCounters(); g.checkAreas() }
func (g *Game) afterTargets(in physics.Inputs) {
	warning, tilt := g.Physics.TiltInput(in.Tilt, g.inChute)
	if warning {
		g.playJingle("S_DANGER")
	}
	if tilt {
		g.playJingle("S_TILT")
		g.music.ReturnPosition = 55
		g.flashes = [30]flash{}
		for n := 1; n <= 38; n++ {
			g.light(n, false)
		}
		g.beginMatrix("TILTTS")
		g.emit("Tilt", "HE_TILTED", 0)
	}
	g.previous = in
	g.runTasks()
	gameplay.Spring(&g.Physics.SpringPosition, g.Physics.SpringValid, in, func(charge uint8) {
		g.Release(charge, uint8(g.clock))
	})
	g.flashTick()
}
func (g *Game) SetOriginalBallSetting(v byte) {
	g.totalBalls = 3
	if v != 0 {
		g.totalBalls = 5
	}
}
func (g *Game) SetHighScore(v Decimal)  { g.top = &v }
func (g *Game) Result() (Decimal, bool) { return g.Score, g.Phase == GameOver }
func (g *Game) PCM() []byte             { return g.AudioPCM }
func (g *Game) InChute() bool           { return g.inChute }
func (g *Game) AttractCue() string      { return "S_NOHIGH" }
func (g *Game) PresentationAudioSync()  { g.audioTick() }
func (g *Game) beatHighScore() bool {
	if g.top == nil || g.beaten || g.Score.Uint64() <= g.top.Uint64() {
		return false
	}
	g.beaten = true
	g.emit("HighScoreBeaten", "CHECKHIGHSCORE", g.Score.Uint64())
	return true
}
