package physics

import (
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func showTable(t *testing.T) *Table {
	t.Helper()
	testinputs.Require(t, "../../TABLE3.PRG")
	b, e := os.ReadFile("../../TABLE3.PRG")
	if e != nil {
		t.Fatal(e)
	}
	v, e := DecodeGameshow(b)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func TestGameshowSourceGeometryAndPhysics(t *testing.T) {
	v := showTable(t)
	g := New(v)
	if g.Raster != (259+33)*16 || g.Ball.PixelX != 299 || g.Ball.PixelY != 530 {
		t.Fatal("SETBALL", g.Ball)
	}
	if e := g.Sync(Inputs{}); e != nil {
		t.Fatal(e)
	}
	if g.Ball.X != 299*1024+30 || g.Ball.Y != 530*1024+24 || g.Ball.VY != 23 || g.Ball.GY != 7 {
		t.Fatal("BALLCODE first sync", g.Ball)
	}
	g.Release(32, 0)
	if g.Ball.VY != -5312 {
		t.Fatal("SPRINGUP", g.Ball)
	}
	if len(v.gravity) != 5 || v.gravity[1] != [2]int16{4, 9} || v.gravity[4] != [2]int16{6, 10} || v.Flippers[2].Kind != 1 || v.Flippers[2].Top != 176 || v.Flippers[2].Height != 51 || v.Flippers[2].Frames != 12 || v.flipStride != 55 {
		t.Fatal("SHOW configuration")
	}
	if v.Sin != table(t).Sin || v.Materials != table(t).Materials {
		t.Fatal("shared source lookup values")
	}
	g = New(v)
	g.moveFlippers(Inputs{Left: true})
	if g.Flippers[0].Frame != 1 || g.Flippers[0].Speed != -68 || g.Flippers[2].Frame != 0 {
		t.Fatal("original held flipper sides", g.Flippers)
	}
	g.Frame()
	g.SetBall(130, 80, 0, 0, false)
	g.checkLevels()
	if !g.Ball.High {
		t.Fatal("upper entry")
	}
	g.SetBall(105, 80, 0, 0, true)
	g.checkLevels()
	if g.Ball.High {
		t.Fatal("lower entry")
	}
	for i, r := range v.targets {
		g.Ball.High = false
		g.Ball.HitX = (r.X1 + r.X2) / 2
		g.Ball.HitY = (r.Y1 + r.Y2) / 2
		g.Events = nil
		g.checkSpringAndTargets()
		if len(g.Events) != 1 || g.Events[0].Object != i {
			t.Fatal("target identity", i, g.Events)
		}
	}
	for i, r := range v.bumper {
		found := false
		for y := r.Y1 - 16; y <= r.Y2; y++ {
			for x := r.X1 - 16; x <= r.X2; x++ {
				g.SetBall(x, y, 0, 0, false)
				if c, ok := g.collision(); ok && c.kind == EventBumperHit && c.object == i {
					found = true
				}
			}
		}
		if !found {
			t.Fatal("original bumper mask", i)
		}
	}
}
