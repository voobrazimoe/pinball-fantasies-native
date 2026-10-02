package partyland

import "testing"

// TestPartyLandFlashSymbolsResolved pins the symbolic _FLASHON operands.
// PLAND.ASM defines DRSPEED=13, LMSPEED=7, PARTY_ON_SPEED=3 and
// SHOOT_AGAIN_SPEED=3; before the extraction fix these reached the shared
// renderer as raw text, strconv.Atoi failed and the error was dropped, so the
// matrix never flashed. Flash() is the observable renderer behaviour.
func TestPartyLandFlashSymbolsResolved(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"DRSPEED", 13},
		{"LMSPEED", 7},
		{"PARTY_ON_SPEED", 3},
		{"SHOOT_AGAIN_SPEED", 3},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			g := newTestGame(t)
			quiet(g)
			index := -1
			for i, c := range timing.Commands {
				if c.Op == "_FLASHON" && c.Arg(0) == tc.raw {
					index = i
					break
				}
			}
			if index < 0 {
				t.Fatalf("source _FLASHON,%s is not in the extracted gameplay stream", tc.raw)
			}
			c := timing.Commands[index]
			if got := c.Num(0); got != tc.want {
				t.Fatalf("_FLASHON,%s resolved to %d, want %d", tc.raw, got, tc.want)
			}
			// Drive the real dispatcher for this command, then the real
			// renderer clock. A zero flash speed never toggles the matrix.
			g.matrix = matrixState{active: true, next: index}
			g.matrixDispatch()
			g.Display.Visit(0, 2, g.matrixNumber)
			for i := 0; i < tc.want; i++ {
				g.Display.Flash()
			}
			if g.Display.On {
				t.Fatalf("_FLASHON,%s (speed %d) did not toggle the matrix", tc.raw, tc.want)
			}
		})
	}
}

// TestPartyLandLastJingleResolved resolves EMPTYJINGLE from the table's own
// S_Empty record instead of the previous runtime hardcode.
func TestPartyLandLastJingleResolved(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	found := false
	for i, c := range timing.Commands {
		if c.Op != "_LASTJINGLE" || c.Arg(0) != "EMPTYJINGLE" {
			continue
		}
		found = true
		if got := c.Num(0); got != 62 {
			t.Fatalf("_LASTJINGLE EMPTYJINGLE resolved to %d, want 62", got)
		}
		g.matrix = matrixState{active: true, next: i}
		g.matrixDispatch()
		if g.Audio.ReturnPosition != 62 {
			t.Fatalf("music return position %d, want 62", g.Audio.ReturnPosition)
		}
		break
	}
	if !found {
		t.Fatal("source _LASTJINGLE,EMPTYJINGLE not extracted")
	}
}

// TestPartyLandNumericOperandsResolved walks the extracted gameplay stream and
// requires a resolved integer for every operand the Party Land VM consumes.
func TestPartyLandNumericOperandsResolved(t *testing.T) {
	numeric := map[string][]int{
		"_WAIT": {0}, "_LASTJINGLE": {0},
		"_RULLGARDIN_UPP": {1}, "_RULLGARDIN_NED": {1},
		"_COUNTDOWN": {0, 1}, "_COUNTDOWN2": {0, 1}, "_FLASHON": {0},
	}
	for i, c := range timing.Commands {
		for _, k := range numeric[c.Op] {
			if _, ok := c.Nums[k]; !ok {
				t.Errorf("command %d: %s operand %d has no resolved value (args %v)", i, c.Op, k, c.Args)
			}
		}
	}
}
