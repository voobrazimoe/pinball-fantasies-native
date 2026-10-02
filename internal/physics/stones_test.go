package physics

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func stonesTable(t *testing.T) *Table {
	t.Helper()
	testinputs.Require(t, "../../TABLE4.PRG")
	b, e := os.ReadFile("../../TABLE4.PRG")
	if e != nil {
		t.Fatal(e)
	}
	v, e := DecodeStones(b)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func TestStonesSourcePhysicsAndTwoFlippers(t *testing.T) {
	v := stonesTable(t)
	g := New(v)
	b, e := os.ReadFile("../../analysis/pf10-assets.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		FirstFlipper string `json:"first_left_flipper_indices_sha256"`
		Geometry     struct {
			Masks     map[string]struct{ SHA256 string }
			Gravities [][2]int16
		}
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	masks := map[string][]byte{"mask12": v.mask12, "mask11": v.mask11, "mask22": v.mask22, "mask13": v.mask13, "mask21": v.mask21, "mask23": v.mask23, "hid2": v.upperForeground}
	for name, raw := range masks {
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != f.Geometry.Masks[name].SHA256 {
			t.Fatal("independent mask", name)
		}
	}
	for i, xy := range f.Geometry.Gravities {
		if v.gravity[i] != xy {
			t.Fatal("source ramp gravity", i)
		}
	}
	if g.Ball.PixelX != 297 || g.Ball.PixelY != 530 || g.Raster != 4672 {
		t.Fatal("SETBALL")
	}
	if e = g.Sync(Inputs{}); e != nil {
		t.Fatal(e)
	}
	if g.Ball.X != 297*1024+30 || g.Ball.Y != 530*1024+24 || g.Ball.VY != 26 || g.Ball.GY != 10 {
		t.Fatal("source first sync", g.Ball)
	}
	if v.Flippers[2].Kind != 0 || v.Flippers[0].Frames != 20 || v.Flippers[1].Frames != 20 || v.flipStride != 30 || len(v.gravity) != 11 {
		t.Fatal("source two flippers and ramp layout")
	}
	if v.Sin != table(t).Sin || v.Materials != table(t).Materials {
		t.Fatal("BALLCODE lookup reuse")
	}
	g = New(v)
	g.moveFlippers(Inputs{Left: true})
	g.animateFlipper(0)
	if g.Flippers[0].Frame != 1 || fmt.Sprintf("%x", sha256.Sum256(g.indices)) != f.FirstFlipper {
		t.Fatal("independent FLIPPRA frame")
	}
	g.Release(32, 0)
	if g.Ball.VY != -5312 {
		t.Fatal("SPRINGUP")
	}
	g.SetBall(42, 20, 0, 0, false)
	g.checkLevels()
	if !g.Ball.High {
		t.Fatal("LEVEL2LISTA")
	}
	g.SetBall(20, 20, 0, 0, true)
	g.checkLevels()
	if g.Ball.High {
		t.Fatal("LEVEL1LISTA")
	}
	for i, r := range v.targets {
		g.Ball.High = false
		g.Ball.HitX = (r.X1 + r.X2) / 2
		g.Ball.HitY = (r.Y1 + r.Y2) / 2
		g.Events = nil
		g.checkSpringAndTargets()
		if len(g.Events) != 1 || g.Events[0].Object != i {
			t.Fatal("ZonLista identity", i, g.Events)
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
			t.Fatal("bumper mask identity", i)
		}
	}
	for _, r := range v.levels[0] {
		g.SetBall(r.X1-8, r.Y1-8, 0, 0, false)
		g.checkLevels()
		if !g.Ball.High {
			t.Fatal("layer entry", r)
		}
	}
}
