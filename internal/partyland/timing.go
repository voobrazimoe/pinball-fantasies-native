package partyland

import (
	"encoding/json"
	"fmt"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"
	"strconv"
	"strings"
)

type matrixCommand struct {
	Op   string
	Args []string
	// Nums carries the extractor-resolved integer for every source numeric
	// operand, keyed by operand index. Args keeps the raw source text for
	// diagnostics and identity lookups.
	Nums map[int]int `json:"nums,omitempty"`
}

// Num returns source-resolved operand i. A missing value is an invariant
// violation and panics; it is never defaulted.
func (c matrixCommand) Num(i int) int {
	v, ok := c.Nums[i]
	if !ok {
		panic("unresolved source operand " + c.Op + "[" + strconv.Itoa(i) + "]=" + c.Arg(i))
	}
	return v
}

// Arg returns operand i, or "" when the source command has no such operand.
func (c matrixCommand) Arg(i int) string {
	if i < 0 || i >= len(c.Args) {
		return ""
	}
	return c.Args[i]
}

type jingleSpec = tablelogic.JingleSpec
type effectSpec struct {
	Jingle, Matrix string
	Priority       uint8
}
type cueSpec = tablelogic.CueSpec
type animationSpec struct{ Header, Frames []uint16 }
type timingContent struct {
	Effects    map[string]effectSpec
	Jingles    map[string]jingleSpec
	Cues       map[string]cueSpec
	Commands   []matrixCommand
	Labels     map[string]int
	Animations map[string]animationSpec
	Scrolls    map[string]uint16
}

var timing = func() timingContent {
	var v timingContent
	if err := json.Unmarshal([]byte(originalTimingData), &v); err != nil {
		panic(err)
	}
	return v
}()

type silentJingle struct {
	Position, ReturnPosition, Priority, LastPriority, JumpCount uint8
	ReadyAnim, ReadyLogic                                       bool
	Active                                                      bool
	entry                                                       uint8
	elapsed                                                     uint32 // 1/3550 second: 50 units per video sync, 71 per tracker tick.
	cue                                                         uint32
}

// matrixConsumer is an isolated extension owned by a source-program adapter.
// Ordinary table programs leave it nil.
type matrixConsumer interface {
	dispatch(matrixCommand) bool
	step(string, *uint16) bool
}

type matrixState struct {
	consumer                matrixConsumer
	sourceProgram           bool
	active                  bool
	next                    int
	op                      string
	remaining               uint16
	textLeft                uint16
	animation               animationSpec
	frame, loops, frameTime uint16
	countDigit              int
	countUnit               uint64
	countTimer              int16
	held                    Decimal
}

// WAITLIST is shared by instances of each original macro call site, not by task.
func (g *Game) waitAt(site string, n uint16, f func()) {
	g.task(func() bool {
		if !g.waitReady(site, n) {
			return false
		}
		f()
		return true
	})
}

// Background orders are 0..5 (spring and every main-music segment).
// Muting also applies to saved returns from jingles and rule-script overrides.
func (g *Game) muteMusicReturn() {
	if g.MusicOff && g.Audio.ReturnPosition <= 5 {
		g.Audio.ReturnPosition = 62
	}
}

// MOD control-flow remains the gameplay clock; optional sample rendering consumes
// the same rational interval without consulting a device.
// PF5 static SDR evidence establishes B-row entry, but exact game-relative phase
// remains unknown; the PF4.5 convention is retained. Deterministic
// callbacks run before electronics at the first 71 Hz boundary after the cue.
func (g *Game) audioTick() {
	g.muteMusicReturn()
	if g.Playback != nil {
		g.AudioPCM = g.Playback.Sync()
	}
	a := g.musicClock()
	a.Sync(timing.Cues, func(kind string, value uint64) {
		label := "JINGLE_HANDLER"
		if kind == "AudioComplete" {
			label = "JINGLEREADY"
		}
		g.emit(kind, label, value)
	})
	g.storeMusicClock(a)
}
func (g *Game) musicClock() tablelogic.MusicClock {
	a := g.Audio
	return tablelogic.MusicClock{Position: a.Position, ReturnPosition: a.ReturnPosition, Priority: a.Priority, LastPriority: a.LastPriority, JumpCount: a.JumpCount, Entry: a.entry, ReadyAnim: a.ReadyAnim, ReadyLogic: a.ReadyLogic, Active: a.Active, Elapsed: a.elapsed, Cue: a.cue}
}
func (g *Game) storeMusicClock(a tablelogic.MusicClock) {
	g.Audio = silentJingle{Position: a.Position, ReturnPosition: a.ReturnPosition, Priority: a.Priority, LastPriority: a.LastPriority, JumpCount: a.JumpCount, entry: a.Entry, ReadyAnim: a.ReadyAnim, ReadyLogic: a.ReadyLogic, Active: a.Active, elapsed: a.Elapsed, cue: a.Cue}
}
func (g *Game) playJingle(label string) bool {
	spec := g.Display.Jingle(label)
	if g.MusicOff && (strings.ToUpper(label) == "S_MAIN" || strings.ToUpper(label) == "S_SPRING") {
		spec.Position = 62
	}
	a := g.musicClock()
	accepted := a.Play(spec, 62, timing.Cues)
	g.storeMusicClock(a)
	if !accepted {
		g.emit("AudioRejected", label, uint64(spec.Priority))
		return false
	}
	g.muteMusicReturn()
	if g.Playback != nil {
		g.Playback.Force(int(spec.Position))
	}
	g.emit("Music", label, uint64(spec.Position))
	return true
}
func (g *Game) effectTiming(label string) bool {
	key := label
	if key == "MULTIBONUS" {
		key = fmt.Sprintf("M%d", g.Multiplier+1)
		if g.Multiplier == 1 {
			key = "M2"
		} else {
			key = fmt.Sprintf("M%d", g.Multiplier+2)
		}
	}
	spec, ok := timing.Effects[key]
	if !ok {
		panic("missing Party Land effect " + key)
	}
	accepted := true
	if spec.Jingle != "0" {
		if !g.inhibitEffect && (!g.special() || label == "LOSTBALL") {
			accepted = g.playJingle(spec.Jingle)
		}
	} else {
		accepted = spec.Priority >= g.Audio.Priority
	}
	return accepted && !g.inhibitEffect && !g.special() && spec.Matrix != "0"
}
func (g *Game) beginMatrix(label string) { g.startMatrix(label, true) }

func (g *Game) startMatrix(label string, reset bool) {
	label = strings.ToUpper(label)
	pos, ok := timing.Labels[label]
	source := false
	if !ok {
		pos, ok = g.Display.Content.Labels[label]
		source = true
	}
	if !ok {
		panic("missing Party Land matrix " + label)
	}
	g.emit("MatrixStarted", label, 0)
	if reset {
		g.Display.StartMatrix()
	} // DO_MATRIX, after effect acceptance.
	g.matrix = matrixState{active: true, next: pos, sourceProgram: source}
	g.matrixDispatch()
}
func (g *Game) matrixJump(label string) {
	pos, ok := timing.Labels[label]
	if g.matrix.sourceProgram {
		pos, ok = g.Display.Content.Labels[label]
	}
	if !ok {
		pos, ok = g.Display.Content.Labels[label]
		g.matrix.sourceProgram = ok
	}
	if !ok {
		panic("missing Party Land matrix branch " + label)
	}
	g.matrix.next = pos
}
func (g *Game) matrixDispatch() {
	// Original _JMP and conditional branches tail-call handlers in the same tick.
	for branches := 0; branches < 32; branches++ {
		m := &g.matrix
		var c matrixCommand
		if m.sourceProgram {
			src := g.Display.Content.Commands[m.next]
			c = matrixCommand{Op: src.Op, Args: src.Args, Nums: src.Nums}
		} else {
			c = timing.Commands[m.next]
		}
		m.next++
		m.op = c.Op
		g.Display.BeginResolved(c.Op, c.Args, c.Nums)
		m.remaining = 1
		g.emit("MatrixCommand", c.Op, 0)
		arg := func(i int) string {
			if i < 0 || i >= len(c.Args) {
				return ""
			}
			return c.Args[i]
		}
		switch c.Op {
		case "0":
			m.active = false
			g.emit("MatrixEnded", "NEXT_A", 0)
			return
		case "_JMP":
			g.matrixJump(arg(0))
			continue
		case "_JBCDZ":
			if g.matrixValue(arg(0)) == 0 {
				g.matrixJump(arg(1))
			}
			continue
		case "_JBONUSX1":
			if g.Multiplier == 1 {
				g.matrixJump(arg(0))
				continue
			}
		case "_CLEAR4":
			m.remaining = 5
		case "_CLEAR2":
			m.remaining = 17
		case "_CLEAR3":
			m.remaining = 81
		case "_WAIT":
			m.remaining = uint16(c.Num(0))
		case "_ANIMATION":
			var ok bool
			a, found := g.Display.Content.Animations[arg(0)]
			ok = found
			m.animation = animationSpec{Header: a.Header, Frames: a.Durations}
			if !ok {
				panic("missing animation " + arg(0))
			}
			m.frame = 0
			m.loops = m.animation.Header[1]
			m.frameTime = 1
		case "_SCROLL":
			text, ok := g.Display.Content.Texts[arg(0)]
			length := uint16(len(text))
			if !ok || length < 21 {
				panic("missing scroll " + arg(0))
			}
			m.textLeft = length - 21 // terminator is tested at SI+20 before SI increments.
		case "_WAITJINGLE", "_WAITJINGLE2": // polled by matrixTick.
		case "_EOSNURR":
			g.spinReady = true
		case "_TSEND":
			g.effectEnded = true
		case "_COUNTDOWN":
			g.modeIntro = false
			g.Display.StartCountdown(c.Num(0), c.Num(1))
			g.ModeTime = g.Display.CountdownRemaining()
		case "_COUNTDOWN2":
			if g.modeIntro {
				g.modeIntro = false
				g.Display.StartCountdown(c.Num(0), c.Num(1))
			} else {
				g.Display.ContinueCountdown()
			}
			g.ModeTime = g.Display.CountdownRemaining()
		case "_JINGLE":
			g.Audio.Priority = 0
			g.Audio.JumpCount = 1
			g.playJingle(arg(0))
		case "_LASTJINGLE":
			g.Audio.ReturnPosition = uint8(c.Num(0))
		case "_RULLGARDIN_UPP":
			m.remaining = uint16(16 - c.Num(1))
		case "_RULLGARDIN_NED":
			m.remaining = uint16(13 + c.Num(1))
		case "_BONUS_X_CALCS":
			g.Display.MutableText("BONUS_X_TEXT", 9)[8] = g.Multiplier + '7'
			original := g.Bonus
			for i := uint8(1); i < g.Multiplier; i++ {
				g.Bonus.Add(original)
			}
			g.emit("BonusMultiplied", c.Op, g.Bonus.Uint64())
		case "_CALC_CYCLO":
			n := uint64(g.Cyclones) * 100000
			g.Bonus.AddNumber(n)
			g.emit("BonusAdded", c.Op, n)
		case "_CALC_HAPPY":
			g.Bonus.Add(g.HappyTotal)
			g.emit("BonusAdded", c.Op, g.HappyTotal.Uint64())
		case "_CALC_MEGA":
			g.Bonus.Add(g.MegaTotal)
			g.emit("BonusAdded", c.Op, g.MegaTotal.Uint64())
		case "_FLORPA":
			m.held = g.Bonus
			m.countDigit = 11
			m.countUnit = 1
			m.countTimer = 0
		case "_WAIT_GAME_ON":
		case "_WAITIFMULTI":
			m.remaining = 2
			if g.Session.PlayerCount > 1 {
				m.remaining = uint16(c.Num(0))
			}
		case "_KOLLA_XXBALL":
			if g.matchBall && g.Lights[51] {
				g.savePlayer() // VARS_2_P_STRUC precedes earned shoot-again.
				g.ExtraBalls--
				g.matrixJump("SHOOT_AGAIN_ONTS")
				continue
			}
			if g.Session.PlayerCount > 1 && g.matchBall && !(g.Lights[51]) {
				if g.nextMatch() {
					g.matrixJump("SHOOT_AGAIN_ONTS")
				} else {
					g.matrixJump("AFTER_XXBALLTS")
					continue
				}
				continue
			}
			if g.matchBall && !g.Lights[51] {
				g.matrixJump("AFTER_XXBALLTS")
				continue
			}
			continue
		case "_BEATEN_MATRIX":
			if g.beatHighScore() {
				g.matrixJump("BEATEN_BH_TS")
			}
			continue
		case "_CHANGE_PLAYER":
			if g.HoldBonus {
				g.Bonus = m.held
			}
			if g.demoChangePlayer() {
				continue
			}
			m.active = false
			if g.changeBall() {
				m.active = true
				continue
			}
			return
		case "_NEW_BALL2":
			g.waitAt("NEW_BALL_TASK", 30, g.newBall)
			continue
		case "_PARTYON":
			g.partyFlash = true
		case "_PARTYONN":
			g.partyFlash = true
			g.Display.MutableText("PARTY_ON_TEXT", 19)[18] = byte(g.Session.CurrentPlayer) + '7'
		case "_SHOOT_AGAIN_ONN":
			g.Display.MutableText("SHOOT_AGAIN_TEXT", 20)[19] = byte(g.Session.CurrentPlayer) + '7'
		case "_KNACKET":
			g.matchRemaining = 22
			g.matchTimer = 0
		case "_CHECK_XXBALLS":
			if g.Session.PlayerCount > 1 {
				if g.selectMatch(uint8(g.matchLast)) {
					g.matrixJump("SHOOT_AGAIN_ONTS")
				} else {
					g.matrixJump("AFTER_XXBALLTS")
					continue
				}
				continue
			}
			if uint16(g.Score[10]) == g.matchLast {
				g.matchBall = true
				g.musicOK = true
				g.emit("ShootAgain", "XXBALLE", 0)
				g.matrixJump("SHOOT_AGAIN_ONTS")
				continue
			}
			g.matrixJump("AFTER_XXBALLTS")
			continue
		case "_CHECK_HIGH":
			// SPINTSEL_IN_HIGH is visited by the frontend, one player per sync.
			g.Phase = GameOver
		case "_2_DEMO_MODE":
			// DOADDTASK clobbers BX. HU_ reads the following TASKLIST
			// slot, not the adjacent matrix stream (whose long wait is
			// unreachable here). DUMRET leaves the installed NODOT idle.
			slot := tablelogic.Add(g.tasks[:], g.taskIDs[:], &g.nextTaskID, func() bool { g.enterDemo(); return true })
			m.active = false
			if slot+1 < len(g.tasks) && g.tasks[slot+1] != nil {
				g.tasks[slot+1]() // HU_ tail call; this is not a task scan.
			}
			return

		case "_DOBEATEN":
			g.ExtraBalls++
			g.light(51, true)
			g.emit("HighScoreExtraBall", "_DOBEATEN", 0)
		default:
			if m.consumer != nil && m.consumer.dispatch(c) {
				return
			}
			if !strings.HasPrefix(c.Op, "_PRINT") && c.Op != "_NUMBER" && c.Op != "_FLASHON" && c.Op != "_FLASHOFF" && c.Op != "_MATRIXLGT" && c.Op != "_PARTYON" && c.Op != "_PARTYONN" && c.Op != "_PARTYOFF" {
				panic("unported matrix command " + c.Op)
			}
		}
		return
	}
	panic("Party Land matrix branch cycle")
}
func (g *Game) matrixValue(label string) uint64 {
	switch label {
	case "BONUSSIFFRORNA":
		return g.Bonus.Uint64()
	case "CYCLONECOUNTERBCD":
		return uint64(g.Cyclones)
	case "HAPPY_HOUR_TOTAL":
		return g.HappyTotal.Uint64()
	case "MEGA_LAUGH_TOTAL":
		return g.MegaTotal.Uint64()
	}
	panic("unknown matrix value " + label)
}
func (g *Game) matrixTick() {
	m := &g.matrix
	if !g.matrixTimeLeft {
		return
	}
	defer g.Display.FlushPrint(g.matrixNumber)
	if !m.active {
		panel := g.Display.TakeIdlePanel(g.inChute)
		if panel {
			g.startMatrix(g.playerPanel(), false) // NODOT calls DO_SPEC_MATRIX.
		}
		if g.Phase == Playing && g.checkHighScore() {
			g.beginMatrix("BEATENTS")
		}
		g.Display.Score(g.Score.String()) // ONLY_SCORE also runs after installing the panel.
		if m.active {
			// NODOT returns SI=0 even after DO_SPEC_MATRIX installed a
			// routine. DO_THE_ANIMATIONS overwrites SISA with that zero,
			// sets DOTRUT=NODOT and dispatches NEXT_A in this same sync.
			g.matrixDispatch()
		}
		return
	}
	if m.op != "_ANIMATION" && m.op != "_SCROLL" && !strings.HasPrefix(m.op, "_PRINT") {
		g.Display.Visit(m.frame, 0, g.matrixNumber)
	}
	done := false
	switch m.op {
	case "_PARTYON":
		done = false
	case "_WAIT_GAME_ON":
		done = !g.Session.SelectionOpen // GONRUT tests ADDPLAYERS, not I_UTSKJUT.
	case "_ANIMATION":
		done = g.Display.StepAnimation(presentation.Animation{Header: m.animation.Header, Durations: m.animation.Frames, Offsets: g.Display.Content.Animations[g.Display.Argument(0)].Offsets}, &m.frame, &m.loops, &m.frameTime)
	case "_SCROLL":
		done = g.Display.StepScroll(&m.textLeft)
	case "_WAITJINGLE", "_WAITJINGLE2":
		done = g.Display.JingleDone(g.Audio.ReadyAnim, m.op == "_WAITJINGLE2")
	case "_COUNTDOWN", "_COUNTDOWN2":
		done = g.Display.StepCountdown(g.Display.Argument(2), false, func(seconds int) {
			if seconds == 2 || seconds == 0 {
				g.ModeTime = uint16(seconds*71 + 1)
				g.modeTick() // READ_SPECIAL_MODE_COUNTER, before PRINTTASK.
			}
		})
		g.ModeTime = g.Display.CountdownRemaining()
	case "_KNACKET":
		if g.matchTimer == 0 {
			g.matchStart()
			g.Audio.JumpCount = 1
			g.music("S_ENDFIG")
			g.Audio.ReturnPosition = 62
			g.matchTimer = 11
			return
		}
		g.matchTimer--
		if g.matchTimer != 0 {
			return
		}
		g.matchTimer = 11
		next := uint16(g.clock % 10)
		if next == g.matchLast {
			next = (next + 1) % 10
		}
		g.Display.MatchStep(g.matchLast, next)
		g.matchLast = next
		g.matchRemaining--
		g.emit("MatchStep", "KNACKRUT2", uint64(next))
		done = g.matchRemaining == 0
		if done {
			g.emit("Match", "_CHECK_XXBALLS", uint64(next))
			if g.anyMatch(uint8(next)) {
				g.matchWin(uint8(g.matchLast))
				g.Audio.Priority = 0
				g.music("S_KNACKET")
				g.Audio.JumpCount = 1
				g.Audio.Priority = 0
				g.Audio.ReturnPosition = 55
			}
		}
	case "_FLORPA":
		m.countTimer++
		if m.countTimer != 4 {
			return
		}
		m.countTimer = 0
		for m.countDigit >= 0 && g.Bonus[m.countDigit] == 0 {
			m.countDigit--
			m.countUnit *= 10
		}
		if m.countDigit < 0 {
			g.Display.FinishBonusField()
			done = true
			break
		}
		g.Bonus[m.countDigit]--
		g.score("DO_FLORPA", m.countUnit)
		g.sound("S_SCORELJUD")
		g.Display.Number(g.Bonus.String(), -32, 6, 8)
		g.Display.Score(g.Score.String())
		if g.Bonus.Uint64() == 0 {
			m.countTimer = -10
		}
	default:
		if m.consumer == nil || !m.consumer.step(m.op, &m.remaining) {
			m.remaining--
		}
		done = m.remaining == 0
	}
	if done {
		var nextOp string
		if m.sourceProgram {
			nextOp = g.Display.Content.Commands[m.next].Op
		} else {
			nextOp = timing.Commands[m.next].Op
		}
		g.Display.FinishRoutine(nextOp != "0")
		g.matrixDispatch()
	}
}

// START_DROP_WHEN_READY decrements the upper bound first, then latches readiness.
// A later effect may clear ReadyLogic; the latch must still release at the minimum.
func (g *Game) waitForAudioDrop(minimum, maximum uint16) {
	g.dropMinimum = minimum
	g.dropMaximum = maximum
	g.dropReady = false
	g.task(func() bool {
		g.dropMaximum--
		release := g.dropMaximum == 0
		if !release {
			if g.dropMinimum != 0 {
				g.dropMinimum--
			}
			if g.dropReady || g.special() || g.Audio.ReadyLogic {
				g.dropReady = true
				release = g.special() || g.dropMinimum == 0
			}
		}
		if !release {
			return false
		}
		g.Audio.ReadyLogic = false
		g.dropMinimum = 65535
		g.dropMaximum = 10
		g.emit("TaskReady", "START_DROP_WHEN_READY", 0)
		g.startDrop()
		return true
	})
}

// These two source tasks occupy real task slots, even though the camera's pixels
// are separate from rule state. Omitting them changes subsequent ADDTASK order.
func (g *Game) cameraDrop() {
	g.ScrollPosition = (g.Physics.Raster >> 4) - 33
	g.task(func() bool {
		g.ScrollPosition = (g.Physics.Raster >> 4) - 33
		g.task(func() bool {
			g.ScrollPosition -= 5
			if g.ScrollPosition < 0 {
				g.ScreenForce = 0
				return true
			}
			g.ScreenForce = g.ScrollPosition
			return false
		})
		return true
	})
}

func (g *Game) waitReady(site string, n uint16) bool {
	if !tablelogic.Wait(g.waitCounters, site, n) {
		return false
	}
	g.emit("TaskReady", site, 0)
	return true
}
func (g *Game) waitUntil(site string, n uint16, bypass func() bool, f func()) {
	g.task(func() bool {
		if !bypass() && !g.waitReady(site, n) {
			return false
		}
		f()
		return true
	})
}
