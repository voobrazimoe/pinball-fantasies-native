package speeddevils

import "testing"

// DO_FLORPA transfers the final digit and sets LB6CNT=-10. PRINT8_TASK
// skips the all-zero BCD without erasing cells during the fourteen-sync hold.
// NO_MORE_NUFFROR owns the later bounded clear, not numeric rendering.
func TestBonusCountdownRetainsFinalDigitUntilSourceClear(t *testing.T) {
	for _, bonus := range []uint64{1, 100000, 123456} {
		g := testGame(t)
		g.Score = number(700)
		g.Bonus = number(bonus)
		g.Display.Clear()
		g.Display.Begin("_FLORPA", nil)
		g.Display.Number(g.Bonus.String(), -32, 6, 8)
		g.matrix = matrix{active: true, op: "_FLORPA", countDigit: 11, countUnit: 1}

		syncs := 0
		for g.Bonus.Uint64() != 0 && syncs < 500 {
			g.matrixTick()
			syncs++
		}
		if g.Bonus.Uint64() != 0 || g.Score.Uint64() != 700+bonus {
			t.Fatalf("bonus %d: remaining=%s score=%s", bonus, g.Bonus, g.Score)
		}
		want := testGame(t).Display
		want.Clear()
		want.Score(g.Score.String())
		finalUnit := uint64(1)
		for n := bonus; n >= 10; n /= 10 {
			finalUnit *= 10
		}
		want.Number(number(finalUnit).String(), -32, 6, 8)
		if !sameBonusField(g.Display.Dots, want.Dots) {
			t.Fatalf("bonus %d: PRINT8_TASK erased a skipped zero cell", bonus)
		}
		for i := 0; i < 13; i++ {
			g.matrixTick()
			if !g.matrix.active || g.matrix.op != "_FLORPA" || !sameBonusField(g.Display.Dots, want.Dots) || g.Score.Uint64() != 700+bonus {
				t.Fatalf("bonus %d: zero hold changed at sync %d", bonus, i+1)
			}
		}
		g.matrixTick() // fourteenth visit: NO_MORE_NUFFROR owns CLEAR_BOX2.
		for y := 6; y < 16; y++ {
			for x := 0; x < 56; x++ {
				if g.Display.Dots[y*160+x] {
					t.Fatal("NO_MORE_NUFFROR final clear", bonus, x, y)
				}
			}
		}
	}
}

// PLIPPA_SCORE uses CODE2's oldbuf cache. Comparing its retained digits with
// a freshly invalidated renderer invents redraws at a BCD width transition.
// This regression concerns the bonus field (source row 6, x=0..63).
func sameBonusField(a, b [160 * 16]bool) bool {
	for y := 6; y < 14; y++ {
		for x := 0; x < 64; x++ {
			if a[y*160+x] != b[y*160+x] {
				return false
			}
		}
	}
	return true
}
