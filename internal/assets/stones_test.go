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

func TestStonesOriginalAssets(t *testing.T) {
	testinputs.Require(t, "../../TABLE4.PRG")
	data, e := os.ReadFile("../../TABLE4.PRG")
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile("../../analysis/pf10-assets.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Indices    string `json:"indices_sha256"`
		RGBA       string `json:"rgba_sha256"`
		Initial    string `json:"initial_rgba_sha256"`
		Ball       string `json:"ball_sha256"`
		Foreground string `json:"foreground_sha256"`
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	p, e := DecodeInitialStones(data)
	if e != nil {
		t.Fatal(e)
	}
	hash := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	if len(p.Playfield.Indices) != 320*576 || hash(p.Playfield.Indices) != f.Indices || hash(p.Playfield.Framebuffer().Pix) != f.RGBA || hash(p.Ball[:]) != f.Ball || hash(p.foreground) != f.Foreground || hash(p.Framebuffer().Pix) != f.Initial {
		t.Fatal("independent original content hashes")
	}
	if p.State.Viewport != image.Rect(0, 259, 320, 576) || p.State.BallOrigin != image.Pt(297, 530) {
		t.Fatal(p.State)
	}
	data[0x4bc10] ^= 1
	if _, e = DecodeStones(data); e == nil {
		t.Fatal("accepts broken FORM")
	}
}
