package gameshow

import (
	"strings"
	"testing"
)

// TestGameshowSourceReachableMatrixRegistered covers the SHOW runtime label
// map (the presentation table 3 content). It is the check that fails without
// the Beaten_bh_TS extractor-range fix.
func TestGameshowSourceReachableMatrixRegistered(t *testing.T) {
	g := game(t)
	// Source-reachable programs are emitted by tools/matrix_reachability.py
	// into the presentation package; assert the ones the SHOW rules name
	// directly, including the shared FANTASIE _beaten_matrix target.
	for _, label := range []string{
		"BEATEN_BH_TS", "BEATENTS", "TILTTS", "PARTY_ONTS", "PARTY_OFFTS",
		"SHOOT_AGAIN_ONTS", "OUT_OF_BALLSTS", "URBANOVERTS", "SHOWHIGHSTS",
		"SPINTS", "RENSA2TS", "RENSATS", "FLASHMATRIXTS",
		"_BONUSX2TS", "_BONUSX3TS", "_BONUSX4TS", "_BONUSX6TS",
		"_BONUSX8TS", "_BONUSX10TS",
	} {
		if _, ok := g.Display.Content.Labels[label]; !ok {
			t.Errorf("SHOW matrix program %s is not registered", label)
		}
	}
}

// TestGameshowBeatenMatrixBranchRegistered covers FANTASIE.ASM _beaten_matrix:
// the shared handler jumps to the table-local Beaten_bh_TS when the high score
// has been beaten.
func TestGameshowBeatenMatrixBranchRegistered(t *testing.T) {
	g := game(t)
	g.Score = number(100000001)
	g.SetHighScore(number(100000000))
	for i, c := range g.Display.Content.Commands {
		if c.Op == "_BEATEN_MATRIX" {
			g.matrix.next = i
			g.matrix.active = true
			g.matrixDispatch()
			if !g.beaten || !g.Lights[31] || g.matrix.next != g.Display.Content.Labels["BEATEN_BH_TS"]+1 {
				t.Fatalf("real beaten branch: beaten=%v extra=%v next=%d", g.beaten, g.Lights[31], g.matrix.next)
			}
			return
		}
	}
	t.Fatal("source _BEATEN_MATRIX opcode absent")
}

// TestGameshowBonusMultiplierFamily walks every BONUSTABLE state. SHOW indexes
// BONUS_ANIMS by bonusmultiplier-1 for values 2,3,4,6,8,10, so each pre-state
// 1,2,3,4,6,8 selects the next program.
func TestGameshowBonusMultiplierFamily(t *testing.T) {
	cases := map[uint8]string{
		1: "_BONUSX2TS", 2: "_BONUSX3TS", 3: "_BONUSX4TS",
		4: "_BONUSX6TS", 6: "_BONUSX8TS", 8: "_BONUSX10TS",
	}
	for pre, want := range cases {
		g := game(t)
		g.Multiplier = pre
		g.timers[2] = 1 // MBcounter: enable the multiplier branch
		g.timers[6] = 0 // CARcounter
		g.Special = true
		g.music.Priority = 0
		g.clockwise()
		started := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == want {
				started = true
			}
		}
		if g.Multiplier != map[uint8]uint8{1: 2, 2: 3, 3: 4, 4: 6, 6: 8, 8: 10}[pre] {
			t.Fatal("wrong multiplier transition", pre, g.Multiplier)
		}
		if !started {
			t.Errorf("multiplier %d: expected matrix %s to start", pre, want)
		}
	}
}

func TestGameshowMatrixPrizeFamilies(t *testing.T) {
	for i, name := range prizeNames {
		g := game(t)
		g.music.Priority = 0
		g.litPrize(i, false)
		want := name + "LITTS"
		found := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == want {
				found = true
			}
		}
		if !found {
			t.Fatal("prize lit", want)
		}
		g = game(t)
		g.TopThree = i >= 3
		base := 0
		if i >= 3 {
			base = 3
		}
		for j := base; j < i; j++ {
			g.Prizes[j] = 2
		}
		g.winPrize()
		want = "YOUWIN" + name + "TS"
		found = false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == want {
				found = true
			}
		}
		if !found {
			t.Fatal("prize won", want)
		}
	}
	g := game(t)
	g.Multiplier = 10
	g.timers[2] = 1
	g.clockwise()
	if g.Multiplier != 10 {
		t.Fatal("maximum multiplier changed")
	}
	for _, e := range g.Events {
		if e.Kind == "MatrixStarted" && strings.HasPrefix(e.Label, "_BONUSX") {
			t.Fatal("maximum multiplier selected another program")
		}
	}
}

func TestGameshowMatrixMoneyManiaStates(t *testing.T) {
	for pre, want := range map[uint16]string{17: "TURBO2TS", 23: "TURBOTS"} {
		g := game(t)
		g.Skills = pre
		g.music.Priority = 0
		g.anotherSkill()
		found := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == want {
				found = true
			}
		}
		if !found {
			t.Fatal("Money Mania parity", pre, want)
		}
	}
}
