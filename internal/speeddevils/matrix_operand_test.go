package speeddevils

import "testing"

// TestSpeedDevilsFlashAndJingleResolved pins the symbolic _FLASHON operands
// (SDEV.ASM defines PARTY_ON_SPEED=3 and SHOOT_AGAIN_SPEED=3) and the recovered
// EMPTYJINGLE value. Before the extraction fix the raw symbols reached
// strconv.Atoi, whose error was dropped, so the flash speed silently became 0.
func TestSpeedDevilsFlashAndJingleResolved(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want int
	}{{"PARTY_ON_SPEED", 3}, {"SHOOT_AGAIN_SPEED", 3}} {
		t.Run(tc.raw, func(t *testing.T) {
			g := testGame(t)
			quiet(g)
			index := -1
			for i, c := range programs.Commands {
				if c.Op == "_FLASHON" && c.Arg(0) == tc.raw {
					index = i
					break
				}
			}
			if index < 0 {
				t.Fatalf("source _FLASHON,%s is not in the extracted stream", tc.raw)
			}
			c := programs.Commands[index]
			if got := c.Num(0); got != tc.want {
				t.Fatalf("_FLASHON,%s resolved to %d, want %d", tc.raw, got, tc.want)
			}
			g.matrix = matrix{active: true, next: index}
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
	t.Run("EMPTYJINGLE", func(t *testing.T) {
		g := testGame(t)
		quiet(g)
		for i, c := range programs.Commands {
			if c.Op != "_LASTJINGLE" || c.Arg(0) != "EMPTYJINGLE" {
				continue
			}
			if got := c.Num(0); got != 55 {
				t.Fatalf("_LASTJINGLE EMPTYJINGLE resolved to %d, want 55", got)
			}
			g.matrix = matrix{active: true, next: i}
			g.matrixDispatch()
			if g.music.ReturnPosition != 55 {
				t.Fatalf("music return position %d, want 55", g.music.ReturnPosition)
			}
			return
		}
		t.Fatal("source _LASTJINGLE,EMPTYJINGLE not extracted")
	})
}

// TestSpeedDevilsWaitExpressionsResolved checks the arithmetic wait operands
// that the SDEV extractor already resolved into typed values.
func TestSpeedDevilsWaitExpressionsResolved(t *testing.T) {
	want := map[string]int{
		"10*2*10":      200,
		"MILESWAIT-16": 24,
		"JMP_T":        3,
		"JMP_T*2":      6,
		"SHOWTIME":     80,
		"15*3-1":       44,
	}
	seen := map[string]bool{}
	for _, c := range programs.Commands {
		if c.Op != "_WAIT" {
			continue
		}
		value, ok := want[c.Arg(0)]
		if !ok {
			continue
		}
		seen[c.Arg(0)] = true
		if got := c.Num(0); got != value {
			t.Errorf("_WAIT,%s resolved to %d, want %d", c.Arg(0), got, value)
		}
	}
	for raw := range want {
		if !seen[raw] {
			t.Errorf("source _WAIT,%s not present in the extracted stream", raw)
		}
	}
}

// TestSpeedDevilsNumericOperandsResolved walks the extracted stream and requires
// a resolved integer for every operand the Speed Devils VM consumes.
func TestSpeedDevilsNumericOperandsResolved(t *testing.T) {
	numeric := map[string][]int{
		"_WAIT": {0}, "_LASTJINGLE": {0},
		"_RULLGARDIN_UPP": {1}, "_RULLGARDIN_NED": {1},
		"_COUNTDOWN": {0, 1}, "_FLASHON": {0},
	}
	for i, c := range programs.Commands {
		for _, k := range numeric[c.Op] {
			if _, ok := c.Nums[k]; !ok {
				t.Errorf("command %d: %s operand %d has no resolved value (args %v)", i, c.Op, k, c.Args)
			}
		}
	}
}
