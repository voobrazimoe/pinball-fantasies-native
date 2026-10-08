//go:build dmoimpl1

package partyland

import (
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
	"reflect"
	"testing"
)

func TestDemoFreshFirstScoredPrefix(t *testing.T) {
	g := newTestGame(t)
	g.Configure(settings.Legacy())
	reference := newTestGame(t)
	reference.Configure(settings.Legacy())
	d := connected(g)
	d.loadToucherOperands(t)
	g.Physics.OnEvent = func(physics.Event) { t.Fatal("canonical OnEvent fallback") }
	g.Physics.BeforeTargets = func() { t.Fatal("canonical BeforeTargets fallback") }
	g.Physics.AfterTargets = func(physics.Inputs) { t.Fatal("canonical AfterTargets fallback") }
	g.Physics.BeforeLate = func() { t.Fatal("canonical BeforeLate fallback") }
	if d.counter != 0 || g.clock != 0 || g.ScoreChanged || g.Score.Uint64() != 0 || g.Bonus.Uint64() != 0 || g.Session.CurrentPlayer != 1 || g.BallNumber != 1 || g.matrix.active || d.expired || d.loosing || d.holdStill || g.Physics.Ball.Lost || g.Physics.Ball.Hold || len(g.waitCounters) != 0 {
		t.Fatal("nonfresh entry")
	}
	for _, task := range g.tasks {
		if task != nil {
			t.Fatal("nonfresh task")
		}
	}
	// Independent reference observation at the same pre-late boundary.
	var referenceEarly physics.Ball
	originalBeforeTargets := reference.Physics.BeforeTargets
	reference.Physics.BeforeTargets = func() { originalBeforeTargets(); referenceEarly = reference.Physics.Ball }
	counts := map[string]int{}
	saved := loadDemoSavedWitness(t)
	for n := 1; n <= 35790; n++ {
		if err := reference.Sync(demoFixedInput(n)); err != nil {
			t.Fatal(err)
		}
		if err := d.sync(demoFixedInput(n), true); err != nil {
			if n != 35790 || d.failure.Producer != "NODOT" || d.failure.Guard != "CHECKHIGHSCORE" || d.failure.Phase != "task/matrix" || g.Physics.Syncs != 35789 {
				t.Fatalf("unexpected boundary %d: %v", n, err)
			}
			if g.Physics.Ball != referenceEarly {
				t.Fatalf("scored-boundary early physics mismatch candidate=%+v reference=%+v", g.Physics.Ball, referenceEarly)
			}
			t.Logf("scored boundary: %v ball=%+v score=%s bonus=%s changed=%t events=%v", err, g.Physics.Ball, g.Score.String(), g.Bonus.String(), g.ScoreChanged, g.Events)
			if g.Score != reference.Score || g.Bonus != reference.Bonus || g.ScoreChanged != reference.ScoreChanged {
				t.Fatal("score reference mismatch")
			}
			counts["BYGEL1"]++
			compareDemoSavedScore(t, saved, n, g)
			assertConnectedSticky(t, d, err)
			break
		}

		if g.Physics.Ball != reference.Physics.Ball || !reflect.DeepEqual(g.Physics.Flippers, reference.Physics.Flippers) || g.Physics.SpringPosition != reference.Physics.SpringPosition || g.Physics.ScreenOffset != reference.Physics.ScreenOffset || g.Physics.Raster != reference.Physics.Raster || g.Physics.ScreenSpeed != reference.Physics.ScreenSpeed || g.Physics.ScreenPosition != reference.Physics.ScreenPosition {
			t.Fatalf("reference trajectory mismatch at %d candidate=%+v reference=%+v", n, g.Physics.Ball, reference.Physics.Ball)
		}
		if d.counter != uint16(n) || g.Random != reference.Random || g.clock != reference.clock || g.Score != reference.Score || g.Bonus != reference.Bonus || g.ScoreChanged != reference.ScoreChanged || g.SkillTime != reference.SkillTime || g.InhibitReverseTime != reference.InhibitReverseTime || g.lastArea != reference.lastArea || g.lastCheck != reference.lastCheck || g.Lights != reference.Lights || g.touchDisabled != reference.touchDisabled || g.Arcade != reference.Arcade || g.Lamps != reference.Lamps || g.flashes != reference.flashes || g.Audio != reference.Audio || !reflect.DeepEqual(g.waitCounters, reference.waitCounters) {
			t.Fatalf("electronics mismatch at %d", n)
		}
		compareDemoSavedWitness(t, saved, n, g)
		if n == 35712 || n == 35713 || n == 35731 || n == 35732 || n == 35790 {
			t.Logf("checkpoint %d ball=%+v score=%s bonus=%s changed=%t touch=%t arcade=%t tasks=%v waits=%v fires=%v flashes=%v events=%v matrix=%+v", n, g.Physics.Ball, g.Score.String(), g.Bonus.String(), g.ScoreChanged, g.touchDisabled, g.Arcade, g.taskIDs, g.waitCounters, d.toucherFires, g.flashes, g.Events, g.matrix)
		}
		if n >= 35712 && n < 35732 && (!g.touchDisabled || g.tasks[0] == nil || g.waitCounters["ENABLETOUCHER"] != uint16(n-35711)) {
			t.Fatal("fresh task age", n)
		}
		if n >= 35732 && (g.touchDisabled || g.tasks[0] != nil || !reflect.DeepEqual(d.toucherFires, []uint16{35732})) {
			t.Fatal("fresh task firing", n)
		}
		if n == 35460 {
			t.Logf("release clock=%d low8=%d ball=%+v spring=%d", g.clock, uint8(g.clock), g.Physics.Ball, g.Physics.SpringPosition)
		}
		for _, e := range g.Events {
			if e.Kind == "Switch" {
				counts[e.Label]++
				t.Logf("callback calculation=%d %s ball=%+v timers=%d/%d score=%s", n, e.Label, g.Physics.Ball, g.SkillTime, g.InhibitReverseTime, g.Score.String())
			}
		}
	}
	if g.Score != Number(50030) || g.Score.String() != "000000050030" || g.Bonus.Uint64() != 0 || !g.ScoreChanged || g.Lights[39] || d.infoCount != 0 || g.lastArea != "BYGEL1" || !g.Arcade || g.Happy || g.Mega || g.Physics.Syncs != 35789 {
		t.Fatal("fresh actual scored state", g.Score.String())
	}
	if !reflect.DeepEqual(counts, map[string]int{"BYGEL12": 2, "BYGEL28": 1, "CLOSE1": 1, "BYGEL9": 1, "BYGEL11": 1, "BYGEL1": 1}) {
		t.Fatal("fresh counts", counts)
	}
	if len(g.Events) < 3 || g.Events[0].Label != "BYGEL1" || g.Events[1].Label != "SBYGEL1" || g.Events[2].Kind != "ScoreAwarded" || g.Events[2].Value != 50030 {
		t.Fatal("actual score sound bookkeeping", g.Events)
	}
}
