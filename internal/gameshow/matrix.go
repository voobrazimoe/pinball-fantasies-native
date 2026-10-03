package gameshow

import (
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"

	"strings"
)

type matrix struct {
	active                                       bool
	next                                         int
	op                                           string
	args                                         []string
	nums                                         map[int]int
	remaining, textLeft, frame, loops, frameTime uint16
	anim                                         presentation.Animation
	scrollPhase                                  uint8
	digit                                        int
	unit                                         uint64
	countTimer                                   int16
	matchRemaining, matchTimer                   uint16
}

func (g *Game) beginMatrix(label string) { g.startMatrix(label, true) }

func (g *Game) startMatrix(label string, reset bool) {
	p, ok := g.Display.Content.Labels[strings.ToUpper(label)]
	if !ok {
		panic("missing SHOW matrix " + label)
	}
	if reset {
		g.Display.StartMatrix()
	} // DO_MATRIX, after effect acceptance.
	g.matrix = matrix{active: true, next: p, scrollPhase: 8}
	g.emit("MatrixStarted", label, 0)
	g.matrixDispatch()
}
func (g *Game) matrixJump(label string) {
	p, ok := g.Display.Content.Labels[label]
	if !ok {
		panic("missing SHOW branch " + label)
	}
	g.matrix.next = p
}
func (g *Game) matrixDispatch() {
	for branches := 0; branches < 32; branches++ {
		m := &g.matrix
		c := g.Display.Content.Commands[m.next]
		m.next++
		m.op = c.Op
		m.args = c.Args
		m.nums = c.Nums
		m.remaining = 1
		g.Display.BeginCommand(c)
		g.emit("MatrixCommand", c.Op, 0)
		arg := func(i int) string { return c.Arg(i) }
		num := func(i int) int { return c.Num(i) }
		switch c.Op {
		case "0":
			m.active = false
			return
		case "_JMP":
			g.matrixJump(arg(0))
			continue
		case "_JBCDZ":
			v := g.matrixNumber(arg(0))
			if strings.Trim(v, "0") == "" {
				g.matrixJump(arg(1))
			}
			continue
		case "_JBONUSX1":
			if g.Multiplier == 1 {
				g.matrixJump(arg(0))
				continue
			}
		case "_WAIT":
			m.remaining = uint16(num(0))
		case "_CLEAR2":
			m.remaining = 17
		case "_CLEAR3":
			m.remaining = 81
		case "_ANIMATION":
			m.anim = g.Display.Content.Animations[arg(0)]
			m.frame = 0
			m.loops = m.anim.Header[1]
			m.frameTime = 1
		case "_SCROLL":
			m.textLeft = uint16(len(g.Display.Content.Texts[arg(0)]) - 21)
		case "_RULLGARDIN_UPP":
			m.remaining = uint16(16 - num(1))
		case "_RULLGARDIN_NED":
			m.remaining = uint16(13 + num(1))
		case "_JINGLE":
			g.Cue(arg(0))
		case "_LASTJINGLE":
			g.music.ReturnPosition = uint8(num(0))
		case "_WAITJINGLE2":
		case "_COUNTDOWN":
			g.Display.StartCountdown(num(0), num(1))
			g.ModeTime = g.Display.CountdownRemaining()
		case "_TURNOFFSPECIALMODE":
			g.Special = false
			g.music.Priority = 0
			g.off(32)
		case "_LIGHTFLASH":
			g.flash(32, 3)
		case "_END_OF_SPIN":
			g.wait("END_OF_SPIN", 100, func() { g.score("END_OF_SPIN", g.spinScore); g.ejectWheel() })
		case "_BONUS_X_CALCS":
			g.Display.WriteBonusMultiplier(g.Multiplier)
			old := g.Bonus
			for i := uint8(1); i < g.Multiplier; i++ {
				g.Bonus.Add(old)
			}
			g.emit("BonusMultiplied", c.Op, g.Bonus.Uint64())
		case "_CALC_CYCLO":
			g.Bonus.Add(number(uint64(g.Skills) * 100000))
			g.emit("BonusAdded", c.Op, uint64(g.Skills)*100000)
		case "_CALC_HAPPY":
			g.Bonus.Add(g.MoneyTotal)
			g.emit("BonusAdded", c.Op, g.MoneyTotal.Uint64())
		case "_FLORPA":
			m.digit = 11
			m.unit = 1
			m.countTimer = 0
		case "_BEATEN_MATRIX":
			if g.beatHighScore() {
				g.matrixJump("BEATEN_BH_TS")
			}
			continue
		case "_DOBEATEN":
			g.light(31, true)
		case "_KOLLA_XXBALL":
			if g.matchBall && g.Lights[31] {
				g.savePlayer() // VARS_2_P_STRUC precedes earned shoot-again.
				g.matrixJump("SHOOT_AGAIN_ONTS")
				continue
			}
			if g.Session.PlayerCount > 1 && g.matchBall && !(g.Lights[31]) {

				if g.nextMatch() {
					g.matrixJump("SHOOT_AGAIN_ONTS")
				} else {
					g.matrixJump("AFTER_XXBALLTS")
				}
				continue
			}
			if g.matchBall && !g.Lights[31] {
				g.matrixJump("AFTER_XXBALLTS")
				continue
			}
			continue // XXBALLE=false uses HU_ on this sync.

		case "_WAITIFMULTI":
			m.remaining = 2
			if g.Session.PlayerCount > 1 {
				m.remaining = uint16(c.Num(0))
			}
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
		case "_KNACKET":
			m.matchRemaining = 15
			m.matchTimer = 0
		case "_CHECK_XXBALLS":
			if g.Session.PlayerCount > 1 {
				if g.selectMatch(uint8(g.matchLast)) {
					g.matrixJump("SHOOT_AGAIN_ONTS")
				} else {
					g.matrixJump("AFTER_XXBALLTS")
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
			g.Phase = GameOver // Frontend visits SPINTSEL_IN_HIGH on subsequent syncs.
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
		case "_SOUND_EFFECT":
			g.sound(arg(0))
		case "_PARTYON":
			g.partyFlash = true
		case "_PARTYOFF":
			// Source only installs WAITRUT1; CLOSE1 owns PARTYFLASH.
		case "_PARTYONN": // Dynamic player digit encoded for PRINT_TEXT.
			for _, label := range []string{"PARTY_ON_TEXT", "SHOOTTHEBALLTEXT"} {
				b := append([]byte(nil), g.Display.Content.Texts[label]...)
				if len(b) > 18 {
					ix := 18
					if label == "SHOOTTHEBALLTEXT" {
						ix = 19
					}
					b[ix] = byte(g.Session.CurrentPlayer) + '7'
					g.Display.Content.Texts[label] = b
				}
			}
		case "_SHOOT_AGAIN_ONN":
			b := append([]byte(nil), g.Display.Content.Texts["SHOOT_AGAIN_TEXT"]...)
			if len(b) > 19 {
				b[19] = byte(g.Session.CurrentPlayer) + '7'
				g.Display.Content.Texts["SHOOT_AGAIN_TEXT"] = b
			}
		case "_WAIT_GAME_ON":
		default:
			if !strings.HasPrefix(c.Op, "_PRINT") && c.Op != "_NUMBER" && c.Op != "_FLASHON" && c.Op != "_FLASHOFF" && c.Op != "_MATRIXLGT" && c.Op != "_PARTYON" && c.Op != "_PARTYOFF" && c.Op != "_SETDECCOR" {
				panic("unported SHOW matrix " + c.Op)
			}
		}
		return
	}
	panic("SHOW matrix branch cycle")
}
func (g *Game) matrixTick() {
	m := &g.matrix
	defer g.Display.FlushPrint(g.matrixNumber)
	if !m.active {
		panel := g.Display.TakeIdlePanel(g.inChute)
		if panel {
			g.startMatrix("SHOWPLAYERSTS", false) // NODOT calls DO_SPEC_MATRIX.
		}
		if g.Phase == Playing && !g.inChute && !g.Special && g.beatHighScore() {
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
	case "_ANIMATION":
		done = g.Display.StepAnimation(m.anim, &m.frame, &m.loops, &m.frameTime)
	case "_SCROLL":
		done = g.Display.StepScroll(&m.textLeft)
	case "_WAITJINGLE2":
		done = g.Display.JingleDone(g.music.ReadyAnim, m.op == "_WAITJINGLE2")
	case "_PARTYON":
		done = false
	case "_WAIT_GAME_ON":
		done = !g.Session.SelectionOpen // GONRUT tests ADDPLAYERS, not I_UTSKJUT.
	case "_COUNTDOWN":
		done = g.Display.StepCountdown(g.Display.Argument(2), false, func(seconds int) {
			if seconds != 0 {
				return
			}
			g.Cue("SJINGLE10")
			g.music.ReturnPosition = 3
			g.MoneyMania = false
			g.music.Priority = 0
			g.music.JumpCount = 1
		})
		g.ModeTime = g.Display.CountdownRemaining()
	case "_KNACKET":
		if m.matchTimer == 0 {
			g.matchStart()
			g.Cue("S_ENDFIG")
			g.music.ReturnPosition = 62
			m.matchTimer = 14
			return
		}
		m.matchTimer--
		if m.matchTimer > 0 {
			return
		}
		m.matchTimer = 14
		next := g.clock % 10
		if next == g.matchLast {
			next = 9 - next
		}
		g.Display.MatchStep(g.matchLast, next)
		g.matchLast = next
		m.matchRemaining--
		done = m.matchRemaining == 0
		g.emit("MatchStep", "KNACKRUT2", uint64(next))
		if done && g.anyMatch(uint8(next)) {
			g.matchWin(uint8(g.matchLast))
			g.Cue("S_KNACKET")
			g.music.ReturnPosition = 55
			g.music.JumpCount = 1
		}
	case "_FLORPA":
		m.countTimer++
		if m.countTimer != 4 {
			return
		}
		m.countTimer = 0
		for m.digit >= 0 && g.Bonus[m.digit] == 0 {
			m.digit--
			m.unit *= 10
		}
		if m.digit < 0 {
			g.Display.FinishBonusField()
			done = true
			break
		}
		g.Bonus[m.digit]--
		g.score("DO_FLORPA", m.unit)
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
		nextOp := g.Display.Content.Commands[m.next].Op
		g.Display.FinishRoutine(nextOp != "0")
		g.matrixDispatch()
	}
}
