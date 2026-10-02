package platform

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestGDITransferRowsStrideChannels(t *testing.T) {
	parent := image.NewRGBA(image.Rect(0, 0, 5, 4))
	f := parent.SubImage(image.Rect(1, 1, 3, 3)).(*image.RGBA)
	f.SetRGBA(1, 1, color.RGBA{R: 1, G: 2, B: 3, A: 255})
	f.SetRGBA(2, 1, color.RGBA{R: 4, G: 5, B: 6, A: 255})
	f.SetRGBA(1, 2, color.RGBA{R: 7, G: 8, B: 9, A: 255})
	f.SetRGBA(2, 2, color.RGBA{R: 10, G: 11, B: 12, A: 255})
	want := []byte{3, 2, 1, 0, 6, 5, 4, 0, 9, 8, 7, 0, 12, 11, 10, 0}
	if got := frameBGRA(nil, f); !bytes.Equal(got, want) {
		t.Fatalf("BGRX transfer: %v", got)
	}
}
func TestClientAspectAtDPIAndAllModes(t *testing.T) {
	for _, dpi := range []int{100, 125, 150} {
		for _, logical := range []image.Point{image.Pt(320, 240), image.Pt(320, 350), image.Pt(320, 609), image.Pt(640, 480)} {
			for _, client := range []image.Point{image.Pt(800*dpi/100, 600*dpi/100), image.Pt(1920*dpi/100, 1080*dpi/100), image.Pt(301, 777)} {
				dst := aspectRect(client, logical)
				if !dst.In(image.Rectangle{Max: client}) || dst.Min.X != (client.X-dst.Dx())/2 || dst.Min.Y != (client.Y-dst.Dy())/2 {
					t.Fatal(client, logical, dst)
				}
				delta := dst.Dx()*logical.Y - dst.Dy()*logical.X
				if delta < -logical.Y || delta > logical.X {
					t.Fatalf("aspect: %v %v %v", client, logical, dst)
				}
			}
		}
	}
}
func TestWin32ModifierBreakAndShortcut(t *testing.T) {
	for _, extended := range []uintptr{0, 1 << 24} {
		k := windowsKeys{}
		shortcut := hostKeys{}
		k.key(0x12, extended, true)
		if k.held() == 0 {
			t.Fatal("Alt flipper")
		}
		e := k.key(13, 1<<29, true)[0]
		if shortcut.down(e.key, e.alt, e.repeat) != toggleFullscreen {
			t.Fatal("Alt+Enter")
		}
		e = k.key(13, 1<<29|1<<30, true)[0]
		if shortcut.down(e.key, e.alt, e.repeat) != ignoreKey {
			t.Fatal("repeat leaked")
		}
		shortcut.enterUp()
		k.key(13, 0, false)
		k.key(0x12, extended, false)
		if k.held() != 0 {
			t.Fatal("stuck Alt")
		}
		e = k.key(13, 0, true)[0]
		if shortcut.down(e.key, e.alt, e.repeat) != dispatchKey {
			t.Fatal("plain Enter")
		}
	}
	k := windowsKeys{}
	k.key(0x10, 0x36<<16, true)
	if k.held() != 2 {
		t.Fatal("right shift")
	}
	k.key(0x11, 1<<24, true)
	k.key(40, 1<<24, true)
	k.key(32, 0, true)
	if k.held() != 14 {
		t.Fatal("held contract")
	}
	k = windowsKeys{}
	if k.held() != 0 {
		t.Fatal("focus reset")
	}
}

func TestWin32ExtendedKeyIdentity(t *testing.T) {
	k := windowsKeys{}
	if e := k.key(13, 1<<24, true)[0]; e.key != 127 {
		t.Fatal("keypad Enter became ordinary Enter")
	}
	if len(k.key(13, 1<<24, false)) != 0 {
		t.Fatal("keypad Enter released ordinary Enter")
	}
	if e := k.key(40, 0, true)[0]; e.key != 127 || k.held() != 0 {
		t.Fatal("keypad down became plunger")
	}
	if e := k.key(40, 1<<24, true)[0]; e.key != 80 || k.held() != 4 {
		t.Fatal("arrow down contract")
	}
	k.key(40, 1<<24, false)
	if k.held() != 0 {
		t.Fatal("arrow release")
	}
	// Physical Q position still maps DOS Q even with an AZERTY virtual A.
	if e := k.key('A', 16<<16, true)[0]; e.key != 16 {
		t.Fatal("physical letter scan mapping")
	}
}
