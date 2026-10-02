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

func TestSourceGriffinPalette(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG")
	b, e := os.ReadFile("../../INTRO.PRG")
	if e != nil {
		t.Fatal(e)
	}
	a, e := DecodeFrontend(b)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("../../analysis/pf8-visual-fixtures.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Griffin struct {
			PaletteSHA256 string `json:"palette_sha256"`
			Colors        map[string][]byte
		}
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(a.Startup[0].Palette)) != f.Griffin.PaletteSHA256 {
		t.Fatal("linked UNPKLBM palette bank conversion")
	}
	for index, rgb := range f.Griffin.Colors {
		var n int
		fmt.Sscan(index, &n)
		for c, v := range rgb {
			if a.Startup[0].Palette[n*3+c] != v || a.Startup[1].Palette[n*3+c] != v {
				t.Fatal("saved pelle1 palette")
			}
		}
	}
}
func TestSharedVGADACConversion(t *testing.T) {
	var p [768]byte
	p[96*3] = 83
	p[0] = 255
	p[1] = 128
	q := VGAPalette(p)
	if q[96*3] != 81 || q[0] != 255 || q[1] != 130 {
		t.Fatal("CMAP >>2 / six-bit expansion")
	}
	if p[96*3] != 83 {
		t.Fatal("raw source palette mutated")
	}
}

func TestSourceBallOccupiedGeometry(t *testing.T) {
	testinputs.Require(t, "../../TABLE1.PRG")
	b, e := os.ReadFile("../../TABLE1.PRG")
	if e != nil {
		t.Fatal(e)
	}
	table, e := DecodeInitialPartyLand(b)
	if e != nil {
		t.Fatal(e)
	}
	bounds := image.Rectangle{}
	for i, v := range table.Ball {
		if v != 0 {
			r := image.Rect(i%16, i/16, i%16+1, i/16+1)
			bounds = bounds.Union(r)
		}
	}
	if bounds.Dx() != 15 || bounds.Dy() != 15 {
		t.Fatal("original source-art ball diameters", bounds)
	}
	// Equal diameters support the 1:1 artwork-geometry scaling policy. This
	// deliberately does not establish a historical monitor active-area aspect.
}
