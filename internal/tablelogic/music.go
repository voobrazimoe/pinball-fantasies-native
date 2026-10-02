package tablelogic

import "strconv"

type JingleSpec struct{ Position, Repeat, Priority uint8 }
type CueSpec struct {
	TrackerTicks uint32 `json:"tracker_ticks"`
	Next         uint8
	Segments     []struct {
		Position uint8
		Start    uint32
	}
}

// MusicClock is FANTASIE's cooperative jingle state, measured in 1/3550s.
// It is independent of queued host audio; the tracker only renders samples.
type MusicClock struct {
	Position, ReturnPosition, Priority, LastPriority, JumpCount, Entry uint8
	ReadyAnim, ReadyLogic, Active                                      bool
	Elapsed, Cue                                                       uint32
}

func (a *MusicClock) Play(s JingleSpec, empty uint8, cues map[string]CueSpec) bool {
	oldRepeat := a.JumpCount
	a.JumpCount = s.Repeat
	if s.Priority < a.Priority {
		return false
	}
	if int8(oldRepeat) <= 0 && a.Position != empty {
		a.ReturnPosition = a.Position
	}
	a.Position, a.Entry = s.Position, s.Position
	a.Priority = s.Priority
	a.ReadyAnim, a.ReadyLogic = false, false
	a.Elapsed = 0
	a.Active = true
	a.Cue = cues[strconv.Itoa(int(s.Position))].TrackerTicks * 71
	return true
}
func (a *MusicClock) Sync(cues map[string]CueSpec, emit func(string, uint64)) {
	if !a.Active {
		return
	}
	a.Elapsed += 50
	for a.Active {
		c, ok := cues[strconv.Itoa(int(a.Entry))]
		if !ok || c.TrackerTicks == 0 {
			a.Active = false
			return
		}
		for _, s := range c.Segments {
			if a.Elapsed >= s.Start*71 {
				a.Position = s.Position
			}
		}
		if a.Elapsed < a.Cue {
			return
		}
		a.Elapsed -= a.Cue
		next := c.Next
		a.JumpCount--
		emit("AudioCue", uint64(a.Position))
		if a.JumpCount == 0 {
			a.ReadyAnim, a.ReadyLogic = true, true
			a.Priority = a.LastPriority
			emit("AudioComplete", uint64(a.Position))
			next = a.ReturnPosition
		} else if int8(a.JumpCount) < 0 {
			a.JumpCount = 0
		}
		a.Entry, a.Position = next, next
		n, ok := cues[strconv.Itoa(int(next))]
		if !ok {
			a.Active = false
			return
		}
		a.Cue = n.TrackerTicks * 71
	}
}
