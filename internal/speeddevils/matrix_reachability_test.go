package speeddevils

import (
	"fmt"
	"testing"
)

// TestSpeedDevilsReachableMatrixRegistered checks the runtime dispatch map
// (programs.Labels) against the source-derived reachable set.
func TestSpeedDevilsReachableMatrixRegistered(t *testing.T) {
	for _, label := range sourceReachableSpeedDevils {
		if _, ok := programs.Labels[label]; !ok {
			t.Errorf("source-reachable matrix program %s is not registered", label)
		}
	}
}

// TestSpeedDevilsDynamicEffectFamilies walks every selector value the source
// can produce: offroadLane picks M2..M9 from the first unlit multiplier lamp,
// and twoLoops picks SSCORE1..12 from the clamped speed value.
func TestSpeedDevilsDynamicEffectFamilies(t *testing.T) {
	for i := 2; i <= 9; i++ {
		g := testGame(t)
		g.effect(fmt.Sprintf("M%d", i)) // must resolve to M<i>TS
	}
	for i := 1; i <= 12; i++ {
		g := testGame(t)
		g.effect(fmt.Sprintf("SSCORE%d", i)) // must resolve to MILESTS
	}
}

func TestSpeedDevilsMatrixRuntimeSelectors(t *testing.T) {
	for n := 42; n <= 49; n++ {
		g := testGame(t)
		g.music.Priority = 0
		g.Lights[4] = true
		g.MBStock = 1
		for j := 42; j < n; j++ {
			g.Lights[j] = true
		}
		g.offroadLane()
		want := fmt.Sprintf("M%dTS", n-40)
		found := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == want {
				found = true
			}
		}
		if !found {
			t.Fatal("offroad selector", want)
		}
	}
	for speed := uint16(0); speed <= 12; speed++ {
		g := testGame(t)
		g.music.Priority = 0
		g.Speed = speed
		g.loop(false)
		found := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == "MILESTS" {
				found = true
			}
		}
		if !found {
			t.Fatal("skill score selector", speed)
		}
	}
}
