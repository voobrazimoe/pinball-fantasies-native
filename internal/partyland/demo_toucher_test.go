//go:build dmoimpl1

package partyland

import (
	"fmt"
	"os/exec"
	"pinballfantasies/internal/physics"
	"reflect"
	"testing"
)

func (d *demoConnected) loadToucherOperands(t *testing.T) {
	t.Helper()
	d.loadGameplayOperands(t) // pinned identities and predecessor admission
	cmd := exec.Command("python3", "../../tools/check_demo_2b9_consumers.py")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("TOUCHER admission: %v %s", err, out)
	}
	d.toucherOperands = true
}

// Structural tests: these injected states are not the reachability proof.
func TestDemoToucherStructural(t *testing.T) {
	for _, arcade := range []bool{false, true} {
		d := connected(newTestGame(t))
		d.loadToucherOperands(t)
		g := d.game
		g.Arcade = arcade
		audio, score, bonus := g.Audio, g.Score, g.Bonus
		demoOK(t, d.consumeTarget(0))
		if !g.touchDisabled || !g.Arcade || g.tasks[0] == nil || g.waitCounters["ENABLETOUCHER"] != 0 || g.Audio != audio || g.Score != score || g.Bonus != bonus || g.ScoreChanged || g.Events[len(g.Events)-1].Label != "S_TOUCH2" {
			t.Fatal("TOUCHER effects")
		}
		if !arcade && (g.flashes[2] != (flash{7, 0, 12}) || g.flashes[3] != (flash{55, 0, 12})) {
			t.Fatal("flash operands", g.flashes)
		}
		before := fmt.Sprintf("%#v", g)
		demoOK(t, d.consumeTarget(0))
		if before != fmt.Sprintf("%#v", g) {
			t.Fatal("repeat not inhibited")
		}
		id := g.taskIDs[0]
		for age := 1; age <= 20; age++ {
			g.runTasks()
			if g.waitCounters["ENABLETOUCHER"] != uint16(age) || g.tasks[0] == nil || g.taskIDs[0] != id || !g.touchDisabled {
				t.Fatal("compare before increment", age)
			}
		}
		g.runTasks()
		if g.tasks[0] != nil || g.touchDisabled || !g.Arcade || g.waitCounters["ENABLETOUCHER"] != 0 {
			t.Fatal("ENABLETOUCHER body/suicide")
		}
	}
}
func TestDemoToucherNestedFailure(t *testing.T) {
	for _, due := range []bool{false, true} {
		d := connected(newTestGame(t))
		d.loadToucherOperands(t)
		g := d.game
		if due {
			demoOK(t, d.consumeTarget(0))
			g.waitCounters["ENABLETOUCHER"] = 20
			d.toucherOperands = false
		} else {
			for i := range g.tasks {
				g.tasks[i] = func() bool { return false }
			}
		}
		before := fmt.Sprintf("%#v %#v", g, g.Physics)
		if due {
			g.runTasks()
		} else {
			_ = d.consumeTarget(0)
		}
		if d.failure == nil || before != fmt.Sprintf("%#v %#v", g, g.Physics) {
			t.Fatal("nested effects before refusal")
		}
		err := d.failure
		if d.consumeTarget(0) != err || d.preflightTarget(0) != err {
			t.Fatal("sticky direct callback")
		}
		assertConnectedSticky(t, d, err)
	}
}
func TestDemoToucherUnknownTarget(t *testing.T) {
	d := connected(newTestGame(t))
	d.loadToucherOperands(t)
	before := d.game.Physics.Ball
	if d.preflightTarget(1) == nil || !reflect.DeepEqual(before, d.game.Physics.Ball) {
		t.Fatal("unadmitted target")
	}
}

func TestDemoToucherTargetConsumption(t *testing.T) {
	for _, supported := range []bool{false, true} {
		d := connected(newTestGame(t))
		d.loadToucherOperands(t)
		g := d.game
		g.Physics.OnEvent = func(physics.Event) { t.Fatal("canonical target callback") }
		// Structural contact; reachability is tested by the fixed fresh script.
		g.Physics.Ball.HitX, g.Physics.Ball.HitY = 132, 199-g.Physics.ScreenOffset
		if !supported {
			for i := range g.tasks {
				g.tasks[i] = func() bool { return false }
			}
		}
		before := fmt.Sprintf("%#v %#v", g, g.Physics)
		err := g.Physics.CandidateTargets(d.preflightTarget, d.consumeTarget)
		if !supported {
			if err == nil || before != fmt.Sprintf("%#v %#v", g, g.Physics) {
				t.Fatal("target writes before nested admission")
			}
			assertConnectedSticky(t, d, err)
			continue
		}
		demoOK(t, err)
		if g.Physics.Ball.HitX != 0 || g.Physics.Ball.HitY != 0 || len(g.Physics.Events) != 1 || g.Physics.Events[0].Kind != physics.EventTargetHit || !g.touchDisabled {
			t.Fatal("source contact/event completion")
		}
		// Another physical contact while disabled consumes contact but makes no
		// second sound, flash, task or callback-body trace, even with all slots full.
		for i := 1; i < len(g.tasks); i++ {
			g.tasks[i] = func() bool { return false }
		}
		events, trace, ids, flashes := len(g.Events), len(d.trace), g.taskIDs, g.flashes
		g.Physics.Ball.HitX, g.Physics.Ball.HitY = 132, 199-g.Physics.ScreenOffset
		demoOK(t, g.Physics.CandidateTargets(d.preflightTarget, d.consumeTarget))
		if len(g.Events) != events || len(d.trace) != trace || g.taskIDs != ids || g.flashes != flashes || g.Physics.Ball.HitX != 0 {
			t.Fatal("inhibited physical contact repeated consumer effects")
		}
	}
}
