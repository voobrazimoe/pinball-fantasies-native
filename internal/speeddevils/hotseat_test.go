package speeddevils

import (
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/tablelogic"
	"testing"
)

func TestHotseatRoundRotation(t *testing.T) {
	for _, count := range []int{1, 2, 3, 8} {
		for _, balls := range []uint8{3, 5} {
			g := testGame(t)
			g.totalBalls = balls
			g.StartPlayers(count)
			for ball := uint8(1); ball <= balls; ball++ {
				for player := 1; player <= count; player++ {
					if g.CurrentPlayer() != player || g.BallNumber != ball {
						t.Fatalf("count=%d got P%d B%d want P%d B%d", count, g.CurrentPlayer(), g.BallNumber, player, ball)
					}
					g.Score.AddNumber(uint64(player) * 1000)
					g.changeBall()
					if ball == balls && player == count {
						if g.BallNumber != balls+1 {
							t.Fatal("last round did not end")
						}
					} else {
						g.newBall()
						if g.Phase == GameOver {
							t.Fatal("early game over")
						}
					}
				}
			}
			for i, score := range g.PlayerScores() {
				if score.Uint64() != uint64(i+1)*1000*uint64(balls) {
					t.Fatalf("P%d score=%s", i+1, score)
				}
			}
		}
	}
}
func TestHotseatStateIsolation(t *testing.T) {
	for _, scroll := range []settings.ScrollMode{settings.ScrollHard, settings.ScrollMedium, settings.ScrollSoft, settings.ScrollOff} {
		g := testGame(t)
		c := settings.Legacy()
		c.ScrollMode = scroll
		g.Configure(c)
		g.StartPlayers(2)
		g.Miles = 7
		g.Speed = 3
		g.Lights[10] = true
		g.Turbo = true
		g.MBStock = 9
		g.Score = tablelogic.Number(1234560)
		g.Jackpot = tablelogic.Number(7777770)
		g.changeBall()
		g.newBall()
		if g.CurrentPlayer() != 2 || g.Miles != 0 || g.Speed != 0 || g.Lights[10] || g.Score.Uint64() != 0 || g.Turbo || g.MBStock != 0 {
			t.Fatal("P1 state leaked into P2")
		}
		if g.Jackpot.Uint64() != 7777770 {
			t.Fatal("shared jackpot was swapped")
		}
		g.Score = tablelogic.Number(890)
		g.changeBall()
		g.newBall()
		if g.CurrentPlayer() != 1 || g.Miles != 7 || g.Speed != 3 || !g.Lights[10] || g.Score.Uint64() != 1234560 || g.Turbo || g.MBStock != 0 {
			t.Fatal("P1 persistent progress was not restored")
		}
		g.Physics.Ball.Hold = true
		g.Frame()
		if g.Physics.Ball.Lost || g.Physics.Stopped {
			t.Fatal("transient physical ball leaked")
		}
		if err := g.Sync(physics.Inputs{}); err != nil {
			t.Fatal(err)
		}
	}
}
func TestHotseatExtraBallAndMatch(t *testing.T) {
	g := testGame(t)
	g.StartPlayers(3)
	g.Lights[55] = true
	g.changeBall()
	if g.CurrentPlayer() != 1 || g.BallNumber != 1 {
		t.Fatal("extra ball advanced round/player")
	}
	g.newBall()
	g.changeBall()
	g.newBall()
	if g.CurrentPlayer() != 2 || g.BallNumber != 1 {
		t.Fatal("rotation after extra ball")
	}
	// Source match checks all saved scores in player order, including more than
	// one matching score; each winner is consumed once.
	g.Session.Players[0].Score = tablelogic.Number(120)
	g.Session.Players[1].Score = tablelogic.Number(130)
	g.Session.Players[2].Score = tablelogic.Number(220)
	g.Score = tablelogic.Number(130)
	if !g.selectMatch(2) || g.CurrentPlayer() != 1 {
		t.Fatal("first match player")
	}
	if !g.nextMatch() || g.CurrentPlayer() != 3 {
		t.Fatal("next match player")
	}
	if g.nextMatch() {
		t.Fatal("match repeated a winner")
	}
}
func TestHotseatResetFactorySlots(t *testing.T) {
	old := testGame(t)
	old.StartPlayers(8)
	old.Score = tablelogic.Number(9999990)
	old.savePlayer()
	g := testGame(t)
	g.StartPlayers(1)
	if g.CurrentPlayer() != 1 || g.PlayerCount() != 1 || g.BallNumber != 1 {
		t.Fatal("new game identity")
	}
	for _, s := range g.Session.Players {
		if s.Score.Uint64() != 0 {
			t.Fatal("old score survived")
		}
	}
	g.StartPlayers(8)
	for _, s := range g.Session.Players {
		if s.Score.Uint64() != 0 {
			t.Fatal("factory slot not cleared")
		}
	}
}

func TestHotseatDrainBonusAndTilt(t *testing.T) {
	for _, tilted := range []bool{false, true} {
		g := testGame(t)
		g.StartPlayers(2)
		g.Score = tablelogic.Number(10000)
		g.Bonus = tablelogic.Number(1000)

		if tilted {
			g.Physics.Stopped = true
		}
		g.drain()
		for i := 0; i < 5000 && (g.CurrentPlayer() != 2 || g.Phase != Playing); i++ {
			if err := g.Sync(physics.Inputs{}); err != nil {
				t.Fatal(err)
			}
		}
		if g.CurrentPlayer() != 2 || g.BallNumber != 1 || g.Phase != Playing {
			t.Fatal("drain did not produce P2 B1", g.CurrentPlayer(), g.BallNumber, g.Phase)
		}
		if g.Score.Uint64() != 0 || g.PlayerScores()[0].Uint64() < 10000 || g.Physics.Stopped {
			t.Fatal("outgoing bonus/tilt leaked")
		}
	}
}

func TestHotseatCompleteGameWithMatchTurns(t *testing.T) {
	g := testGame(t)
	g.StartPlayers(3)
	normal := 0
	draining := false
	for tick := 0; tick < 30000 && g.Phase != GameOver; tick++ {
		if g.Phase == Playing && !draining {
			if !g.matchBall {
				if g.CurrentPlayer() != normal%3+1 || int(g.BallNumber) != normal/3+1 {
					t.Fatal("normal sequence before game over", normal, g.CurrentPlayer(), g.BallNumber)
				}
				normal++
			} else if normal != 9 {
				t.Fatal("match began before all required turns")
			}
			g.Score.AddNumber(1000)

			g.drain()
			draining = true
		}
		if g.Phase == NewBall {
			draining = false
		}
		if err := g.Sync(physics.Inputs{}); err != nil {
			t.Fatal(err)
		}
	}
	if g.Phase != GameOver || normal != 9 {
		t.Fatal("game over incomplete", normal, g.Phase)
	}
}
