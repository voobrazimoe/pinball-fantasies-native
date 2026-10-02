package physics

import (
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func speedTable(t *testing.T) *Table {
	t.Helper()
	testinputs.Require(t, "../../TABLE2.PRG")
	b, e := os.ReadFile("../../TABLE2.PRG")
	if e != nil {
		t.Fatal(e)
	}
	table, e := DecodeSpeedDevils(b)
	if e != nil {
		t.Fatal(e)
	}
	return table
}
func TestSpeedDevilsSetBallAndIntegrator(t *testing.T) {
	g := New(speedTable(t))
	if g.Ball.PixelX != 300 || g.Ball.PixelY != 530 || g.Ball.VX != 10 || g.Ball.GY != 8 {
		t.Fatal(g.Ball)
	}
	if e := g.Sync(Inputs{}); e != nil {
		t.Fatal(e)
	}
	// Three original moves: Y += 0,8,16; X += 10 three times.
	if g.Ball.X != 300*1024+30 || g.Ball.Y != 530*1024+24 || g.Ball.VY != 23 || g.Ball.GY != 7 {
		t.Fatal(g.Ball)
	}
	g.Release(32, 0)
	if g.Ball.VY != -5312 {
		t.Fatal("SPRINGUP")
	}
	for i := 0; i < 100; i++ {
		if e := g.Sync(Inputs{Left: i%17 < 4, Right: i%23 < 4}); e != nil {
			t.Fatal(e)
		}
		g.Frame()
	}
}
func TestSpeedDevilsContentGeometry(t *testing.T) {
	table := speedTable(t)
	if len(table.bumper) != 4 || len(table.targets) != 6 || table.Flippers[2].Top != 168 || table.Flippers[2].Height != 53 || table.flipStride != 42 {
		t.Fatal("SDEV geometry")
	}
	if table.Sin != speedTable(t).Sin {
		t.Fatal("immutable decode")
	}
	g := New(table)
	g.moveFlippers(Inputs{Left: true})
	if g.Flippers[0].Frame != 1 || g.Flippers[0].Speed != -68 {
		t.Fatal(g.Flippers[0])
	}
	g.Frame()
	g.SetBall(75, 120, 0, 0, false)
	g.checkLevels()
	if !g.Ball.High {
		t.Fatal("level2 upper entry")
	}
	g.SetBall(75, 150, 0, 0, true)
	g.checkLevels()
	if g.Ball.High {
		t.Fatal("level1 lower exit")
	}
	g.SetBall(20, 576, 0, 0, false)
	if e := g.Sync(Inputs{}); e != nil {
		t.Fatal(e)
	}
	if !g.Stopped || len(g.Events) == 0 || g.Events[len(g.Events)-1].Kind != EventDrain {
		t.Fatal("drain")
	}
	// Contacts at SDEV B/U/R and N/I/N zones produce the table-local identity.
	for i, r := range table.targets {
		g.Ball.High = false
		g.Ball.HitX = (r.X1 + r.X2) / 2
		g.Ball.HitY = (r.Y1 + r.Y2) / 2
		g.Events = nil
		g.checkSpringAndTargets()
		if len(g.Events) != 1 || g.Events[0].Object != i {
			t.Fatalf("target %d: %v", i, g.Events)
		}
	}
}

func TestSpeedDevilsCollisionResponse(t *testing.T) {
	g := New(speedTable(t))
	// Both linked tables contain exactly the same MAT_TABLE and sine data.
	p := table(t)
	if g.Table.Materials != p.Materials || g.Table.Sin != p.Sin {
		t.Fatal("shared BALLCODE lookup content differs")
	}
	g.Ball.VX, g.Ball.VY = 1024, 0
	if err := g.respond(contact{angle: 0, count: 1, material: 6}); err != nil {
		t.Fatal(err)
	}
	// Original steel bounce: normal 2048, coefficient 450, signed-high-word
	// reconstruction -442. The Speed Devils SETBALL origin remains unchanged.
	if g.Ball.VX != -442 || g.Ball.VY != 0 || g.Ball.X != 300*1024 {
		t.Fatal("integer steel response", g.Ball)
	}
	// Exercise original Speed Devils collision masks, rather than a synthetic
	// collision map: each source bumper region must resolve its own identity.
	for index, r := range g.Table.bumper {
		found := false
		for y := r.Y1 - 16; y <= r.Y2 && !found; y++ {
			for x := r.X1 - 16; x <= r.X2; x++ {
				g.SetBall(x, y, 0, 0, false)
				c, hit := g.collision()
				if hit && c.kind == EventBumperHit && c.object == index {
					found = true
					break
				}
			}
		}
		if !found {
			t.Fatal("original bumper collision missing", index)
		}
	}
}
