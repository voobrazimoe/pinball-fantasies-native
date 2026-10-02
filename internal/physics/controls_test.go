package physics

import "testing"

func TestTiltMakeLatchThresholdAndReset(t *testing.T) {
	g := &Game{AllowFlip: true}
	if w, tilt := g.TiltInput(true, false); w || tilt || g.TiltCounter != 60 {
		t.Fatal("first push")
	}
	g.TiltInput(true, false)
	if g.TiltCounter != 60 {
		t.Fatal("held key repeated")
	}
	g.TiltInput(false, false)
	if w, tilt := g.TiltInput(true, false); !w || tilt || g.TiltCounter != 120 {
		t.Fatal("second push warning")
	}
	g.TiltInput(false, false)
	if _, tilt := g.TiltInput(true, false); !tilt || !g.Tilted || g.AllowFlip {
		t.Fatal("third push tilt")
	}
	g.ResetTilt()
	if g.Tilted || g.TiltCounter != 0 || !g.AllowFlip {
		t.Fatal("new-ball reset")
	}
	g.TiltInput(false, false)
	g.TiltInput(true, true)
	if g.TiltCounter != 0 {
		t.Fatal("chute should inhibit tilt")
	}
}
func TestTablePushIntegerMotion(t *testing.T) {
	g := &Game{}
	for i := 0; i < 4; i++ {
		g.push(true)
	}
	if g.ScreenPosition != 2048 || g.ScreenOffset != 4 || g.ScreenSpeed != 0 {
		t.Fatal("push clamp")
	}
	for i := 0; i < 11; i++ {
		g.push(false)
	}
	if g.ScreenPosition != 0 || g.ScreenOffset != 0 || g.ScreenSpeed != 0 {
		t.Fatal("return clamp")
	}
}
