package speeddevils

import (
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"

	"strconv"
	"strings"
)

type matrix struct {
	active                     bool
	next                       int
	op                         string
	remaining, textLeft        uint16
	animation                  animation
	frame, loops, frameTime    uint16
	scrollPhase                uint8
	countDigit                 int
	countUnit                  uint64
	countTimer                 int16
	held                       Decimal
	matchRemaining, matchTimer uint16
}

func (g *Game) beginMatrix(label string) { g.startMatrix(label, true) }

func (g *Game) startMatrix(label string, reset bool) {
	pos, ok := programs.Labels[strings.ToUpper(label)]
	if !ok {
		panic("missing SDEV matrix " + label)
	}
	if reset {
		g.Display.StartMatrix()
	} // DO_MATRIX, after effect acceptance.
	g.matrix = matrix{active: true, next: pos, scrollPhase: 8}
	g.emit("MatrixStarted", label, 0)
	g.matrixDispatch()
}
func (g *Game) matrixJump(label string) {
	pos, ok := programs.Labels[label]
	if !ok {
		panic("missing matrix branch " + label)
	}
	g.matrix.next = pos
}
func (g *Game) matrixValue(label string) uint64 {
	switch label {
	case "BONUSSIFFRORNA":
		return g.Bonus.Uint64()
	case "CYCLONECOUNTERBCD":
		return uint64(g.Miles)
	case "OR_TOTAL":
		return g.OffRoadTotal.Uint64()
	case "TM_TOTAL":
		return g.TurboTotal.Uint64()
	}
	panic("unknown SDEV matrix value " + label)
}

// Num returns source-resolved operand i. A missing value is an invariant
// violation and panics; it is never defaulted.
func (c command) Num(i int) int {
	v, ok := c.Nums[i]
	if !ok {
		panic("unresolved source operand " + c.Op + "[" + strconv.Itoa(i) + "]=" + c.Arg(i))
	}
	return v
}

// Arg returns operand i, or "" when the source command has no such operand.
func (c command) Arg(i int) string {
	if i < 0 || i >= len(c.Args) {
		return ""
	}
	return c.Args[i]
}

func (g *Game) matrixDispatch() {
	for branches := 0; branches < 32; branches++ {
		m := &g.matrix
		c := programs.Commands[m.next]
		m.next++
		m.op = c.Op
		g.Display.BeginResolved(c.Op, c.Args, c.Nums)
		m.remaining = 1
		g.emit("MatrixCommand", c.Op, 0)
		arg := func(i int) string { return c.Arg(i) }
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
		case "_CLEAR1":
			m.remaining = 1
		case "_CLEAR4":
			m.remaining = 5
		case "_CLEAR2":
			m.remaining = 17
		case "_CLEAR3":
			m.remaining = 81
		case "_WAIT":
			m.remaining = uint16(c.Num(0))
		case "_ANIMATION":
			a := g.Display.Content.Animations[arg(0)]
			m.animation = animation{Header: a.Header, Frames: a.Durations}
			m.frame = 0
			m.loops = m.animation.Header[1]
			m.frameTime = 1
		case "_SCROLL":
			m.textLeft = uint16(len(g.Display.Content.Texts[arg(0)])) - 21
		case "_WAITJINGLE", "_WAITJINGLE2":
		case "_COUNTDOWN":
			g.Display.StartCountdown(c.Num(0), c.Num(1))
			g.ModeTime = g.Display.CountdownRemaining()
		case "_COUNTDOWNCONTINUE":
			g.Display.ContinueCountdown()
			g.ModeTime = g.Display.CountdownRemaining()
			g.inhibitCountdown = false
		case "_JINGLE":
			g.Cue(arg(0))
		case "_LASTJINGLE":
			g.music.ReturnPosition = uint8(c.Num(0))
		case "_RULLGARDIN_UPP":
			m.remaining = uint16(16 - c.Num(1))
		case "_RULLGARDIN_NED":
			m.remaining = uint16(13 + c.Num(1))
		case "_BONUS_X_CALCS":
			g.Display.MutableText("BONUS_X_TEXT", 9)[8] = g.Multiplier + '7'
			old := g.Bonus
			for i := uint8(1); i < g.Multiplier; i++ {
				g.Bonus.Add(old)
			}
			g.emit("BonusMultiplied", c.Op, g.Bonus.Uint64())
		case "_CALC_CYCLO":
			g.Bonus.Add(number(uint64(g.Miles) * 100000))
			g.emit("BonusAdded", c.Op, uint64(g.Miles)*100000)
		case "_CALC_HAPPY":
			g.Bonus.Add(g.OffRoadTotal)
			g.emit("BonusAdded", c.Op, g.OffRoadTotal.Uint64())
		case "_CALC_MEGA":
			g.Bonus.Add(g.TurboTotal)
			g.emit("BonusAdded", c.Op, g.TurboTotal.Uint64())
		case "_FLORPA":
			m.held = g.Bonus
			m.countDigit = 11
			m.countUnit = 1
			m.countTimer = 0
		case "_WAIT_GAME_ON":
		case "_WAITIFMULTI":
			// FANTASIE single-player WAITIFMULTI explicitly loads SISA=2.
			m.remaining = 2
			if g.Session.PlayerCount > 1 {
				m.remaining = uint16(c.Num(0))
			}
		case "_KOLLA_XXBALL":
			if g.matchBall && g.Lights[55] {
				g.savePlayer() // VARS_2_P_STRUC precedes earned shoot-again.
				g.matrixJump("SHOOT_AGAIN_ONTS")
				continue
			}
			if g.Session.PlayerCount > 1 && g.matchBall && !(g.Lights[55]) {
				if g.nextMatch() {
					g.matrixJump("SHOOT_AGAIN_ONTS")
				} else {
					g.matrixJump("AFTER_XXBALLTS")
					continue
				}
				continue
			}
			if g.matchBall && !g.Lights[55] {
				g.matrixJump("AFTER_XXBALLTS")
				continue
			}
			continue
		case "_BEATEN_MATRIX":
			if g.beatHighScore() {
				g.matrixJump("BEATEN_BH_TS")
			}
			continue
		case "_DOBEATEN":
			g.light(55, true)
			g.emit("HighScoreExtraBall", c.Op, 0)
		case "_CHANGE_PLAYER":
			m.active = false
			if g.changeBall() {
				m.active = true
				continue
			}
			return
		case "_NEW_BALL2":
			g.wait("NEW_BALL_TASK", 60, g.newBall)
			continue
		case "_PARTYON":
			g.partyFlash = true
		case "_PARTYONN":
			// SDEV only substitutes the digit; _PARTYON owns PARTYFLASH.
			g.Display.MutableText("PARTY_ON_TEXT", 19)[18] = byte(g.Session.CurrentPlayer) + '7'
		case "_SHOOT_AGAIN_ONN":
			g.Display.MutableText("SHOOT_AGAIN_TEXT", 20)[19] = byte(g.Session.CurrentPlayer) + '7'
		case "_KNACKET":
			m.matchRemaining = 18
			m.matchTimer = 0
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
			slot := tablelogic.Add(g.tasks[:], g.ids[:], &g.nextID, func() bool { g.enterDemo(); return true })
			m.active = false
			if slot+1 < len(g.tasks) && g.tasks[slot+1] != nil {
				g.tasks[slot+1]() // HU_ tail call; this is not a task scan.
			}
			return

		case "_TURNOFFSPECIALMODE":
			g.Special = false
			g.music.Priority = 0
		case "_RETURN_OF_THE_EVIL_SUPERMODE":
			if g.inhibitCountdown {
				g.inhibitCountdown = false
				g.Turbo = true
				g.matrixJump("BACK_2_TURBOTS")
				continue
			}
			// INGA_KONSTIGHETER preserves NODOT/SISA=0 inherited
			// from the completed routine; NORMAL_END points at zero.
			m.active = false
			return
		case "_TURNONTURBO":
			g.off(9)
			g.startTurbo()
			// SDEV _turnonturbo tail-jumps TURBOTS even when GetTheGoal
			// only queues Goliat because another special mode is active.
			g.matrixJump("TURBOTS")
			continue
		case "_ADD50MILLION":
			g.score("SUPERJACK", 50_000_000)
		case "_SOUND_EFFECT":
			g.sound(arg(0))
		default:
			if !strings.HasPrefix(c.Op, "_PRINT") && c.Op != "_NUMBER" && c.Op != "_FLASHON" && c.Op != "_FLASHOFF" && c.Op != "_MATRIXLGT" && c.Op != "_PARTYON" && c.Op != "_PARTYONN" && c.Op != "_PARTYOFF" && c.Op != "_SETDECCOR" {
				panic("unported SDEV matrix command " + c.Op)
			}
		}
		return
	}
	panic("SDEV matrix branch cycle")
}
func (g *Game) matrixTick() {
	m := &g.matrix
	defer g.Display.FlushPrint(g.matrixNumber)
	if !m.active {
		panel := g.Display.TakeIdlePanel(g.inChute)
		if panel {
			g.startMatrix("SHOWPLAYERSTS", false) // NODOT calls DO_SPEC_MATRIX.
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
		done = g.Display.JingleDone(g.music.ReadyAnim, m.op == "_WAITJINGLE2")
	case "_COUNTDOWN", "_COUNTDOWNCONTINUE":
		done = g.Display.StepCountdown(g.Display.Argument(2), g.inhibitCountdown, func(seconds int) {
			if seconds != 0 {
				return
			}
			if g.Turbo {
				g.Cue("SJINGLE20")
			}
			if g.OffRoad {
				g.Cue("SJINGLE23")
			}
			g.Turbo, g.OffRoad = false, false
			g.music.Priority = 0
			g.music.JumpCount = 1
			g.music.ReturnPosition = 3
		})
		g.ModeTime = g.Display.CountdownRemaining()
	case "_KNACKET":
		if m.matchTimer == 0 {
			g.matchStart()
			g.Cue("S_ENDFIG")
			g.music.ReturnPosition = 55
			m.matchTimer = 13
			return
		}
		m.matchTimer--
		if m.matchTimer != 0 {
			return
		}
		m.matchTimer = 13
		g.Display.MatchStep(g.matchLast, g.clock%10)
		g.matchLast = g.clock % 10
		m.matchRemaining--
		done = m.matchRemaining == 0
		if done && g.anyMatch(uint8(g.matchLast)) {
			g.matchWin(uint8(g.matchLast))
			g.Cue("S_KNACKET")
			g.music.JumpCount = 1
			g.music.Priority = 0
			g.music.ReturnPosition = 55
		}
		g.emit("MatchStep", "KNACKRUT2", uint64(g.matchLast))
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
		m.remaining--
		done = m.remaining == 0
	}
	if done {
		nextOp := programs.Commands[m.next].Op
		g.Display.FinishRoutine(nextOp != "0")
		g.matrixDispatch()
	}
}
