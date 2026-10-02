package physics

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

type referenceFrame struct {
	Ball     Ball
	Raster   int16
	Flippers [3][3]int16
	Events   [][4]int
	RGBA     string `json:"rgba_sha256"`
}
type referenceCase struct {
	Name        string
	Start       []int
	Left, Right bool
	Release     int `json:"release_at"`
	Frames      []referenceFrame
}

func table(t *testing.T) *Table {
	t.Helper()
	testinputs.Require(t, "../../TABLE1.PRG")
	d, e := os.ReadFile("../../TABLE1.PRG")
	if e != nil {
		t.Fatal(e)
	}
	tb, e := DecodePartyLand(d)
	if e != nil {
		t.Fatal(e)
	}
	return tb
}
func checkFrame(t *testing.T, g *Game, w referenceFrame, n int) {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(g.Frame().Pix)); got != w.RGBA {
		t.Fatalf("sync %d framebuffer got %s want %s", n, got, w.RGBA)
	}
	if g.Ball != w.Ball || g.Raster != w.Raster {
		t.Fatalf("sync %d got %+v raster %d; want %+v raster %d", n, g.Ball, g.Raster, w.Ball, w.Raster)
	}
	for i, f := range g.Flippers {
		if [3]int16{f.Speed, f.Angle, f.Frame} != w.Flippers[i] {
			t.Fatalf("sync %d flipper %d", n, i)
		}
	}
	if len(g.Events) != len(w.Events) {
		t.Fatalf("sync %d events: got %v want %v", n, g.Events, w.Events)
	}
	for i, e := range g.Events {
		if [4]int{int(e.Kind), e.Object, int(e.X), int(e.Y)} != w.Events[i] {
			t.Fatalf("sync %d event: %+v", n, e)
		}
	}
}
func TestOriginalTrajectories(t *testing.T) {
	raw, e := os.ReadFile(testinputs.Generated(t, "reference_pf3.py", "TABLE1.PRG"))
	if e != nil {
		t.Fatal(e)
	}
	var cases []referenceCase
	if e = json.Unmarshal(raw, &cases); e != nil {
		t.Fatal(e)
	}
	tb := table(t)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			g := New(tb)
			g.Ball = Ball{X: int32(c.Start[0]) * 1024, Y: int32(c.Start[1]) * 1024, VX: int16(c.Start[2]), VY: int16(c.Start[3]), GY: 8, PixelX: int16(c.Start[0]), PixelY: int16(c.Start[1]), High: c.Start[4] != 0}
			checkFrame(t, g, c.Frames[0], 0)
			for n := 1; n < len(c.Frames); n++ {
				if n-1 == c.Release {
					g.Release(32, 0)
				}
				if e := g.Sync(Inputs{Left: c.Left, Right: c.Right}); e != nil {
					t.Fatal(e)
				}
				checkFrame(t, g, c.Frames[n], n)
			}
		})
	}
}
func TestDataAndRegression(t *testing.T) {
	tb := table(t)
	if tb.Sin[0] != 0 || tb.Sin[512] != 16384 || tb.Sin[1024] != 0 || tb.Sin[1536] != -16384 {
		t.Fatal("original trig lookup")
	}
	g := New(tb)
	if g.Ball.X != 297*1024 || g.Ball.Y != 530*1024 || g.Ball.VX != 10 || g.Ball.VY != 0 || g.Ball.GY != 8 {
		t.Fatal("PF2 SETBALL")
	}
	hash := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	if hash(tb.Initial.Framebuffer().Pix) != "b921c9933cf6f49b605d78e28257d77c6c5f420ca84afc03a8dac2edda81f7b8" || hash(tb.Initial.Playfield.Framebuffer().Pix) != "2d90631a3348512de700102621e1da16c42eb844dd1eb3ea3382c7fa12775753" {
		t.Fatal("PF1/PF2")
	}
	a, b := New(tb), New(tb)
	for n := 0; n < 600; n++ {
		if n == 100 {
			a.Release(32, 0)
			b.Release(32, 0)
		}
		input := Inputs{Left: n%80 < 40, Right: n%101 < 30}
		if err := a.Sync(input); err != nil {
			t.Fatal(err)
		}
		if err := b.Sync(input); err != nil {
			t.Fatal(err)
		}
		sa, _ := json.Marshal(a.Ball)
		sb, _ := json.Marshal(b.Ball)
		if !bytes.Equal(sa, sb) || a.Raster != b.Raster || !bytes.Equal(a.mask12, b.mask12) {
			t.Fatalf("repeat %d", n)
		}
		if !bytes.Equal(a.Frame().Pix, b.Frame().Pix) {
			t.Fatalf("repeat framebuffer %d", n)
		}
	}
	if _, e := DecodePartyLand(nil); e == nil {
		t.Fatal("unvalidated data")
	}
}
func TestIntegerBoundaries(t *testing.T) {
	if q, e := divide(-1023, 1024); e != nil || q != 0 {
		t.Fatal("signed division")
	}
	if _, e := divide(33554432, 1024); e == nil {
		t.Fatal("quotient overflow")
	}
	if _, e := divide(1, 0); e == nil {
		t.Fatal("zero divisor")
	}
	b := int16(32767)
	b++
	if b != -32768 {
		t.Fatal("word wrap")
	}
}

func TestOriginalFlipperUpBranch(t *testing.T) {
	g := New(table(t))
	g.Ball.Hold = true
	if err := g.step(Inputs{Left: true}); err != nil {
		t.Fatal(err)
	}
	f := g.Flippers[0]
	if f.Speed != -68 || f.Angle != 68 || f.Frame != 1 {
		t.Fatalf("TILT0/UpFlip JLE: %+v", f)
	}
	if err := g.step(Inputs{Left: true}); err != nil {
		t.Fatal(err)
	}
	f = g.Flippers[0]
	if f.Speed != -75 || f.Angle != 143 || f.Frame != 2 {
		t.Fatalf("second UpFlip: %+v", f)
	}
}

func TestDirectSourceResponse(t *testing.T) {
	g := New(table(t))
	g.Ball.VX = 1024
	g.Ball.VY = 0
	if err := g.respond(contact{angle: 0, count: 1, material: 6}); err != nil {
		t.Fatal(err)
	}
	// SIN[512]=16384: normal=2048, bounce=450 gives -883, then
	// the signed high-word reconstruction gives -442. No penetration push.
	if g.Ball.VX != -442 || g.Ball.VY != 0 || g.Ball.X != 297*1024 {
		t.Fatalf("steel wall: %+v", g.Ball)
	}
	g = New(table(t))
	g.Ball.VX = 1024
	g.Ball.VY = 0
	g.pendingBumper = true
	if err := g.respond(contact{angle: 0, count: 1, material: 7}); err != nil {
		t.Fatal(err)
	}
	// -2048-7000=-9048; -9048 - trunc(-9048*256/400)=-3258.
	if g.Ball.VX != -1629 || !g.pendingBumper {
		t.Fatalf("active bumper: %+v", g.Ball)
	}
}

func TestOriginalRelease(t *testing.T) {
	g := New(table(t))
	g.Release(32, 255)
	if g.Ball.VX != 0 || g.Ball.VY != -5567 || g.Ball.Rotation != 15 {
		t.Fatal("SPRINGUP full charge/jitter")
	}
	g.SpringValid = false
	before := g.Ball
	g.Release(16, 0)
	if g.Ball != before {
		t.Fatal("invalid spring release")
	}
}
