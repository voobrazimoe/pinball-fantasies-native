//go:build matrixdebug

package partyland

import (
	"pinballfantasies/internal/physics"
	"testing"
)

func TestDiagnosticSideLaneUsesPhysicalAwardAndDrain(t *testing.T) {
	for _, lane := range []int{-1, 1} {
		g := newTestGame(t)
		if g.TestSideLaneExtraBall(lane) {
			t.Fatal("placement must require leaving chute")
		}
		g.inChute = false
		g.Physics.SpringValid = false
		if !g.TestSideLaneExtraBall(lane) {
			t.Fatal("diagnostic placement rejected")
		}
		awarded, lost, newBall := false, -1, -1
		for tick := 0; tick < 2000; tick++ {
			if err := g.Sync(physics.Inputs{}); err != nil {
				t.Fatal(err)
			}
			for _, event := range g.Events {
				if event.Kind == "BallLost" {
					lost = tick
				}
				if event.Kind == "NewBall" {
					newBall = tick
				}
			}
			if g.ExtraBalls > 0 || g.Lights[51] {
				awarded = true
			}
			if newBall >= 0 {
				break
			}
		}
		if !awarded || lost < 0 || newBall < 0 {
			t.Fatalf("lane=%d XB=%t lost=%d newBall=%d phase=%d ball=%+v", lane, awarded, lost, newBall, g.Phase, g.Physics.Ball)
		}
		// BALL_LOSTTS's CLEARIT5 + PRINT13 + WAIT80 precede bonus.
		if newBall-lost < 86 {
			t.Fatal("physical reproducer bypassed BALL_LOSTTS WAIT80", lane, lost, newBall)
		}
		t.Logf("lane=%d drain=%d newball=%d", lane, lost, newBall)
	}
}
