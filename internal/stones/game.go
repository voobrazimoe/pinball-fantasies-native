// Package stones implements STONES table-local rules on the native shared engine.
package stones

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
	Name                string
	X, Y, Width, Height int
	High                bool
	OpenedRef           [4]int `json:"opened_ref"`
	ClosedRef           [4]int `json:"closed_ref"`
}
type content struct {
	Effects      map[string]effectSpec
	Jingles      map[string]tablelogic.JingleSpec
	Cues         map[string]tablelogic.CueSpec
	Lamps        map[string]lamp
	LampOrder    []int `json:"lamp_order"`
	Gates        []gate
	AreaRefs     map[string][2]int `json:"area_refs"`
	AreaHandlers map[uint16]string `json:"area_handlers"`
}

var programs = func() content {
	var c content
	if e := json.Unmarshal([]byte(contentJSON), &c); e != nil {
		panic(e)
	}
	return c
}()

type flash struct{ lamp, counter, speed uint8 }
type Game struct {
	Session                                                                                                           hotseat.Session[PlayerState]
	areas                                                                                                             map[string][]areaRecord
	Physics                                                                                                           *physics.Game
	Display                                                                                                           *presentation.Display
	Playback                                                                                                          *audio.Player
	AudioPCM                                                                                                          []byte
	Score, Bonus, Jackpot, TowerValue, VaultValue, WellValue, SkillScore, SulpScore, GhostTotal, GrimTotal, heldBonus Decimal
	BallNumber, Multiplier, ExtraBalls, GhostCounter, SkillKey                                                        uint8
	Phase                                                                                                             Phase
	Tick                                                                                                              uint64
	Lights, Lamps                                                                                                     [45]bool
	Events                                                                                                            []Event
	Special, GhostHunt, Grim, TowerHunt, MultiDemon                                                                   bool
	TowerStage, ComboStage                                                                                            uint8
	Screams, NextJump                                                                                                 uint16
	ModeTime                                                                                                          uint16
	ScreenForce                                                                                                       int16
	MusicOff                                                                                                          bool
	tasks                                                                                                             [20]func() bool
	ids                                                                                                               [20]uint64
	nextID                                                                                                            uint64
	waits                                                                                                             map[string]uint16
	flashes                                                                                                           [64]flash
	palette                                                                                                           [768]byte
	music                                                                                                             tablelogic.MusicClock
	matrix                                                                                                            matrix
	timers                                                                                                            [6]uint16 // left ramp, multi demon, SULP, MB, combo, tower hunt
	previous                                                                                                          physics.Inputs
	clock                                                                                                             uint16
	inChute, partyFlash, beaten, matchBall, holdBonus, holdMulti, inhibit, wasSpecial, grimBackup, fixTeleport        bool
	inhEff                                                                                                            bool
	keyDisabled, keyRollDisabled, ripDisabled, ripRollDisabled, sbDisabled, ghostInhibit, ghostFlashing               bool
	sbGuards                                                                                                          [11]bool
	kickback                                                                                                          bool
	matrixOn                                                                                                          bool
	syncSulp, syncTower                                                                                               uint8
	towerOpen, highVault                                                                                              bool
	towerHuntOrig                                                                                                     bool
	captured                                                                                                          string
	towerNext                                                                                                         int
	bonusPointer                                                                                                      uint8
	lastCheck                                                                                                         string
	totalBalls                                                                                                        uint8
	top                                                                                                               *Decimal
}

func New(t *physics.Table, data []byte) *Game {
	g := &Game{Physics: physics.New(t), Display: presentation.New(4, data), BallNumber: 1, totalBalls: 3, Jackpot: number(10_000_000), inChute: true, ScreenForce: -1, waits: map[string]uint16{}}
	g.areas = decodeAreas(data, programs.AreaRefs, programs.AreaHandlers)
	g.Display.UseSourceMatrixOff()
	g.Physics.SpringGraphics = g.Display.SpringGraphics()
	g.music.ReadyAnim = true
	g.music.ReadyLogic = true
	g.resetBall()
	g.Physics.OnEvent = g.consume
	g.Physics.BeforeTargets = func() { g.updateCounters(); g.checkAreas() }
	g.Physics.AfterTargets = g.afterTargets
	g.Physics.BeforeLate = func() { g.matrixTick(); g.matrixPalette(false) }
	g.Physics.ScrollForce = func() int16 { return g.ScreenForce }
	g.Cue("S_SPRING1")
	g.music.ReturnPosition = 0
	g.Session.Initialize(1, g.SavePlayerState())
	return g
}
func (g *Game) emit(k, l string, n uint64)  { g.Events = append(g.Events, Event{g.Tick, k, l, n}) }
func (g *Game) score(l string, n uint64)    { g.Score.Add(number(n)); g.emit("ScoreAwarded", l, n) }
func (g *Game) award(l string, n, b uint64) { g.score(l, n); g.Bonus.Add(number(b)) }
func (g *Game) effect(l string) bool        { return g.effectCore(l, false) }
func (g *Game) effectCore(l string, force bool) bool {
	l = strings.ToUpper(l)
	s, ok := programs.Effects[l]
	if !ok {
		panic("unknown STONES effect " + l)
	}
	accepted := s.Priority >= g.music.Priority
	if s.Jingle != "0" && (!g.Special && !g.inhEff || force || l == "LOSTBALL") {
		accepted = g.playJingle(s.Jingle)
	}
	g.award(l, s.Score, s.Bonus)
	if (force || accepted && !g.Special && !g.inhEff) && s.Matrix != "0" {
		g.beginMatrix(s.Matrix)
		return true
	}
	return false
}
func (g *Game) task(f func() bool) { tablelogic.Add(g.tasks[:], g.ids[:], &g.nextID, f) }
func (g *Game) wait(site string, n uint16, f func()) {
	g.task(func() bool {
		if !tablelogic.Wait(g.waits, site, n) {
			return false
		}
		f()
		return true
	})
}
func (g *Game) light(n int, on bool) { g.Lights[n] = on; g.lamp(n, on) }
func (g *Game) lamp(n int, on bool) {
	g.Lamps[n] = on
	g.applyLamp(&g.palette, n, on)
	g.emit("LampChanged", fmt.Sprint(n), 0)
}
func (g *Game) endFlash(n int) {
	for i, f := range g.flashes {
		if int(f.lamp) == n {
			g.flashes[i] = flash{}
			return
		}
	}
}
func (g *Game) flash(n, s int) {
	g.endFlash(n)
	for i := range g.flashes {
		if g.flashes[i].lamp == 0 {
			g.flashes[i] = flash{uint8(n), 0, uint8(s)}
			return
		}
	}
}
func (g *Game) off(n int) { g.endFlash(n); g.light(n, false) }
func (g *Game) flashTick() {
	for i := range g.flashes {
		f := &g.flashes[i]
		if f.lamp == 0 {
			continue
		}
		if f.counter == f.speed*2 {
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
func (g *Game) gate(name string, open bool) {
	for _, c := range programs.Gates {
		if c.Name == name {
			b := g.Display.StridedRecord(c.ClosedRef)
			if open {
				b = g.Display.StridedRecord(c.OpenedRef)
			}
			g.Physics.PatchMask(c.High, c.X, c.Y, c.Width, c.Height, b)
			g.emit("Gate", name, 0)
			return
		}
	}
	panic(name)
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
		g.Display.Flash() // MATRIX_BLINKOR precedes DO_TASKS.
		tablelogic.Run(g.tasks[:], g.ids[:])
		g.flashTick()
		if g.Phase == GameOver {
			return nil // Frontend continues DEMOMODE NODOT in this sync.
		}
		g.matrixTick()
		g.matrixPalette(false)
		return nil
	}
	if g.Physics.AllowFlip && (in.Left && !g.previous.Left || in.Right && !g.previous.Right) {
		g.sound("SFLIPPUPP")
		g.rotateGroups()
	}
	return g.Physics.Sync(in)
}
func (g *Game) afterTargets(in physics.Inputs) {
	warning, tilt := g.Physics.TiltInput(in.Tilt, g.inChute)
	if warning {
		g.playJingle("S_DANGER")
	}
	if tilt {
		g.playJingle("S_TILT")
		g.music.ReturnPosition = 52
		g.flashes = [64]flash{}
		for n := 1; n <= 44; n++ {
			g.light(n, false)
		}
		g.beginMatrix("TILTTS")
	}
	g.previous = in
	g.Display.Flash() // MATRIX_BLINKOR precedes DO_TASKS.
	tablelogic.Run(g.tasks[:], g.ids[:])
	gameplay.Spring(&g.Physics.SpringPosition, g.Physics.SpringValid, in, func(charge uint8) {
		g.Release(charge, uint8(g.clock))
	})
	g.flashTick()
}
func (g *Game) consume(e physics.Event) {
	switch e.Kind {
	case physics.EventBumperHit:
		g.score("BUMPER", 1000)
		g.sound("SBUMPER")
		g.addGhost()
	case physics.EventSlingshotHit:
		g.score("KICKER", 500)
		g.sound("SKICKER")
	case physics.EventTargetHit:
		g.touch(e.Object)
	case physics.EventDrain:
		g.drain()
	}
}
func (g *Game) Release(c, j uint8) {
	if g.Phase == Playing || g.Phase == NewBall {
		g.Physics.Release(c, j)
		if c > 0 && g.Playback != nil {
			g.Playback.SoundEffectVolume(g.soundSpec("SFJADER"), c*2)
		}
	}
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
	return true
}

func (g *Game) synced(n, s int, c uint8) {
	g.flash(n, s)
	for i := range g.flashes {
		if int(g.flashes[i].lamp) == n {
			g.flashes[i].counter = c
			return
		}
	}
}
