package partyland

import (
	"bytes"
	"testing"
)

func TestHighScoreBranchExtraBall(t *testing.T) {
	for _, duringDrain := range []bool{false, true} {
		g := newTestGame(t)
		quiet(g)
		g.Physics.SpringValid = false
		g.SetHighScore(Number(50_000_000))
		g.Score = Number(50_000_001)
		g.ScoreChanged = true
		if duringDrain {
			g.drain()
			for i := 0; i < 2000 && g.ExtraBalls == 0; i++ {
				ticks(t, g, 1)
			}
		} else {
			ticks(t, g, 2)
		}
		if !g.alreadyBeaten || g.ExtraBalls != 1 || !g.Lights[51] {
			t.Fatalf("high-score award drain=%v already=%v balls=%d", duringDrain, g.alreadyBeaten, g.ExtraBalls)
		}
		old := g.ExtraBalls
		g.checkHighScore()
		if g.ExtraBalls != old {
			t.Fatal("top score awarded twice")
		}
	}
}
func TestHighScoreStrictAndOptIn(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.Physics.SpringValid = false
	g.Score = Number(50_000_000)
	g.SetHighScore(Number(50_000_000))
	if g.checkHighScore() {
		t.Fatal("tie beats top score")
	}
	g.Score = Number(50_000_001)
	g.Physics.SpringValid = true
	if g.checkHighScore() {
		t.Fatal("chute triggered beat")
	}
	g.Physics.SpringValid = false
	if !g.checkHighScore() {
		t.Fatal("greater score did not beat")
	}
	g = newTestGame(t)
	g.Score = Number(999_999_999_999)
	if g.checkHighScore() {
		t.Fatal("direct PF4 branch enabled")
	}
}
func TestAttractRenderDoesNotMutateSession(t *testing.T) {
	g := newTestGame(t)
	a := g.AttractFrame(0)
	b := g.AttractFrame(100)
	if bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("attract camera did not scroll")
	}
	if g.Tick != 0 || g.Physics.Syncs != 0 || g.Physics.Raster != (259+33)*16 {
		t.Fatal("attract render mutated game")
	}
}
