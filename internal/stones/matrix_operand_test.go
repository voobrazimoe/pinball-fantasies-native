package stones

import (
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"

	"pinballfantasies/internal/presentation"
)

// TestStonesWaitExpressionOperandsResolved reproduces the live regression
// operand. STONES.ASM:1475 (SKILLSHOTTS) and STONES.ASM:1598 (MILLIONPLUSTS)
// both write `DW _WAIT,2*60`. The pre-fix extractor kept the raw expression and
// stones.numeric() panicked with "STONES integer 2*60" on the first dispatch.
func TestStonesWaitExpressionOperandsResolved(t *testing.T) {
	for _, program := range []string{"SKILLSHOTTS", "MILLIONPLUSTS"} {
		t.Run(program, func(t *testing.T) {
			g := game(t)
			g.beginMatrix(program)
			var wait presentation.Command
			found := false
			for n := 0; n < 2000 && g.matrix.active; n++ {
				if g.matrix.op == "_WAIT" && g.matrix.args[0] == "2*60" {
					wait = presentation.Command{Op: g.matrix.op, Args: g.matrix.args, Nums: g.matrix.nums}
					found = true
					break
				}
				g.matrixTick()
			}
			if !found {
				t.Fatalf("%s never reached the source _WAIT,2*60", program)
			}
			if wait.Args[0] != "2*60" {
				t.Fatalf("source provenance lost: %q", wait.Args[0])
			}
			if got := wait.Num(0); got != 120 {
				t.Fatalf("_WAIT,2*60 resolved to %d, want 120", got)
			}
			// Drive the real scheduler through the fully resolved wait.
			for n := 0; n < 120; n++ {
				g.matrixTick()
			}
			if g.matrix.op == "_WAIT" && g.matrix.args[0] == "2*60" {
				t.Fatal("_WAIT,2*60 did not expire after its 120 source ticks")
			}
		})
	}
}

// TestStonesPackedTowerOperandsResolved pins the three packed Tower row
// selectors PF10.1 discovered. Their source text stays for provenance while the
// runtime consumes the resolved row.
func TestStonesPackedTowerOperandsResolved(t *testing.T) {
	testinputs.Require(t, "../../TABLE4.PRG")
	b, err := os.ReadFile("../../TABLE4.PRG")
	if err != nil {
		t.Fatal(err)
	}
	d := presentation.New(4, b)
	want := map[string]int{"87-16+2": 73, "61-16+2": 47, "36-16": 20}
	for _, c := range d.Content.Commands {
		if c.Op != "_TOWER" {
			continue
		}
		value, ok := want[c.Args[0]]
		if !ok {
			t.Fatalf("unexpected source Tower operand %q", c.Args[0])
		}
		if got := c.Num(0); got != value {
			t.Fatalf("_TOWER %s resolved to %d, want %d", c.Args[0], got, value)
		}
		delete(want, c.Args[0])
	}
	if len(want) != 0 {
		t.Fatalf("source Tower operands missing from the generated stream: %v", want)
	}
}

// TestStonesReachableProgramsOperandSmoke runs the real Stones VM over every
// registered program. It is a bounded dispatch smoke: each program starts at
// its source entry and ticks until it terminates or the bound expires, so every
// operand the VM can consume in that program is exercised.
func TestStonesReachableProgramsOperandSmoke(t *testing.T) {
	testinputs.Require(t, "../../TABLE4.PRG")
	b, err := os.ReadFile("../../TABLE4.PRG")
	if err != nil {
		t.Fatal(err)
	}
	d := presentation.New(4, b)
	for label := range d.Content.Labels {
		if label == "LAST_POS" {
			continue
		}
		t.Run(label, func(t *testing.T) {
			g := game(t)
			g.beginMatrix(label)
			for n := 0; n < 4000 && g.matrix.active; n++ {
				g.matrixTick()
			}
		})
	}
}

// TestStonesNumericOperandsResolved walks the extracted Stones gameplay stream
// and requires a resolved integer for every operand the Stones VM consumes.
func TestStonesNumericOperandsResolved(t *testing.T) {
	testinputs.Require(t, "../../TABLE4.PRG")
	b, err := os.ReadFile("../../TABLE4.PRG")
	if err != nil {
		t.Fatal(err)
	}
	d := presentation.New(4, b)
	numeric := map[string][]int{
		"_WAIT": {0}, "_TOWER": {0}, "_LASTJINGLE": {0},
		"_RULLGARDIN_UPP": {1}, "_RULLGARDIN_NED": {1},
		"_COUNTDOWN": {0, 1}, "_FLASHON": {0},
	}
	for i, c := range d.Content.Commands {
		for _, k := range numeric[c.Op] {
			if _, ok := c.Nums[k]; !ok {
				t.Errorf("command %d: %s operand %d (%q) has no resolved value", i, c.Op, k, c.Arg(k))
				continue
			}
			if want, ok := d.Content.Positions[c.Arg(k)]; ok && c.Num(k) != want {
				t.Errorf("command %d: %s operand %d (%q) resolved %d, source positions %d",
					i, c.Op, k, c.Arg(k), c.Num(k), want)
			}
		}
	}
}
