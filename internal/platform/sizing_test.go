package platform

import (
	"image"
	"os"
	"strings"
	"testing"
)

func TestInitialWindowSize(t *testing.T) {
	for _, tc := range []struct{ desktop, want image.Point }{
		{image.Pt(1920, 1080), image.Pt(640, 766)},
		{image.Pt(2560, 1440), image.Pt(960, 1149)},
		// Usable bounds exclude ordinary desktop panels/decorations.
		{image.Pt(1920, 1040), image.Pt(640, 766)},
		{image.Pt(2560, 1400), image.Pt(960, 1149)},
		{image.Pt(3840, 2160), image.Pt(1280, 1532)},
		{image.Pt(800, 600), image.Pt(640, 480)},
	} {
		if got := initialWindowSize(tc.desktop); got != tc.want {
			t.Errorf("%v: got %v want %v", tc.desktop, got, tc.want)
		}
	}
	for _, desktop := range []image.Point{image.Pt(480, 320), image.Pt(1366, 768), image.Pt(5120, 1440)} {
		got := initialWindowSize(desktop)
		if got.X > desktop.X*17/20 || got.Y > desktop.Y*17/20 {
			t.Fatalf("does not fit %v: %v", desktop, got)
		}
	}
}

func TestFrontendNeverResizesPhysicalWindow(t *testing.T) {
	b, e := os.ReadFile("frontend.go")
	if e != nil {
		t.Fatal(e)
	}
	s := string(b)
	if strings.Contains(s, "SDL_SetWindowSize") || strings.Count(s, "C.SDL_CreateWindow(") != 1 {
		t.Fatal("frontend must own one persistent physical window")
	}
	if logicalSize(image.Pt(640, 240)) != image.Pt(640, 480) {
		t.Fatal("selector scanline aspect")
	}
}

func TestNativePresentationAspects(t *testing.T) {
	for _, h := range []int{240, 350, 609} {
		if logicalSize(image.Pt(320, h)) != image.Pt(320, h) {
			t.Fatal(h)
		}
	}
}
