package presentation

import (
	"reflect"
	"testing"
)

func TestPlayerBallPreservesLiveCommand(t *testing.T) {
	for table := 1; table <= 4; table++ {
		d := original(t, table)
		d.SetPlayers(2, 3, 1)
		d.Begin("_FLASHON", []string{"1"})
		d.Visit(0, 0, nil)
		d.Flash()
		before := *d
		d.ShowPlayerBall("000000123450")
		after := *d
		after.Dots = before.Dots
		after.On = before.On
		after.flashing = before.flashing
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("table %d idle handoff mutated live command state", table)
		}
		for tick := 0; tick < 710; tick++ {
			d.Flash()
			if !d.On {
				t.Fatalf("table %d idle handoff left illumination off/flashing", table)
			}
		}
		first := d.Dots
		d.SetPlayers(3, 3, 1)
		d.ShowPlayerBall("000000000000")
		if d.Dots == first || d.Dots == [DotWidth * DotHeight]bool{} {
			t.Fatalf("table %d did not repaint zero-score player", table)
		}
	}
}
