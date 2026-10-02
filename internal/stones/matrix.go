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

func (g *Game) beginMatrix(label string) {
	p, ok := g.Display.Content.Labels[strings.ToUpper(label)]
	if !ok {
		panic("missing STONES program " + label)
	}
	g.Display.Begin("_FLASHOFF", nil)
	g.Display.Visit(0, 0, g.matrixNumber)
	g.matrixPalette(true)
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
		if strings.HasPrefix(c.Op, "_PRINT") {
			g.Display.Visit(0, 0, g.matrixNumber)
		}
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
			m.left = num(0)
		case "_WAITIFMULTI":
			m.left = 2
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
			g.ModeTime = uint16((num(0)*10+num(1))*71 + 1)
		case "_COUNTDOWNCONTINUE":
			g.inhibit = false
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
			if !g.TowerHunt {
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
				if g.holdBonus {
					g.Bonus = g.heldBonus
				}
				if g.nextMatch() {
					g.jump("SHOOT_AGAIN_ONTS")
				} else {
					g.jump("AFTER_XXBALLSTS")
				}
				continue
			}
			if g.matchBall {
				if g.ExtraBalls > 0 {
					g.ExtraBalls--
					g.jump("SHOOT_AGAIN_ONTS")
				} else {
					g.jump("AFTER_XXBALLSTS")
				}
				continue
			}
		case "_CHANGE_PLAYER":
			m.active = false // The outgoing loss program must not clear the incoming idle panel.
			g.changeBall()
			return
		case "_NEW_BALL2":
			g.wait("NEW_BALL_TASK", 30, g.newBall)
		case "_KNACKET":
			m.matchStep = 0
			m.matchTimer = 22
			m.matchLast = int(g.clock % 10)
			g.matchStart()
			g.Cue("S_ENDFIG")
			g.music.ReturnPosition = 52
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
		case "_2_DEMO_MODE":
			g.endGame()
			return
		case "_SOUND_EFFECT":
			g.sound(arg(0))
		case "_PARTYON":
			g.partyFlash = true
		case "_PARTYOFF":
			g.partyFlash = false
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
	if !m.active {
		if g.Phase == Playing && !g.inChute && !g.Special && g.beatHighScore() {
			g.beginMatrix("BEATENTS")
		}
		return
	}
	g.Display.Visit(m.frame, 2, g.matrixNumber)
	done := false
	switch m.op {
	case "_ANIMATION":
		done = tablelogic.Animation(m.anim.Header, m.anim.Durations, &m.frame, &m.loops, &m.frameTime)
	case "_SCROLL":
		done = tablelogic.Scroll(&m.textLeft, &m.scroll)
	case "_WAITJINGLE", "_WAITJINGLE2":
		done = g.music.ReadyAnim
	case "_PARTYON":
		done = false
	case "_WAIT_GAME_ON":
		done = !g.inChute
	case "_TOWER":
		m.towerRow--
		g.Display.TowerWindow(0x4a1f0, m.towerRow)
		stop, ok := m.nums[0]
		if !ok {
			panic("STONES unresolved source operand _TOWER[0]=" + m.args[0])
		}
		done = m.towerRow == stop
	case "_COUNTDOWN", "_COUNTDOWNCONTINUE":
		if !g.inhibit && g.ModeTime > 0 {
			g.ModeTime--
		}
		g.Display.Countdown(g.matrixNumber(m.args[2]), int(g.ModeTime/71))
		done = g.ModeTime == 0
		if done {
			g.Cue("SJINGLE20")
			g.music.ReturnPosition = 3
			g.Grim = false
			g.GhostHunt = false
			g.music.JumpCount = 1
		}
	case "_KNACKET":
		m.matchTimer--
		if m.matchTimer == 0 {
			m.matchTimer = matchTimes[m.matchStep]
			previous := m.matchLast
			m.matchLast = (m.matchLast + 9) % 10
			g.Display.MatchStep(uint16(previous), uint16(m.matchLast))
			m.matchStep++
			done = m.matchStep == len(matchTimes)
			if done && g.anyMatch(uint8(m.matchLast)) {
				g.matchWin(uint8(m.matchLast))
				g.Cue("S_KNACKET")
				g.music.ReturnPosition = 52
				g.music.JumpCount = 1
			}
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
		g.matrixDispatch()
	}
}
