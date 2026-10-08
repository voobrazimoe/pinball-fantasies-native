package frontend

import (
	"crypto/sha256"
	"fmt"
	"os"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/tablelogic"
	"testing"
)

// This scripted plunger/flipper run gives the same score, game length, PCM and
// final frame on every supported layout as on canonical A. It does not reach
// the edition differences (startup pictures, factory initials, S_EMPTY
// priority), which the per-profile tests cover.
func TestPrivateLayoutsPlayLikeCanonical(t *testing.T) {
	layouts := []struct{ env, profile string }{
		{"PF_RUNTIME_DATA", datalayout.RetailProfile},
		{"PF_POWERPACK_DATA", datalayout.PowerPackProfile},
		{"PF_DELUXE_CD_ALT_DATA", datalayout.DeluxeCDAltProfile},
		{"PF_DELUXE_CD_DATA", datalayout.DeluxeCDProfile},
	}
	var want [4]string
	for i, layout := range layouts {
		dir := os.Getenv(layout.env)
		if dir == "" {
			t.Skip("supply private A/B/C/D installations")
		}
		r := loadCompatible(t, stageInstallation(t, dir))
		if r.ProfileID != layout.profile {
			t.Fatalf("%s: profile %s", layout.env, r.ProfileID)
		}
		for frame := 0; frame < 3000; frame++ {
			if err := r.Update(Input{}); err != nil {
				t.Fatalf("%s: frontend %d: %v", layout.profile, frame, err)
			}
		}
		for table := range r.Model.Factories {
			got := scriptedGame(t, r, table)
			if i == 0 {
				want[table] = got
			} else if got != want[table] {
				t.Errorf("%s table %d: %s, canonical %s", layout.profile, table+1, got, want[table])
			}
		}
	}
}

func scriptedGame(t *testing.T, r *Runtime, table int) string {
	t.Helper()
	factory := r.Model.Factories[table]
	if factory == nil && table == 0 {
		factory = r.Model.Factory
	}
	s, err := factory(tablelogic.Decimal{})
	if err != nil {
		t.Fatal(err)
	}
	if g, ok := s.(interface{ StartPlayers(int) }); ok {
		g.StartPlayers(1)
	}
	pcm := sha256.New()
	var score tablelogic.Decimal
	done := false
	sync := 0
	for ; sync < 40000 && !done; sync++ {
		in := physics.Inputs{Down: sync%600 < 90, Release: sync%600 == 90}
		in.Left = (sync/7)%9 == 0
		in.Right = (sync/11)%8 == 0
		if err := s.Sync(in); err != nil {
			t.Fatalf("table %d sync %d: %v", table+1, sync, err)
		}
		pcm.Write(s.PCM())
		score, done = s.Result()
	}
	if !done || score == (tablelogic.Decimal{}) {
		t.Fatalf("table %d: game did not finish with a score (%s after %d syncs)", table+1, score.String(), sync)
	}
	frame := sha256.Sum256(s.Frame().Pix)
	return fmt.Sprintf("score=%s syncs=%d pcm=%x frame=%x", score.String(), sync, pcm.Sum(nil)[:8], frame[:8])
}
