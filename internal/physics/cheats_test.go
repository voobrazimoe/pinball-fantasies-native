package physics

import "testing"

func TestTiltDisabledPreservesPhysicalPush(t *testing.T) {
	g := &Game{TiltDisabled: true, AllowFlip: true}
	for i := 0; i < 4; i++ {
		g.TiltInput(true, false)
		g.TiltInput(false, false)
	}
	if g.TiltCounter != 0 || g.Tilted || g.tiltLatched {
		t.Fatal("TILTLOGIC must exit before latch and counter writes")
	}
	g.push(true)
	if g.ScreenPosition != 600 {
		t.Fatal("EARTHQUAKE must retain BALLCODE/TILT0 push")
	}
	g.TiltDisabled = false
	for i := 0; i < 3; i++ {
		g.TiltInput(true, false)
		g.TiltInput(false, false)
	}
	if !g.Tilted {
		t.Fatal("FAIRPLAY must restore tilt")
	}
}

func TestFairPlaySourcePhysicsCadence(t *testing.T) {
	// FANTASIE LATE_RASTER_INTERRUPT executes KOLLAKULAN twice, skips
	// DO_BALL_CALCULATIONS iff SHIFTKEYS bit2 is set. VBLANK still has two
	// calculations. An unobstructed ball lets displacement count these passes.
	for _, fast := range []bool{false, true} {
		g := New(table(t))
		g.mask12 = make([]byte, len(g.mask12))
		g.mask22 = make([]byte, len(g.mask22))
		g.FastBall = fast
		g.Ball = Ball{X: 160 * 1024, Y: 200 * 1024, VX: 10, PixelX: 160, PixelY: 200, High: true}
		calls := 0
		g.BeforeLate = func() {
			calls++
			if g.Ball.X != 160*1024+20 {
				t.Fatal("VBLANK cadence")
			}
		}
		if err := g.Sync(Inputs{}); err != nil {
			t.Fatal(err)
		}
		steps := int32(3)
		if fast {
			steps = 4
		}
		if calls != 1 || g.Ball.X != 160*1024+steps*10 {
			t.Fatal("SHIFTKEYS late-raster cadence", fast, g.Ball.X, calls)
		}
	}
}
