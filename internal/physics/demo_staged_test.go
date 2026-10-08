//go:build dmoimpl1

package physics

import (
	"errors"
	"reflect"
	"testing"
)

func TestDemoLateCollisionGateBeforeWrites(t *testing.T) {
	g := New(table(t))
	// Explicit structural bumper fixture, not a reachable input-only prefix.
	g.SetBall(211, 238, 0, 0, false)
	probe := *g
	probe.Events = nil
	probe.collision()
	if !probe.pendingBumper {
		t.Fatal("structural bumper fixture drift")
	}
	sentinel := errors.New("unsupported event")
	before := *g
	err := g.CandidateStage("late.step", Inputs{}, false, func(_, consumer string) error {
		if consumer == "collision/OnEvent" {
			return sentinel
		}
		return nil
	})
	if err != sentinel || !reflect.DeepEqual(before, *g) {
		t.Fatal("collision effects before gate", err)
	}
}
