package stones

import (
	"bytes"
	"testing"
)

func TestStonesBonusSourceWrite(t *testing.T) {
	for _, mult := range []uint8{1, 2, 3, 4, 6, 8, 10} {
		g := game(t)
		g.Multiplier = mult
		g.Bonus = number(100)
		want := []byte{mult + '7', ' '}
		if mult == 10 {
			want = []byte{'8', '7'}
		}
		// Real source program, including _BONUS_X_CALCS before printing the row.
		if mult == 1 {
			// The normal ball-lost program skips X1; dispatch the source opcode
			// directly to pin its unconditional NO_X_BONUS write as well.
			for i, c := range g.Display.Content.Commands {
				if c.Op == "_BONUS_X_CALCS" {
					g.matrix = matrix{active: true, next: i}
					g.matrixDispatch()
					break
				}
			}
		} else {
			g.beginMatrix("BALL_LOSTTS")
		}
		found := false
		for n := 0; n < 2000 && g.matrix.active; n++ {
			if g.matrix.op == "_BONUS_X_CALCS" {
				found = true
				break
			}
			g.music.ReadyAnim = true
			g.music.ReadyLogic = true
			g.matrixTick()
		}
		if !found {
			t.Fatal("bonus calculation absent")
		}
		if !bytes.Equal(g.Display.Content.Texts["BONUS_X_TEXT"][8:10], want) {
			t.Fatal("multiplier bytes", mult)
		}
		if string(game(t).Display.Content.Texts["BONUS_X_TEXT"]) != "BONUS*X*??\x00" {
			t.Fatal("new game contaminated")
		}
	}
}

func TestStonesPutInTextSourceZeroSuppression(t *testing.T) {
	for _, tc := range []struct {
		value uint16
		want  string
	}{{0, "***"}, {2, "**2"}, {10, "*10"}, {100, "100"}, {999, "999"}, {1002, "**2"}} {
		g := game(t)
		g.Screams, g.NextJump = tc.value, tc.value
		g.writeMilesText()
		g.writeJumpText()
		if got := g.Display.SourceText("MILES_TEXT")[4:7]; got != tc.want {
			t.Fatalf("MILES %d: %q != %q", tc.value, got, tc.want)
		}
		label := "JUMP_AT_TEXT"
		untouched := "JUMP_AT_TEXT2"
		if tc.value > 10 {
			label, untouched = untouched, label
		}
		if got := g.Display.SourceText(label)[:3]; got != tc.want {
			t.Fatalf("JUMP %d: %q != %q", tc.value, got, tc.want)
		}
		if got := g.Display.SourceText(untouched)[:3]; got != "XXX" {
			t.Fatal("unselected source buffer mutated", untouched, got)
		}
		if got := game(t).Display.SourceText("MILES_TEXT")[4:7]; got != "XXX" {
			t.Fatal("mutable text contaminated a new session")
		}
	}
}

func TestStonesScreamAwardWritesBeforeMatrix(t *testing.T) {
	g := game(t)
	g.Screams = 1
	g.NextJump = 10
	g.scream()
	if got := g.Display.SourceText("MILES_TEXT"); got != "    **2 SCREAMS\x00" {
		t.Fatal("actual scream path retained leading zeros", got)
	}
	if got := g.Display.SourceText("JUMP_AT_TEXT"); got != "*10 LITES EXTRA BALL\x00" {
		t.Fatal("source threshold formatting", got)
	}
	if g.Display.SourceText("JUMP_AT_TEXT2")[:3] != "XXX" {
		t.Fatal("source's other threshold buffer was written")
	}
}
