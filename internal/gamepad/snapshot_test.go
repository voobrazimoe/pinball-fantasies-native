package gamepad

import "testing"

func TestXInputButtonsAndTriggers(t *testing.T) {
	s := XInputSnapshot(0xffff, 255, 0)
	if s.Buttons != 0x7fdf || s.LeftTrigger != 32767 || s.RightTrigger != 0 {
		t.Fatalf("XInput mapping: %+v", s)
	}
	for _, c := range []struct {
		native   uint16
		standard int
	}{{0x1000, A}, {0x2000, B}, {0x4000, X}, {0x8000, Y}, {0x100, LeftShoulder}, {0x200, RightShoulder}, {1, Up}, {2, Down}, {4, Left}, {8, Right}, {0x10, Start}, {0x20, Back}} {
		if XInputSnapshot(c.native, 0, 0).Buttons != 1<<c.standard {
			t.Fatal("button mismatch", c)
		}
	}
}
func TestPollingConnectionAndEdges(t *testing.T) {
	var tracker Tracker
	s := Snapshot{Buttons: 1 << X, LeftTrigger: 20000}
	first := tracker.Sample(s)
	if len(first) != 1 || first[0].Kind != 2 || first[0].A != 1<<X|1<<15 {
		t.Fatal("initial holds not seeded", first)
	}
	if len(tracker.Sample(s)) != 0 {
		t.Fatal("poll repeated make")
	}
	events := tracker.Sample(Snapshot{})
	if len(events) != 2 || events[0] != (Event{Kind: 0, A: X, B: 0}) || events[1] != (Event{Kind: 1, A: 4, B: 0}) {
		t.Fatal("lost release", events)
	}
	if len(tracker.Disconnect()) != 1 || len(tracker.Disconnect()) != 0 {
		t.Fatal("disconnect repeated")
	}
	if tracker.Sample(s)[0].Kind != 2 {
		t.Fatal("reconnect lacked baseline")
	}
}
