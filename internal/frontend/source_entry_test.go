package frontend

import (
	"fmt"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/tablelogic"
	"testing"
)

func TestSourceScoreEntryVisitsAndRetainedPixels(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			r := hotseatRuntime(t)
			m := r.Model
			key(t, m, Key(int(F1)+table-1))
			key(t, m, F1)
			d := m.Session.(interface{ MatrixDisplay() *presentation.Display }).MatrixDisplay()
			for i := range d.Dots {
				d.Dots[i] = i%3 == 0
			}
			before := d.Dots
			high := m.Scores[table-1][0].Digits
			high.AddNumber(1)
			m.Mode = GameEnd
			m.sourceScoreStage = 1
			m.Counter = 1
			m.scoreQueue = []tablelogic.Decimal{{}, {}, high}
			m.scorePlayer = 0
			for player := 1; player <= 2; player++ {
				update(t, m, Input{})
				if m.Mode != GameEnd || m.scorePlayer != player {
					t.Fatal("SPINTSEL checked more than one player", player, m.scorePlayer, m.Mode)
				}
			}
			update(t, m, Input{})
			if m.Mode != Initials || m.scorePlayer != 3 {
				t.Fatal("third player qualification")
			}
			for y := 14; y < 16; y++ {
				for x := 0; x < 160; x++ {
					i := y*160 + x
					if d.Dots[i] != before[i] {
						t.Fatal("HAJJSKAR erased retained rows", x, y)
					}
				}
			}
			key(t, m, 30) // GET_IT_FROM_KEYBOARD clears SCAN_CODE on its own visit.
			if m.Entered != 0 {
				t.Fatal("setup accepted an old make")
			}
			update(t, m, Input{Keys: []Key{30, 48}})
			if m.Entered != 1 || m.Entry[0] != 'B' {
				t.Fatal("READ_KEYBOARDET did not consume one scan-code slot", m.Entry)
			}
			key(t, m, 30)
			key(t, m, 46)
			if m.Mode != EntryWait || m.Counter != 60 {
				t.Fatal("WAIT_A_LITTLE install")
			}
			for tick := 1; tick <= 58; tick++ {
				update(t, m, Input{})
				if tick < 58 && m.Mode != EntryWait {
					t.Fatal("early SPINTSEL restart", tick)
				}
				if tick == 30 {
					want := *d
					want.Dots = before
					want.Text(want.SourceText("STJAERNOR"), 0, 1, 13)
					if d.Dots != want.Dots {
						t.Fatal("counter30 source star writes")
					}
				}
			}
			if m.Mode != GameEnd || m.Counter != 1 || m.scorePlayer != 3 {
				t.Fatal("counter2 did not install SPINTSEL for its next visit")
			}
		})
	}
}

func TestSourceSelectionClosesOnActualLaunch(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			r := hotseatRuntime(t)
			m := r.Model
			key(t, m, Key(int(F1)+table-1))
			key(t, m, F1)
			g := m.Session.(interface {
				Session
				PlayerCount() int
				PlayerSelectionReady() bool
				InChute() bool
			})
			for i := 0; i < 100; i++ {
				update(t, m, Input{})
			}
			g.Release(32, 0)
			for i := 0; i < 500 && g.PlayerSelectionReady(); i++ {
				update(t, m, Input{})
			}
			if g.InChute() || g.PlayerSelectionReady() {
				t.Fatal("CLOSE1 did not close first-ball selection")
			}
			key(t, m, F8)
			if g.PlayerCount() != 1 {
				t.Fatal("launch accepted player-count change")
			}
		})
	}
}

func TestNonQualifyingGameShowsSourceGameOverPromptly(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			r := hotseatRuntime(t)
			m := r.Model
			m.Settings.Balls = 0
			key(t, m, Key(int(F1)+table-1))
			key(t, m, F1)
			var p *physics.Game
			var playing func() bool
			switch g := m.Session.(type) {
			case *partyland.Game:
				p = g.Physics
				g.BallNumber = 3
				playing = func() bool {
					g.ScoreChanged = true // Last ball has scored; PLAND otherwise gives a free replay.
					return g.Phase == partyland.Playing || g.Phase == partyland.NewBall
				}
			case *speeddevils.Game:
				p = g.Physics
				g.BallNumber = 3
				playing = func() bool { return g.Phase == speeddevils.Playing }
			case *gameshow.Game:
				p = g.Physics
				g.BallNumber = 3
				playing = func() bool { return g.Phase == gameshow.Playing }
			case *stones.Game:
				p = g.Physics
				g.BallNumber = 3
				playing = func() bool { return g.Phase == stones.Playing }
			}
			// Last ordinary ball is a test setup. Drains, bonus, match, optional
			// match-earned turn, qualification and demo handoff use real source ticks.
			for tick := 0; tick < 3000 && m.Mode != TableAttract; tick++ {
				if m.Mode == Playing && playing() && !p.Ball.Hold {
					p.SetBall(160, 577, 0, 0, false)
					p.Ball.Hold = false
				}
				update(t, m, Input{})
				if m.Mode == Initials {
					t.Fatal("unexpected qualification")
				}
			}
			if m.Mode != TableAttract || m.gameOverTimeline == nil {
				t.Fatal("source ending did not reach demo")
			}
			d := m.gameOverTimeline.Display()
			if d.CurrentOperation() != "_WAIT" {
				t.Fatal("missing first demo WAIT20", d.CurrentOperation())
			}
			found := false
			for visit := 0; visit < 40; visit++ {
				update(t, m, Input{})
				if d.CurrentOperation() == "_PRINT13" {
					found = true
					break
				}
			}
			if !found {
				t.Fatal("blank matrix persisted beyond source GAME OVER boundary")
			}
		})
	}
}
