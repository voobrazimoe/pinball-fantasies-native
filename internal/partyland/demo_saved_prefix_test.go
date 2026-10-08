//go:build dmoimpl1

package partyland

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type demoSavedRow struct {
	Start, End int
	State      map[string]any
}

func loadDemoSavedWitness(t *testing.T) []demoSavedRow {
	t.Helper()
	p := os.Getenv("PF_DEMO_RESEARCH_WITNESS")
	if p == "" {
		t.Log("saved research witness NOT AVAILABLE; independent A comparison remains active")
		return nil
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var v struct {
		Replay struct{ Witness struct{ Rows []demoSavedRow } } `json:"deterministic_replay"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Replay.Witness.Rows) == 0 {
		t.Fatal("empty saved witness")
	}
	return v.Replay.Witness.Rows
}
func demoSavedState(t *testing.T, rows []demoSavedRow, n int) map[string]any {
	t.Helper()
	for _, r := range rows {
		if r.Start <= n && n <= r.End {
			return r.State
		}
	}
	t.Fatal("missing saved calculation", n)
	return nil
}
func compareDemoSavedWitness(t *testing.T, rows []demoSavedRow, n int, g *Game) {
	t.Helper()
	if rows == nil {
		return
	}
	want := demoSavedState(t, rows, n)
	values := map[string]any{"ball": g.Physics.Ball, "spring": g.Physics.SpringPosition, "spring_valid": g.Physics.SpringValid, "score": g.Score.Uint64(), "SCORECHANGED": g.ScoreChanged, "totals": [4]uint64{g.Bonus.Uint64(), uint64(g.Cyclones), g.HappyTotal.Uint64(), g.MegaTotal.Uint64()}, "light39": g.Lights[39], "lights": g.Lights, "last_area": g.lastArea, "last_check": g.lastCheck, "skill_time": g.SkillTime, "inhibit_reverse_time": g.InhibitReverseTime, "Arcade": g.Arcade, "touch_disabled": g.touchDisabled, "flippers": g.Physics.Flippers, "waits": g.waitCounters}
	raw, _ := json.Marshal(values)
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	for key, v := range got {
		if !reflect.DeepEqual(v, want[key]) {
			t.Fatalf("saved witness calculation=%d field=%s got=%v want=%v", n, key, v, want[key])
		}
	}
}
func compareDemoSavedScore(t *testing.T, rows []demoSavedRow, n int, g *Game) {
	t.Helper()
	if rows == nil {
		return
	}
	want := demoSavedState(t, rows, n)
	if want["score"] != float64(g.Score.Uint64()) || want["SCORECHANGED"] != g.ScoreChanged || want["light39"] != g.Lights[39] {
		t.Fatal("saved actual scored event differs")
	}
	t.Logf("saved deterministic witness matches completed1..35789 and actual score at%d", n)
}
