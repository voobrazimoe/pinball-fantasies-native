package gamepad

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"pinballfantasies/internal/frontend"
	"testing"
)

func TestContextualHints(t *testing.T) {
	if Hints(frontend.Playing, false, true, false) != nil || Hints(frontend.Initials, true, true, false) != nil {
		t.Fatal("hints during play or initials")
	}
	full := Hints(frontend.Selector, false, false, false)
	demo := Hints(frontend.Selector, false, false, true)
	if len(full) != 6 || len(demo) != 3 || full[2].Button != X || full[2].Text != "BILLION DOLLAR GAMESHOW" {
		t.Fatal("table identities", full, demo)
	}
	if len(Hints(frontend.Playing, true, true, false)) != 6 || len(Hints(frontend.Playing, true, false, false)) != 5 {
		t.Fatal("player selection window")
	}
}

func TestGlyphIdentityAndDisconnect(t *testing.T) {
	p := New()
	p.Connect(0)
	p.SetFamily(PlayStation)
	for b, g := range []Glyph{Cross, Circle, Square, Triangle} {
		if p.Glyph(b) != g {
			t.Fatal(b, p.Glyph(b))
		}
	}
	p.SetGlyph(A, Circle)
	if p.Glyph(A) != Circle {
		t.Fatal("host override")
	}
	p.SetFamily(Nintendo)
	if p.Glyph(A) != LetterB || p.Glyph(X) != LetterY {
		t.Fatal("Nintendo positions")
	}
	p.Disconnect()
	if p.Connected() || p.Glyph(A) != LetterA {
		t.Fatal("stale identity")
	}
}

func TestOverlayOwnsPixelsAndProtectsMatrix(t *testing.T) {
	p := New()
	src := image.NewRGBA(image.Rect(0, 0, 320, 240))
	draw.Draw(src, src.Rect, image.NewUniform(color.RGBA{90, 110, 130, 255}), image.Point{}, draw.Src)
	original := append([]byte(nil), src.Pix...)
	rows := Hints(frontend.Selector, false, false, false)
	if Overlay(nil, src, p, rows) != src {
		t.Fatal("disconnected hints")
	}
	p.Connect(0)
	p.SetFamily(PlayStation)
	dst := Overlay(nil, src, p, rows)
	if dst == src || !bytes.Equal(src.Pix, original) || bytes.Equal(dst.Pix, src.Pix) {
		t.Fatal("source frame changed or no overlay")
	}
	if dst.Rect != src.Rect {
		t.Fatal("hints resized the frame", dst.Rect)
	}
	// The upper half, with the dot-matrix display, stays untouched.
	half := src.Rect.Dy() / 2 * src.Stride
	if !bytes.Equal(dst.Pix[:half], src.Pix[:half]) {
		t.Fatal("hints cover the upper frame")
	}
	if Overlay(dst, src, p, rows) != dst {
		t.Fatal("host allocation not reused")
	}
	if Overlay(dst, src, p, Hints(frontend.Playing, false, false, false)) != src {
		t.Fatal("visible during ball play")
	}
	if c := dst.RGBAAt(12, src.Rect.Max.Y-6); c.R == 12 || c == src.RGBAAt(12, src.Rect.Max.Y-6) {
		t.Fatal("panel missing or opaque", c)
	}
}

func TestSelectorHintsKeepFrameAndScanlineAspect(t *testing.T) {
	p := New()
	p.Connect(0)
	src := image.NewRGBA(image.Rect(0, 0, 640, 240))
	for y := 0; y < 240; y++ {
		for x := 0; x < 640; x++ {
			src.SetRGBA(x, y, color.RGBA{uint8(x), uint8(y), 99, 255})
		}
	}
	dst := Overlay(nil, src, p, Hints(frontend.Options, false, false, false))
	if dst.Rect != src.Rect {
		t.Fatal("hints resized the frame", dst.Rect)
	}
	for y := 0; y < 120; y++ {
		for x := 0; x < 640; x++ {
			if dst.RGBAAt(x, y) != src.RGBAAt(x, y) {
				t.Fatal("covered upper selector", x, y)
			}
		}
	}
	// Panel pixels come in horizontal pairs, matching the doubled scanlines.
	ink := 0
	for y := 120; y < 240; y++ {
		for x := 20; x < 620; x += 2 {
			if dst.RGBAAt(x, y) == (color.RGBA{240, 245, 250, 255}) {
				ink++
				if dst.RGBAAt(x+1, y) != dst.RGBAAt(x, y) {
					t.Fatal("panel not doubled horizontally", x, y)
				}
			}
		}
	}
	if ink == 0 {
		t.Fatal("no hint text drawn")
	}
}
