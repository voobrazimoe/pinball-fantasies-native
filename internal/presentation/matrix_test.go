package presentation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"testing"
)

func original(t *testing.T, n int) *Display {
	t.Helper()
	testinputs.Require(t, fmt.Sprintf("../../TABLE%d.PRG", n))
	b, e := os.ReadFile(fmt.Sprintf("../../TABLE%d.PRG", n))
	if e != nil {
		t.Fatal(e)
	}
	return New(n, b)
}
func dotHash(d *Display) string {
	b := make([]byte, len(d.Dots))
	for i, on := range d.Dots {
		if on {
			b[i] = 1
		}
	}
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
func TestSourceVideoAndComposition(t *testing.T) {
	if CRTCHeight(HighCRTC) != 350 || (int(HighCRTC[1])+1)*4 != 320 || int(HighCRTC[0x13])*8 != 336 {
		t.Fatal("CRTC geometry")
	}
	if FieldHeight+MatrixHeight != FrameHeight || DotWidth*2 != Width || DotHeight*2+1 != MatrixHeight {
		t.Fatal("split geometry")
	}
	d := original(t, 1)
	var p [768]byte
	field := image.NewRGBA(image.Rect(0, 0, 320, 576))
	field.Pix[0] = 17
	f := Compose(field, d, p, 96, 242)
	if f.Rect != image.Rect(0, 0, 320, 350) || f.Pix[0] != 17 || f.RGBAAt(0, 317).R != 0 {
		t.Fatal("composition is not visible field + bottom split")
	}
}
func TestOriginalMatrixBitmapFixtures(t *testing.T) {
	b, e := os.ReadFile("../../analysis/pf8-visual-fixtures.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixtures map[string]struct {
		Animations   map[string][]string
		SpringSHA256 string `json:"spring_sha256"`
	}
	if e = json.Unmarshal(b, &fixtures); e != nil {
		t.Fatal(e)
	}
	for n := 1; n <= 2; n++ {
		d := original(t, n)
		f := fixtures[fmt.Sprint(n)]
		if fmt.Sprintf("%x", sha256.Sum256(d.SpringGraphics())) != f.SpringSHA256 {
			t.Fatal("spring source data")
		}
		for label, hashes := range f.Animations {
			d.Clear()
			for frame, want := range hashes {
				d.Bitmap(d.Content.Animations[label].Offsets[frame])
				if dotHash(d) != want {
					t.Fatalf("table %d %s frame %d", n, label, frame)
				}
			}
		}
	}
}
func TestTextGlyphPositionClippingAndScroll(t *testing.T) {
	for n := 1; n <= 2; n++ {
		d := original(t, n)
		d.Text("A", 2, 1, 13)
		font := d.Content.Fonts["13"]
		for row := 0; row < 13; row++ {
			for col := 0; col < 8; col++ {
				want := d.data[font+10*13+row]&(128>>uint(col)) != 0
				if d.Dots[(row+1)*160+col+2] != want {
					t.Fatal("original PRINT glyph")
				}
			}
		}
		d.Clear()
		d.Text("A", -4, 1, 13)
		for row := 0; row < 13; row++ {
			for col := 0; col < 4; col++ {
				if d.Dots[(row+1)*160+col] != (d.data[font+10*13+row]&(128>>uint(col+4)) != 0) {
					t.Fatal("left clipping")
				}
			}
		}
		d.Clear()
		d.Begin("_SCROLL", []string{"SCROLL_TEXT1"})
		d.Visit(0, 2, nil)
		one := d.Dots
		d.Visit(0, 2, nil)
		for row := 0; row < 16; row++ {
			for col := 0; col < 158; col++ {
				if d.Dots[row*160+col] != one[row*160+col+2] {
					t.Fatal("source scroller cadence")
				}
			}
		}
		// Rendering is never a timing authority.
		frozen := *d
		var p [768]byte
		out := image.NewRGBA(image.Rect(0, 0, 320, 350))
		d.Draw(out, p, 96, 242)
		d.Draw(out, p, 96, 242)
		if !reflect.DeepEqual(frozen, *d) {
			t.Fatal("Draw advanced state")
		}
	}
}
func TestSourceScoreAndAttractPrograms(t *testing.T) {
	for n := 1; n <= 2; n++ {
		d := original(t, n)
		d.Score("000000000000")
		for y := 0; y < 16; y++ {
			for x := 0; x < 152; x++ {
				if d.Dots[y*160+x] {
					t.Fatal("leading zeros printed")
				}
			}
		}
		var names, scores [4]string
		a := d.Attract(50, names, scores)
		b := d.Attract(50, names, scores)
		if a.Dots != b.Dots || a.Dots == ([2560]bool{}) {
			t.Fatal("attract source text/fixed counter")
		}
		if len(d.Content.Attract) < 40 {
			t.Fatal("ShowHighsTS macro missing")
		}
	}
}

func TestSourcePrintNumberAndGameOver(t *testing.T) {
	for n := 1; n <= 2; n++ {
		d := original(t, n)
		d.Begin("_PRINT13_NUMBER_CENT", []string{"FEMHUNDRATUSEN", "2*2*SW/4+TOTCENT"})
		d.Visit(0, 0, func(string) string { return "000000500000" })
		want := original(t, n)
		// PRINT_NUMBER's SCASB counts seven bytes for six visible digits;
		// DI=336+16 - 7*2, then skips six 4-byte source glyph cells.
		want.Text("500000", 52, 1, 13)
		if d.Dots != want.Dots {
			t.Fatal("centered BCD address/leading zero conversion")
		}
		d.Clear()
		d.Begin("_PRINT13_NUMBER_CENT", []string{"FEMHUNDRATUSEN", "2*2*SW/4+TOTCENT"})
		d.Visit(0, 0, func(string) string { return "000000000000" })
		if d.Dots != ([2560]bool{}) {
			t.Fatal("PRINT_NUMBER's all-zero BCD must stay blank")
		}
		var names, scores [4]string
		a, b := d.GameOver(68, "123456", names, scores), d.GameOver(68, "123456", names, scores)
		if a.Dots != b.Dots || a.Dots == ([2560]bool{}) {
			t.Fatal("single-player ShowIt score/frozen counter")
		}
		if d.GameOver(305, "123456", names, scores).Dots != d.Attract(0, names, scores).Dots {
			t.Fatal("UrbanOverTS did not transition to ShowHighsTS")
		}
	}
}

func TestSourceMatchWinningFlash(t *testing.T) {
	d := original(t, 1)
	d.MatchWin(0)
	for i := 0; i < 2; i++ {
		d.Flash()
		if !d.On {
			t.Fatal("early gladgnu flash")
		}
	}
	d.Flash()
	if d.On {
		t.Fatal("MATRIX_CNT=3 did not flash")
	}
	for i := 0; i < 3; i++ {
		d.Flash()
	}
	if !d.On {
		t.Fatal("MATRIX_SPEED=3 cadence")
	}
}

func TestSourceTableGameOverClearVariants(t *testing.T) {
	var names, scores [4]string
	party := original(t, 1).GameOver(63, "123456", names, scores)
	speed := original(t, 2).GameOver(63, "123456", names, scores)
	// UrbanOverTS expands CLEARIT to _CLEAR4 on Party Land and _CLEAR1 on
	// Speed Devils: the first removes interleaved rows, the second clears at once.
	if party.Dots == ([2560]bool{}) || speed.Dots != ([2560]bool{}) {
		t.Fatal("table-specific CLEARIT expansion")
	}
}

func TestSourceMatrixOffUsesLiteralDAC(t *testing.T) {
	d := original(t, 1)
	d.Dots[0] = true
	d.On = false
	var p [768]byte
	p[96*3] = 81
	p[96*3+1] = 81
	p[96*3+2] = 81
	frame := image.NewRGBA(image.Rect(0, 0, 320, 350))
	d.Draw(frame, p, 96, 242)
	if got := frame.RGBAAt(0, 319); got.R != 85 || got.G != 85 || got.B != 85 {
		t.Fatal("MATRIXOFF must not be recalculated as lamp percentages", got)
	}
	if frame.RGBAAt(2, 319).R != 81 {
		t.Fatal("unlit lattice changed")
	}
}

// Linked SCROLLE (Party 71d6..727e) sets AL=242/AH=96 and
// shifts existing glyphs. Literal stores are run boundaries: A row1
// starts at dot1 and clears at dot6, giving five illuminated dots.
func TestSourceScrollerRunBoundaries(t *testing.T) {
	want := []byte{0, 124, 254, 254, 238, 238, 254, 254, 254, 238, 238, 238, 238, 238, 0, 0}
	for n := 1; n <= 2; n++ {
		d := original(t, n)
		if !reflect.DeepEqual(d.Content.ScrollFont["65"], want) {
			t.Fatal("SCROLLE boundary integration")
		}
	}
}

// The unscaled DOSBox-X capture has repeated horizontal samples, verified by
// compare_pf8.py. This checks recovered artwork, not elapsed host time.
func TestSourceScrollerAgainstOriginalCapture(t *testing.T) {
	testinputs.Require(t, "../../analysis/pf8-checkpoints/verified-x-party-scroll-long.png")
	f, e := os.Open("../../analysis/pf8-checkpoints/verified-x-party-scroll-long.png")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	im, e := png.Decode(f)
	if e != nil {
		t.Fatal(e)
	}
	for n := 1; n <= 2; n++ {
		d := original(t, n)
		text := original(t, 1).Content.Texts["SCROLL_TEXT1"]
		d.GlyphText(string(text), -675, 0, d.Content.ScrollFont)
		for y := 0; y < 16; y++ {
			for x := 0; x < 160; x++ {
				r, _, _, _ := im.At(x*4, 319+y*2).RGBA()
				if d.Dots[y*160+x] != (r > 150*257) {
					t.Fatalf("table%d scroller dot %d,%d", n, x, y)
				}
			}
		}
	}
}

func TestGearAnimationAgainstLiveOriginalCaptures(t *testing.T) {
	// The external data-only probe dispatches the original GEARTS. It retains
	// original code, assets, renderer and callback cadence; see its provenance.
	base := "../../analysis/pf8-runtime-validation/bitmap-data-isolated/"
	testinputs.Require(t, "../../TABLE1.PRG", "../../TABLE2.PRG", base+"matches.json")
	raw, e := os.ReadFile(base + "matches.json")
	if e != nil {
		t.Fatal(e)
	}
	var captures [][]json.RawMessage
	if e = json.Unmarshal(raw, &captures); e != nil {
		t.Fatal(e)
	}
	d := original(t, 2)
	var pal [768]byte
	pal[98*3], pal[98*3+1], pal[98*3+2] = 81, 81, 81 // original DAC20
	pal = MatrixPalette(pal, 128)
	seen := map[int]bool{}
	for _, entry := range captures {
		var name string
		var frame int
		if e = json.Unmarshal(entry[0], &name); e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(entry[2], &frame); e != nil {
			t.Fatal(e)
		}
		seen[frame] = true
		d.Clear()
		for i := 0; i <= frame; i++ {
			d.Bitmap(d.Content.Animations["_GEAR"].Offsets[i])
		}
		out := image.NewRGBA(image.Rect(0, 0, 320, 350))
		d.Draw(out, pal, 98, 128)
		f, e := os.Open(base + name)
		if e != nil {
			t.Fatal(e)
		}
		im, e := png.Decode(f)
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
		if im.Bounds() != image.Rect(0, 0, 640, 33) {
			t.Fatal("raw matrix crop changed")
		}
		for y := 0; y < 33; y++ {
			for x := 0; x < 640; x++ {
				want := out.RGBAAt(x/2, y+317)
				r, g, b, a := im.At(x, y).RGBA()
				if r != uint32(want.R)*257 || g != uint32(want.G)*257 || b != uint32(want.B)*257 || a != 65535 {
					t.Fatalf("%s pixel%d,%d differs from original", name, x, y)
				}
			}
		}
	}
	if len(captures) != 69 || len(seen) != 5 {
		t.Fatal("live probe must cover all five distinct GEAR frames")
	}
}

func TestSourceNormalVideoAndComposition(t *testing.T) {
	if CRTCHeight(NormalCRTC) != 480 || NormalCRTC[9]&128 == 0 || int(NormalCRTC[0x18])+int(NormalCRTC[7]&16)*16 != 413 {
		t.Fatal("SET240/SETSPLIT")
	}
	field := image.NewRGBA(image.Rect(0, 0, 320, 207))
	field.Pix[0] = 17
	d := original(t, 1)
	var p [768]byte
	f := ComposeHeight(field, d, p, 96, 242, 207)
	if f.Rect != image.Rect(0, 0, 320, 240) || f.Pix[0] != 17 || f.RGBAAt(0, 207).R != 0 {
		t.Fatal("normal split")
	}
}
