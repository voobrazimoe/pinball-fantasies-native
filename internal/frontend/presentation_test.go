package frontend

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func TestSourceGriffinHeldFade(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 100; i++ {
		if e = r.Update(Input{}); e != nil {
			t.Fatal(e)
		}
	}
	raw, e := os.ReadFile("../../analysis/pf8-visual-fixtures.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Griffin struct {
			Hold string `json:"hold_rgba_sha256"`
		}
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(r.Frame().Pix)) != f.Griffin.Hold {
		t.Fatal("fade must retain 19/20 in VGA DAC units")
	}
}

func TestIndependentSidebarFontRaster(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	r.Model.Mode = Selector
	r.Model.Reveal = 60
	r.Model.SidebarTick = 700
	frame := r.Frame()
	region := image.NewRGBA(image.Rect(0, 0, 112, 90))
	draw.Draw(region, region.Rect, frame, image.Pt(8, 95), draw.Src)
	raw, e := os.ReadFile("../../analysis/pf8-bios-font.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Hash string `json:"independent_sidebar_rgba_sha256"`
	}
	if e = json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(region.Pix)) != fixture.Hash {
		t.Fatalf("independent sidebar raster hash: %x", sha256.Sum256(region.Pix))
	}
	// LOGGAN's final callback restores the source crop, including its 640px stride.
	r.Model.SidebarTick = 1199
	restored := r.Frame()
	logo := r.View.Art.Logo.Frame()
	for y := 95; y < 185; y++ {
		for x := 16; x < 128; x++ {
			if restored.RGBAAt(x, y) != logo.RGBAAt(x, y) {
				t.Fatal("LOGGAN restored the wrong source scanline")
			}
		}
	}
}
func TestFrontendDownHoldReleaseAndSpace(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, k := range []Key{Space, F1, Enter} {
		if e = r.Update(Input{Keys: []Key{k}}); e != nil {
			t.Fatal(e)
		}
	}
	for i := 0; i < 150; i++ {
		if e = r.Update(Input{}); e != nil {
			t.Fatal(e)
		}
	}
	g := r.Model.Session.(*partyland.Game)
	for i := 0; i < 4; i++ {
		if e = r.Update(Input{Down: true}); e != nil {
			t.Fatal(e)
		}
	}
	if g.Physics.SpringPosition != 4 {
		t.Fatal("frontend held Down lost")
	}
	if e = r.Update(Input{Tilt: true}); e != nil {
		t.Fatal(e)
	}
	if g.Physics.SpringPosition != 4 || g.Physics.Tilted {
		t.Fatal("Space launches or tilts in chute")
	}
	if e = r.Update(Input{Release: true}); e != nil {
		t.Fatal(e)
	}
	if g.Physics.SpringPosition != 0 || g.Physics.Ball.VY >= 0 {
		t.Fatal("Down key-up release")
	}
	if e = r.Update(Input{Keys: []Key{P}}); e != nil {
		t.Fatal(e)
	}
	dots, syncs, pos := g.Display.Dots, g.Physics.Syncs, g.Physics.SpringPosition
	for i := 0; i < 20; i++ {
		if e = r.Update(Input{Down: true, Release: true, Tilt: true}); e != nil {
			t.Fatal(e)
		}
	}
	if g.Display.Dots != dots || g.Physics.Syncs != syncs || g.Physics.SpringPosition != pos {
		t.Fatal("pause advanced matrix/plunger")
	}
}

func TestOriginalSidebarCRTCPan(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	r.Model.Mode = Selector
	r.Model.Reveal = 0
	r.Model.SidebarTick = 16
	final := r.Frame()
	for tick := 0; tick < 16; tick++ {
		r.Model.SidebarTick = tick
		got := r.Frame()
		shift := (16 - tick) * 8 * 4
		for i := 0; i < len(got.Pix)-shift; i++ {
			if got.Pix[i] != final.Pix[i+shift] {
				t.Fatal("creatretf CRTC start address", tick, i)
			}
		}
	}
}

func TestSelectorSharedRasterPaletteCapture(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	r.Model.Mode = Selector
	r.Model.Reveal = 60
	r.Model.SidebarTick = 20
	raw, e := os.ReadFile("../../analysis/pf8-bios-font.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Hash string `json:"selector_frame_rgba_sha256"`
	}
	if e = json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(r.Frame().Pix)) != fixture.Hash {
		t.Fatal("SHOWPICS raster banks must also color the sidebar")
	}
}

func TestSidebarTypingCallbackBounds(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	r.Model.Mode = Selector
	r.Model.Reveal = 60
	r.Model.SidebarTick = 510
	first := r.Frame()
	r.Model.SidebarTick = 629
	last := r.Frame()
	r.Model.SidebarTick = 700
	held := r.Frame()
	for y := 97; y < 105; y++ {
		for x := 24; x < 32; x++ {
			if first.RGBAAt(x, y) != first.RGBAAt(8, 95) {
				t.Fatal("first WRITEROMCHAR callback wrote two characters")
			}
		}
	}
	if !bytes.Equal(last.Pix, held.Pix) {
		t.Fatal("callback 629 must write the final character, not overrun the 120-byte text")
	}
}

func TestSourceTitleUnwrittenPixelsUsePaletteZero(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	// INTRO: BIOS mode12h clears index0; UNPKLBM draws at y150;
	// fade20 applies the image palette to the entire surface (19/20 hold).
	r.Model.Segment = 14
	r.Model.SegmentTick = 19
	frame := r.Frame()
	pal := r.View.Art.Startup[7].Palette
	for _, p := range []image.Point{{0, 0}, {639, 0}, {0, 479}, {639, 479}} {
		c := frame.RGBAAt(p.X, p.Y)
		want := [3]byte{}
		for i := range want {
			v := int(pal[i]>>2) * 19 / 20
			want[i] = byte((v << 2) | (v >> 4))
		}
		if [3]byte{c.R, c.G, c.B} != want || c.R == 0 {
			t.Fatal("BIOS index0 backdrop must fade with the title palette")
		}
	}
}

func TestSourceDigitalIllusionsCRTCStart(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	r.Model.Segment, r.Model.SegmentTick = 4, 9
	im := r.Frame()
	// INTRO vikingpos=240, vikingadr=80*247; lower IFF starts at365.
	// Last upper row is overwritten by the lower IFF at logical row118.
	for _, p := range []struct{ x, y, asset, sy int }{{17, 0, 3, 7}, {17, 117, 3, 124}, {17, 118, 2, 0}, {17, 239, 2, 121}} {
		c := r.View.Art.Startup[p.asset].Frame().RGBAAt(p.x, p.sy)
		want := c
		fade := func(v byte) byte { dac := (63 + 9*int(v>>2)) / 10; return byte(dac<<2 | dac>>4) }
		want.R, want.G, want.B = fade(c.R), fade(c.G), fade(c.B)
		if got := im.RGBAAt(p.x, p.y); got != want {
			t.Fatalf("CRTC crop at%d,%d: %v != %v", p.x, p.y, got, want)
		}
	}
}

func TestSourceSelectorTextList(t *testing.T) {
	m, _, _ := setup(t)
	m.Mode = Selector
	v := &View{}
	for cycle := 0; cycle < 12; cycle++ {
		previous := m.Page
		m.advanceSelector(false)
		if m.TextPage != cycle%10+1 {
			t.Fatal("TEXTPEK must wrap after ten pages")
		}
		lines := v.textPage(m)
		if len(lines) != 12 {
			t.Fatal("WRITEPAGE consumes twelve rows")
		}
		if highTextPage(m.TextPage) {
			want := "     PARTY LAND         "
			if previous == 1 {
				want = "     BILLION DOLLAR     "
			}
			if lines[1] != want || len(lines[2]) != 24 {
				t.Fatalf("WAITEND selected wrong HITEXT or fixed columns: %q", lines)
			}
		}
		m.advanceSelector(false)
	}
}

func TestSelectorMonoTextAgainstIsolatedOriginal(t *testing.T) {
	base := "../../analysis/pf8-runtime-validation/"
	testinputs.Require(t, "../../TABLE1.PRG", "../../TABLE2.PRG", base+"selector-text-scores.json")
	raw, e := os.ReadFile(base + "selector-text-scores.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct{ Tables map[string]string }
	if e = json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	for i := 1; i <= 2; i++ {
		b, e := hex.DecodeString(fixture.Tables[fmt.Sprint(i)])
		if e != nil {
			t.Fatal(e)
		}
		r.Model.Scores[i-1], e = DecodeScores(b)
		if e != nil {
			t.Fatal(e)
		}
	}
	r.Model.Mode, r.Model.Page, r.Model.PreviousPage = SelectorText, 1, 0
	r.Model.TextPage, r.Model.TextTick, r.Model.TextExitTick = 1, 45, 3
	r.Model.Counter, r.Model.SidebarTick = 0, 700
	frame := r.Frame()
	f, e := os.Open(base + "normal-isolated/text-0000.png")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	im, e := png.Decode(f)
	if e != nil {
		t.Fatal(e)
	}
	if im.Bounds() != image.Rect(0, 0, 640, 480) {
		t.Fatal("oracle dimensions")
	}
	// SHOWTEXT's fade3b20 step3: DAC levels16/20. Verify the full text/logo
	// region; independently evolving sidebar callback phase is excluded.
	for y := 0; y < 240; y++ {
		for x := 128; x < 640; x++ {
			want := frame.RGBAAt(x, y)
			for duplicate := 0; duplicate < 2; duplicate++ {
				r, g, b, a := im.At(x, y*2+duplicate).RGBA()
				if r != uint32(want.R)*257 || g != uint32(want.G)*257 || b != uint32(want.B)*257 || a != 65535 {
					t.Fatalf("text frame pixel%d,%d", x, y)
				}
			}
		}
	}
}
