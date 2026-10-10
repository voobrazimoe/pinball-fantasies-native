package gamepad

import (
	"pinballfantasies/internal/frontend"
	"testing"
)

func TestPlungerTapAndDuplicate(t *testing.T) {
	p := New()
	p.Mode(frontend.Playing)
	p.Button(X, true)
	if !p.Controls().Down {
		t.Fatal("charge missing")
	}
	if p.Button(X, true).Gameplay.Release {
		t.Fatal("repeat launched")
	}
	if !p.Button(X, false).Gameplay.Release {
		t.Fatal("short tap lost release edge")
	}
	if p.Controls().Down || p.Button(X, false).Gameplay.Release {
		t.Fatal("duplicate release")
	}
}

func TestTriggerHysteresisAndAliases(t *testing.T) {
	p := New()
	p.Mode(frontend.Playing)
	p.Axis(4, 16000)
	p.Button(LeftShoulder, true)
	p.Axis(4, 14000)
	if !p.Controls().Left {
		t.Fatal("threshold chatter")
	}
	p.Axis(4, 12000)
	if !p.Controls().Left {
		t.Fatal("trigger released shoulder")
	}
	p.Axis(5, 32767)
	p.Button(LeftShoulder, false)
	if p.Controls().Left || !p.Controls().Right {
		t.Fatal("sides or aliases mixed")
	}
	p.Axis(5, 0)
	if p.Controls().Right {
		t.Fatal("right stuck")
	}
}

func TestFocusAndModeRequireFreshPress(t *testing.T) {
	for _, boundary := range []string{"focus", "mode", "connection"} {
		t.Run(boundary, func(t *testing.T) {
			p := New()
			p.Mode(frontend.Playing)
			p.Button(X, true)
			p.Button(LeftShoulder, true)
			p.Axis(5, 32767)
			switch boundary {
			case "focus":
				p.Focus(false)
				p.Button(Y, true)
				p.Focus(true)
			case "mode":
				p.Mode(frontend.Paused)
				p.Mode(frontend.Playing)
			case "connection":
				p.Connect(1<<X | 1<<LeftShoulder | 1<<rightTrigger)
			}
			p.Button(X, true)
			p.Button(LeftShoulder, true)
			p.Axis(5, 30000)
			// Releasing another alias must not resurrect a suppressed hold.
			p.Axis(4, 32767)
			p.Axis(4, 0)
			c := p.Controls()
			if c.Left || c.Right || c.Down || c.Tilt {
				t.Fatalf("held control leaked: %+v", c)
			}
			if p.Button(X, false).Gameplay.Release {
				t.Fatal("boundary fired plunger")
			}
			p.Button(X, true)
			if !p.Controls().Down {
				t.Fatal("fresh press suppressed")
			}
		})
	}
}

func TestDisconnectNeverLaunches(t *testing.T) {
	p := New()
	p.Mode(frontend.Playing)
	p.Button(X, true)
	p.Axis(4, 32767)
	p.Disconnect()
	if p.Controls().Down || p.Controls().Left || p.Button(X, false).Gameplay.Release {
		t.Fatal("disconnect leaked input")
	}
}

func TestMenuMappings(t *testing.T) {
	checks := []struct {
		mode   frontend.Mode
		button int
		key    frontend.Key
	}{
		{frontend.Startup, A, frontend.Space},
		{frontend.Selector, A, frontend.F1}, {frontend.Selector, B, frontend.F2},
		{frontend.Selector, X, frontend.F3}, {frontend.Selector, Y, frontend.F4},
		{frontend.Selector, Start, frontend.F5},
		{frontend.Options, Up, frontend.Up}, {frontend.Options, Down, frontend.Down},
		{frontend.Options, A, frontend.Enter}, {frontend.Options, B, frontend.Escape},
		{frontend.TableAttract, A, frontend.F1}, {frontend.Playing, A, frontend.Enter},
		{frontend.Playing, Start, frontend.P}, {frontend.Paused, Start, frontend.Space},
		{frontend.QuitQuestion, A, 21}, {frontend.QuitQuestion, B, 49},
		{frontend.Closing, Back, frontend.Escape},
	}
	for _, c := range checks {
		p := New()
		p.Mode(c.mode)
		in := p.Button(c.button, true)
		if len(in.Keys) != 1 || in.Keys[0] != c.key {
			t.Fatalf("mode=%v button=%d: %+v", c.mode, c.button, in)
		}
		if len(p.Button(c.button, true).Keys) != 0 {
			t.Fatal("duplicate menu make")
		}
	}
}

func TestSelectorHoldDoesNotChargeNextMode(t *testing.T) {
	p := New()
	p.Mode(frontend.Selector)
	p.Button(X, true)
	p.Mode(frontend.Playing)
	if p.Controls().Down || p.Button(X, false).Gameplay.Release {
		t.Fatal("table selection became a launch")
	}
}
