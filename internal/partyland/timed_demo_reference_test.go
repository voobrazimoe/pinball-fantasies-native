//go:build demodev

package partyland

import (
	"bytes"
	"os"
	"testing"

	"pinballfantasies/internal/physics"
)

// The shipped demo layer is cross-checked against the bounded research
// candidate on every prefix the candidate admits. Private: both the demo and
// canonical A are required, and nothing is written.
func timedDemoReference(t *testing.T) *ExperimentalDemo {
	t.Helper()
	demo, a := os.Getenv("PF_10MIN_DEMO_DATA"), os.Getenv("PF_RUNTIME_DATA")
	if demo == "" || a == "" {
		t.Skip("set PF_10MIN_DEMO_DATA and PF_RUNTIME_DATA for the research reference")
	}
	d, err := LoadExperimentalDemo(demo, a)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func sameDemoBall(a, b physics.Ball) bool {
	a.Hold, b.Hold = false, false // source HOLDSTILL vs native capture hold
	return a == b
}

func TestTimedDemoMatchesReferenceNoInput(t *testing.T) {
	d, g := timedDemoReference(t), timedDemoFixture(t)
	for i := 1; !d.Done(); i++ {
		if err := d.Sync(physics.Inputs{}); err != nil {
			t.Fatal(i, err)
		}
		if err := g.Sync(physics.Inputs{}); err != nil {
			t.Fatal(err)
		}
		c, _, _ := g.Demo()
		if c != d.core.counter || !sameDemoBall(d.core.game.Physics.Ball, g.Physics.Ball) || d.core.game.Score != g.Score || d.core.game.matrix.op != g.matrix.op {
			t.Fatal("diverged at", i)
		}
		// The fade snapshots the live DAC at FADE; the candidate's is from load.
		if !g.timed.fade && (i%11 == 0 || i >= DemoThreshold) && !bytes.Equal(d.Frame().Pix, g.Frame().Pix) {
			t.Fatal("frame differs at", i)
		}
	}
	if !g.DemoFinished() || d.core.terminal.Calculation != 37047 {
		t.Fatal(g.DemoFinished(), d.core.terminal)
	}
}

// Played prefixes run until the candidate refuses an unproved transition; the
// shipped layer continues canonically from there.
func TestTimedDemoMatchesReferencePlayedPrefix(t *testing.T) {
	for _, period := range []int{300, 400, 500, 700} {
		d, g := timedDemoReference(t), timedDemoFixture(t)
		i := 1
		for ; i <= 5000; i++ {
			in := physics.Inputs{Down: i%period < 60, Release: i%period == 60, Left: (i/7)%5 == 0, Right: (i/11)%4 == 0}
			if d.Sync(in) != nil {
				break
			}
			if err := g.Sync(in); err != nil {
				t.Fatal(err)
			}
			if !sameDemoBall(d.core.game.Physics.Ball, g.Physics.Ball) || d.core.game.Score != g.Score || d.core.game.Phase != g.Phase {
				t.Fatalf("period %d diverged at %d", period, i)
			}
		}
		if i < 800 {
			t.Fatalf("period %d: reference prefix unexpectedly short (%d)", period, i)
		}
		for j := 0; j < 2000; j++ {
			if err := g.Sync(physics.Inputs{Left: j%40 < 5}); err != nil {
				t.Fatal(err)
			}
		}
	}
}
