//go:build dmoimpl1

package partyland

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type demoCollisionWitness struct {
	Suffix struct{ Rows []map[string]any }
	Native []map[string]any `json:"native_task_reference"`
}

func loadDemoCollisionWitness(t *testing.T, needed bool) demoCollisionWitness {
	t.Helper()
	var w demoCollisionWitness
	if !needed {
		return w
	}
	path := os.Getenv("PF_DEMO_COLLISION_WITNESS")
	if path == "" {
		t.Log("collision saved witness NOT AVAILABLE")
		return w
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &w); e != nil || len(w.Suffix.Rows) != 122 || len(w.Native) != 122 {
		t.Fatal("collision artifact", e)
	}
	return w
}
func compareDemoCollisionWitness(t *testing.T, w demoCollisionWitness, n int, d *demoConnected) {
	t.Helper()
	g := d.game
	if len(w.Suffix.Rows) == 0 {
		return
	}
	row := w.Suffix.Rows[n-35877]
	native := w.Native[n-35877]
	if row["calculation"] != float64(n) || (n != 35877 && native["calculation"] != float64(n)) {
		t.Fatal("witness indexing")
	}
	got := map[string]any{"expired": d.expired, "HOLDSTILL": d.holdStill, "BALL_DOWN": g.Physics.Ball.Lost, "LOOSING": d.loosing, "I_UTSKJUT": g.inChute, "PARTYFLASH_after": g.partyFlash, "VISAKEYS_after": d.visaKeys, "matrix_op": g.matrix.op, "matrix_remaining": g.matrix.remaining}
	raw, _ := json.Marshal(got)
	_ = json.Unmarshal(raw, &got)
	for k, v := range got {
		if !reflect.DeepEqual(row[k], v) {
			t.Fatalf("linked witness %d %s got=%v want=%v", n, k, v, row[k])
		}
	}
	// Cursor normalization maps compact typed nodes to their actual linked sites.
	cursors := []int{0x1b459, 0x1b45b, 0x1b461, 0x1b465, 0x1b467, 0x1b49f, 0x1b4cb, 0x1b4eb, 0x1b50b, 0x1b531, 0x1b533, 0x1b535, 0x1b537, 0x1b53b}
	cursor := 0x1b890
	if n < 35998 {
		cursor = cursors[g.matrix.next]
	}
	if row["cursor"] != float64(cursor) {
		t.Fatal("linked cursor", n, cursor, row["cursor"])
	}
	waits := row["WAITLIST"].(map[string]any)
	for site, v := range waits {
		if float64(g.waitCounters[site]) != v {
			t.Fatal("linked wait", n, site, g.waitCounters, v)
		}
	}
	sites := row["tasklist_after"].([]any)
	for i, site := range sites {
		if (g.tasks[i] == nil) != (site == "") {
			t.Fatal("linked live slot", n, i, site)
		}
	}
	ball := g.Physics.Ball
	ball.Hold = d.holdStill // representation normalization only
	projected := map[string]any{"ball": ball, "score": g.Score.Uint64(), "SCORECHANGED": g.ScoreChanged, "totals": [4]uint64{g.Bonus.Uint64(), uint64(g.Cyclones), g.HappyTotal.Uint64(), g.MegaTotal.Uint64()}, "lights": g.Lights, "flippers": g.Physics.Flippers, "touch_disabled": g.touchDisabled, "Arcade": g.Arcade}
	raw, _ = json.Marshal(projected)
	_ = json.Unmarshal(raw, &projected)
	for k, v := range projected {
		if !reflect.DeepEqual(native[k], v) {
			t.Fatalf("native witness %d %s got=%v want=%v", n, k, v, native[k])
		}
	}
}
func assertDemoEquality(t *testing.T, d *demoConnected) {
	t.Helper()
	g := d.game
	if d.counter != 35998 || d.calls != 35998 || g.Physics.Syncs != 35998 || !d.expired || !d.holdStill || d.loosing || g.Physics.Ball.Lost || g.Physics.Ball.Hold || !g.inChute || g.ScoreChanged || g.Score != Number(50030) || g.Bonus != Number(0) || g.partyFlash || d.visaKeys || d.cueRequests != 1 || d.handoffs != 1 || d.drains != 1 || g.matrix.next != 1 || g.matrix.op != "_CLEAR4" || g.matrix.remaining != 4 || len(d.drainFires) != 2 || d.drainFires[0].Site != "SOUNDRINNER" || d.drainFires[0].Counter != 35882 || d.drainFires[1].Counter != 35998 {
		t.Fatalf("collision state d=%+v ball=%+v matrix=%+v fires=%v", d, g.Physics.Ball, g.matrix, d.drainFires)
	}
	if !d.scoredDueExpiry || d.scoredDueMatrix.op != "_CLEAR4" || d.scoredDueMatrix.next != 1 || d.scoredDueMatrix.remaining != 5 || d.scoredDueAudio.Position != 13 || d.scoredDueAudio.Priority != 255 || g.Audio.Position != 13 || g.Audio.Priority != 255 || g.Audio.JumpCount != 0 || g.Audio.ReturnPosition != 0 {
		t.Fatal("expiry admission before due reset, spring request rejected by priority", d.scoredDueMatrix, d.scoredDueAudio, g.Audio)
	}
	if g.waitCounters["SOUNDNEWBALL"] != 0 || g.waitCounters["SETBALL"] != 1 || g.waitCounters["SOUNDBRICKUPP"] != 1 || g.waitCounters["NEW_BALL_TASK"] != 0 {
		t.Fatal("same calculation scan", g.waitCounters)
	}
	if g.Display.Content.Commands[1].Args[0] != "PLAYERSTEXT" || g.Display.Content.Commands[1].Nums[1] != 336 || g.Display.Content.Commands[2].Args[0] != "DEMO_BALLSTEXT" {
		t.Fatal("source reset program replaced expiry")
	}
	var ops []string
	for _, e := range g.Events {
		if e.Kind == "MatrixCommand" {
			ops = append(ops, e.Label)
		}
	}
	if !reflect.DeepEqual(ops, []string{"_CLEAR4", "_CLEAR4"}) {
		t.Fatal("expiry install then reset install, before sole matrix visit", ops)
	}
	t.Logf("35998 COMPLETE: expiry installed then replaced via real NEW_BALL_TASK; matrix clear remaining4; score=%s waits=%v", g.Score.String(), g.waitCounters)
}
