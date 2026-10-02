package gameshow

import (
	"testing"
)

// TestGameshowWaitExpressionsResolved reproduces the two unexercised SHOW
// arithmetic operands. SHOW.ASM:1039 BEATEN_BH_TS writes `DW _WAIT,10*2*15`
// and SHOW.ASM:1051 BEATENTS writes `DW _WAIT,10*2*5`. The pre-fix dispatcher
// ran strconv.Atoi on the raw text and panicked.
func TestGameshowWaitExpressionsResolved(t *testing.T) {
	cases := []struct {
		program string
		raw     string
		want    int
	}{
		{"BEATEN_BH_TS", "10*2*15", 300},
		{"BEATENTS", "10*2*5", 100},
	}
	for _, tc := range cases {
		t.Run(tc.program, func(t *testing.T) {
			g := game(t)
			g.beginMatrix(tc.program)
			found := false
			for n := 0; n < 2000 && g.matrix.active; n++ {
				if g.matrix.op == "_WAIT" && g.matrix.args[0] == tc.raw {
					found = true
					break
				}
				g.matrixTick()
			}
			if !found {
				t.Fatalf("%s never reached the source _WAIT,%s", tc.program, tc.raw)
			}
			if got := g.matrix.remaining; got != uint16(tc.want) {
				t.Fatalf("_WAIT,%s scheduled %d ticks, want %d", tc.raw, got, tc.want)
			}
			for n := 0; n < tc.want; n++ {
				g.matrixTick()
			}
			if g.matrix.op == "_WAIT" && g.matrix.args[0] == tc.raw {
				t.Fatalf("_WAIT,%s did not expire", tc.raw)
			}
		})
	}
}

// TestGameshowNumericOperandsResolved walks the extracted Gameshow gameplay
// stream and requires a resolved integer for every operand the VM consumes.
func TestGameshowNumericOperandsResolved(t *testing.T) {
	g := game(t)
	numeric := map[string][]int{
		"_WAIT": {0}, "_LASTJINGLE": {0},
		"_RULLGARDIN_UPP": {1}, "_RULLGARDIN_NED": {1},
		"_COUNTDOWN": {0, 1}, "_FLASHON": {0},
	}
	for i, c := range g.Display.Content.Commands {
		for _, k := range numeric[c.Op] {
			if _, ok := c.Nums[k]; !ok {
				t.Errorf("command %d: %s operand %d (%q) has no resolved value", i, c.Op, k, c.Arg(k))
				continue
			}
			if want, ok := g.Display.Content.Positions[c.Arg(k)]; ok && c.Num(k) != want {
				t.Errorf("command %d: %s operand %d (%q) resolved %d, source positions %d",
					i, c.Op, k, c.Arg(k), c.Num(k), want)
			}
		}
	}
}
