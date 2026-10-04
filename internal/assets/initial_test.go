package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func TestPartyLandInitialState(t *testing.T) {
	s := PartyLandInitialState()
	if s.Viewport != image.Rect(0, 259, 320, 576) || s.Viewport.Size() != image.Pt(320, 317) {
		t.Fatalf("initial viewport: %v", s.Viewport)
	}
	if s.BallOrigin != image.Pt(297, 530) || s.BallHigh {
		t.Fatalf("SETBALL state: %+v", s)
	}
	if s.BallOrigin.Sub(s.Viewport.Min) != image.Pt(297, 271) {
		t.Fatal("ball hotspot/origin in viewport")
	}
}

func TestPartyLandInitialReference(t *testing.T) {
	testinputs.Require(t, "../../TABLE1.PRG")
	data, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.ReadFile("../../analysis/pf2-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		BallSHA       string `json:"ball_sha256"`
		ForegroundSHA string `json:"foreground_sha256"`
		RGBASHA       string `json:"rgba_sha256"`
		PF1SHA        string `json:"pf1_rgba_sha256"`
	}
	if err := json.Unmarshal(f, &fixture); err != nil {
		t.Fatal(err)
	}
	hash := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	p, err := DecodeInitialPartyLand(data)
	if err != nil {
		t.Fatal(err)
	}
	if hash(p.Ball[:]) != fixture.BallSHA || hash(p.foreground) != fixture.ForegroundSHA {
		t.Fatal("original ball graphics or lower-plane foreground differs")
	}
	// Pin the graphic's shape independently of the extracted offset map.
	if !bytes.Equal(p.Ball[:16], []byte{0, 0, 0, 0, 0, 21, 21, 21, 28, 36, 0, 0, 0, 0, 0, 0}) {
		t.Fatal("sprite top row / hotspot")
	}
	if !bytes.Equal(p.Ball[15*16:], make([]byte, 16)) {
		t.Fatal("nominal transparent bottom row")
	}
	baseline := p.Playfield.Framebuffer()
	frame := p.Framebuffer()
	if hash(frame.Pix) != fixture.RGBASHA {
		t.Fatal("initial framebuffer differs from independent Python composition")
	}
	if hash(baseline.Pix) != fixture.PF1SHA || hash(p.Playfield.Framebuffer().Pix) != fixture.PF1SHA {
		t.Fatal("PF2 modified PF1 baseline")
	}
	if !bytes.Equal(frame.Pix, p.Framebuffer().Pix) {
		t.Fatal("static rerender differs")
	}
	// Every sprite pixel must obey foreground ordering, including pixels at
	// non-byte-aligned x=297. Transparent pixels retain the background.
	visible, masked, hidden := 0, 0, 0
	for i, c := range p.Ball {
		x, y := 297+i%16, 530+i/16
		want := baseline.Pix[baseline.PixOffset(x, y) : baseline.PixOffset(x, y)+4]
		if c != 0 {
			if p.foreground[y*40+x/8]&(128>>uint(x&7)) != 0 {
				masked++
			} else {
				visible++
				want = append(append([]byte(nil), p.Playfield.Palette[int(c)*3:int(c)*3+3]...), 255)
			}
			hx := 282 + i%16
			if p.foreground[y*40+hx/8]&(128>>uint(hx&7)) != 0 {
				hidden++
			}
		}
		o := frame.PixOffset(x, y-259)
		if !bytes.Equal(frame.Pix[o:o+4], want) {
			t.Fatalf("pixel %d,%d ordering", x, y)
		}
	}
	if visible != 152 || masked != 25 || hidden != 177 {
		t.Fatalf("visible=%d masked=%d earlier hidden=%d", visible, masked, hidden)
	}
	// Decoded state owns its content; original bytes may be released/changed.
	data[foregroundOffset] ^= 255
	if !bytes.Equal(frame.Pix, p.Framebuffer().Pix) {
		t.Fatal("retained input alias")
	}
	if _, err := DecodeInitialPartyLand(data); err != nil {
		t.Fatal("compatible foreground rejected", err)
	}
	if _, err := DecodeInitialPartyLand(nil); err == nil {
		t.Fatal("truncated installation accepted")
	}
}
