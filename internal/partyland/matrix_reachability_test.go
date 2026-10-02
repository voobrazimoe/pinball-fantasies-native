package partyland

import "testing"

// TestPartyLandReachableMatrixRegistered checks the runtime dispatch map that
// the table actually uses (timing.Labels), not the render-only presentation
// map, against the source-derived reachable set.
func TestPartyLandReachableMatrixRegistered(t *testing.T) {
	for _, label := range sourceReachablePartyLand {
		if _, ok := timing.Labels[label]; !ok {
			t.Errorf("source-reachable matrix program %s is not registered", label)
		}
	}
}

// TestPartyLandBonusMultiplierFamily walks every source multiplier state.
// PLAND.ASM selects EFFECT M2/M4/M6/M8 from LIGHTSTATUS 47/49/50/48 in that
// order, so the pre-trigger multiplier is 1/2/4/6.
func TestPartyLandBonusMultiplierFamily(t *testing.T) {
	cases := map[uint8]string{1: "BONUSX2TS", 2: "BONUSX4TS", 4: "BONUSX6TS", 6: "BONUSX8TS"}
	for mult, want := range cases {
		g := newTestGame(t)
		g.Multiplier = mult
		g.effect("MULTIBONUS", 10000, 5000) // must resolve, must not panic
		started := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == want {
				started = true
			}
		}
		if !started {
			t.Errorf("multiplier %d: expected matrix %s to start", mult, want)
		}
	}
}

func TestPartyLandMatrixLetterFamily(t *testing.T) {
	for i, label := range []string{"PARTY_PTS", "PARTY_ATS", "PARTY_RTS", "PARTY_TTS", "PARTY_YTS"} {
		g := newTestGame(t)
		g.party(i)
		found := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == label {
				found = true
			}
		}
		if !found {
			t.Fatal("letter matrix", i, label)
		}
	}
}

func TestPartyLandMatrixStateSelectedFamilies(t *testing.T) {
	for mode, want := range []string{"JACKPOTTS", "JACKPOT_SPECIAL_HH_TS", "JACKPOT_SPECIAL_ML_TS"} {
		g := newTestGame(t)
		g.JackpotNormal = true
		g.Happy = mode == 1
		g.Mega = mode == 2
		g.dragon()
		found := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == want {
				found = true
			}
		}
		if !found {
			t.Fatal("jackpot state", mode, want)
		}
	}
	for _, alternate := range []bool{false, true} {
		g := newTestGame(t)
		g.arcadeCrazy = alternate
		g.crazy()
		found := false
		for _, e := range g.Events {
			if e.Kind == "MatrixStarted" && e.Label == "CRAZYLETTERTS" {
				found = true
			}
		}
		if !found {
			t.Fatal("crazy alternate", alternate)
		}
	}
}
