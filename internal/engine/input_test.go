package engine

import (
	"image"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/tablelogic"
	"testing"
	"time"
)

type inputSession struct {
	valid    bool
	position uint8
	launches []uint8
	inputs   []physics.Inputs
}

func (s *inputSession) Sync(in physics.Inputs) error {
	s.inputs = append(s.inputs, in)
	gameplay.Spring(&s.position, s.valid, in, func(charge uint8) { s.launches = append(s.launches, charge); s.valid = false })
	return nil
}
func (s *inputSession) Release(uint8, uint8)               {}
func (s *inputSession) Frame() *image.RGBA                 { return image.NewRGBA(image.Rect(0, 0, 1, 1)) }
func (s *inputSession) Result() (tablelogic.Decimal, bool) { return tablelogic.Decimal{}, false }
func (s *inputSession) PCM() []byte                        { return []byte{0, 0, 0, 0} }
func (s *inputSession) Cue(string)                         {}
func (s *inputSession) InChute() bool                      { return s.valid }
func inputEngine(t *testing.T) (*Engine, *inputSession) {
	t.Helper()
	m, err := frontend.New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := &inputSession{valid: true}
	m.Mode = frontend.Playing
	m.Session = s
	return New(&frontend.Runtime{Model: m}, 0), s
}
func tickEngine(t *testing.T, e *Engine) {
	t.Helper()
	if err := e.Advance(e.runner.Next.UnixNano(), nil); err != nil {
		t.Fatal(err)
	}
}
func TestNativeInputEdgesRelativeMotionAndInvalidSpring(t *testing.T) {
	e, s := inputEngine(t)
	e.SetAction(Left, true)
	e.SetAction(Right, true)
	e.PlungerDelta(3)
	e.PlungerDelta(5)
	tickEngine(t, e)
	if s.position != 1 || !s.inputs[0].Left || !s.inputs[0].Right {
		t.Fatal("accumulated movement/held controls", s)
	}
	e.PlungerDelta(8000)
	tickEngine(t, e)
	if s.position != 2 {
		t.Fatal("motion accelerated source spring", s.position)
	}
	e.PlungerFire()
	tickEngine(t, e)
	if len(s.launches) != 0 {
		t.Fatal("fire ran before following source task")
	}
	tickEngine(t, e)
	tickEngine(t, e)
	if len(s.launches) != 1 || s.launches[0] != 2 {
		t.Fatal("fire edge repeated or altered charge", s.launches)
	}
	e.PlungerDelta(7)
	e.PlungerFire()
	tickEngine(t, e)
	s.valid = true
	e.PlungerDelta(1)
	tickEngine(t, e)
	if s.position != 0 || len(s.launches) != 1 {
		t.Fatal("invalid spring banked movement/fire", s)
	}
	e.PlungerDelta(7)
	tickEngine(t, e)
	if s.position != 1 {
		t.Fatal("valid remainder lost")
	}
	e.SetAction(Spring, true)
	e.SetAction(Spring, false)
	tickEngine(t, e)
	tickEngine(t, e)
	if len(s.launches) != 2 {
		t.Fatal("short keyboard release not consumed once", s.launches)
	}
}
func TestNativeLifecycleClearsScheduledFireAndReanchors(t *testing.T) {
	e, s := inputEngine(t)
	e.SetAction(Spring, true)
	tickEngine(t, e)
	e.PlungerFire()
	tickEngine(t, e)
	if err := e.Suspend(); err != nil {
		t.Fatal(err)
	}
	before := len(s.inputs)
	e.SetAction(Left, true)
	e.Key(uint8(frontend.P))
	e.PlungerDelta(100)
	e.PlungerFire()
	if err := e.Advance(int64(time.Hour), nil); err != nil {
		t.Fatal(err)
	}
	if len(s.inputs) != before {
		t.Fatal("background catchup")
	}
	if err := e.Resume(int64(time.Hour)); err != nil {
		t.Fatal(err)
	}
	tickEngine(t, e)
	if e.runner.Runtime.Model.Mode != frontend.Paused || len(s.inputs) != before {
		t.Fatal("focus gain implicitly unpaused")
	}
	e.Key(uint8(frontend.Enter))
	tickEngine(t, e)
	tickEngine(t, e)
	if len(s.launches) != 0 || s.inputs[len(s.inputs)-1] != (physics.Inputs{}) {
		t.Fatal("stale lifecycle input", s)
	}
	e.SetAction(Left, true)
	if err := e.Resume(e.last); err != nil {
		t.Fatal(err)
	}
	tickEngine(t, e)
	if !s.inputs[len(s.inputs)-1].Left {
		t.Fatal("duplicate active resume cleared valid hold")
	}
}
