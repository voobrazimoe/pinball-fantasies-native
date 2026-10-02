package assets

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func TestFrontendContent(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG")
	b, e := os.ReadFile("../../INTRO.PRG")
	if e != nil {
		t.Fatal(e)
	}
	a, e := DecodeFrontend(b)
	if e != nil {
		t.Fatal(e)
	}
	if a.Logo.Width != 640 || a.Logo.Height != 256 || a.Font.Height != 28 || a.HighLogo.Height != 200 {
		t.Fatal("asset dimensions")
	}
	for i, p := range a.Tables {
		if p.Width != 640 || p.Height < 95 || len(p.Indices) != p.Width*p.Height {
			t.Fatalf("table %d", i)
		}
	}
	again, e := DecodeFrontend(b)
	if e != nil {
		t.Fatal(e)
	}
	if sha256.Sum256(a.Logo.Frame().Pix) != sha256.Sum256(again.Logo.Frame().Pix) {
		t.Fatal("nondeterministic frontend decode")
	}
	b[0] ^= 1
	if _, e := DecodeFrontend(b); e == nil {
		t.Fatal("unknown build accepted")
	}
}
func TestFrontendIFFValidation(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG")
	b, e := os.ReadFile("../../INTRO.PRG")
	if e != nil {
		t.Fatal(e)
	}
	for _, off := range []int{-1, 0, len(b) - 2} {
		if _, e := decodeFrontendIFF(b, off); e == nil {
			t.Fatal("bad FORM accepted")
		}
	}
	for _, n := range []int{12, 80, 500} {
		if _, e := decodeFrontendIFF(b[:0x6b70+n], 0x6b70); e == nil {
			t.Fatal("truncated IFF accepted")
		}
	}
}
func TestFrontendIndependentFixture(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG")
	b, e := os.ReadFile("../../INTRO.PRG")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("../../analysis/pf6-art-fixtures.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture []struct {
		Offset, Width, Height int
		Indices, RGBA         string
	}
	if e := json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}
	for _, f := range fixture {
		p, e := decodeFrontendIFF(b, f.Offset)
		if e != nil {
			t.Fatal(e)
		}
		if p.Width != f.Width || p.Height != f.Height || fmt.Sprintf("%x", sha256.Sum256(p.Indices)) != f.Indices || fmt.Sprintf("%x", sha256.Sum256(p.Frame().Pix)) != f.RGBA {
			t.Fatalf("IFF at %x differs from independent oracle", f.Offset)
		}
	}
}

// Pin only hashes of the previously accepted instruction strings. The strings
// themselves now come from INTRO.PRG, including spacing and punctuation.
func TestIntroSidebarRecordParity(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG")
	data, err := os.ReadFile("../../INTRO.PRG")
	if err != nil {
		t.Fatal(err)
	}
	a, err := DecodeFrontend(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ text, hash string }{
		{a.SidebarInfo, "49309e672f96348d01d18d308220522d035442c392539976b3f2207a4d6797a9"}, {a.OptionsInfo, "a1e05a13b984d8cb0dd35e46b876ebcbf82ffbcbf55f9258ce36f0011b913d46"},
	} {
		if len(tc.text) != 120 || fmt.Sprintf("%x", sha256.Sum256([]byte(tc.text))) != tc.hash {
			t.Fatal("intro instruction record parity")
		}
	}
}
