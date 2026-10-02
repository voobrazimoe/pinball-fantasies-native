package speeddevils

import (
	"pinballfantasies/internal/physics"
	"testing"
)

func TestSourceSpringTiltAndNextBall(t *testing.T) {
	g := testGame(t)
	for i := 0; i < 32; i++ {
		g.gameplayControl(physics.Inputs{Down: true})
	}
	if g.Physics.SpringPosition != 32 || g.Frame().Rect.Dy() != 350 {
		t.Fatal("charge/composition")
	}
	g.gameplayControl(physics.Inputs{Release: true})
	if g.Physics.SpringPosition != 0 || g.Physics.Ball.VY != -5312 {
		t.Fatal("release")
	}
	g.inChute = false
	g.Physics.Ball.Hold = true
	for _, held := range []bool{true, false, true, false, true} {
		g.gameplayControl(physics.Inputs{Tilt: held})
	}
	if !g.Physics.Tilted || g.Physics.AllowFlip || g.music.ReturnPosition != 62 {
		t.Fatal("SDEV source tilt/lastjingle")
	}
	for i := 0; i < 8; i++ {
		if e := g.Sync(physics.Inputs{Left: true}); e != nil {
			t.Fatal(e)
		}
	}
	if g.matrix.op != "_WAIT" {
		t.Fatal("tilt matrix wait")
	}
	g.newBall()
	if g.Physics.Tilted || g.Physics.TiltCounter != 0 || !g.Physics.AllowFlip {
		t.Fatal("tilt reset")
	}
}
