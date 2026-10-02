package gameshow

import (
	"bytes"
	"testing"
)

func TestGameshowSkillSourceWrite(t *testing.T) {
	g := game(t)
	pristine := append([]byte(nil), g.Display.Content.Texts["SKILLTEXT"]...)
	for _, c := range []struct {
		before  uint16
		want    string
		changed bool
	}{{0, "", false}, {5, "", false}, {6, "", false}, {11, "", false}, {12, "*18\x00", true}, {18, "*24\x00", true}, {98, "*102", true}} {
		g.Skills = c.before
		g.anotherSkill()
		b := g.Display.Content.Texts["SKILLTEXT"]
		if c.changed {
			if g.Display.SourceText("SKILLTEXT")[16:] != c.want {
				t.Fatalf("skills %d: %q want %q", c.before, b[16:], c.want)
			}
		} else if !bytes.Equal(b, pristine) {
			t.Fatal("source no-write branch")
		}
	}
	// Across balls, player skill counter and DATA text remain live.
	before := append([]byte(nil), g.Display.Content.Texts["SKILLTEXT"]...)
	skills := g.Skills
	g.resetTable()
	if g.Skills != skills || !bytes.Equal(g.Display.Content.Texts["SKILLTEXT"], before) {
		t.Fatal("ball reset erased skill state")
	}
	if !bytes.Equal(game(t).Display.Content.Texts["SKILLTEXT"], pristine) {
		t.Fatal("new game contaminated")
	}
}
func TestGameshowBonusSourceWrite(t *testing.T) {
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
