//go:build dmoimpl1

package partyland

import (
	"fmt"
	"pinballfantasies/internal/physics"

	"reflect"
	"testing"
)

func connectedQuiet(t *testing.T) *demoConnected {
	g := newTestGame(t)
	// Structural execution fixture, not a reachable demo prefix/witness.
	g.Physics.SetBall(100, 400, 100, 0, false)
	return connected(g)
}

func TestDemoConnectedCalculation(t *testing.T) {
	d := connectedQuiet(t)
	before := d.game.Physics.Ball
	demoOK(t, d.sync(physics.Inputs{}, false))
	if d.counter != 1 || d.game.Random != 1 || d.game.Physics.Ball == before {
		t.Fatal("missing real calculation")
	}
	if !reflect.DeepEqual(d.order, []string{"early.step", "early.step", "early.finish", "ElectronicsCalculation", "targets", "task/matrix", "late.step", "complete"}) {
		t.Fatal(d.order)
	}
	if !reflect.DeepEqual(d.trace, []string{"UPDATE_COUNTERS", "timer", "electronics-no-callback", "KEYTASK-empty", "DO_TASKS", "matrix-budget"}) {
		t.Fatal(d.trace)
	}
}

func assertConnectedSticky(t *testing.T, d *demoConnected, err error) {
	t.Helper()
	before := fmt.Sprintf("%#v %#v %#v %#v", d, d.demoCore, d.game, d.game.Physics)
	for _, got := range []error{d.loadResetOperands(nil), d.preflightChild("SETBALL"), d.preflightNewBall(), d.loadPartyPresentation(nil, nil), d.looseBall(), d.electronicsPrefix(), d.taskMatrixSuffix(true), d.matrixVisit(), d.stage("late.step", physics.Inputs{}), d.queue(demoTask{Site: "after-failure"}), d.install(demoWait(1))} {
		if got != err {
			t.Fatal("direct consumer bypassed sticky failure")
		}
	}
	for i := 0; i < 3; i++ {
		if d.sync(physics.Inputs{Left: true}, true) != err {
			t.Fatal("not sticky")
		}
	}
	if before != fmt.Sprintf("%#v %#v %#v %#v", d, d.demoCore, d.game, d.game.Physics) {
		t.Fatal("post-failure mutation")
	}
}

func TestDemoConnectedHold(t *testing.T) {
	d := connectedQuiet(t)
	d.counter = 35997
	// Compare the expiry calculation to exactly its two shared early steps.
	control := connectedQuiet(t)
	demoOK(t, control.stage("early.step", physics.Inputs{}))
	demoOK(t, control.stage("early.step", physics.Inputs{}))
	demoOK(t, control.stage("early.finish", physics.Inputs{}))
	demoOK(t, d.sync(physics.Inputs{}, false))
	if !d.holdStill || d.game.Physics.Ball.Hold || d.game.Physics.Ball != control.game.Physics.Ball {
		t.Fatal("expiry did not hold late physics")
	}
	before := d.game.Physics.Ball
	d.game.Physics.Ball.Hold = true
	d.game.Physics.Ball.Hold = false // native capture release
	demoOK(t, d.sync(physics.Inputs{}, false))
	if !d.holdStill || d.game.Physics.Ball != before {
		t.Fatal("native release cleared source hold")
	}
	demoOK(t, d.queue(demoTask{Site: "source-release", Action: demoReleaseSourceHold}))
	demoOK(t, d.sync(physics.Inputs{}, false))
	if d.holdStill || d.game.Physics.Ball == before || d.counter != 36000 {
		t.Fatal("typed source release did not open late movement")
	}
}

func TestDemoConnectedReject(t *testing.T) {
	for _, kind := range []string{"area", "target", "counter", "nested-task", "unowned-task", "unowned-matrix"} {
		t.Run(kind, func(t *testing.T) {
			d := connectedQuiet(t)
			d.game.Physics.Ball.Hold = true
			switch kind {
			case "area":
				d.game.Physics.SetBall(100, 20, 0, 0, false)
			case "target":
				d.game.Physics.Ball.HitX, d.game.Physics.Ball.HitY = 150, 283
			case "counter":
				d.game.TunnelTime = 721
			case "nested-task":
				demoOK(t, d.queue(demoTask{Site: "bad", Action: demoReplaceMatrix, Program: demoExpiryProgram()[1:]}))
			case "unowned-task":
				d.game.task(func() bool { t.Fatal("canonical task invoked"); return true })
			case "unowned-matrix":
				demoOK(t, d.install(demoWait(10)))
				d.ownedMatrix = false
			}
			score := d.game.Score
			events := append([]Event(nil), d.game.Events...)
			before := d.game.Physics.Ball
			err := d.sync(physics.Inputs{}, false)
			if err == nil {
				t.Fatal("unsupported consumer admitted")
			}
			if d.game.lastCheck != "" || !reflect.DeepEqual(d.game.Events, events) || len(d.game.Physics.Events) != 0 || d.game.Score != score {
				t.Fatal("callback effects")
			}
			if kind == "target" && (d.game.Physics.Ball.HitX != before.HitX || d.game.Physics.Ball.HitY != before.HitY) {
				t.Fatal("target input consumed")
			}
			if kind == "counter" && (d.counter != 0 || d.game.TunnelTime != 721 || d.game.Random != 0) {
				t.Fatal("nested counter effects")
			}
			if kind == "nested-task" && d.game.waitCounters["bad"] != 0 {
				t.Fatal("due WAIT mutated before rejection")
			}
			assertConnectedSticky(t, d, err)
		})
	}
}

func TestDemoConnectedTypedMatrixAndCapture(t *testing.T) {
	d := connectedQuiet(t)
	demoOK(t, d.install(demoWait(10)))
	demoOK(t, d.queue(demoTask{Site: "preserve", Action: demoPreserve}))
	demoOK(t, d.sync(physics.Inputs{}, true))
	if d.game.tasks[0] != nil || d.game.matrix.remaining != 9 || d.counter != 1 {
		t.Fatal("typed suffix not connected")
	}
	d.holdStill, d.game.Physics.Ball.Hold = true, true
	before := d.game.Physics.Ball
	demoOK(t, d.queue(demoTask{Site: "release", Action: demoReleaseSourceHold}))
	demoOK(t, d.sync(physics.Inputs{}, false))
	if d.holdStill || !d.game.Physics.Ball.Hold || d.game.Physics.Ball != before {
		t.Fatal("source release changed native capture")
	}
}

func TestDemoConnectedMatrixRejectBeforeLate(t *testing.T) {
	d := connectedQuiet(t)
	d.holdStill = true
	demoOK(t, d.install(demoExpiryProgram()))
	for d.game.matrix.remaining > 1 {
		demoOK(t, d.sync(physics.Inputs{}, true))
	}
	before, timer := d.game.matrix, d.counter
	err := d.sync(physics.Inputs{}, true)
	if err == nil || d.failure.Consumer != "NEXT_A" || d.failure.Phase != "task/matrix" || !reflect.DeepEqual(d.game.matrix, before) || d.counter != timer+1 {
		t.Fatal("matrix dispatch boundary", err)
	}
	if d.order[len(d.order)-1] != "task/matrix" {
		t.Fatal("late continuation after rejected matrix")
	}
	assertConnectedSticky(t, d, err)
	if d.electronicsPrefix() != err || d.taskMatrixSuffix(true) != err || d.stage("late.step", physics.Inputs{}) != err {
		t.Fatal("direct sticky entry")
	}
	if d.counter != timer+1 {
		t.Fatal("direct post-failure counting")
	}
}

func TestDemoConnectedNestedFailureStopsTaskSuffix(t *testing.T) {
	d := connectedQuiet(t)
	d.holdStill = true
	demoOK(t, d.queue(demoTask{Site: "bad", Delay: 1, Action: demoReplaceMatrix, Program: demoExpiryProgram()[1:]}))
	demoOK(t, d.queue(demoTask{Site: "later", Delay: 1, Action: demoReleaseSourceHold}))
	demoOK(t, d.sync(physics.Inputs{}, false))
	ids := d.game.taskIDs
	err := d.sync(physics.Inputs{}, false)
	if err == nil || !d.holdStill || d.game.waitCounters["bad"] != 1 || d.game.waitCounters["later"] != 1 || d.game.taskIDs != ids || d.game.tasks[0] == nil || d.game.tasks[1] == nil {
		t.Fatal("nested failure continued shared task scan effects")
	}
	assertConnectedSticky(t, d, err)
}
