package gamepad

// Snapshot is standardized controller state for polling hosts (XInput).
type Snapshot struct {
	Buttons                   uint16
	LeftTrigger, RightTrigger int
}
type Event struct{ Kind, A, B int }
type Tracker struct {
	connected bool
	previous  Snapshot
}

func (t *Tracker) Disconnect() []Event {
	if !t.connected {
		return nil
	}
	t.connected = false
	return []Event{{Kind: 3}}
}
func (t *Tracker) Sample(s Snapshot) []Event {
	if !t.connected {
		t.connected = true
		t.previous = s
		mask := int(s.Buttons)
		if s.LeftTrigger > 12000 {
			mask |= 1 << 15
		}
		if s.RightTrigger > 12000 {
			mask |= 1 << 16
		}
		return []Event{{Kind: 2, A: mask}}
	}
	var events []Event
	for b := 0; b < 15; b++ {
		if (s.Buttons^t.previous.Buttons)&(1<<b) != 0 {
			value := 0
			if s.Buttons&(1<<b) != 0 {
				value = 1
			}
			events = append(events, Event{Kind: 0, A: b, B: value})
		}
	}
	if s.LeftTrigger != t.previous.LeftTrigger {
		events = append(events, Event{Kind: 1, A: 4, B: s.LeftTrigger})
	}
	if s.RightTrigger != t.previous.RightTrigger {
		events = append(events, Event{Kind: 1, A: 5, B: s.RightTrigger})
	}
	t.previous = s
	return events
}

func XInputSnapshot(buttons uint16, left, right uint8) Snapshot {
	// XINPUT_GAMEPAD bit identities are unrelated to SDL button indices.
	bits := [15]uint16{0x1000, 0x2000, 0x4000, 0x8000, 0x20, 0, 0x10, 0x40, 0x80, 0x100, 0x200, 1, 2, 4, 8}
	var standard uint16
	for i, bit := range bits {
		if buttons&bit != 0 {
			standard |= 1 << i
		}
	}
	return Snapshot{standard, int(left) * 32767 / 255, int(right) * 32767 / 255}
}
