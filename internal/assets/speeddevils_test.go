package assets

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func TestSpeedDevilsOriginalAssets(t *testing.T) {
	testinputs.Require(t, "../../TABLE2.PRG")
	data, e := os.ReadFile("../../TABLE2.PRG")
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile("../../analysis/pf7-assets.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Indices string `json:"indices_sha256"`
		RGBA    string `json:"rgba_sha256"`
		Initial string `json:"initial_rgba_sha256"`
		Ball    string `json:"ball_sha256"`
	}
	if e = json.Unmarshal(b, &fixture); e != nil {
		t.Fatal(e)
	}
	p, e := DecodeSpeedDevils(data)
	if e != nil {
		t.Fatal(e)
	}
	hash := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	if hash(p.Indices) != fixture.Indices || hash(p.Framebuffer().Pix) != fixture.RGBA {
		t.Fatal("original playfield fixture differs")
	}
	initial, e := DecodeInitialSpeedDevils(data)
	if e != nil {
		t.Fatal(e)
	}
	if initial.State.BallOrigin != image.Pt(300, 530) || initial.State.Viewport != image.Rect(0, 259, 320, 576) {
		t.Fatal(initial.State)
	}
	if hash(initial.Ball[:]) != fixture.Ball || hash(initial.Framebuffer().Pix) != fixture.Initial {
		t.Fatal("original initial frame differs")
	}
	data[0x50730] ^= 1
	if _, e := DecodeSpeedDevils(data); e == nil {
		t.Fatal("must reject broken FORM")
	}
}
