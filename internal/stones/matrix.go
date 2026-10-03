package stones

import (
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"

	"strings"
)

type matrix struct {
	active                            bool
	next                              int
	op                                string
	args                              []string
	nums                              map[int]int
	left                              int
	frame, loops, frameTime, textLeft uint16
	scroll                            uint8
	anim                              presentation.Animation
	digit                             int
	unit                              uint64
	countTimer                        int
	matchStep, matchTimer, matchLast  int
	towerRow                          int
}

func (g *Game) beginMatrix(label string) { g.startMatrix(label, true) }

func (g *Game) startMatrix(label string, reset bool) {
	p, ok := g.Display.Content.Labels[strings.ToUpper(label)]
	if !ok {
		panic("missing STONES program " + label)
	}
	g.matrixPalette(true)
	if reset {
		g.Display.StartMatrix()
	} // DO_MATRIX, after effect acceptance.
	g.matrix = matrix{active: true, next: p, scroll: 8}
	g.emit("MatrixStarted", label, 0)
	g.matrixDispatch()
}
func (g *Game) jump(label string) {
	p, ok := g.Display.Content.Labels[label]
	if !ok {
		panic(label)
	}
	g.matrix.next = p
}
func (g *Game) matrixDispatch() {
	for branches := 0; branches < 40; branches++ {
		m := &g.matrix
		c := g.Display.Content.Commands[m.next]
		m.next++
		m.op = c.Op
		m.args = c.Args
		m.nums = c.Nums
		m.left = 1
		g.Display.BeginCommand(c)
		arg := func(i int) string { return c.Arg(i) }
		num := func(i int) int { return c.Num(i) }
		switch c.Op {
		case "0":
			m.active = false
			return
		case "_JMP":
			g.jump(arg(0))
			continue
		case "_JBCDZ":
			if strings.Trim(g.matrixNumber(arg(0)), "0") == "" {
				g.jump(arg(1))
			}
			continue
		case "_JBONUSX1":
			if g.Multiplier == 1 {
				g.jump(arg(0))
				continue
			}
		case "_FLASHOFF", "_MATRIXLGT", "_FLASHON":
			g.Display.Visit(0, 0, g.matrixNumber)
			g.matrixPalette(true)
		case "_WAIT":
			m.left = presentation.WordWaitTicks(num(0))
		case "_WAITIFMULTI":
			m.left = 2
			if g.Session.PlayerCount > 1 {
				m.left = presentation.WordWaitTicks(num(0))
			}
		case "_CLEAR2":
			m.left = 17
		case "_CLEAR3":
			m.left = 81
		case "_ANIMATION":
			m.anim = g.Display.Content.Animations[arg(0)]
			m.frame = 0
			m.loops = m.anim.Header[1]
			m.frameTime = 1
		case "_SCROLL":
			m.textLeft = uint16(len(g.Display.Content.Texts[arg(0)]) - 21)
		case "_RULLGARDIN_UPP":
			m.left = 16 - num(1)
		case "_RULLGARDIN_NED":
			m.left = 13 + num(1)
		case "_JINGLE":
			g.Cue(arg(0))
		case "_LASTJINGLE":
			g.music.ReturnPosition = uint8(num(0))
		case "_WAITJINGLE", "_WAITJINGLE2":
		case "_TURNONSPECIALMODE":
			g.Special = true
		case "_TURNOFFSPECIALMODE":
			g.Special = false
			g.music.Priority = 0
		case "_TURNONOFFROADMODE":
			g.GhostHunt = true
		case "_TURNOFFOFFROADMODE":
			g.GhostHunt = false
		case "_TURNONTURBOMODE":
			g.Grim = true
		case "_TURNOFFTURBOMODE":
			g.Grim = false
		case "_COUNTDOWN":
			g.Display.StartCountdown(num(0), num(1))
			g.ModeTime = g.Display.CountdownRemaining()
		case "_COUNTDOWNCONTINUE":
			g.inhibit = false
			g.Display.ContinueCountdown()
			g.ModeTime = g.Display.CountdownRemaining()
		case "_TOWER":
			m.towerRow = 152
		case "_TOWEREND":
			g.ejectTask("TOWER")
		case "_WELLEND":
			g.ejectTask("WELL")
		case "_VAULTEND":
			if !g.fixTeleport {
				g.ejectTask("VAULT")
			}
			g.fixTeleport = false
		case "_TILTED":
			if g.captured == "TOWER" {
				g.task(func() bool { g.eject("TOWER"); return true })
			} else {
				g.ejectTask(g.captured)
			}
		case "_JACKEND":
			g.off(9)
			g.off(10)
			if !g.towerHuntOrig {
				g.updateTower()
			}
		case "_GRIMOFF":
			g.off(38)
		case "_BONUS_X_CALCS":
			g.Display.WriteBonusMultiplier(g.Multiplier)
			v := g.Bonus
			for i := uint8(1); i < g.Multiplier; i++ {
				g.Bonus.Add(v)
			}
		case "_CALC_CYCLO":
			g.Bonus.Add(number(uint64(g.Screams) * 100000))
		case "_CALC_HAPPY":
			g.Bonus.Add(g.GhostTotal)
		case "_CALC_MEGA":
			g.Bonus.Add(g.GrimTotal)
		case "_FLORPA":
			m.digit = 11
			m.unit = 1
			m.countTimer = 0
			g.heldBonus = g.Bonus
		case "_BEATEN_MATRIX":
			if g.beatHighScore() {
				g.jump("BEATEN_BH_TS")
			}
			continue
		case "_DOBEATEN":
			g.ExtraBalls++
		case "_KOLLA_XXBALL":
			if g.Session.PlayerCount > 1 && g.matchBall && !(g.ExtraBalls > 0) {
				if g.nextMatch() {
					g.jump("SHOOT_AGAIN_ONTS")
				} else {
					g.jump("AFTER_XXBALLSTS")
				}
				continue
			}
			if g.matchBall {
				if g.ExtraBalls > 0 {
					g.savePlayer() // VARS_2_P_STRUC precedes earned shoot-again.
					g.ExtraBalls--
					g.jump("SHOOT_AGAIN_ONTS")
				} else {
					g.jump("AFTER_XXBALLSTS")
				}
				continue
			}
			continue // XXBALLE=false uses HU_ on this sync.

		case "_CHANGE_PLAYER":
			m.active = false
			if g.changeBall() {
				m.active = true
				continue
			}
			return
		case "_NEW_BALL2":
			g.wait("NEW_BALL_TASK", 30, g.newBall)
			continue // HU_ tail-calls the next handler.
		case "_KNACKET":
			m.matchStep = 0
			m.matchTimer = 0 // _KNACKET installs KNACKRUT1; first visit initializes it.
		case "_CHECK_XXBALLS":
			if g.Session.PlayerCount > 1 {
				if g.selectMatch(uint8(m.matchLast)) {
					g.jump("SHOOT_AGAIN_ONTS")
				} else {
					g.jump("AFTER_XXBALLSTS")
				}
				continue
			}
			if uint8(m.matchLast) == g.Score[10] {
				g.matchBall = true
				g.jump("SHOOT_AGAIN_ONTS")
			} else {
				g.jump("AFTER_XXBALLSTS")
			}
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
		case "_PARTYONN":
			b := append([]byte(nil), g.Display.Content.Texts["SHOOTTHEBALLTEXT"]...)
			b[19] = byte(g.Session.CurrentPlayer) + '7'
			g.Display.Content.Texts["SHOOTTHEBALLTEXT"] = b
		case "_SHOOT_AGAIN_ONN":
			b := append([]byte(nil), g.Display.Content.Texts["SHOOT_AGAIN_TEXT"]...)
			b[19] = byte(g.Session.CurrentPlayer) + '7'
			g.Display.Content.Texts["SHOOT_AGAIN_TEXT"] = b
		case "_WAIT_GAME_ON":
		case "_SETDECCOR", "_SETLOOP", "_LOOP_", "_INIT_SCORE", "_SHOW_SCORE":
		default:
			if !strings.HasPrefix(c.Op, "_PRINT") && c.Op != "_FLASHON" && c.Op != "_FLASHOFF" && c.Op != "_MATRIXLGT" {
				panic("unported STONES command " + c.Op)
			}
		}
		return
	}
	panic("STONES program branch cycle")
}

var matchTimes = []int{22, 28, 25, 25, 22, 19, 18, 15, 13, 11, 9, 9, 8, 8, 7, 7, 6, 6, 6, 6, 6, 5, 5, 5, 5, 5, 5, 4, 4, 4, 4, 4, 4, 4, 3, 3}

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
	case "_WAITJINGLE", "_WAITJINGLE2":
		done = g.Display.JingleDone(g.music.ReadyAnim, m.op == "_WAITJINGLE2")
	case "_PARTYON":
		done = false
	case "_WAIT_GAME_ON":
		done = !g.Session.SelectionOpen // GONRUT tests ADDPLAYERS, not I_UTSKJUT.
	case "_TOWER":
		m.towerRow--
		g.Display.TowerWindow(0x4a1f0, m.towerRow)
		stop, ok := m.nums[0]
		if !ok {
			panic("STONES unresolved source operand _TOWER[0]=" + m.args[0])
		}
		done = m.towerRow == stop
	case "_COUNTDOWN", "_COUNTDOWNCONTINUE":
		done = g.Display.StepCountdown(g.Display.Argument(2), g.inhibit, func(seconds int) {
			if seconds != 0 {
				return
			}
			g.Cue("SJINGLE20")
			g.music.ReturnPosition = 3
			g.Grim, g.GhostHunt = false, false
			g.music.Priority = 0
			g.music.JumpCount = 1
		})
		g.ModeTime = g.Display.CountdownRemaining()
	case "_KNACKET":
		if m.matchTimer == 0 {
			m.matchLast = int(g.clock % 10)
			g.matchStart()
			g.Cue("S_ENDFIG")
			g.music.ReturnPosition = 52
			m.matchTimer = matchTimes[0]
			return
		}
		previous := m.matchLast
		m.matchTimer--
		if m.matchTimer == 0 {
			m.matchTimer = matchTimes[m.matchStep]
			m.matchLast = (m.matchLast + 9) % 10
			m.matchStep++
			done = m.matchStep == len(matchTimes)
		}
		// STONES KNACKRUT2 clears/queues its current digit on every visit,
		// including non-expiry visits, unlike the other three tables.
		g.Display.MatchStep(uint16(previous), uint16(m.matchLast))
		if done && g.anyMatch(uint8(m.matchLast)) {
			g.matchWin(uint8(m.matchLast))
			g.Cue("S_KNACKET")
			g.music.ReturnPosition = 52
			g.music.JumpCount = 1
			g.music.Priority = 0
		}
	case "_FLORPA":
		m.countTimer++
		if m.countTimer == 4 {
			m.countTimer = 0
			for m.digit >= 0 && g.Bonus[m.digit] == 0 {
				m.digit--
				m.unit *= 10
			}
			if m.digit < 0 {
				g.Display.FinishBonusField()
				done = true
			} else {
				g.Bonus[m.digit]--
				g.score("DO_FLORPA", m.unit)
				g.sound("S_SCORELJUD")
				g.Display.Number(g.Bonus.String(), -32, 6, 8)
				g.Display.Score(g.Score.String())
				if g.Bonus.Uint64() == 0 {
					m.countTimer = -10
				}
			}
		}
	default:
		m.left--
		done = m.left <= 0
	}
	if done {
		nextOp := g.Display.Content.Commands[m.next].Op
		g.Display.FinishRoutine(nextOp != "0")
		g.matrixDispatch()
	}
}
