package frontend

import (
	"fmt"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/tablelogic"
	"strings"
	"testing"
)

// Read the logical matrix, never Frame's idle rendering fallback. The outgoing
// ball is placed below the drain only after the incoming idle checks finish.
type handoffFixture struct {
	hotseatSession
	physics *physics.Game
	display *presentation.Display
	ball    func() uint8
	score   func() tablelogic.Decimal
	seed    func([]uint64)
	drain   func()
	extra   func()
	newBall func() bool
	chute   func() bool
}

func handoffTable(t *testing.T, table, count int, scroll settings.ScrollMode) handoffFixture {
	t.Helper()
	r := hotseatRuntime(t)
	c := r.Model.Settings
	c.ScrollMode = scroll
	r.Model.Settings = c
	key(t, r.Model, Key(int(F1)+table-1))
	key(t, r.Model, Key(int(F1)+count-1))
	h := handoffFixture{hotseatSession: r.Model.Session.(hotseatSession)}
	switch table {
	case 1:
		g := r.Model.Session.(*partyland.Game)
		h.physics, h.display = g.Physics, g.Display
		h.ball = func() uint8 { return g.BallNumber }
		h.score = func() tablelogic.Decimal { return g.Score }
		h.seed = func(scores []uint64) {
			for i, v := range scores {
				g.Session.Players[i].Score = tablelogic.Number(v)
			}
			g.Score = tablelogic.Number(scores[0])
			g.Bonus = tablelogic.Decimal{}
			g.Session.SelectionOpen = false
		}
		h.drain = func() {
			g.ScoreChanged = true
			g.Physics.SetBall(160, 577, 0, 0, false)
			g.Physics.Ball.Hold = false
		}
		h.extra = func() { g.ExtraBalls = 1; g.Lights[51] = true }
		h.newBall = func() bool { return g.Phase == partyland.NewBall }
		h.chute = g.InChute
	case 2:
		g := r.Model.Session.(*speeddevils.Game)
		h.physics, h.display = g.Physics, g.Display
		h.ball = func() uint8 { return g.BallNumber }
		h.score = func() tablelogic.Decimal { return g.Score }
		h.seed = func(scores []uint64) {
			for i, v := range scores {
				g.Session.Players[i].Score = tablelogic.Number(v)
			}
			g.Score = tablelogic.Number(scores[0])
			g.Bonus = tablelogic.Decimal{}
			g.Session.SelectionOpen = false
		}
		h.drain = func() {

			g.Physics.SetBall(160, 577, 0, 0, false)
			g.Physics.Ball.Hold = false
		}
		h.extra = func() { g.Lights[55] = true }
		h.newBall = func() bool { return g.Phase == speeddevils.NewBall }
		h.chute = g.InChute
	case 3:
		g := r.Model.Session.(*gameshow.Game)
		h.physics, h.display = g.Physics, g.Display
		h.ball = func() uint8 { return g.BallNumber }
		h.score = func() tablelogic.Decimal { return g.Score }
		h.seed = func(scores []uint64) {
			for i, v := range scores {
				g.Session.Players[i].Score = tablelogic.Number(v)
			}
			g.Score = tablelogic.Number(scores[0])
			g.Bonus = tablelogic.Decimal{}
			g.Session.SelectionOpen = false
		}
		h.drain = func() {

			g.Physics.SetBall(160, 577, 0, 0, false)
			g.Physics.Ball.Hold = false
		}
		h.extra = func() { g.Lights[31] = true }
		h.newBall = func() bool { return g.Phase == gameshow.NewBall }
		h.chute = g.InChute
	case 4:
		g := r.Model.Session.(*stones.Game)
		h.physics, h.display = g.Physics, g.Display
		h.ball = func() uint8 { return g.BallNumber }
		h.score = func() tablelogic.Decimal { return g.Score }
		h.seed = func(scores []uint64) {
			for i, v := range scores {
				g.Session.Players[i].Score = tablelogic.Number(v)
			}
			g.Score = tablelogic.Number(scores[0])
			g.Bonus = tablelogic.Decimal{}
			g.Session.SelectionOpen = false
		}
		h.drain = func() {

			g.Physics.SetBall(160, 577, 0, 0, false)
			g.Physics.Ball.Hold = false
		}
		h.extra = func() { g.ExtraBalls = 1 }
		h.newBall = func() bool { return g.Phase == stones.NewBall }
		h.chute = g.InChute
	}
	return h
}

func sourceIdleDots(d *presentation.Display, score tablelogic.Decimal) [presentation.DotWidth * presentation.DotHeight]bool {
	// Independent expected drawing, using the source text and typed coordinates.
	// No synthesized zero or hand-written player/ball strings.
	expected := *d
	expected.Clear()
	for _, label := range []string{"PLAYERSTEXT", "BALLSTEXT"} {
		for _, c := range d.Content.Commands[d.Content.Labels["SHOWPLAYERSTS"]+1:] {
			if c.Op == "0" {
				break
			}
			if c.Arg(0) == label {
				x, y := presentation.PositionValue(c.Num(1))
				expected.Text(strings.TrimRight(d.SourceText(label), "\x00"), x, y, 5)
				break
			}
		}
	}
	expected.InvalidateScore() // Independent redraw starts with UPDAT_SCORE.
	expected.Score(score.String())
	return expected.Dots
}

func assertIdleHandoff(t *testing.T, h handoffFixture, player int, ball uint8, score tablelogic.Decimal) {
	t.Helper()
	if h.CurrentPlayer() != player || h.ball() != ball || h.score() != score {
		t.Fatalf("got P%d/B%d score %s; want P%d/B%d score %s", h.CurrentPlayer(), h.ball(), h.score(), player, ball, score)
	}
	if !h.display.On {
		t.Fatal("incoming logical dots exist but matrix illumination is OFF before launch")
	}
	if h.display.Dots != sourceIdleDots(h.display, score) {
		lit := 0
		for _, dot := range h.display.Dots {
			if dot {
				lit++
			}
		}
		t.Fatalf("logical matrix does not show incoming P%d/B%d/score before launch (%d lit dots)", player, ball, lit)
	}
}

// Check the composed user-visible frame too: logical dots alone can pass while
// the outgoing _MATRIXLGT,0 palette makes every dot indistinguishable.
func assertVisibleHandoff(t *testing.T, h handoffFixture) {
	t.Helper()
	frame := h.Frame()
	matrixY := h.physics.Settings.MatrixY()
	var lit, unlit int = -1, -1
	for i, on := range h.display.Dots {
		if on && lit < 0 {
			lit = i
		}
		if !on && unlit < 0 {
			unlit = i
		}
	}
	if lit < 0 || unlit < 0 {
		t.Fatal("matrix lacks score/player glyphs or background")
	}
	pixel := func(i int) [3]uint8 {
		c := frame.RGBAAt(2*(i%presentation.DotWidth), matrixY+2+2*(i/presentation.DotWidth))
		return [3]uint8{c.R, c.G, c.B}
	}
	if pixel(lit) == pixel(unlit) {
		t.Fatal("rendered handoff matrix is blank despite logical glyphs")
	}
}

func syncHandoff(t *testing.T, h handoffFixture) {
	t.Helper()
	if err := h.Sync(physics.Inputs{}); err != nil {
		t.Fatal(err)
	}
}

// Only the outgoing player is launched. Incoming presentation assertions run
// before any Down/Release input for that player, including before SETBALL.
func launchOutgoing(t *testing.T, h handoffFixture) {
	t.Helper()
	for charge := 0; charge < 32; charge++ {
		if err := h.Sync(physics.Inputs{Down: true}); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.Sync(physics.Inputs{Release: true}); err != nil {
		t.Fatal(err)
	}
	if h.physics.Ball.VY >= 0 {
		t.Fatal("outgoing plunger did not launch")
	}
}

func TestPlayerHandoffMatrixBeforeLaunch(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for _, count := range []int{1, 2, 3} {
			for _, scroll := range []settings.ScrollMode{settings.ScrollHard, settings.ScrollOff} {
				for _, zero := range []bool{false, true} {
					t.Run(fmt.Sprintf("table%d/players%d/scroll%d/zero%t", table, count, scroll, zero), func(t *testing.T) {
						h := handoffTable(t, table, count, scroll)
						scores := make([]uint64, count)
						for i := range scores {
							scores[i] = uint64(i+1) * 123450
						}
						if zero {
							if count > 1 {
								scores[1] = 0
							} else {
								scores[0] = 0
							}
						}
						h.seed(scores)
						// Finish first-ball selection without launching, then drive real loss and
						// bonus tasks. No newBall/changeBall calls can bypass the scheduler.
						for tick := 0; tick < 200; tick++ {
							syncHandoff(t, h)
						}
						expectedScores := append([]tablelogic.Decimal(nil), h.PlayerScores()...)
						for turn := 0; turn < count; turn++ {
							outgoing := h.CurrentPlayer()
							oldBall := h.ball()
							next, ball := outgoing%count+1, oldBall
							if outgoing == count {
								ball++
							}
							launchOutgoing(t, h)
							h.drain()
							ticks := 0
							for ; ticks < 10000; ticks++ {
								syncHandoff(t, h)
								if h.CurrentPlayer() == next && h.ball() == ball {
									break
								}
							}
							if ticks == 10000 {
								t.Fatal("loss/bonus did not commit handoff")
							}
							expectedScores[outgoing-1] = h.PlayerScores()[outgoing-1]
							// PLAND/STONES NEW_BALL_TASK waits 30; SDEV/SHOW wait 60.
							// WHEN_NEW_BALL_RESET then executes SHOWPLAYERSTS before
							// NODOT paints score. Identity alone does not repaint dots.
							if !h.physics.Ball.Hold {
								t.Fatal("incoming physical ball launched during handoff")
							}
							for wait := 0; wait < 710; wait++ { // ten seconds at the source 71 Hz clock
								syncHandoff(t, h)
								if wait >= 150 {
									assertIdleHandoff(t, h, next, ball, expectedScores[next-1])
								}
								if wait == 150 || wait == 709 {
									assertVisibleHandoff(t, h)
								}
								if (wait >= 150 && !h.chute()) || h.physics.SpringPosition != 0 {
									t.Fatal("idle wait launched gameplay")
								}
							}
							if h.newBall() || h.physics.Stopped {
								t.Fatal("normal source SETBALL delay did not finish")
							}
						}
					})
				}
			}
		}
	}
}

func TestExtraBallMatrixBeforeLaunch(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for _, count := range []int{1, 2, 3} {
			for _, scroll := range []settings.ScrollMode{settings.ScrollHard, settings.ScrollOff} {
				t.Run(fmt.Sprintf("table%d/players%d/scroll%d", table, count, scroll), func(t *testing.T) {
					h := handoffTable(t, table, count, scroll)
					scores := make([]uint64, count)
					scores[0] = 123450
					h.seed(scores)
					for tick := 0; tick < 200; tick++ {
						syncHandoff(t, h)
					}
					h.extra()
					launchOutgoing(t, h)
					h.drain()
					ticks := 0
					for ; ticks < 10000; ticks++ {
						syncHandoff(t, h)
						if h.newBall() {
							break
						}
					}
					if ticks == 10000 {
						t.Fatal("extra-ball new-ball task did not run")
					}
					score := h.score()
					// FANTASIE WHEN_NEW_BALL_RESET preserves PARTYFLASH on all tables;
					// each shoot-again program installs persistent PARTYRUT.
					expected := sourceIdleDots(h.display, score)
					if table >= 1 && table <= 4 {
						d := *h.display
						d.Clear()
						for _, c := range d.Content.Commands[d.Content.Labels["SHOOT_AGAIN_ONTS"]:] {
							if c.Op == "_PRINT13" {
								x, y := presentation.PositionValue(c.Num(1))
								d.Text(strings.TrimRight(d.SourceText(c.Arg(0)), "\x00"), x, y, 13)
								break
							}
						}
						expected = d.Dots
					}
					visibleTicks := 0
					for wait := 0; wait < 710; wait++ {
						if h.display.On {
							visibleTicks++
						}
						if h.CurrentPlayer() != 1 || h.ball() != 1 || h.score() != score {
							t.Fatal("extra ball rotated player/round or changed score")
						}
						if h.display.Dots != expected {
							t.Fatal("extra-ball logical matrix missing before launch")
						}
						nonblank := false
						for _, dot := range h.display.Dots {
							nonblank = nonblank || dot
						}
						if !nonblank || (wait >= 150 && !h.chute()) || h.physics.SpringPosition != 0 {
							t.Fatal("extra-ball wait blank or launched")
						}
						if h.display.On && (wait < 10 || wait == 150 || wait == 709) {
							assertVisibleHandoff(t, h)
						}
						syncHandoff(t, h)
					}
					if visibleTicks == 0 {
						t.Fatal("extra-ball message remained dark throughout idle wait")
					}
				})
			}
		}
	}
}
