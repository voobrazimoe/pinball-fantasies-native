package source

import (
	"bytes"
	"image"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/tablelogic"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"testing"
	"time"
)

type session struct{ inputs []physics.Inputs }

func (s *session) Sync(in physics.Inputs) error       { s.inputs = append(s.inputs, in); return nil }
func (s *session) Release(uint8, uint8)               {}
func (s *session) Frame() *image.RGBA                 { return image.NewRGBA(image.Rect(0, 0, 1, 1)) }
func (s *session) Result() (tablelogic.Decimal, bool) { return tablelogic.Decimal{}, false }
func (s *session) PCM() []byte                        { return []byte{byte(len(s.inputs))} }
func (s *session) Cue(string)                         {}
func fixture(t *testing.T) (*Runner, *session, *time.Time) {
	t.Helper()
	m, err := frontend.New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := &session{}
	m.Mode = frontend.Playing
	m.Session = s
	now := time.Unix(1, 0)
	r := New(&frontend.Runtime{Model: m}, func() time.Time { return now })
	return r, s, &now
}
func advance(t *testing.T, r *Runner) {
	t.Helper()
	if err := r.Advance(nil, nil); err != nil {
		t.Fatal(err)
	}
}
func TestSuspendResumeClearsControlsAndDoesNotCatchUp(t *testing.T) {
	r, s, now := fixture(t)
	r.Submit(frontend.Input{Gameplay: gameplay.Controls{Left: true, Right: true, Down: true, Tilt: true, MouseFire: true, MouseY: 8}})
	r.Suspend()
	*now = now.Add(time.Hour)
	advance(t, r)
	if len(s.inputs) != 0 {
		t.Fatal("suspended source advanced")
	}
	r.Resume()
	advance(t, r)
	if len(s.inputs) != 1 || s.inputs[0] != (physics.Inputs{}) {
		t.Fatal("resume catchup or stuck controls", s.inputs)
	}
	if !r.Next.Equal(now.Add(time.Second / 71)) {
		t.Fatal("resume deadline", r.Next)
	}
}
func TestFocusLossKeepsDesktopPauseAndClearsEdges(t *testing.T) {
	r, s, now := fixture(t)
	r.Submit(frontend.Input{Gameplay: gameplay.Controls{Down: true, MouseFire: true}})
	r.Submit(frontend.Input{FocusLost: true})
	advance(t, r)
	if !r.Suspended || r.Runtime.Model.Mode != frontend.Paused || len(s.inputs) != 0 {
		t.Fatal("focus loss advanced gameplay")
	}
	*now = now.Add(time.Minute)
	advance(t, r)
	r.Resume()
	// The original desktop pause remains until the user makes a resume key.
	r.Submit(frontend.Input{Keys: []frontend.Key{frontend.Enter}})
	advance(t, r)
	*now = r.Next
	advance(t, r)
	if len(s.inputs) != 1 || s.inputs[0] != (physics.Inputs{}) {
		t.Fatal("focus loss retained controls", s.inputs)
	}
}
func TestEdgesConsumedOnceAndSlowHostCannotDropTicks(t *testing.T) {
	r, s, now := fixture(t)
	r.Submit(frontend.Input{Gameplay: gameplay.Controls{Left: true, MouseY: 24, MouseFire: true, Release: true}})
	*now = now.Add(9 * (time.Second / 71))
	var pcm []byte
	err := r.Advance(nil, func(b []byte) error { pcm = append(pcm, b...); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(s.inputs) != 10 || !bytes.Equal(pcm, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
		t.Fatal("dropped source tick/PCM", s.inputs, pcm)
	}
	for i, in := range s.inputs {
		if !in.Left {
			t.Fatal("held lost")
		}
		if in.MouseFire != (i == 1) || (i > 0 && (in.MouseY != 0 || in.Release)) {
			t.Fatal("edge repeated", i, in)
		}
	}
}
func TestActualFrontendAndTableCadences(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../TABLE1.PRG", "../../TABLE2.PRG", "../../TABLE3.PRG", "../../TABLE4.PRG")
	rt, err := frontend.Load("../..", nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1, 0)
	r := New(rt, func() time.Time { return now })
	before := now
	advance(t, r)
	if r.Next.Sub(before) != time.Second/60 {
		t.Fatal("INTRO cadence")
	}
	// Table mode is sufficient to verify the scheduling boundary; a fake session
	// makes the test independent of attract/game-start delay counters.
	rt.Model.Mode = frontend.Playing
	rt.Model.Session = &session{}
	now = r.Next
	before = now
	advance(t, r)
	if r.Next.Sub(before) != time.Second/71 {
		t.Fatal("table cadence")
	}
}

func TestSuspendCancelsScheduledMouseFire(t *testing.T) {
	r, s, now := fixture(t)
	r.Submit(frontend.Input{Gameplay: gameplay.Controls{MouseFire: true}})
	advance(t, r)
	if s.inputs[0].MouseFire {
		t.Fatal("SPRINGUP ran in SPRINGSTEEN task")
	}
	r.Suspend()
	*now = now.Add(time.Hour)
	r.Resume()
	advance(t, r)
	if len(s.inputs) != 2 || s.inputs[1].MouseFire {
		t.Fatal("pending fire survived suspend", s.inputs)
	}
}
func TestFocusLossAndGainInSamePollKeepsPauseWithoutStalling(t *testing.T) {
	r, s, now := fixture(t)
	r.Submit(frontend.Input{FocusLost: true})
	r.Resume()
	advance(t, r)
	if r.Suspended || r.Runtime.Model.Mode != frontend.Paused || len(s.inputs) != 0 {
		t.Fatal("batched focus lifecycle")
	}
	*now = r.Next
	r.Submit(frontend.Input{Keys: []frontend.Key{frontend.Enter}})
	advance(t, r)
	*now = r.Next
	advance(t, r)
	if len(s.inputs) != 1 {
		t.Fatal("focus gain stalled source")
	}
}

func TestFocusLossDoesNotMutateIntroTracker(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../TABLE1.PRG", "../../TABLE2.PRG", "../../TABLE3.PRG", "../../TABLE4.PRG")
	rt, err := frontend.Load("../..", nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1, 0)
	r := New(rt, func() time.Time { return now })
	advance(t, r)
	before := *rt.Player
	r.Submit(frontend.Input{FocusLost: true})
	now = r.Next
	advance(t, r)
	now = now.Add(time.Hour)
	advance(t, r)
	if !r.Suspended || !reflect.DeepEqual(before, *rt.Player) || len(rt.PCM) != 0 {
		t.Fatal("focus lifecycle changed tracker state")
	}
	r.Resume()
	advance(t, r)
	if rt.Player.Frames != before.Frames+audio.Rate/60 {
		t.Fatal("INTRO resume caught up", before.Frames, rt.Player.Frames)
	}
}

func TestDeadlinesThatBecomeDueDuringPCMSinkAreDrained(t *testing.T) {
	r, s, now := fixture(t)
	count := 0
	if err := r.Advance(nil, func(pcm []byte) error {
		count++
		if count == 1 {
			*now = now.Add(9 * (time.Second / 71))
		}
		if len(pcm) != 1 || int(pcm[0]) != count {
			t.Fatal("PCM out of source order", count, pcm)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if count != 10 || len(s.inputs) != 10 || !r.Next.After(*now) {
		t.Fatal("returned to host before draining due PCM", count, r.Next, *now)
	}
}
