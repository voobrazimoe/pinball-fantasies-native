// Package speeddevils implements SDEV table rules around the shared native
// BALLCODE engine, FANTASIE task primitives and four-channel tracker.
package speeddevils

import (
	"encoding/json"
	"fmt"
	"pinballfantasies/internal/audio"
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
type rectangle struct {
	x1, y1, x2, y2 int16
	label          string
}

func (r rectangle) contains(x, y int16) bool {
	return uint16(x) >= uint16(r.x1) && uint16(x) <= uint16(r.x2) && uint16(y) >= uint16(r.y1) && uint16(y) <= uint16(r.y2)
}

type flash struct{ number, counter, speed uint8 }
type lamp struct {
	start int
	rgb   []byte
}
type effectSpec struct {
	Jingle, Matrix string
	Priority       uint8
	Score, Bonus   uint64
}
type jingleSpec = tablelogic.JingleSpec
type command struct {
	Op   string
	Args []string
	// Nums carries the extractor-resolved integer for every source numeric
	// operand, keyed by operand index. Args keeps the raw source text for
	// diagnostics and identity lookups; the runtime reads only Nums.
	Nums map[int]int `json:"nums,omitempty"`
}
type animation struct{ Header, Frames []uint16 }
type cue = tablelogic.CueSpec
type lampSpec struct{ Offset int }
type content struct {
	Attract    [][4]int
	Lamps      map[string]lampSpec
	Effects    map[string]effectSpec
	Jingles    map[string]jingleSpec
	Commands   []command
	Labels     map[string]int
	Animations map[string]animation
	Scrolls    map[string]uint16
	Cues       map[string]cue
}

var programs = func() content {
	var p content
	if e := json.Unmarshal([]byte(contentJSON), &p); e != nil {
		panic(e)
	}
	return p
}()

type Game struct {
	Session     hotseat.Session[PlayerState]
	MusicOff    bool
	Display     *presentation.Display
	lampPalette [768]byte

	Physics                                                                         *physics.Game
	Playback                                                                        *audio.Player
	AudioPCM                                                                        []byte
	Score, Bonus, Jackpot, OffRoadTotal, TurboTotal                                 Decimal
	BallNumber, Multiplier                                                          uint8
	Phase                                                                           Phase
	Tick                                                                            uint64
	Lights, Lamps                                                                   [68]bool
	Events                                                                          []Event
	Gear, Speed, Miles, NextJump, NextOffRoad, Position, PositionAvailable, CarPart uint16
	MBStock, MBCollected                                                            uint8
	HoldBonus, Turbo, OffRoad, Special                                              bool
	ModeTime                                                                        uint16
	ScreenForce                                                                     int16
	tasks                                                                           [20]func() bool
	ids                                                                             [20]uint64
	nextID                                                                          uint64
	waits                                                                           map[string]uint16
	flashes                                                                         [64]flash
	content                                                                         [68]lamp
	touchBusy                                                                       [6]bool
	pitDown                                                                         [3]uint16
	pitAll, gearDown, loopHigh, loopLow, jackDown                                   uint16
	loopH, loopL, jump                                                              bool
	lastArea, lastCheck                                                             string
	previous                                                                        physics.Inputs
	totalBalls                                                                      uint8
	snackDisabled, scoreChanged, beaten                                             bool
	top                                                                             *Decimal
	matchBall                                                                       bool
	clock, matchLast                                                                uint16
	inChute                                                                         bool
	posSync, scrollPhase                                                            uint8
	inhibitCountdown                                                                bool
	matrix                                                                          matrix
	music                                                                           music
	font5, font13                                                                   []byte
}

func New(t *physics.Table, data []byte) *Game {
	g := &Game{inChute: true, scrollPhase: 8, Physics: physics.New(t), BallNumber: 1, Multiplier: 1, totalBalls: 3, NextJump: 30, NextOffRoad: 40, ScreenForce: -1, waits: map[string]uint16{}}
	g.Display = presentation.New(2, data)
	g.Physics.SpringGraphics = g.Display.SpringGraphics()
	g.Jackpot = number(5_000_000)
	g.music.ReadyAnim = true
	g.music.ReadyLogic = true
	g.music.ReturnPosition = 0
	g.font13 = append([]byte(nil), data[0x1f170:0x1f170+42*13]...)
	g.font5 = append([]byte(nil), data[0x1f680:0x1f680+36*5]...)
	// Palette lamp records are located independently below during content loading.
	g.loadLamps(data)
	g.lampPalette = g.paletteFor(g.Lamps)
	g.Physics.OnEvent = g.consume
	g.Physics.BeforeTargets = g.beforeTargets
	g.Physics.AfterTargets = g.afterTargets
	g.Physics.BeforeLate = g.presentationTick
	g.Physics.ScrollForce = func() int16 { return g.ScreenForce }
	g.Cue("S_SPRING")
	g.Session.Initialize(1, g.SavePlayerState())
	return g
}
func (g *Game) emit(kind, label string, n uint64) {
	g.Events = append(g.Events, Event{g.Tick, kind, label, n})
}
func (g *Game) score(label string, n uint64) {
	g.Score.Add(number(n))
	g.scoreChanged = true
	g.emit("ScoreAwarded", label, n)
}
func (g *Game) effect(label string) {
	label = strings.ToUpper(label)
	s, ok := programs.Effects[label]
	if !ok {
		panic("unknown SDEV effect " + label)
	}
	accepted := s.Priority >= g.music.Priority
	if s.Jingle != "0" {
		if !g.Special || label == "LOSTBALL" {
			accepted = g.playJingle(s.Jingle)
		}
	}
	g.score(label, s.Score)
	g.Bonus.Add(number(s.Bonus))
	if accepted && !g.Special && s.Matrix != "0" {
		g.beginMatrix(s.Matrix)
	}
}
func (g *Game) light(n int, on bool) { g.Lights[n] = on; g.lamp(n, on) }
func (g *Game) lamp(n int, on bool) {
	g.Lamps[n] = on
	g.applyLamp(&g.lampPalette, n, on)
	g.emit("LampChanged", fmt.Sprint(n), boolNumber(on))
}
func boolNumber(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}
func (g *Game) flash(n, speed int) {
	g.endFlash(n)
	for i := range g.flashes {
		if g.flashes[i].number == 0 {
			g.flashes[i] = flash{uint8(n), 0, uint8(speed)}
			return
		}
	}
}
func (g *Game) syncFlash(n int, invert bool) {
	g.flash(n, 15)
	for i := range g.flashes {
		f := &g.flashes[i]
		if int(f.number) == n {
			f.counter = g.posSync
			if invert {
				f.counter = (f.counter + 15) % 30
			}
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
func (g *Game) off(n int) { g.endFlash(n); g.light(n, false) }
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
func (g *Game) task(f func() bool) { tablelogic.Add(g.tasks[:], g.ids[:], &g.nextID, f) }
func (g *Game) runTasks()          { tablelogic.Run(g.tasks[:], g.ids[:]) }
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
func down(v *uint16, f func()) {
	if *v != 0 {
		*v--
		if *v == 0 && f != nil {
			f()
		}
	}
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
		g.emit("Sound", "SFJADER", uint64(charge)*2)
		if g.Playback != nil && charge > 0 {
			g.Playback.SoundEffectVolume(sounds["SFJADER"], charge*2)
		}
	}
}
func (g *Game) consume(e physics.Event) {
	switch e.Kind {
	case physics.EventBumperHit:
		g.score("BUMPER", 1030)
		g.sound("SBUMPER")
		g.addOffRoad()
	case physics.EventSlingshotHit:
		g.score("KICKER", 510)
		g.sound("SKICKER")
	case physics.EventTargetHit:
		g.touch(e.Object)
	case physics.EventDrain:
		g.drain()
	}
}
func (g *Game) beforeTargets() {
	g.posSync = (g.posSync + 1) % 30
	for i := range g.pitDown {
		index := i
		down(&g.pitDown[i], func() { g.endFlash(index + 6); g.lamp(index+6, true) })
	}
	down(&g.pitAll, func() {
		for n := 6; n <= 8; n++ {
			g.off(n)
		}
	})
	down(&g.gearDown, g.resetGear)
	down(&g.loopHigh, nil)
	down(&g.loopLow, nil)
	down(&g.jackDown, func() { g.off(15) })
	g.checkAreas()
}
func (g *Game) afterTargets(in physics.Inputs) {
	g.tiltControl(in)
	if g.Physics.AllowFlip && (in.Left && !g.previous.Left || in.Right && !g.previous.Right) {
		g.rotate(16)
		g.rotate(19)
		g.touchBusy = [6]bool{}
		if g.pitAll == 0 {
			g.pitDown = [3]uint16{}
			g.rotate(6)
		}
	}
	g.previous = in
	g.runTasks()
	g.springControl(in)
	g.flashTick()
}
func (g *Game) rotate(base int) {
	old := [3]bool{g.Lights[base], g.Lights[base+1], g.Lights[base+2]}
	for i := 0; i < 3; i++ {
		g.light(base+i, old[(i+1)%3])
	}
}
func (g *Game) addOffRoad() {
	if g.OffRoad {
		g.OffRoadTotal.Add(number(100000))
	}
}
func (g *Game) addTurbo() {
	if g.Turbo {
		g.TurboTotal.Add(number(5_000_000))
	}
}
func (g *Game) jackAdd() { g.Jackpot.Add(number(100000)) }
func (g *Game) SetOriginalBallSetting(setting byte) {
	g.totalBalls = 3
	if setting != 0 {
		g.totalBalls = 5
	}
}
func (g *Game) SetHighScore(top Decimal) { g.top = &top }
func (g *Game) Result() (Decimal, bool)  { return g.Score, g.Phase == GameOver }
func (g *Game) InChute() bool            { return g.inChute }
func (g *Game) PCM() []byte              { return g.AudioPCM }
func (g *Game) PresentationAudioSync()   { g.audioTick() }

func (g *Game) AttractCue() string { return "S_NOHIGH" }
