//go:build dmoimpl1

package partyland

import (
	"fmt"
	"pinballfantasies/internal/physics"
	"reflect"
	"strings"
	"testing"
)

// demoNativeCandidate deliberately has no Game.Sync fallback, no configurable
// allowlist and no public host entry. Aggregate canonical gameplay consumers
// are never admitted. The shared early physics operations are the only envelope.
type demoNativeCandidate struct {
	game        *Game
	failure     *demoUnsupported
	calculation uint64
	admitted    []string
}

func (d *demoNativeCandidate) gate(producer, consumer string) error {
	if d.failure != nil {
		return d.failure
	}
	switch consumer {
	case "physics.Sync", "physics.step", "checkRamps/checkLevels":
		d.admitted = append(d.admitted, consumer)
		return nil
	}
	d.failure = &demoUnsupported{Producer: producer, Consumer: consumer,
		Calculation: d.calculation,
		Guard:       fmt.Sprintf("phase=%d lost=%t hold=%t high=%t", d.game.Phase, d.game.Physics.Ball.Lost, d.game.Physics.Ball.Hold, d.game.Physics.Ball.High),
		Reason:      "aggregate canonical gameplay consumer has no admitted demo implementation"}
	return d.failure
}

func (d *demoNativeCandidate) sync(input physics.Inputs) error {
	if d.failure != nil {
		return d.failure
	}
	d.calculation++
	err := d.game.Physics.SyncWithGate(input, d.gate)
	if err != nil && d.failure == nil {
		d.failure = &demoUnsupported{Producer: "physics.Sync", Consumer: "physics arithmetic", Calculation: d.calculation, Guard: "native error", Reason: err.Error()}
	}
	if d.failure != nil {
		return d.failure
	}
	return nil
}

func TestDemoNativeGateBeforeTargets(t *testing.T) {
	g := newTestGame(t)
	canonical := newTestGame(t)
	d := &demoNativeCandidate{game: g}
	random, tick, score := g.Random, g.Tick, g.Score
	events := append([]Event(nil), g.Events...)
	err := d.sync(physics.Inputs{}) // actual initial native physics, no teleport/hold
	if err == nil || d.failure.Consumer != "BeforeTargets" {
		t.Fatalf("wrong boundary: %v", err)
	}
	if !reflect.DeepEqual(d.admitted, []string{"physics.Sync", "physics.step", "physics.step", "checkRamps/checkLevels"}) {
		t.Fatal(d.admitted)
	}
	if g.Random != random || g.Tick != tick || g.Score != score || !reflect.DeepEqual(g.Events, events) || g.Physics.Syncs != 0 {
		t.Fatal("blocked electronics or suffix executed")
	}
	if g.Physics.Ball == canonical.Physics.Ball {
		t.Fatal("admitted early physics did not execute")
	}
	for _, field := range []string{"UNSUPPORTED_DEMO_TRANSITION", "producer=physics.Sync", "consumer=BeforeTargets", "state/guard=", "calculation=1", "reason="} {
		if !strings.Contains(err.Error(), field) {
			t.Fatal(err)
		}
	}
	before := fmt.Sprintf("%#v %#v", g, g.Physics)
	for i := 0; i < 3; i++ {
		if d.sync(physics.Inputs{Left: true}) != err {
			t.Fatal("not sticky")
		}
	}
	if before != fmt.Sprintf("%#v %#v", g, g.Physics) || d.calculation != 1 {
		t.Fatal("post-failure mutation")
	}
	if err := canonical.Physics.Sync(physics.Inputs{}); err != nil {
		t.Fatal(err)
	}
	if canonical.Random != random+1 || canonical.Physics.Syncs != 1 {
		t.Fatal("canonical sequence restricted")
	}
}

func TestDemoNativeGateDrainBeforeCallback(t *testing.T) {
	g := newTestGame(t)
	// Structural drain-boundary test, explicitly not a reachable demo witness.
	g.Physics.SetBall(100, 576, 0, 0, false)
	d := &demoNativeCandidate{game: g}
	before := g.Score
	if err := d.sync(physics.Inputs{}); err == nil || d.failure.Consumer != "drain/OnEvent" {
		t.Fatalf("%v", err)
	}
	if g.Phase != Playing || g.Physics.Stopped || len(g.Physics.Events) != 0 || g.Score != before {
		t.Fatal("drain side effects began")
	}
	for _, task := range g.tasks {
		if task != nil {
			t.Fatal("drain task inserted")
		}
	}
}

func TestDemoNativeGateRejectsAllGameplayAggregates(t *testing.T) {
	for _, consumer := range []string{"OnEvent", "BeforeTargets", "checkSpringAndTargets/OnEvent", "AfterTargets", "BeforeLate", "scroll/ScrollForce", "unknown"} {
		d := &demoNativeCandidate{game: &Game{Physics: &physics.Game{}}}
		if d.gate("test", consumer) == nil {
			t.Fatal(consumer)
		}
	}
}

func TestDemoGateStickyMatrixEntry(t *testing.T) {
	d := newDemoCore()
	demoOK(t, d.install(demoExpiryProgram()))
	err := d.unsupported("area", "unadmitted")
	before := d.game.matrix
	if d.matrixVisit() != err || !reflect.DeepEqual(d.game.matrix, before) {
		t.Fatal("matrix advanced after sticky error")
	}
}
