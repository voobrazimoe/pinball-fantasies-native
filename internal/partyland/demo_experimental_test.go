//go:build demodev

package partyland

import (
	"bytes"
	"os"
	"pinballfantasies/internal/physics"
	"reflect"
	"testing"
)

func experimentalFixture(t *testing.T) *ExperimentalDemo {
	t.Helper()
	demo, canonical := os.Getenv("PF_10MIN_DEMO_DATA"), os.Getenv("PF_RUNTIME_DATA")
	if demo == "" || canonical == "" {
		t.Skip("owner-local demo and canonical inputs NOT AVAILABLE")
	}
	g, err := LoadExperimentalDemo(demo, canonical)
	if err != nil {
		t.Fatal(err)
	}
	return g
}
func TestExperimentalHostConsumesCandidateWithoutExtraTicks(t *testing.T) {
	live := experimentalFixture(t)
	silent := experimentalFixture(t)
	nonzero, moved := false, false
	first := live.core.game.Physics.Ball
	// Fixed short real keyboard smoke, not a reachability search.
	for n := 1; n <= 600; n++ {
		in := physics.Inputs{Down: n >= 2 && n < 26, Release: n == 26, Left: n%47 < 13, Right: n%53 < 11}
		e1, e2 := live.Sync(in), silent.core.sync(in, true)
		if (e1 == nil) != (e2 == nil) {
			t.Fatal("semantic/playback mismatch", n, e1, e2)
		}
		a, b := live.core, silent.core
		if a.calls != b.calls || a.counter != b.counter || a.game.Physics.Ball != b.game.Physics.Ball || a.game.Score != b.game.Score || a.game.Audio != b.game.Audio || a.game.clock != b.game.clock || !reflect.DeepEqual(a.game.waitCounters, b.game.waitCounters) {
			t.Fatal("extra tick or playback changed semantics", n)
		}
		if e1 != nil {
			t.Log(e1)
			break
		}
		if a.game.Physics.Ball != first {
			moved = true
		}
		if !bytes.Equal(live.PCM(), make([]byte, len(live.PCM()))) {
			nonzero = true
		}
	}
	if !moved || !nonzero || live.Calculations() < 50 {
		t.Fatal("no real movement/PCM/progression", live.Calculations(), moved, nonzero)
	}
	f := live.Frame()
	colored := false
	for i := 0; i < len(f.Pix); i += 4 {
		if f.Pix[i] != 0 || f.Pix[i+1] != 0 || f.Pix[i+2] != 0 {
			colored = true
			break
		}
	}
	if !colored {
		t.Fatal("black first frame")
	}
}
func TestExperimentalUnsupportedIsStickyAndDoesNotTick(t *testing.T) {
	d := experimentalFixture(t)
	if err := d.Sync(physics.Inputs{}); err != nil {
		t.Fatal(err)
	}
	before := d.core.game.Physics.Ball
	counter := d.core.counter
	err := d.Sync(physics.Inputs{Tilt: true})
	if err == nil || d.Diagnostic() == "" {
		t.Fatal("missing refusal")
	}
	for n := 0; n < 10; n++ {
		if d.Sync(physics.Inputs{}) != err {
			t.Fatal("nonsticky failure")
		}
	}
	if d.core.counter != counter || d.core.game.Physics.Ball != before || len(d.PCM()) != 0 {
		t.Fatal("continued after refusal")
	}
	if d.core.failure.Phase != "keyboard admission" {
		t.Fatal("stale diagnostic phase", d.core.failure)
	}
}
func TestExperimentalRejectsCommercialAndIncompleteDemo(t *testing.T) {
	canonical := os.Getenv("PF_RUNTIME_DATA")
	if canonical == "" {
		t.Skip("owner-local inputs NOT AVAILABLE")
	}
	if _, err := LoadExperimentalDemo(canonical, canonical); err == nil {
		t.Fatal("commercial input accepted as demo")
	}
	if _, err := LoadExperimentalDemo(t.TempDir(), canonical); err == nil {
		t.Fatal("missing demo accepted")
	}
}

// Structural expiry integration only: the seed is confined to this test. The
// interactive loader never exposes a timer/state/task override.
func TestExperimentalExpiryOutputStaysBlackThroughQuit(t *testing.T) {
	d := experimentalFixture(t)
	d.core.counter = 35997
	d.core.game.Physics.Ball.Hold = true
	sawFade, sawTrailingWait := false, false
	for n := 0; n < 5000 && !d.Done(); n++ {
		if err := d.Sync(physics.Inputs{}); err != nil {
			t.Fatal(err)
		}
		if d.core.fade.Active {
			sawFade = true
		}
		if d.core.fade.Active && d.core.game.matrix.op == "_WAIT" {
			sawTrailingWait = true
			if d.fadeVolume != 0 {
				t.Fatal("audio regained volume in trailing WAIT")
			}
			f := d.Frame()
			for i := 0; i < len(f.Pix); i += 4 {
				if f.Pix[i] != 0 || f.Pix[i+1] != 0 || f.Pix[i+2] != 0 {
					t.Fatal("FADE regained brightness in trailing WAIT")
				}
			}
		}
	}
	if !sawFade || !sawTrailingWait || !d.Done() || d.core.terminal.Status != 0 {
		t.Fatal("missing source expiry/QUIT", sawFade, sawTrailingWait, d.core.terminal)
	}
}
