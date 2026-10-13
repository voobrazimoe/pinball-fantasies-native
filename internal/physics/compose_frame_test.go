package physics_test

import (
	"bytes"
	"image"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/demodata"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
)

// demoTable decodes the bundled demo's Party Land, so this runs everywhere.
func demoTable(t *testing.T) (*physics.Table, []byte) {
	t.Helper()
	dir, cleanup, err := demodata.Extract()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	intro, err := os.ReadFile(filepath.Join(dir, "INTRO.PRG"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "TABLE1.PRG"))
	if err != nil {
		t.Fatal(err)
	}
	id, err := datalayout.DetectDemoInstallation(intro, raw)
	if err != nil {
		t.Fatal(err)
	}
	data, err := datalayout.PreparePRGForProfile(id, "TABLE1.PRG", raw)
	if err != nil {
		t.Fatal(err)
	}
	table, err := physics.DecodePartyLand(data)
	if err != nil {
		t.Fatal(err)
	}
	return table, data
}

// ComposeFrame must match the snapshot pipeline pixel for pixel, including
// transparent rows past the table, SCREENPOSY black, spring and ball masks.
func TestComposeFrameMatchesSnapshotPipeline(t *testing.T) {
	table, data := demoTable(t)
	display := presentation.New(1, data)
	rng := rand.New(rand.NewSource(1))
	var reused *image.RGBA
	for i := 0; i < 400; i++ {
		g := physics.New(table)
		g.SpringGraphics = display.SpringGraphics()
		c := settings.Legacy()
		c.Resolution = byte(rng.Intn(2))
		c.ScrollMode = settings.ScrollMode(rng.Intn(4))
		g.Settings = c
		g.PresentationFullTable = rng.Intn(4) == 0
		g.Raster = int16(rng.Intn(700*16)) - 64*16
		g.ScreenOffset = int16(rng.Intn(17) - 8)
		g.Ball.PixelX = int16(rng.Intn(360) - 20)
		g.Ball.PixelY = int16(rng.Intn(620) - 20)
		g.Ball.High = rng.Intn(3) == 0
		g.Ball.Hold = rng.Intn(4) == 0
		g.SpringPosition = uint8(rng.Intn(33))
		if rng.Intn(5) == 0 {
			g.SpringGraphics = nil
		}
		for f := range g.Flippers {
			if n := g.Flippers[f].Frames; n > 0 {
				g.Flippers[f].Frame = int16(rng.Intn(int(n)))
			}
		}
		var p [768]byte
		rng.Read(p[:])
		off, on := byte(96), byte(242)
		want := presentation.ComposeNative(g.FramePalette(p), display, p, off, on, g.PresentationSettings(), g.ScreenOffset)
		got := g.ComposeFrame(nil, p, display, off, on)
		reused = g.ComposeFrame(reused, p, display, off, on)
		if got.Rect != want.Rect || !bytes.Equal(got.Pix, want.Pix) {
			t.Fatalf("state %d (%+v full=%v raster=%d offset=%d ball=%+v): fresh frame differs", i, c, g.PresentationFullTable, g.Raster, g.ScreenOffset, g.Ball)
		}
		if reused.Rect != want.Rect || !bytes.Equal(reused.Pix, want.Pix) {
			t.Fatalf("state %d: reused frame differs", i)
		}
	}
}

func TestComposeFrameReuseDoesNotAllocate(t *testing.T) {
	table, data := demoTable(t)
	display := presentation.New(1, data)
	g := physics.New(table)
	g.SpringGraphics = display.SpringGraphics()
	var p [768]byte
	dst := g.ComposeFrame(nil, p, display, 96, 242)
	if n := testing.AllocsPerRun(20, func() { dst = g.ComposeFrame(dst, p, display, 96, 242) }); n > 1 {
		t.Fatalf("%v allocations per reused frame", n)
	}
}
