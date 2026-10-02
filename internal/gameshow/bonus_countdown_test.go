package gameshow

import "testing"

// DO_FLORPA reaches zero, transfers every digit to score and holds the empty
// bonus field for fourteen syncs. Compare repaints, not only the BCD value.
func TestBonusCountdownClearsFinalDigit(t *testing.T) {
	for _, bonus := range []uint64{1, 100000, 123456} {
		g := game(t)
		g.Score = number(700)
		g.Bonus = number(bonus)
		g.Display.Clear()
		g.Display.Begin("_FLORPA", nil)
		g.Display.Number(g.Bonus.String(), -32, 6, 8)
		g.matrix = matrix{active: true, op: "_FLORPA", digit: 11, unit: 1}

		syncs := 0
		for g.Bonus.Uint64() != 0 && syncs < 500 {
			g.matrixTick()
			syncs++
		}
		if g.Bonus.Uint64() != 0 || g.Score.Uint64() != 700+bonus {
			t.Fatalf("bonus %d: remaining=%s score=%s", bonus, g.Bonus, g.Score)
		}
		want := game(t).Display
		want.Clear()
		want.Score(g.Score.String())
		if g.Display.Dots != want.Dots {
			t.Fatalf("bonus %d: stale digit at zero", bonus)
		}
		for i := 0; i < 13; i++ {
			g.matrixTick()
			if !g.matrix.active || g.matrix.op != "_FLORPA" || g.Display.Dots != want.Dots || g.Score.Uint64() != 700+bonus {
				t.Fatalf("bonus %d: zero hold changed at sync %d", bonus, i+1)
			}
		}
	}
}
