//go:build dmoimpl1

package partyland

import (
	"fmt"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
	"reflect"
	"testing"
)

func TestDemoFreshCompletedScoredCalculation(t *testing.T) { demoFreshScored(t, false) }
func TestDemoFreshScoredDrainContinuation(t *testing.T)    { demoFreshScored(t, true) }
func TestDemoFreshInfoContinuation(t *testing.T)           { demoFreshScored(t, true, true) }
func demoFreshScored(t *testing.T, continuation bool, info ...bool) {
	showInfo := len(info) > 0 && info[0]
	g := newTestGame(t)
	g.Configure(settings.Legacy())
	reference := newTestGame(t)
	reference.Configure(settings.Legacy())
	d := connected(g)
	if showInfo {
		d.loadInfoOperands(t)
	} else if continuation {
		d.loadScoredDrainOperands(t)
	} else {
		d.loadHighScoreOperands(t)
	}
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
	// Reference uses the same verified native factory policy, independently
	// through its ordinary consumer. It never supplies candidate awards/state.
	reference.SetHighScore(Number(50000000))
	counts := map[string]int{}
	saved := loadDemoSavedWitness(t)
	limit := 35877
	if continuation {
		limit = 35998
	}
	collision := loadDemoCollisionWitness(t, continuation)
	for n := 1; n <= limit; n++ {
		if err := reference.Sync(demoFixedInput(n)); err != nil {
			t.Fatal(err)
		}
		if err := d.sync(demoFixedInput(n), true); err != nil {
			if continuation {
				t.Fatalf("continuation boundary %d: %v", n, err)
			}
			if n != 35877 || d.failure.Producer != "LOOSE_BALL" || d.failure.Consumer != "scored drain" || d.failure.Phase != "drain handoff" || d.failure.Calculation != 35877 || g.Physics.Syncs != 35876 || d.counter != 35876 || d.oldExpired || d.expired || !d.loosing || !d.holdStill || g.Bonus != Number(0) || g.Score != Number(50030) || !g.ScoreChanged {
				t.Fatalf("wrong next boundary at %d: %v", n, err)
			}
			if d.snapshots[0].Ball.PixelY < 576 || !d.snapshots[0].ScoreChanged || !d.snapshots[0].Lost || len(d.snapshots) != 2 {
				t.Fatal("real scored drain", d.snapshots)
			}
			for _, event := range g.Events {
				if event.Label == "S_LOSTBALL" || event.Kind == "HighScoreBeaten" {
					t.Fatal("unsupported drain/reward effects", event)
				}
			}
			t.Logf("NEXT boundary %d: %v ball=%+v score=%s changed=%t syncs=%d", n, err, g.Physics.Ball, g.Score.String(), g.ScoreChanged, g.Physics.Syncs)
			assertConnectedSticky(t, d, err)

			break
		}

		if continuation && n >= 35877 {
			compareDemoCollisionWitness(t, collision, n, d)
			if g.Random != uint8(35876%256) {
				t.Fatal("source UPDATE_COUNTERS recurrence", n, g.Random)
			}
			if n == 35877 && (d.drains != 1 || d.oldExpired || d.expired || len(d.snapshots) < 3 || d.snapshots[0].Ball.PixelY < 576 || !g.ScoreChanged || g.Score != Number(50030)) {
				t.Fatal("real drain")
			}
			if n == 35967 && (len(d.bonusNodes) != 5 || g.tasks[0] == nil || g.waitCounters["NEW_BALL_TASK"] != 0 || g.Session.Load().Score != Number(50030)) {
				t.Fatal("source producer")
			}
			if n == 35998 {
				assertDemoEquality(t, d)
				break
			}
			if n >= 35967 {
				continue
			} // A's full-game progression differs here; saved linked witness covers every later calculation.
			// The native reference keeps its own Hold during execution.
		}

		refBall := reference.Physics.Ball
		if continuation && n >= 35877 {
			refBall.Hold = false
		}
		if g.Physics.Ball != refBall || ((!continuation || n < 35877) && (!reflect.DeepEqual(g.Physics.Flippers, reference.Physics.Flippers) || g.Physics.SpringPosition != reference.Physics.SpringPosition || g.Physics.ScreenOffset != reference.Physics.ScreenOffset || g.Physics.Raster != reference.Physics.Raster || g.Physics.ScreenSpeed != reference.Physics.ScreenSpeed || g.Physics.ScreenPosition != reference.Physics.ScreenPosition)) {
			t.Fatalf("reference trajectory mismatch %d ball=%+v reference=%+v", n, g.Physics.Ball, refBall)
		}
		if d.counter != uint16(n) || g.Random != reference.Random || g.clock != reference.clock || g.Score != reference.Score || g.Bonus != reference.Bonus || g.ScoreChanged != reference.ScoreChanged || g.SkillTime != reference.SkillTime || g.InhibitReverseTime != reference.InhibitReverseTime || g.lastArea != reference.lastArea || g.lastCheck != reference.lastCheck || g.Lights != reference.Lights || g.touchDisabled != reference.touchDisabled || g.Arcade != reference.Arcade || g.Lamps != reference.Lamps || g.flashes != reference.flashes || g.Audio != reference.Audio || !reflect.DeepEqual(g.waitCounters, reference.waitCounters) {
			t.Fatalf("electronics mismatch %d clock=%d/%d random=%d/%d score=%s/%s waits=%v/%v audio=%+v/%+v area=%s/%s check=%s/%s flags=%v/%v", n, g.clock, reference.clock, g.Random, reference.Random, g.Score.String(), reference.Score.String(), g.waitCounters, reference.waitCounters, g.Audio, reference.Audio, g.lastArea, reference.lastArea, g.lastCheck, reference.lastCheck, []bool{g.touchDisabled, g.Arcade}, []bool{reference.touchDisabled, reference.Arcade})
		}
		if g.matrix.active != reference.matrix.active || demoReferenceMatrixOp(g.matrix.op) != reference.matrix.op || g.matrix.remaining != reference.matrix.remaining || g.matrix.frame != reference.matrix.frame || g.matrix.frameTime != reference.matrix.frameTime || g.matrix.loops != reference.matrix.loops || g.alreadyBeaten != reference.alreadyBeaten || g.inChute != reference.inChute || g.Happy != reference.Happy || g.Mega != reference.Mega || g.inhibitEffect != reference.inhibitEffect || ((!continuation || n < 35877) && g.Physics.AllowFlip != reference.Physics.AllowFlip) || g.PukeForbidden != reference.PukeForbidden {
			t.Fatalf("matrix/flags mismatch %d candidate=%+v reference=%+v flags=%v/%v", n, g.matrix, reference.matrix, []bool{g.alreadyBeaten, g.inChute, g.Happy, g.Mega, g.inhibitEffect, g.Physics.AllowFlip, g.PukeForbidden}, []bool{reference.alreadyBeaten, reference.inChute, reference.Happy, reference.Mega, reference.inhibitEffect, reference.Physics.AllowFlip, reference.PukeForbidden})
		}
		if d.ownedMatrix && (!continuation || n < 35877) {
			label := "PARTY_OFFTS"
			if g.Display.Content.Commands[1].Op == "_PRINT5" {
				label = "SHOWPLAYERSTS"
			}
			base, ok := timing.Labels[label]
			if reference.matrix.sourceProgram {
				base, ok = reference.Display.Content.Labels[label]
			}
			if !ok || g.matrix.next != reference.matrix.next-base {
				t.Fatal("normalized matrix cursor", n, label, g.matrix.next, reference.matrix.next)
			}
		}

		for i, task := range g.tasks {
			if (task == nil) != (reference.tasks[i] == nil) || g.taskIDs[i] != reference.taskIDs[i] {
				t.Fatal("task scan mismatch", n, i)
			}
		}
		if n == 35789 && (g.Score != Number(0) || g.Physics.Syncs != 35789) {
			t.Fatal("preaward checkpoint")
		}
		if n == 35790 {
			if g.Score != Number(50030) || g.Score.String() != "000000050030" || !g.ScoreChanged || g.Bonus != Number(0) || g.Physics.Syncs != 35790 || g.Physics.Ball.X != 4325 || g.Physics.Ball.Y != 459054 || g.Physics.Ball.VY != 794 || g.Physics.Ball.Rotation != 1580 || d.infoCount != 1 || g.matrix.active || g.matrix.op != "0" || g.matrix.next != 4 || g.alreadyBeaten {
				t.Fatal("completed award and first late physics")
			}
			if len(g.Events) < 3 || g.Events[0].Label != "BYGEL1" || g.Events[1].Label != "SBYGEL1" || g.Events[2].Kind != "ScoreAwarded" || g.Events[2].Value != 50030 {
				t.Fatal("award order")
			}
		}
		if n < 35877 {
			compareDemoSavedWitness(t, saved, n, g)
		}
		if n == 35712 || n == 35713 || n == 35731 || n == 35732 || n == 35790 {
			t.Logf("checkpoint %d ball=%+v score=%s bonus=%s changed=%t touch=%t arcade=%t tasks=%v waits=%v fires=%v flashes=%v events=%v matrix=%+v", n, g.Physics.Ball, g.Score.String(), g.Bonus.String(), g.ScoreChanged, g.touchDisabled, g.Arcade, g.taskIDs, g.waitCounters, d.toucherFires, g.flashes, g.Events, g.matrix)
		}
		if n >= 35712 && n < 35732 && (!g.touchDisabled || g.tasks[0] == nil || g.waitCounters["ENABLETOUCHER"] != uint16(n-35711)) {
			t.Fatal("fresh task age", n)
		}
		if n >= 35732 && n < 35877 && (g.touchDisabled || g.tasks[0] != nil || !reflect.DeepEqual(d.toucherFires, []uint16{35732})) {
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
	if continuation {
		limit := 38000
		if showInfo {
			limit = 50000
		}
		for n := 35999; n <= limit; n++ {
			beforeSyncs := g.Physics.Syncs
			oldMatrix, oldDots, oldWaits := g.matrix, g.Display.Dots, fmt.Sprint(g.waitCounters)
			expectedMusic := g.musicClock()
			expectedMusic.Sync(timing.Cues, func(string, uint64) {})
			if showInfo && n == 36725 && (d.infoCount != 720 || !g.Physics.SpringValid || !d.expired || g.matrix.active) {
				t.Fatal("actual SHOW_HI_ETC entry")
			}

			err := d.sync(demoFixedInput(n), true)
			if err != nil {
				t.Logf("NEXT unsupported calculation=%d completed=%d: %v", n, beforeSyncs, err)
				if d.failure == nil || g.Physics.Syncs != beforeSyncs {
					t.Fatal("failed calculation completed late physics")
				}
				assertConnectedSticky(t, d, err)
				break
			}
			if g.Physics.Syncs != uint64(n) || d.counter != uint16(n) || d.drains != 1 || g.Random != uint8((35876+n-35998)%256) || g.clock != uint16((uint64(n)*1030)%65536) || g.Score != Number(50030) || g.Bonus != Number(0) {
				t.Fatal("post equality completed calculation", n)
			}
			if showInfo && n == 36725 && (d.infoCount != 721 || !d.expired || d.infoStarts != 0 || !reflect.DeepEqual(oldMatrix, g.matrix) || oldDots != g.Display.Dots || expectedMusic != g.musicClock() || oldWaits != fmt.Sprint(g.waitCounters)) {
				t.Fatal("spring guard source writes")
			}
			if showInfo && (n == 36725 || (d.infoStarts > 0 && !g.matrix.active && n < 39000)) {
				t.Logf("INFO checkpoint calculation=%d starts=%d count=%d active=%t cursor=%d op=%s spring=%t expired=%t hold=%t jackpot=%d", n, d.infoStarts, d.infoCount, g.matrix.active, g.matrix.next, g.matrix.op, g.Physics.SpringValid, d.expired, d.holdStill, g.Jackpot.Uint64())
			}
			if n == limit {
				if showInfo {
					t.Logf("BOUNDED supported through %d; info starts=%d; count=%d; next unsupported NOT REACHED", n, d.infoStarts, d.infoCount)
					break
				}
				t.Fatal("no next unsupported consumer in bounded continuation")
			}
		}
	}

	if showInfo && (d.failure != nil || g.Physics.Syncs != 50000 || d.infoStarts != 0 || d.infoCount != 13996 || !d.expired) {
		t.Fatal("fresh guarded continuation", d.failure, g.Physics.Syncs, d.infoStarts, d.infoCount)
	}
	if !continuation && d.infoCount != 87 {
		t.Fatal("scored idle count", d.infoCount)
	}
	if (!continuation && d.failure == nil) || (continuation && !showInfo && d.failure == nil) || g.Score != Number(50030) || g.Bonus != Number(0) || (!continuation && !g.ScoreChanged) || g.alreadyBeaten {
		t.Fatal("missing boundary or lost award")
	}
	if !reflect.DeepEqual(counts, map[string]int{"BYGEL12": 2, "BYGEL28": 1, "CLOSE1": 1, "BYGEL9": 1, "BYGEL11": 1, "BYGEL1": 1}) {
		t.Fatal("fixed replay callbacks", counts)
	}
}

func demoReferenceMatrixOp(op string) string {
	switch op {
	case "_DEMO_PARTY_FLASHOFF":
		return "_FLASHOFF"

	}
	return op
}
