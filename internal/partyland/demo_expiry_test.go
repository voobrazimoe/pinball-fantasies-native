//go:build dmoimpl1

package partyland

import (
	"fmt"
	"os"
	"pinballfantasies/internal/physics"

	"pinballfantasies/internal/tablelogic"
	"reflect"
	"testing"
)

func expiryFixture(t *testing.T) *demoCore {
	t.Helper()
	d := connectedQuiet(t).demoCore
	// Malformed-consumer tests must own the mutable metadata map too.
	fonts := make(map[string]int, len(d.game.Display.Content.Fonts))
	for k, v := range d.game.Display.Content.Fonts {
		fonts[k] = v
	}
	d.game.Display.Content.Fonts = fonts
	b := pinnedDemoBytes(t)
	a, e := os.ReadFile("../../TABLE1.PRG")
	if e != nil {
		t.Fatal(e)
	}
	demoOK(t, d.loadExpiry(b, a))
	d.game.tasks = [50]func() bool{}
	d.counter = 35997
	return d
}
func TestDemoExpiryComplete(t *testing.T) {
	for _, score := range []uint64{0, 1234567890} {
		t.Run(fmt.Sprint(score), func(t *testing.T) {
			d := expiryFixture(t)
			d.game.Score = tablelogic.Number(score)
			counts := map[string]int{}
			var order []string
			changed := map[int]bool{}
			for d.terminal == nil && d.visits < 1100 {
				before := d.game.Display.Dots
				demoOK(t, d.calculation(false, true))
				m := d.game.matrix
				if len(order) == 0 || order[len(order)-1] != m.op {
					order = append(order, m.op)
					t.Log(d.counter, m.op, m.next, m.remaining, m.textLeft)
				}
				counts[m.op]++
				if before != d.game.Display.Dots && m.op == "_SCROLL" {
					changed[m.next] = true
				}
				if m.op == "_PRINT13_NUMBER" {
					want := newTestGame(t).Display
					want.Dots = before
					want.BeginCommand(demoExpiryProgram()[3])
					want.FlushPrint(d.game.matrixNumber)
					if want.Dots != d.game.Display.Dots {
						t.Fatal("score PRINTTASK")
					}
				}
			}
			if d.terminal == nil || d.terminal.Status != 0 || d.counter != 37047 || d.visits != 1050 || d.failure != nil {
				t.Fatalf("terminal %v visits %d timer %d failure %v", d.terminal, d.visits, d.counter, d.failure)
			}
			want := []string{"_CLEAR4", "_SCROLL", "_FLASHON", "_PRINT13_NUMBER", "_WAIT", "_FLASHOFF", "_SCROLL", "_DEMO_FADE", "_WAIT", "_DEMO_QUIT"}
			if !reflect.DeepEqual(order, want) || !changed[2] || !changed[7] || counts["_WAIT"] != 200 || counts["_DEMO_FADE"] != 256 || d.fade.Palette != [768]byte{} || len(d.fade.Volume) != 16 {
				t.Fatal("program consumers", order, counts, changed)
			}
			before := expirySnapshot(d)
			for i := 0; i < 5; i++ {
				demoOK(t, d.calculation(false, true))
				demoOK(t, d.electronics(true))
				demoOK(t, d.install(demoPartyProgram()))
				demoOK(t, (&demoConnected{demoCore: d}).sync(physics.Inputs{}, true))
			}
			if !reflect.DeepEqual(before, expirySnapshot(d)) {
				t.Fatal("post QUIT mutation")
			}
		})
	}
}
func expirySnapshot(d *demoCore) string {
	return fmt.Sprintf("%#v %#v %#v %#v %#v %#v", *d, *d.game, *d.game.Physics, *d.game.Display, d.game.waitCounters, d.game.Events)
}

// structuralCalculation returns the typed normal result separately from sticky
// unsupported failure. This is the matrix candidate API, not an input witness.
func (d *demoCore) structuralCalculation(paused, budget bool) (*demoTerminal, error) {
	err := d.calculation(paused, budget)
	return d.terminal, err
}
func advanceExpiry(t *testing.T, d *demoCore, op string) {
	t.Helper()
	for i := 0; i < 1100; i++ {
		if d.game.matrix.op == op {
			return
		}
		demoOK(t, d.calculation(false, true))
	}
	t.Fatal("missing stage", op)
}
func TestDemoExpiryStarvationPauseInterruption(t *testing.T) {
	for _, op := range []string{"_CLEAR4", "_SCROLL", "_FLASHON", "_PRINT13_NUMBER", "_WAIT", "_FLASHOFF", "_DEMO_FADE"} {
		t.Run(op, func(t *testing.T) {
			d := expiryFixture(t)
			advanceExpiry(t, d, op)
			if op == "_SCROLL" || op == "_DEMO_FADE" || op == "_WAIT" {
				for i := 0; i < 9; i++ {
					demoOK(t, d.calculation(false, true))
				}
			}
			cursor, dots, fade, visits := d.game.matrix, d.game.Display.Dots, d.fade, d.visits
			for i := 0; i < 3; i++ {
				demoOK(t, d.calculation(false, false))
			}
			if !reflect.DeepEqual(cursor, d.game.matrix) || dots != d.game.Display.Dots || !reflect.DeepEqual(fade, d.fade) || visits != d.visits {
				t.Fatal("budget advances consumer")
			}
			before := expirySnapshot(d)
			for i := 0; i < 3; i++ {
				r, e := d.structuralCalculation(true, true)
				demoOK(t, e)
				if r != nil {
					t.Fatal(r)
				}
			}
			if before != expirySnapshot(d) {
				t.Fatal("pause mutation")
			}
			f := demoFlash(d.game.Display)
			demoOK(t, d.install(demoWait(2)))
			if d.game.matrix.next != 1 || d.game.matrix.op != "_WAIT" || d.fade.Active || demoFlash(d.game.Display).Enabled || !d.game.Display.On || demoFlash(d.game.Display).Count != f.Count || demoFlash(d.game.Display).Phase != f.Phase {
				t.Fatal("replacement ownership")
			}
			for i := 0; i < 4; i++ {
				demoOK(t, d.calculation(false, true))
			}
			if d.game.matrix.op != "_WAIT" || d.terminal != nil {
				t.Fatal("old cursor restored")
			}
		})
	}
}
func TestDemoExpiryFlashAndScoreMoment(t *testing.T) {
	d := expiryFixture(t)
	advanceExpiry(t, d, "_SCROLL")
	for d.game.matrix.textLeft > 0 {
		demoOK(t, d.calculation(false, true))
	}
	d.game.Display.On = false
	demoOK(t, d.calculation(false, true))
	if f := demoFlash(d.game.Display); f != (demoFlashState{1, 1, true, true, false}) {
		t.Fatal("FLASHON dispatch changes On", f)
	}
	// Reads live score at PRINTTASK, after NEXT_A dispatch, never at equality.
	d.game.Score = tablelogic.Number(987654321)
	dots := d.game.Display.Dots
	demoOK(t, d.calculation(false, true))
	if d.game.Display.On {
		t.Fatal("first blink phase")
	}
	want := newTestGame(t).Display
	want.Dots = dots
	want.BeginCommand(demoExpiryProgram()[3])
	want.FlushPrint(d.game.matrixNumber)
	if want.Dots != d.game.Display.Dots {
		t.Fatal("live score read")
	}
	dots = d.game.Display.Dots
	cursor := d.game.matrix
	d.game.Score = tablelogic.Number(1)
	demoOK(t, d.calculation(false, false))
	if !d.game.Display.On || d.game.Display.Dots != dots || !reflect.DeepEqual(cursor, d.game.matrix) {
		t.Fatal("budget suppressed flash or repeated PRINTTASK")
	}
	advanceExpiry(t, d, "_FLASHOFF")
	f := demoFlash(d.game.Display)
	if f.Enabled || !f.On || f.Speed != 1 {
		t.Fatal("FLASHOFF", f)
	}
}
func TestDemoExpiryFadeDAC(t *testing.T) {
	d := expiryFixture(t)
	advanceExpiry(t, d, "_DEMO_FADE")
	initial := d.fade.Palette
	if initial == [768]byte{} || initial != d.fade.Snapshot {
		t.Fatal("palette capture")
	}
	original := d.game.lampPalette
	d.fade.Palette[0] ^= 1
	if d.game.lampPalette != original {
		t.Fatal("palette alias")
	}
	d.fade.Palette = initial
	for remaining := uint16(255); ; remaining-- {
		demoOK(t, d.calculation(false, true))
		for i, v := range initial {
			if d.fade.Palette[i] != byte(uint16(v)*remaining>>8) {
				t.Fatal("DAC MUL high byte", remaining, i)
			}
		}
		if remaining == 0 {
			break
		}
	}
	for i, v := range d.fade.Volume {
		if v != uint16(240-i*16) {
			t.Fatal("AL6 CX", d.fade.Volume)
		}
	}
	if d.game.matrix.op != "_WAIT" || d.game.matrix.remaining != 100 {
		t.Fatal("FADE boundary")
	}
}
func TestDemoExpiryRestart(t *testing.T) {
	d := expiryFixture(t)
	for d.counter < 36172 {
		demoOK(t, d.calculation(false, true))
	}
	old := d.game.matrix
	demoOK(t, d.queue(demoTask{Site: "structural-expiry-restart", Action: demoReplaceMatrix, Program: demoExpiryProgram()}))
	before := d.visits
	demoOK(t, d.calculation(false, true))
	if d.counter != 36173 || d.game.matrix.next != 1 || d.game.matrix.remaining != 4 || old.next == d.game.matrix.next {
		t.Fatal("restart entry")
	}
	for d.terminal == nil {
		demoOK(t, d.calculation(false, true))
		if d.counter > 37300 {
			t.Fatal("restart terminal missing")
		}
	}
	t.Logf("restart terminal=%d admitted=%d", d.counter, d.visits-before)
	// 1050 ordinary visits minus two visits inherited from half-completed glyph.
	if d.counter != 37220 || d.visits-before != 1048 {
		t.Fatal("restart phase", d.terminal, d.visits-before)
	}
}
func TestDemoExpiryUnsupportedStillFallible(t *testing.T) {
	d := expiryFixture(t)
	advanceExpiry(t, d, "_SCROLL")
	d.game.Display.Content.Commands[2].Op = "BYGEL12"
	for d.game.matrix.textLeft != 0 {
		demoOK(t, d.calculation(false, true))
	}
	cursor, dots := d.game.matrix, d.game.Display.Dots
	r, err := d.structuralCalculation(false, true)
	if err == nil || r != nil || d.failure.Producer != "NEXT_A" || !reflect.DeepEqual(cursor, d.game.matrix) || dots != d.game.Display.Dots {
		t.Fatal("preflight", r, err)
	}
	before := expirySnapshot(d)
	r, next := d.structuralCalculation(false, true)
	if next != err || r != nil || before != expirySnapshot(d) {
		t.Fatal("sticky failure")
	}
}

func TestDemoExpiryPriorityAndRestartAllStages(t *testing.T) {
	for _, op := range []string{"_SCROLL", "_DEMO_FADE", "_WAIT"} {
		t.Run(op, func(t *testing.T) {
			d := expiryFixture(t)
			advanceExpiry(t, d, op)
			for i := 0; i < 5; i++ {
				demoOK(t, d.calculation(false, true))
			}
			// Direct DO_MATRIX does not gain a synthetic priority guard from expiry.
			d.game.Audio.Priority = 255
			d.game.inhibitEffect = true
			old := d.game.matrix
			demoOK(t, d.install(demoExpiryProgram()))
			if d.game.matrix.next != 1 || d.game.matrix.remaining != 5 || d.game.matrix.op != "_CLEAR4" || old.next == 1 || d.game.Audio.Priority != 255 || d.game.inhibitEffect != true {
				t.Fatal("direct replacement priority/cursor")
			}
			for d.terminal == nil {
				demoOK(t, d.calculation(false, true))
				if d.visits > 2200 {
					t.Fatal("restart hang")
				}
			}
			if d.game.matrix.op != "_DEMO_QUIT" {
				t.Fatal("resumed displaced cursor")
			}
		})
	}
}

func TestDemoExpiryEveryCommandBoundary(t *testing.T) {
	d := expiryFixture(t)
	boundaries := []uint16{35998, 36002, 36315, 36316, 36317, 36417, 36418, 36691, 36947, 37047}
	for i, at := range boundaries {
		for d.counter < at {
			r, e := d.structuralCalculation(false, true)
			demoOK(t, e)
			if r != nil && d.counter != 37047 {
				t.Fatal("premature QUIT")
			}
		}
		if d.game.matrix.next != i+1 {
			t.Fatal("source command boundary", at, d.game.matrix)
		}
		if i == 4 || i == 8 {
			if d.game.matrix.remaining != 100 {
				t.Fatal("WAIT dispatch")
			}
		}
		cursor, dots, fade := d.game.matrix, d.game.Display.Dots, d.fade
		for n := 0; n < 2; n++ {
			r, e := d.structuralCalculation(true, true)
			demoOK(t, e)
			if i == 9 && r != d.terminal {
				t.Fatal("terminal result lost")
			}
		}
		if !reflect.DeepEqual(cursor, d.game.matrix) || dots != d.game.Display.Dots || !reflect.DeepEqual(fade, d.fade) {
			t.Fatal("pause at command", i)
		}
	}
	// Conformance mismatch is explicit evidence, never a fabricated 1042 deadline.
	if d.terminal.Visits == 1042 || d.terminal.Calculation == 37039 {
		t.Fatal("truncated reference cadence substituted")
	}
}

func TestDemoExpiryMissingPinnedConsumer(t *testing.T) {
	for _, op := range []string{"_SCROLL", "_PRINT13_NUMBER"} {
		t.Run(op, func(t *testing.T) {
			d := expiryFixture(t)
			if op == "_SCROLL" {
				advanceExpiry(t, d, "_CLEAR4")
				for d.game.matrix.remaining > 1 {
					demoOK(t, d.calculation(false, true))
				}
				delete(d.game.Display.Content.Texts, "DEMO_EXPIRY_TEXT1")
			} else {
				advanceExpiry(t, d, "_FLASHON")
				delete(d.game.Display.Content.Fonts, "13")
			}
			before := d.game.matrix
			r, e := d.structuralCalculation(false, true)
			if e == nil || r != nil || !reflect.DeepEqual(before, d.game.matrix) {
				t.Fatal("missing consumer advanced", op, r, e)
			}
		})
	}
}
